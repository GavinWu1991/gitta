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
	Use:   "start [sprint-id]",
	Short: "Create and activate a new sprint, or activate an existing sprint",
	Long: `Create a new sprint directory, set up the current sprint link, and calculate the end date.
Alternatively, activate an existing sprint in Ready or Planning status.

If sprint-id is not provided, the next sequential sprint name will be auto-generated
(e.g., Sprint-01, Sprint-02, etc.) and a new sprint will be created.

If sprint-id is provided and matches an existing sprint in Ready or Planning status,
that sprint will be activated (transitioned to Active status).

Examples:
  gitta sprint start                    # Auto-generate next sprint name and create
  gitta sprint start Sprint-02         # Create sprint with specific name
  gitta sprint start 24                # Activate existing sprint (partial match)
  gitta sprint start Sprint_24         # Activate existing sprint (full match)
  gitta sprint start --duration 3w    # Create sprint with 3-week duration
  gitta sprint start --start-date 2025-02-01  # Create sprint starting on specific date
  gitta sprint start --dry-run         # Show what would be done without making changes`,
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

		sprintID := ""
		if len(args) > 0 {
			sprintID = args[0]
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		sprintRepo := filesystem.NewDefaultRepository()
		sprintsDir := filepath.Join(repoPath, "sprints")

		// If sprint ID provided, try to activate existing sprint
		if sprintID != "" {
			statusService := services.NewSprintStatusService(sprintRepo, repoPath)

			if dryRun {
				// Check if sprint exists and can be activated
				targetSprint, err := sprintRepo.ResolveSprintByID(ctx, sprintsDir, sprintID)
				if err != nil {
					return fmt.Errorf("[DRY RUN] sprint %q not found in Ready or Planning status: %w", sprintID, err)
				}

				currentStatus, err := sprintRepo.ReadSprintStatus(ctx, targetSprint.DirectoryPath)
				if err != nil {
					return fmt.Errorf("[DRY RUN] failed to read sprint status: %w", err)
				}

				activeSprint, _ := sprintRepo.FindActiveSprint(ctx, sprintsDir)

				if jsonOutput {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					return enc.Encode(map[string]interface{}{
						"dry_run": true,
						"action":  "activate",
						"target": map[string]interface{}{
							"sprint_id": filepath.Base(targetSprint.DirectoryPath),
							"status":    currentStatus.String(),
						},
						"would_archive": activeSprint != nil,
					})
				}

				fmt.Printf("[DRY RUN] Would activate sprint: %s\n", filepath.Base(targetSprint.DirectoryPath))
				if activeSprint != nil {
					fmt.Printf("[DRY RUN] Would archive current active sprint: %s\n", filepath.Base(activeSprint.DirectoryPath))
				}
				return nil
			}

			// Activate existing sprint
			result, err := statusService.ActivateSprint(ctx, sprintID)
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				output := map[string]interface{}{
					"activated": map[string]interface{}{
						"name":   result.Activated.Name,
						"path":   result.Activated.DirectoryPath,
						"status": "active",
					},
				}
				if result.Archived != nil {
					output["archived"] = map[string]interface{}{
						"name":   result.Archived.Name,
						"path":   result.Archived.DirectoryPath,
						"status": "archived",
					}
				}
				output["current_link"] = filepath.Join(sprintsDir, "Current")
				return enc.Encode(output)
			}

			fmt.Printf("Activated sprint: %s\n", result.Activated.Name)
			if result.Archived != nil {
				fmt.Printf("Archived previous active sprint: %s\n", result.Archived.Name)
			}
			fmt.Printf("Current sprint link updated.\n")
			return nil
		}

		// No sprint ID provided, create new sprint
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
			Name:      "",
			Duration:  duration,
			StartDate: startDate,
		}

		// Create service
		startService := services.NewSprintStartService(sprintRepo, repoPath)

		if dryRun {
			// For dry run, just show what would be created
			existing, err := sprintRepo.ListSprints(ctx, sprintsDir)
			if err != nil {
				return fmt.Errorf("[DRY RUN] failed to list sprints: %w", err)
			}

			nextName := generateNextSprintName(existing)

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"dry_run": true,
					"action":  "create",
					"would_create": map[string]interface{}{
						"name": nextName,
					},
				})
			}

			fmt.Printf("[DRY RUN] Would create sprint: %s\n", nextName)
			return nil
		}

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

