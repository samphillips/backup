package main

import (
	"fmt"

	"github.com/samphillips/backup/internal/config"
)

func main() {
	config, err := config.ParseConfig()

	if err != nil {
		return
	}

	fmt.Println(config)
}
