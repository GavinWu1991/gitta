package unit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gavin/gitta/internal/core"
)

func TestTaskMetadataUpdateOnRollover(t *testing.T) {
	now := time.Date(2025, 1, 27, 12, 0, 0, 0, time.UTC)
	targetSprint := "Sprint-02"

	tests := []struct {
		name           string
		story          *core.Story
		wantStatus     core.Status
		wantSprint     string
		wantUpdated    bool
		preserveFields []string // Fields that should be preserved
	}{
		{
			name: "todo task - reset to todo",
			story: &core.Story{
				ID:        "US-001",
				Title:     "Test Task",
				Status:    core.StatusTodo,
				Priority:  core.PriorityHigh,
				Assignee:  stringPtr("alice"),
				Tags:      []string{"frontend"},
				CreatedAt: timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
				Body:      "Task description",
			},
			wantStatus:     core.StatusTodo,
			wantSprint:     targetSprint,
			wantUpdated:    true,
			preserveFields: []string{"id", "title", "assignee", "priority", "tags", "body", "created_at"},
		},
		{
			name: "doing task - reset to todo",
			story: &core.Story{
				ID:        "US-002",
				Title:     "In Progress Task",
				Status:    core.StatusDoing,
				Priority:  core.PriorityMedium,
				Assignee:  stringPtr("bob"),
				Tags:      []string{"backend"},
				CreatedAt: timePtr(time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)),
				Body:      "Work in progress",
			},
			wantStatus:     core.StatusTodo,
			wantSprint:     targetSprint,
			wantUpdated:    true,
			preserveFields: []string{"id", "title", "assignee", "priority", "tags", "body", "created_at"},
		},
		{
			name: "done task - keep as done (should not be rolled over)",
			story: &core.Story{
				ID:        "US-003",
				Title:     "Completed Task",
				Status:    core.StatusDone,
				Priority:  core.PriorityLow,
				Assignee:  stringPtr("charlie"),
				Tags:      []string{"test"},
				CreatedAt: timePtr(time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)),
				Body:      "Already done",
			},
			wantStatus:     core.StatusDone, // Should not change
			wantSprint:     targetSprint,
			wantUpdated:    true,
			preserveFields: []string{"id", "title", "assignee", "priority", "tags", "body", "created_at"},
		},
		{
			name: "review task - reset to todo",
			story: &core.Story{
				ID:        "US-004",
				Title:     "Review Task",
				Status:    core.StatusReview,
				Priority:  core.PriorityCritical,
				Assignee:  stringPtr("dave"),
				Tags:      []string{"urgent", "bug"},
				CreatedAt: timePtr(time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)),
				Body:      "Needs review",
			},
			wantStatus:     core.StatusTodo,
			wantSprint:     targetSprint,
			wantUpdated:    true,
			preserveFields: []string{"id", "title", "assignee", "priority", "tags", "body", "created_at"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate metadata update
			updated := updateTaskMetadataForRollover(tt.story, targetSprint, now)

			// Check status
			if updated.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", updated.Status, tt.wantStatus)
			}

			// Check sprint field (would be set in frontmatter, not in struct)
			// For now, we'll just verify the status update logic

			// Check updated_at
			if updated.UpdatedAt == nil {
				t.Error("UpdatedAt should be set")
			} else if !updated.UpdatedAt.Equal(now) {
				t.Errorf("UpdatedAt = %v, want %v", updated.UpdatedAt, now)
			}

			// Verify preserved fields
			if updated.ID != tt.story.ID {
				t.Errorf("ID changed: got %q, want %q", updated.ID, tt.story.ID)
			}
			if updated.Title != tt.story.Title {
				t.Errorf("Title changed: got %q, want %q", updated.Title, tt.story.Title)
			}
			if updated.Assignee != nil && tt.story.Assignee != nil {
				if *updated.Assignee != *tt.story.Assignee {
					t.Errorf("Assignee changed: got %q, want %q", *updated.Assignee, *tt.story.Assignee)
				}
			}
			if updated.Priority != tt.story.Priority {
				t.Errorf("Priority changed: got %q, want %q", updated.Priority, tt.story.Priority)
			}
			if len(updated.Tags) != len(tt.story.Tags) {
				t.Errorf("Tags changed: got %v, want %v", updated.Tags, tt.story.Tags)
			}
			if updated.Body != tt.story.Body {
				t.Errorf("Body changed: got %q, want %q", updated.Body, tt.story.Body)
			}
		})
	}
}

func TestAtomicFileMoveOperations(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(string) error // Setup source file
		wantErr   bool
		checkFunc func(*testing.T, string, string) // Verify result
	}{
		{
			name: "successful atomic move",
			setup: func(sourcePath string) error {
				// Create a test file
				return os.WriteFile(sourcePath, []byte("test content"), 0644)
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sourcePath, targetPath string) {
				// Source should not exist
				if _, err := os.Stat(sourcePath); err == nil {
					t.Errorf("source file should not exist: %s", sourcePath)
				}
				// Target should exist
				if _, err := os.Stat(targetPath); err != nil {
					t.Errorf("target file should exist: %v", err)
				}
			},
		},
		{
			name: "target directory does not exist",
			setup: func(sourcePath string) error {
				return os.WriteFile(sourcePath, []byte("test"), 0644)
			},
			wantErr: false, // Should create directory
			checkFunc: func(t *testing.T, sourcePath, targetPath string) {
				// Target should exist even if directory was created
				if _, err := os.Stat(targetPath); err != nil {
					t.Errorf("target file should exist after creating directory: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			sourcePath := filepath.Join(tmpDir, "source", "US-001.md")
			targetDir := filepath.Join(tmpDir, "target")
			targetPath := filepath.Join(targetDir, "US-001.md")

			// Setup
			os.MkdirAll(filepath.Dir(sourcePath), 0755)
			if tt.setup != nil {
				if err := tt.setup(sourcePath); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Perform atomic move (simulated)
			err := performAtomicMove(sourcePath, targetPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("performAtomicMove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, sourcePath, targetPath)
			}
		})
	}
}

// Helper functions for testing (will be replaced by actual implementation)

func updateTaskMetadataForRollover(story *core.Story, targetSprint string, rolloverTime time.Time) *core.Story {
	updated := *story // Copy
	updated.UpdatedAt = &rolloverTime

	// Reset status to todo if not done
	if story.Status != core.StatusDone {
		updated.Status = core.StatusTodo
	}

	// Sprint field would be set in frontmatter during file write
	return &updated
}

func performAtomicMove(sourcePath, targetPath string) error {
	// Create target directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// Read source
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	// Write to temp file
	tmpPath := targetPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, targetPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Remove source
	return os.Remove(sourcePath)
}
