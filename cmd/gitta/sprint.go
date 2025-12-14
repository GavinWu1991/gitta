package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/infra/git"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
	"github.com/gavin/gitta/pkg/ui"
	"github.com/gavin/gitta/ui/tui"
	"github.com/spf13/cobra"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Manage sprints (time-bounded work periods)",
	Long: `Manage sprints for organizing tasks by time periods.

Sprints are time-bounded work periods that help organize tasks. Use 'sprint start' to
create a new sprint, 'sprint close' to close a sprint and rollover unfinished tasks,
and 'sprint burndown' to view sprint progress over time.`,
}

var sprintStartCmd = &cobra.Command{
	Use:   "start [sprint-name]",
	Short: "Create and activate a new sprint",
	Long: `Create a new sprint directory, set up the current sprint link, and calculate the end date.

If sprint-name is not provided, the next sequential sprint name will be auto-generated
(e.g., Sprint-01, Sprint-02, etc.).

Examples:
  gitta sprint start                    # Auto-generate next sprint name
  gitta sprint start Sprint-02         # Create sprint with specific name
  gitta sprint start --duration 3w    # Create sprint with 3-week duration
  gitta sprint start --start-date 2025-02-01  # Create sprint starting on specific date`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		repoPath, err := findRepoRoot()
		if err != nil {
			return fmt.Errorf("not a git repository: %w", err)
		}

		sprintName := ""
		if len(args) > 0 {
			sprintName = args[0]
		}

		duration, _ := cmd.Flags().GetString("duration")
		startDateStr, _ := cmd.Flags().GetString("start-date")

		var startDate *time.Time
		if startDateStr != "" {
			parsed, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				return fmt.Errorf("invalid start-date format (use YYYY-MM-DD): %w", err)
			}
			startDate = &parsed
		}

		req := core.StartSprintRequest{
			Name:      sprintName,
			Duration:  duration,
			StartDate: startDate,
		}

		// Create service
		sprintRepo := filesystem.NewDefaultRepository()
		startService := services.NewSprintStartService(sprintRepo, repoPath)

		sprint, err := startService.StartSprint(ctx, req)
		if err != nil {
			return err
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(sprint); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
			return nil
		}

		fmt.Printf("✓ Created sprint: %s\n", sprint.Name)
		fmt.Printf("✓ Start date: %s\n", sprint.StartDate.Format("2006-01-02"))
		fmt.Printf("✓ End date: %s\n", sprint.EndDate.Format("2006-01-02"))
		fmt.Printf("✓ Duration: %s\n", sprint.Duration)
		fmt.Printf("✓ Current sprint link updated\n")
		return nil
	},
}

