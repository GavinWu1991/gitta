package integration

import (
	"time"

	ggit "github.com/go-git/go-git/v5"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// testCommitOptions returns CommitOptions with default author for tests.
// This ensures commits work in CI environments where Git user config may not be set.
func testCommitOptions() *ggit.CommitOptions {
	return &ggit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	}
}

// testCommitOptionsGogit returns CommitOptions for gogit (alias) package.
func testCommitOptionsGogit() *gogit.CommitOptions {
	return &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	}
}
