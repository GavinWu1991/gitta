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

func TestSprintCloseCommand(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setup     func(string) error
		req       core.RolloverRequest
		wantErr   bool
		checkFunc func(*testing.T, string, string, string)
	}{
		{
			name: "rollover unfinished tasks",
			setup: func(repoPath string) error {
				// Create source sprint with tasks
				sourceSprint := filepath.Join(repoPath, "sprints", "Sprint-01")
				os.MkdirAll(sourceSprint, 0755)

				// Create target sprint
				targetSprint := filepath.Join(repoPath, "sprints", "Sprint-02")
				os.MkdirAll(targetSprint, 0755)

				// Create unfinished task files
				task1 := filepath.Join(sourceSprint, "US-001.md")
				task2 := filepath.Join(sourceSprint, "US-002.md")
				task3 := filepath.Join(sourceSprint, "US-003.md")

				// Write task files with different statuses
				os.WriteFile(task1, []byte(`---
id: US-001
title: Task 1
status: todo
---
Task 1 description
`), 0644)

				os.WriteFile(task2, []byte(`---
id: US-002
title: Task 2
status: doing
---
Task 2 description
`), 0644)

				os.WriteFile(task3, []byte(`---
id: US-003
title: Task 3
status: done
---
Task 3 description
`), 0644)

				// Set current sprint link
				sprintsDir := filepath.Join(repoPath, "sprints")
				filesystem.CreateCurrentSprintLink(sourceSprint, filepath.Join(sprintsDir, ".current-sprint"))

				return nil
			},
			req: core.RolloverRequest{
				SourceSprintPath: filepath.Join("sprints", "Sprint-01"),
				TargetSprintPath: filepath.Join("sprints", "Sprint-02"),
				SelectedTaskIDs:  []string{"US-001", "US-002"},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, repoPath, sourceSprintPath, targetSprintPath string) {
				// Verify tasks moved
				sourceTask1 := filepath.Join(sourceSprintPath, "US-001.md")
				targetTask1 := filepath.Join(targetSprintPath, "US-001.md")

				if _, err := os.Stat(sourceTask1); err == nil {
					t.Errorf("source task US-001 should not exist")
				}
				if _, err := os.Stat(targetTask1); err != nil {
					t.Errorf("target task US-001 should exist: %v", err)
				}
			},
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
				if err := tt.setup(testRepo); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Resolve paths
			req := tt.req
			req.SourceSprintPath = filepath.Join(testRepo, req.SourceSprintPath)
			req.TargetSprintPath = filepath.Join(testRepo, req.TargetSprintPath)

			// Create service
			storyRepo := filesystem.NewDefaultRepository()
			sprintRepo := filesystem.NewDefaultRepository()
			parser := filesystem.NewMarkdownParser()
			closeService := services.NewSprintCloseService(storyRepo, sprintRepo, parser, testRepo)

			// Test CloseSprint first
			unfinished, err := closeService.CloseSprint(ctx, req.SourceSprintPath)
			if err != nil {
				t.Fatalf("CloseSprint failed: %v", err)
			}

			// Should find 2 unfinished tasks (US-001, US-002), not US-003 (done)
			if len(unfinished) != 2 {
				t.Errorf("expected 2 unfinished tasks, got %d", len(unfinished))
			}

			// Test RolloverTasks
			err = closeService.RolloverTasks(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("RolloverTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, testRepo, req.SourceSprintPath, req.TargetSprintPath)
			}
		})
	}
}
