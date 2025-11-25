package unit

import (
	"strings"
	"testing"
	"time"

	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestValidateStory_ValidStory(t *testing.T) {
	assignee := "alice"
	createdAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 20, 14, 30, 0, 0, time.UTC)

	story := &core.Story{
		ID:        "US-001",
		Title:     "Valid Story",
		Assignee:  &assignee,
		Priority:  core.PriorityHigh,
		Status:    core.StatusDoing,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		Tags:      []string{"backend", "api"},
		Body:      "Test body",
	}

	errors := services.ValidateStory(story)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateStory_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name  string
		story *core.Story
		want  int // expected number of errors
	}{
		{
			name: "missing ID",
			story: &core.Story{
				Title:    "Test",
				Priority: core.PriorityMedium, // Set valid defaults
				Status:   core.StatusTodo,
			},
			want: 1,
		},
		{
			name: "missing Title",
			story: &core.Story{
				ID:       "US-001",
				Priority: core.PriorityMedium, // Set valid defaults
				Status:   core.StatusTodo,
			},
			want: 1,
		},
		{
			name:  "nil story",
			story: nil,
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := services.ValidateStory(tt.story)
			if len(errors) != tt.want {
				t.Errorf("Expected %d errors, got %d: %v", tt.want, len(errors), errors)
			}
		})
	}
}

func TestValidateStory_InvalidIDFormat(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid ID", "US-001", true},
		{"valid ID with more digits", "BG-12345", true},
		{"invalid: lowercase", "us-001", false},
		{"invalid: no dash", "US001", false},
		{"invalid: wrong format", "001-US", false},
		{"invalid: too short", "U-1", false},
		{"invalid: too long", "US-12345678901234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			story := &core.Story{
				ID:       tt.id,
				Title:    "Test",
				Priority: core.PriorityMedium, // Set valid defaults
				Status:   core.StatusTodo,
			}
			errors := services.ValidateStory(story)
			hasIDError := false
			for _, err := range errors {
				if err.Field == "id" {
					hasIDError = true
					break
				}
			}
			if tt.valid && hasIDError {
				t.Errorf("Expected valid ID '%s', but got error: %v", tt.id, errors)
			}
			if !tt.valid && !hasIDError {
				t.Errorf("Expected invalid ID '%s' to produce error, but got none", tt.id)
			}
		})
	}
}

func TestValidateStory_InvalidEnumValues(t *testing.T) {
	tests := []struct {
		name  string
		story *core.Story
		field string
	}{
		{
			name: "invalid Priority",
			story: &core.Story{
				ID:       "US-001",
				Title:    "Test",
				Priority: core.Priority("invalid"),
			},
			field: "priority",
		},
		{
			name: "invalid Status",
			story: &core.Story{
				ID:     "US-001",
				Title:  "Test",
				Status: core.Status("invalid"),
			},
			field: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := services.ValidateStory(tt.story)
			found := false
			for _, err := range errors {
				if err.Field == tt.field {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected validation error for field '%s', but got none", tt.field)
			}
		})
	}
}

func TestValidateStory_DateValidation(t *testing.T) {
	createdAt := time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC) // Before createdAt

	story := &core.Story{
		ID:        "US-001",
		Title:     "Test",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	errors := services.ValidateStory(story)
	found := false
	for _, err := range errors {
		if err.Field == "created_at" && err.Rule == "range" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected validation error for date range (created_at > updated_at), but got none")
	}
}

func TestValidateStory_TagsValidation(t *testing.T) {
	tests := []struct {
		name  string
		tags  []string
		valid bool
	}{
		{"valid tags", []string{"backend", "api"}, true},
		{"too many tags", make([]string, 21), false},
		{"empty tag", []string{""}, false},
		{"duplicate tags", []string{"backend", "backend"}, false},
		{"tag too long", []string{strings.Repeat("a", 31)}, false},
		{"invalid characters", []string{"tag with spaces"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			story := &core.Story{
				ID:    "US-001",
				Title: "Test",
				Tags:  tt.tags,
			}
			errors := services.ValidateStory(story)
			hasTagError := false
			for _, err := range errors {
				if err.Field == "tags" {
					hasTagError = true
					break
				}
			}
			if tt.valid && hasTagError {
				t.Errorf("Expected valid tags, but got error: %v", errors)
			}
			if !tt.valid && !hasTagError {
				t.Errorf("Expected invalid tags to produce error, but got none")
			}
		})
	}
}
