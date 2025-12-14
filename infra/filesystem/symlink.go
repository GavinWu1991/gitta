//go:build !windows
// +build !windows

package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LinkType represents the type of link mechanism used.
type LinkType int

const (
	// LinkTypeSymlink represents a Unix symbolic link.
	LinkTypeSymlink LinkType = iota
	// LinkTypeJunction represents a Windows junction link.
	LinkTypeJunction
	// LinkTypeTextConfig represents a text-based configuration file fallback.
	LinkTypeTextConfig
)

// CreateCurrentSprintLink creates a link to the current sprint directory using
// the best available mechanism for the platform.
// Returns the link type used and any error encountered.
func CreateCurrentSprintLink(targetPath, linkPath string) (LinkType, error) {
	// Normalize paths
	targetPath = filepath.Clean(targetPath)
	linkPath = filepath.Clean(linkPath)

	// Verify target exists
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return LinkTypeTextConfig, fmt.Errorf("target path does not exist: %w", err)
	}
	if !targetInfo.IsDir() {
		return LinkTypeTextConfig, errors.New("target path must be a directory")
	}

	// Unix-like systems: try symlink
	if err := os.Symlink(targetPath, linkPath); err == nil {
		return LinkTypeSymlink, nil
	}
	// Fallback to text config
	return LinkTypeTextConfig, createTextConfig(targetPath, linkPath)
}

// ReadCurrentSprintLink reads the target path from a current sprint link.
// Supports symlinks, junctions, and text config files.
func ReadCurrentSprintLink(linkPath string) (string, LinkType, error) {
	linkPath = filepath.Clean(linkPath)

	// Try reading as symlink first
	target, err := os.Readlink(linkPath)
	if err == nil {
		// Resolve relative symlinks
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(linkPath), target)
		}
		return filepath.Clean(target), LinkTypeSymlink, nil
	}

	// Check if it's a text config file
	textConfigPath := linkPath + ".txt"
	if info, err := os.Stat(textConfigPath); err == nil && !info.IsDir() {
		data, err := os.ReadFile(textConfigPath)
		if err != nil {
			return "", LinkTypeTextConfig, fmt.Errorf("failed to read text config: %w", err)
		}
		target := strings.TrimSpace(string(data))
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(textConfigPath), target)
		}
		return filepath.Clean(target), LinkTypeTextConfig, nil
	}

	// Junctions are Windows-only, handled in symlink_windows.go

	return "", LinkTypeTextConfig, fmt.Errorf("failed to read link at %s: %w", linkPath, err)
}

// createJunction is a stub for non-Windows platforms (junctions are Windows-only).
func createJunction(targetPath, linkPath string) error {
	return errors.New("junctions are only supported on Windows")
}

// readJunction is a stub for non-Windows platforms.
func readJunction(linkPath string) (string, error) {
	return "", errors.New("junctions are only supported on Windows")
}

// createTextConfig creates a text-based configuration file as a fallback.
func createTextConfig(targetPath, linkPath string) error {
	// Use .txt extension for text config
	textConfigPath := linkPath + ".txt"

	// Remove existing file if it exists
	os.Remove(textConfigPath)

	// Write target path to file
	// Use relative path if possible
	relPath, err := filepath.Rel(filepath.Dir(textConfigPath), targetPath)
	if err != nil {
		// Fallback to absolute path
		relPath = targetPath
	}

	return os.WriteFile(textConfigPath, []byte(relPath), 0644)
}
