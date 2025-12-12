package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "gitta")
	// Add .exe extension on Windows
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	buildCmd := exec.Command("go", "build", "-o", binPath, "../../cmd/gitta")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	// Verify binary exists before trying to execute
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("binary not found at %s: %v", binPath, err)
	}

	tests := []struct {
		name        string
		args        []string
		wantExit    int
		checkOutput func(t *testing.T, output string)
	}{
		{
			name:     "version command succeeds",
			args:     []string{"version"},
			wantExit: 0,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "gitta") {
					t.Errorf("output should contain 'gitta', got: %s", output)
				}
			},
		},
		{
			name:     "version --json returns valid JSON",
			args:     []string{"version", "--json"},
			wantExit: 0,
			checkOutput: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("output should be valid JSON: %v", err)
				}
				// Check required fields
				requiredFields := []string{"version", "goVersion"}
				for _, field := range requiredFields {
					if _, ok := result[field]; !ok {
						t.Errorf("JSON missing required field: %s", field)
					}
				}
			},
		},
		{
			name:     "--help shows usage",
			args:     []string{"--help"},
			wantExit: 0,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Usage:") {
					t.Errorf("help output should contain 'Usage:', got: %s", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tt.args...)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("command failed: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d. Output: %s", exitCode, tt.wantExit, string(output))
			}

			if tt.checkOutput != nil {
				tt.checkOutput(t, string(output))
			}
		})
	}
}
