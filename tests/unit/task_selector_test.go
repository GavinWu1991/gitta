package unit

import (
	"testing"

	"github.com/gavin/gitta/internal/core"
)

func TestTaskSelector(t *testing.T) {
	// Basic test structure for TUI selector
	// Full TUI testing would require terminal simulation
	tests := []struct {
		name          string
		tasks         []*core.Story
		selectedIDs   []string
		wantCancelled bool
	}{
		{
			name: "select single task",
			tasks: []*core.Story{
				{ID: "US-001", Title: "Task 1", Status: core.StatusTodo},
				{ID: "US-002", Title: "Task 2", Status: core.StatusDoing},
			},
			selectedIDs:   []string{"US-001"},
			wantCancelled: false,
		},
		{
			name: "select multiple tasks",
			tasks: []*core.Story{
				{ID: "US-001", Title: "Task 1", Status: core.StatusTodo},
				{ID: "US-002", Title: "Task 2", Status: core.StatusDoing},
				{ID: "US-003", Title: "Task 3", Status: core.StatusReview},
			},
			selectedIDs:   []string{"US-001", "US-003"},
			wantCancelled: false,
		},
		{
			name: "cancel selection",
			tasks: []*core.Story{
				{ID: "US-001", Title: "Task 1", Status: core.StatusTodo},
			},
			selectedIDs:   []string{},
			wantCancelled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a placeholder test - actual TUI testing would require
			// terminal simulation or mocking Bubble Tea
			// For now, we'll test the logic that processes selections
			selected := filterTasksByIDs(tt.tasks, tt.selectedIDs)

			if tt.wantCancelled {
				if len(selected) != 0 {
					t.Errorf("expected cancelled selection, got %d tasks", len(selected))
				}
			} else {
				if len(selected) != len(tt.selectedIDs) {
					t.Errorf("selected %d tasks, want %d", len(selected), len(tt.selectedIDs))
				}
			}
		})
	}
}

// filterTasksByIDs is a helper for testing task selection logic
func filterTasksByIDs(tasks []*core.Story, ids []string) []*core.Story {
	if len(ids) == 0 {
		return nil // Cancelled
	}

	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	var selected []*core.Story
	for _, task := range tasks {
		if idMap[task.ID] {
			selected = append(selected, task)
		}
	}
	return selected
}
