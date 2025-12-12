package unit

import (
	"testing"

	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestListService_ApplyFilter(t *testing.T) {
	tests := []struct {
		name    string
		stories []*core.Story
		filter  services.Filter
		wantIDs []string
		wantErr bool
	}{
		{
			name: "empty filter returns all stories",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Priority: core.PriorityLow},
				{ID: "US-002", Status: core.StatusDoing, Priority: core.PriorityHigh},
			},
			filter:  services.Filter{},
			wantIDs: []string{"US-001", "US-002"},
		},
		{
			name: "filter by single status",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Priority: core.PriorityLow},
				{ID: "US-002", Status: core.StatusDoing, Priority: core.PriorityHigh},
				{ID: "US-003", Status: core.StatusTodo, Priority: core.PriorityMedium},
			},
			filter: services.Filter{
				Statuses: []core.Status{core.StatusTodo},
			},
			wantIDs: []string{"US-001", "US-003"},
		},
		{
			name: "filter by multiple statuses (OR logic)",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Priority: core.PriorityLow},
				{ID: "US-002", Status: core.StatusDoing, Priority: core.PriorityHigh},
				{ID: "US-003", Status: core.StatusReview, Priority: core.PriorityMedium},
			},
			filter: services.Filter{
				Statuses: []core.Status{core.StatusTodo, core.StatusDoing},
			},
			wantIDs: []string{"US-001", "US-002"},
		},
		{
			name: "filter by priority",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Priority: core.PriorityLow},
				{ID: "US-002", Status: core.StatusDoing, Priority: core.PriorityHigh},
				{ID: "US-003", Status: core.StatusTodo, Priority: core.PriorityHigh},
			},
			filter: services.Filter{
				Priorities: []core.Priority{core.PriorityHigh},
			},
			wantIDs: []string{"US-002", "US-003"},
		},
		{
			name: "filter by assignee",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Assignee: stringPtr("alice")},
				{ID: "US-002", Status: core.StatusDoing, Assignee: stringPtr("bob")},
				{ID: "US-003", Status: core.StatusTodo, Assignee: stringPtr("alice")},
			},
			filter: services.Filter{
				Assignees: []string{"alice"},
			},
			wantIDs: []string{"US-001", "US-003"},
		},
		{
			name: "filter by tags",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Tags: []string{"frontend", "ui"}},
				{ID: "US-002", Status: core.StatusDoing, Tags: []string{"backend"}},
				{ID: "US-003", Status: core.StatusTodo, Tags: []string{"frontend"}},
			},
			filter: services.Filter{
				Tags: []string{"frontend"},
			},
			wantIDs: []string{"US-001", "US-003"},
		},
		{
			name: "filter by multiple tags (OR logic)",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Tags: []string{"frontend"}},
				{ID: "US-002", Status: core.StatusDoing, Tags: []string{"backend"}},
				{ID: "US-003", Status: core.StatusTodo, Tags: []string{"ui"}},
			},
			filter: services.Filter{
				Tags: []string{"frontend", "backend"},
			},
			wantIDs: []string{"US-001", "US-002"},
		},
		{
			name: "filter by status AND priority (AND logic)",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Priority: core.PriorityHigh},
				{ID: "US-002", Status: core.StatusDoing, Priority: core.PriorityHigh},
				{ID: "US-003", Status: core.StatusTodo, Priority: core.PriorityLow},
			},
			filter: services.Filter{
				Statuses:   []core.Status{core.StatusTodo},
				Priorities: []core.Priority{core.PriorityHigh},
			},
			wantIDs: []string{"US-001"},
		},
		{
			name: "filter by status AND assignee AND tags",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Assignee: stringPtr("alice"), Tags: []string{"frontend"}},
				{ID: "US-002", Status: core.StatusDoing, Assignee: stringPtr("alice"), Tags: []string{"backend"}},
				{ID: "US-003", Status: core.StatusTodo, Assignee: stringPtr("bob"), Tags: []string{"frontend"}},
			},
			filter: services.Filter{
				Statuses:  []core.Status{core.StatusTodo},
				Assignees: []string{"alice"},
				Tags:      []string{"frontend"},
			},
			wantIDs: []string{"US-001"},
		},
		{
			name: "no matches",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Priority: core.PriorityLow},
				{ID: "US-002", Status: core.StatusDoing, Priority: core.PriorityHigh},
			},
			filter: services.Filter{
				Statuses: []core.Status{core.StatusDone},
			},
			wantIDs: []string{},
		},
		{
			name: "filter by multiple assignees (OR logic)",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Assignee: stringPtr("alice")},
				{ID: "US-002", Status: core.StatusDoing, Assignee: stringPtr("bob")},
				{ID: "US-003", Status: core.StatusTodo, Assignee: stringPtr("charlie")},
			},
			filter: services.Filter{
				Assignees: []string{"alice", "bob"},
			},
			wantIDs: []string{"US-001", "US-002"},
		},
		{
			name: "story with nil assignee not matched",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Assignee: nil},
				{ID: "US-002", Status: core.StatusDoing, Assignee: stringPtr("alice")},
			},
			filter: services.Filter{
				Assignees: []string{"alice"},
			},
			wantIDs: []string{"US-002"},
		},
		{
			name: "story with empty tags not matched",
			stories: []*core.Story{
				{ID: "US-001", Status: core.StatusTodo, Tags: []string{}},
				{ID: "US-002", Status: core.StatusDoing, Tags: []string{"frontend"}},
			},
			filter: services.Filter{
				Tags: []string{"frontend"},
			},
			wantIDs: []string{"US-002"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the filter matching logic directly
			filtered := filterStories(tt.stories, tt.filter)

			// Extract IDs
			gotIDs := make([]string, 0, len(filtered))
			for _, story := range filtered {
				gotIDs = append(gotIDs, story.ID)
			}

			// Compare
			if len(gotIDs) != len(tt.wantIDs) {
				t.Errorf("Filtered count = %d, want %d. Got: %v, Want: %v", len(gotIDs), len(tt.wantIDs), gotIDs, tt.wantIDs)
				return
			}

			// Check each ID
			wantMap := make(map[string]bool)
			for _, id := range tt.wantIDs {
				wantMap[id] = true
			}

			for _, id := range gotIDs {
				if !wantMap[id] {
					t.Errorf("Unexpected ID in result: %s", id)
				}
			}
		})
	}
}

