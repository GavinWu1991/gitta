package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInitService_Initialize_Defaults(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".git"))

	svc := NewInitService()
	res, err := svc.Initialize(context.Background(), dir, InitOptions{})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	expectExists(t, filepath.Join(dir, "tasks", "sprints", "Sprint-01"))
	expectExists(t, filepath.Join(dir, "tasks", "backlog"))
	expectExists(t, filepath.Join(dir, "tasks", "sprints", "Sprint-01", "US-001.md"))
	expectExists(t, filepath.Join(dir, "tasks", "backlog", "US-002.md"))

	if res == nil || res.SprintDir == "" || res.BacklogDir == "" {
		t.Fatalf("expected result with directories, got %+v", res)
	}
	if len(res.BackupPaths) != 0 {
		t.Fatalf("expected no backups, got %v", res.BackupPaths)
	}
}

func TestInitService_Initialize_ExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".git"))
	mustMkdir(t, filepath.Join(dir, "tasks", "sprints", "Sprint-01"))

	svc := NewInitService()
	_, err := svc.Initialize(context.Background(), dir, InitOptions{})
	if err == nil {
		t.Fatalf("expected error for existing workspace without force")
	}
	if err != nil && err.Error() == "" {
		t.Fatalf("expected descriptive error, got empty")
	}
}

func TestInitService_Initialize_ForceBackups(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".git"))
	mustMkdir(t, filepath.Join(dir, "tasks", "sprints", "Sprint-01"))
	mustMkdir(t, filepath.Join(dir, "tasks", "backlog"))

	svc := &initService{
		templates: initTemplateFS,
		now: func() time.Time {
			return time.Unix(1700000000, 0)
		},
	}

	res, err := svc.Initialize(context.Background(), dir, InitOptions{Force: true, ExampleSprint: "Sprint-01"})
	if err != nil {
		t.Fatalf("Initialize force: %v", err)
	}
	if len(res.BackupPaths) != 2 {
		t.Fatalf("expected 2 backups, got %d (%v)", len(res.BackupPaths), res.BackupPaths)
	}

	for _, backup := range res.BackupPaths {
		if _, err := os.Stat(backup); err != nil {
			t.Fatalf("expected backup %s to exist: %v", backup, err)
		}
	}

	expectExists(t, filepath.Join(dir, "tasks", "sprints", "Sprint-01", "US-001.md"))
	expectExists(t, filepath.Join(dir, "tasks", "backlog", "US-002.md"))
}

func TestInitService_Initialize_InvalidSprintName(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".git"))

	svc := NewInitService()
	_, err := svc.Initialize(context.Background(), dir, InitOptions{ExampleSprint: "../bad"})
	if err == nil {
		t.Fatalf("expected invalid sprint name error")
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func expectExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}
