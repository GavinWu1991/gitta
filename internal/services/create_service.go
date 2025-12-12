package services

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gavin/gitta/internal/core"
)

//go:embed create_templates/story-template.md
var createTemplateFS embed.FS

// CreateService handles story creation with template support and editor integration.
type CreateService interface {
	// CreateStory creates a new story file from a template, generates unique ID,
	// and launches the user's editor. Returns the created story and file path.
	CreateStory(ctx context.Context, req CreateStoryRequest) (*core.Story, string, error)
}

// CreateStoryRequest contains parameters for creating a new story.
type CreateStoryRequest struct {
	Title    string
	Prefix   string // ID prefix (default: "STORY")
	Template string // Template path (optional, uses default if empty)
	Editor   string // Editor command (optional, uses $EDITOR env var)
	Status   core.Status
	Priority core.Priority
	Assignee *string
	Tags     []string
}

type createService struct {
	idGenerator core.IDGenerator
	parser      core.StoryParser
	storyRepo   core.StoryRepository
	storyDir    string
}

// NewCreateService creates a new CreateService instance.
func NewCreateService(
	idGenerator core.IDGenerator,
	parser core.StoryParser,
	storyRepo core.StoryRepository,
	storyDir string,
) CreateService {
	return &createService{
		idGenerator: idGenerator,
		parser:      parser,
		storyRepo:   storyRepo,
		storyDir:    storyDir,
	}
}

// CreateStory implements CreateService.CreateStory.
func (s *createService) CreateStory(ctx context.Context, req CreateStoryRequest) (*core.Story, string, error) {
	// Validate context
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("context cancelled: %w", err)
	}

	// Validate request
	if strings.TrimSpace(req.Title) == "" {
		return nil, "", fmt.Errorf("title is required")
	}

	// Set defaults
	if req.Prefix == "" {
		req.Prefix = "US" // Default to "US" to match existing pattern
	}
	if req.Status == "" {
		req.Status = core.StatusTodo
	}
	if req.Priority == "" {
		req.Priority = core.PriorityMedium
	}

	// Generate unique ID
	id, err := s.idGenerator.GenerateNextID(ctx, req.Prefix)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate ID: %w", err)
	}

	// Load template
	var templateContent []byte
	if req.Template != "" {
		// Use custom template from file
		var err error
		templateContent, err = os.ReadFile(req.Template)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read template: %w", err)
		}
	} else {
		// Use embedded default template
		var err error
		templateContent, err = createTemplateFS.ReadFile("create_templates/story-template.md")
		if err != nil {
			return nil, "", fmt.Errorf("failed to read embedded template: %w", err)
		}
	}

	// Parse template
	tmpl, err := template.New("story").Parse(string(templateContent))
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	now := time.Now()
	templateData := struct {
		ID        string
		Title     string
		Status    string
		Priority  string
		Assignee  string
		CreatedAt string
		Tags      []string
	}{
		ID:        id,
		Title:     req.Title,
		Status:    string(req.Status),
		Priority:  string(req.Priority),
		CreatedAt: now.Format(time.RFC3339),
		Tags:      req.Tags,
	}
	if req.Assignee != nil {
		templateData.Assignee = *req.Assignee
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Create story file path
	fileName := fmt.Sprintf("%s.md", id)
	filePath := filepath.Join(s.storyDir, fileName)

	// Ensure story directory exists
	if err := os.MkdirAll(s.storyDir, 0755); err != nil {
		return nil, "", &core.IOError{
			Operation: "create",
			FilePath:  s.storyDir,
			Cause:     err,
		}
	}

	// Write story file atomically (temp file + rename)
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		return nil, "", &core.IOError{
			Operation: "write",
			FilePath:  filePath,
			Cause:     err,
		}
	}
	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile)
		return nil, "", &core.IOError{
			Operation: "write",
			FilePath:  filePath,
			Cause:     err,
		}
	}

	// Parse the written file to get the Story struct
	story, err := s.parser.ReadStory(ctx, filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse created story: %w", err)
	}

	// Launch editor if specified
	if req.Editor != "" || os.Getenv("EDITOR") != "" {
		editor := req.Editor
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			editor = "vi" // Default fallback
		}

		// Launch editor
		cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("%s %s", editor, filePath))
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			// Editor may have exited with non-zero code, but file might still be modified
			// Check if file was modified
			stat, statErr := os.Stat(filePath)
			if statErr != nil {
				return nil, "", fmt.Errorf("editor failed and file check failed: %w", err)
			}
			// If file exists and was modified, consider it a success
			if stat.ModTime().After(now) {
				// File was modified, re-read it
				story, err = s.parser.ReadStory(ctx, filePath)
				if err != nil {
					return nil, "", fmt.Errorf("failed to parse story after editor: %w", err)
				}
			} else {
				return nil, "", fmt.Errorf("failed to launch editor: %w", err)
			}
		} else {
			// Editor exited successfully, re-read the file
			story, err = s.parser.ReadStory(ctx, filePath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse story after editor: %w", err)
			}
		}
	}

	// Validate story
	validationErrors := s.parser.ValidateStory(story)
	if len(validationErrors) > 0 {
		return nil, "", fmt.Errorf("story validation failed: %s", validationErrors[0].Message)
	}

	return story, filePath, nil
}
