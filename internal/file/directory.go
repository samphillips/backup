package file

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/samphillips/backup/internal/logging"
)

// ScanDirectory obtains the details of all files and directories in a given directory recursively
func ScanDirectory(dirPath string) map[string]os.FileInfo {
	files := map[string]os.FileInfo{}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		shortPath := strings.TrimPrefix(path, dirPath)
		files[shortPath] = info
		return nil
	})

	if err != nil {
		logging.Fatal("Error encountered while scanning directory %s: %s", dirPath, err)
	}

	return files
}
