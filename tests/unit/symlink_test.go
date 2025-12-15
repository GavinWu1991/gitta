package unit

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
)

func TestCreateCurrentSprintLink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows-specific tests should be in a separate file")
	}

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	linkPath := filepath.Join(tmpDir, "Current") // Use "Current" instead of ".current-sprint"

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	tests := []struct {
		name     string
		target   string
		link     string
		wantType filesystem.LinkType
		wantErr  bool
		setup    func() // Additional setup before test
		cleanup  func() // Cleanup after test
	}{
		{
			name:     "create symlink successfully",
			target:   targetDir,
			link:     linkPath,
			wantType: filesystem.LinkTypeSymlink,
			wantErr:  false,
		},
		{
			name:     "target does not exist",
			target:   filepath.Join(tmpDir, "nonexistent"),
			link:     linkPath,
			wantType: filesystem.LinkTypeTextConfig,
			wantErr:  true,
		},
		{
			name:     "target is a file not directory",
			target:   filepath.Join(tmpDir, "file.txt"),
			link:     linkPath,
			wantType: filesystem.LinkTypeTextConfig,
			wantErr:  true,
			setup: func() {
				os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup previous link
			os.Remove(tt.link)
			os.Remove(tt.link + ".txt")

			if tt.setup != nil {
				tt.setup()
			}

			gotType, err := filesystem.CreateCurrentSprintLink(tt.target, tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCurrentSprintLink(%q, %q) error = %v, wantErr %v", tt.target, tt.link, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotType != tt.wantType {
					// On some systems, symlink might fail and fallback to text config
					// That's acceptable behavior
					if gotType != filesystem.LinkTypeTextConfig {
						t.Errorf("CreateCurrentSprintLink(%q, %q) = %v, want %v", tt.target, tt.link, gotType, tt.wantType)
					}
				}

				// Verify link can be read (using sprintsDir, not linkPath)
				sprintsDir := filepath.Dir(tt.link)
				readTarget, readType, readErr := filesystem.ReadCurrentSprintLink(sprintsDir)
				if readErr != nil {
					t.Errorf("ReadCurrentSprintLink(%q) error = %v", sprintsDir, readErr)
					return
				}
				if readType != gotType {
					t.Errorf("ReadCurrentSprintLink returned type %v, expected %v", readType, gotType)
				}
				// Normalize paths for comparison
				absTarget, _ := filepath.Abs(tt.target)
				absReadTarget, _ := filepath.Abs(readTarget)
				if absTarget != absReadTarget {
					t.Errorf("ReadCurrentSprintLink returned target %q, expected %q", absReadTarget, absTarget)
				}
			}

			if tt.cleanup != nil {
				tt.cleanup()
			}
		})
	}
}

func TestReadCurrentSprintLink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows-specific tests should be in a separate file")
	}

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	linkPath := filepath.Join(tmpDir, "Current") // Use "Current" instead of ".current-sprint"

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	tests := []struct {
		name       string
		setup      func() // Create link before reading
		wantTarget string
		wantType   filesystem.LinkType
		wantErr    bool
	}{
		{
			name: "read symlink",
			setup: func() {
				os.Symlink(targetDir, linkPath)
			},
			wantTarget: targetDir,
			wantType:   filesystem.LinkTypeSymlink,
			wantErr:    false,
		},
		{
			name: "read text config",
			setup: func() {
				relPath, _ := filepath.Rel(tmpDir, targetDir)
				os.WriteFile(linkPath+".txt", []byte(relPath), 0644)
			},
			wantTarget: targetDir,
			wantType:   filesystem.LinkTypeTextConfig,
			wantErr:    false,
		},
		{
			name: "link does not exist",
			setup: func() {
				// No setup - link doesn't exist
			},
			wantTarget: "",
			wantType:   filesystem.LinkTypeTextConfig,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			os.Remove(linkPath)
			os.Remove(linkPath + ".txt")

			if tt.setup != nil {
				tt.setup()
			}

			// Use sprintsDir instead of linkPath
			sprintsDir := tmpDir
			gotTarget, gotType, err := filesystem.ReadCurrentSprintLink(sprintsDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCurrentSprintLink(%q) error = %v, wantErr %v", sprintsDir, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotType != tt.wantType {
					t.Errorf("ReadCurrentSprintLink(%q) type = %v, want %v", sprintsDir, gotType, tt.wantType)
				}
				// Normalize paths for comparison
				absTarget, _ := filepath.Abs(tt.wantTarget)
				absGotTarget, _ := filepath.Abs(gotTarget)
				if absTarget != absGotTarget {
					t.Errorf("ReadCurrentSprintLink(%q) target = %q, want %q", sprintsDir, absGotTarget, absTarget)
				}
			}
		})
	}
}
