package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gavin/gitta/internal/core/workspace"
)

// MigrateOptions configures migration behavior.
type MigrateOptions struct {
	Force  bool // Overwrite existing targets (creates backups)
	DryRun bool // Simulate without making changes
}

// MigrateResult reports migration outcome.
type MigrateResult struct {
	Success          bool
	MovedDirectories []string
	BackupPaths      []string
	Errors           []error
}

// MigrateService handles migration from legacy layout to consolidated tasks/ structure.
type MigrateService interface {
	CanMigrate(ctx context.Context, repoPath string) (bool, error)
	Migrate(ctx context.Context, repoPath string, opts MigrateOptions) (*MigrateResult, error)
}

type migrateService struct{}

// NewMigrateService creates a MigrateService.
func NewMigrateService() MigrateService {
	return &migrateService{}
}

// CanMigrate reports whether repoPath is a Git repo using legacy structure.
func (m *migrateService) CanMigrate(ctx context.Context, repoPath string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if !isGitRepo(repoPath) {
		return false, ErrNotGitRepository
	}
	structure, err := workspace.DetectStructure(ctx, repoPath)
	if err != nil {
		return false, err
	}
	return structure == workspace.Legacy, nil
}

// Migrate performs migration from legacy to consolidated structure.
func (m *migrateService) Migrate(ctx context.Context, repoPath string, opts MigrateOptions) (*MigrateResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !isGitRepo(repoPath) {
		return nil, ErrNotGitRepository
	}

	legacyBacklog := filepath.Join(repoPath, "backlog")
	legacySprints := filepath.Join(repoPath, "sprints")

	if !dirExists(legacyBacklog) || !dirExists(legacySprints) {
		return nil, fmt.Errorf("legacy structure incomplete: missing backlog or sprints")
	}

	structure, err := workspace.DetectStructure(ctx, repoPath)
	if err != nil && !errors.Is(err, workspace.ErrInconsistentStructure) {
		return nil, err
	}
	if structure == workspace.Consolidated && err == nil {
		return nil, ErrAlreadyConsolidated
	}

	targetTasks := filepath.Join(repoPath, "tasks")
	targetBacklog := filepath.Join(targetTasks, "backlog")
	targetSprints := filepath.Join(targetTasks, "sprints")

	conflicts := existingTargets([]string{targetBacklog, targetSprints})
	result := &MigrateResult{
		MovedDirectories: []string{},
		BackupPaths:      []string{},
		Errors:           []error{},
	}

	if len(conflicts) > 0 && !opts.Force {
		return result, ErrMigrationConflict
	}

	if opts.DryRun {
		result.Success = true
		result.MovedDirectories = []string{"backlog -> tasks/backlog", "sprints -> tasks/sprints"}
		return result, nil
	}

	// Prepare tasks dir
	if err := os.MkdirAll(targetTasks, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}

	// Backup conflicts if force
	if len(conflicts) > 0 && opts.Force {
		for _, c := range conflicts {
			backup := fmt.Sprintf("%s.backup-%d", c, time.Now().Unix())
			if err := os.Rename(c, backup); err != nil {
				return nil, fmt.Errorf("failed to backup %s: %w", c, err)
			}
			result.BackupPaths = append(result.BackupPaths, backup)
		}
	}

	// Move legacy directories
	if err := os.Rename(legacyBacklog, targetBacklog); err != nil {
		return nil, fmt.Errorf("failed to move backlog: %w", err)
	}
	if err := os.Rename(legacySprints, targetSprints); err != nil {
		return nil, fmt.Errorf("failed to move sprints: %w", err)
	}

	result.Success = true
	result.MovedDirectories = []string{targetBacklog, targetSprints}
	return result, nil
}

func isGitRepo(repoPath string) bool {
	info, err := os.Stat(filepath.Join(repoPath, ".git"))
	return err == nil && info.IsDir()
}

func existingTargets(paths []string) []string {
	var found []string
	for _, p := range paths {
		if dirExists(p) {
			found = append(found, p)
		}
	}
	return found
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
