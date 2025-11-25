package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	expectedModulePath = "github.com/gavin/gitta"
	minGoVersion       = "1.21"
)

// ValidateModule checks that the current directory has a valid go.mod
// with the expected module path and Go version.
func ValidateModule() error {
	// Check if go.mod exists
	goModPath := "go.mod"
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found. Run 'go mod init %s' first", expectedModulePath)
	}

	// Read go.mod content
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Check module path
	if !strings.Contains(string(content), expectedModulePath) {
		return fmt.Errorf("go.mod module path does not match expected '%s'", expectedModulePath)
	}

	// Check Go version
	if !strings.Contains(string(content), "go 1.21") && !strings.Contains(string(content), "go 1.22") {
		return fmt.Errorf("go.mod requires Go %s or higher", minGoVersion)
	}

	// Check GOEXPERIMENT is not set
	if exp := os.Getenv("GOEXPERIMENT"); exp != "" {
		return fmt.Errorf("GOEXPERIMENT is set to '%s' but must be disabled per constitution", exp)
	}

	return nil
}

// CheckExistingModule detects if go.mod already exists and provides guidance.
func CheckExistingModule() (bool, error) {
	goModPath := "go.mod"
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return false, nil
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		return true, fmt.Errorf("failed to read existing go.mod: %w", err)
	}

	// Check if module path matches
	if strings.Contains(string(content), expectedModulePath) {
		return true, nil // Already initialized correctly
	}

	// Extract current module path
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "module ") {
			currentPath := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			return true, fmt.Errorf(
				"go.mod already exists with module path '%s', expected '%s'. "+
					"Either update go.mod or initialize in a different directory",
				currentPath, expectedModulePath,
			)
		}
	}

	return true, fmt.Errorf("go.mod exists but module path not found")
}

// VerifyGoVersion checks that the installed Go version meets requirements.
func VerifyGoVersion() error {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check Go version: %w", err)
	}

	versionStr := string(output)
	// Go version output format: "go version go1.21.5 darwin/amd64"
	if !strings.Contains(versionStr, "go1.21") && !strings.Contains(versionStr, "go1.22") && !strings.Contains(versionStr, "go1.23") {
		return fmt.Errorf("Go version must be %s or higher, found: %s", minGoVersion, versionStr)
	}

	return nil
}
