package core

import (
	"context"
	"errors"
	"time"
)

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

	// CreateBranch creates a new branch from the current HEAD (or current commit if detached).
	// Returns ErrBranchExists if the branch already exists locally.
	CreateBranch(ctx context.Context, repoPath, branchName string) error

	// CheckoutBranch checks out an existing branch or creates it if missing.
	// If the branch exists remotely but not locally, implementations should create a local
	// tracking branch. If force is false and the working tree has uncommitted changes,
	// implementations should return ErrUncommittedChanges.
	CheckoutBranch(ctx context.Context, repoPath, branchName string, force bool) error
}

var (
	// ErrInvalidCommit indicates the commit hash is invalid or doesn't exist.
	ErrInvalidCommit = errors.New("invalid commit hash")
	// ErrInsufficientHistory indicates insufficient Git history for analysis.
	ErrInsufficientHistory = errors.New("insufficient Git history")
)

// CommitSnapshot represents the state of files in a sprint directory at a specific commit.
type CommitSnapshot struct {
	// CommitHash is the Git commit hash (full SHA-1).
	CommitHash string
	// CommitDate is the commit timestamp.
	CommitDate time.Time
	// Author is the commit author name.
	Author string
	// Message is the commit message.
	Message string
	// Files is a map of file paths (relative to repo root) to Story entities.
	// Only includes files in the sprint directory.
	Files map[string]*Story
}

// AnalyzeHistoryRequest contains parameters for analyzing Git commit history.
type AnalyzeHistoryRequest struct {
	// RepoPath is the filesystem path to the Git repository root.
	RepoPath string
	// SprintDir is the relative path to the sprint directory from repo root.
	// Example: "sprints/Sprint-01"
	SprintDir string
	// StartDate is the start date of the analysis period.
	StartDate time.Time
	// EndDate is the end date of the analysis period.
	EndDate time.Time
	// IncludeMergeCommits indicates whether to include merge commits.
	// Default: false (only include regular commits).
	IncludeMergeCommits bool
}

// GitHistoryAnalyzer provides Git commit history traversal and file state reconstruction.
type GitHistoryAnalyzer interface {
	// AnalyzeSprintHistory traverses Git commits affecting a sprint directory.
	AnalyzeSprintHistory(ctx context.Context, req AnalyzeHistoryRequest) ([]CommitSnapshot, error)
	// ReconstructFileState reconstructs the file tree state at a specific commit.
	ReconstructFileState(ctx context.Context, repoPath string, commitHash string, dirPath string) (map[string]*Story, error)
}
