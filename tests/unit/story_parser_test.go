package unit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
)

func TestReadStory_ValidStory(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	// Get path to test file
	testFile := filepath.Join("..", "..", "testdata", "parser", "valid-story.md")
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	story, err := parser.ReadStory(ctx, absPath)
	if err != nil {
		t.Fatalf("ReadStory failed: %v", err)
	}

	// Verify required fields
	if story.ID != "US-001" {
		t.Errorf("Expected ID 'US-001', got '%s'", story.ID)
	}
	if story.Title != "Implement user authentication" {
		t.Errorf("Expected Title 'Implement user authentication', got '%s'", story.Title)
	}

	// Verify optional fields
	if story.Assignee == nil || *story.Assignee != "alice" {
		t.Errorf("Expected Assignee 'alice', got %v", story.Assignee)
	}
	if story.Priority != core.PriorityHigh {
		t.Errorf("Expected Priority 'high', got '%s'", story.Priority)
	}
	if story.Status != core.StatusDoing {
		t.Errorf("Expected Status 'doing', got '%s'", story.Status)
	}

	// Verify body content
	if story.Body == "" {
		t.Error("Expected non-empty body content")
	}
	if !contains(story.Body, "OAuth2 authentication") {
		t.Errorf("Body should contain 'OAuth2 authentication', got: %s", story.Body)
	}
}

func TestReadStory_MissingFields(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "parser", "missing-fields.md")
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	story, err := parser.ReadStory(ctx, absPath)
	if err != nil {
		t.Fatalf("ReadStory failed: %v", err)
	}

	// Verify required fields are present
	if story.ID != "US-003" {
		t.Errorf("Expected ID 'US-003', got '%s'", story.ID)
	}

	// Verify default values for missing optional fields
	if story.Priority != core.PriorityMedium {
		t.Errorf("Expected default Priority 'medium', got '%s'", story.Priority)
	}
	if story.Status != core.StatusTodo {
		t.Errorf("Expected default Status 'todo', got '%s'", story.Status)
	}
	if story.Assignee != nil {
		t.Errorf("Expected nil Assignee, got %v", story.Assignee)
	}
}

func TestReadStory_NoFrontmatter(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "parser", "no-frontmatter.md")
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	story, err := parser.ReadStory(ctx, absPath)
	// Should handle gracefully - either error or return story with defaults
	if err == nil {
		// If no error, verify defaults are applied
		if story.Priority != core.PriorityMedium {
			t.Errorf("Expected default Priority 'medium', got '%s'", story.Priority)
		}
		if story.Status != core.StatusTodo {
			t.Errorf("Expected default Status 'todo', got '%s'", story.Status)
		}
		if story.Body == "" {
			t.Error("Expected body content to be preserved")
		}
	}
}

func TestReadStory_EmptyFile(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "parser", "empty-file.md")
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	story, err := parser.ReadStory(ctx, absPath)
	// Empty file should either error or return story with empty/default values
	if err == nil {
		if story.ID != "" || story.Title != "" {
			t.Errorf("Empty file should have empty required fields, got ID='%s', Title='%s'", story.ID, story.Title)
		}
	}
}

