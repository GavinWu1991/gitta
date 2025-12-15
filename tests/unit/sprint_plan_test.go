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

func TestSprintIDGeneration(t *testing.T) {
	ctx := context.Background()
	testRepo := t.TempDir()
	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	tests := []struct {
		name            string
		existingSprints []string
		wantID          string
	}{
		{
			name:            "no existing sprints",
			existingSprints: []string{},
			wantID:          "Sprint_01",
		},
		{
			name:            "one existing sprint",
			existingSprints: []string{"!Sprint_01_Login"},
			wantID:          "Sprint_02",
		},
		{
			name:            "multiple sprints with gaps",
			existingSprints: []string{"!Sprint_01", "+Sprint_03", "@Sprint_05"},
			wantID:          "Sprint_06",
		},
		{
			name:            "sprints with different prefixes",
			existingSprints: []string{"!Sprint_10_Active", "+Sprint_11_Ready", "@Sprint_12_Planning", "~Sprint_13_Archived"},
			wantID:          "Sprint_14",
		},
		{
			name:            "sprints with underscore format",
			existingSprints: []string{"!Sprint_24_Login"},
			wantID:          "Sprint_25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create existing sprint directories
			for _, sprintName := range tt.existingSprints {
				sprintDir := filepath.Join(sprintsDir, sprintName)
				os.MkdirAll(sprintDir, 0755)
			}

			sprintRepo := filesystem.NewDefaultRepository()
			planService := services.NewSprintPlanService(sprintRepo, testRepo)

			req := services.CreatePlanningSprintRequest{
				Description: "Test Sprint",
			}

			sprint, err := planService.CreatePlanningSprint(ctx, req)
			if err != nil {
				t.Fatalf("CreatePlanningSprint() error = %v", err)
			}

			if sprint.Name != tt.wantID {
				t.Errorf("CreatePlanningSprint() generated ID = %q, want %q", sprint.Name, tt.wantID)
			}
		})
	}
}

func TestSprintNameValidation(t *testing.T) {
	ctx := context.Background()
	testRepo := t.TempDir()
	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	tests := []struct {
		name        string
		description string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid description",
			description: "Dashboard Redesign",
			wantErr:     false,
		},
		{
			name:        "description with !",
			description: "Important! Feature",
			wantErr:     true,
			errContains: "status prefix characters",
		},
		{
			name:        "description with +",
			description: "Plus+ Feature",
			wantErr:     true,
			errContains: "status prefix characters",
		},
		{
			name:        "description with @",
			description: "At@ Mention",
			wantErr:     true,
			errContains: "status prefix characters",
		},
		{
			name:        "description with ~",
			description: "Home~ Directory",
			wantErr:     true,
			errContains: "status prefix characters",
		},
		{
			name:        "empty description",
			description: "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprintRepo := filesystem.NewDefaultRepository()
			planService := services.NewSprintPlanService(sprintRepo, testRepo)

			req := services.CreatePlanningSprintRequest{
				ID:          "Sprint_99",
				Description: tt.description,
			}

			_, err := planService.CreatePlanningSprint(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePlanningSprint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !errorContains(err.Error(), tt.errContains) {
					t.Errorf("CreatePlanningSprint() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestCreatePlanningSprint_DuplicateID(t *testing.T) {
	ctx := context.Background()
	testRepo := t.TempDir()
	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	// Create existing sprint with same ID (Active status)
	existingSprintDir := filepath.Join(sprintsDir, "!Sprint_25_Payment")
	os.MkdirAll(existingSprintDir, 0755)
	os.MkdirAll(filepath.Join(existingSprintDir, ".gitta"), 0755)
	os.WriteFile(filepath.Join(existingSprintDir, ".gitta", "status"), []byte("active\n"), 0644)

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
	if err != core.ErrSprintExists && !errorContains(err.Error(), "already exists") {
		t.Errorf("CreatePlanningSprint() error = %v, want ErrSprintExists or 'already exists'", err)
	}
}

func TestCreatePlanningSprint_Success(t *testing.T) {
	ctx := context.Background()
	testRepo := t.TempDir()
	sprintsDir := filepath.Join(testRepo, "sprints")
	os.MkdirAll(sprintsDir, 0755)

	sprintRepo := filesystem.NewDefaultRepository()
	planService := services.NewSprintPlanService(sprintRepo, testRepo)

	req := services.CreatePlanningSprintRequest{
		ID:          "Sprint_27",
		Description: "Dashboard",
	}

	sprint, err := planService.CreatePlanningSprint(ctx, req)
	if err != nil {
		t.Fatalf("CreatePlanningSprint() error = %v", err)
	}

	// Verify sprint directory was created with Planning prefix
	expectedDir := filepath.Join(sprintsDir, "@Sprint_27_Dashboard")
	if sprint.DirectoryPath != expectedDir {
		t.Errorf("CreatePlanningSprint() directory = %q, want %q", sprint.DirectoryPath, expectedDir)
	}

	// Verify directory exists
	if _, err := os.Stat(expectedDir); err != nil {
		t.Errorf("Sprint directory does not exist: %v", err)
	}

	// Verify status file was created
	status, err := sprintRepo.ReadSprintStatus(ctx, expectedDir)
	if err != nil {
		t.Fatalf("ReadSprintStatus() error = %v", err)
	}
	if status != core.StatusPlanning {
		t.Errorf("ReadSprintStatus() = %v, want %v", status, core.StatusPlanning)
	}
}
