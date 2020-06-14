package config

import (
	"flag"
)

// Config contains the validated flags
type Config struct {
	srcDir string
	dstDir string
}

// ParseConfig parses the command line flags and validates them
func ParseConfig() (Config, error) {
	var srcDir, dstDir string
	flag.StringVar(&srcDir, "src-dir", "", "(Required) The absolute directory path you wish to back up")
	flag.StringVar(&dstDir, "dst-dir", "", "(Required) The absolute directory that the source directory will be backed up to")
	flag.Parse()

	if srcDir == "" || dstDir == "" {

		flag.PrintDefaults()
	}

	return Config{
		srcDir: srcDir,
		dstDir: dstDir,
	}, nil
}
