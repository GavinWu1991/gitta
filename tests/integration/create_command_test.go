package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestCreateCommand_BasicCreation(t *testing.T) {
	repoPath := setupRepo(t)
	storyDir := filepath.Join(repoPath, "backlog")

	// Create service dependencies
	idGenerator := filesystem.NewIDCounter(repoPath)
	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)
	createService := services.NewCreateService(idGenerator, parser, storyRepo, storyDir)

	req := services.CreateStoryRequest{
		Title:    "Test Story",
		Prefix:   "US",
		Status:   core.StatusTodo,
		Priority: core.PriorityMedium,
	}

	ctx := context.Background()
	story, filePath, err := createService.CreateStory(ctx, req)
	if err != nil {
		t.Fatalf("CreateStory() error = %v, want nil", err)
	}

	// Verify story was created
	if story == nil {
		t.Fatal("CreateStory() returned nil story")
	}

	// Verify ID format
	if !strings.HasPrefix(story.ID, "US-") {
		t.Errorf("CreateStory() ID = %v, want prefix US-", story.ID)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("CreateStory() file does not exist: %v", filePath)
	}

	// Verify file can be read back
	readStory, err := parser.ReadStory(ctx, filePath)
	if err != nil {
		t.Errorf("CreateStory() created file cannot be read: %v", err)
	} else if readStory.ID != story.ID {
		t.Errorf("CreateStory() read back ID = %v, want %v", readStory.ID, story.ID)
	}
}

func TestCreateCommand_WithEditor(t *testing.T) {
	repoPath := setupRepo(t)
	storyDir := filepath.Join(repoPath, "backlog")

	// Create a mock editor script
	mockEditor := filepath.Join(repoPath, "mock-editor.sh")
	editorScript := `#!/bin/sh
# Mock editor that adds a comment
echo "" >> "$1"
echo "# Edited by mock editor" >> "$1"
`
	if err := os.WriteFile(mockEditor, []byte(editorScript), 0755); err != nil {
		t.Fatalf("Failed to create mock editor: %v", err)
	}

	idGenerator := filesystem.NewIDCounter(repoPath)
	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)
	createService := services.NewCreateService(idGenerator, parser, storyRepo, storyDir)

	req := services.CreateStoryRequest{
		Title:    "Test with Editor",
		Prefix:   "US",
		Editor:   mockEditor,
		Status:   core.StatusTodo,
		Priority: core.PriorityMedium,
	}

	ctx := context.Background()
	story, filePath, err := createService.CreateStory(ctx, req)
	if err != nil {
		t.Fatalf("CreateStory() with editor error = %v", err)
	}

	// Verify story was created
	if story == nil {
		t.Fatal("Story is nil")
	}

	// Verify editor modified the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read story file: %v", err)
	}
	if !strings.Contains(string(content), "Edited by mock editor") {
		t.Error("Editor did not modify the file")
	}
}

func TestCreateCommand_ConcurrentCreation(t *testing.T) {
	repoPath := setupRepo(t)
	storyDir := filepath.Join(repoPath, "backlog")

	idGenerator := filesystem.NewIDCounter(repoPath)
	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)

	// Create multiple stories concurrently
	const numStories = 5
	results := make(chan struct {
		story *services.CreateStoryRequest
		id    string
		err   error
	}, numStories)

	ctx := context.Background()
	for i := 0; i < numStories; i++ {
		go func(index int) {
			createService := services.NewCreateService(idGenerator, parser, storyRepo, storyDir)
			req := services.CreateStoryRequest{
				Title:    fmt.Sprintf("Concurrent Story %d", index),
				Prefix:   "US",
				Status:   core.StatusTodo,
				Priority: core.PriorityMedium,
			}
			story, _, err := createService.CreateStory(ctx, req)
			results <- struct {
				story *services.CreateStoryRequest
				id    string
				err   error
			}{&req, story.ID, err}
		}(i)
	}

	// Collect results
	ids := make(map[string]bool)
	for i := 0; i < numStories; i++ {
		result := <-results
		if result.err != nil {
			t.Errorf("CreateStory() concurrent error = %v", result.err)
			continue
		}
		if ids[result.id] {
			t.Errorf("Duplicate ID generated: %s", result.id)
		}
		ids[result.id] = true
	}

	// Verify we got unique IDs
	if len(ids) != numStories {
		t.Errorf("Expected %d unique IDs, got %d", numStories, len(ids))
	}
}
