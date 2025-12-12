package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/services"
)

func TestMoveCommand_BasicMove(t *testing.T) {
	repoPath := setupRepo(t)
	writeStory(t, filepath.Join(repoPath, "backlog", "US-001.md"), "US-001", "Test Story")

	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)
	moveService := services.NewMoveService(parser, storyRepo, repoPath)

	ctx := context.Background()
	err := moveService.MoveStory(ctx, "US-001", "sprints/2025-01", false)
	if err != nil {
		t.Fatalf("MoveStory() error = %v", err)
	}

	// Verify file was moved
	targetPath := filepath.Join(repoPath, "sprints", "2025-01", "US-001.md")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Story file not found at target: %v", targetPath)
	}

	// Verify original is removed
	originalPath := filepath.Join(repoPath, "backlog", "US-001.md")
	if _, err := os.Stat(originalPath); err == nil {
		t.Error("Original story file still exists")
	}
}

func TestMoveCommand_StoryNotFound(t *testing.T) {
	repoPath := setupRepo(t)

	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)
	moveService := services.NewMoveService(parser, storyRepo, repoPath)

	ctx := context.Background()
	err := moveService.MoveStory(ctx, "US-999", "sprints/2025-01", false)
	if err == nil {
		t.Error("MoveStory() expected error for non-existent story")
	}
}
