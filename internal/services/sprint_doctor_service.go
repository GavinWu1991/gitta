package services

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gavin/gitta/internal/core"
)

// Inconsistency represents a detected inconsistency between visual indicator and status file.
type Inconsistency struct {
	SprintPath    string            // Path to sprint directory
	FolderName    string            // Current folder name
	FolderStatus  core.SprintStatus // Status inferred from folder prefix
	StatusFile    core.SprintStatus // Status from .gitta/status file
	ExpectedName  string            // Expected folder name
	HasStatusFile bool              // Whether .gitta/status exists
}

// RepairResult contains the result of repair operations.
type RepairResult struct {
	RepairedCount int     // Number of sprints repaired
	FailedCount   int     // Number of repairs that failed
	Errors        []error // Errors encountered during repair
}

// SprintDoctorService handles detection and repair of sprint status inconsistencies.
type SprintDoctorService interface {
	// DetectInconsistencies scans all sprints and detects inconsistencies.
	DetectInconsistencies(ctx context.Context) ([]Inconsistency, error)
	// RepairInconsistencies repairs detected inconsistencies.
	RepairInconsistencies(ctx context.Context, inconsistencies []Inconsistency) (*RepairResult, error)
}

type sprintDoctorService struct {
	sprintRepo core.SprintRepository
	repoPath   string
}

// NewSprintDoctorService creates a new SprintDoctorService instance.
func NewSprintDoctorService(sprintRepo core.SprintRepository, repoPath string) SprintDoctorService {
	return &sprintDoctorService{
		sprintRepo: sprintRepo,
		repoPath:   repoPath,
	}
}

// DetectInconsistencies implements SprintDoctorService.DetectInconsistencies.
func (s *sprintDoctorService) DetectInconsistencies(ctx context.Context) ([]Inconsistency, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	paths, err := resolveWorkspacePaths(ctx, s.repoPath)
	if err != nil {
		return nil, err
	}
	sprintsDir := paths.SprintsPath
	existingSprints, err := s.sprintRepo.ListSprints(ctx, sprintsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list sprints: %w", err)
	}

	var inconsistencies []Inconsistency

	for _, sprintName := range existingSprints {
		select {
		case <-ctx.Done():
			return inconsistencies, ctx.Err()
		default:
		}

		sprintPath := filepath.Join(sprintsDir, sprintName)

		// Extract status from folder name prefix
		folderStatus := extractStatusFromName(sprintName)

		// Read status from .gitta/status file
		fileStatus, err := s.sprintRepo.ReadSprintStatus(ctx, sprintPath)
		hasStatusFile := err == nil

		// If status file doesn't exist, infer from folder name (not an inconsistency)
		if !hasStatusFile {
			// No status file - this is okay, status is inferred from folder name
			continue
		}

		// Compare folder status with file status
		if folderStatus != fileStatus {
			// Inconsistency detected
			// Parse folder name to get ID and description
			_, id, desc, err := parseSprintFolderNameForPlan(sprintName)
			if err != nil {
				// Skip if we can't parse
				continue
			}

			// Build expected folder name based on status file
			expectedName := buildSprintFolderName(fileStatus, id, desc)

			inconsistencies = append(inconsistencies, Inconsistency{
				SprintPath:    sprintPath,
				FolderName:    sprintName,
				FolderStatus:  folderStatus,
				StatusFile:    fileStatus,
				ExpectedName:  expectedName,
				HasStatusFile: true,
			})
		}
	}

	return inconsistencies, nil
}

// RepairInconsistencies implements SprintDoctorService.RepairInconsistencies.
func (s *sprintDoctorService) RepairInconsistencies(ctx context.Context, inconsistencies []Inconsistency) (*RepairResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &RepairResult{
		Errors: []error{},
	}

	for _, inc := range inconsistencies {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Parse folder name to get ID and description
		_, id, desc, err := parseSprintFolderNameForPlan(inc.FolderName)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Errorf("failed to parse folder name %q: %w", inc.FolderName, err))
			continue
		}

		// Rename folder to match status file
		oldPath := inc.SprintPath

		if err := s.sprintRepo.RenameSprintWithPrefix(ctx, oldPath, inc.StatusFile, id, desc); err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Errorf("failed to rename %q to %q: %w", inc.FolderName, inc.ExpectedName, err))
			continue
		}

		result.RepairedCount++
	}

	return result, nil
}

// extractStatusFromName extracts sprint status from folder name prefix.
func extractStatusFromName(name string) core.SprintStatus {
	if len(name) == 0 {
		return core.StatusActive
	}

	switch name[0] {
	case '!':
		return core.StatusActive
	case '+':
		return core.StatusReady
	case '@':
		return core.StatusPlanning
	case '~':
		return core.StatusArchived
	default:
		return core.StatusActive // Default for backward compatibility
	}
}
