package file

import (
	"crypto/md5"
	"encoding/hex"
	"io"
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

// GenerateBackupDetails determines the list of directories, files and symlinks to create in the
// backup location
func GenerateBackupDetails(srcIndex, dstIndex map[string]os.FileInfo, srcDir, dstDir string) ([]string, []string, map[string]string) {
	bar := progress.Start(len(srcIndex) + 1)

	var files, directories []string
	symlinks := map[string]string{}

	for srcPath, srcFile := range srcIndex {
		bar.Increment()
		if dstFile, ok := dstIndex[srcPath]; ok {
			if srcFile.IsDir() {
				logging.Debug("Skipping %s as directory already exists at backup location", srcPath)
				continue
			}

			if srcFile.Mode()&os.ModeSymlink != 0 {
				srcLink, err := os.Readlink(filepath.Join(srcDir, srcPath))
				if err != nil {
					logging.Warn("Error reading file %s symlink: %s", srcPath, err)
					continue
				}
				dstLink, err := os.Readlink(filepath.Join(dstDir, srcPath))
				if err != nil || strings.TrimPrefix(srcLink, srcDir) != strings.TrimPrefix(dstLink, dstDir) {
					logging.Debug("Marking symlink at %s for backup", srcPath)
					if strings.HasPrefix(srcLink, srcDir) {
						srcLink = filepath.Join(dstDir, strings.TrimPrefix(srcLink, srcDir))
					}
					symlinks[srcPath] = srcLink
					continue
				}
			}

			if srcFile.Size() == dstFile.Size() {
				srcSum, err := hashFile(filepath.Join(srcDir, srcPath))

				if err != nil {
					logging.Warn("Could not calculate md5 hashsum of file: %s", srcPath)
				}

				dstSum, err := hashFile(filepath.Join(dstDir, srcPath))

				if err != nil {
					logging.Warn("Could not calculate md5 hashsum of file: %s", srcPath)
				}

				if srcSum != dstSum {
					logging.Debug("Marking %s for backup as file hashsum is different to file at backup location", srcPath)
					files = append(files, srcPath)
				} else {
					logging.Debug("Skipping %s as the file has not changed", srcPath)
				}
			} else {
				logging.Debug("Marking %s for backup as file size is different to file at backup location", srcPath)
				files = append(files, srcPath)
			}
		} else {
			if srcFile.IsDir() {
				logging.Debug("Marking %s for creation as directory does not exist at backup location", srcPath)
				directories = append(directories, srcPath)
			} else {
				if srcFile.Mode()&os.ModeSymlink != 0 {
					link, err := os.Readlink(filepath.Join(srcDir, srcPath))
					if err != nil {
						logging.Warn("Error reading file %s symlink: %s", srcPath, err)
						continue
					}
					logging.Debug("Marking symlink at %s for backup", srcPath)
					symlinks[srcPath] = strings.TrimPrefix(link, srcDir)
					continue
				}

				logging.Debug("Marking %s for backup as file does not exist at backup location", srcPath)
				files = append(files, srcPath)
			}
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
