package unit

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestCreateService_CreateStory(t *testing.T) {
	tests := []struct {
		name          string
		req           services.CreateStoryRequest
		wantIDPattern string
		wantErr       bool
		setupFunc     func(t *testing.T, tmpDir string)
	}{
		{
			name: "create story with default prefix",
			req: services.CreateStoryRequest{
				Title:    "Test Story",
				Prefix:   "US",
				Status:   core.StatusTodo,
				Priority: core.PriorityMedium,
			},
			wantIDPattern: "US-1",
			wantErr:       false,
		},
		{
			name: "create story with custom prefix",
			req: services.CreateStoryRequest{
				Title:    "Bug Fix",
				Prefix:   "BG",
				Status:   core.StatusTodo,
				Priority: core.PriorityHigh,
			},
			wantIDPattern: "BG-1",
			wantErr:       false,
		},
		{
			name: "create story with assignee and tags",
			req: services.CreateStoryRequest{
				Title:    "Feature Story",
				Prefix:   "US",
				Status:   core.StatusDoing,
				Priority: core.PriorityHigh,
				Assignee: stringPtrHelper("alice"),
				Tags:     []string{"frontend", "ui"},
			},
			wantIDPattern: "US-1",
			wantErr:       false,
		},
		{
			name: "create story with custom template",
			req: services.CreateStoryRequest{
				Title:    "Custom Template Story",
				Prefix:   "US",
				Template: "", // Will be set in setupFunc
				Status:   core.StatusTodo,
				Priority: core.PriorityMedium,
			},
			wantIDPattern: "US-1",
			wantErr:       false,
			setupFunc: func(t *testing.T, tmpDir string) {
				// Create custom template
				templatePath := filepath.Join(tmpDir, "custom-template.md")
				templateContent := `---
id: {{.ID}}
title: {{.Title}}
status: {{.Status}}
---
# {{.Title}}
Custom template content
`
				if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
					t.Fatalf("Failed to create custom template: %v", err)
				}
			},
		},
		{
			name: "empty title should fail validation",
			req: services.CreateStoryRequest{
				Title:    "",
				Prefix:   "US",
				Status:   core.StatusTodo,
				Priority: core.PriorityMedium,
			},
			wantErr: true,
		},
		{
			name: "invalid prefix should fail",
			req: services.CreateStoryRequest{
				Title:    "Test Story",
				Prefix:   "invalid",
				Status:   core.StatusTodo,
				Priority: core.PriorityMedium,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			storyDir := filepath.Join(tmpDir, "stories")
			if err := os.MkdirAll(storyDir, 0755); err != nil {
				t.Fatalf("Failed to create story directory: %v", err)
			}

			// Setup if needed
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
				if tt.req.Template != "" {
					tt.req.Template = filepath.Join(tmpDir, "custom-template.md")
				}
			}

			// Create service dependencies
			idGenerator := filesystem.NewIDCounter(tmpDir)
			parser := filesystem.NewMarkdownParser()
			storyRepo := filesystem.NewRepository(parser)

			// Create service (we'll need to implement this)
			// For now, this test will fail until CreateService is implemented
			createService := services.NewCreateService(idGenerator, parser, storyRepo, storyDir)

			ctx := context.Background()
			story, filePath, err := createService.CreateStory(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateStory() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateStory() error = %v, want nil", err)
				return
			}

			if story == nil {
				t.Fatal("CreateStory() returned nil story")
			}

			// Verify ID pattern
			if story.ID != tt.wantIDPattern {
				t.Errorf("CreateStory() ID = %v, want %v", story.ID, tt.wantIDPattern)
			}

			// Verify title
			if story.Title != tt.req.Title {
				t.Errorf("CreateStory() Title = %v, want %v", story.Title, tt.req.Title)
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
		})
	}
}

func TestCreateService_EditorIntegration(t *testing.T) {
	// Skip on Windows - editor integration is Unix-focused
	if runtime.GOOS == "windows" {
		t.Skip("Skipping editor integration test on Windows")
	}

	// This test verifies editor integration (without actually launching an editor)
	// We'll use a cross-platform mock editor approach
	tmpDir := t.TempDir()
	storyDir := filepath.Join(tmpDir, "stories")
	if err := os.MkdirAll(storyDir, 0755); err != nil {
		t.Fatalf("Failed to create story directory: %v", err)
	}

	// Create a cross-platform mock editor
	mockEditor := createMockEditor(t, tmpDir)

	idGenerator := filesystem.NewIDCounter(tmpDir)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	story, filePath, err := createService.CreateStory(ctx, req)
	if err != nil {
		t.Fatalf("CreateStory() with editor error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Story file was not created")
	}

	// Verify story was created
	if story == nil {
		t.Fatal("Story is nil")
	}

	// Verify editor modified the file (check for the comment)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read story file: %v", err)
	}
	if !strings.Contains(string(content), "Edited by mock editor") {
		t.Error("Editor did not modify the file")
	}
}

// createMockEditor creates a cross-platform mock editor for testing
func createMockEditor(t *testing.T, tmpDir string) string {
	t.Helper()

	// Create a shell script
	mockEditor := filepath.Join(tmpDir, "mock-editor.sh")
	editorScript := `#!/bin/sh
# Mock editor that just adds a comment to the file
echo "" >> "$1"
echo "# Edited by mock editor" >> "$1"
`
	if err := os.WriteFile(mockEditor, []byte(editorScript), 0755); err != nil {
		t.Fatalf("Failed to create mock editor: %v", err)
	}

	// Return the script path - it will be executed via sh -c in create_service.go
	return mockEditor
}

// stringPtrHelper creates a string pointer (helper function for tests)
func stringPtrHelper(s string) *string {
	return &s
}
