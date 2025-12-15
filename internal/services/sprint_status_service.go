package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/core/workspace"
)

// SprintActivationResult contains the result of activating a sprint.
type SprintActivationResult struct {
	Activated *core.Sprint // Sprint that was activated
	Archived  *core.Sprint // Sprint that was archived (nil if none)
}

// SprintStatusService handles sprint status transitions.
type SprintStatusService interface {
	// ActivateSprint activates a sprint by transitioning it from Ready or Planning status to Active,
	// automatically archiving any currently active sprint.
	ActivateSprint(ctx context.Context, sprintID string) (*SprintActivationResult, error)
}

type sprintStatusService struct {
	sprintRepo core.SprintRepository
	repoPath   string
}

// NewSprintStatusService creates a new SprintStatusService instance.
func NewSprintStatusService(sprintRepo core.SprintRepository, repoPath string) SprintStatusService {
	return &sprintStatusService{
		sprintRepo: sprintRepo,
		repoPath:   repoPath,
	}
}

// ActivateSprint implements SprintStatusService.ActivateSprint.
func (s *sprintStatusService) ActivateSprint(ctx context.Context, sprintID string) (*SprintActivationResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	paths, err := resolveWorkspacePaths(ctx, s.repoPath)
	if err != nil {
		return nil, err
	}
	sprintsDir := paths.SprintsPath

	// Resolve sprint by ID (must be in Ready or Planning status)
	targetSprint, err := s.sprintRepo.ResolveSprintByID(ctx, sprintsDir, sprintID)
	if err != nil {
		return nil, fmt.Errorf("sprint %q not found in Ready or Planning status: %w", sprintID, err)
	}

	// Read current status
	currentStatus, err := s.sprintRepo.ReadSprintStatus(ctx, targetSprint.DirectoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sprint status: %w", err)
	}

	// Validate transition is allowed
	if err := core.ValidateTransition(currentStatus, core.StatusActive); err != nil {
		return nil, fmt.Errorf("cannot activate sprint: %w", err)
	}

	// Find current active sprint (if any)
	var archivedSprint *core.Sprint
	activeSprint, err := s.sprintRepo.FindActiveSprint(ctx, sprintsDir)
	if err == nil && activeSprint != nil {
		// Archive the currently active sprint
		// Update status file first (truth layer)
		if err := s.sprintRepo.WriteSprintStatus(ctx, activeSprint.DirectoryPath, core.StatusArchived); err != nil {
			return nil, fmt.Errorf("failed to update status for active sprint: %w", err)
		}

		// Parse folder name to get ID and description
		folderName := filepath.Base(activeSprint.DirectoryPath)
		_, id, desc, err := parseSprintFolderName(folderName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse active sprint folder name: %w", err)
		}

		// Rename folder with archived prefix
		if err := s.sprintRepo.RenameSprintWithPrefix(ctx, activeSprint.DirectoryPath, core.StatusArchived, id, desc); err != nil {
			// Status file already updated, but folder rename failed
			// This is recoverable via doctor command
			return nil, fmt.Errorf("failed to archive active sprint (status updated, but folder rename failed): %w", err)
		}

		archivedSprint = &core.Sprint{
			Name:          id,
			DirectoryPath: workspace.ResolveSprintPath(s.repoPath, core.StatusArchived.Prefix()+id+getDescSuffix(desc), paths.Structure),
		}
	}

	// Activate target sprint
	// Update status file first (truth layer)
	if err := s.sprintRepo.WriteSprintStatus(ctx, targetSprint.DirectoryPath, core.StatusActive); err != nil {
		return nil, fmt.Errorf("failed to update status for target sprint: %w", err)
	}

	// Parse folder name to get ID and description
	folderName := filepath.Base(targetSprint.DirectoryPath)
	_, id, desc, err := parseSprintFolderName(folderName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target sprint folder name: %w", err)
	}

	// Rename folder with active prefix
	if err := s.sprintRepo.RenameSprintWithPrefix(ctx, targetSprint.DirectoryPath, core.StatusActive, id, desc); err != nil {
		// Status file already updated, but folder rename failed
		// This is recoverable via doctor command
		return nil, fmt.Errorf("failed to activate sprint (status updated, but folder rename failed): %w", err)
	}

	// Calculate new active path after rename
	newActivePath := workspace.ResolveSprintPath(s.repoPath, core.StatusActive.Prefix()+id+getDescSuffix(desc), paths.Structure)

	// Update Current link to point to newly activated sprint
	if err := s.sprintRepo.UpdateCurrentLink(ctx, sprintsDir, newActivePath); err != nil {
		// Sprint is activated, but Current link update failed
		// This is recoverable via doctor command
		return nil, fmt.Errorf("failed to update Current link (sprint activated, but link update failed): %w", err)
	}

	// Update archived sprint path if it was archived
	if archivedSprint != nil {
		archivedFolderName := filepath.Base(archivedSprint.DirectoryPath)
		_, archivedID, archivedDesc, _ := parseSprintFolderName(archivedFolderName)
		archivedSprint.DirectoryPath = filepath.Join(filepath.Dir(archivedSprint.DirectoryPath), core.StatusArchived.Prefix()+archivedID+getDescSuffix(archivedDesc))
	}

	return &SprintActivationResult{
		Activated: &core.Sprint{
			Name:          id,
			DirectoryPath: newActivePath,
		},
		Archived: archivedSprint,
	}, nil
}

// parseSprintFolderName parses a sprint folder name and extracts the status prefix, ID, and description.
// This is a duplicate of filesystem.ParseFolderName to avoid circular dependency.
func parseSprintFolderName(name string) (core.SprintStatus, string, string, error) {
	if len(name) == 0 {
		return core.StatusActive, "", "", fmt.Errorf("folder name cannot be empty")
	}

	// Extract status from prefix
	var status core.SprintStatus
	switch name[0] {
	case '!':
		status = core.StatusActive
	case '+':
		status = core.StatusReady
	case '@':
		status = core.StatusPlanning
	case '~':
		status = core.StatusArchived
	default:
		return core.StatusActive, "", "", fmt.Errorf("folder name must start with a valid status prefix (!, +, @, ~)")
	}

	// Remove prefix
	rest := name[1:]

	// Find last underscore
	lastUnderscore := strings.LastIndex(rest, "_")
	if lastUnderscore == -1 {
		return status, rest, "", nil
	}

	// Check if part after last underscore is numeric
	afterLastUnderscore := rest[lastUnderscore+1:]
	isNumeric := true
	for _, r := range afterLastUnderscore {
		if r < '0' || r > '9' {
			isNumeric = false
			break
		}
	}

	if isNumeric {
		// Numeric part is part of ID
		return status, rest, "", nil
	}

	// Split at last underscore
	sprintID := rest[:lastUnderscore]
	description := rest[lastUnderscore+1:]

	return status, sprintID, description, nil
}

// getDescSuffix returns the description suffix for folder name construction.
func getDescSuffix(desc string) string {
	if desc == "" {
		return ""
	}
	return "_" + desc
}
