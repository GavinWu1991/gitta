package services

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/gavin/gitta/internal/core"
)

// ListService orchestrates story listing and status derivation for the CLI.
type ListService interface {
	// ListSprintTasks returns stories from the current Sprint directory with derived status.
	ListSprintTasks(ctx context.Context, repoPath string) ([]*StoryWithStatus, error)
	// ListAllTasks returns stories from Sprint and backlog directories with derived status.
	ListAllTasks(ctx context.Context, repoPath string) ([]*StoryWithStatus, []*StoryWithStatus, error)
}

type listService struct {
	storyRepo    core.StoryRepository
	statusEngine StatusEngine
	gitRepo      core.GitRepository
}

// StoryWithStatus couples a story with its derived status and origin.
type StoryWithStatus struct {
	Story  *core.Story
	Status core.Status
	Source string
}

// NewListService constructs a ListService with the provided dependencies.
func NewListService(storyRepo core.StoryRepository, gitRepo core.GitRepository) ListService {
	return &listService{
		storyRepo:    storyRepo,
		statusEngine: NewStatusEngineWithRepository(gitRepo),
		gitRepo:      gitRepo,
	}
}

func (s *listService) ListSprintTasks(ctx context.Context, repoPath string) ([]*StoryWithStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sprintPath, err := s.storyRepo.FindCurrentSprint(ctx, filepath.Join(repoPath, "sprints"))
	if err != nil {
		return nil, fmt.Errorf("failed to find current Sprint: %w", err)
	}

	stories, err := s.storyRepo.ListStories(ctx, sprintPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list Sprint stories: %w", err)
	}

	if err := s.deriveStatuses(ctx, repoPath, stories); err != nil {
		return nil, err
	}

	sortStories(stories)
	return toStoryWithStatus(stories, "Sprint"), nil
}

func (s *listService) ListAllTasks(ctx context.Context, repoPath string) ([]*StoryWithStatus, []*StoryWithStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	var sprintStories []*core.Story
	var backlogStories []*core.Story

	// Try Sprint stories first; if not found, continue with backlog only.
	if sprintPath, err := s.storyRepo.FindCurrentSprint(ctx, filepath.Join(repoPath, "sprints")); err == nil {
		if listed, listErr := s.storyRepo.ListStories(ctx, sprintPath); listErr == nil {
			sprintStories = listed
		} else {
			return nil, nil, fmt.Errorf("failed to list Sprint stories: %w", listErr)
		}
	}

	// Backlog stories (missing backlog directory returns empty slice).
	if listed, err := s.storyRepo.ListStories(ctx, filepath.Join(repoPath, "backlog")); err == nil {
		backlogStories = listed
	} else {
		return nil, nil, fmt.Errorf("failed to list backlog stories: %w", err)
	}

	stories := append([]*core.Story{}, sprintStories...)
	stories = append(stories, backlogStories...)

	if len(stories) == 0 {
		return toStoryWithStatus(sprintStories, "Sprint"), toStoryWithStatus(backlogStories, "Backlog"), nil
	}

	if err := s.deriveStatuses(ctx, repoPath, stories); err != nil {
		return nil, nil, err
	}

	// Sort within groups to maintain clear separation.
	sortStories(sprintStories)
	sortStories(backlogStories)

	return toStoryWithStatus(sprintStories, "Sprint"), toStoryWithStatus(backlogStories, "Backlog"), nil
}

func (s *listService) deriveStatuses(ctx context.Context, repoPath string, stories []*core.Story) error {
	if len(stories) == 0 {
		return nil
	}

	branchList, err := s.gitRepo.GetBranchList(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to derive task status: %w", err)
	}

	statuses, err := s.statusEngine.DeriveStatusBatch(ctx, stories, branchList, repoPath)
	if err != nil {
		return fmt.Errorf("failed to derive task status: %w", err)
	}

	for i, status := range statuses {
		stories[i].Status = status
	}
	return nil
}

func sortStories(stories []*core.Story) {
	sort.SliceStable(stories, func(i, j int) bool {
		return stories[i].ID < stories[j].ID
	})
}

func toStoryWithStatus(stories []*core.Story, source string) []*StoryWithStatus {
	withStatus := make([]*StoryWithStatus, 0, len(stories))
	for _, story := range stories {
		withStatus = append(withStatus, &StoryWithStatus{
			Story:  story,
			Status: story.Status,
			Source: source,
		})
	}
	return withStatus
}