var sprintCloseCmd = &cobra.Command{
	Use:   "close [target-sprint]",
	Short: "Close current sprint and rollover unfinished tasks",
	Long: `Close the current sprint, identify unfinished tasks, and provide interactive
selection for rolling over tasks to the next sprint.

Examples:
  gitta sprint close                    # Interactive close with TUI selection
  gitta sprint close Sprint-02          # Close and rollover to Sprint-02
  gitta sprint close --target-sprint Sprint-02 --all  # Rollover all unfinished tasks
  gitta sprint close --skip             # Close without rollover`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		repoPath, err := findRepoRoot()
		if err != nil {
			return fmt.Errorf("not a git repository: %w", err)
		}

		// Get flags
		targetSprint, _ := cmd.Flags().GetString("target-sprint")
		rolloverAll, _ := cmd.Flags().GetBool("all")
		skip, _ := cmd.Flags().GetBool("skip")

		// If target sprint provided as argument, use it
		if len(args) > 0 && targetSprint == "" {
			targetSprint = args[0]
		}

		// Create services
		storyRepo := filesystem.NewDefaultRepository()
		sprintRepo := filesystem.NewDefaultRepository()
		parser := filesystem.NewMarkdownParser()
		closeService := services.NewSprintCloseService(storyRepo, sprintRepo, parser, repoPath)

		// Find current sprint
		sprintsDir := filepath.Join(repoPath, "sprints")
		currentSprintPath, err := storyRepo.FindCurrentSprint(ctx, sprintsDir)
		if err != nil {
			return fmt.Errorf("no current sprint found: %w", err)
		}

		// Identify unfinished tasks
		unfinished, err := closeService.CloseSprint(ctx, currentSprintPath)
		if err != nil {
			return fmt.Errorf("failed to close sprint: %w", err)
		}

		if len(unfinished) == 0 {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"closed_sprint":     filepath.Base(currentSprintPath),
					"unfinished_tasks":  0,
					"rolled_over_tasks": []string{},
				})
			}
			fmt.Printf("✓ Sprint %s closed (no unfinished tasks)\n", filepath.Base(currentSprintPath))
			return nil
		}

		// Handle skip flag
		if skip {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"closed_sprint":     filepath.Base(currentSprintPath),
					"unfinished_tasks":  len(unfinished),
					"rolled_over_tasks": []string{},
					"skipped":           true,
				})
			}
			fmt.Printf("✓ Sprint %s closed (skipped rollover)\n", filepath.Base(currentSprintPath))
			fmt.Printf("  Found %d unfinished tasks (not rolled over)\n", len(unfinished))
			return nil
		}

		// Determine target sprint
		if targetSprint == "" {
			return fmt.Errorf("target sprint required (use --target-sprint or provide as argument)")
		}

		targetSprintPath := filepath.Join(sprintsDir, targetSprint)
		targetExists, err := sprintRepo.SprintExists(ctx, targetSprintPath)
		if err != nil {
			return fmt.Errorf("failed to check target sprint: %w", err)
		}
		if !targetExists {
			return fmt.Errorf("target sprint %q does not exist", targetSprint)
		}

		// Select tasks to rollover
		var selectedIDs []string
		if rolloverAll {
			// Rollover all unfinished tasks
			for _, story := range unfinished {
				selectedIDs = append(selectedIDs, story.ID)
			}
		} else {
			// Interactive TUI selection
			selectedIDs, err = tui.SelectTasks(ctx, unfinished)
			if err != nil {
				return fmt.Errorf("task selection cancelled: %w", err)
			}
		}

		if len(selectedIDs) == 0 {
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"closed_sprint":     filepath.Base(currentSprintPath),
					"unfinished_tasks":  len(unfinished),
					"rolled_over_tasks": []string{},
				})
			}
			fmt.Printf("✓ Sprint %s closed (no tasks selected for rollover)\n", filepath.Base(currentSprintPath))
			return nil
		}

		// Perform rollover
		rolloverReq := core.RolloverRequest{
			SourceSprintPath: currentSprintPath,
			TargetSprintPath: targetSprintPath,
			SelectedTaskIDs:  selectedIDs,
		}

		if err := closeService.RolloverTasks(ctx, rolloverReq); err != nil {
			return fmt.Errorf("failed to rollover tasks: %w", err)
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]interface{}{
				"closed_sprint":     filepath.Base(currentSprintPath),
				"unfinished_tasks":  len(unfinished),
				"rolled_over_tasks": selectedIDs,
				"target_sprint":     targetSprint,
			})
		}

		fmt.Printf("✓ Sprint %s closed\n", filepath.Base(currentSprintPath))
		fmt.Printf("✓ Rolled over %d tasks to %s\n", len(selectedIDs), targetSprint)
		return nil
	},
}

