package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ggit "github.com/go-git/go-git/v5"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/services"
	"github.com/gavin/gitta/pkg/ui"
)

func TestListCommand_DefaultSprint(t *testing.T) {
	repoPath := setupRepo(t)
	writeStory(t, filepath.Join(repoPath, "sprints", "Sprint-01", "US-001.md"), "US-001", "Sprint task")

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := git.NewRepository()
	svc := services.NewListService(storyRepo, gitRepo)

	stories, err := svc.ListSprintTasks(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stories) != 1 {
		t.Fatalf("expected 1 sprint story, got %d", len(stories))
	}
}

func TestListCommand_All_IncludesBacklog(t *testing.T) {
	repoPath := setupRepo(t)
	writeStory(t, filepath.Join(repoPath, "sprints", "Sprint-01", "US-001.md"), "US-001", "Sprint task")
	writeStory(t, filepath.Join(repoPath, "backlog", "BL-001.md"), "BL-001", "Backlog task")

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := git.NewRepository()
	svc := services.NewListService(storyRepo, gitRepo)

	sprintStories, backlogStories, err := svc.ListAllTasks(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sprintStories) != 1 || len(backlogStories) != 1 {
		t.Fatalf("expected 1 sprint and 1 backlog story, got %d and %d", len(sprintStories), len(backlogStories))
	}
}

func TestListCommand_All_BacklogOnly(t *testing.T) {
	repoPath := setupRepo(t)
	writeStory(t, filepath.Join(repoPath, "backlog", "BL-002.md"), "BL-002", "Backlog only")

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := git.NewRepository()
	svc := services.NewListService(storyRepo, gitRepo)

	sprintStories, backlogStories, err := svc.ListAllTasks(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sprintStories) != 0 {
		t.Fatalf("expected no sprint stories, got %d", len(sprintStories))
	}
	if len(backlogStories) != 1 {
		t.Fatalf("expected backlog story, got %d", len(backlogStories))
	}
}

func TestListCommand_FormatSnapshot(t *testing.T) {
	repoPath := setupRepo(t)
	writeStory(t, filepath.Join(repoPath, "sprints", "Sprint-01", "US-001.md"), "US-001", "Sprint task")
	writeStory(t, filepath.Join(repoPath, "backlog", "BL-001.md"), "BL-001", "Backlog task")

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := git.NewRepository()
	svc := services.NewListService(storyRepo, gitRepo)

	sprintStories, backlogStories, err := svc.ListAllTasks(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sections := map[string][]ui.DisplayStory{
		"Sprint":  toDisplayStoriesForTest(sprintStories),
		"Backlog": toDisplayStoriesForTest(backlogStories),
	}
	output := ui.RenderStorySections(sections)
	if output == "" {
		t.Fatalf("expected formatted output, got empty string")
	}
	if !containsAll(output, []string{"Sprint", "Backlog", "US-001", "BL-001", "╭"}) {
		t.Fatalf("formatted output missing expected markers: %s", output)
	}
}

func TestListCommand_PerformanceSanity(t *testing.T) {
	repoPath := setupRepo(t)
	// create 120 backlog stories
	for i := 0; i < 120; i++ {
		writeStory(t, filepath.Join(repoPath, "backlog", fmt.Sprintf("BL-%03d.md", i)), fmt.Sprintf("BL-%03d", i), "Backlog task")
	}

	storyRepo := filesystem.NewDefaultRepository()
	gitRepo := git.NewRepository()
	svc := services.NewListService(storyRepo, gitRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sprintStories, backlogStories, err := svc.ListAllTasks(ctx, repoPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sprintStories) != 0 {
		t.Fatalf("expected no sprint stories, got %d", len(sprintStories))
	}
	if len(backlogStories) != 120 {
		t.Fatalf("expected 120 backlog stories, got %d", len(backlogStories))
	}
}

func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := ggit.PlainInit(dir, false); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	return dir
}

func writeStory(t *testing.T, path, id, title string) {
	t.Helper()
	content := `---
id: ` + id + `
title: ` + title + `
---

Body
`
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write story: %v", err)
	}
}

func toDisplayStoriesForTest(stories []*services.StoryWithStatus) []ui.DisplayStory {
	display := make([]ui.DisplayStory, 0, len(stories))
	for _, s := range stories {
		display = append(display, ui.DisplayStory{
			Source:   s.Source,
			Story:    s.Story,
			Priority: s.Story.Priority,
			Status:   s.Status,
		})
	}
	return display
}

func containsAll(haystack string, needles []string) bool {
	for _, n := range needles {
		if !strings.Contains(haystack, n) {
			return false
		}
	}
	return true
}
