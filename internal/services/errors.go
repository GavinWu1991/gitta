package services

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidStory indicates that the story provided is nil or invalid.
	ErrInvalidStory = errors.New("story is nil or invalid")

	// ErrInvalidInput indicates that required input is missing or invalid.
	ErrInvalidInput = errors.New("required input is missing or invalid")

	// ErrContextCancelled indicates that the context was cancelled during operation.
	ErrContextCancelled = errors.New("context was cancelled")

	// ErrAssigneeUpdateFailed indicates that updating the assignee field failed.
	ErrAssigneeUpdateFailed = errors.New("failed to update assignee")

	// ErrNotGitRepository indicates the working directory is not a Git repository.
	ErrNotGitRepository = errors.New("not a git repository")

	// ErrWorkspaceExists indicates gitta workspace folders already exist and overwrite is disallowed.
	ErrWorkspaceExists = errors.New("gitta workspace already exists")

	// ErrAlreadyConsolidated indicates the repository is already using the consolidated tasks/ structure.
	ErrAlreadyConsolidated = errors.New("repository already uses consolidated structure")

	// ErrMigrationConflict indicates migration targets already exist without --force.
	ErrMigrationConflict = errors.New("migration target directories already exist")
)

// AssigneeUpdateError wraps an assignee update failure with file context.
type AssigneeUpdateError struct {
	FilePath string
	Cause    error
}

func (e *AssigneeUpdateError) Error() string {
	return fmt.Sprintf("failed to update assignee for %s: %v", e.FilePath, e.Cause)
}

func (e *AssigneeUpdateError) Unwrap() error {
	return e.Cause
}
