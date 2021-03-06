package config

import (
	"path/filepath"
	"strings"

	"github.com/jpillora/opts"
	"github.com/samphillips/backup/internal/logging"
)

// Config contains the validated flags
type Config struct {
	SrcDir          string `opts:"mode=arg,help=(Required) The absolute directory path you wish to back up"`
	DstDir          string `opts:"mode=arg,help=(Required) The absolute directory that the source directory will be backed up to"`
	Fast            bool   `opts:"help=Assume files of the same size are equal and don't do a hashsum check to test contents equality"`
	Mirror          bool   `opts:"help=Ensure backup location is a mirror of the source location (This will remove any files in the destination that do not exist at the source)"`
	IncludeSymlinks bool   `opts:"help=Also backup any symlinks found (If the symlink target is also in the source directory the backup symlink will target the backed-up file)"`
	Verbose         bool   `opts:"help=Enable debug logging"`
}

// ParseConfig parses the command line flags and validates them
func ParseConfig() Config {
	c := Config{}
	opts.Parse(&c)

	var err error

	c.SrcDir, err = filepath.Abs(c.SrcDir)

	if err != nil {
		logging.Fatal("Could not resolve absolute path for source directory: %s", err)
	}

	c.DstDir, err = filepath.Abs(c.DstDir)

	if err != nil {
		logging.Fatal("Could not resolve absolute path for source directory: %s", err)
	}

	if !strings.HasSuffix(c.SrcDir, "/") {
		c.SrcDir += "/"
	}

	if !strings.HasSuffix(c.DstDir, "/") {
		c.DstDir += "/"
	}

	return c
}
