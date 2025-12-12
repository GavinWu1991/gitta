package main

import (
	"github.com/spf13/cobra"
)

// storyCmd is the parent command for all story-related operations.
var storyCmd = &cobra.Command{
	Use:   "story",
	Short: "Manage stories (create, list, status, move)",
	Long:  "Commands for creating, listing, updating, and moving stories.",
}

func init() {
	// Register story subcommands
	storyCmd.AddCommand(createCmd)
	storyCmd.AddCommand(statusCmd)
	storyCmd.AddCommand(moveCmd)
	// Note: list is a separate top-level command, not under story
}
