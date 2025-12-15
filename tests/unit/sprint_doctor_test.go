package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestDetectInconsistencies(t *testing.T) {
	ctx := context.Background()
	testRepo := t.TempDir()
	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	tests := []struct {
		name              string
		setup             func()
		wantCount         int
		wantInconsistency bool
	}{
		{
			name: "no inconsistencies",
			setup: func() {
				// Create consistent sprint
				sprintDir := filepath.Join(sprintsDir, "!Sprint_24_Login")
				os.MkdirAll(sprintDir, 0755)
				os.MkdirAll(filepath.Join(sprintDir, ".gitta"), 0755)
				os.WriteFile(filepath.Join(sprintDir, ".gitta", "status"), []byte("active\n"), 0644)
			},
			wantCount: 0,
		},
		{
			name: "folder prefix mismatch",
			setup: func() {
				// Create sprint with Active prefix but Ready status file
				sprintDir := filepath.Join(sprintsDir, "!Sprint_24_Login")
				os.MkdirAll(sprintDir, 0755)
				os.MkdirAll(filepath.Join(sprintDir, ".gitta"), 0755)
				os.WriteFile(filepath.Join(sprintDir, ".gitta", "status"), []byte("ready\n"), 0644)
			},
			wantCount:         1,
			wantInconsistency: true,
		},
		{
			name: "missing status file",
			setup: func() {
				// Create sprint without status file (not an inconsistency - inferred from folder)
				sprintDir := filepath.Join(sprintsDir, "+Sprint_25_Payment")
				os.MkdirAll(sprintDir, 0755)
			},
			wantCount: 0, // Missing status file is not an inconsistency
		},
		{
			name: "multiple inconsistencies",
			setup: func() {
				// Sprint 1: Active prefix, Ready status
				sprint1Dir := filepath.Join(sprintsDir, "!Sprint_24_Login")
				os.MkdirAll(sprint1Dir, 0755)
				os.MkdirAll(filepath.Join(sprint1Dir, ".gitta"), 0755)
				os.WriteFile(filepath.Join(sprint1Dir, ".gitta", "status"), []byte("ready\n"), 0644)

				// Sprint 2: Planning prefix, Archived status
				sprint2Dir := filepath.Join(sprintsDir, "@Sprint_25_Payment")
				os.MkdirAll(sprint2Dir, 0755)
				os.MkdirAll(filepath.Join(sprint2Dir, ".gitta"), 0755)
				os.WriteFile(filepath.Join(sprint2Dir, ".gitta", "status"), []byte("archived\n"), 0644)
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up and setup
			os.RemoveAll(sprintsDir)
			os.MkdirAll(sprintsDir, 0755)
			tt.setup()

			sprintRepo := filesystem.NewDefaultRepository()
			doctorService := services.NewSprintDoctorService(sprintRepo, testRepo)

			inconsistencies, err := doctorService.DetectInconsistencies(ctx)
			if err != nil {
				t.Fatalf("DetectInconsistencies() error = %v", err)
			}

			if len(inconsistencies) != tt.wantCount {
				t.Errorf("DetectInconsistencies() count = %d, want %d", len(inconsistencies), tt.wantCount)
			}

			if tt.wantInconsistency && len(inconsistencies) > 0 {
				inc := inconsistencies[0]
				if inc.FolderStatus == inc.StatusFile {
					t.Errorf("DetectInconsistencies() inconsistency should have different folder and file status")
				}
			}
		})
	}
}

func TestRepairInconsistencies(t *testing.T) {
	ctx := context.Background()
	testRepo := t.TempDir()
	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create inconsistent sprint: Active prefix, Ready status file
	sprintDir := filepath.Join(sprintsDir, "!Sprint_24_Login")
	os.MkdirAll(sprintDir, 0755)
	os.MkdirAll(filepath.Join(sprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(sprintDir, ".gitta", "status"), []byte("ready\n"), 0644)

	sprintRepo := filesystem.NewDefaultRepository()
	doctorService := services.NewSprintDoctorService(sprintRepo, testRepo)

	// Detect inconsistencies
	inconsistencies, err := doctorService.DetectInconsistencies(ctx)
	if err != nil {
		t.Fatalf("DetectInconsistencies() error = %v", err)
	}

	if len(inconsistencies) != 1 {
		t.Fatalf("DetectInconsistencies() count = %d, want 1", len(inconsistencies))
	}

	// Repair inconsistencies
	result, err := doctorService.RepairInconsistencies(ctx, inconsistencies)
	if err != nil {
		t.Fatalf("RepairInconsistencies() error = %v", err)
	}

	if result.RepairedCount != 1 {
		t.Errorf("RepairInconsistencies() repaired = %d, want 1", result.RepairedCount)
	}

	// Verify folder was renamed
	expectedDir := filepath.Join(sprintsDir, "+Sprint_24_Login")
	if _, err := os.Stat(expectedDir); err != nil {
		t.Errorf("Expected sprint directory %q does not exist after repair", expectedDir)
	}

	// Verify old folder doesn't exist
	if _, err := os.Stat(sprintDir); err == nil {
		t.Errorf("Old sprint directory %q still exists after repair", sprintDir)
	}

	// Verify status file still has Ready status
	status, err := sprintRepo.ReadSprintStatus(ctx, expectedDir)
	if err != nil {
		t.Fatalf("ReadSprintStatus() error = %v", err)
	}
	if status != core.StatusReady {
		t.Errorf("ReadSprintStatus() = %v, want %v", status, core.StatusReady)
	}
}

func TestRepairInconsistencies_Multiple(t *testing.T) {
	ctx := context.Background()
	testRepo := t.TempDir()
	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create multiple inconsistent sprints
	sprint1Dir := filepath.Join(sprintsDir, "!Sprint_24_Login")
	os.MkdirAll(sprint1Dir, 0755)
	os.MkdirAll(filepath.Join(sprint1Dir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(sprint1Dir, ".gitta", "status"), []byte("ready\n"), 0644)

	sprint2Dir := filepath.Join(sprintsDir, "@Sprint_25_Payment")
	os.MkdirAll(sprint2Dir, 0755)
	os.MkdirAll(filepath.Join(sprint2Dir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(sprint2Dir, ".gitta", "status"), []byte("archived\n"), 0644)

	sprintRepo := filesystem.NewDefaultRepository()
	doctorService := services.NewSprintDoctorService(sprintRepo, testRepo)

	// Detect inconsistencies
	inconsistencies, err := doctorService.DetectInconsistencies(ctx)
	if err != nil {
		t.Fatalf("DetectInconsistencies() error = %v", err)
	}

	if len(inconsistencies) != 2 {
		t.Fatalf("DetectInconsistencies() count = %d, want 2", len(inconsistencies))
	}

	// Repair inconsistencies
	result, err := doctorService.RepairInconsistencies(ctx, inconsistencies)
	if err != nil {
		t.Fatalf("RepairInconsistencies() error = %v", err)
	}

	if result.RepairedCount != 2 {
		t.Errorf("RepairInconsistencies() repaired = %d, want 2", result.RepairedCount)
	}

	// Verify both folders were renamed
	expected1Dir := filepath.Join(sprintsDir, "+Sprint_24_Login")
	expected2Dir := filepath.Join(sprintsDir, "~Sprint_25_Payment")

	if _, err := os.Stat(expected1Dir); err != nil {
		t.Errorf("Expected sprint directory %q does not exist", expected1Dir)
	}
	if _, err := os.Stat(expected2Dir); err != nil {
		t.Errorf("Expected sprint directory %q does not exist", expected2Dir)
	}
}
