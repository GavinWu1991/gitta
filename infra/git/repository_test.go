package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// createTempRepo initializes an empty repository for tests.
func createTempRepo(t *testing.T) (*git.Repository, string) {
	t.Helper()
	repo, err := git.PlainInit(t.TempDir(), false)
	if err != nil {
		t.Fatalf("failed to init temp repo: %v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	return repo, worktree.Filesystem.Root()
}

// helper to create a commit with a file and return its hash.
func commitFile(t *testing.T, repo *git.Repository, path string, message string) plumbing.Hash {
	t.Helper()
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(path, "file.txt"), []byte(message), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := wt.Add("file.txt"); err != nil {
		t.Fatalf("add: %v", err)
	}
	hash, err := wt.Commit(message, &git.CommitOptions{})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	return hash
}

func TestRepository_GetBranchList_LocalAndRemote(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	repoImpl := NewRepository()

	// initial commit on main
	mainHash := commitFile(t, repo, repoPath, "main")

	// create local branch feature
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("feature"),
		Create: true,
	}); err != nil {
		t.Fatalf("checkout feature: %v", err)
	}
	commitFile(t, repo, repoPath, "feature")

	// create remote tracking ref for origin/main
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		mainHash,
	)); err != nil {
		t.Fatalf("set remote ref: %v", err)
	}

	branches, err := repoImpl.GetBranchList(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("GetBranchList error: %v", err)
	}

	var names []string
	var current string
	for _, b := range branches {
		names = append(names, b.Name)
		if b.IsCurrent {
			current = b.Name
		}
	}

	assertContains(t, names, "master")
	assertContains(t, names, "feature")
	assertContains(t, names, "origin/main")
	if current != "feature" {
		t.Fatalf("expected current branch feature, got %s", current)
	}
}

func TestRepository_GetBranchList_NonGitPath(t *testing.T) {
	repoImpl := NewRepository()
	_, err := repoImpl.GetBranchList(context.Background(), t.TempDir())
	if err == nil {
		t.Fatalf("expected error for non-git path")
	}
	if err != ErrNotGitRepository {
		t.Fatalf("expected ErrNotGitRepository, got %v", err)
	}
}

func TestRepository_GetBranchList_EmptyRepo(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	_ = repo // unused except for creation
	repoImpl := NewRepository()

	branches, err := repoImpl.GetBranchList(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 0 {
		t.Fatalf("expected no branches in empty repo, got %d", len(branches))
	}
}

func assertContains(t *testing.T, items []string, target string) {
	t.Helper()
	for _, it := range items {
		if it == target {
			return
		}
	}
	t.Fatalf("expected %s in %v", target, items)
}

func TestRepository_CheckBranchMerged_Merged(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	impl := NewRepository()

	// initial commit on master
	masterHash := commitFile(t, repo, repoPath, "master")

	// create feature branch and commit
	wt, _ := repo.Worktree()
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("feature"),
		Create: true,
	}); err != nil {
		t.Fatalf("checkout feature: %v", err)
	}
	featureHash := commitFile(t, repo, repoPath, "feature")

	// fast-forward master to feature (simulate merge)
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewBranchReferenceName("master"),
		featureHash,
	)); err != nil {
		t.Fatalf("ff master: %v", err)
	}

	// origin/main at merged state
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		featureHash,
	)); err != nil {
		t.Fatalf("set origin main: %v", err)
	}

	merged, err := impl.CheckBranchMerged(context.Background(), repoPath, "feature")
	if err != nil {
		t.Fatalf("CheckBranchMerged: %v", err)
	}
	if !merged {
		t.Fatalf("expected merged=true, got false (masterHash=%s featureHash=%s)", masterHash, featureHash)
	}
}

func TestRepository_CheckBranchMerged_NotMerged(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	impl := NewRepository()

	masterHash := commitFile(t, repo, repoPath, "master")

	// feature branch commit
	wt, _ := repo.Worktree()
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("feature"),
		Create: true,
	}); err != nil {
		t.Fatalf("checkout feature: %v", err)
	}
	featureHash := commitFile(t, repo, repoPath, "feature")

	// origin/main remains at masterHash
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		masterHash,
	)); err != nil {
		t.Fatalf("set origin main: %v", err)
	}

	merged, err := impl.CheckBranchMerged(context.Background(), repoPath, "feature")
	if err != nil {
		t.Fatalf("CheckBranchMerged: %v", err)
	}
	if merged {
		t.Fatalf("expected merged=false, got true (masterHash=%s featureHash=%s)", masterHash, featureHash)
	}
}

