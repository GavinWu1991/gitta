package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitta",
	Short: "Git Task Assistant - Manage backlog stories via Git-backed workflows",
	Long: `Gitta is a lightweight task management tool that stores tasks as Markdown files
in your Git repository. It uses branch state to track task progress automatically.

For more information, see: https://github.com/gavin/gitta/docs/cli/`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no subcommand provided
		cmd.Help()
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON output")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug|info|warn|error)")

	// Register subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
