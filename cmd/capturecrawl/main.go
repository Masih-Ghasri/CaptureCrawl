package main

import (
	"os"

	"github.com/Masih-Ghasri/CaptureCrawl/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
