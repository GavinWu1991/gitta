package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	gittagit "github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestStatusEngine_Integration(t *testing.T) {
	repoPath := t.TempDir()
	repo, err := gogit.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	// Create initial commit on main
	if err := os.WriteFile(filepath.Join(repoPath, "readme.md"), []byte("# Test Repo"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := wt.Add("readme.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	mainHash, err := wt.Commit("Initial commit", &gogit.CommitOptions{})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Create a feature branch with a new commit (not merged)
	featBranchName := plumbing.NewBranchReferenceName("feat/US-001")
	featRef := plumbing.NewHashReference(featBranchName, mainHash)
	if err := repo.Storer.SetReference(featRef); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	// Checkout feature branch and make a new commit
	if err := wt.Checkout(&gogit.CheckoutOptions{
		Branch: featBranchName,
		Create: false,
	}); err != nil {
		t.Fatalf("checkout branch: %v", err)
	}

	// Make a commit on feature branch
	if err := os.WriteFile(filepath.Join(repoPath, "feature.md"), []byte("# Feature"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := wt.Add("feature.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	featHash, err := wt.Commit("Feature commit", &gogit.CommitOptions{})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Update branch reference to point to new commit
	featRef = plumbing.NewHashReference(featBranchName, featHash)
	if err := repo.Storer.SetReference(featRef); err != nil {
		t.Fatalf("update branch ref: %v", err)
	}

	// Add origin/main remote ref
	if err := repo.Storer.SetReference(plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		mainHash,
	)); err != nil {
		t.Fatalf("set remote ref: %v", err)
	}

	// Get branch list using real GitRepository
	gitRepo := gittagit.NewRepository()
	ctx := context.Background()
	branchList, err := gitRepo.GetBranchList(ctx, repoPath)
	if err != nil {
		t.Fatalf("GetBranchList: %v", err)
	}

	// Create StatusEngine with real GitRepository
	engine := services.NewStatusEngineWithRepository(gitRepo)

	// Test: Story with no matching branch should return Todo
	storyNoBranch := &core.Story{
		ID:    "US-999",
		Title: "Non-existent story",
	}
	status, err := engine.DeriveStatus(ctx, storyNoBranch, branchList, repoPath)
	if err != nil {
		t.Fatalf("DeriveStatus failed: %v", err)
	}
	if status != core.StatusTodo {
		t.Errorf("Expected StatusTodo for story with no branch, got %v", status)
	}

	// Test: Story with local branch should return Doing
	storyWithBranch := &core.Story{
		ID:    "US-001",
		Title: "Test story",
	}
	status, err = engine.DeriveStatus(ctx, storyWithBranch, branchList, repoPath)
	if err != nil {
		t.Fatalf("DeriveStatus failed: %v", err)
	}
	if status != core.StatusDoing {
		t.Errorf("Expected StatusDoing for story with local branch, got %v", status)
	}
}

func TestStatusEngine_EmptyRepository(t *testing.T) {
	repoPath := t.TempDir()
	_, err := gogit.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	// Empty repo - no commits
	gitRepo := gittagit.NewRepository()
	ctx := context.Background()
	branchList, err := gitRepo.GetBranchList(ctx, repoPath)
	if err != nil {
		t.Fatalf("GetBranchList: %v", err)
	}

	engine := services.NewStatusEngineWithRepository(gitRepo)
	story := &core.Story{
		ID:    "US-001",
		Title: "Test story",
	}

	// Should return Todo (no branches in empty repo)
	status, err := engine.DeriveStatus(ctx, story, branchList, repoPath)
	if err != nil {
		t.Fatalf("DeriveStatus failed: %v", err)
	}
	if status != core.StatusTodo {
		t.Errorf("Expected StatusTodo for empty repository, got %v", status)
	}
}

func TestStatusEngine_NoRemotes(t *testing.T) {
	repoPath := t.TempDir()
	repo, err := gogit.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	// Create initial commit
	if err := os.WriteFile(filepath.Join(repoPath, "readme.md"), []byte("# Test"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := wt.Add("readme.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	hash, err := wt.Commit("Initial commit", &gogit.CommitOptions{})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Create feature branch (no remote configured)
	featBranchName := plumbing.NewBranchReferenceName("feat/US-001")
	if err := repo.Storer.SetReference(plumbing.NewHashReference(featBranchName, hash)); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	gitRepo := gittagit.NewRepository()
	ctx := context.Background()
	branchList, err := gitRepo.GetBranchList(ctx, repoPath)
	if err != nil {
		t.Fatalf("GetBranchList: %v", err)
	}

	engine := services.NewStatusEngineWithRepository(gitRepo)
	story := &core.Story{
		ID:    "US-001",
		Title: "Test story",
	}

	// Should return Doing (local branch, no remote)
	status, err := engine.DeriveStatus(ctx, story, branchList, repoPath)
	if err != nil {
		t.Fatalf("DeriveStatus failed: %v", err)
	}
	if status != core.StatusDoing {
		t.Errorf("Expected StatusDoing for local branch without remote, got %v", status)
	}
}

func TestStatusEngine_DetachedHEAD(t *testing.T) {
	repoPath := t.TempDir()
	repo, err := gogit.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	// Create initial commit
	if err := os.WriteFile(filepath.Join(repoPath, "readme.md"), []byte("# Test"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := wt.Add("readme.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	hash, err := wt.Commit("Initial commit", &gogit.CommitOptions{})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Checkout specific commit (detached HEAD)
	if err := wt.Checkout(&gogit.CheckoutOptions{
		Hash: hash,
	}); err != nil {
		t.Fatalf("checkout commit: %v", err)
	}

	gitRepo := gittagit.NewRepository()
	ctx := context.Background()
	branchList, err := gitRepo.GetBranchList(ctx, repoPath)
	if err != nil {
		t.Fatalf("GetBranchList: %v", err)
	}

	engine := services.NewStatusEngineWithRepository(gitRepo)
	story := &core.Story{
		ID:    "US-001",
		Title: "Test story",
	}

	// Should handle detached HEAD gracefully (return Todo if no matching branch)
	status, err := engine.DeriveStatus(ctx, story, branchList, repoPath)
	if err != nil {
		t.Fatalf("DeriveStatus failed: %v", err)
	}
	if status != core.StatusTodo {
		t.Errorf("Expected StatusTodo for story with no branch in detached HEAD state, got %v", status)
	}
}
