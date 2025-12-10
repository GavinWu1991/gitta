package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
