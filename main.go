package main

import (
	"os"

	"github.com/biwakonbu/zeus/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
