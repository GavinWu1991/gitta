package unit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestGitHistoryTraversal(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(string) ([]string, error) // Returns commit hashes
		sprintDir string
		wantCount int
		wantErr   bool
	}{
		{
			name: "single commit affecting sprint",
			setup: func(repoPath string) ([]string, error) {
				repo, err := git.PlainInit(repoPath, false)
				if err != nil {
					return nil, err
				}

				wt, _ := repo.Worktree()

				// Create initial commit (required for git)
				readme := filepath.Join(repoPath, "README.md")
				if err := os.WriteFile(readme, []byte("# Test\n"), 0644); err != nil {
					return nil, err
				}
				relReadme, _ := filepath.Rel(repoPath, readme)
				if _, err := wt.Add(relReadme); err != nil {
					return nil, err
				}
				if _, err := wt.Commit("Initial commit", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				}); err != nil {
					return nil, err
				}

				sprintDir := filepath.Join(repoPath, "sprints", "Sprint-01")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					return nil, err
				}

				// Create a task file
				taskFile := filepath.Join(sprintDir, "US-001.md")
				if err := os.WriteFile(taskFile, []byte("---\nid: US-001\ntitle: Task 1\n---\n"), 0644); err != nil {
					return nil, err
				}
				relTask, _ := filepath.Rel(repoPath, taskFile)
				if _, err := wt.Add(relTask); err != nil {
					return nil, err
				}
				commit, err := wt.Commit("Add task", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				})
				if err != nil {
					return nil, err
				}

				return []string{commit.String()}, nil
			},
			sprintDir: "sprints/Sprint-01",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple commits affecting sprint",
			setup: func(repoPath string) ([]string, error) {
				repo, err := git.PlainInit(repoPath, false)
				if err != nil {
					return nil, err
				}

				wt, _ := repo.Worktree()

				// Create initial commit (required for git)
				readme := filepath.Join(repoPath, "README.md")
				if err := os.WriteFile(readme, []byte("# Test\n"), 0644); err != nil {
					return nil, err
				}
				relReadme, _ := filepath.Rel(repoPath, readme)
				if _, err := wt.Add(relReadme); err != nil {
					return nil, err
				}
				if _, err := wt.Commit("Initial commit", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				}); err != nil {
					return nil, err
				}

				sprintDir := filepath.Join(repoPath, "sprints", "Sprint-01")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					return nil, err
				}

				var commits []string

				// First commit
				task1 := filepath.Join(sprintDir, "US-001.md")
				if err := os.WriteFile(task1, []byte("---\nid: US-001\ntitle: Task 1\n---\n"), 0644); err != nil {
					return nil, err
				}
				relTask1, _ := filepath.Rel(repoPath, task1)
				if _, err := wt.Add(relTask1); err != nil {
					return nil, err
				}
				commit1, err := wt.Commit("Add task 1", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				})
				if err != nil {
					return nil, err
				}
				commits = append(commits, commit1.String())

				// Second commit
				task2 := filepath.Join(sprintDir, "US-002.md")
				if err := os.WriteFile(task2, []byte("---\nid: US-002\ntitle: Task 2\n---\n"), 0644); err != nil {
					return nil, err
				}
				relTask2, _ := filepath.Rel(repoPath, task2)
				if _, err := wt.Add(relTask2); err != nil {
					return nil, err
				}
				commit2, err := wt.Commit("Add task 2", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				})
				if err != nil {
					return nil, err
				}
				commits = append(commits, commit2.String())

				return commits, nil
			},
			sprintDir: "sprints/Sprint-01",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "commits outside sprint directory ignored",
			setup: func(repoPath string) ([]string, error) {
				repo, err := git.PlainInit(repoPath, false)
				if err != nil {
					return nil, err
				}

				wt, _ := repo.Worktree()

				// Create initial commit (required for git)
				readme := filepath.Join(repoPath, "README.md")
				if err := os.WriteFile(readme, []byte("# Test\n"), 0644); err != nil {
					return nil, err
				}
				relReadme, _ := filepath.Rel(repoPath, readme)
				if _, err := wt.Add(relReadme); err != nil {
					return nil, err
				}
				if _, err := wt.Commit("Initial commit", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				}); err != nil {
					return nil, err
				}

				// Commit outside sprint
				otherFile := filepath.Join(repoPath, "other.md")
				if err := os.WriteFile(otherFile, []byte("other"), 0644); err != nil {
					return nil, err
				}
				relOther, _ := filepath.Rel(repoPath, otherFile)
				if _, err := wt.Add(relOther); err != nil {
					return nil, err
				}
				if _, err := wt.Commit("Add other file", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				}); err != nil {
					return nil, err
				}

				// Commit in sprint
				sprintDir := filepath.Join(repoPath, "sprints", "Sprint-01")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					return nil, err
				}
				taskFile := filepath.Join(sprintDir, "US-001.md")
				if err := os.WriteFile(taskFile, []byte("---\nid: US-001\ntitle: Task 1\n---\n"), 0644); err != nil {
					return nil, err
				}
				relTask, _ := filepath.Rel(repoPath, taskFile)
				if _, err := wt.Add(relTask); err != nil {
					return nil, err
				}
				commit, err := wt.Commit("Add task", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				})
				if err != nil {
					return nil, err
				}

				return []string{commit.String()}, nil
			},
			sprintDir: "sprints/Sprint-01",
			wantCount: 1, // Only the sprint commit
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup repo
			commits, err := tt.setup(tmpDir)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// This is a placeholder test - actual implementation will use GitHistoryAnalyzer
			// For now, we'll verify the setup worked
			if len(commits) != tt.wantCount {
				t.Errorf("expected %d commits, got %d", tt.wantCount, len(commits))
			}
		})
	}
}

func TestFileStateReconstruction(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(string) (string, error) // Returns commit hash
		sprintDir string
		wantFiles []string
		wantErr   bool
	}{
		{
			name: "reconstruct single file state",
			setup: func(repoPath string) (string, error) {
				repo, err := git.PlainInit(repoPath, false)
				if err != nil {
					return "", err
				}

				wt, _ := repo.Worktree()

				// Create initial commit (required for git)
				readme := filepath.Join(repoPath, "README.md")
				if err := os.WriteFile(readme, []byte("# Test\n"), 0644); err != nil {
					return "", err
				}
				relReadme, _ := filepath.Rel(repoPath, readme)
				if _, err := wt.Add(relReadme); err != nil {
					return "", err
				}
				if _, err := wt.Commit("Initial commit", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				}); err != nil {
					return "", err
				}

				sprintDir := filepath.Join(repoPath, "sprints", "Sprint-01")
				if err := os.MkdirAll(sprintDir, 0755); err != nil {
					return "", err
				}

				taskFile := filepath.Join(sprintDir, "US-001.md")
				if err := os.WriteFile(taskFile, []byte("---\nid: US-001\ntitle: Task 1\n---\n"), 0644); err != nil {
					return "", err
				}
				relTask, _ := filepath.Rel(repoPath, taskFile)
				if _, err := wt.Add(relTask); err != nil {
					return "", err
				}
				commit, err := wt.Commit("Add task", &git.CommitOptions{
					Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
				})
				if err != nil {
					return "", err
				}

				return commit.String(), nil
			},
			sprintDir: "sprints/Sprint-01",
			wantFiles: []string{"sprints/Sprint-01/US-001.md"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			commitHash, err := tt.setup(tmpDir)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// This is a placeholder test - actual implementation will use ReconstructFileState
			// For now, we'll verify the setup worked
			if commitHash == "" {
				t.Error("expected commit hash")
			}
		})
	}
}
