package main

import (
	"os"

	"github.com/deparrow/dpc/app"
)

func main() {
	rootCmd := NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
