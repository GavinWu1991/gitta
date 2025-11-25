package main

import (
	"fmt"
	"os"
)

var (
	jsonOutput bool
	logLevel   string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
