package git

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gavin/gitta/internal/core"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
)

// HistoryAnalyzer implements core.GitHistoryAnalyzer using go-git.
type HistoryAnalyzer struct {
	parser core.StoryParser
}

// NewHistoryAnalyzer creates a new GitHistoryAnalyzer instance.
func NewHistoryAnalyzer(parser core.StoryParser) *HistoryAnalyzer {
	return &HistoryAnalyzer{parser: parser}
}

// AnalyzeSprintHistory implements core.GitHistoryAnalyzer.AnalyzeSprintHistory.
func (h *HistoryAnalyzer) AnalyzeSprintHistory(ctx context.Context, req core.AnalyzeHistoryRequest) ([]core.CommitSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(req.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrInvalidCommit, err)
	}

	// Get HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("%w: no HEAD commit", core.ErrInsufficientHistory)
	}

	// Iterate through commits
	cIter, err := repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer cIter.Close()

	var snapshots []core.CommitSnapshot
	visitedDates := make(map[string]bool) // Track dates to avoid duplicates

	err = cIter.ForEach(func(c *object.Commit) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		commitDate := c.Author.When
		dateKey := commitDate.Format("2006-01-02")

		// Filter by date range
		if commitDate.Before(req.StartDate) || commitDate.After(req.EndDate.Add(24*time.Hour)) {
			return nil // Skip commits outside date range
		}

		// Filter merge commits if requested
		if !req.IncludeMergeCommits && c.NumParents() > 1 {
			return nil
		}

		// Check if commit affects sprint directory
		affected, err := h.commitAffectsSprintDir(ctx, c, req.SprintDir)
		if err != nil {
			return err
		}
		if !affected {
			return nil
		}

		// Reconstruct file state at this commit
		files, err := h.ReconstructFileState(ctx, req.RepoPath, c.Hash.String(), req.SprintDir)
		if err != nil {
			return fmt.Errorf("failed to reconstruct file state at commit %s: %w", c.Hash.String()[:7], err)
		}

		// Group by date - only keep one snapshot per day (latest commit of the day)
		if !visitedDates[dateKey] || len(snapshots) == 0 || snapshots[len(snapshots)-1].CommitDate.Before(commitDate) {
			snapshot := core.CommitSnapshot{
				CommitHash: c.Hash.String(),
				CommitDate: commitDate,
				Author:     c.Author.Name,
				Message:    strings.Split(c.Message, "\n")[0], // First line only
				Files:      files,
			}

			if visitedDates[dateKey] && len(snapshots) > 0 {
				// Replace the last snapshot for this date
				snapshots[len(snapshots)-1] = snapshot
			} else {
				snapshots = append(snapshots, snapshot)
				visitedDates[dateKey] = true
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(snapshots)-1; i < j; i, j = i+1, j-1 {
		snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
	}

	return snapshots, nil
}

// ReconstructFileState implements core.GitHistoryAnalyzer.ReconstructFileState.
func (h *HistoryAnalyzer) ReconstructFileState(ctx context.Context, repoPath string, commitHash string, dirPath string) (map[string]*core.Story, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrInvalidCommit, err)
	}

	hash := plumbing.NewHash(commitHash)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrInvalidCommit, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	files := make(map[string]*core.Story)

	// Walk the tree to find files in the sprint directory
	err = tree.Files().ForEach(func(f *object.File) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if file is in the sprint directory
		if !strings.HasPrefix(f.Name, dirPath+"/") {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(f.Name, ".md") {
			return nil
		}

		// Read file content from Git
		reader, err := f.Reader()
		if err != nil {
			return nil // Skip files that can't be read
		}
		defer reader.Close()

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(reader); err != nil {
			return nil // Skip files that can't be read
		}

		// Parse story from bytes
		story, err := h.parseStoryFromBytes(buf.Bytes(), f.Name)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		// Use relative path from repo root
		files[f.Name] = story
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// commitAffectsSprintDir checks if a commit affects files in the sprint directory.
func (h *HistoryAnalyzer) commitAffectsSprintDir(ctx context.Context, commit *object.Commit, sprintDir string) (bool, error) {
	// Get parent commit for comparison
	if commit.NumParents() == 0 {
		// Initial commit - check if it has files in sprint dir
		tree, err := commit.Tree()
		if err != nil {
			return false, err
		}

		affected := false
		err = tree.Files().ForEach(func(f *object.File) error {
			if strings.HasPrefix(f.Name, sprintDir+"/") && strings.HasSuffix(f.Name, ".md") {
				affected = true
				return fmt.Errorf("found") // Break loop
			}
			return nil
		})
		return affected, nil
	}

	// Compare with parent commit
	parent, err := commit.Parent(0)
	if err != nil {
		return false, err
	}

	parentTree, err := parent.Tree()
	if err != nil {
		return false, err
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return false, err
	}

	changes, err := object.DiffTree(parentTree, commitTree)
	if err != nil {
		return false, err
	}

	for _, change := range changes {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		// Check if change affects sprint directory
		from := change.From.Name
		to := change.To.Name

		if (from != "" && strings.HasPrefix(from, sprintDir+"/")) ||
			(to != "" && strings.HasPrefix(to, sprintDir+"/")) {
			return true, nil
		}
	}

	return false, nil
}

// parseStoryFromBytes parses a Story from byte content (used for Git tree objects).
func (h *HistoryAnalyzer) parseStoryFromBytes(data []byte, filePath string) (*core.Story, error) {
	// Use Goldmark to parse frontmatter
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)

	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := md.Convert(data, &buf, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("failed to parse Markdown: %w", err)
	}

	// Extract metadata
	metaData := meta.Get(ctx)

	// Extract body content
	body := extractBodyFromBytes(data)

	// Unmarshal frontmatter into Story struct
	var story core.Story
	if len(metaData) > 0 {
		yamlData, err := yaml.Marshal(metaData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
		}

		if err := yaml.Unmarshal(yamlData, &story); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML frontmatter: %w", err)
		}
	}

	// Set body content
	story.Body = body

	// Apply defaults
	applyDefaultsToStory(&story)

	return &story, nil
}

// extractBodyFromBytes extracts the Markdown body content after the frontmatter delimiter.
func extractBodyFromBytes(data []byte) string {
	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return ""
	}

	if strings.TrimSpace(lines[0]) != "---" {
		return content
	}

	bodyStart := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			bodyStart = i + 1
			break
		}
	}

	if bodyStart == -1 || bodyStart >= len(lines) {
		return ""
	}

	bodyLines := lines[bodyStart:]
	body := strings.Join(bodyLines, "\n")
	return strings.TrimPrefix(body, "\n")
}

// applyDefaultsToStory sets default values for missing optional fields.
func applyDefaultsToStory(story *core.Story) {
	if story.Priority == "" {
		story.Priority = core.PriorityMedium
	}
	if story.Status == "" {
		story.Status = core.StatusTodo
	}
	if story.Tags == nil {
		story.Tags = []string{}
	}
}
