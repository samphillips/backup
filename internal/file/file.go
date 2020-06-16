package file

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/samphillips/backup/internal/logging"
	"github.com/samphillips/backup/internal/progress"
)

// HashFile generates the md5 sum hash string of a file
func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashBytes := hash.Sum(nil)[:16]
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}

type srcDetails struct {
	srcPath string
	srcFile os.FileInfo
}

type backupDetails struct {
	files       []string
	directories []string
	symlinks    map[string]string
}

func worker(dstIndex map[string]os.FileInfo, srcDir, dstDir string, skipHashsum bool, jobs <-chan srcDetails, results chan<- backupDetails) {
	b := backupDetails{
		files:       []string{},
		directories: []string{},
		symlinks:    map[string]string{},
	}

	for j := range jobs {
		if dstFile, ok := dstIndex[j.srcPath]; ok {
			if j.srcFile.IsDir() {
				logging.Debug("Skipping %s as directory already exists at backup location", j.srcPath)
				continue
			}

			if j.srcFile.Mode()&os.ModeSymlink != 0 {
				srcLink, err := os.Readlink(filepath.Join(srcDir, j.srcPath))
				if err != nil {
					logging.Warn("Error reading file %s symlink: %s", j.srcPath, err)
					continue
				}
				dstLink, err := os.Readlink(filepath.Join(dstDir, j.srcPath))
				if err != nil || strings.TrimPrefix(srcLink, srcDir) != strings.TrimPrefix(dstLink, dstDir) {
					logging.Debug("Marking symlink at %s for backup", j.srcPath)
					if strings.HasPrefix(srcLink, srcDir) {
						srcLink = filepath.Join(dstDir, strings.TrimPrefix(srcLink, srcDir))
					}
					b.symlinks[j.srcPath] = srcLink
					continue
				}
			}

			if j.srcFile.Size() == dstFile.Size() {
				if skipHashsum {
					logging.Debug("Skipping %s as the file size has not changed and hashsum skip is enabled", j.srcPath)
					continue
				}

				srcSum, err := hashFile(filepath.Join(srcDir, j.srcPath))

				if err != nil {
					logging.Warn("Could not calculate md5 hashsum of file: %s", j.srcPath)
				}

				dstSum, err := hashFile(filepath.Join(dstDir, j.srcPath))

				if err != nil {
					logging.Warn("Could not calculate md5 hashsum of file: %s", j.srcPath)
				}

				if srcSum != dstSum {
					logging.Debug("Marking %s for backup as file hashsum is different to file at backup location", j.srcPath)
					b.files = append(b.files, j.srcPath)
				} else {
					logging.Debug("Skipping %s as the file has not changed", j.srcPath)
				}
			} else {
				logging.Debug("Marking %s for backup as file size is different to file at backup location", j.srcPath)
				b.files = append(b.files, j.srcPath)
			}
		} else {
			if j.srcFile.IsDir() {
				logging.Debug("Marking %s for creation as directory does not exist at backup location", j.srcPath)
				b.directories = append(b.directories, j.srcPath)
			} else {
				if j.srcFile.Mode()&os.ModeSymlink != 0 {
					link, err := os.Readlink(filepath.Join(srcDir, j.srcPath))
					if err != nil {
						logging.Warn("Error reading file %s symlink: %s", j.srcPath, err)
						continue
					}
					logging.Debug("Marking symlink at %s for backup", j.srcPath)
					b.symlinks[j.srcPath] = strings.TrimPrefix(link, srcDir)
					continue
				}

				logging.Debug("Marking %s for backup as file does not exist at backup location", j.srcPath)
				b.files = append(b.files, j.srcPath)
			}
		}
	}

	results <- b
}

// GenerateBackupDetails determines the list of directories, files and symlinks to create in the
// backup location
func GenerateBackupDetails(srcIndex, dstIndex map[string]os.FileInfo, srcDir, dstDir string, skipHashsum bool) ([]string, []string, map[string]string) {
	var files, directories []string
	symlinks := map[string]string{}

	numWorkers := int(math.Ceil(float64(len(srcIndex)) / 100.0))
	jobs := make(chan srcDetails, numWorkers)
	results := make(chan backupDetails, numWorkers)

	for w := 0; w < numWorkers; w++ {
		go worker(dstIndex, srcDir, dstDir, skipHashsum, jobs, results)
	}

	bar := progress.Start(len(srcIndex) + 1 + numWorkers)
	for srcPath, srcFile := range srcIndex {
		bar.Increment()
		jobs <- srcDetails{
			srcPath: srcPath,
			srcFile: srcFile,
		}
	}

	close(jobs)

	for w := 0; w < numWorkers; w++ {
		bar.Increment()
		r := <-results
		files = append(files, r.files...)
		directories = append(directories, r.directories...)
		for k, v := range r.symlinks {
			symlinks[k] = v
		}
	}
	bar.Increment()
	bar.Finish()

	return files, directories, symlinks
}

// CopyFile copies the source file to the destination file
func CopyFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
