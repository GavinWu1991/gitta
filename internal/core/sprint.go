package core

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrSprintExists indicates a sprint with the given name already exists.
	ErrSprintExists = errors.New("sprint already exists")
	// ErrSprintNotFound indicates the requested sprint could not be found.
	ErrSprintNotFound = errors.New("sprint not found")
)

// Sprint represents a time-bounded work period with associated directory containing task files.
type Sprint struct {
	// Name is the sprint identifier (e.g., "Sprint-01", "Sprint-02").
	Name string
	// StartDate is the sprint start date.
	StartDate time.Time
	// EndDate is the sprint end date (calculated from start + duration).
	EndDate time.Time
	// Duration is the sprint duration string (e.g., "2w", "14d").
	Duration string
	// DirectoryPath is the filesystem path to the sprint directory.
	DirectoryPath string
	// CreatedAt is the sprint creation timestamp.
	CreatedAt time.Time
	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time
}

// BurndownDataPoint represents a snapshot of remaining work at a specific point in time,
// reconstructed from Git history.
type BurndownDataPoint struct {
	// Date is the date of the snapshot (time component ignored).
	Date time.Time
	// RemainingPoints is the total remaining story points.
	RemainingPoints int
	// RemainingTasks is the count of incomplete tasks.
	RemainingTasks int
	// TotalPoints is the total points at sprint start (optional).
	TotalPoints *int
	// TotalTasks is the total tasks at sprint start (optional).
	TotalTasks *int
}

// StartSprintRequest contains parameters for creating a new sprint.
type StartSprintRequest struct {
	// Name is the sprint identifier (e.g., "Sprint-01", "Sprint-02").
	// If empty, auto-generates next sequential name.
	Name string
	// Duration is the sprint duration string (e.g., "2w", "14d").
	// Format: <number><unit> where unit is "w" (weeks) or "d" (days).
	// Default: "2w" if empty.
	Duration string
	// StartDate is the sprint start date.
	// If nil, defaults to today's date.
	StartDate *time.Time
}

// RolloverRequest contains parameters for rolling over tasks from one sprint to another.
type RolloverRequest struct {
	// SourceSprintPath is the filesystem path to the source sprint directory.
	SourceSprintPath string
	// TargetSprintPath is the filesystem path to the target sprint directory.
	TargetSprintPath string
	// SelectedTaskIDs is the list of task IDs to rollover.
	SelectedTaskIDs []string
}

// SprintRepository defines operations for sprint directory management and current sprint link.
type SprintRepository interface {
	// CreateSprint creates a new sprint directory at sprintDir and returns the sprint entity.
	CreateSprint(ctx context.Context, sprintDir string, name string, startDate time.Time, duration string) (*Sprint, error)
	// FindCurrentSprint returns the path to the current sprint directory.
	FindCurrentSprint(ctx context.Context, sprintsDir string) (string, error)
	// SetCurrentSprint creates/updates the current sprint link to point to the given sprint.
	SetCurrentSprint(ctx context.Context, sprintsDir string, sprintPath string) error
	// ListSprints returns all sprint directories in lexicographic order.
	ListSprints(ctx context.Context, sprintsDir string) ([]string, error)
	// SprintExists checks if a sprint directory exists.
	SprintExists(ctx context.Context, sprintPath string) (bool, error)
}

// SprintService provides sprint management operations.
type SprintService interface {
	// StartSprint creates a new sprint, sets up current sprint link, and calculates end date.
	StartSprint(ctx context.Context, req StartSprintRequest) (*Sprint, error)
	// CloseSprint identifies unfinished tasks in the current sprint.
	CloseSprint(ctx context.Context, sprintPath string) ([]*Story, error)
	// RolloverTasks moves selected tasks from source sprint to target sprint with metadata updates.
	RolloverTasks(ctx context.Context, req RolloverRequest) error
	// GenerateBurndown analyzes Git history and generates burndown data points.
	GenerateBurndown(ctx context.Context, sprintPath string) ([]BurndownDataPoint, error)
}

// ValidateSprintName validates a sprint name format.
// Returns an error if the name is invalid.
func ValidateSprintName(name string) error {
	if name == "" {
		return errors.New("sprint name cannot be empty")
	}
	// Check for path separators
	if containsAny(name, "/", "\\", "..") {
		return errors.New("sprint name cannot contain path separators or parent directory references")
	}
	// Basic validation - can be extended with pattern matching if needed
	return nil
}

// ParseDuration parses a duration string (e.g., "2w", "14d") and returns the number of days.
// Returns an error if the format is invalid.
func ParseDuration(duration string) (int, error) {
	if duration == "" {
		return 14, nil // Default: 2 weeks
	}

	var days int
	var unit string

	// Parse number and unit
	if len(duration) < 2 {
		return 0, errors.New("invalid duration format: must be <number><unit>")
	}

	// Extract unit (last character)
	unit = duration[len(duration)-1:]
	// Extract number (everything except last character)
	_, err := fmt.Sscanf(duration[:len(duration)-1], "%d", &days)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %w", err)
	}

	if days <= 0 {
		return 0, errors.New("duration must be positive")
	}

	switch unit {
	case "w", "W":
		return days * 7, nil
	case "d", "D":
		return days, nil
	default:
		return 0, fmt.Errorf("invalid duration unit: %s (must be 'w' or 'd')", unit)
	}
}

// CalculateEndDate calculates the sprint end date from start date and duration.
func CalculateEndDate(startDate time.Time, duration string) (time.Time, error) {
	days, err := ParseDuration(duration)
	if err != nil {
		return time.Time{}, err
	}
	return startDate.AddDate(0, 0, days), nil
}

// containsAny checks if the string contains any of the given substrings.
func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(substr) > 0 && len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
