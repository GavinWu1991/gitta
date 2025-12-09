package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/gavin/gitta/internal/core"
)

// StatusEngine derives story status from Git branch state.
//
// StatusEngine automatically determines task status (Todo, Doing, Review, Done)
// based on Git branch state. It uses a configurable branch naming pattern to
// match story IDs to branch names, then checks branch existence, merge status,
// and remote tracking to derive the appropriate status.
//
// Example usage:
//
//	engine := services.NewStatusEngine()
//	gitRepo := git.NewRepository()
//	branchList, err := gitRepo.GetBranchList(ctx, repoPath)
//	if err != nil {
//		return err
//	}
//
//	status, err := engine.DeriveStatus(ctx, story, branchList, repoPath)
//	if err != nil {
//		return err
//	}
//	story.Status = status
type StatusEngine interface {
	// DeriveStatus derives the status for a single story based on Git branch state.
	// Returns the derived Status enum value, or an error if derivation fails.
	//
	// The derivation follows this priority order:
	//  1. Explicit Frontmatter status (if set, takes precedence)
	//  2. Branch existence (no branch → Todo)
	//  3. Merge status (merged → Done)
	//  4. Remote branch existence (on remote → Review, local only → Doing)
	DeriveStatus(
		ctx context.Context,
		story *core.Story,
		branchList []core.Branch,
		repoPath string,
	) (core.Status, error)

	// DeriveStatusBatch derives status for multiple stories in a single operation.
	// Uses the same branch list for all stories (more efficient than multiple calls).
	//
	// Returns an error on first failure (fail-fast behavior).
	// Respects context cancellation between story processing.
	DeriveStatusBatch(
		ctx context.Context,
		stories []*core.Story,
		branchList []core.Branch,
		repoPath string,
	) ([]core.Status, error)
}

// statusEngine is the implementation of StatusEngine.
type statusEngine struct {
	gitRepo core.GitRepository
	config  StatusEngineConfig
}

// NewStatusEngine creates a new StatusEngine instance.
// Uses default configuration (can be overridden via Viper).
func NewStatusEngine() StatusEngine {
	return &statusEngine{
		gitRepo: nil, // Will be set via SetGitRepository or constructor parameter
		config:  loadConfig(),
	}
}

// NewStatusEngineWithRepository creates a StatusEngine with a GitRepository.
// Useful for dependency injection in tests.
func NewStatusEngineWithRepository(repo core.GitRepository) StatusEngine {
	return &statusEngine{
		gitRepo: repo,
		config:  loadConfig(),
	}
}

// DeriveStatus derives the status for a single story based on Git branch state.
func (e *statusEngine) DeriveStatus(
	ctx context.Context,
	story *core.Story,
	branchList []core.Branch,
	repoPath string,
) (core.Status, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("%w: %w", ErrContextCancelled, err)
	}

	// Input validation
	if story == nil {
		return "", fmt.Errorf("%w: story cannot be nil", ErrInvalidStory)
	}

	if story.ID == "" {
		return "", fmt.Errorf("%w: story ID is required", ErrInvalidInput)
	}

	if repoPath == "" {
		return "", fmt.Errorf("%w: repository path cannot be empty", ErrInvalidInput)
	}

	// Priority 1: Check explicit Frontmatter status
	if story.Status != "" {
		// Validate status is a valid enum value
		validStatuses := map[core.Status]bool{
			core.StatusTodo:   true,
			core.StatusDoing:  true,
			core.StatusReview: true,
			core.StatusDone:   true,
		}
		if validStatuses[story.Status] {
			return story.Status, nil
		}
		// Invalid status in Frontmatter - continue with derivation
	}

	// Priority 2: Check branch existence
	matchingBranch := branchMatcher(story.ID, branchList, e.config)
	if matchingBranch == nil {
		return core.StatusTodo, nil
	}

	// Priority 3: Check merge status (merged branches are always Done)
	if e.gitRepo != nil {
		merged, err := e.gitRepo.CheckBranchMerged(ctx, repoPath, matchingBranch.Name)
		if err != nil {
			// If merge check fails, continue with other checks
			// (branch might not exist in remote, or repo might be in unusual state)
			// Error is logged but doesn't block status derivation
		} else if merged {
			return core.StatusDone, nil
		}
	}

	// Priority 4: Check remote branch existence
	if checkRemoteBranchExists(matchingBranch.Name, branchList) {
		return core.StatusReview, nil
	}

	// Branch exists locally only
	return core.StatusDoing, nil
}

// DeriveStatusBatch derives status for multiple stories in a single operation.
func (e *statusEngine) DeriveStatusBatch(
	ctx context.Context,
	stories []*core.Story,
	branchList []core.Branch,
	repoPath string,
) ([]core.Status, error) {
	if stories == nil {
		return nil, fmt.Errorf("%w: stories slice cannot be nil", ErrInvalidInput)
	}

	statuses := make([]core.Status, len(stories))
	for i, story := range stories {
		// Check context cancellation before each story
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrContextCancelled, err)
		}

		status, err := e.DeriveStatus(ctx, story, branchList, repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to derive status for story %q: %w", story.ID, err)
		}
		statuses[i] = status
	}

	return statuses, nil
}

// branchMatcher matches a story ID to a branch name using configurable pattern.
// Returns the matching branch or nil if no match found.
func branchMatcher(storyID string, branchList []core.Branch, config StatusEngineConfig) *core.Branch {
	if storyID == "" {
		return nil
	}

	// Build expected branch name from pattern
	expectedBranchName := config.BranchPrefix + storyID

	// Handle no-prefix pattern (exact match)
	if config.BranchPrefix == "" {
		expectedBranchName = storyID
	}

	for i := range branchList {
		branch := &branchList[i]
		// Only match local branches for the initial match
		if branch.Type != core.BranchTypeLocal {
			continue
		}

		var matches bool
		if config.CaseSensitive {
			matches = branch.Name == expectedBranchName
		} else {
			matches = strings.EqualFold(branch.Name, expectedBranchName)
		}

		if matches {
			return branch
		}
	}

	return nil
}

// checkRemoteBranchExists checks if a branch exists on remote by looking for
// a corresponding remote branch in the branch list.
func checkRemoteBranchExists(branchName string, branchList []core.Branch) bool {
	for i := range branchList {
		branch := &branchList[i]
		if branch.Type == core.BranchTypeRemote && branch.Name == branchName {
			return true
		}
	}
	return false
}
