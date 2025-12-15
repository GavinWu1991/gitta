package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/internal/services"
)

func TestInitServiceIntegration_CreatesWorkspace(t *testing.T) {
	repoPath := setupRepo(t)
	svc := services.NewInitService()

	result, err := svc.Initialize(context.Background(), repoPath, services.InitOptions{
		ExampleSprint: "Sprint-09",
	})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	expectPathExists(t, filepath.Join(repoPath, "tasks", "sprints", "Sprint-09"))
	expectPathExists(t, filepath.Join(repoPath, "tasks", "backlog"))
	expectPathExists(t, filepath.Join(repoPath, "tasks", "sprints", "Sprint-09", "US-001.md"))
	expectPathExists(t, filepath.Join(repoPath, "tasks", "backlog", "US-002.md"))

	if len(result.Created) != 2 {
		t.Fatalf("expected 2 created files, got %d", len(result.Created))
	}
}

func expectPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}
