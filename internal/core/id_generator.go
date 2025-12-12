package core

import (
	"context"
	"errors"
)

// IDGenerator generates unique story IDs with conflict detection.
// Implementations must handle concurrent access safely using file locking
// or equivalent synchronization mechanisms.
type IDGenerator interface {
	// GenerateNextID generates the next unique ID for the given prefix.
	// The ID format is "{prefix}-{number}" where number is an incrementing integer.
	// Returns an error if ID generation fails (lock timeout, counter corruption).
	GenerateNextID(ctx context.Context, prefix string) (string, error)
}

var (
	// ErrLockTimeout indicates that lock acquisition timed out after maximum wait period.
	ErrLockTimeout = errors.New("failed to acquire ID counter lock: timeout")

	// ErrCounterCorrupted indicates that the counter file is corrupted or unreadable.
	ErrCounterCorrupted = errors.New("ID counter file corrupted, please fix manually")

	// ErrInvalidPrefix indicates that the provided prefix does not match the required format.
	ErrInvalidPrefix = errors.New("invalid prefix format: must be 2+ uppercase letters")
)