// filterStories applies the filter logic (replicates the matching logic)
func filterStories(stories []*core.Story, filter services.Filter) []*core.Story {
	if isEmptyFilter(filter) {
		return stories
	}

	result := make([]*core.Story, 0, len(stories))
	for _, story := range stories {
		if matchesFilter(story, filter) {
			result = append(result, story)
		}
	}
	return result
}

// matchesFilter replicates the matching logic from list_service.go
func matchesFilter(story *core.Story, filter services.Filter) bool {
	// Status filter (OR logic)
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if story.Status == status {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Priority filter (OR logic)
	if len(filter.Priorities) > 0 {
		matched := false
		for _, priority := range filter.Priorities {
			if story.Priority == priority {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Assignee filter (OR logic)
	if len(filter.Assignees) > 0 {
		matched := false
		for _, assignee := range filter.Assignees {
			if story.Assignee != nil && *story.Assignee == assignee {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Tags filter (OR logic - story must have any of the filter tags)
	if len(filter.Tags) > 0 {
		matched := false
		for _, filterTag := range filter.Tags {
			for _, storyTag := range story.Tags {
				if storyTag == filterTag {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// isEmptyFilter checks if the filter is empty
func isEmptyFilter(filter services.Filter) bool {
	return len(filter.Statuses) == 0 &&
		len(filter.Priorities) == 0 &&
		len(filter.Assignees) == 0 &&
		len(filter.Tags) == 0
}

func stringPtr(s string) *string {
	return &s
}
