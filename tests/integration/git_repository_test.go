package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	gittagit "github.com/gavin/gitta/infra/git"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestGetBranchListIntegration(t *testing.T) {
	repoPath := t.TempDir()
	repo, err := gogit.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	// Commit on default branch (master/main depending on platform; go-git defaults to master).
	if err := os.WriteFile(filepath.Join(repoPath, "readme.md"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := wt.Add("readme.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	hash, err := wt.Commit("init", testCommitOptionsGogit())
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Add origin/main remote ref pointing to the commit.
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		hash,
	)); err != nil {
		t.Fatalf("set remote ref: %v", err)
	}

	impl := gittagit.NewRepository()
	branches, err := impl.GetBranchList(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("GetBranchList: %v", err)
	}
	if len(branches) == 0 {
		t.Fatalf("expected branches, got none")
	}
	var hasLocal, hasRemote bool
	for _, b := range branches {
		if b.Type == 0 && b.Name != "" { // local
			hasLocal = true
		}
		if b.Type == 1 && b.RemoteName == "origin" {
			hasRemote = true
		}
	}
	if !hasLocal {
		t.Fatalf("expected at least one local branch")
	}
	if !hasRemote {
		t.Fatalf("expected origin remote branch")
	}
}

func TestCheckBranchMergedDetachedHeadIntegration(t *testing.T) {
	repoPath := t.TempDir()
	repo, err := gogit.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	// master commit
	if err := os.WriteFile(filepath.Join(repoPath, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if _, err := wt.Add("a.txt"); err != nil {
		t.Fatalf("add a: %v", err)
	}
	masterHash, err := wt.Commit("init", testCommitOptionsGogit())
	if err != nil {
		t.Fatalf("commit init: %v", err)
	}

	// feature branch
	if err := wt.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("feature"),
		Create: true,
	}); err != nil {
		t.Fatalf("checkout feature: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}
	if _, err := wt.Add("b.txt"); err != nil {
		t.Fatalf("add b: %v", err)
	}
	featureHash, err := wt.Commit("feature", testCommitOptionsGogit())
	if err != nil {
		t.Fatalf("commit feature: %v", err)
	}

	// origin/main at master
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		masterHash,
	)); err != nil {
		t.Fatalf("set origin main: %v", err)
	}

	// detach HEAD to feature commit
	if err := wt.Checkout(&gogit.CheckoutOptions{
		Hash:  featureHash,
		Force: true,
	}); err != nil {
		t.Fatalf("detach head: %v", err)
	}

	impl := gittagit.NewRepository()

	merged, err := impl.CheckBranchMerged(context.Background(), repoPath, "feature")
	if err != nil {
		t.Fatalf("CheckBranchMerged: %v", err)
	}
	if merged {
		t.Fatalf("expected merged=false before origin update")
	}

	// simulate merge by fast-forwarding origin/main
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
