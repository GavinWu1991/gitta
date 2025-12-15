package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
)

func TestParseFolderName(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		wantStatus core.SprintStatus
		wantID     string
		wantDesc   string
		wantErr    bool
	}{
		{"Active sprint", "!Sprint_24_Login", core.StatusActive, "Sprint_24", "Login", false},
		{"Ready sprint", "+Sprint_25_Payment", core.StatusReady, "Sprint_25", "Payment", false},
		{"Planning sprint", "@Sprint_26_Dashboard", core.StatusPlanning, "Sprint_26", "Dashboard", false},
		{"Archived sprint", "~Sprint_23_Onboarding", core.StatusArchived, "Sprint_23", "Onboarding", false},
		{"Active no description", "!Sprint_24", core.StatusActive, "Sprint_24", "", false},
		{"Ready no description", "+Sprint_25", core.StatusReady, "Sprint_25", "", false},
		{"Planning no description", "@Sprint_26", core.StatusPlanning, "Sprint_26", "", false},
		{"Archived no description", "~Sprint_23", core.StatusArchived, "Sprint_23", "", false},
		{"No prefix", "Sprint_24", core.StatusActive, "", "", true},
		{"Empty name", "", core.StatusActive, "", "", true},
		{"Invalid prefix", "#Sprint_24", core.StatusActive, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, id, desc, err := filesystem.ParseFolderName(tt.folderName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFolderName(%q) error = %v, wantErr %v", tt.folderName, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if status != tt.wantStatus {
					t.Errorf("ParseFolderName(%q) status = %v, want %v", tt.folderName, status, tt.wantStatus)
				}
				if id != tt.wantID {
					t.Errorf("ParseFolderName(%q) id = %q, want %q", tt.folderName, id, tt.wantID)
				}
				if desc != tt.wantDesc {
					t.Errorf("ParseFolderName(%q) desc = %q, want %q", tt.folderName, desc, tt.wantDesc)
				}
			}
		})
	}
}

func TestBuildFolderName(t *testing.T) {
	tests := []struct {
		name       string
		status     core.SprintStatus
		id         string
		desc       string
		wantFolder string
	}{
		{"Active with desc", core.StatusActive, "Sprint_24", "Login", "!Sprint_24_Login"},
		{"Ready with desc", core.StatusReady, "Sprint_25", "Payment", "+Sprint_25_Payment"},
		{"Planning with desc", core.StatusPlanning, "Sprint_26", "Dashboard", "@Sprint_26_Dashboard"},
		{"Archived with desc", core.StatusArchived, "Sprint_23", "Onboarding", "~Sprint_23_Onboarding"},
		{"Active no desc", core.StatusActive, "Sprint_24", "", "!Sprint_24"},
		{"Ready no desc", core.StatusReady, "Sprint_25", "", "+Sprint_25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filesystem.BuildFolderName(tt.status, tt.id, tt.desc)
			if got != tt.wantFolder {
				t.Errorf("BuildFolderName(%v, %q, %q) = %q, want %q", tt.status, tt.id, tt.desc, got, tt.wantFolder)
			}
		})
	}
}

func TestExtractStatus(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		wantStatus core.SprintStatus
	}{
		{"Active prefix", "!Sprint_24", core.StatusActive},
		{"Ready prefix", "+Sprint_25", core.StatusReady},
		{"Planning prefix", "@Sprint_26", core.StatusPlanning},
		{"Archived prefix", "~Sprint_23", core.StatusArchived},
		{"No prefix", "Sprint_24", core.StatusActive}, // Defaults to Active
		{"Empty", "", core.StatusActive},              // Defaults to Active
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filesystem.ExtractStatus(tt.folderName)
			if got != tt.wantStatus {
				t.Errorf("ExtractStatus(%q) = %v, want %v", tt.folderName, got, tt.wantStatus)
			}
		})
	}
}

func TestReadWriteSprintStatus(t *testing.T) {
	tmpDir := t.TempDir()
	sprintDir := filepath.Join(tmpDir, "test-sprint")
	if err := os.MkdirAll(sprintDir, 0755); err != nil {
		t.Fatalf("Failed to create sprint directory: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name   string
		status core.SprintStatus
	}{
		{"Write Active", core.StatusActive},
		{"Write Ready", core.StatusReady},
		{"Write Planning", core.StatusPlanning},
		{"Write Archived", core.StatusArchived},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write status
			if err := filesystem.WriteSprintStatus(ctx, sprintDir, tt.status); err != nil {
				t.Fatalf("WriteSprintStatus() error = %v", err)
			}

			// Verify file exists
			if !filesystem.StatusFileExists(ctx, sprintDir) {
				t.Error("StatusFileExists() = false, want true")
			}

			// Read status
			got, err := filesystem.ReadSprintStatus(ctx, sprintDir)
			if err != nil {
				t.Fatalf("ReadSprintStatus() error = %v", err)
			}
			if got != tt.status {
				t.Errorf("ReadSprintStatus() = %v, want %v", got, tt.status)
			}
		})
	}
}

func TestReadSprintStatusFromFolderName(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	tests := []struct {
		name       string
		folderName string
		wantStatus core.SprintStatus
	}{
		{"Active from folder", "!Sprint_24", core.StatusActive},
		{"Ready from folder", "+Sprint_25", core.StatusReady},
		{"Planning from folder", "@Sprint_26", core.StatusPlanning},
		{"Archived from folder", "~Sprint_23", core.StatusArchived},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprintDir := filepath.Join(tmpDir, tt.folderName)
			if err := os.MkdirAll(sprintDir, 0755); err != nil {
				t.Fatalf("Failed to create sprint directory: %v", err)
			}

			// Read status (should infer from folder name since no .gitta/status file)
			got, err := filesystem.ReadSprintStatus(ctx, sprintDir)
			if err != nil {
				t.Fatalf("ReadSprintStatus() error = %v", err)
			}
			if got != tt.wantStatus {
				t.Errorf("ReadSprintStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestValidateSprintNameForStatusPrefix(t *testing.T) {
	tests := []struct {
		name       string
		sprintName string
		wantErr    bool
	}{
		{"Valid name", "Sprint_24_Login", false},
		{"Valid name with spaces", "Sprint 24 Login", false},
		{"Contains !", "Sprint_24_!Important", true},
		{"Contains +", "Sprint_24_+Bonus", true},
		{"Contains @", "Sprint_24_@Mention", true},
		{"Contains ~", "Sprint_24_~Home", true},
		{"Multiple invalid", "Sprint_24_!+@~", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filesystem.ValidateSprintName(tt.sprintName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSprintName(%q) error = %v, wantErr %v", tt.sprintName, err, tt.wantErr)
			}
		})
	}
}
