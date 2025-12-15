package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/core/workspace"
	"github.com/gavin/gitta/internal/services"
)

var (
	createTitle    string
	createPrefix   string
	createTemplate string
	createEditor   string
	createStatus   string
	createPriority string
	createAssignee string
	createTags     []string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new story",
	Long:  "Create a new story file from a template, generate unique ID, and open editor.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		// Validate required flags
		if createTitle == "" {
			return fmt.Errorf("--title is required")
		}

		// Get repository path
		repoPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine working directory: %w", err)
		}

		// Determine story directory using workspace structure (default consolidated).
		structure, err := workspace.DetectStructure(ctx, repoPath)
		if err != nil {
			return fmt.Errorf("failed to detect workspace structure: %w", err)
		}
		storyDir := workspace.ResolveBacklogPath(repoPath, structure)
		if _, err := os.Stat(storyDir); os.IsNotExist(err) {
			// Fall back to repo root if backlog missing (keeps backward compatibility for unusual layouts).
			storyDir = repoPath
		}

		// Create service dependencies
		idGenerator := filesystem.NewIDCounter(repoPath)
		parser := filesystem.NewMarkdownParser()
		storyRepo := filesystem.NewRepository(parser)
		createService := services.NewCreateService(idGenerator, parser, storyRepo, storyDir)

		// Parse status and priority
		var status core.Status
		if createStatus != "" {
			status = core.Status(createStatus)
		} else {
			status = core.StatusTodo
		}

		var priority core.Priority
		if createPriority != "" {
			priority = core.Priority(createPriority)
		} else {
			priority = core.PriorityMedium
		}

		// Build request
		req := services.CreateStoryRequest{
			Title:    createTitle,
			Prefix:   createPrefix,
			Template: createTemplate,
			Editor:   createEditor,
			Status:   status,
			Priority: priority,
			Tags:     createTags,
		}
		if createAssignee != "" {
			req.Assignee = &createAssignee
		}

		// Create story
		story, filePath, err := createService.CreateStory(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create story: %w", err)
		}

		// Output result
		if jsonOutput {
			output := map[string]interface{}{
				"id":    story.ID,
				"file":  filePath,
				"title": story.Title,
			}
			if createEditor != "" || os.Getenv("EDITOR") != "" {
				editor := createEditor
				if editor == "" {
					editor = os.Getenv("EDITOR")
				}
				output["editor"] = editor
			}
			jsonBytes, err := json.Marshal(output)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))
		} else {
			fmt.Printf("Created story %s: %s\n", story.ID, story.Title)
			fmt.Printf("File: %s\n", filePath)
			if createEditor != "" || os.Getenv("EDITOR") != "" {
				editor := createEditor
				if editor == "" {
					editor = os.Getenv("EDITOR")
				}
				fmt.Printf("Opened in editor: %s\n", editor)
			}
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createTitle, "title", "", "Story title/name (required)")
	createCmd.Flags().StringVar(&createPrefix, "prefix", "US", "ID prefix (e.g., US, BUG, TS)")
	createCmd.Flags().StringVar(&createTemplate, "template", "", "Path to custom template file (default: built-in template)")
	createCmd.Flags().StringVar(&createEditor, "editor", "", "Editor command to use (default: $EDITOR env var, fallback: vi)")
	createCmd.Flags().StringVar(&createStatus, "status", "todo", "Initial status (todo, doing, review, done)")
	createCmd.Flags().StringVar(&createPriority, "priority", "medium", "Initial priority (low, medium, high, critical)")
	createCmd.Flags().StringVar(&createAssignee, "assignee", "", "Initial assignee")
	createCmd.Flags().StringArrayVar(&createTags, "tag", []string{}, "Initial tags (can be specified multiple times)")

	// Mark title as required
	createCmd.MarkFlagRequired("title")
}
