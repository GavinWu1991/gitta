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
