package main

import (
	"os"
	"path/filepath"

	"github.com/samphillips/backup/internal/config"
	"github.com/samphillips/backup/internal/file"
	"github.com/samphillips/backup/internal/logging"
)

func main() {
	config := config.ParseConfig()

	if config.Verbose {
		logging.SetLogLevel(logging.DEBUG)
	}

	srcSDChan := make(chan map[string]os.FileInfo)
	dstSDChan := make(chan map[string]os.FileInfo)

	logging.Debug("Scanning source and destination directories")
	go func() {
		srcIndex := file.ScanDirectory(config.SrcDir)
		srcSDChan <- srcIndex
	}()
	go func() {
		dstIndex := file.ScanDirectory(config.DstDir)
		dstSDChan <- dstIndex
	}()

	var srcIndex, dstIndex map[string]os.FileInfo

	for i := 0; i < 2; i++ {
		select {
		case srcIndex = <-srcSDChan:
			logging.Debug("Finished scanning source directory")
		case dstIndex = <-dstSDChan:
			logging.Debug("Finished scanning destination directory")
		}
	}

	close(srcSDChan)
	close(dstSDChan)

	files, directories, _ := file.GenerateBackupDetails(srcIndex, dstIndex, config.SrcDir, config.DstDir)

	for _, dir := range directories {
		os.MkdirAll(filepath.Join(config.DstDir, dir), os.ModePerm)
	}

	for _, f := range files {
		err := file.CopyFile(filepath.Join(config.SrcDir, f), filepath.Join(config.DstDir, f))
		if err != nil {
			logging.Error("Failed to copy file %s: %s", filepath.Join(config.SrcDir, f), err)
		}
	}

	if config.Mirror {
		for dstPath := range dstIndex {
			if _, ok := srcIndex[dstPath]; !ok {
				err := os.RemoveAll(dstPath)
				if err != nil {
					logging.Error("Failed to remove file %s: %s", filepath.Join(config.DstDir, dstPath), err)
				}
			}
		}
	}
}
