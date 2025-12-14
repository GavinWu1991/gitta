package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gavin/gitta/internal/core"
)

// SprintBurndownService handles burndown chart generation from Git history.
type SprintBurndownService interface {
	// GenerateBurndown analyzes Git history and generates burndown data points.
	GenerateBurndown(ctx context.Context, sprintPath string) ([]core.BurndownDataPoint, error)
}

type sprintBurndownService struct {
	gitAnalyzer core.GitHistoryAnalyzer
	sprintRepo  core.SprintRepository
	storyRepo   core.StoryRepository
	repoPath    string
}

// NewSprintBurndownService creates a new SprintBurndownService instance.
func NewSprintBurndownService(
	gitAnalyzer core.GitHistoryAnalyzer,
	sprintRepo core.SprintRepository,
	storyRepo core.StoryRepository,
	repoPath string,
) SprintBurndownService {
	return &sprintBurndownService{
		gitAnalyzer: gitAnalyzer,
		sprintRepo:  sprintRepo,
		storyRepo:   storyRepo,
		repoPath:    repoPath,
	}
}

// GenerateBurndown implements SprintBurndownService.GenerateBurndown.
func (s *sprintBurndownService) GenerateBurndown(ctx context.Context, sprintPath string) ([]core.BurndownDataPoint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Get sprint information (we need start/end dates)
	// For now, we'll need to read sprint metadata or infer from directory structure
	// Let's get the sprint name from path
	sprintName := filepath.Base(sprintPath)

	// Get sprint start date - we'll need to read it from sprint metadata or infer
	// For MVP, let's try to get it from the sprint directory creation or use a default
	startDate := time.Now().AddDate(0, 0, -14) // Default: 2 weeks ago
	endDate := time.Now()

	// Calculate relative sprint directory path from repo root
	sprintDir, err := filepath.Rel(s.repoPath, sprintPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate relative sprint path: %w", err)
	}

	// Analyze Git history for the sprint
	req := core.AnalyzeHistoryRequest{
		RepoPath:            s.repoPath,
		SprintDir:           sprintDir,
		StartDate:           startDate,
		EndDate:             endDate,
		IncludeMergeCommits: false,
	}

	snapshots, err := s.gitAnalyzer.AnalyzeSprintHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze sprint history: %w", err)
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("%w: no commits found for sprint %s", core.ErrInsufficientHistory, sprintName)
	}

	// Calculate initial totals from first snapshot
	firstSnapshot := snapshots[0]
	initialTasks := len(firstSnapshot.Files)
	initialPoints := s.calculateTotalPoints(firstSnapshot.Files)

	// Generate burndown data points
	var dataPoints []core.BurndownDataPoint

	for _, snapshot := range snapshots {
		// Calculate remaining work
		remainingTasks := s.countIncompleteTasks(snapshot.Files)
		remainingPoints := s.calculateRemainingPoints(snapshot.Files)

		// Normalize date to midnight for grouping
		date := time.Date(
			snapshot.CommitDate.Year(),
			snapshot.CommitDate.Month(),
			snapshot.CommitDate.Day(),
			0, 0, 0, 0,
			snapshot.CommitDate.Location(),
		)

		dataPoint := core.BurndownDataPoint{
			Date:            date,
			RemainingPoints: remainingPoints,
			RemainingTasks:  remainingTasks,
			TotalPoints:     &initialPoints,
			TotalTasks:      &initialTasks,
		}

		dataPoints = append(dataPoints, dataPoint)
	}

	// Fill in missing days with previous day's values
	dataPoints = s.fillMissingDays(dataPoints, startDate, endDate)

	return dataPoints, nil
}

// calculateTotalPoints calculates total story points from files.
// For now, we'll use task count as points (can be enhanced later).
func (s *sprintBurndownService) calculateTotalPoints(files map[string]*core.Story) int {
	return len(files)
}

// calculateRemainingPoints calculates remaining story points from incomplete tasks.
func (s *sprintBurndownService) calculateRemainingPoints(files map[string]*core.Story) int {
	count := 0
	for _, story := range files {
		if story.Status != core.StatusDone {
			count++
		}
	}
	return count
}

// countIncompleteTasks counts tasks that are not done.
func (s *sprintBurndownService) countIncompleteTasks(files map[string]*core.Story) int {
	count := 0
	for _, story := range files {
		if story.Status == "" || story.Status != core.StatusDone {
			count++
		}
	}
	return count
}

// fillMissingDays fills in missing days with the previous day's values.
func (s *sprintBurndownService) fillMissingDays(
	dataPoints []core.BurndownDataPoint,
	startDate, endDate time.Time,
) []core.BurndownDataPoint {
	if len(dataPoints) == 0 {
		return dataPoints
	}

	// Create a map of existing dates
	dateMap := make(map[string]core.BurndownDataPoint)
	for _, dp := range dataPoints {
		dateKey := dp.Date.Format("2006-01-02")
		dateMap[dateKey] = dp
	}

	// Fill in missing days
	var filled []core.BurndownDataPoint
	currentDate := startDate
	lastValue := dataPoints[0]

	for !currentDate.After(endDate) {
		dateKey := currentDate.Format("2006-01-02")
		if dp, exists := dateMap[dateKey]; exists {
			filled = append(filled, dp)
			lastValue = dp
		} else {
			// Use last known value
			dp := lastValue
			dp.Date = currentDate
			filled = append(filled, dp)
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return filled
}
