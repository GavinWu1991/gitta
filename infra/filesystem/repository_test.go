package filesystem

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gavin/gitta/internal/core"
)

func TestFindCurrentSprint_SelectsHighestLexicographically(t *testing.T) {
	dir := t.TempDir()
	sprintsDir := filepath.Join(dir, "sprints")
	requireNoError(t, os.MkdirAll(filepath.Join(sprintsDir, "Sprint-01"), 0o755))
	requireNoError(t, os.MkdirAll(filepath.Join(sprintsDir, "sprint-03"), 0o755))
	requireNoError(t, os.MkdirAll(filepath.Join(sprintsDir, "Sprint-02"), 0o755))

	repo := NewDefaultRepository()
	sprint, err := repo.FindCurrentSprint(context.Background(), sprintsDir)
	if err != nil {
		t.Fatalf("expected sprint dir, got error: %v", err)
	}

	expected := filepath.Join(sprintsDir, "sprint-03")
	if sprint != expected {
		t.Fatalf("expected %s, got %s", expected, sprint)
	}
}

func TestFindCurrentSprint_NoSprints(t *testing.T) {
	dir := t.TempDir()
	repo := NewDefaultRepository()
	_, err := repo.FindCurrentSprint(context.Background(), dir)
	if err == nil {
		t.Fatalf("expected error when no Sprint directories found")
	}
}

func TestListStories_ReturnsStories(t *testing.T) {
	dir := t.TempDir()
	storiesDir := filepath.Join(dir, "sprints", "Sprint-01")
	requireNoError(t, os.MkdirAll(storiesDir, 0o755))

	writeStory(t, filepath.Join(storiesDir, "US-001.md"), "US-001", "Title 1")
	writeStory(t, filepath.Join(storiesDir, "US-002.md"), "US-002", "Title 2")

	repo := NewDefaultRepository()
	stories, err := repo.ListStories(context.Background(), storiesDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stories) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(stories))
	}
}

func TestListStories_MissingDirectoryReturnsEmpty(t *testing.T) {
	repo := NewDefaultRepository()
	stories, err := repo.ListStories(context.Background(), "/non/existent/path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(stories) != 0 {
		t.Fatalf("expected empty slice, got %d", len(stories))
	}
}

func TestListStories_ContextCancellation(t *testing.T) {
	dir := t.TempDir()
	storiesDir := filepath.Join(dir, "sprints", "Sprint-01")
	requireNoError(t, os.MkdirAll(storiesDir, 0o755))
	writeStory(t, filepath.Join(storiesDir, "US-001.md"), "US-001", "Title 1")

	repo := NewDefaultRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := repo.ListStories(ctx, storiesDir)
	if err == nil {
		t.Fatalf("expected context cancellation error")
	}
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

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListStories_SkipsInvalidFilesButReturnsError(t *testing.T) {
	dir := t.TempDir()
	storiesDir := filepath.Join(dir, "sprints", "Sprint-01")
	requireNoError(t, os.MkdirAll(storiesDir, 0o755))

	// Valid story
	writeStory(t, filepath.Join(storiesDir, "US-001.md"), "US-001", "Title 1")
	// Invalid story (malformed YAML frontmatter)
	invalid := `---
id: [missing-closing
`
	if err := os.WriteFile(filepath.Join(storiesDir, "US-002.md"), []byte(invalid), 0o644); err != nil {
		t.Fatalf("failed to write invalid story: %v", err)
	}

	repo := NewDefaultRepository()
	stories, err := repo.ListStories(context.Background(), storiesDir)
	if len(stories) != 1 {
		t.Fatalf("expected 1 valid story, got %d", len(stories))
	}
	if err == nil {
		t.Fatalf("expected error due to invalid story")
	}
}

func TestFindCurrentSprint_RespectsContext(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := NewDefaultRepository()
	_, err := repo.FindCurrentSprint(ctx, dir)
	if err == nil {
		t.Fatalf("expected context cancellation error")
	}
}

// Ensure default parser applies defaults for missing optional fields.
func TestListStories_AppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	storiesDir := filepath.Join(dir, "sprints", "Sprint-01")
	requireNoError(t, os.MkdirAll(storiesDir, 0o755))
	writeStory(t, filepath.Join(storiesDir, "US-010.md"), "US-010", "Title 10")

	repo := NewDefaultRepository()
	stories, err := repo.ListStories(context.Background(), storiesDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stories) != 1 {
		t.Fatalf("expected 1 story, got %d", len(stories))
	}
	if stories[0].Priority != core.PriorityMedium {
		t.Fatalf("expected default priority, got %s", stories[0].Priority)
	}
	if stories[0].Status != core.StatusTodo {
		t.Fatalf("expected default status, got %s", stories[0].Status)
	}
	if stories[0].Tags == nil {
		t.Fatalf("expected tags slice initialized")
	}
}

