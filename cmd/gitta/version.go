package main

import (
	"fmt"
	"time"

	"github.com/gavin/gitta/internal/version"
	"github.com/spf13/cobra"
)

var (
	// These will be set via ldflags during build
	buildVersion = "dev"
	buildCommit  = ""
	buildDate    = ""
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Print version information including semantic version, commit SHA,
build date, and Go runtime version. Use --json for machine-readable output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := version.NewInfo()

		// Override with build-time values if available
		if buildVersion != "dev" {
			info.Version = buildVersion
		}
		if buildCommit != "" {
			info.Commit = buildCommit
		}
		if buildDate != "" {
			// Parse build date (expected format: RFC3339)
			if t, err := parseBuildDate(buildDate); err == nil {
				info.BuildDate = t
			}
		}

		if jsonOutput {
			jsonStr, err := info.FormatJSON()
			if err != nil {
				return fmt.Errorf("failed to format version as JSON: %w", err)
			}
			fmt.Println(jsonStr)
		} else {
			fmt.Print(info.Format())
		}
		return nil
	},
}

func parseBuildDate(dateStr string) (time.Time, error) {
	// Parse RFC3339 format (e.g., "2024-01-15T10:30:00Z")
	return time.Parse(time.RFC3339, dateStr)
}
