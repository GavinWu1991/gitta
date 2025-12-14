package filesystem

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gavin/gitta/internal/core"
)

// Repository is a filesystem-backed implementation of core.StoryRepository.
// It scans directories for Markdown story files and parses them using a provided
// core.StoryParser. All operations respect context cancellation.
type Repository struct {
	parser core.StoryParser
}

var storyIDPattern = regexp.MustCompile(`^[A-Z]{2}-[0-9]+$`)

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

// FindStoryByID searches Sprint directories (current Sprint first) then backlog for a story ID.
func (r *Repository) FindStoryByID(ctx context.Context, repoPath, storyID string) (*core.Story, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}

	if storyID == "" || !storyIDPattern.MatchString(storyID) {
		return nil, "", fmt.Errorf("%w: invalid story ID format %q", core.ErrInvalidPath, storyID)
	}

	// Search current Sprint first.
	sprintsDir := filepath.Join(repoPath, "sprints")
	if sprintPath, err := r.FindCurrentSprint(ctx, sprintsDir); err == nil {
		if story, path, err := r.findStoryInDir(ctx, sprintPath, storyID); err != nil {
			return nil, "", err
		} else if story != nil {
			return story, path, nil
		}
	}

	// Search other Sprint directories (lexicographic order) excluding the current one.
	entries, err := os.ReadDir(sprintsDir)
	if err == nil {
		for _, entry := range entries {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			default:
			}

			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasPrefix(strings.ToLower(name), "sprint") {
				continue
			}
			sprintPath := filepath.Join(sprintsDir, name)
			if story, path, err := r.findStoryInDir(ctx, sprintPath, storyID); err != nil {
				return nil, "", err
			} else if story != nil {
				return story, path, nil
			}
		}
	}

	// Search backlog.
	backlogPath := filepath.Join(repoPath, "backlog")
	if story, path, err := r.findStoryInDir(ctx, backlogPath, storyID); err != nil {
		return nil, "", err
	} else if story != nil {
		return story, path, nil
	}

	return nil, "", core.ErrStoryNotFound
}

// FindStoryByPath reads and validates a story from a Markdown file path.
func (r *Repository) FindStoryByPath(ctx context.Context, filePath string) (*core.Story, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, core.ErrInvalidPath
		}
		return nil, &core.IOError{
			Operation: "stat",
			FilePath:  filePath,
			Cause:     err,
		}
	}

	if info.IsDir() {
		return nil, core.ErrInvalidPath
	}

	if ext := strings.ToLower(filepath.Ext(filePath)); ext != ".md" {
		return nil, core.ErrInvalidPath
	}

	story, err := r.parser.ReadStory(ctx, filePath)
	if err != nil {
		return nil, err
	}
	if validationErrors := r.parser.ValidateStory(story); len(validationErrors) > 0 {
		return nil, fmt.Errorf("story validation failed for %s: %s", filePath, validationErrors[0].Message)
	}

	return story, nil
}

// CreateSprint creates a new sprint directory and returns the sprint entity.
// The sprintDir parameter should be the full path to the sprint directory to create.
func (r *Repository) CreateSprint(ctx context.Context, sprintDir string, name string, startDate time.Time, duration string) (*core.Sprint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Validate sprint name
	if err := core.ValidateSprintName(name); err != nil {
		return nil, fmt.Errorf("invalid sprint name: %w", err)
	}

	// Check if sprint already exists
	exists, err := r.SprintExists(ctx, sprintDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check if sprint exists: %w", err)
	}
	if exists {
		return nil, core.ErrSprintExists
	}

	// Create sprint directory
	if err := os.MkdirAll(sprintDir, 0755); err != nil {
		return nil, &core.IOError{
			Operation: "create",
			FilePath:  sprintDir,
			Cause:     err,
		}
	}

	// Calculate end date
	endDate, err := core.CalculateEndDate(startDate, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate end date: %w", err)
	}

	return &core.Sprint{
		Name:          name,
		StartDate:     startDate,
		EndDate:       endDate,
		Duration:      duration,
		DirectoryPath: sprintDir,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

// SetCurrentSprint creates/updates the current sprint link to point to the given sprint.
func (r *Repository) SetCurrentSprint(ctx context.Context, sprintsDir string, sprintPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	linkPath := filepath.Join(sprintsDir, ".current-sprint")
	_, err := CreateCurrentSprintLink(sprintPath, linkPath)
	return err
}

// ListSprints returns all sprint directories in lexicographic order.
func (r *Repository) ListSprints(ctx context.Context, sprintsDir string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(sprintsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, &core.IOError{
			Operation: "read",
			FilePath:  sprintsDir,
			Cause:     err,
		}
	}

	var sprintDirs []string
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
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

	sort.SliceStable(sprintDirs, func(i, j int) bool {
		return strings.ToLower(sprintDirs[i]) < strings.ToLower(sprintDirs[j])
	})

	return sprintDirs, nil
}

// SprintExists checks if a sprint directory exists.
func (r *Repository) SprintExists(ctx context.Context, sprintPath string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	info, err := os.Stat(sprintPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, &core.IOError{
			Operation: "stat",
			FilePath:  sprintPath,
			Cause:     err,
		}
	}

	return info.IsDir(), nil
}

// findStoryInDir searches a directory for a story ID and returns the story and file path.
func (r *Repository) findStoryInDir(ctx context.Context, dirPath, storyID string) (*core.Story, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", nil
		}
		return nil, "", &core.IOError{
			Operation: "read",
			FilePath:  dirPath,
			Cause:     err,
		}
	}

	if !info.IsDir() {
		return nil, "", &core.IOError{
			Operation: "read",
			FilePath:  dirPath,
			Cause:     fmt.Errorf("path is not a directory"),
		}
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, "", &core.IOError{
			Operation: "read",
			FilePath:  dirPath,
			Cause:     err,
		}
	}

	var parseErrors []error

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
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

		if story.ID == storyID {
			return story, filePath, nil
		}
	}

	if len(parseErrors) > 0 {
		return nil, "", errors.Join(parseErrors...)
	}

	return nil, "", nil
}