// Small timeout to ensure context is checked regularly.
func TestListStories_ContextCheckedDuringScan(t *testing.T) {
	dir := t.TempDir()
	storiesDir := filepath.Join(dir, "sprints", "Sprint-01")
	requireNoError(t, os.MkdirAll(storiesDir, 0o755))

	// Create several files to iterate over
	for i := 0; i < 5; i++ {
		writeStory(t, filepath.Join(storiesDir, fmt.Sprintf("US-%03d.md", i)), fmt.Sprintf("US-%03d", i), "Title")
	}

	repo := NewDefaultRepository()
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Nanosecond)

	_, err := repo.ListStories(ctx, storiesDir)
	if err == nil {
		t.Fatalf("expected context cancellation or deadline exceeded")
	}
}

func TestFindStoryByID_SearchesSprintThenBacklog(t *testing.T) {
	dir := t.TempDir()
	repoPath := dir

	// Backlog story only; Sprint directory exists but without target story.
	writeStory(t, filepath.Join(repoPath, "sprints", "Sprint-01", "US-100.md"), "US-100", "Other task")
	writeStory(t, filepath.Join(repoPath, "backlog", "US-200.md"), "US-200", "Target task")

	repo := NewDefaultRepository()
	story, path, err := repo.FindStoryByID(context.Background(), repoPath, "US-200")
	if err != nil {
		t.Fatalf("expected story, got error: %v", err)
	}
	if story.ID != "US-200" {
		t.Fatalf("expected story ID US-200, got %s", story.ID)
	}
	if !strings.HasSuffix(path, filepath.Join("backlog", "US-200.md")) {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestFindStoryByID_NotFound(t *testing.T) {
	repo := NewDefaultRepository()
	_, _, err := repo.FindStoryByID(context.Background(), t.TempDir(), "US-999")
	if err == nil {
		t.Fatalf("expected ErrStoryNotFound")
	}
	if !errors.Is(err, core.ErrStoryNotFound) {
		t.Fatalf("expected ErrStoryNotFound, got %v", err)
	}
}

func TestFindStoryByPath_ValidatesMarkdown(t *testing.T) {
	dir := t.TempDir()
	repo := NewDefaultRepository()

	// Non-existent
	if _, err := repo.FindStoryByPath(context.Background(), filepath.Join(dir, "missing.md")); !errors.Is(err, core.ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath for missing file, got %v", err)
	}

	// Not markdown
	txtPath := filepath.Join(dir, "file.txt")
	requireNoError(t, os.WriteFile(txtPath, []byte("hi"), 0o644))
	if _, err := repo.FindStoryByPath(context.Background(), txtPath); !errors.Is(err, core.ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath for non-markdown, got %v", err)
	}

	// Valid
	mdPath := filepath.Join(dir, "task.md")
	writeStory(t, mdPath, "US-300", "Valid story")
	story, err := repo.FindStoryByPath(context.Background(), mdPath)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if story.ID != "US-300" {
		t.Fatalf("expected ID US-300, got %s", story.ID)
	}
}
