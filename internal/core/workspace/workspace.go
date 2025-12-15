package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Structure describes how gitta task directories are organized within a repo.
type Structure int

const (
	// Legacy places backlog/ and sprints/ at repository root.
	Legacy Structure = iota
	// Consolidated places task folders under tasks/backlog and tasks/sprints.
	Consolidated
)

var (
	ErrInvalidRepository     = errors.New("invalid repository path")
	ErrInconsistentStructure = errors.New("inconsistent workspace structure")
)

// Paths holds resolved workspace directories for the detected structure.
type Paths struct {
	BacklogPath string
	SprintsPath string
	Structure   Structure
}

// DetectStructure determines whether the repo uses consolidated or legacy layout.
// Precedence:
// 1) tasks/backlog + tasks/sprints => Consolidated
// 2) backlog + sprints => Legacy
// 3) Neither present => default to Consolidated for new repos
// If only one of the paired dirs exists, returns ErrInconsistentStructure.
func DetectStructure(ctx context.Context, repoPath string) (Structure, error) {
	if err := ctx.Err(); err != nil {
		return Consolidated, err
	}
	if repoPath == "" {
		return Consolidated, ErrInvalidRepository
	}
	info, err := os.Stat(repoPath)
	if err != nil || !info.IsDir() {
		return Consolidated, ErrInvalidRepository
	}

	consolidatedBacklog := dirExists(filepath.Join(repoPath, "tasks", "backlog"))
	consolidatedSprints := dirExists(filepath.Join(repoPath, "tasks", "sprints"))
	legacyBacklog := dirExists(filepath.Join(repoPath, "backlog"))
	legacySprints := dirExists(filepath.Join(repoPath, "sprints"))

	consolidatedExists := consolidatedBacklog || consolidatedSprints
	legacyExists := legacyBacklog || legacySprints

	switch {
	case consolidatedExists && legacyExists:
		return Consolidated, fmt.Errorf("%w: both legacy and consolidated structures detected", ErrInconsistentStructure)
	case consolidatedExists:
		return Consolidated, nil
	case legacyExists:
		return Legacy, nil
	default:
		// No structure found; default to new consolidated layout.
		return Consolidated, nil
	}
}

// ResolveBacklogPath returns the backlog directory for the given structure.
func ResolveBacklogPath(repoPath string, structure Structure) string {
	if structure == Consolidated {
		return filepath.Join(repoPath, "tasks", "backlog")
	}
	return filepath.Join(repoPath, "backlog")
}

// ResolveSprintsPath returns the sprints parent directory for the given structure.
func ResolveSprintsPath(repoPath string, structure Structure) string {
	if structure == Consolidated {
		return filepath.Join(repoPath, "tasks", "sprints")
	}
	return filepath.Join(repoPath, "sprints")
}

// ResolveSprintPath returns the directory for a specific sprint.
func ResolveSprintPath(repoPath, sprintName string, structure Structure) string {
	return filepath.Join(ResolveSprintsPath(repoPath, structure), sprintName)
}

// BuildPaths resolves all workspace paths for the detected structure.
func BuildPaths(repoPath string, structure Structure) Paths {
	return Paths{
		BacklogPath: ResolveBacklogPath(repoPath, structure),
		SprintsPath: ResolveSprintsPath(repoPath, structure),
		Structure:   structure,
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
