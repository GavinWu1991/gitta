package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"

	"github.com/gavin/gitta/internal/core"
)

// StartService orchestrates starting work on a story by creating/checking out a branch
// and optionally updating the assignee field in the story file.
type StartService interface {
	// Start begins work on a task by creating/checking out the feature branch and
	// optionally updating the assignee. Returns the story and branch name.
	Start(ctx context.Context, repoPath, taskIdentifier string, assignee *string) (*core.Story, string, error)
}

type startService struct {
	storyRepo core.StoryRepository
	gitRepo   core.GitRepository
	parser    core.StoryParser
	config    StatusEngineConfig
}

var assigneePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,50}$`)

// NewStartService constructs a StartService with the provided dependencies.
func NewStartService(storyRepo core.StoryRepository, gitRepo core.GitRepository, parser core.StoryParser) StartService {
	return &startService{
		storyRepo: storyRepo,
		gitRepo:   gitRepo,
		parser:    parser,
		config:    loadConfig(),
	}
}

func (s *startService) Start(ctx context.Context, repoPath, taskIdentifier string, assignee *string) (*core.Story, string, error) {
	// Context check
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrContextCancelled, err)
	}

	if repoPath == "" || taskIdentifier == "" {
		return nil, "", fmt.Errorf("%w: repository path and task identifier are required", ErrInvalidInput)
	}

	story, storyPath, err := s.resolveStory(ctx, repoPath, taskIdentifier)
	if err != nil {
		return nil, "", err
	}

	branchName := s.config.BranchPrefix + story.ID

	// Checkout (and create if needed) the branch.
	if err := s.gitRepo.CheckoutBranch(ctx, repoPath, branchName, false); err != nil {
		return nil, "", err
	}

	// Assignee update (optional).
	assigneeValue, err := s.resolveAssignee(ctx, repoPath, assignee)
	if err != nil {
		return story, branchName, err
	}

	if assigneeValue != "" {
		// Update story copy to avoid mutating caller data.
		assign := assigneeValue
		story.Assignee = &assign

		if validationErrors := s.parser.ValidateStory(story); len(validationErrors) > 0 {
			return story, branchName, fmt.Errorf("%w: %s", ErrInvalidInput, validationErrors[0].Message)
		}

		if err := s.parser.WriteStory(ctx, storyPath, story); err != nil {
			return story, branchName, &AssigneeUpdateError{
				FilePath: storyPath,
				Cause:    err,
			}
		}
	}

	return story, branchName, nil
}

func (s *startService) resolveStory(ctx context.Context, repoPath, identifier string) (*core.Story, string, error) {
	// Detect file path (absolute or relative to repoPath).
	candidatePath := identifier
	if !filepath.IsAbs(candidatePath) {
		candidatePath = filepath.Join(repoPath, candidatePath)
	}

	if info, err := os.Stat(candidatePath); err == nil && !info.IsDir() && strings.EqualFold(filepath.Ext(candidatePath), ".md") {
		story, err := s.storyRepo.FindStoryByPath(ctx, candidatePath)
		if err != nil {
			return nil, "", err
		}
		return story, candidatePath, nil
	}

	// Treat as ID otherwise.
	story, filePath, err := s.storyRepo.FindStoryByID(ctx, repoPath, identifier)
	if err != nil {
		return nil, "", err
	}
	return story, filePath, nil
}

func (s *startService) resolveAssignee(ctx context.Context, repoPath string, assignee *string) (string, error) {
	if assignee != nil && *assignee != "" {
		if !assigneePattern.MatchString(*assignee) {
			return "", fmt.Errorf("%w: assignee must contain only alphanumeric characters, hyphens, and underscores", ErrInvalidInput)
		}
		return *assignee, nil
	}

	// Try Git repository config first.
	userName, err := s.readGitUserName(repoPath)
	if err != nil {
		return "", err
	}

	if assigneePattern.MatchString(userName) {
		return userName, nil
	}

	// No valid assignee resolved.
	return "", nil
}

func (s *startService) readGitUserName(repoPath string) (string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", nil
	}

	if cfg, err := repo.Config(); err == nil && cfg != nil && cfg.User.Name != "" {
		return cfg.User.Name, nil
	}

	globalCfg, err := config.LoadConfig(config.GlobalScope)
	if err == nil && globalCfg.User.Name != "" {
		return globalCfg.User.Name, nil
	}

	return "", nil
}
