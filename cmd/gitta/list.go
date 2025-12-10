package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/services"
	"github.com/gavin/gitta/pkg/ui"
)

var (
	listAll bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Sprint tasks",
	Long:  "Display current Sprint tasks in a formatted table. Use --all to include backlog tasks.",
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
		listService := services.NewListService(storyRepo, gitRepo)

		if listAll {
			sprintStories, backlogStories, err := listService.ListAllTasks(ctx, repoPath)
			if err != nil {
				return fmt.Errorf("list --all: %w", err)
			}

			if len(sprintStories)+len(backlogStories) == 0 {
				fmt.Println("No tasks found.")
				return nil
			}

			sections := map[string][]ui.DisplayStory{
				"Sprint":  toDisplayStories(sprintStories),
				"Backlog": toDisplayStories(backlogStories),
			}

			output := ui.RenderStorySections(sections)
			fmt.Println(output)
			return nil
		}

		stories, err := listService.ListSprintTasks(ctx, repoPath)
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}

		if len(stories) == 0 {
			fmt.Println("No Sprint tasks found.")
			return nil
		}

		sections := map[string][]ui.DisplayStory{
			"Sprint": toDisplayStories(stories),
		}

		output := ui.RenderStorySections(sections)
		fmt.Println(output)
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "Include backlog tasks")
}

func toDisplayStories(stories []*services.StoryWithStatus) []ui.DisplayStory {
	display := make([]ui.DisplayStory, 0, len(stories))
	for _, s := range stories {
		display = append(display, ui.DisplayStory{
			Source:   s.Source,
			Story:    s.Story,
			Priority: s.Story.Priority,
			Status:   s.Status,
		})
	}
	return display
}
