package git

import "errors"

var (
	// ErrNotGitRepository indicates the path is not a Git repository.
	ErrNotGitRepository = errors.New("not a git repository")
	// ErrBranchNotFound indicates the requested branch does not exist.
	ErrBranchNotFound = errors.New("branch not found")
	// ErrEmptyRepository indicates the repository has no commits.
	ErrEmptyRepository = errors.New("empty repository")
)
