package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestStatusCommand_BasicUpdate(t *testing.T) {
	repoPath := setupRepo(t)
	writeStory(t, filepath.Join(repoPath, "backlog", "US-001.md"), "US-001", "Test Story")

	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)
	updateService := services.NewUpdateService(parser, storyRepo, repoPath)

	ctx := context.Background()
	err := updateService.UpdateStatus(ctx, "US-001", core.StatusDoing)
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	// Verify status was updated
	story, _, err := storyRepo.FindStoryByID(ctx, repoPath, "US-001")
	if err != nil {
		t.Fatalf("Failed to find updated story: %v", err)
	}
	if story.Status != core.StatusDoing {
		t.Errorf("Status = %v, want %v", story.Status, core.StatusDoing)
	}
	if story.UpdatedAt == nil {
		t.Error("UpdatedAt was not set")
	}
}

func TestStatusCommand_StoryNotFound(t *testing.T) {
	repoPath := setupRepo(t)

	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)
	updateService := services.NewUpdateService(parser, storyRepo, repoPath)

	ctx := context.Background()
	err := updateService.UpdateStatus(ctx, "US-999", core.StatusDoing)
	if err == nil {
		t.Error("UpdateStatus() expected error for non-existent story")
	}
}
