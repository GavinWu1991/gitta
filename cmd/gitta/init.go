package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gavin/gitta/internal/services"
)

var (
	initForce         bool
	initExampleSprint string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gitta workspace in the current Git repository",
	Long: `Set up the gitta workspace structure (sprints/backlog) with example tasks.

By default, creates tasks/sprints/Sprint-01/ and tasks/backlog/ with sample stories. Use --example-sprint
to customize the Sprint folder name and --force to back up and recreate existing gitta folders.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		repoPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine working directory: %w", err)
		}

		svc := services.NewInitService()
		result, err := svc.Initialize(ctx, repoPath, services.InitOptions{
			Force:         initForce,
			ExampleSprint: initExampleSprint,
		})
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNotGitRepository):
				return fmt.Errorf("init: %w", err)
			case errors.Is(err, services.ErrWorkspaceExists):
				return err
			case errors.Is(err, services.ErrInvalidInput):
				return err
			default:
				return fmt.Errorf("init failed: %w", err)
			}
		}

		fmt.Println("✅ Gitta initialized successfully!")
		fmt.Println()
		fmt.Println("Created directories:")
		fmt.Printf("  - %s\n", relPath(repoPath, result.SprintDir))
		fmt.Printf("  - %s\n", relPath(repoPath, result.BacklogDir))
		fmt.Println()
		fmt.Println("Created example tasks:")
		for _, f := range result.Created {
			fmt.Printf("  - %s\n", relPath(repoPath, f))
		}
		if len(result.BackupPaths) > 0 {
			fmt.Println()
			fmt.Println("Backups:")
			for _, b := range result.BackupPaths {
				fmt.Printf("  - %s\n", relPath(repoPath, b))
			}
		}
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1) gitta list")
		fmt.Println("  2) gitta list --all")
		fmt.Println("  3) Edit example tasks or create new ones")

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Backup existing gitta directories and recreate")
	initCmd.Flags().StringVar(&initExampleSprint, "example-sprint", "Sprint-01", "Sprint name for example tasks")
}

func relPath(repoRoot, target string) string {
	if rel, err := filepath.Rel(repoRoot, target); err == nil {
		return rel
	}
	return target
}
