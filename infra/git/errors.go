package git

import "errors"

var (
	// ErrNotGitRepository indicates the path is not a Git repository.
	ErrNotGitRepository = errors.New("not a git repository")
	// ErrBranchNotFound indicates the requested branch does not exist.
	ErrBranchNotFound = errors.New("branch not found")
	// ErrBranchExists indicates the branch already exists locally.
	ErrBranchExists = errors.New("branch already exists")
	// ErrUncommittedChanges indicates the working tree has uncommitted changes that block checkout.
	ErrUncommittedChanges = errors.New("uncommitted changes detected")
	// ErrEmptyRepository indicates the repository has no commits.
	ErrEmptyRepository = errors.New("empty repository")
)
