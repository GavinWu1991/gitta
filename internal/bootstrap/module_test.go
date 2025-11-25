package bootstrap

import (
	"os"
	"testing"
)

func TestValidateModule(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (string, func())
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid go.mod exists",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				// Create valid go.mod
				goModContent := `module github.com/gavin/gitta

go 1.21
`
				os.WriteFile("go.mod", []byte(goModContent), 0644)

				return tmpDir, func() {
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
		{
			name: "go.mod missing",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)
				return tmpDir, func() {
					os.Chdir(oldDir)
				}
			},
			expectError: true,
			errorMsg:    "go.mod not found",
		},
		{
			name: "wrong module path",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				goModContent := `module github.com/wrong/path

go 1.21
`
				os.WriteFile("go.mod", []byte(goModContent), 0644)

				return tmpDir, func() {
					os.Chdir(oldDir)
				}
			},
			expectError: true,
			errorMsg:    "module path does not match",
		},
		{
			name: "Go version too old",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				goModContent := `module github.com/gavin/gitta

go 1.20
`
				os.WriteFile("go.mod", []byte(goModContent), 0644)

				return tmpDir, func() {
					os.Chdir(oldDir)
				}
			},
			expectError: true,
			errorMsg:    "requires Go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := tt.setup(t)
			defer cleanup()

			err := ValidateModule()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("error message should contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCheckExistingModule(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (string, func())
		expectExists bool
		expectError  bool
		errorMsg     string
	}{
		{
			name: "no go.mod exists",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)
				return tmpDir, func() {
					os.Chdir(oldDir)
				}
			},
			expectExists: false,
			expectError:  false,
		},
		{
			name: "correct go.mod exists",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				goModContent := `module github.com/gavin/gitta

go 1.21
`
				os.WriteFile("go.mod", []byte(goModContent), 0644)

				return tmpDir, func() {
					os.Chdir(oldDir)
				}
			},
			expectExists: true,
			expectError:  false,
		},
		{
			name: "wrong module path in existing go.mod",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				goModContent := `module github.com/other/project

go 1.21
`
				os.WriteFile("go.mod", []byte(goModContent), 0644)

				return tmpDir, func() {
					os.Chdir(oldDir)
				}
			},
			expectExists: true,
			expectError:  true,
			errorMsg:     "module path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := tt.setup(t)
			defer cleanup()

			exists, err := CheckExistingModule()
			if exists != tt.expectExists {
				t.Errorf("expected exists=%v, got %v", tt.expectExists, exists)
			}
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("error message should contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
