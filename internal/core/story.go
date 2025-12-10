package core

import (
	"context"
	"fmt"
	"time"
)

// Story represents a user story or task stored as a Markdown file with YAML frontmatter.
// It contains both structured metadata (from YAML frontmatter) and narrative content
// (from Markdown body). Stories are the primary data structure used throughout the
// Gitta system for task management.
type Story struct {
	// Required fields
	ID    string `yaml:"id"`    // Story identifier (e.g., "US-001", "US-002")
	Title string `yaml:"title"` // Story title/name

	// Optional metadata fields
	Assignee  *string    `yaml:"assignee,omitempty"`   // Assigned developer (nil if unset)
	Priority  Priority   `yaml:"priority,omitempty"`   // Priority level (default: Medium)
	Status    Status     `yaml:"status,omitempty"`     // Current status (default: Todo)
	CreatedAt *time.Time `yaml:"created_at,omitempty"` // Creation timestamp (nil if unset)
	UpdatedAt *time.Time `yaml:"updated_at,omitempty"` // Last update timestamp (nil if unset)
	Tags      []string   `yaml:"tags,omitempty"`       // Tags for categorization

	// Content
	Body string `yaml:"-"` // Markdown body content (not in frontmatter)
}

// Priority represents the priority level of a story.
// Priority levels are used to indicate the relative importance of a story
// and can be used for filtering and sorting.
type Priority string

const (
	PriorityLow      Priority = "low"      // Low priority story
	PriorityMedium   Priority = "medium"   // Medium priority (default)
	PriorityHigh     Priority = "high"     // High priority story
	PriorityCritical Priority = "critical" // Critical priority story
)

// Status represents the current workflow status of a story.
// Status values track the progress of a story through the development workflow.
// Status may be derived from Git branch state in future features (StatusEngine),
// but the parser allows explicit status in frontmatter for manual override.
type Status string

const (
	StatusTodo   Status = "todo"   // Story exists, no associated branch
	StatusDoing  Status = "doing"  // Branch created (feat/<story-id>)
	StatusReview Status = "review" // PR submitted
	StatusDone   Status = "done"   // Branch merged to main
)

// ParseError represents an error that occurred during parsing of a story file.
// ParseError is returned when YAML frontmatter is malformed or cannot be parsed.
// It includes the file path, line number (if available), and a descriptive message.
type ParseError struct {
	FilePath string // Path to the file that failed to parse
	Line     int    // Line number where error occurred (0 if unknown)
	Message  string // Specific error message
	Cause    error  // Underlying error (if any)
}

// Error implements the error interface for ParseError.
func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error in %s at line %d: %s", e.FilePath, e.Line, e.Message)
	}
	return fmt.Sprintf("parse error in %s: %s", e.FilePath, e.Message)
}

// Unwrap returns the underlying error, if any.
func (e *ParseError) Unwrap() error {
	return e.Cause
}

// ValidationError represents a validation error for a story field.
// ValidationError provides structured information about validation failures,
// including which field failed, what rule was violated, and a human-readable message.
// Multiple ValidationErrors may be returned for a single story if multiple
// validation rules are violated.
type ValidationError struct {
	Field   string // Field name (e.g., "id", "priority")
	Rule    string // Validation rule violated (e.g., "required", "format", "enum")
	Message string // Human-readable error message
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' (rule: %s): %s", e.Field, e.Rule, e.Message)
}

// IOError represents an I/O error that occurred during file operations.
// IOError wraps underlying filesystem errors (e.g., permission denied, file not found)
// with context about the operation being performed and the file path.
type IOError struct {
	Operation string // Operation type (e.g., "read", "write")
	FilePath  string // Path to the file
	Cause     error  // Underlying os error
}

// Error implements the error interface for IOError.
func (e *IOError) Error() string {
	return fmt.Sprintf("I/O error during %s of %s: %v", e.Operation, e.FilePath, e.Cause)
}

// Unwrap returns the underlying error.
func (e *IOError) Unwrap() error {
	return e.Cause
}

// StoryParser defines operations for reading and writing story files.
// StoryParser is the core interface for parsing Markdown story files with YAML frontmatter.
// Implementations should handle file I/O, YAML parsing, Markdown body extraction,
// and validation according to the contract defined in contracts/parser.md.
type StoryParser interface {
	// ReadStory reads a Markdown file and parses it into a Story struct.
	// It extracts YAML frontmatter metadata and Markdown body content.
	// Returns an error if the file cannot be read or parsed.
	ReadStory(ctx context.Context, filePath string) (*Story, error)

	// WriteStory writes a Story struct to a Markdown file.
	// It marshals the story metadata to YAML frontmatter and writes the body content.
	// Uses atomic writes to prevent corruption. Returns an error if validation fails
	// or the file cannot be written.
	WriteStory(ctx context.Context, filePath string, story *Story) error

	// ValidateStory validates a Story struct against business rules.
	// Returns a slice of ValidationErrors describing any validation failures.
	// An empty slice indicates the story is valid.
	ValidateStory(story *Story) []ValidationError
}

// StoryRepository defines operations for listing and discovering story files.
// Implementations encapsulate directory scanning and Sprint discovery logic.
// All methods must respect context cancellation for long-running operations.
type StoryRepository interface {
	// ListStories scans a directory for Markdown story files and returns parsed stories.
	// Returns an empty slice (not error) when the directory is empty.
	// Should skip files that fail to parse while returning an aggregated error when
	// critical failures occur (e.g., unreadable directory).
	ListStories(ctx context.Context, dirPath string) ([]*Story, error)

	// FindCurrentSprint determines the current Sprint directory path within sprintsDir.
	// Implementations should use lexicographic ordering (case-insensitive) to pick the
	// highest Sprint name when multiple exist.
	FindCurrentSprint(ctx context.Context, sprintsDir string) (string, error)
}
