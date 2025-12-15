package core

import (
	"errors"
	"fmt"
	"strings"
)

// SprintStatus represents the lifecycle state of a sprint.
type SprintStatus int

const (
	// StatusActive represents an active sprint (prefix: "!", ASCII 33).
	StatusActive SprintStatus = iota
	// StatusReady represents a ready sprint prepared for activation (prefix: "+", ASCII 43).
	StatusReady
	// StatusPlanning represents a planning sprint for future work (prefix: "@", ASCII 64).
	StatusPlanning
	// StatusArchived represents an archived sprint (prefix: "~", ASCII 126).
	StatusArchived
)

// String returns the lowercase string representation of the status.
func (s SprintStatus) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusReady:
		return "ready"
	case StatusPlanning:
		return "planning"
	case StatusArchived:
		return "archived"
	default:
		return "unknown"
	}
}

// Prefix returns the ASCII prefix character for the status.
func (s SprintStatus) Prefix() string {
	switch s {
	case StatusActive:
		return "!"
	case StatusReady:
		return "+"
	case StatusPlanning:
		return "@"
	case StatusArchived:
		return "~"
	default:
		return ""
	}
}

// ASCIIValue returns the ASCII code for the status prefix character.
func (s SprintStatus) ASCIIValue() int {
	switch s {
	case StatusActive:
		return 33
	case StatusReady:
		return 43
	case StatusPlanning:
		return 64
	case StatusArchived:
		return 126
	default:
		return 0
	}
}

// ParseStatus parses a status string and returns the corresponding SprintStatus.
// The parsing is case-insensitive and trims whitespace.
// Returns an error if the status string is invalid.
func ParseStatus(s string) (SprintStatus, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "active":
		return StatusActive, nil
	case "ready":
		return StatusReady, nil
	case "planning":
		return StatusPlanning, nil
	case "archived":
		return StatusArchived, nil
	default:
		return StatusActive, fmt.Errorf("invalid status: %q (must be one of: active, ready, planning, archived)", s)
	}
}

// ValidateTransition validates whether a status transition is allowed.
// Returns an error if the transition is invalid.
func ValidateTransition(from, to SprintStatus) error {
	// Archived sprints are read-only
	if from == StatusArchived {
		return errors.New("cannot transition from archived status (archived sprints are read-only)")
	}

	// Active → Ready/Planning is not allowed (must archive first)
	if from == StatusActive && (to == StatusReady || to == StatusPlanning) {
		return fmt.Errorf("cannot transition from active to %s (active sprint must be archived first)", to.String())
	}

	// Valid transitions:
	// - Planning → Active
	// - Ready → Active
	// - Active → Archived
	// - Any → Archived (for future manual archiving)
	if from == StatusPlanning && to == StatusActive {
		return nil
	}
	if from == StatusReady && to == StatusActive {
		return nil
	}
	if from == StatusActive && to == StatusArchived {
		return nil
	}
	if to == StatusArchived {
		return nil // Allow any status to Archived (for future manual archiving)
	}

	// Same status is valid (no-op)
	if from == to {
		return nil
	}

	return fmt.Errorf("invalid transition from %s to %s", from.String(), to.String())
}
