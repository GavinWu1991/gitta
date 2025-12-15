package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gavin/gitta/internal/services"
)

var (
	migrateForce bool
	migrateDry   bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate legacy backlog/sprints to consolidated tasks/ structure",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		repoPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine working directory: %w", err)
		}

		svc := services.NewMigrateService()
		result, err := svc.Migrate(ctx, repoPath, services.MigrateOptions{
			Force:  migrateForce,
			DryRun: migrateDry,
		})
		if err != nil {
			return err
		}

		if jsonOutput {
			out := map[string]interface{}{
				"success":           result.Success,
				"moved_directories": result.MovedDirectories,
				"backup_paths":      result.BackupPaths,
			}
			data, _ := json.Marshal(out)
			fmt.Println(string(data))
			return nil
		}

		if migrateDry {
			fmt.Println("ℹ️ Dry run: would migrate backlog/ and sprints/ to tasks/ structure.")
			return nil
		}

		fmt.Println("✅ Migration completed successfully!")
		if len(result.MovedDirectories) > 0 {
			fmt.Println("\nMoved directories:")
			for _, d := range result.MovedDirectories {
				fmt.Printf("  - %s\n", d)
			}
		}
		if len(result.BackupPaths) > 0 {
			fmt.Println("\nBackups:")
			for _, b := range result.BackupPaths {
				fmt.Printf("  - %s\n", b)
			}
		}
		fmt.Println("\nNext steps:")
		fmt.Println("  1) Verify migration: gitta list")
		fmt.Println("  2) Commit changes: git add tasks/ && git commit -m \"Migrate to consolidated structure\"")
		return nil
	},
}

func init() {
	migrateCmd.Flags().BoolVar(&migrateForce, "force", false, "Overwrite existing tasks/ directories (creates backups)")
	migrateCmd.Flags().BoolVar(&migrateDry, "dry-run", false, "Simulate migration without making changes")
}
