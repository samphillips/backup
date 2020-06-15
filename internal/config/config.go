package config

import (
	"strings"

	"github.com/jpillora/opts"
)

// Config contains the validated flags
type Config struct {
	SrcDir  string `opts:"mode=arg,help=(Required) The absolute directory path you wish to back up"`
	DstDir  string `opts:"mode=arg,help=(Required) The absolute directory that the source directory will be backed up to"`
	Verbose bool   `opts:"help=Enable debug logging"`
}

// ParseConfig parses the command line flags and validates them
func ParseConfig() Config {
	c := Config{}
	opts.Parse(&c)

	if !strings.HasSuffix(c.SrcDir, "/") {
		c.SrcDir += "/"
	}

	if !strings.HasSuffix(c.DstDir, "/") {
		c.DstDir += "/"
	}

	return c
}