var sprintBurndownCmd = &cobra.Command{
	Use:   "burndown [sprint-name]",
	Short: "Generate burndown chart from Git history",
	Long: `Generate a burndown chart showing sprint progress over time by analyzing
Git commit history and reconstructing daily remaining work.

Examples:
  gitta sprint burndown                  # Burndown for current sprint
  gitta sprint burndown Sprint-01        # Burndown for specific sprint
  gitta sprint burndown --format json    # Output as JSON
  gitta sprint burndown --format csv     # Output as CSV`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		repoPath, err := findRepoRoot()
		if err != nil {
			return fmt.Errorf("not a git repository: %w", err)
		}

		// Get flags
		sprintName, _ := cmd.Flags().GetString("sprint")
		format, _ := cmd.Flags().GetString("format")
		pointsOnly, _ := cmd.Flags().GetBool("points-only")
		tasksOnly, _ := cmd.Flags().GetBool("tasks-only")

		// If sprint name provided as argument, use it
		if len(args) > 0 && sprintName == "" {
			sprintName = args[0]
		}

		// Create services
		storyRepo := filesystem.NewDefaultRepository()
		sprintRepo := filesystem.NewDefaultRepository()
		parser := filesystem.NewMarkdownParser()
		gitAnalyzer := git.NewHistoryAnalyzer(parser)
		burndownService := services.NewSprintBurndownService(gitAnalyzer, sprintRepo, storyRepo, repoPath)

		// Find sprint path
		sprintsDir := filepath.Join(repoPath, "sprints")
		var sprintPath string

		if sprintName == "" {
			// Use current sprint
			currentSprintPath, err := storyRepo.FindCurrentSprint(ctx, sprintsDir)
			if err != nil {
				return fmt.Errorf("no current sprint found: %w", err)
			}
			sprintPath = currentSprintPath
		} else {
			sprintPath = filepath.Join(sprintsDir, sprintName)
			exists, err := sprintRepo.SprintExists(ctx, sprintPath)
			if err != nil {
				return fmt.Errorf("failed to check sprint: %w", err)
			}
			if !exists {
				return fmt.Errorf("sprint %q does not exist", sprintName)
			}
		}

		// Generate burndown data
		dataPoints, err := burndownService.GenerateBurndown(ctx, sprintPath)
		if err != nil {
			if err == core.ErrInsufficientHistory {
				return fmt.Errorf("insufficient Git history for burndown analysis: %w", err)
			}
			return fmt.Errorf("failed to generate burndown: %w", err)
		}

		// Output based on format
		switch format {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(dataPoints)

		case "csv":
			fmt.Println(ui.FormatBurndownCSV(dataPoints))
			return nil

		case "ascii", "":
			showPoints := pointsOnly || (!pointsOnly && !tasksOnly)
			showTasks := tasksOnly || (!pointsOnly && !tasksOnly)
			chart := ui.RenderBurndownChart(dataPoints, showPoints, showTasks)
			fmt.Println(chart)
			return nil

		default:
			return fmt.Errorf("invalid format: %s (supported: ascii, json, csv)", format)
		}
	},
}

func init() {
	// Sprint start flags
	sprintStartCmd.Flags().StringP("duration", "d", "2w", "Sprint duration (e.g., '2w', '14d')")
	sprintStartCmd.Flags().String("start-date", "", "Sprint start date (YYYY-MM-DD format, defaults to today)")

	// Sprint close flags
	sprintCloseCmd.Flags().StringP("target-sprint", "t", "", "Target sprint name for rollover")
	sprintCloseCmd.Flags().Bool("all", false, "Rollover all unfinished tasks without prompting")
	sprintCloseCmd.Flags().Bool("skip", false, "Skip rollover, just close the sprint")

	// Sprint burndown flags
	sprintBurndownCmd.Flags().StringP("sprint", "s", "", "Sprint name to analyze (alternative to positional argument)")
	sprintBurndownCmd.Flags().String("format", "ascii", "Output format (ascii, json, csv)")
	sprintBurndownCmd.Flags().Bool("points-only", false, "Show only story points (hide task count)")
	sprintBurndownCmd.Flags().Bool("tasks-only", false, "Show only task count (hide story points)")

	// Register subcommands
	sprintCmd.AddCommand(sprintStartCmd)
	sprintCmd.AddCommand(sprintCloseCmd)
	sprintCmd.AddCommand(sprintBurndownCmd)
}

// findRepoRoot finds the Git repository root by walking up from current directory.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository")
		}
		dir = parent
	}
}
