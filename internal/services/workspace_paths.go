package services

import (
	"context"

	"github.com/gavin/gitta/internal/core/workspace"
)

// resolveWorkspacePaths detects the repository structure and returns resolved paths.
func resolveWorkspacePaths(ctx context.Context, repoPath string) (workspace.Paths, error) {
	structure, err := workspace.DetectStructure(ctx, repoPath)
	if err != nil {
		return workspace.Paths{}, err
	}
	return workspace.BuildPaths(repoPath, structure), nil
}
