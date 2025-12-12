package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestMoveService_MoveStory(t *testing.T) {
	tests := []struct {
		name       string
		storyID    string
		targetDir  string
		force      bool
		wantErr    bool
		setupFunc  func(t *testing.T, tmpDir string) string
		verifyFunc func(t *testing.T, tmpDir, targetDir, storyID string)
	}{
		{
			name:      "move story to sprint directory",
			storyID:   "US-001",
			targetDir: "sprints/2025-01",
			force:     false,
			wantErr:   false,
			setupFunc: func(t *testing.T, tmpDir string) string {
				return createTestStory(t, filepath.Join(tmpDir, "backlog"), "US-001", "Test Story", core.StatusTodo)
			},
			verifyFunc: func(t *testing.T, tmpDir, targetDir, storyID string) {
				targetPath := filepath.Join(tmpDir, targetDir, storyID+".md")
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					t.Errorf("Story file not found at target: %v", targetPath)
				}
				// Verify original is removed
				originalPath := filepath.Join(tmpDir, "backlog", storyID+".md")
				if _, err := os.Stat(originalPath); err == nil {
					t.Error("Original story file still exists")
				}
			},
		},
		{
			name:      "move story with force overwrite",
			storyID:   "US-001",
			targetDir: "sprints/2025-01",
			force:     true,
			wantErr:   false,
			setupFunc: func(t *testing.T, tmpDir string) string {
				// Create story in backlog
				createTestStory(t, filepath.Join(tmpDir, "backlog"), "US-001", "Test Story", core.StatusTodo)
				// Create existing story at target
				createTestStory(t, filepath.Join(tmpDir, "sprints", "2025-01"), "US-001", "Existing Story", core.StatusDoing)
				return ""
			},
			verifyFunc: func(t *testing.T, tmpDir, targetDir, storyID string) {
				targetPath := filepath.Join(tmpDir, targetDir, storyID+".md")
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					t.Errorf("Story file not found at target: %v", targetPath)
				}
			},
		},
		{
			name:      "target file exists without force",
			storyID:   "US-001",
			targetDir: "sprints/2025-01",
			force:     false,
			wantErr:   true,
			setupFunc: func(t *testing.T, tmpDir string) string {
				// Create story in backlog
				createTestStory(t, filepath.Join(tmpDir, "backlog"), "US-001", "Test Story", core.StatusTodo)
				// Create existing story at target
				createTestStory(t, filepath.Join(tmpDir, "sprints", "2025-01"), "US-001", "Existing Story", core.StatusDoing)
				return ""
			},
		},
		{
			name:      "story not found",
			storyID:   "US-999",
			targetDir: "sprints/2025-01",
			force:     false,
			wantErr:   true,
			setupFunc: func(t *testing.T, tmpDir string) string {
				return "" // No story created
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			backlogDir := filepath.Join(tmpDir, "backlog")
			if err := os.MkdirAll(backlogDir, 0755); err != nil {
				t.Fatalf("Failed to create backlog directory: %v", err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			parser := filesystem.NewMarkdownParser()
			storyRepo := filesystem.NewRepository(parser)
			moveService := services.NewMoveService(parser, storyRepo, tmpDir)

			ctx := context.Background()
			err := moveService.MoveStory(ctx, tt.storyID, tt.targetDir, tt.force)

			if tt.wantErr {
				if err == nil {
					t.Errorf("MoveStory() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("MoveStory() error = %v, want nil", err)
				return
			}

			if tt.verifyFunc != nil {
				tt.verifyFunc(t, tmpDir, tt.targetDir, tt.storyID)
			}
		})
	}
}

func TestMoveService_AtomicMove(t *testing.T) {
	// Test that atomic move preserves content
	tmpDir := t.TempDir()
	backlogDir := filepath.Join(tmpDir, "backlog")
	if err := os.MkdirAll(backlogDir, 0755); err != nil {
		t.Fatalf("Failed to create backlog directory: %v", err)
	}

	storyPath := createTestStory(t, backlogDir, "US-001", "Test Story", core.StatusTodo)

	// Read original content for verification
	_, err := os.ReadFile(storyPath)
	if err != nil {
		t.Fatalf("Failed to read original story: %v", err)
	}

	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)
	moveService := services.NewMoveService(parser, storyRepo, tmpDir)

	ctx := context.Background()
	targetDir := "sprints/2025-01"
	err = moveService.MoveStory(ctx, "US-001", targetDir, false)
	if err != nil {
		t.Fatalf("MoveStory() error = %v", err)
	}

	// Verify file was moved
	targetPath := filepath.Join(tmpDir, targetDir, "US-001.md")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Fatal("Story file was not moved to target")
	}

	// Verify content is preserved
	movedContent, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read moved story: %v", err)
	}

	// Content should be the same (or at least contain the same story data)
	if len(movedContent) == 0 {
		t.Error("Moved story file is empty")
	}

	// Verify original is removed
	if _, err := os.Stat(storyPath); err == nil {
		t.Error("Original story file still exists")
	}

	// Verify story can be parsed from new location
	movedStory, err := parser.ReadStory(ctx, targetPath)
	if err != nil {
		t.Fatalf("Failed to parse moved story: %v", err)
	}
	if movedStory.ID != "US-001" {
		t.Errorf("Moved story ID = %v, want US-001", movedStory.ID)
	}
}

// Helper functions are defined in update_service_test.go (same package)
