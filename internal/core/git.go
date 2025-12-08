package core

import "context"

// BranchType represents the type of a Git branch.
type BranchType int

const (
	// BranchTypeLocal represents a local branch.
	BranchTypeLocal BranchType = iota
	// BranchTypeRemote represents a remote tracking branch.
	BranchTypeRemote
)

// Branch describes a Git branch with minimal metadata needed by the domain.
type Branch struct {
	// Name is the branch name (e.g., "main", "feature-1").
	Name string
	// Type indicates whether this is a local or remote branch.
	Type BranchType
	// RemoteName is set when Type is BranchTypeRemote (e.g., "origin").
	RemoteName string
	// IsCurrent reports whether this is the currently checked-out branch.
	IsCurrent bool
	// CommitHash is the tip commit SHA (may be empty if unavailable).
	CommitHash string
}

// GitRepository defines read-only Git operations used by the domain.
type GitRepository interface {
	// GetBranchList returns all branches (local and remote) in the repository.
	// Returns an error if the path is not a Git repository or if branch listing fails.
	GetBranchList(ctx context.Context, repoPath string) ([]Branch, error)

	// CheckBranchMerged reports whether the given branch has been merged into the
	// "origin" remote's default branch (origin/main or origin/master). If origin
	// is missing, the result is false without error.
	CheckBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error)
}
