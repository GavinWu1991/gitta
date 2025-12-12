package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/services"
)

var (
	moveTo    string
	moveForce bool
)

var moveCmd = &cobra.Command{
	Use:   "move <story-id>",
	Short: "Move story file",
	Long:  "Move a story file to a different directory atomically while preserving all content.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		storyID := args[0]

		// Validate target directory
		if moveTo == "" {
			return fmt.Errorf("--to is required")
		}

		// Get repository path
		repoPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine working directory: %w", err)
		}

		// Create service dependencies
		parser := filesystem.NewMarkdownParser()
		storyRepo := filesystem.NewRepository(parser)
		moveService := services.NewMoveService(parser, storyRepo, repoPath)

		// Find story to get source path
		story, sourcePath, err := storyRepo.FindStoryByID(ctx, repoPath, storyID)
		if err != nil {
			return fmt.Errorf("story not found: %s", storyID)
		}

		// Move story
		err = moveService.MoveStory(ctx, storyID, moveTo, moveForce)
		if err != nil {
			return fmt.Errorf("failed to move story: %w", err)
		}

		// Build target path for output
		targetPath := filepath.Join(repoPath, moveTo, filepath.Base(sourcePath))

		// Output result
		if jsonOutput {
			output := map[string]interface{}{
				"id":   story.ID,
				"from": sourcePath,
				"to":   targetPath,
			}
			jsonBytes, err := json.Marshal(output)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))
		} else {
			fmt.Printf("Moved %s: %s -> %s\n", storyID, sourcePath, targetPath)
		}

		return nil
	},
}

func init() {
	moveCmd.Flags().StringVar(&moveTo, "to", "", "Target directory path (required)")
	moveCmd.Flags().BoolVar(&moveForce, "force", false, "Overwrite existing file at destination")
	moveCmd.MarkFlagRequired("to")
}
