package services_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ggit "github.com/go-git/go-git/v5"

	"github.com/gavin/gitta/infra/filesystem"
	gitrepo "github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/services"
)

func TestStartService_StartByID_CreatesBranch(t *testing.T) {
	repoPath, repo := setupRepoWithStory(t, "sprints/Sprint-01/US-001.md")

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := gitrepo.NewRepository()
	parser := filesystem.NewMarkdownParser()
	svc := services.NewStartService(storyRepo, gitRepo, parser)

	story, branch, err := svc.Start(context.Background(), repoPath, "US-001", nil)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if story.ID != "US-001" {
		t.Fatalf("expected story ID US-001, got %s", story.ID)
	}
	if branch != "feat/US-001" {
		t.Fatalf("expected branch feat/US-001, got %s", branch)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if head.Name().Short() != "feat/US-001" {
		t.Fatalf("expected head feat/US-001, got %s", head.Name().Short())
	}
}

func TestStartService_StartByPath_AssigneeUpdate(t *testing.T) {
	repoPath, _ := setupRepoWithStory(t, "backlog/US-002.md")

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := gitrepo.NewRepository()
	parser := filesystem.NewMarkdownParser()
	svc := services.NewStartService(storyRepo, gitRepo, parser)

	assignee := "alice"
	story, _, err := svc.Start(context.Background(), repoPath, filepath.Join(repoPath, "backlog", "US-002.md"), &assignee)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if story.Assignee == nil || *story.Assignee != "alice" {
		t.Fatalf("expected assignee alice, got %v", story.Assignee)
	}

	// Verify persisted to file.
	parsed, err := parser.ReadStory(context.Background(), filepath.Join(repoPath, "backlog", "US-002.md"))
	if err != nil {
		t.Fatalf("read story: %v", err)
	}
	if parsed.Assignee == nil || *parsed.Assignee != "alice" {
		t.Fatalf("expected persisted assignee alice, got %v", parsed.Assignee)
	}
}

func TestStartService_UncommittedChanges(t *testing.T) {
	repoPath, repo := setupRepoWithStory(t, "sprints/Sprint-02/US-003.md")

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	// create uncommitted change
	if err := os.WriteFile(filepath.Join(repoPath, "dirty.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}
	if _, err := wt.Add("dirty.txt"); err != nil {
		t.Fatalf("add: %v", err)
	}
	// no commit -> uncommitted change present

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := gitrepo.NewRepository()
	parser := filesystem.NewMarkdownParser()
	svc := services.NewStartService(storyRepo, gitRepo, parser)

	_, _, err = svc.Start(context.Background(), repoPath, "US-003", nil)
	if err == nil {
		t.Fatalf("expected error due to uncommitted changes")
	}
	if !errors.Is(err, gitrepo.ErrUncommittedChanges) {
		t.Fatalf("expected ErrUncommittedChanges, got %v", err)
	}
}

func setupRepoWithStory(t *testing.T, storyRelPath string) (string, *ggit.Repository) {
	t.Helper()
	repoPath := t.TempDir()
	repo, err := ggit.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoPath, "readme.md"), []byte("init"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	if _, err := wt.Add("readme.md"); err != nil {
		t.Fatalf("add readme: %v", err)
	}
	if _, err := wt.Commit("init", &ggit.CommitOptions{}); err != nil {
		t.Fatalf("commit: %v", err)
	}

	storyPath := filepath.Join(repoPath, storyRelPath)
	if err := os.MkdirAll(filepath.Dir(storyPath), 0o755); err != nil {
		t.Fatalf("mkdir story dir: %v", err)
	}
	content := `---
id: ` + storyIDFromPath(storyRelPath) + `
title: Title
---

Body
`
	if err := os.WriteFile(storyPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write story: %v", err)
	}

	wt, err = repo.Worktree()
	if err != nil {
		t.Fatalf("worktree story: %v", err)
	}
	if rel, err := filepath.Rel(repoPath, storyPath); err == nil {
		if _, err := wt.Add(rel); err != nil {
			t.Fatalf("add story: %v", err)
		}
	} else {
		t.Fatalf("rel path: %v", err)
	}
	if _, err := wt.Commit("add story", &ggit.CommitOptions{}); err != nil {
		t.Fatalf("commit story: %v", err)
	}

	return repoPath, repo
}

func storyIDFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
