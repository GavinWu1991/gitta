package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gavin/gitta/internal/core"
)

var (
	idPattern       = regexp.MustCompile(`^[A-Z]{2}-[0-9]+$`) // Story ID pattern: 2 uppercase letters, dash, digits
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)  // Username pattern: alphanumeric, hyphens, underscores
	tagPattern      = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)  // Tag pattern: alphanumeric, hyphens, underscores
)

// ValidateStory validates a Story struct against business rules and returns validation errors.
// It checks all required fields, format constraints, enum values, and business rules.
// Returns a slice of ValidationErrors describing any violations. An empty slice
// indicates the story is valid and ready for use.
//
// Validation rules:
//   - ID: Required, pattern ^[A-Z]{2}-[0-9]+$, length 3-20
//   - Title: Required, non-empty after trim, max 200 characters
//   - Assignee: Optional, if set: length 1-50, valid username format
//   - Priority: Optional, if set: must be valid enum (low, medium, high, critical)
//   - Status: Optional, if set: must be valid enum (todo, doing, review, done)
//   - Dates: CreatedAt <= UpdatedAt if both set
//   - Tags: 0-20 tags, each 1-30 characters, valid format, no duplicates
func ValidateStory(story *core.Story) []core.ValidationError {
	var errors []core.ValidationError

	if story == nil {
		return []core.ValidationError{
			{
				Field:   "story",
				Rule:    "required",
				Message: "story cannot be nil",
			},
		}
	}

	// Validate ID
	if story.ID == "" {
		errors = append(errors, core.ValidationError{
			Field:   "id",
			Rule:    "required",
			Message: "story ID is required",
		})
	} else {
		if len(story.ID) < 3 || len(story.ID) > 20 {
			errors = append(errors, core.ValidationError{
				Field:   "id",
				Rule:    "length",
				Message: "story ID must be between 3 and 20 characters",
			})
		}
		if !idPattern.MatchString(story.ID) {
			errors = append(errors, core.ValidationError{
				Field:   "id",
				Rule:    "format",
				Message: "story ID must match pattern ^[A-Z]{2}-[0-9]+$ (e.g., US-001, BUG-12345)",
			})
		}
	}

	// Validate Title
	titleTrimmed := strings.TrimSpace(story.Title)
	if titleTrimmed == "" {
		errors = append(errors, core.ValidationError{
			Field:   "title",
			Rule:    "required",
			Message: "story title is required",
		})
	} else if len(titleTrimmed) > 200 {
		errors = append(errors, core.ValidationError{
			Field:   "title",
			Rule:    "length",
			Message: "story title must be 200 characters or less",
		})
	}

	// Validate Assignee (optional)
	if story.Assignee != nil && *story.Assignee != "" {
		assignee := strings.TrimSpace(*story.Assignee)
		if len(assignee) < 1 || len(assignee) > 50 {
			errors = append(errors, core.ValidationError{
				Field:   "assignee",
				Rule:    "length",
				Message: "assignee must be between 1 and 50 characters",
			})
		}
		if !usernamePattern.MatchString(assignee) {
			errors = append(errors, core.ValidationError{
				Field:   "assignee",
				Rule:    "format",
				Message: "assignee must contain only alphanumeric characters, hyphens, and underscores",
			})
		}
	}

	// Validate Priority (empty is valid, will get default)
	if story.Priority != "" {
		validPriorities := map[core.Priority]bool{
			core.PriorityLow:      true,
			core.PriorityMedium:   true,
			core.PriorityHigh:     true,
			core.PriorityCritical: true,
		}
		if !validPriorities[story.Priority] {
			errors = append(errors, core.ValidationError{
				Field:   "priority",
				Rule:    "enum",
				Message: "priority must be one of: low, medium, high, critical",
			})
		}
	}

	// Validate Status (empty is valid, will get default)
	if story.Status != "" {
		validStatuses := map[core.Status]bool{
			core.StatusTodo:   true,
			core.StatusDoing:  true,
			core.StatusReview: true,
			core.StatusDone:   true,
		}
		if !validStatuses[story.Status] {
			errors = append(errors, core.ValidationError{
				Field:   "status",
				Rule:    "enum",
				Message: "status must be one of: todo, doing, review, done",
			})
		}
	}

	// Validate dates
	if story.CreatedAt != nil && story.UpdatedAt != nil {
		if story.CreatedAt.After(*story.UpdatedAt) {
			errors = append(errors, core.ValidationError{
				Field:   "created_at",
				Rule:    "range",
				Message: "created_at must be before or equal to updated_at",
			})
		}
	}

	// Validate Tags
	if len(story.Tags) > 20 {
		errors = append(errors, core.ValidationError{
			Field:   "tags",
			Rule:    "count",
			Message: "story cannot have more than 20 tags",
		})
	}

	// Validate individual tags and deduplicate
	seenTags := make(map[string]bool)
	for i, tag := range story.Tags {
		tagTrimmed := strings.TrimSpace(tag)
		if tagTrimmed == "" {
			errors = append(errors, core.ValidationError{
				Field:   "tags",
				Rule:    "format",
				Message: fmt.Sprintf("tag at index %d is empty", i),
			})
			continue
		}
		if len(tagTrimmed) < 1 || len(tagTrimmed) > 30 {
			errors = append(errors, core.ValidationError{
				Field:   "tags",
				Rule:    "length",
				Message: fmt.Sprintf("tag at index %d must be between 1 and 30 characters", i),
			})
		}
		if !tagPattern.MatchString(tagTrimmed) {
			errors = append(errors, core.ValidationError{
				Field:   "tags",
				Rule:    "format",
				Message: fmt.Sprintf("tag at index %d must contain only alphanumeric characters, hyphens, and underscores", i),
			})
		}
		if seenTags[tagTrimmed] {
			errors = append(errors, core.ValidationError{
				Field:   "tags",
				Rule:    "uniqueness",
				Message: fmt.Sprintf("duplicate tag: %s", tagTrimmed),
			})
		}
		seenTags[tagTrimmed] = true
	}

	return errors
}
