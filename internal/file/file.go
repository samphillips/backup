package file

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"

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
				srcSum, err := hashFile(srcDir + srcPath)

				if err != nil {
					logging.Warn("Could not calculate md5 hashsum of file: %s", srcDir+srcPath)
				}

				dstSum, err := hashFile(dstDir + srcPath)

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
