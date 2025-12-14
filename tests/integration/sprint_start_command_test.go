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

func TestSprintStartCommand(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		req       core.StartSprintRequest
		wantErr   bool
		setup     func(string) // Setup function that receives repoPath
		checkFunc func(*testing.T, *core.Sprint, string)
	}{
		{
			name: "create sprint with auto-generated name",
			req: core.StartSprintRequest{
				Duration: "2w",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sprint *core.Sprint, repoPath string) {
				if sprint.Name == "" {
					t.Error("sprint name should be auto-generated")
				}
				sprintDir := filepath.Join(repoPath, "sprints", sprint.Name)
				expectPathExists(t, sprintDir)
			},
		},
		{
			name: "create sprint with specific name",
			req: core.StartSprintRequest{
				Name:     "Sprint-01",
				Duration: "2w",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sprint *core.Sprint, repoPath string) {
				if sprint.Name != "Sprint-01" {
					t.Errorf("sprint name = %q, want %q", sprint.Name, "Sprint-01")
				}
				sprintDir := filepath.Join(repoPath, "sprints", "Sprint-01")
				expectPathExists(t, sprintDir)
			},
		},
		{
			name: "create sprint with custom duration",
			req: core.StartSprintRequest{
				Name:     "Sprint-02",
				Duration: "3w",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sprint *core.Sprint, repoPath string) {
				if sprint.Duration != "3w" {
					t.Errorf("sprint duration = %q, want %q", sprint.Duration, "3w")
				}
			},
		},
		{
			name: "sprint already exists",
			req: core.StartSprintRequest{
				Name:     "Sprint-01",
				Duration: "2w",
			},
			wantErr: true,
			setup: func(repoPath string) {
				// Create sprint directory first
				sprintDir := filepath.Join(repoPath, "sprints", "Sprint-01")
				os.MkdirAll(sprintDir, 0755)
			},
		},
		{
			name: "invalid sprint name",
			req: core.StartSprintRequest{
				Name:     "Sprint/Invalid",
				Duration: "2w",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := setupRepo(t)
			defer os.RemoveAll(testRepo)

			// Create sprints directory
			sprintsDir := filepath.Join(testRepo, "sprints")
			os.MkdirAll(sprintsDir, 0755)

			if tt.setup != nil {
				tt.setup(testRepo)
			}

			sprintRepo := filesystem.NewDefaultRepository()
			service := services.NewSprintStartService(sprintRepo, testRepo)
			sprint, err := service.StartSprint(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("StartSprint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && sprint != nil {
				if tt.checkFunc != nil {
					tt.checkFunc(t, sprint, testRepo)
				}

				// Verify current sprint link
				linkPath := filepath.Join(sprintsDir, ".current-sprint")
				currentSprint, _, err := filesystem.ReadCurrentSprintLink(linkPath)
				if err != nil {
					// Try text config fallback
					currentSprint, _, err = filesystem.ReadCurrentSprintLink(linkPath)
				}
				if err == nil {
					expectedPath := filepath.Join(sprintsDir, sprint.Name)
					absCurrent, _ := filepath.Abs(currentSprint)
					absExpected, _ := filepath.Abs(expectedPath)
					if absCurrent != absExpected {
						t.Errorf("current sprint link = %q, want %q", absCurrent, absExpected)
					}
				}
			}
		})
	}
}
