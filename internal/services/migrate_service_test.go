package services

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrateService_Migrate_Success(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "backlog"))
	mustMkdir(t, filepath.Join(repo, "sprints"))

	svc := NewMigrateService()
	res, err := svc.Migrate(context.Background(), repo, MigrateOptions{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success")
	}
	expectExists(t, filepath.Join(repo, "tasks", "backlog"))
	expectExists(t, filepath.Join(repo, "tasks", "sprints"))
	if dirExists(filepath.Join(repo, "backlog")) || dirExists(filepath.Join(repo, "sprints")) {
		t.Fatalf("legacy directories should be moved")
	}
}

func TestMigrateService_Migrate_DryRun(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "backlog"))
	mustMkdir(t, filepath.Join(repo, "sprints"))

	svc := NewMigrateService()
	res, err := svc.Migrate(context.Background(), repo, MigrateOptions{DryRun: true})
	if err != nil {
		t.Fatalf("dry-run migrate: %v", err)
	}
	if !res.Success || len(res.MovedDirectories) == 0 {
		t.Fatalf("expected dry-run success with planned moves")
	}
	// No moves should have happened.
	if !dirExists(filepath.Join(repo, "backlog")) || !dirExists(filepath.Join(repo, "sprints")) {
		t.Fatalf("dry-run should not move directories")
	}
}

func TestMigrateService_Migrate_ConflictWithoutForce(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "backlog"))
	mustMkdir(t, filepath.Join(repo, "sprints"))
	mustMkdir(t, filepath.Join(repo, "tasks", "backlog"))

	svc := NewMigrateService()
	_, err := svc.Migrate(context.Background(), repo, MigrateOptions{})
	if err == nil {
		t.Fatalf("expected conflict error")
	}
}

func TestMigrateService_Migrate_ForceBackups(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "backlog"))
	mustMkdir(t, filepath.Join(repo, "sprints"))
	mustMkdir(t, filepath.Join(repo, "tasks", "backlog"))

	svc := NewMigrateService()
	res, err := svc.Migrate(context.Background(), repo, MigrateOptions{Force: true})
	if err != nil {
		t.Fatalf("migrate force: %v", err)
	}
	if len(res.BackupPaths) == 0 {
		t.Fatalf("expected backups created")
	}
}

func TestMigrateService_Migrate_AlreadyConsolidated(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "tasks", "backlog"))
	mustMkdir(t, filepath.Join(repo, "tasks", "sprints"))

	svc := NewMigrateService()
	_, err := svc.Migrate(context.Background(), repo, MigrateOptions{})
	if err == nil {
		t.Fatalf("expected already consolidated error")
	}
}
