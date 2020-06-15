package main

import (
	"github.com/samphillips/backup/internal/config"
	"github.com/samphillips/backup/internal/logging"
)

func main() {
	config := config.ParseConfig()

	if config.Verbose {
		logging.SetLogLevel(logging.DEBUG)
	}
}
