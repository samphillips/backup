package main

import (
	"os"
	"path/filepath"

	"github.com/samphillips/backup/internal/config"
	"github.com/samphillips/backup/internal/file"
	"github.com/samphillips/backup/internal/logging"
	"github.com/samphillips/backup/internal/progress"
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

	logging.Info("Determining files to be backed up")
	files, directories := file.GenerateBackupDetails(srcIndex, dstIndex, config.SrcDir, config.DstDir)

	logging.Info("Creating new directories")
	bar := progress.Start(len(directories) + 1)
	for _, dir := range directories {
		bar.Increment()
		logging.Debug("Create directory %s", filepath.Join(config.DstDir, dir))
		os.MkdirAll(filepath.Join(config.DstDir, dir), os.ModePerm)
	}
	bar.Increment()
	bar.Finish()

	logging.Info("Copying files")

	bar = progress.Start(len(files) + 1)
	for _, f := range files {
		bar.Increment()
		logging.Debug("Copying %s to backup location %s", filepath.Join(config.SrcDir, f), filepath.Join(config.DstDir, f))
		err := file.CopyFile(filepath.Join(config.SrcDir, f), filepath.Join(config.DstDir, f))
		if err != nil {
			logging.Error("Failed to copy file %s: %s", filepath.Join(config.SrcDir, f), err)
		}
	}
	bar.Increment()
	bar.Finish()

	if config.Mirror {
		logging.Info("Removing excess files in backup directory")
		bar = progress.Start(len(dstIndex) + 1)
		for dstPath := range dstIndex {
			bar.Increment()
			if _, ok := srcIndex[dstPath]; !ok {
				logging.Debug("Removing file %s", filepath.Join(config.DstDir, dstPath))
				err := os.RemoveAll(dstPath)
				if err != nil {
					logging.Error("Failed to remove file %s: %s", filepath.Join(config.DstDir, dstPath), err)
				}
			}
		}
		bar.Increment()
		bar.Finish()
	}
}
