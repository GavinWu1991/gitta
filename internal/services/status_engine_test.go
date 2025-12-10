package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gavin/gitta/internal/core"
)

// mockGitRepository is a mock implementation of core.GitRepository for testing.
type mockGitRepository struct {
	branches       []core.Branch
	mergedBranches map[string]bool
	err            error
}

func (m *mockGitRepository) GetBranchList(ctx context.Context, repoPath string) ([]core.Branch, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.branches, nil
}

func (m *mockGitRepository) CheckBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	merged, ok := m.mergedBranches[branchName]
	if !ok {
		return false, nil
	}
	return merged, nil
}

func (m *mockGitRepository) CreateBranch(ctx context.Context, repoPath, branchName string) error {
	return nil
}

func (m *mockGitRepository) CheckoutBranch(ctx context.Context, repoPath, branchName string, force bool) error {
	return nil
}

func TestDeriveStatus_Todo(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		want      core.Status
		wantError bool
	}{
		{
			name: "no matching branch",
			story: &core.Story{
				ID:    "US-001",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-002", Type: core.BranchTypeLocal},
			},
			want:      core.StatusTodo,
			wantError: false,
		},
		{
			name: "empty branch list",
			story: &core.Story{
				ID:    "US-001",
				Title: "Test Story",
			},
			branches:  []core.Branch{},
			want:      core.StatusTodo,
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_Doing(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		want      core.Status
		wantError bool
	}{
		{
			name: "local branch exists, not pushed to remote",
			story: &core.Story{
				ID:    "US-002",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-002", Type: core.BranchTypeLocal},
			},
			want:      core.StatusDoing,
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_Review(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		want      core.Status
		wantError bool
	}{
		{
			name: "branch exists locally and on remote, not merged",
			story: &core.Story{
				ID:    "US-003",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-003", Type: core.BranchTypeLocal},
				{Name: "feat/US-003", Type: core.BranchTypeRemote, RemoteName: "origin"},
			},
			want:      core.StatusReview,
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_Done(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		merged    map[string]bool
		want      core.Status
		wantError bool
	}{
		{
			name: "branch merged into main",
			story: &core.Story{
				ID:    "US-004",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-004", Type: core.BranchTypeLocal},
			},
			merged:    map[string]bool{"feat/US-004": true},
			want:      core.StatusDone,
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: tt.merged,
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_ExplicitStatusOverride(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		want      core.Status
		wantError bool
	}{
		{
			name: "explicit status in Frontmatter overrides derivation",
			story: &core.Story{
				ID:     "US-005",
				Title:  "Test Story",
				Status: core.StatusDone, // Explicit status
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-005", Type: core.BranchTypeLocal}, // Would normally be Doing
			},
			want:      core.StatusDone, // Explicit status takes precedence
			wantError: false,
		},
		{
			name: "explicit status todo overrides branch existence",
			story: &core.Story{
				ID:     "US-006",
				Title:  "Test Story",
				Status: core.StatusTodo, // Explicit status
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-006", Type: core.BranchTypeLocal}, // Branch exists
			},
			want:      core.StatusTodo, // Explicit status takes precedence
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_CustomPrefixPatterns(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		prefix    string
		want      core.Status
		wantError bool
	}{
		{
			name: "custom prefix 'feature/'",
			story: &core.Story{
				ID:    "US-005",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feature/US-005", Type: core.BranchTypeLocal},
			},
			prefix:    "feature/",
			want:      core.StatusDoing,
			wantError: false,
		},
		{
			name: "custom prefix 'task/'",
			story: &core.Story{
				ID:    "TASK-123",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "task/TASK-123", Type: core.BranchTypeLocal},
			},
			prefix:    "task/",
			want:      core.StatusDoing,
			wantError: false,
		},
		{
			name: "no prefix (exact match)",
			story: &core.Story{
				ID:    "US-006",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "US-006", Type: core.BranchTypeLocal},
			},
			prefix:    "",
			want:      core.StatusDoing,
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			// Create engine with custom config
			engine := engineWithRepo.(*statusEngine)
			engine.config.BranchPrefix = tt.prefix

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name          string
		story         *core.Story
		branches      []core.Branch
		caseSensitive bool
		want          core.Status
		wantError     bool
	}{
		{
			name: "case-sensitive match (exact case)",
			story: &core.Story{
				ID:    "US-001",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-001", Type: core.BranchTypeLocal},
			},
			caseSensitive: true,
			want:          core.StatusDoing,
			wantError:     false,
		},
		{
			name: "case-sensitive no match (wrong case)",
			story: &core.Story{
				ID:    "US-001",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "FEAT/US-001", Type: core.BranchTypeLocal},
			},
			caseSensitive: true,
			want:          core.StatusTodo,
			wantError:     false,
		},
		{
			name: "case-insensitive match",
			story: &core.Story{
				ID:    "US-001",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "FEAT/US-001", Type: core.BranchTypeLocal},
			},
			caseSensitive: false,
			want:          core.StatusDoing,
			wantError:     false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			engine := engineWithRepo.(*statusEngine)
			engine.config.CaseSensitive = tt.caseSensitive

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_MultipleBranchMatches(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		want      core.Status
		wantError bool
	}{
		{
			name: "multiple branches match same pattern (return first)",
			story: &core.Story{
				ID:    "US-005",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-005", Type: core.BranchTypeLocal},
				{Name: "feature/US-005", Type: core.BranchTypeLocal},
				{Name: "task/US-005", Type: core.BranchTypeLocal},
			},
			want:      core.StatusDoing,
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_SimilarStoryIDs(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		want      core.Status
		wantError bool
	}{
		{
			name: "US-10 should not match US-100",
			story: &core.Story{
				ID:    "US-10",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-100", Type: core.BranchTypeLocal},
			},
			want:      core.StatusTodo,
			wantError: false,
		},
		{
			name: "US-100 should not match US-10",
			story: &core.Story{
				ID:    "US-100",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-10", Type: core.BranchTypeLocal},
			},
			want:      core.StatusTodo,
			wantError: false,
		},
		{
			name: "US-10 should match feat/US-10 exactly",
			story: &core.Story{
				ID:    "US-10",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-10", Type: core.BranchTypeLocal},
			},
			want:      core.StatusDoing,
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		story     *core.Story
		branches  []core.Branch
		want      core.Status
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil story",
			story:     nil,
			branches:  []core.Branch{},
			want:      "",
			wantError: true,
			errorMsg:  "story is nil or invalid",
		},
		{
			name: "story with missing ID",
			story: &core.Story{
				ID:    "",
				Title: "Test Story",
			},
			branches:  []core.Branch{},
			want:      "",
			wantError: true,
			errorMsg:  "story ID is required",
		},
		{
			name: "empty branch list",
			story: &core.Story{
				ID:    "US-001",
				Title: "Test Story",
			},
			branches:  []core.Branch{},
			want:      core.StatusTodo,
			wantError: false,
		},
		{
			name: "empty repository path",
			story: &core.Story{
				ID:    "US-001",
				Title: "Test Story",
			},
			branches:  []core.Branch{},
			want:      "",
			wantError: true,
			errorMsg:  "repository path cannot be empty",
		},
		{
			name: "story ID with special characters",
			story: &core.Story{
				ID:    "US-001@test",
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-001@test", Type: core.BranchTypeLocal},
			},
			want:      core.StatusDoing,
			wantError: false,
		},
		{
			name: "very long story ID",
			story: &core.Story{
				ID:    "US-" + strings.Repeat("X", 200),
				Title: "Test Story",
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
			},
			want:      core.StatusTodo,
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			testRepoPath := repoPath
			if tt.name == "empty repository path" {
				testRepoPath = ""
			}

			got, err := engineWithRepo.DeriveStatus(ctx, tt.story, tt.branches, testRepoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError {
				if err == nil {
					t.Errorf("DeriveStatus() expected error, got nil")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("DeriveStatus() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if got != tt.want {
					t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDeriveStatus_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	story := &core.Story{
		ID:    "US-001",
		Title: "Test Story",
	}
	branches := []core.Branch{
		{Name: "main", Type: core.BranchTypeLocal},
	}

	mockRepo := &mockGitRepository{
		branches:       branches,
		mergedBranches: make(map[string]bool),
	}
	engineWithRepo := NewStatusEngineWithRepository(mockRepo)

	_, err := engineWithRepo.DeriveStatus(ctx, story, branches, ".")
	if err == nil {
		t.Error("DeriveStatus() expected error on cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context was cancelled") {
		t.Errorf("DeriveStatus() error = %v, want error containing 'context was cancelled'", err)
	}
}

func TestDeriveStatusBatch(t *testing.T) {
	tests := []struct {
		name      string
		stories   []*core.Story
		branches  []core.Branch
		want      []core.Status
		wantError bool
	}{
		{
			name: "multiple stories with different statuses",
			stories: []*core.Story{
				{ID: "US-001", Title: "Story 1"},                          // No branch -> Todo
				{ID: "US-002", Title: "Story 2"},                          // Local branch -> Doing
				{ID: "US-003", Title: "Story 3", Status: core.StatusDone}, // Explicit status -> Done
			},
			branches: []core.Branch{
				{Name: "main", Type: core.BranchTypeLocal},
				{Name: "feat/US-002", Type: core.BranchTypeLocal},
			},
			want:      []core.Status{core.StatusTodo, core.StatusDoing, core.StatusDone},
			wantError: false,
		},
		{
			name:      "nil stories slice",
			stories:   nil,
			branches:  []core.Branch{},
			want:      nil,
			wantError: true,
		},
		{
			name:      "empty stories slice",
			stories:   []*core.Story{},
			branches:  []core.Branch{},
			want:      []core.Status{},
			wantError: false,
		},
	}

	ctx := context.Background()
	repoPath := "."

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockGitRepository{
				branches:       tt.branches,
				mergedBranches: make(map[string]bool),
			}
			engineWithRepo := NewStatusEngineWithRepository(mockRepo)

			got, err := engineWithRepo.DeriveStatusBatch(ctx, tt.stories, tt.branches, repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("DeriveStatusBatch() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if len(got) != len(tt.want) {
					t.Errorf("DeriveStatusBatch() returned %d statuses, want %d", len(got), len(tt.want))
					return
				}
				for i, wantStatus := range tt.want {
					if got[i] != wantStatus {
						t.Errorf("DeriveStatusBatch() status[%d] = %v, want %v", i, got[i], wantStatus)
					}
				}
			}
		})
	}
}

func TestDeriveStatusBatch_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	stories := []*core.Story{
		{ID: "US-001", Title: "Story 1"},
		{ID: "US-002", Title: "Story 2"},
	}
	branches := []core.Branch{
		{Name: "main", Type: core.BranchTypeLocal},
	}

	mockRepo := &mockGitRepository{
		branches:       branches,
		mergedBranches: make(map[string]bool),
	}
	engineWithRepo := NewStatusEngineWithRepository(mockRepo)

	_, err := engineWithRepo.DeriveStatusBatch(ctx, stories, branches, ".")
	if err == nil {
		t.Error("DeriveStatusBatch() expected error on cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context was cancelled") {
		t.Errorf("DeriveStatusBatch() error = %v, want error containing 'context was cancelled'", err)
	}
}

func TestDeriveStatusBatch_FailFast(t *testing.T) {
	ctx := context.Background()
	repoPath := "."

	stories := []*core.Story{
		{ID: "US-001", Title: "Story 1"},
		{ID: "", Title: "Story 2"}, // Invalid - missing ID
		{ID: "US-003", Title: "Story 3"},
	}
	branches := []core.Branch{
		{Name: "main", Type: core.BranchTypeLocal},
	}

	mockRepo := &mockGitRepository{
		branches:       branches,
		mergedBranches: make(map[string]bool),
	}
	engineWithRepo := NewStatusEngineWithRepository(mockRepo)

	_, err := engineWithRepo.DeriveStatusBatch(ctx, stories, branches, repoPath)
	if err == nil {
		t.Error("DeriveStatusBatch() expected error for invalid story, got nil")
	}
	if !strings.Contains(err.Error(), "US-002") && !strings.Contains(err.Error(), "story ID is required") {
		t.Errorf("DeriveStatusBatch() error = %v, want error mentioning story ID requirement", err)
	}
}

func TestDeriveStatus_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a large branch list (simulating 1000 branches)
	branches := make([]core.Branch, 1000)
	branches[0] = core.Branch{Name: "main", Type: core.BranchTypeLocal}
	for i := 1; i < 1000; i++ {
		branches[i] = core.Branch{
			Name: "feat/OTHER-" + string(rune('0'+i%10)),
			Type: core.BranchTypeLocal,
		}
	}
	// Add the target branch at the end
	branches[999] = core.Branch{Name: "feat/US-001", Type: core.BranchTypeLocal}

	story := &core.Story{
		ID:    "US-001",
		Title: "Test Story",
	}

	mockRepo := &mockGitRepository{
		branches:       branches,
		mergedBranches: make(map[string]bool),
	}
	engineWithRepo := NewStatusEngineWithRepository(mockRepo)

	ctx := context.Background()
	repoPath := "."

	// Benchmark single story derivation
	start := time.Now()
	_, err := engineWithRepo.DeriveStatus(ctx, story, branches, repoPath)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("DeriveStatus failed: %v", err)
	}

	// Target: <100ms per story (per SC-001)
	if duration > 100*time.Millisecond {
		t.Errorf("DeriveStatus took %v, expected <100ms", duration)
	}
}

func TestDeriveStatusBatch_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create 100 stories
	stories := make([]*core.Story, 100)
	for i := 0; i < 100; i++ {
		stories[i] = &core.Story{
			ID:    fmt.Sprintf("US-%03d", i+1),
			Title: fmt.Sprintf("Story %d", i+1),
		}
	}

	// Create branch list with some matching branches
	branches := []core.Branch{
		{Name: "main", Type: core.BranchTypeLocal},
	}
	for i := 0; i < 50; i++ {
		branches = append(branches, core.Branch{
			Name: fmt.Sprintf("feat/US-%03d", i+1),
			Type: core.BranchTypeLocal,
		})
	}

	mockRepo := &mockGitRepository{
		branches:       branches,
		mergedBranches: make(map[string]bool),
	}
	engineWithRepo := NewStatusEngineWithRepository(mockRepo)

	ctx := context.Background()
	repoPath := "."

	// Benchmark batch processing
	start := time.Now()
	_, err := engineWithRepo.DeriveStatusBatch(ctx, stories, branches, repoPath)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("DeriveStatusBatch failed: %v", err)
	}

	// Target: <5s for 100 stories (per SC-002)
	if duration > 5*time.Second {
		t.Errorf("DeriveStatusBatch took %v for 100 stories, expected <5s", duration)
	}
}
