package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	ggit "github.com/go-git/go-git/v5"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/services"
)

func TestStartCommand_CreatesBranchAndChecksOut(t *testing.T) {
	repoPath := setupRepo(t)
	storyPath := filepath.Join(repoPath, "sprints", "Sprint-01", "US-010.md")
	writeStory(t, storyPath, "US-010", "Start task")
	commitFileToRepo(t, repoPath, storyPath, "add story")

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := git.NewRepository()
	parser := filesystem.NewMarkdownParser()
	svc := services.NewStartService(storyRepo, gitRepo, parser)

	story, branch, err := svc.Start(context.Background(), repoPath, "US-010", nil)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if story.ID != "US-010" {
		t.Fatalf("expected story ID US-010, got %s", story.ID)
	}
	if branch != "feat/US-010" {
		t.Fatalf("expected branch feat/US-010, got %s", branch)
	}

	repo, err := ggit.PlainOpen(repoPath)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	head, err := repo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if head.Name().Short() != "feat/US-010" {
		t.Fatalf("expected head feat/US-010, got %s", head.Name().Short())
	}
}

func TestStartCommand_AssigneeUpdate(t *testing.T) {
	repoPath := setupRepo(t)
	storyPath := filepath.Join(repoPath, "backlog", "US-011.md")
	writeStory(t, storyPath, "US-011", "Backlog task")
	commitFileToRepo(t, repoPath, storyPath, "add backlog story")

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := git.NewRepository()
	parser := filesystem.NewMarkdownParser()
	svc := services.NewStartService(storyRepo, gitRepo, parser)

	assignee := "bob"
	story, _, err := svc.Start(context.Background(), repoPath, storyPath, &assignee)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if story.Assignee == nil || *story.Assignee != "bob" {
		t.Fatalf("expected assignee bob, got %v", story.Assignee)
	}

	data, err := os.ReadFile(storyPath)
	if err != nil {
		t.Fatalf("read story: %v", err)
	}
	if !containsAll(string(data), []string{"assignee: bob"}) {
		t.Fatalf("expected assignee persisted, content: %s", string(data))
	}
}

func commitFileToRepo(t *testing.T, repoPath, filePath, message string) {
	t.Helper()
	repo, err := ggit.PlainOpen(repoPath)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	rel, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}
	if _, err := wt.Add(rel); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := wt.Commit(message, testCommitOptions()); err != nil {
		t.Fatalf("commit: %v", err)
	}
}
