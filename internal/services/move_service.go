package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gavin/gitta/internal/core"
)

// MoveService handles atomic story file moves.
type MoveService interface {
	// MoveStory moves a story file to a different directory atomically.
	MoveStory(ctx context.Context, storyID string, targetDir string, force bool) error
}

type moveService struct {
	parser    core.StoryParser
	storyRepo core.StoryRepository
	repoPath  string
}

// NewMoveService creates a new MoveService instance.
func NewMoveService(parser core.StoryParser, storyRepo core.StoryRepository, repoPath string) MoveService {
	return &moveService{
		parser:    parser,
		storyRepo: storyRepo,
		repoPath:  repoPath,
	}
}

// MoveStory implements MoveService.MoveStory.
func (s *moveService) MoveStory(ctx context.Context, storyID string, targetDir string, force bool) error {
	// Validate context
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Validate target directory (prevent path traversal)
	if err := validatePath(targetDir); err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	// Find story by ID
	story, sourcePath, err := s.storyRepo.FindStoryByID(ctx, s.repoPath, storyID)
	if err != nil {
		if err == core.ErrStoryNotFound {
			return fmt.Errorf("story not found: %s", storyID)
		}
		return fmt.Errorf("failed to find story: %w", err)
	}

	// Resolve target directory (relative to repo root)
	targetDirPath := filepath.Join(s.repoPath, targetDir)
	if !filepath.IsAbs(targetDir) {
		targetDirPath = filepath.Join(s.repoPath, targetDir)
	}

	// Create target directory if needed
	if err := os.MkdirAll(targetDirPath, 0755); err != nil {
		return &core.IOError{
			Operation: "create",
			FilePath:  targetDirPath,
			Cause:     err,
		}
	}

	// Build target file path
	fileName := filepath.Base(sourcePath)
	targetPath := filepath.Join(targetDirPath, fileName)

	// Check if target file exists
	if _, err := os.Stat(targetPath); err == nil {
		if !force {
			return fmt.Errorf("target file exists: %s (use --force to overwrite)", targetPath)
		}
	}

	// Read story content (already have story from FindStoryByID)
	// Write to temp file at destination
	tmpTargetPath := targetPath + ".tmp"
	if err := s.parser.WriteStory(ctx, tmpTargetPath, story); err != nil {
		return fmt.Errorf("failed to write to target: %w", err)
	}

	// Rename temp file to final name (atomic)
	if err := os.Rename(tmpTargetPath, targetPath); err != nil {
		// Clean up temp file on error
		os.Remove(tmpTargetPath)
		return &core.IOError{
			Operation: "rename",
			FilePath:  targetPath,
			Cause:     err,
		}
	}

	// Remove source file only after successful write
	if err := os.Remove(sourcePath); err != nil {
		// If removal fails, we still have the file at target, which is acceptable
		// Log warning but don't fail the operation
		return fmt.Errorf("failed to remove source file (target file created): %w", err)
	}

	return nil
}

// validatePath validates that the path doesn't contain path traversal attacks.
func validatePath(path string) error {
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Check for absolute paths (should be relative to repo)
	if filepath.IsAbs(path) {
		// Allow absolute paths that are within repo, but warn
		// For now, we'll allow it but could add stricter validation
	}

	return nil
}
