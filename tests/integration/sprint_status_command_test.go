package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
)

func TestSprintActivation_ReadyToActive(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create a Ready sprint
	readySprintDir := filepath.Join(sprintsDir, "+Sprint_24_Login")
	os.MkdirAll(readySprintDir, 0755)
	os.MkdirAll(filepath.Join(readySprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(readySprintDir, ".gitta", "status"), []byte("ready\n"), 0644)

	sprintRepo := filesystem.NewDefaultRepository()
	statusService := services.NewSprintStatusService(sprintRepo, testRepo)

	// Activate the sprint
	result, err := statusService.ActivateSprint(ctx, "24")
	if err != nil {
		t.Fatalf("ActivateSprint() error = %v", err)
	}

	// Verify sprint was activated
	if result.Activated == nil {
		t.Fatal("ActivateSprint() result.Activated is nil")
	}

	// Verify folder was renamed to Active prefix
	activeSprintDir := filepath.Join(sprintsDir, "!Sprint_24_Login")
	expectPathExists(t, activeSprintDir)

	// Verify status file was updated
	status, err := sprintRepo.ReadSprintStatus(ctx, activeSprintDir)
	if err != nil {
		t.Fatalf("ReadSprintStatus() error = %v", err)
	}
	if status != core.StatusActive {
		t.Errorf("ReadSprintStatus() = %v, want %v", status, core.StatusActive)
	}

	// Verify Current link points to active sprint
	currentPath, _, err := filesystem.ReadCurrentSprintLink(sprintsDir)
	if err != nil {
		t.Fatalf("ReadCurrentSprintLink() error = %v", err)
	}
	absCurrent, _ := filepath.Abs(currentPath)
	absActive, _ := filepath.Abs(activeSprintDir)
	if absCurrent != absActive {
		t.Errorf("Current link = %q, want %q", absCurrent, absActive)
	}
}

func TestSprintActivation_PlanningToActive(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create a Planning sprint
	planningSprintDir := filepath.Join(sprintsDir, "@Sprint_25_Payment")
	os.MkdirAll(planningSprintDir, 0755)
	os.MkdirAll(filepath.Join(planningSprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(planningSprintDir, ".gitta", "status"), []byte("planning\n"), 0644)

	sprintRepo := filesystem.NewDefaultRepository()
	statusService := services.NewSprintStatusService(sprintRepo, testRepo)

	// Activate the sprint
	result, err := statusService.ActivateSprint(ctx, "Sprint_25")
	if err != nil {
		t.Fatalf("ActivateSprint() error = %v", err)
	}

	// Verify sprint was activated
	if result.Activated == nil {
		t.Fatal("ActivateSprint() result.Activated is nil")
	}

	// Verify folder was renamed to Active prefix
	activeSprintDir := filepath.Join(sprintsDir, "!Sprint_25_Payment")
	expectPathExists(t, activeSprintDir)

	// Verify status file was updated
	status, err := sprintRepo.ReadSprintStatus(ctx, activeSprintDir)
	if err != nil {
		t.Fatalf("ReadSprintStatus() error = %v", err)
	}
	if status != core.StatusActive {
		t.Errorf("ReadSprintStatus() = %v, want %v", status, core.StatusActive)
	}
}

func TestSprintActivation_WithExistingActiveSprint(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create an Active sprint
	activeSprintDir := filepath.Join(sprintsDir, "!Sprint_23_Onboarding")
	os.MkdirAll(activeSprintDir, 0755)
	os.MkdirAll(filepath.Join(activeSprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(activeSprintDir, ".gitta", "status"), []byte("active\n"), 0644)

	// Create Current link pointing to active sprint
	filesystem.CreateCurrentSprintLink(activeSprintDir, filepath.Join(sprintsDir, "Current"))

	// Create a Ready sprint
	readySprintDir := filepath.Join(sprintsDir, "+Sprint_24_Login")
	os.MkdirAll(readySprintDir, 0755)
	os.MkdirAll(filepath.Join(readySprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(readySprintDir, ".gitta", "status"), []byte("ready\n"), 0644)

	sprintRepo := filesystem.NewDefaultRepository()
	statusService := services.NewSprintStatusService(sprintRepo, testRepo)

	// Activate the Ready sprint
	result, err := statusService.ActivateSprint(ctx, "24")
	if err != nil {
		t.Fatalf("ActivateSprint() error = %v", err)
	}

	// Verify previous active sprint was archived
	if result.Archived == nil {
		t.Fatal("ActivateSprint() result.Archived is nil (expected previous active sprint to be archived)")
	}

	// Verify previous sprint folder was renamed to Archived prefix
	archivedSprintDir := filepath.Join(sprintsDir, "~Sprint_23_Onboarding")
	expectPathExists(t, archivedSprintDir)

	// Verify archived sprint status file
	archivedStatus, err := sprintRepo.ReadSprintStatus(ctx, archivedSprintDir)
	if err != nil {
		t.Fatalf("ReadSprintStatus() error = %v", err)
	}
	if archivedStatus != core.StatusArchived {
		t.Errorf("Archived sprint status = %v, want %v", archivedStatus, core.StatusArchived)
	}

	// Verify new sprint was activated
	newActiveSprintDir := filepath.Join(sprintsDir, "!Sprint_24_Login")
	expectPathExists(t, newActiveSprintDir)

	// Verify Current link points to new active sprint
	currentPath, _, err := filesystem.ReadCurrentSprintLink(sprintsDir)
	if err != nil {
		t.Fatalf("ReadCurrentSprintLink() error = %v", err)
	}
	absCurrent, _ := filepath.Abs(currentPath)
	absActive, _ := filepath.Abs(newActiveSprintDir)
	if absCurrent != absActive {
		t.Errorf("Current link = %q, want %q", absCurrent, absActive)
	}
}

func TestSprintActivation_SprintNotFound(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	sprintRepo := filesystem.NewDefaultRepository()
	statusService := services.NewSprintStatusService(sprintRepo, testRepo)

	// Try to activate non-existent sprint
	_, err := statusService.ActivateSprint(ctx, "99")
	if err == nil {
		t.Fatal("ActivateSprint() expected error for non-existent sprint")
	}
	if err != core.ErrSprintNotFound && err.Error() == "" {
		t.Errorf("ActivateSprint() error = %v, want ErrSprintNotFound or error message", err)
	}
}

func TestSprintActivation_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create an Archived sprint (cannot be activated)
	archivedSprintDir := filepath.Join(sprintsDir, "~Sprint_23_Onboarding")
	os.MkdirAll(archivedSprintDir, 0755)
	os.MkdirAll(filepath.Join(archivedSprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(archivedSprintDir, ".gitta", "status"), []byte("archived\n"), 0644)

	sprintRepo := filesystem.NewDefaultRepository()
	statusService := services.NewSprintStatusService(sprintRepo, testRepo)

	// Try to activate archived sprint
	_, err := statusService.ActivateSprint(ctx, "23")
	if err == nil {
		t.Fatal("ActivateSprint() expected error for archived sprint")
	}
	// Should get validation error about invalid transition
}

func TestSprintPlanCommand_Success(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	sprintRepo := filesystem.NewDefaultRepository()
	planService := services.NewSprintPlanService(sprintRepo, testRepo)

	req := services.CreatePlanningSprintRequest{
		Description: "Dashboard",
	}

	sprint, err := planService.CreatePlanningSprint(ctx, req)
	if err != nil {
		t.Fatalf("CreatePlanningSprint() error = %v", err)
	}

	// Verify sprint was created with Planning prefix
	expectedDir := filepath.Join(sprintsDir, "@"+sprint.Name+"_Dashboard")
	if sprint.DirectoryPath != expectedDir {
		t.Errorf("CreatePlanningSprint() directory = %q, want %q", sprint.DirectoryPath, expectedDir)
	}

	// Verify directory exists
	expectPathExists(t, expectedDir)

	// Verify status file contains "planning"
	status, err := sprintRepo.ReadSprintStatus(ctx, expectedDir)
	if err != nil {
		t.Fatalf("ReadSprintStatus() error = %v", err)
	}
	if status != core.StatusPlanning {
		t.Errorf("ReadSprintStatus() = %v, want %v", status, core.StatusPlanning)
	}
}

func TestSprintPlanCommand_DuplicateID(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create existing sprint with same ID
	existingSprintDir := filepath.Join(sprintsDir, "!Sprint_25_Payment")
	os.MkdirAll(existingSprintDir, 0755)

	sprintRepo := filesystem.NewDefaultRepository()
	planService := services.NewSprintPlanService(sprintRepo, testRepo)

	req := services.CreatePlanningSprintRequest{
		ID:          "Sprint_25",
		Description: "New Feature",
	}

	_, err := planService.CreatePlanningSprint(ctx, req)
	if err == nil {
		t.Fatal("CreatePlanningSprint() expected error for duplicate ID")
	}
}

func TestSprintPlanCommand_InvalidName(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	sprintRepo := filesystem.NewDefaultRepository()
	planService := services.NewSprintPlanService(sprintRepo, testRepo)

	req := services.CreatePlanningSprintRequest{
		Description: "Feature! Important",
	}

	_, err := planService.CreatePlanningSprint(ctx, req)
	if err == nil {
		t.Fatal("CreatePlanningSprint() expected error for invalid name")
	}
}

func TestDoctorCommand_DetectionMode(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create inconsistent sprint: Active prefix, Ready status file
	sprintDir := filepath.Join(sprintsDir, "!Sprint_24_Login")
	os.MkdirAll(sprintDir, 0755)
	os.MkdirAll(filepath.Join(sprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(sprintDir, ".gitta", "status"), []byte("ready\n"), 0644)

	// Create consistent sprint for comparison
	consistentDir := filepath.Join(sprintsDir, "+Sprint_25_Payment")
	os.MkdirAll(consistentDir, 0755)
	os.MkdirAll(filepath.Join(consistentDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(consistentDir, ".gitta", "status"), []byte("ready\n"), 0644)

	sprintRepo := filesystem.NewDefaultRepository()
	doctorService := services.NewSprintDoctorService(sprintRepo, testRepo)

	inconsistencies, err := doctorService.DetectInconsistencies(ctx)
	if err != nil {
		t.Fatalf("DetectInconsistencies() error = %v", err)
	}

	if len(inconsistencies) != 1 {
		t.Errorf("DetectInconsistencies() count = %d, want 1", len(inconsistencies))
	}

	if len(inconsistencies) > 0 {
		inc := inconsistencies[0]
		if inc.FolderStatus == inc.StatusFile {
			t.Error("DetectInconsistencies() inconsistency should have different folder and file status")
		}
		if inc.ExpectedName != "+Sprint_24_Login" {
			t.Errorf("DetectInconsistencies() expected name = %q, want %q", inc.ExpectedName, "+Sprint_24_Login")
		}
	}
}

func TestDoctorCommand_RepairMode(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

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
	expectPathExists(t, expectedDir)

	// Verify old folder doesn't exist
	if _, err := os.Stat(sprintDir); err == nil {
		t.Errorf("Old sprint directory %q still exists after repair", sprintDir)
	}
}

func TestDoctorCommand_CurrentLinkValidation(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create Active sprint
	activeSprintDir := filepath.Join(sprintsDir, "!Sprint_24_Login")
	os.MkdirAll(activeSprintDir, 0755)
	os.MkdirAll(filepath.Join(activeSprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(activeSprintDir, ".gitta", "status"), []byte("active\n"), 0644)

	// Create Current link pointing to active sprint
	filesystem.CreateCurrentSprintLink(activeSprintDir, filepath.Join(sprintsDir, "Current"))

	sprintRepo := filesystem.NewDefaultRepository()
	doctorService := services.NewSprintDoctorService(sprintRepo, testRepo)

	// Detect inconsistencies (should find none, and Current link should be valid)
	inconsistencies, err := doctorService.DetectInconsistencies(ctx)
	if err != nil {
		t.Fatalf("DetectInconsistencies() error = %v", err)
	}

	if len(inconsistencies) != 0 {
		t.Errorf("DetectInconsistencies() count = %d, want 0", len(inconsistencies))
	}

	// Verify Current link is readable
	currentPath, _, err := filesystem.ReadCurrentSprintLink(sprintsDir)
	if err != nil {
		t.Fatalf("ReadCurrentSprintLink() error = %v", err)
	}
	absCurrent, _ := filepath.Abs(currentPath)
	absActive, _ := filepath.Abs(activeSprintDir)
	if absCurrent != absActive {
		t.Errorf("Current link = %q, want %q", absCurrent, absActive)
	}
}

func TestSprintActivation_AlreadyActive(t *testing.T) {
	ctx := context.Background()
	testRepo := setupRepo(t)
	defer os.RemoveAll(testRepo)

	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create an Active sprint
	activeSprintDir := filepath.Join(sprintsDir, "!Sprint_24_Login")
	os.MkdirAll(activeSprintDir, 0755)
	os.MkdirAll(filepath.Join(activeSprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(activeSprintDir, ".gitta", "status"), []byte("active\n"), 0644)

	sprintRepo := filesystem.NewDefaultRepository()
	statusService := services.NewSprintStatusService(sprintRepo, testRepo)

	// Try to activate already active sprint
	// This should fail because ResolveSprintByID only finds Ready or Planning sprints
	_, err := statusService.ActivateSprint(ctx, "24")
	if err == nil {
		t.Fatal("ActivateSprint() expected error for already active sprint")
	}
}
