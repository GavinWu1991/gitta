package services

import (
	"context"
	"fmt"
	"time"

	"github.com/gavin/gitta/internal/core"
)

// UpdateService handles atomic story status updates.
type UpdateService interface {
	// UpdateStatus updates a story's status atomically.
	UpdateStatus(ctx context.Context, storyID string, newStatus core.Status) error
}

type updateService struct {
	parser    core.StoryParser
	storyRepo core.StoryRepository
	repoPath  string
}

// NewUpdateService creates a new UpdateService instance.
func NewUpdateService(parser core.StoryParser, storyRepo core.StoryRepository, repoPath string) UpdateService {
	return &updateService{
		parser:    parser,
		storyRepo: storyRepo,
		repoPath:  repoPath,
	}
}

// UpdateStatus implements UpdateService.UpdateStatus.
func (s *updateService) UpdateStatus(ctx context.Context, storyID string, newStatus core.Status) error {
	// Validate context
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Validate status
	validStatuses := map[core.Status]bool{
		core.StatusTodo:   true,
		core.StatusDoing:  true,
		core.StatusReview: true,
		core.StatusDone:   true,
	}
	if !validStatuses[newStatus] {
		return fmt.Errorf("invalid status: %s (valid: todo, doing, review, done)", newStatus)
	}

	// Find story by ID
	story, filePath, err := s.storyRepo.FindStoryByID(ctx, s.repoPath, storyID)
	if err != nil {
		if err == core.ErrStoryNotFound {
			return fmt.Errorf("story not found: %s", storyID)
		}
		return fmt.Errorf("failed to find story: %w", err)
	}

	// Update status and timestamp
	story.Status = newStatus
	now := time.Now()
	story.UpdatedAt = &now

	// Validate story
	validationErrors := s.parser.ValidateStory(story)
	if len(validationErrors) > 0 {
		return fmt.Errorf("story validation failed: %s", validationErrors[0].Message)
	}

	// Write atomically (using parser's WriteStory which handles atomic writes)
	if err := s.parser.WriteStory(ctx, filePath, story); err != nil {
		return fmt.Errorf("failed to update story: %w", err)
	}

	return nil
}
