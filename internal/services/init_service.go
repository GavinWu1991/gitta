package services

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed init_templates/*.md
var initTemplateFS embed.FS

// InitOptions configures init behavior.
type InitOptions struct {
	ExampleSprint string
	Force         bool
}

// InitResult captures the filesystem changes performed by init.
type InitResult struct {
	SprintDir   string
	BacklogDir  string
	Created     []string
	BackupPaths []string
}

// InitService initializes a repository with gitta workspace structure.
type InitService interface {
	Initialize(ctx context.Context, repoPath string, opts InitOptions) (*InitResult, error)
}

type initService struct {
	templates fs.FS
	now       func() time.Time
}

// NewInitService constructs an InitService with embedded templates.
func NewInitService() InitService {
	return &initService{
		templates: initTemplateFS,
		now:       time.Now,
	}
}

// Initialize creates the gitta workspace in repoPath using the provided options.
func (s *initService) Initialize(ctx context.Context, repoPath string, opts InitOptions) (*InitResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrContextCancelled, err)
	}
	if repoPath == "" {
		return nil, fmt.Errorf("%w: repository path is required", ErrInvalidInput)
	}

	if err := s.ensureGitRepo(repoPath); err != nil {
		return nil, err
	}

	sprintName := opts.ExampleSprint
	if sprintName == "" {
		sprintName = "Sprint-01"
	}
	if err := validateSprintName(sprintName); err != nil {
		return nil, err
	}

	sprintsRoot := filepath.Join(repoPath, "sprints")
	sprintDir := filepath.Join(sprintsRoot, sprintName)
	backlogDir := filepath.Join(repoPath, "backlog")

	existing := s.existingTargets([]string{sprintDir, backlogDir})
	if len(existing) > 0 && !opts.Force {
		return nil, fmt.Errorf("%w: found existing directories %v; rerun with --force to recreate", ErrWorkspaceExists, toBaseNames(existing))
	}

	var backups []string
	if opts.Force {
		for _, path := range existing {
			backup, err := s.backup(path)
			if err != nil {
				return nil, err
			}
			backups = append(backups, backup)
		}
	}

	if err := os.MkdirAll(sprintDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create sprint directory %s: %w", sprintDir, err)
	}
	if err := os.MkdirAll(backlogDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create backlog directory %s: %w", backlogDir, err)
	}

	created := []string{}
	if err := s.writeTemplate("US-001.md", sprintDir, &created); err != nil {
		return nil, err
	}
	if err := s.writeTemplate("US-002.md", backlogDir, &created); err != nil {
		return nil, err
	}

	return &InitResult{
		SprintDir:   sprintDir,
		BacklogDir:  backlogDir,
		Created:     created,
		BackupPaths: backups,
	}, nil
}

func (s *initService) ensureGitRepo(repoPath string) error {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: .git directory not found in %s", ErrNotGitRepository, repoPath)
		}
		return fmt.Errorf("failed to stat .git directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: .git exists but is not a directory at %s", ErrNotGitRepository, repoPath)
	}
	return nil
}

func validateSprintName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: example sprint name must be non-empty", ErrInvalidInput)
	}
	if strings.Contains(name, string(os.PathSeparator)) || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("%w: sprint name must not contain path separators", ErrInvalidInput)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("%w: sprint name must not contain parent path segments", ErrInvalidInput)
	}
	return nil
}

func (s *initService) existingTargets(paths []string) []string {
	var existing []string
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			existing = append(existing, p)
		}
	}
	return existing
}

func (s *initService) backup(path string) (string, error) {
	backup := fmt.Sprintf("%s.backup-%d", path, s.now().Unix())
	if err := os.Rename(path, backup); err != nil {
		return "", fmt.Errorf("failed to backup %s: %w", path, err)
	}
	return backup, nil
}

func (s *initService) writeTemplate(templateName, destDir string, created *[]string) error {
	sourcePath := filepath.Join("init_templates", templateName)
	data, err := fs.ReadFile(s.templates, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templateName, err)
	}
	destPath := filepath.Join(destDir, templateName)
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", destPath, err)
	}
	*created = append(*created, destPath)
	return nil
}

func toBaseNames(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, filepath.Base(p))
	}
	return out
}
