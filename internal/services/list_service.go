package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/gavin/gitta/internal/core"
)

// ListService orchestrates story listing and status derivation for the CLI.
type ListService interface {
	// ListSprintTasks returns stories from the current Sprint directory with derived status.
	ListSprintTasks(ctx context.Context, repoPath string) ([]*StoryWithStatus, error)
	// ListAllTasks returns stories from Sprint and backlog directories with derived status.
	ListAllTasks(ctx context.Context, repoPath string) ([]*StoryWithStatus, []*StoryWithStatus, error)
	// ListStories lists stories matching the given filter criteria.
	ListStories(ctx context.Context, repoPath string, filter Filter) ([]*StoryWithStatus, error)
}

// Filter represents filtering criteria for story lists.
type Filter struct {
	Statuses   []core.Status
	Priorities []core.Priority
	Assignees  []string
	Tags       []string
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

	paths, err := resolveWorkspacePaths(ctx, repoPath)
	if err != nil {
		return nil, err
	}

	sprintPath, err := s.storyRepo.FindCurrentSprint(ctx, paths.SprintsPath)
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

	paths, err := resolveWorkspacePaths(ctx, repoPath)
	if err != nil {
		return nil, nil, err
	}

	var sprintStories []*core.Story
	var backlogStories []*core.Story

	// Try Sprint stories first; if not found, continue with backlog only.
	if sprintPath, err := s.storyRepo.FindCurrentSprint(ctx, paths.SprintsPath); err == nil {
		if listed, listErr := s.storyRepo.ListStories(ctx, sprintPath); listErr == nil {
			sprintStories = listed
		} else {
			return nil, nil, fmt.Errorf("failed to list Sprint stories: %w", listErr)
		}
	}

	// Backlog stories (missing backlog directory returns empty slice).
	if listed, err := s.storyRepo.ListStories(ctx, paths.BacklogPath); err == nil {
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

// ListStories lists stories matching the given filter criteria.
func (s *listService) ListStories(ctx context.Context, repoPath string, filter Filter) ([]*StoryWithStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	paths, err := resolveWorkspacePaths(ctx, repoPath)
	if err != nil {
		return nil, err
	}

	// Collect all stories from sprint and backlog
	var allStories []*core.Story

	// Sprint stories
	if sprintPath, err := s.storyRepo.FindCurrentSprint(ctx, paths.SprintsPath); err == nil {
		if listed, err := s.storyRepo.ListStories(ctx, sprintPath); err == nil {
			allStories = append(allStories, listed...)
		}
	}

	// Backlog stories
	if listed, err := s.storyRepo.ListStories(ctx, paths.BacklogPath); err == nil {
		allStories = append(allStories, listed...)
	}

	if len(allStories) == 0 {
		return []*StoryWithStatus{}, nil
	}

	// Derive statuses
	if err := s.deriveStatuses(ctx, repoPath, allStories); err != nil {
		return nil, err
	}

	// Apply filters
	filtered := s.ApplyFilter(allStories, filter)

	// Sort
	sortStories(filtered)

	// Convert to StoryWithStatus (determine source based on directory)
	result := make([]*StoryWithStatus, 0, len(filtered))
	for _, story := range filtered {
		// Determine source (simplified - in production, track source during collection)
		source := "Sprint" // Default, could be enhanced to track actual source
		result = append(result, &StoryWithStatus{
			Story:  story,
			Status: story.Status,
			Source: source,
		})
	}

	return result, nil
}

// ApplyFilter applies the filter criteria to stories and returns matching stories.
// Filter logic: AND between fields, OR within fields.
func (s *listService) ApplyFilter(stories []*core.Story, filter Filter) []*core.Story {
	if isEmptyFilter(filter) {
		return stories
	}

	result := make([]*core.Story, 0, len(stories))
	for _, story := range stories {
		if matchesFilter(story, filter) {
			result = append(result, story)
		}
	}
	return result
}

// matchesFilter checks if a story matches the filter criteria.
func matchesFilter(story *core.Story, filter Filter) bool {
	// Status filter (OR logic)
	if len(filter.Statuses) > 0 {
		matched := false
		for _, status := range filter.Statuses {
			if story.Status == status {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Priority filter (OR logic)
	if len(filter.Priorities) > 0 {
		matched := false
		for _, priority := range filter.Priorities {
			if story.Priority == priority {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Assignee filter (OR logic)
	if len(filter.Assignees) > 0 {
		matched := false
		for _, assignee := range filter.Assignees {
			if story.Assignee != nil && *story.Assignee == assignee {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Tags filter (OR logic - story must have any of the filter tags)
	if len(filter.Tags) > 0 {
		matched := false
		for _, filterTag := range filter.Tags {
			for _, storyTag := range story.Tags {
				if storyTag == filterTag {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// isEmptyFilter checks if the filter is empty (no filtering).
func isEmptyFilter(filter Filter) bool {
	return len(filter.Statuses) == 0 &&
		len(filter.Priorities) == 0 &&
		len(filter.Assignees) == 0 &&
		len(filter.Tags) == 0
}
