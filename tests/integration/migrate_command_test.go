package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/internal/services"
)

func TestMigrateCommand_Success(t *testing.T) {
	repoPath := setupRepo(t)
	mustMkdirPath(t, filepath.Join(repoPath, "backlog"))
	mustMkdirPath(t, filepath.Join(repoPath, "sprints"))

	svc := services.NewMigrateService()
	res, err := svc.Migrate(context.Background(), repoPath, services.MigrateOptions{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success")
	}
	expectPathExists(t, filepath.Join(repoPath, "tasks", "backlog"))
	expectPathExists(t, filepath.Join(repoPath, "tasks", "sprints"))
}

func TestMigrateCommand_DryRun(t *testing.T) {
	repoPath := setupRepo(t)
	mustMkdirPath(t, filepath.Join(repoPath, "backlog"))
	mustMkdirPath(t, filepath.Join(repoPath, "sprints"))

	svc := services.NewMigrateService()
	res, err := svc.Migrate(context.Background(), repoPath, services.MigrateOptions{DryRun: true})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected dry-run success")
	}
	// Legacy dirs remain
	expectPathExists(t, filepath.Join(repoPath, "backlog"))
	expectPathExists(t, filepath.Join(repoPath, "sprints"))
}

func mustMkdirPath(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
