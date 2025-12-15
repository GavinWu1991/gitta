package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gavin/gitta/internal/core"
)

// CreatePlanningSprintRequest contains parameters for creating a planning sprint.
type CreatePlanningSprintRequest struct {
	ID          string // Optional: auto-generate if empty
	Description string // Required: sprint description/name
}

// SprintPlanService handles creation of planning sprints.
type SprintPlanService interface {
	// CreatePlanningSprint creates a new sprint in Planning status.
	CreatePlanningSprint(ctx context.Context, req CreatePlanningSprintRequest) (*core.Sprint, error)
}

type sprintPlanService struct {
	sprintRepo core.SprintRepository
	repoPath   string
}

// NewSprintPlanService creates a new SprintPlanService instance.
func NewSprintPlanService(sprintRepo core.SprintRepository, repoPath string) SprintPlanService {
	return &sprintPlanService{
		sprintRepo: sprintRepo,
		repoPath:   repoPath,
	}
}

// CreatePlanningSprint implements SprintPlanService.CreatePlanningSprint.
func (s *sprintPlanService) CreatePlanningSprint(ctx context.Context, req CreatePlanningSprintRequest) (*core.Sprint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Validate description
	if req.Description == "" {
		return nil, fmt.Errorf("sprint description cannot be empty")
	}

	// Validate description doesn't contain status prefix characters
	if err := validateSprintNameForStatusPrefix(req.Description); err != nil {
		return nil, fmt.Errorf("invalid sprint description: %w", err)
	}

	sprintsDir := filepath.Join(s.repoPath, "sprints")

	// Generate sprint ID if not provided
	sprintID := req.ID
	if sprintID == "" {
		existing, err := s.sprintRepo.ListSprints(ctx, sprintsDir)
		if err != nil {
			return nil, fmt.Errorf("failed to list existing sprints: %w", err)
		}
		sprintID = s.generateNextSprintID(existing)
	}

	// Check if sprint with this ID already exists (check all status prefixes)
	existingSprints, err := s.sprintRepo.ListSprints(ctx, sprintsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing sprints: %w", err)
	}

	for _, existingName := range existingSprints {
		// Parse existing sprint name to get ID
		_, existingID, _, err := parseSprintFolderNameForPlan(existingName)
		if err == nil && existingID == sprintID {
			return nil, fmt.Errorf("%w: sprint with ID %q already exists", core.ErrSprintExists, sprintID)
		}
		// Also check if the name without prefix matches
		// Handle cases where sprint might not have prefix yet
		if len(existingName) > 0 && (existingName[0] == '!' || existingName[0] == '+' || existingName[0] == '@' || existingName[0] == '~') {
			// Already handled by parseSprintFolderNameForPlan
		} else if existingName == sprintID {
			// Exact match without prefix
			return nil, fmt.Errorf("%w: sprint with ID %q already exists", core.ErrSprintExists, sprintID)
		}
	}

	// Build folder name with Planning prefix
	folderName := buildSprintFolderName(core.StatusPlanning, sprintID, req.Description)
	sprintDir := filepath.Join(sprintsDir, folderName)

	// Check if directory already exists
	exists, err := s.sprintRepo.SprintExists(ctx, sprintDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check if sprint exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: sprint %q already exists", core.ErrSprintExists, folderName)
	}

	// Create sprint directory (use current time as start date, empty duration for planning sprints)
	// Note: CreateSprint expects a time.Time, but for planning sprints we don't need dates yet
	// We'll use a zero time and the sprint won't have dates until activated
	var zeroTime time.Time
	sprint, err := s.sprintRepo.CreateSprint(ctx, sprintDir, sprintID, zeroTime, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create sprint directory: %w", err)
	}

	// Write Planning status to .gitta/status file
	if err := s.sprintRepo.WriteSprintStatus(ctx, sprintDir, core.StatusPlanning); err != nil {
		// Cleanup on error
		os.RemoveAll(sprintDir)
		return nil, fmt.Errorf("failed to write sprint status: %w", err)
	}

	return &core.Sprint{
		Name:          sprintID,
		DirectoryPath: sprintDir,
		CreatedAt:     sprint.CreatedAt,
		UpdatedAt:     sprint.UpdatedAt,
	}, nil
}

// generateNextSprintID generates the next sequential sprint ID based on existing sprints.
func (s *sprintPlanService) generateNextSprintID(existing []string) string {
	if len(existing) == 0 {
		return "Sprint_01"
	}

	maxNum := 0
	numberPattern := regexp.MustCompile(`Sprint[_-](\d+)`)

	for _, name := range existing {
		// Remove status prefix if present
		if len(name) > 0 && (name[0] == '!' || name[0] == '+' || name[0] == '@' || name[0] == '~') {
			name = name[1:]
		}

		// Try to extract number from "Sprint_XX" or "Sprint-XX" pattern
		matches := numberPattern.FindStringSubmatch(name)
		if len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				if num > maxNum {
					maxNum = num
				}
			}
		}
	}

	// Generate next number
	nextNum := maxNum + 1
	return fmt.Sprintf("Sprint_%02d", nextNum)
}

// validateSprintNameForStatusPrefix validates that a sprint name/description doesn't contain status prefix characters.
func validateSprintNameForStatusPrefix(name string) error {
	invalidChars := []string{"!", "+", "@", "~"}
	var found []string

	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			found = append(found, char)
		}
	}

	if len(found) > 0 {
		return fmt.Errorf("sprint name/description cannot contain status prefix characters: %s", strings.Join(found, ", "))
	}

	return nil
}

// buildSprintFolderName constructs a sprint folder name from status, ID, and description.
func buildSprintFolderName(status core.SprintStatus, id string, desc string) string {
	if desc == "" {
		return status.Prefix() + id
	}
	return status.Prefix() + id + "_" + desc
}

// parseSprintFolderNameForPlan parses a sprint folder name and extracts the status prefix, ID, and description.
// This is a duplicate of the function in sprint_status_service.go to avoid circular dependency.
func parseSprintFolderNameForPlan(name string) (core.SprintStatus, string, string, error) {
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
