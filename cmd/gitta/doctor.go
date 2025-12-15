package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
	"github.com/gavin/gitta/internal/services"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Detect and repair sprint status inconsistencies",
	Long: `Detect and repair inconsistencies between visual indicators (folder name prefixes)
and authoritative status files (.gitta/status).

Scans all sprints and compares folder name prefixes with .gitta/status files.
Reports all inconsistencies found. Use --fix to automatically repair them.

Examples:
  gitta doctor                    # Check for inconsistencies (report only)
  gitta doctor --fix              # Check and automatically fix
  gitta doctor --sprint Sprint_24 # Check specific sprint only
  gitta doctor --json             # Output result as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		repoPath, err := findRepoRoot()
		if err != nil {
			return fmt.Errorf("not a git repository: %w", err)
		}

		fix, _ := cmd.Flags().GetBool("fix")
		sprintPath, _ := cmd.Flags().GetString("sprint")

		sprintRepo := filesystem.NewDefaultRepository()
		doctorService := services.NewSprintDoctorService(sprintRepo, repoPath)

		var inconsistencies []services.Inconsistency
		if sprintPath != "" {
			// Check specific sprint
			sprintsDir := filepath.Join(repoPath, "sprints")
			fullSprintPath := filepath.Join(sprintsDir, sprintPath)
			// For single sprint, we'll detect all and filter, or implement a single-sprint check
			// For now, detect all and filter
			all, err := doctorService.DetectInconsistencies(ctx)
			if err != nil {
				return err
			}
			for _, inc := range all {
				if inc.SprintPath == fullSprintPath || filepath.Base(inc.SprintPath) == sprintPath {
					inconsistencies = append(inconsistencies, inc)
				}
			}
		} else {
			// Check all sprints
			var err error
			inconsistencies, err = doctorService.DetectInconsistencies(ctx)
			if err != nil {
				return fmt.Errorf("failed to detect inconsistencies: %w", err)
			}
		}

		// Validate Current link
		sprintsDir := filepath.Join(repoPath, "sprints")
		currentPath, _, err := filesystem.ReadCurrentSprintLink(sprintsDir)
		currentLinkValid := err == nil && currentPath != ""
		if currentLinkValid {
			// Verify Current link points to an Active sprint
			currentStatus, err := sprintRepo.ReadSprintStatus(ctx, currentPath)
			if err == nil && currentStatus != core.StatusActive {
				// Current link points to non-active sprint - this is an inconsistency
				// We could add this to inconsistencies list, but for now just note it
				currentLinkValid = false
			}
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			output := map[string]interface{}{
				"status":             "ok",
				"sprints_checked":    len(inconsistencies), // Count of inconsistent sprints
				"inconsistencies":    inconsistencies,
				"current_link_valid": currentLinkValid,
			}
			if len(inconsistencies) > 0 {
				output["status"] = "inconsistencies_found"
			}
			return enc.Encode(output)
		}

		// Human-readable output
		fmt.Println("Checking sprint status consistency...")

		if len(inconsistencies) == 0 {
			fmt.Println("✓ All sprints are consistent")
			if currentLinkValid {
				fmt.Println("✓ Current link points to active sprint")
			} else {
				fmt.Println("✗ Current link is invalid or missing")
			}
			return nil
		}

		fmt.Printf("✗ Found %d inconsistencies:\n\n", len(inconsistencies))
		for i, inc := range inconsistencies {
			fmt.Printf("%d. Sprint %q\n", i+1, filepath.Base(inc.SprintPath))
			fmt.Printf("   - Folder name: %s (%s)\n", inc.FolderName, inc.FolderStatus.String())
			fmt.Printf("   - Status file: %s\n", inc.StatusFile.String())
			fmt.Printf("   → Should rename to: %s\n\n", inc.ExpectedName)
		}

		if !fix {
			fmt.Println("Run with --fix to repair these issues.")
			return fmt.Errorf("inconsistencies found")
		}

		// Repair inconsistencies
		fmt.Println("Repairing inconsistencies...")
		result, err := doctorService.RepairInconsistencies(ctx, inconsistencies)
		if err != nil {
			return fmt.Errorf("failed to repair inconsistencies: %w", err)
		}

		for i, inc := range inconsistencies {
			if i < result.RepairedCount {
				fmt.Printf("✓ Renamed %q to %q\n", inc.FolderName, inc.ExpectedName)
			}
		}

		if result.FailedCount > 0 {
			fmt.Printf("\n✗ %d repairs failed:\n", result.FailedCount)
			for _, repairErr := range result.Errors {
				fmt.Printf("  - %v\n", repairErr)
			}
			return fmt.Errorf("some repairs failed")
		}

		if result.RepairedCount > 0 {
			fmt.Printf("\n✓ All inconsistencies repaired\n")
		}

		return nil
	},
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "Automatically repair detected inconsistencies")
	doctorCmd.Flags().String("sprint", "", "Check specific sprint only (default: check all sprints)")
	rootCmd.AddCommand(doctorCmd)
}
