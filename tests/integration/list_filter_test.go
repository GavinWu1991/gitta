package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

// setupRepo and writeStory are defined in list_command_test.go (same package)

func TestListCommand_WithFilters(t *testing.T) {
	repoPath := setupRepo(t)

	// Create stories with different attributes
	writeStoryWithAttrs(t, filepath.Join(repoPath, "backlog", "US-001.md"), "US-001", "Todo High Story", "todo", "high", "alice", []string{"frontend"})
	writeStoryWithAttrs(t, filepath.Join(repoPath, "backlog", "US-002.md"), "US-002", "Doing High Story", "doing", "high", "bob", []string{"backend"})
	writeStoryWithAttrs(t, filepath.Join(repoPath, "backlog", "US-003.md"), "US-003", "Todo Low Story", "todo", "low", "alice", []string{"frontend"})
	writeStoryWithAttrs(t, filepath.Join(repoPath, "backlog", "US-004.md"), "US-004", "Doing Medium Story", "doing", "medium", "charlie", []string{"ui"})

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := git.NewRepository()
	svc := services.NewListService(storyRepo, gitRepo)

	ctx := context.Background()

	tests := []struct {
		name    string
		filter  services.Filter
		wantIDs []string
		wantErr bool
	}{
		{
			name: "filter by status todo",
			filter: services.Filter{
				Statuses: []core.Status{core.StatusTodo},
			},
			wantIDs: []string{"US-001", "US-003"},
		},
		{
			name: "filter by priority high",
			filter: services.Filter{
				Priorities: []core.Priority{core.PriorityHigh},
			},
			wantIDs: []string{"US-001", "US-002"},
		},
		{
			name: "filter by assignee alice",
			filter: services.Filter{
				Assignees: []string{"alice"},
			},
			wantIDs: []string{"US-001", "US-003"},
		},
		{
			name: "filter by tag frontend",
			filter: services.Filter{
				Tags: []string{"frontend"},
			},
			wantIDs: []string{"US-001", "US-003"},
		},
		{
			name: "filter by status AND priority (AND logic)",
			filter: services.Filter{
				Statuses:   []core.Status{core.StatusTodo},
				Priorities: []core.Priority{core.PriorityHigh},
			},
			wantIDs: []string{"US-001"},
		},
		{
			name: "filter by status OR priority (multiple values)",
			filter: services.Filter{
				Statuses:   []core.Status{core.StatusTodo, core.StatusDoing},
				Priorities: []core.Priority{core.PriorityHigh},
			},
			wantIDs: []string{"US-001", "US-002"},
		},
		{
			name: "filter by multiple tags (OR logic)",
			filter: services.Filter{
				Tags: []string{"frontend", "backend"},
			},
			wantIDs: []string{"US-001", "US-002", "US-003"},
		},
		{
			name: "complex filter: status AND assignee AND tag",
			filter: services.Filter{
				Statuses:  []core.Status{core.StatusTodo},
				Assignees: []string{"alice"},
				Tags:      []string{"frontend"},
			},
			wantIDs: []string{"US-001", "US-003"},
		},
		{
			name:    "empty filter returns all",
			filter:  services.Filter{},
			wantIDs: []string{"US-001", "US-002", "US-003", "US-004"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stories, err := svc.ListStories(ctx, repoPath, tt.filter)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ListStories() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ListStories() error = %v, want nil", err)
			}

			gotIDs := make([]string, 0, len(stories))
			for _, s := range stories {
				gotIDs = append(gotIDs, s.Story.ID)
			}

			if len(gotIDs) != len(tt.wantIDs) {
				t.Errorf("ListStories() count = %d, want %d. Got: %v, Want: %v", len(gotIDs), len(tt.wantIDs), gotIDs, tt.wantIDs)
				return
			}

			// Check that all expected IDs are present
			wantMap := make(map[string]bool)
			for _, id := range tt.wantIDs {
				wantMap[id] = true
			}

			for _, id := range gotIDs {
				if !wantMap[id] {
					t.Errorf("Unexpected ID in result: %s", id)
				}
				delete(wantMap, id)
			}

			// Check that all expected IDs were found
			for id := range wantMap {
				t.Errorf("Missing expected ID: %s", id)
			}
		})
	}
}

func writeStoryWithAttrs(t *testing.T, path, id, title, status, priority, assignee string, tags []string) {
	t.Helper()
	content := `---
id: ` + id + `
title: ` + title + `
status: ` + status + `
priority: ` + priority + `
assignee: ` + assignee + `
tags:`
	for _, tag := range tags {
		content += "\n  - " + tag
	}
	content += `
---
Body
`
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write story: %v", err)
	}
}