func TestReadStory_FileNotFound(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	_, err := parser.ReadStory(ctx, "/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	var ioErr *core.IOError
	if !errors.As(err, &ioErr) {
		t.Errorf("Expected IOError, got %T: %v", err, err)
	}
}

func TestWriteStory_NewFile(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-story.md")

	assignee := "test-user"
	story := &core.Story{
		ID:       "TS-001",
		Title:    "Test Story",
		Assignee: &assignee,
		Priority: core.PriorityHigh,
		Status:   core.StatusTodo,
		Body:     "## Description\n\nThis is a test story.",
	}

	err := parser.WriteStory(ctx, testFile, story)
	if err != nil {
		t.Fatalf("WriteStory failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}

	// Read it back and verify
	readStory, err := parser.ReadStory(ctx, testFile)
	if err != nil {
		t.Fatalf("Failed to read written story: %v", err)
	}

	if readStory.ID != story.ID {
		t.Errorf("Expected ID '%s', got '%s'", story.ID, readStory.ID)
	}
	if readStory.Title != story.Title {
		t.Errorf("Expected Title '%s', got '%s'", story.Title, readStory.Title)
	}
	if readStory.Body != story.Body {
		t.Errorf("Expected Body '%s', got '%s'", story.Body, readStory.Body)
	}
}

func TestWriteStory_RoundTrip(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	// Read existing story
	testFile := filepath.Join("..", "..", "testdata", "parser", "valid-story.md")
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	originalStory, err := parser.ReadStory(ctx, absPath)
	if err != nil {
		t.Fatalf("Failed to read original story: %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "roundtrip-test.md")

	err = parser.WriteStory(ctx, tmpFile, originalStory)
	if err != nil {
		t.Fatalf("WriteStory failed: %v", err)
	}

	// Read it back
	readStory, err := parser.ReadStory(ctx, tmpFile)
	if err != nil {
		t.Fatalf("Failed to read written story: %v", err)
	}

	// Verify all fields match
	if readStory.ID != originalStory.ID {
		t.Errorf("ID mismatch: expected '%s', got '%s'", originalStory.ID, readStory.ID)
	}
	if readStory.Title != originalStory.Title {
		t.Errorf("Title mismatch: expected '%s', got '%s'", originalStory.Title, readStory.Title)
	}
	// Normalize line endings for cross-platform compatibility (Windows uses CRLF, Unix uses LF)
	normalizedOriginal := strings.ReplaceAll(originalStory.Body, "\r\n", "\n")
	normalizedRead := strings.ReplaceAll(readStory.Body, "\r\n", "\n")
	if normalizedRead != normalizedOriginal {
		t.Errorf("Body mismatch: expected '%s', got '%s'", normalizedOriginal, normalizedRead)
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestReadStory_ConcurrentReads(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "parser", "valid-story.md")
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Test concurrent reads from multiple goroutines
	const numGoroutines = 10
	const numReadsPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numReadsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numReadsPerGoroutine; j++ {
				story, err := parser.ReadStory(ctx, absPath)
				if err != nil {
					errors <- err
					return
				}
				if story == nil || story.ID != "US-001" {
					errors <- fmt.Errorf("unexpected story content")
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent read error: %v", err)
	}
}

func TestReadStory_LargeFile(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	// Create a large file (approaching 10MB limit)
	tmpDir := t.TempDir()
	largeFile := filepath.Join(tmpDir, "large-story.md")

	// Create frontmatter
	frontmatter := `---
id: LG-001
title: Large Story File
---

`

	// Create large body content (about 9MB to stay under 10MB limit)
	body := strings.Repeat("# Content\n\nThis is a large story file with lots of content.\n\n", 100000)
	content := frontmatter + body

	if err := os.WriteFile(largeFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	// Verify file size is close to limit but under
	info, err := os.Stat(largeFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() > 10*1024*1024 {
		t.Fatalf("Test file too large: %d bytes", info.Size())
	}

	// Read the large file
	story, err := parser.ReadStory(ctx, largeFile)
	if err != nil {
		t.Fatalf("Failed to read large file: %v", err)
	}

	if story.ID != "LG-001" {
		t.Errorf("Expected ID 'LG-001', got '%s'", story.ID)
	}
	if len(story.Body) == 0 {
		t.Error("Expected non-empty body content")
	}
}

func TestReadStory_SpecialCharacters(t *testing.T) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	tests := []struct {
		name    string
		content string
		wantID  string
		wantErr bool
	}{
		{
			name: "unicode in title",
			content: `---
id: UC-001
title: "测试标题 with émojis 🎉"
---

Body content`,
			wantID: "UC-001",
		},
		{
			name: "newlines in field values",
			content: `---
id: NL-001
title: "Title with\nnewline"
---

Body`,
			wantID: "NL-001",
		},
		{
			name: "quotes in YAML",
			content: `---
id: QT-001
title: 'Title with "quotes" inside'
---

Body`,
			wantID: "QT-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			story, err := parser.ReadStory(ctx, testFile)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if story.ID != tt.wantID {
				t.Errorf("Expected ID '%s', got '%s'", tt.wantID, story.ID)
			}
		})
	}
}
