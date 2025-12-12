package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"

	"github.com/spf13/cobra"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
	"github.com/gavin/gitta/pkg/ui"
)

var (
	listAll      bool
	listStatus   []string
	listPriority []string
	listAssignee []string
	listTag      []string
	listSort     string
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

		// Build filter from flags
		filter, err := buildFilter(listStatus, listPriority, listAssignee, listTag)
		if err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}

		// If filters are specified, use filtered listing
		if hasFilters(filter) {
			stories, err := listService.ListStories(ctx, repoPath, filter)
			if err != nil {
				return fmt.Errorf("list: %w", err)
			}

			if len(stories) == 0 {
				if jsonOutput {
					fmt.Println(`{"stories":[],"total":0,"filtered":true}`)
				} else {
					fmt.Println("No tasks found matching filters.")
				}
				return nil
			}

			// Sort if specified
			if listSort != "" {
				stories = sortStoriesByField(stories, listSort)
			}

			// Output in JSON or formatted table
			if jsonOutput {
				outputJSON(stories, true)
				return nil
			}

			// Group by source for display
			sections := groupBySource(stories)
			output := ui.RenderStorySections(sections)
			fmt.Println(output)
			return nil
		}

		// No filters - use existing behavior
		if listAll {
			sprintStories, backlogStories, err := listService.ListAllTasks(ctx, repoPath)
			if err != nil {
				return fmt.Errorf("list --all: %w", err)
			}

			if len(sprintStories)+len(backlogStories) == 0 {
				if jsonOutput {
					fmt.Println(`{"stories":[],"total":0,"filtered":false}`)
				} else {
					fmt.Println("No tasks found.")
				}
				return nil
			}

			allStories := append(sprintStories, backlogStories...)
			if jsonOutput {
				outputJSON(allStories, false)
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
			if jsonOutput {
				fmt.Println(`{"stories":[],"total":0,"filtered":false}`)
			} else {
				fmt.Println("No Sprint tasks found.")
			}
			return nil
		}

		if jsonOutput {
			outputJSON(stories, false)
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
	listCmd.Flags().StringArrayVar(&listStatus, "status", []string{}, "Filter by status (can specify multiple: --status todo --status doing)")
	listCmd.Flags().StringArrayVar(&listPriority, "priority", []string{}, "Filter by priority")
	listCmd.Flags().StringArrayVar(&listAssignee, "assignee", []string{}, "Filter by assignee")
	listCmd.Flags().StringArrayVar(&listTag, "tag", []string{}, "Filter by tags (story must have any tag)")
	listCmd.Flags().StringVar(&listSort, "sort", "id", "Sort field (id, title, status, priority, created_at)")
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

// buildFilter constructs a Filter from command-line flags with validation.
func buildFilter(statuses, priorities, assignees, tags []string) (services.Filter, error) {
	filter := services.Filter{}

	// Validate and parse statuses
	validStatuses := map[string]bool{
		"todo":   true,
		"doing":  true,
		"review": true,
		"done":   true,
	}
	for _, s := range statuses {
		if !validStatuses[s] {
			return filter, fmt.Errorf("invalid status: %s (valid: todo, doing, review, done)", s)
		}
		filter.Statuses = append(filter.Statuses, core.Status(s))
	}

	// Validate and parse priorities
	validPriorities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}
	for _, p := range priorities {
		if !validPriorities[p] {
			return filter, fmt.Errorf("invalid priority: %s (valid: low, medium, high, critical)", p)
		}
		filter.Priorities = append(filter.Priorities, core.Priority(p))
	}

	// Validate assignees (username pattern: alphanumeric, hyphens, underscores)
	assigneePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	for _, a := range assignees {
		if !assigneePattern.MatchString(a) {
			return filter, fmt.Errorf("invalid assignee: %s (must be alphanumeric with hyphens/underscores)", a)
		}
	}
	filter.Assignees = assignees

	// Validate tags (same pattern as assignees)
	tagPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	for _, t := range tags {
		if !tagPattern.MatchString(t) {
			return filter, fmt.Errorf("invalid tag: %s (must be alphanumeric with hyphens/underscores)", t)
		}
	}
	filter.Tags = tags

	return filter, nil
}

// hasFilters checks if any filter criteria are specified.
func hasFilters(filter services.Filter) bool {
	return len(filter.Statuses) > 0 ||
		len(filter.Priorities) > 0 ||
		len(filter.Assignees) > 0 ||
		len(filter.Tags) > 0
}

// groupBySource groups stories by their source (Sprint/Backlog).
func groupBySource(stories []*services.StoryWithStatus) map[string][]ui.DisplayStory {
	sections := make(map[string][]ui.DisplayStory)
	for _, s := range stories {
		source := s.Source
		if source == "" {
			source = "Sprint" // Default
		}
		sections[source] = append(sections[source], ui.DisplayStory{
			Source:   s.Source,
			Story:    s.Story,
			Priority: s.Story.Priority,
			Status:   s.Status,
		})
	}
	return sections
}

// sortStoriesByField sorts stories by the specified field.
func sortStoriesByField(stories []*services.StoryWithStatus, field string) []*services.StoryWithStatus {
	// Create a copy to avoid modifying original
	result := make([]*services.StoryWithStatus, len(stories))
	copy(result, stories)

	switch field {
	case "title":
		sort.Slice(result, func(i, j int) bool {
			return result[i].Story.Title < result[j].Story.Title
		})
	case "status":
		sort.Slice(result, func(i, j int) bool {
			return string(result[i].Status) < string(result[j].Status)
		})
	case "priority":
		sort.Slice(result, func(i, j int) bool {
			return string(result[i].Story.Priority) < string(result[j].Story.Priority)
		})
	case "created_at":
		sort.Slice(result, func(i, j int) bool {
			if result[i].Story.CreatedAt == nil {
				return false
			}
			if result[j].Story.CreatedAt == nil {
				return true
			}
			return result[i].Story.CreatedAt.Before(*result[j].Story.CreatedAt)
		})
	default: // "id" or unknown
		sort.Slice(result, func(i, j int) bool {
			return result[i].Story.ID < result[j].Story.ID
		})
	}
	return result
}

// outputJSON outputs stories in JSON format.
func outputJSON(stories []*services.StoryWithStatus, filtered bool) {
	type storyJSON struct {
		ID        string   `json:"id"`
		Title     string   `json:"title"`
		Status    string   `json:"status"`
		Priority  string   `json:"priority"`
		Assignee  *string  `json:"assignee"`
		Tags      []string `json:"tags"`
		CreatedAt *string  `json:"created_at,omitempty"`
		UpdatedAt *string  `json:"updated_at,omitempty"`
	}

	storyList := make([]storyJSON, 0, len(stories))
	for _, s := range stories {
		sj := storyJSON{
			ID:       s.Story.ID,
			Title:    s.Story.Title,
			Status:   string(s.Status),
			Priority: string(s.Story.Priority),
			Tags:     s.Story.Tags,
		}
		if s.Story.Assignee != nil {
			sj.Assignee = s.Story.Assignee
		}
		if s.Story.CreatedAt != nil {
			createdStr := s.Story.CreatedAt.Format("2006-01-02T15:04:05Z")
			sj.CreatedAt = &createdStr
		}
		if s.Story.UpdatedAt != nil {
			updatedStr := s.Story.UpdatedAt.Format("2006-01-02T15:04:05Z")
			sj.UpdatedAt = &updatedStr
		}
		storyList = append(storyList, sj)
	}

	output := map[string]interface{}{
		"stories":  storyList,
		"total":    len(storyList),
		"filtered": filtered,
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}
