package filesystem

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gavin/gitta/internal/core"
)

// Repository is a filesystem-backed implementation of core.StoryRepository.
// It scans directories for Markdown story files and parses them using a provided
// core.StoryParser. All operations respect context cancellation.
type Repository struct {
	parser core.StoryParser
}

// NewRepository constructs a Repository with the provided parser.
func NewRepository(parser core.StoryParser) *Repository {
	return &Repository{parser: parser}
}

// NewDefaultRepository constructs a Repository using the MarkdownParser.
func NewDefaultRepository() *Repository {
	return &Repository{parser: NewMarkdownParser()}
}

// ListStories scans dirPath for Markdown story files and returns parsed stories.
// Missing directories return an empty slice without error to allow graceful handling
// for optional locations (e.g., backlog).
func (r *Repository) ListStories(ctx context.Context, dirPath string) ([]*core.Story, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	info, statErr := os.Stat(dirPath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return []*core.Story{}, nil
		}
		return nil, &core.IOError{
			Operation: "read",
			FilePath:  dirPath,
			Cause:     statErr,
		}
	}

	if !info.IsDir() {
		return nil, &core.IOError{
			Operation: "read",
			FilePath:  dirPath,
			Cause:     fmt.Errorf("path is not a directory"),
		}
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, &core.IOError{
			Operation: "read",
			FilePath:  dirPath,
			Cause:     err,
		}
	}

	var stories []*core.Story
	var parseErrors []error

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return stories, ctx.Err()
		default:
		}

		if entry.IsDir() {
			continue
		}

		if !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		story, parseErr := r.parser.ReadStory(ctx, filePath)
		if parseErr != nil {
			parseErrors = append(parseErrors, parseErr)
			continue
		}
		if validationErrors := r.parser.ValidateStory(story); len(validationErrors) > 0 {
			parseErrors = append(parseErrors, fmt.Errorf("validation failed for %s: %s", filePath, validationErrors[0].Message))
			continue
		}
		stories = append(stories, story)
	}

	if len(parseErrors) > 0 {
		return stories, errors.Join(parseErrors...)
	}

	return stories, nil
}

// FindCurrentSprint locates the current Sprint directory within sprintsDir using
// case-insensitive lexicographic ordering to select the highest Sprint name.
func (r *Repository) FindCurrentSprint(ctx context.Context, sprintsDir string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	entries, err := os.ReadDir(sprintsDir)
	if err != nil {
		return "", &core.IOError{
			Operation: "read",
			FilePath:  sprintsDir,
			Cause:     err,
		}
	}

	var sprintDirs []string
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(strings.ToLower(name), "sprint") {
			sprintDirs = append(sprintDirs, name)
		}
	}

	if len(sprintDirs) == 0 {
		return "", fmt.Errorf("no Sprint directories found in %s", sprintsDir)
	}

	sort.SliceStable(sprintDirs, func(i, j int) bool {
		return strings.ToLower(sprintDirs[i]) < strings.ToLower(sprintDirs[j])
	})

	return filepath.Join(sprintsDir, sprintDirs[len(sprintDirs)-1]), nil
}
