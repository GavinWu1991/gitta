package unit

import (
	"testing"

	"github.com/gavin/gitta/internal/core"
)

func TestIdentifyUnfinishedTasks(t *testing.T) {
	tests := []struct {
		name           string
		stories        []*core.Story
		wantUnfinished []string // Story IDs
	}{
		{
			name: "all tasks done",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusDone},
				{ID: "US-002", Status: core.StatusDone},
			},
			wantUnfinished: []string{},
		},
		{
			name: "all tasks unfinished",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo},
				{ID: "US-002", Status: core.StatusDoing},
				{ID: "US-003", Status: core.StatusReview},
			},
			wantUnfinished: []string{"US-001", "US-002", "US-003"},
		},
		{
			name: "mixed finished and unfinished",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusDone},
				{ID: "US-002", Status: core.StatusTodo},
				{ID: "US-003", Status: core.StatusDoing},
				{ID: "US-004", Status: core.StatusDone},
				{ID: "US-005", Status: core.StatusReview},
			},
			wantUnfinished: []string{"US-002", "US-003", "US-005"},
		},
		{
			name:           "empty stories list",
			stories:        []*core.Story{},
			wantUnfinished: []string{},
		},
		{
			name: "status not set (defaults to todo)",
			stories: []*core.Story{
				{ID: "US-001", Status: ""}, // Empty status should be treated as unfinished
				{ID: "US-002", Status: core.StatusDone},
			},
			wantUnfinished: []string{"US-001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unfinished := identifyUnfinishedTasks(tt.stories)

			if len(unfinished) != len(tt.wantUnfinished) {
				t.Errorf("identifyUnfinishedTasks() returned %d tasks, want %d", len(unfinished), len(tt.wantUnfinished))
				return
			}

			// Create maps for comparison
			gotMap := make(map[string]bool)
			for _, story := range unfinished {
				gotMap[story.ID] = true
			}
			wantMap := make(map[string]bool)
			for _, id := range tt.wantUnfinished {
				wantMap[id] = true
			}

			for id := range gotMap {
				if !wantMap[id] {
					t.Errorf("unexpected unfinished task: %s", id)
				}
			}
			for id := range wantMap {
				if !gotMap[id] {
					t.Errorf("missing unfinished task: %s", id)
				}
			}
		})
	}
}

// identifyUnfinishedTasks is a helper function that identifies unfinished tasks.
// This will be implemented in the service.
func identifyUnfinishedTasks(stories []*core.Story) []*core.Story {
	var unfinished []*core.Story
	for _, story := range stories {
		// Unfinished = status is not "done"
		// Also handle empty status (defaults to todo, which is unfinished)
		if story.Status == "" || story.Status != core.StatusDone {
			unfinished = append(unfinished, story)
		}
	}
	return unfinished
}