func TestRepository_CheckBranchMerged_NoOrigin(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	impl := NewRepository()

	// create initial commit
	commitFile(t, repo, repoPath, "init")

	// create branch commit
	wt, _ := repo.Worktree()
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("feature"),
		Create: true,
	}); err != nil {
		t.Fatalf("checkout feature: %v", err)
	}
	commitFile(t, repo, repoPath, "feature")

	merged, err := impl.CheckBranchMerged(context.Background(), repoPath, "feature")
	if err != nil {
		t.Fatalf("CheckBranchMerged: %v", err)
	}
	if merged {
		t.Fatalf("expected merged=false when no origin remote")
	}
}

func TestRepository_CreateBranch_And_Checkout(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	impl := NewRepository()

	commitFile(t, repo, repoPath, "init")

	if err := impl.CheckoutBranch(context.Background(), repoPath, "feat/US-001", false); err != nil {
		t.Fatalf("CheckoutBranch: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if head.Name().Short() != "feat/US-001" {
		t.Fatalf("expected head feat/US-001, got %s", head.Name().Short())
	}
}

func TestRepository_CheckoutBranch_UncommittedChanges(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	impl := NewRepository()

	commitFile(t, repo, repoPath, "init")

	// create uncommitted change on tracked file
	if err := os.WriteFile(filepath.Join(repoPath, "file.txt"), []byte("modified"), 0o644); err != nil {
		t.Fatalf("modify tracked file: %v", err)
	}

	err := impl.CheckoutBranch(context.Background(), repoPath, "feat/US-002", false)
	if err == nil {
		t.Fatalf("expected error due to uncommitted changes")
	}
	if err != ErrUncommittedChanges {
		t.Fatalf("expected ErrUncommittedChanges, got %v", err)
	}
}

func TestRepository_CheckoutBranch_RemoteTracking(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	impl := NewRepository()

	hash := commitFile(t, repo, repoPath, "init")

	// Add remote ref for origin/feat/US-003
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "feat/US-003"),
		hash,
	)); err != nil {
		t.Fatalf("set remote ref: %v", err)
	}

	if err := impl.CheckoutBranch(context.Background(), repoPath, "feat/US-003", false); err != nil {
		t.Fatalf("CheckoutBranch remote: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if head.Name().Short() != "feat/US-003" {
		t.Fatalf("expected head feat/US-003, got %s", head.Name().Short())
	}
}

func TestRepository_CheckBranchMerged_MissingBranch(t *testing.T) {
	_, repoPath := createTempRepo(t)
	impl := NewRepository()

	_, err := impl.CheckBranchMerged(context.Background(), repoPath, "missing")
	if err != ErrBranchNotFound {
		t.Fatalf("expected ErrBranchNotFound, got %v", err)
	}
}

func TestRepository_CheckBranchMerged_DetachedHead(t *testing.T) {
	repo, repoPath := createTempRepo(t)
	impl := NewRepository()

	// initial commit on master
	masterHash := commitFile(t, repo, repoPath, "master")

	// create feature branch and commit
	wt, _ := repo.Worktree()
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("feature"),
		Create: true,
	}); err != nil {
		t.Fatalf("checkout feature: %v", err)
	}
	featureHash := commitFile(t, repo, repoPath, "feature")

	// origin/main at master
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		masterHash,
	)); err != nil {
		t.Fatalf("set origin main: %v", err)
	}

	// detach HEAD to feature commit
	if err := wt.Checkout(&git.CheckoutOptions{
		Hash:  featureHash,
		Force: true,
	}); err != nil {
		t.Fatalf("detach head: %v", err)
	}

	// Should still evaluate merge status (false)
	merged, err := impl.CheckBranchMerged(context.Background(), repoPath, "feature")
	if err != nil {
		t.Fatalf("CheckBranchMerged: %v", err)
	}
	if merged {
		t.Fatalf("expected merged=false in detached state")
	}

	// Now fast-forward origin/main to feature (simulate merge)
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		featureHash,
	)); err != nil {
		t.Fatalf("ff origin main: %v", err)
	}

	merged, err = impl.CheckBranchMerged(context.Background(), repoPath, "feature")
	if err != nil {
		t.Fatalf("CheckBranchMerged: %v", err)
	}
	if !merged {
		t.Fatalf("expected merged=true after origin fast-forward")
	}
}
