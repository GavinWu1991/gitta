//go:build windows
// +build windows

package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modkernel32             = syscall.NewLazyDLL("kernel32.dll")
	procCreateSymbolicLinkW = modkernel32.NewProc("CreateSymbolicLinkW")
)

const (
	SYMBOLIC_LINK_FLAG_DIRECTORY = 0x1
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

// createJunction creates a Windows junction link.
func createJunction(targetPath, linkPath string) error {
	// Remove existing link if it exists
	os.Remove(linkPath)

	// Create junction using syscall
	targetPtr, err := syscall.UTF16PtrFromString(targetPath)
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	linkPtr, err := syscall.UTF16PtrFromString(linkPath)
	if err != nil {
		return fmt.Errorf("invalid link path: %w", err)
	}

	// Call CreateSymbolicLinkW
	ret, _, err := procCreateSymbolicLinkW.Call(
		uintptr(unsafe.Pointer(linkPtr)),
		uintptr(unsafe.Pointer(targetPtr)),
		SYMBOLIC_LINK_FLAG_DIRECTORY,
	)
	if ret == 0 {
		if err != nil && err.Error() != "The operation completed successfully." {
			return fmt.Errorf("failed to create junction: %w", err)
		}
		// Check last error
		if errno, ok := err.(syscall.Errno); ok && errno != 0 {
			return fmt.Errorf("failed to create junction: %w", errno)
		}
	}

	return nil
}

// readJunction reads the target of a Windows junction link.
func readJunction(linkPath string) (string, error) {
	// For junctions, we can use os.Readlink on Windows 10+
	target, err := os.Readlink(linkPath)
	if err != nil {
		return "", err
	}
	return target, nil
}

// CreateCurrentSprintLink creates a link to the current sprint directory using
// the best available mechanism for Windows.
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

	// Try junction first (works without admin privileges)
	if err := createJunction(targetPath, linkPath); err == nil {
		return LinkTypeJunction, nil
	}
	// Try symlink (requires admin or Developer Mode)
	if err := os.Symlink(targetPath, linkPath); err == nil {
		return LinkTypeSymlink, nil
	}
	// Fallback to text config
	return LinkTypeTextConfig, createTextConfig(targetPath, linkPath)
}

// ReadCurrentSprintLink reads the target path from a current sprint link on Windows.
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

	// Check if it's a junction
	if target, err := readJunction(linkPath); err == nil {
		return target, LinkTypeJunction, nil
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

	return "", LinkTypeTextConfig, fmt.Errorf("failed to read link at %s: %w", linkPath, err)
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
