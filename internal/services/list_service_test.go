package services

import (
	"context"
	"errors"
	"testing"

	"github.com/gavin/gitta/internal/core"
)

type fakeStoryRepo struct {
	sprintPath     string
	findErr        error
	storyLists     map[string][]*core.Story
	listErr        error
	listErrPerPath map[string]error
}

func (f *fakeStoryRepo) ListStories(ctx context.Context, dirPath string) ([]*core.Story, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := f.listErrPerPath[dirPath]; err != nil {
		return nil, err
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.storyLists[dirPath], nil
}

func (f *fakeStoryRepo) FindCurrentSprint(ctx context.Context, sprintsDir string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if f.findErr != nil {
		return "", f.findErr
	}
	return f.sprintPath, nil
}

func (f *fakeStoryRepo) FindStoryByID(ctx context.Context, repoPath, storyID string) (*core.Story, string, error) {
	return nil, "", core.ErrStoryNotFound
}

func (f *fakeStoryRepo) FindStoryByPath(ctx context.Context, filePath string) (*core.Story, error) {
	return nil, core.ErrInvalidPath
}

type fakeGitRepo struct {
	branches []core.Branch
	err      error
}

func (f *fakeGitRepo) GetBranchList(ctx context.Context, repoPath string) ([]core.Branch, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.branches, nil
}

func (f *fakeGitRepo) CheckBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error) {
	return false, nil
}

func (f *fakeGitRepo) CreateBranch(ctx context.Context, repoPath, branchName string) error {
	return f.err
}

func (f *fakeGitRepo) CheckoutBranch(ctx context.Context, repoPath, branchName string, force bool) error {
	return f.err
}

func TestListSprintTasks_ReturnsStoriesWithStatus(t *testing.T) {
	repo := &fakeStoryRepo{
		sprintPath: "sprints/Sprint-01",
		storyLists: map[string][]*core.Story{
			"sprints/Sprint-01": {
				{ID: "US-002", Title: "Second"},
				{ID: "US-001", Title: "First"},
			},
		},
		listErrPerPath: map[string]error{},
	}
	gitRepo := &fakeGitRepo{}

	service := NewListService(repo, gitRepo)
	stories, err := service.ListSprintTasks(context.Background(), ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stories) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(stories))
	}
	if stories[0].Story.ID != "US-001" {
		t.Fatalf("expected stories sorted by ID, got %s first", stories[0].Story.ID)
	}
	if stories[0].Source != "Sprint" {
		t.Fatalf("expected source Sprint, got %s", stories[0].Source)
	}
	if stories[0].Status == "" {
		t.Fatalf("expected derived status to be set")
	}
}

func TestListSprintTasks_FailsWhenNoSprint(t *testing.T) {
	repo := &fakeStoryRepo{
		findErr:        errors.New("no sprint"),
		storyLists:     map[string][]*core.Story{},
		listErrPerPath: map[string]error{},
	}
	gitRepo := &fakeGitRepo{}

	service := NewListService(repo, gitRepo)
	_, err := service.ListSprintTasks(context.Background(), ".")
	if err == nil {
		t.Fatalf("expected error when sprint not found")
	}
}

func TestListAllTasks_BacklogOnly(t *testing.T) {
	repo := &fakeStoryRepo{
		findErr: errors.New("no sprint"),
		storyLists: map[string][]*core.Story{
			"backlog": {
				{ID: "BL-001", Title: "Backlog task"},
			},
		},
		listErrPerPath: map[string]error{},
	}
	gitRepo := &fakeGitRepo{}

	service := NewListService(repo, gitRepo)
	sprintStories, backlogStories, err := service.ListAllTasks(context.Background(), ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sprintStories) != 0 {
		t.Fatalf("expected no sprint stories, got %d", len(sprintStories))
	}
	if len(backlogStories) != 1 {
		t.Fatalf("expected backlog story, got %d", len(backlogStories))
	}
	if backlogStories[0].Source != "Backlog" {
		t.Fatalf("expected backlog source label")
	}
}

func TestListAllTasks_CombinesSprintAndBacklog(t *testing.T) {
	repo := &fakeStoryRepo{
		sprintPath: "sprints/Sprint-01",
		storyLists: map[string][]*core.Story{
			"sprints/Sprint-01": {
				{ID: "US-002", Title: "Second"},
			},
			"backlog": {
				{ID: "BL-001", Title: "Backlog"},
			},
		},
		listErrPerPath: map[string]error{},
	}
	gitRepo := &fakeGitRepo{}

	service := NewListService(repo, gitRepo)
	sprintStories, backlogStories, err := service.ListAllTasks(context.Background(), ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sprintStories) != 1 || len(backlogStories) != 1 {
		t.Fatalf("expected 1 sprint and 1 backlog story, got %d and %d", len(sprintStories), len(backlogStories))
	}
	if sprintStories[0].Story.ID != "US-002" || backlogStories[0].Story.ID != "BL-001" {
		t.Fatalf("unexpected IDs in result")
	}
}
