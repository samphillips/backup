package file

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/samphillips/backup/internal/logging"
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

// GenerateBackupDetails determines the list of directories and files to create in the backup
// location as well as the total size of the files to copy
func GenerateBackupDetails(srcIndex, dstIndex map[string]os.FileInfo, srcDir, dstDir string) (filesToCopy, directoriesToCreate []string, totalSize int64) {
	for srcPath, srcFile := range srcIndex {
		logging.Info(srcPath)
		if dstFile, ok := dstIndex[srcPath]; ok {
			if srcFile.IsDir() {
				continue
			}

			if srcFile.Size() == dstFile.Size() {
				srcSum, err := hashFile(filepath.Join(srcDir + srcPath))

				if err != nil {
					logging.Warn("Could not calculate md5 hashsum of file: %s", srcDir+srcPath)
				}

				dstSum, err := hashFile(filepath.Join(dstDir + srcPath))

				if err != nil {
					logging.Warn("Could not calculate md5 hashsum of file: %s", dstDir+srcPath)
				}

				if srcSum != dstSum {
					filesToCopy = append(filesToCopy, srcPath)
					totalSize += srcFile.Size()
				}
			} else {
				filesToCopy = append(filesToCopy, srcPath)
				totalSize += srcFile.Size()
			}
		} else {
			if srcFile.IsDir() {
				directoriesToCreate = append(directoriesToCreate, srcPath)
			} else {
				filesToCopy = append(filesToCopy, srcPath)
				totalSize += srcFile.Size()
			}
		}
	}
	return
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