// generateNextSprintName generates the next sequential sprint name (helper for dry-run).
func generateNextSprintName(existing []string) string {
	if len(existing) == 0 {
		return "Sprint-01"
	}

	maxNum := 0
	for _, name := range existing {
		var num int
		if _, err := fmt.Sscanf(name, "Sprint-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
		// Also try pattern with status prefix
		if _, err := fmt.Sscanf(name, "!Sprint-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
		if _, err := fmt.Sscanf(name, "+Sprint-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
		if _, err := fmt.Sscanf(name, "@Sprint-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
		if _, err := fmt.Sscanf(name, "~Sprint-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	nextNum := maxNum + 1
	return fmt.Sprintf("Sprint-%02d", nextNum)
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

var sprintPlanCmd = &cobra.Command{
	Use:   "plan <name>",
	Short: "Create a new planning sprint",
	Long: `Create a new sprint in Planning status for future work.

The sprint will be created with the @ prefix and appear in the middle section
of the sprint list. You can later activate it using 'sprint start <id>'.

Examples:
  gitta sprint plan "Dashboard Redesign"     # Create planning sprint with auto-generated ID
  gitta sprint plan "Payment" --id Sprint_25  # Create planning sprint with specific ID
  gitta sprint plan "Login Feature" --json    # Output result as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		repoPath, err := findRepoRoot()
		if err != nil {
			return fmt.Errorf("not a git repository: %w", err)
		}

		description := args[0]
		sprintID, _ := cmd.Flags().GetString("id")

		sprintRepo := filesystem.NewDefaultRepository()
		planService := services.NewSprintPlanService(sprintRepo, repoPath)

		req := services.CreatePlanningSprintRequest{
			ID:          sprintID,
			Description: description,
		}

		sprint, err := planService.CreatePlanningSprint(ctx, req)
		if err != nil {
			return err
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]interface{}{
				"sprint": map[string]interface{}{
					"id":          sprint.Name,
					"name":        filepath.Base(sprint.DirectoryPath),
					"path":        sprint.DirectoryPath,
					"status":      "planning",
					"description": description,
				},
			})
		}

		fmt.Printf("Created planning sprint: %s\n", filepath.Base(sprint.DirectoryPath))
		fmt.Printf("Sprint will appear in the Planning section.\n")
		return nil
	},
}

var sprintBoardCmd = &cobra.Command{
	Use:   "board",
	Short: "Display interactive kanban board (spike prototype)",
	Long: `Display an interactive three-column kanban board TUI for evaluating Bubble Tea capabilities.

This is a research spike prototype that demonstrates:
- Three-column layout rendering
- Hardcoded task data display
- Cursor navigation with arrow keys

Use arrow keys to navigate between columns (←/→) and tasks (↑/↓).
Press 'q' or Esc to quit.

Examples:
  gitta sprint board              # Launch the interactive board TUI`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		// Launch the board TUI
		if err := tui.ShowBoard(ctx); err != nil {
			return fmt.Errorf("board display failed: %w", err)
		}

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
	sprintStartCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")

	// Sprint close flags
	sprintCloseCmd.Flags().StringP("target-sprint", "t", "", "Target sprint name for rollover")
	sprintCloseCmd.Flags().Bool("all", false, "Rollover all unfinished tasks without prompting")
	sprintCloseCmd.Flags().Bool("skip", false, "Skip rollover, just close the sprint")

	// Sprint plan flags
	sprintPlanCmd.Flags().String("id", "", "Specify sprint ID manually (default: auto-generate next sequential number)")

	// Sprint burndown flags
	sprintBurndownCmd.Flags().StringP("sprint", "s", "", "Sprint name to analyze (alternative to positional argument)")
	sprintBurndownCmd.Flags().String("format", "ascii", "Output format (ascii, json, csv)")
	sprintBurndownCmd.Flags().Bool("points-only", false, "Show only story points (hide task count)")
	sprintBurndownCmd.Flags().Bool("tasks-only", false, "Show only task count (hide story points)")

	var sprintPlanCmd = &cobra.Command{
		Use:   "plan <name>",
		Short: "Create a new planning sprint",
		Long: `Create a new sprint in Planning status for future work.

The sprint will be created with the @ prefix and appear in the middle section
of the sprint list. You can later activate it using 'sprint start <id>'.

Examples:
  gitta sprint plan "Dashboard Redesign"     # Create planning sprint with auto-generated ID
  gitta sprint plan "Payment" --id Sprint_25  # Create planning sprint with specific ID
  gitta sprint plan "Login Feature" --json    # Output result as JSON`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			repoPath, err := findRepoRoot()
			if err != nil {
				return fmt.Errorf("not a git repository: %w", err)
			}

			description := args[0]
			sprintID, _ := cmd.Flags().GetString("id")

			sprintRepo := filesystem.NewDefaultRepository()
			planService := services.NewSprintPlanService(sprintRepo, repoPath)

			req := services.CreatePlanningSprintRequest{
				ID:          sprintID,
				Description: description,
			}

			sprint, err := planService.CreatePlanningSprint(ctx, req)
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]interface{}{
					"sprint": map[string]interface{}{
						"id":          sprint.Name,
						"name":        filepath.Base(sprint.DirectoryPath),
						"path":        sprint.DirectoryPath,
						"status":      "planning",
						"description": description,
					},
				})
			}

			fmt.Printf("Created planning sprint: %s\n", filepath.Base(sprint.DirectoryPath))
			fmt.Printf("Sprint will appear in the Planning section.\n")
			return nil
		},
	}

	// Register subcommands
	sprintCmd.AddCommand(sprintStartCmd)
	sprintCmd.AddCommand(sprintPlanCmd)
	sprintCmd.AddCommand(sprintCloseCmd)
	sprintCmd.AddCommand(sprintBurndownCmd)
	sprintCmd.AddCommand(sprintBoardCmd)
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
