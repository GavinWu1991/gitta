package git

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/gavin/gitta/internal/core"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Repository is the go-git backed implementation of core.GitRepository.
type Repository struct{}

// NewRepository constructs a new GitRepository implementation.
func NewRepository() *Repository {
	return &Repository{}
}

// GetBranchList returns all branches (local and remote) in the repository.
func (r *Repository) GetBranchList(ctx context.Context, repoPath string) ([]core.Branch, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, ErrNotGitRepository
	}

	// Determine current HEAD (may be detached or missing).
	headRef, _ := repo.Head()
	var headName plumbing.ReferenceName
	if headRef != nil {
		headName = headRef.Name()
	}

	refs, err := repo.References()
	if err != nil {
		return nil, ErrNotGitRepository
	}
	defer refs.Close()

	var branches []core.Branch
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if ref.Name().IsBranch() || ref.Name().IsRemote() {
			branch := branchFromRef(ref, ref.Name() == headName)
			branches = append(branches, branch)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Empty repositories are allowed; return empty slice without error.
	return branches, nil
}

// CheckBranchMerged reports whether the given branch has been merged into
// the origin remote's default branch.
func (r *Repository) CheckBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return false, ErrNotGitRepository
	}

	// Resolve branch ref (prefer local heads, then origin/<name>).
	var branchRef *plumbing.Reference
	if ref, err := repo.Reference(plumbing.NewBranchReferenceName(branchName), true); err == nil {
		branchRef = ref
	} else if ref, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", branchName), true); err == nil {
		branchRef = ref
	} else {
		return false, ErrBranchNotFound
	}

	branchCommit, err := repo.CommitObject(branchRef.Hash())
	if err != nil {
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			return false, ErrEmptyRepository
		}
		return false, err
	}

	// Find origin default branch (main then master).
	targetRef, _ := repo.Reference(plumbing.NewRemoteReferenceName("origin", "main"), true)
	if targetRef == nil {
		targetRef, _ = repo.Reference(plumbing.NewRemoteReferenceName("origin", "master"), true)
	}
	if targetRef == nil {
		// No origin remote -> treated as not merged.
		return false, nil
	}

	targetCommit, err := repo.CommitObject(targetRef.Hash())
	if err != nil {
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			return false, ErrEmptyRepository
		}
		return false, err
	}

	isAnc, err := isAncestor(ctx, targetCommit, branchCommit)
	if err != nil {
		return false, err
	}
	return isAnc, nil
}

// CreateBranch creates a new branch from the current HEAD (or current commit if detached).
func (r *Repository) CreateBranch(ctx context.Context, repoPath, branchName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return ErrNotGitRepository
	}

	head, err := repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return ErrEmptyRepository
		}
		return err
	}

	branchRef := plumbing.NewBranchReferenceName(branchName)
	if _, err := repo.Reference(branchRef, true); err == nil {
		return ErrBranchExists
	}

	ref := plumbing.NewHashReference(branchRef, head.Hash())
	if err := repo.Storer.SetReference(ref); err != nil {
		return err
	}

	return nil
}

// CheckoutBranch checks out an existing branch or creates it if missing.
func (r *Repository) CheckoutBranch(ctx context.Context, repoPath, branchName string, force bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return ErrNotGitRepository
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	if !force {
		if status, err := worktree.Status(); err == nil && !status.IsClean() {
			return ErrUncommittedChanges
		}
	}

	localRef := plumbing.NewBranchReferenceName(branchName)
	remoteRef := plumbing.NewRemoteReferenceName("origin", branchName)

	// Check if branch exists locally.
	if _, err := repo.Reference(localRef, true); err == nil {
		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: localRef,
			Force:  force,
		}); err != nil {
			if !force && strings.Contains(err.Error(), "worktree contains") {
				return ErrUncommittedChanges
			}
			return err
		}
		return nil
	}

	// If remote branch exists, create local tracking branch.
	if ref, err := repo.Reference(remoteRef, true); err == nil {
		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: localRef,
			Hash:   ref.Hash(),
			Create: true,
			Force:  force,
		}); err != nil {
			if !force && strings.Contains(err.Error(), "worktree contains") {
				return ErrUncommittedChanges
			}
			return err
		}
		return nil
	}

	// Create from current HEAD if no branch exists.
	head, err := repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return ErrEmptyRepository
		}
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: localRef,
		Hash:   head.Hash(),
		Create: true,
		Force:  force,
	}); err != nil {
		if !force && strings.Contains(err.Error(), "worktree contains") {
			return ErrUncommittedChanges
		}
		return err
	}

	return nil
}

// isAncestor reports whether ancestor is reachable from target's history.
func isAncestor(ctx context.Context, target, ancestor *object.Commit) (bool, error) {
	if target.Hash == ancestor.Hash {
		return true, nil
	}

	queue := []*object.Commit{target}
	visited := map[plumbing.Hash]bool{}

	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		c := queue[0]
		queue = queue[1:]
		if visited[c.Hash] {
			continue
		}
		visited[c.Hash] = true

		iter := c.Parents()
		for {
			parent, err := iter.Next()
			if err == object.ErrParentNotFound || errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return false, err
			}
			if parent.Hash == ancestor.Hash {
				return true, nil
			}
			queue = append(queue, parent)
		}
	}
	return false, nil
}
