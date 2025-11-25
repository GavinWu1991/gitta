package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// Info holds build metadata for the application.
type Info struct {
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	BuildDate time.Time `json:"buildDate"`
	GoVersion string    `json:"goVersion"`
}

// NewInfo creates a new Info struct with default values.
// Version, Commit, and BuildDate should be set via ldflags during build.
func NewInfo() Info {
	return Info{
		Version:   "dev",
		Commit:    "",
		BuildDate: time.Now(),
		GoVersion: runtime.Version(),
	}
}

// Format returns a human-readable version string.
func (i Info) Format() string {
	commit := i.Commit
	if commit == "" {
		commit = "<unknown>"
	}
	return fmt.Sprintf("gitta v%s\ncommit: %s\nbuild date: %s\ngo: %s\n",
		i.Version, commit, i.BuildDate.Format(time.RFC3339), i.GoVersion)
}

// FormatJSON returns a JSON representation of the version info.
func (i Info) FormatJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal version info: %w", err)
	}
	return string(jsonBytes), nil
}
