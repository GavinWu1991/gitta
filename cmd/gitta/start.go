package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/services"
)

var (
	startAssignee string
)

var startCmd = &cobra.Command{
	Use:   "start <task-id|file-path>",
	Short: "Start work on a task by creating/checking out its feature branch",
	Long:  "Create and checkout the feature branch for the given task (ID or file path) and optionally update the assignee field.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		repoPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine working directory: %w", err)
		}

		storyRepo := filesystem.NewDefaultRepository()
		gitRepo := git.NewRepository()
		parser := filesystem.NewMarkdownParser()
		startService := services.NewStartService(storyRepo, gitRepo, parser)

		story, branchName, startErr := startService.Start(ctx, repoPath, args[0], valuePtr(startAssignee))
		var assigneeUpdateErr *services.AssigneeUpdateError
		if startErr != nil {
			if errors.As(startErr, &assigneeUpdateErr) {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", assigneeUpdateErr)
			} else {
				return fmt.Errorf("start: %w", startErr)
			}
		}

		fmt.Printf("Started work on %s: switched to branch %s\n", story.ID, branchName)
		if startAssignee != "" && assigneeUpdateErr == nil {
			fmt.Printf("Updated assignee to %s\n", startAssignee)
		}
		return nil
	},
}

func init() {
	startCmd.Flags().StringVar(&startAssignee, "assignee", "", "Explicit assignee to set in the task file")
}

func valuePtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
