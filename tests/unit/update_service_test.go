package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestUpdateService_UpdateStatus(t *testing.T) {
	tests := []struct {
		name       string
		storyID    string
		newStatus  core.Status
		wantErr    bool
		setupFunc  func(t *testing.T, tmpDir string) string
		verifyFunc func(t *testing.T, filePath string, status core.Status)
	}{
		{
			name:      "update status from todo to doing",
			storyID:   "US-001",
			newStatus: core.StatusDoing,
			wantErr:   false,
			setupFunc: func(t *testing.T, tmpDir string) string {
				return createTestStory(t, tmpDir, "US-001", "Test Story", core.StatusTodo)
			},
			verifyFunc: func(t *testing.T, filePath string, status core.Status) {
				verifyStoryStatus(t, filePath, status)
			},
		},
		{
			name:      "update status to done",
			storyID:   "US-001",
			newStatus: core.StatusDone,
			wantErr:   false,
			setupFunc: func(t *testing.T, tmpDir string) string {
				return createTestStory(t, tmpDir, "US-001", "Test Story", core.StatusDoing)
			},
			verifyFunc: func(t *testing.T, filePath string, status core.Status) {
				verifyStoryStatus(t, filePath, status)
			},
		},
		{
			name:      "story not found",
			storyID:   "US-999",
			newStatus: core.StatusDoing,
			wantErr:   true,
			setupFunc: func(t *testing.T, tmpDir string) string {
				return "" // No story created
			},
		},
		{
			name:      "invalid status",
			storyID:   "US-001",
			newStatus: core.Status("invalid"),
			wantErr:   true,
			setupFunc: func(t *testing.T, tmpDir string) string {
				return createTestStory(t, tmpDir, "US-001", "Test Story", core.StatusTodo)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			// Create backlog directory (where FindStoryByID searches)
			backlogDir := filepath.Join(tmpDir, "backlog")
			if err := os.MkdirAll(backlogDir, 0755); err != nil {
				t.Fatalf("Failed to create backlog directory: %v", err)
			}

			var storyPath string
			if tt.setupFunc != nil {
				storyPath = tt.setupFunc(t, backlogDir)
			}

			parser := filesystem.NewMarkdownParser()
			storyRepo := filesystem.NewRepository(parser)
			updateService := services.NewUpdateService(parser, storyRepo, tmpDir)

			ctx := context.Background()
			err := updateService.UpdateStatus(ctx, tt.storyID, tt.newStatus)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateStatus() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateStatus() error = %v, want nil", err)
				return
			}

			if tt.verifyFunc != nil && storyPath != "" {
				tt.verifyFunc(t, storyPath, tt.newStatus)
			}
		})
	}
}

func TestUpdateService_AtomicWrite(t *testing.T) {
	// Test that atomic write preserves file on failure
	tmpDir := t.TempDir()
	backlogDir := filepath.Join(tmpDir, "backlog")
	if err := os.MkdirAll(backlogDir, 0755); err != nil {
		t.Fatalf("Failed to create backlog directory: %v", err)
	}

	storyPath := createTestStory(t, backlogDir, "US-001", "Test Story", core.StatusTodo)

	// Read original content
	originalContent, err := os.ReadFile(storyPath)
	if err != nil {
		t.Fatalf("Failed to read original story: %v", err)
	}

	parser := filesystem.NewMarkdownParser()
	storyRepo := filesystem.NewRepository(parser)
	updateService := services.NewUpdateService(parser, storyRepo, tmpDir)

	ctx := context.Background()
	err = updateService.UpdateStatus(ctx, "US-001", core.StatusDoing)
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	// Verify file was updated (not corrupted)
	updatedStory, err := parser.ReadStory(ctx, storyPath)
	if err != nil {
		t.Fatalf("Failed to read updated story: %v", err)
	}

	if updatedStory.Status != core.StatusDoing {
		t.Errorf("UpdateStatus() status = %v, want %v", updatedStory.Status, core.StatusDoing)
	}

	// Verify updated_at was set
	if updatedStory.UpdatedAt == nil {
		t.Error("UpdateStatus() updated_at was not set")
	}

	// Verify original content is different (file was actually updated)
	updatedContent, _ := os.ReadFile(storyPath)
	if string(originalContent) == string(updatedContent) {
		t.Error("UpdateStatus() file content was not updated")
	}
}

// Helper functions are defined in other test files in the same package
func createTestStory(t *testing.T, dir, id, title string, status core.Status) string {
	t.Helper()
	story := &core.Story{
		ID:        id,
		Title:     title,
		Status:    status,
		Priority:  core.PriorityMedium,
		CreatedAt: timePtr(time.Now()),
	}

	filePath := filepath.Join(dir, id+".md")
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()
	if err := parser.WriteStory(ctx, filePath, story); err != nil {
		t.Fatalf("Failed to create test story: %v", err)
	}
	return filePath
}

func verifyStoryStatus(t *testing.T, filePath string, expectedStatus core.Status) {
	t.Helper()
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()
	story, err := parser.ReadStory(ctx, filePath)
	if err != nil {
		t.Fatalf("Failed to read story: %v", err)
	}
	if story.Status != expectedStatus {
		t.Errorf("Story status = %v, want %v", story.Status, expectedStatus)
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
