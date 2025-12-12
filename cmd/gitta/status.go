package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

var statusStatus string

var statusCmd = &cobra.Command{
	Use:   "status <story-id>",
	Short: "Update story status",
	Long:  "Update a story's status atomically with data corruption prevention.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		storyID := args[0]

		// Validate status
		if statusStatus == "" {
			return fmt.Errorf("--status is required")
		}

		// Get repository path
		repoPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine working directory: %w", err)
		}

		// Create service dependencies
		parser := filesystem.NewMarkdownParser()
		storyRepo := filesystem.NewRepository(parser)
		updateService := services.NewUpdateService(parser, storyRepo, repoPath)

		// Parse status
		newStatus := core.Status(statusStatus)

		// Update status
		err = updateService.UpdateStatus(ctx, storyID, newStatus)
		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		// Read updated story for output
		story, _, readErr := storyRepo.FindStoryByID(ctx, repoPath, storyID)
		if readErr != nil {
			// Story was updated but we can't read it back - still report success
			if jsonOutput {
				fmt.Printf(`{"id":"%s","status":"%s"}\n`, storyID, statusStatus)
			} else {
				fmt.Printf("Updated %s: status -> %s\n", storyID, statusStatus)
			}
			return nil
		}

		// Output result
		if jsonOutput {
			output := map[string]interface{}{
				"id":         story.ID,
				"status":     string(story.Status),
				"updated_at": nil,
			}
			if story.UpdatedAt != nil {
				output["updated_at"] = story.UpdatedAt.Format("2006-01-02T15:04:05Z")
			}
			jsonBytes, err := json.Marshal(output)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))
		} else {
			fmt.Printf("Updated %s: status -> %s\n", storyID, statusStatus)
		}

		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusStatus, "status", "", "New status value (todo, doing, review, done) (required)")
	statusCmd.MarkFlagRequired("status")
}
