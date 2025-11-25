package filesystem

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"

	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

// MarkdownParser implements core.StoryParser for reading and writing Markdown story files.
// It uses Goldmark with the meta extension to parse YAML frontmatter and extract
// Markdown body content. MarkdownParser is thread-safe for concurrent read operations.
type MarkdownParser struct {
	md goldmark.Markdown
}

// NewMarkdownParser creates a new MarkdownParser instance.
// The parser is configured with Goldmark's meta extension to extract YAML frontmatter.
// Returns a ready-to-use parser instance.
func NewMarkdownParser() *MarkdownParser {
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)
	return &MarkdownParser{md: md}
}

// ReadStory reads a Markdown file and parses it into a Story struct.
// It extracts YAML frontmatter metadata and Markdown body content.
// Missing optional fields are set to default values (Priority: Medium, Status: Todo).
// Returns an error if the file cannot be read, exceeds size limits, or contains
// malformed YAML frontmatter.
func (p *MarkdownParser) ReadStory(ctx context.Context, filePath string) (*core.Story, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, &core.IOError{
			Operation: "read",
			FilePath:  filePath,
			Cause:     err,
		}
	}

	// Check file size
	if len(data) > maxFileSize {
		return nil, &core.IOError{
			Operation: "read",
			FilePath:  filePath,
			Cause:     fmt.Errorf("file size %d exceeds maximum %d bytes", len(data), maxFileSize),
		}
	}

	// Parse with Goldmark to extract frontmatter
	var buf bytes.Buffer
	context := parser.NewContext()
	if err := p.md.Convert(data, &buf, parser.WithContext(context)); err != nil {
		return nil, &core.ParseError{
			FilePath: filePath,
			Message:  fmt.Sprintf("failed to parse Markdown: %v", err),
			Cause:    err,
		}
	}

	// Extract metadata from context
	metaData := meta.Get(context)

	// Extract body content (everything after frontmatter)
	body := extractBody(string(data))

	// Unmarshal frontmatter into Story struct
	var story core.Story
	if len(metaData) > 0 {
		// Convert metaData map to YAML bytes for unmarshaling
		yamlData, err := yaml.Marshal(metaData)
		if err != nil {
			return nil, &core.ParseError{
				FilePath: filePath,
				Message:  fmt.Sprintf("failed to marshal frontmatter: %v", err),
				Cause:    err,
			}
		}

		if err := yaml.Unmarshal(yamlData, &story); err != nil {
			return nil, &core.ParseError{
				FilePath: filePath,
				Message:  fmt.Sprintf("failed to unmarshal YAML frontmatter: %v", err),
				Cause:    err,
			}
		}
	}

	// Set body content
	story.Body = body

	// Apply default values for missing optional fields
	applyDefaults(&story)

	return &story, nil
}

// extractBody extracts the Markdown body content after the frontmatter delimiter.
func extractBody(content string) string {
	// Look for frontmatter delimiter
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return ""
	}

	// Check if first line is frontmatter delimiter
	if strings.TrimSpace(lines[0]) != "---" {
		// No frontmatter, entire content is body
		return content
	}

	// Find the closing delimiter
	bodyStart := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			bodyStart = i + 1
			break
		}
	}

	if bodyStart == -1 || bodyStart >= len(lines) {
		// No closing delimiter found or body is empty
		return ""
	}

	// Join lines from bodyStart onwards
	bodyLines := lines[bodyStart:]
	body := strings.Join(bodyLines, "\n")

	// Remove leading newline if present
	body = strings.TrimPrefix(body, "\n")

	return body
}

// applyDefaults sets default values for missing optional fields.
func applyDefaults(story *core.Story) {
	if story.Priority == "" {
		story.Priority = core.PriorityMedium
	}
	if story.Status == "" {
		story.Status = core.StatusTodo
	}
	if story.Tags == nil {
		story.Tags = []string{}
	}
	// Assignee, CreatedAt, UpdatedAt remain nil if not set (pointer types)
}

// WriteStory writes a Story struct to a Markdown file.
// It validates the story before writing, then marshals metadata to YAML frontmatter
// and writes the body content. Uses atomic writes (temp file + rename) to prevent
// corruption. Preserves line endings from the original file if updating an existing file.
// Creates parent directories if they don't exist. Returns an error if validation fails
// or the file cannot be written.
func (p *MarkdownParser) WriteStory(ctx context.Context, filePath string, story *core.Story) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if story == nil {
		return &core.IOError{
			Operation: "write",
			FilePath:  filePath,
			Cause:     fmt.Errorf("story is nil"),
		}
	}

	// Validate story before writing
	validationErrors := p.ValidateStory(story)
	if len(validationErrors) > 0 {
		// Return first validation error as ParseError
		// In a real implementation, we might want to return all errors
		return &core.ParseError{
			FilePath: filePath,
			Message:  fmt.Sprintf("validation failed: %s", validationErrors[0].Message),
		}
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &core.IOError{
			Operation: "write",
			FilePath:  filePath,
			Cause:     fmt.Errorf("failed to create directory %s: %w", dir, err),
		}
	}

	// Detect line endings from existing file if it exists
	lineEnding := "\n" // Default to Unix line endings
	if existingData, err := os.ReadFile(filePath); err == nil {
		lineEnding = detectLineEnding(string(existingData))
	}

	// Marshal frontmatter to YAML
	frontmatterData, err := yaml.Marshal(story)
	if err != nil {
		return &core.ParseError{
			FilePath: filePath,
			Message:  fmt.Sprintf("failed to marshal story to YAML: %v", err),
			Cause:    err,
		}
	}

	// Build file content: frontmatter + body
	var content strings.Builder
	content.WriteString("---")
	content.WriteString(lineEnding)
	content.Write(frontmatterData)
	content.WriteString("---")
	content.WriteString(lineEnding)
	content.WriteString(lineEnding)
	content.WriteString(story.Body)

	// Normalize line endings in content
	contentStr := normalizeLineEndings(content.String(), lineEnding)

	// Atomic write: write to temp file, then rename
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(contentStr), 0644); err != nil {
		return &core.IOError{
			Operation: "write",
			FilePath:  filePath,
			Cause:     fmt.Errorf("failed to write temp file: %w", err),
		}
	}

	// Rename temp file to target (atomic operation)
	if err := os.Rename(tmpFile, filePath); err != nil {
		// Clean up temp file on error
		os.Remove(tmpFile)
		return &core.IOError{
			Operation: "write",
			FilePath:  filePath,
			Cause:     fmt.Errorf("failed to rename temp file: %w", err),
		}
	}

	return nil
}

// detectLineEnding detects the line ending style used in the content.
func detectLineEnding(content string) string {
	if strings.Contains(content, "\r\n") {
		return "\r\n" // Windows
	}
	return "\n" // Unix/Mac
}

// normalizeLineEndings normalizes all line endings in content to the specified line ending.
func normalizeLineEndings(content, lineEnding string) string {
	// Replace all line ending variations with the target line ending
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if lineEnding != "\n" {
		content = strings.ReplaceAll(content, "\n", lineEnding)
	}
	return content
}

// ValidateStory validates a Story struct against business rules.
// It delegates to the service layer validator to check all field constraints.
// Returns a slice of ValidationErrors describing any violations. An empty slice
// indicates the story is valid.
//
// Note: This creates a dependency on internal/services, which is a minor architectural
// trade-off. In a production system, validation could be injected as a dependency
// to maintain strict hexagonal boundaries.
func (p *MarkdownParser) ValidateStory(story *core.Story) []core.ValidationError {
	return services.ValidateStory(story)
}
