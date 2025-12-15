package services

import (
	"context"
	"fmt"
	"time"

	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/core/workspace"
)

// SprintStartService handles sprint creation and activation.
type SprintStartService interface {
	// StartSprint creates a new sprint, sets up current sprint link, and calculates end date.
	StartSprint(ctx context.Context, req core.StartSprintRequest) (*core.Sprint, error)
}

type sprintStartService struct {
	sprintRepo core.SprintRepository
	repoPath   string
}

// NewSprintStartService creates a new SprintStartService instance.
func NewSprintStartService(sprintRepo core.SprintRepository, repoPath string) SprintStartService {
	return &sprintStartService{
		sprintRepo: sprintRepo,
		repoPath:   repoPath,
	}
}

// StartSprint implements SprintStartService.StartSprint.
func (s *sprintStartService) StartSprint(ctx context.Context, req core.StartSprintRequest) (*core.Sprint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	paths, err := resolveWorkspacePaths(ctx, s.repoPath)
	if err != nil {
		return nil, err
	}

	// Determine sprint name
	sprintName := req.Name
	if sprintName == "" {
		// Auto-generate next sequential name
		existing, err := s.sprintRepo.ListSprints(ctx, paths.SprintsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to list existing sprints: %w", err)
		}
		sprintName = s.generateNextSprintName(existing)
	}

	// Validate sprint name
	if err := core.ValidateSprintName(sprintName); err != nil {
		return nil, fmt.Errorf("invalid sprint name: %w", err)
	}

	// Determine start date
	startDate := time.Now()
	if req.StartDate != nil {
		startDate = *req.StartDate
	}

	// Determine duration
	duration := req.Duration
	if duration == "" {
		duration = "2w" // Default: 2 weeks
	}

	// Build sprint directory path
	sprintsDir := paths.SprintsPath
	sprintDir := workspace.ResolveSprintPath(s.repoPath, sprintName, paths.Structure)

	// Check if sprint already exists
	exists, err := s.sprintRepo.SprintExists(ctx, sprintDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check if sprint exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: sprint %q already exists", core.ErrSprintExists, sprintName)
	}

	// Create sprint directory
	sprint, err := s.sprintRepo.CreateSprint(ctx, sprintDir, sprintName, startDate, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create sprint: %w", err)
	}

	// Set current sprint link
	if err := s.sprintRepo.SetCurrentSprint(ctx, sprintsDir, sprintDir); err != nil {
		// If link creation fails, we still have the sprint directory, but log the error
		// For now, we'll return an error to ensure the link is created
		return nil, fmt.Errorf("failed to create current sprint link: %w", err)
	}

	return sprint, nil
}

// generateNextSprintName generates the next sequential sprint name based on existing sprints.
func (s *sprintStartService) generateNextSprintName(existing []string) string {
	if len(existing) == 0 {
		return "Sprint-01"
	}

	// Find the highest number
	maxNum := 0
	for _, name := range existing {
		// Try to extract number from "Sprint-XX" pattern
		var num int
		if _, err := fmt.Sscanf(name, "Sprint-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	// Generate next number
	nextNum := maxNum + 1
	return fmt.Sprintf("Sprint-%02d", nextNum)
}
