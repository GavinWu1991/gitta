package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gavin/gitta/internal/core"
)

// ReadSprintStatus reads the sprint status from the .gitta/status file.
// If the file doesn't exist, it infers the status from the folder name prefix.
// Returns the status and an error if the status cannot be determined.
func ReadSprintStatus(ctx context.Context, sprintDir string) (core.SprintStatus, error) {
	if err := ctx.Err(); err != nil {
		return core.StatusActive, err
	}

	statusFilePath := filepath.Join(sprintDir, ".gitta", "status")

	// Try to read from status file first (authoritative source)
	if data, err := os.ReadFile(statusFilePath); err == nil {
		statusStr := strings.TrimSpace(string(data))
		status, err := core.ParseStatus(statusStr)
		if err != nil {
			return core.StatusActive, fmt.Errorf("invalid status in %s: %w", statusFilePath, err)
		}
		return status, nil
	}

	// If status file doesn't exist, infer from folder name prefix
	dirName := filepath.Base(sprintDir)
	status := ExtractStatus(dirName)
	if status == core.StatusActive && !strings.HasPrefix(dirName, "!") {
		// If we couldn't determine status from prefix, return error
		return core.StatusActive, fmt.Errorf("cannot determine status for sprint %s: no .gitta/status file and folder name doesn't have status prefix", sprintDir)
	}

	return status, nil
}

// WriteSprintStatus writes the sprint status to the .gitta/status file.
// Creates the .gitta directory if it doesn't exist.
// Returns an error if the write operation fails.
func WriteSprintStatus(ctx context.Context, sprintDir string, status core.SprintStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	statusDir := filepath.Join(sprintDir, ".gitta")
	statusFilePath := filepath.Join(statusDir, "status")

	// Create .gitta directory if it doesn't exist
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		return &core.IOError{
			Operation: "create",
			FilePath:  statusDir,
			Cause:     err,
		}
	}

	// Write status to file
	statusStr := status.String() + "\n"
	if err := os.WriteFile(statusFilePath, []byte(statusStr), 0644); err != nil {
		return &core.IOError{
			Operation: "write",
			FilePath:  statusFilePath,
			Cause:     err,
		}
	}

	return nil
}

// StatusFileExists checks if the .gitta/status file exists for a sprint.
func StatusFileExists(ctx context.Context, sprintDir string) bool {
	if err := ctx.Err(); err != nil {
		return false
	}

	statusFilePath := filepath.Join(sprintDir, ".gitta", "status")
	_, err := os.Stat(statusFilePath)
	return err == nil
}

// ParseFolderName parses a sprint folder name and extracts the status prefix, ID, and description.
// Format: {PREFIX}{Sprint_ID}_{Description}
// Returns: status, sprint ID, description, error
// The ID can contain underscores (e.g., "Sprint_24"), so we split from the right.
func ParseFolderName(name string) (core.SprintStatus, string, string, error) {
	if len(name) == 0 {
		return core.StatusActive, "", "", fmt.Errorf("folder name cannot be empty")
	}

	// Extract status from prefix
	status := ExtractStatus(name)
	if status == core.StatusActive && !strings.HasPrefix(name, "!") {
		return core.StatusActive, "", "", fmt.Errorf("folder name must start with a valid status prefix (!, +, @, ~)")
	}

	// Remove prefix to get the rest of the name
	prefix := status.Prefix()
	rest := name[len(prefix):]

	// Find the last underscore to separate ID from description
	// This handles cases like "Sprint_24_Login" where ID is "Sprint_24" and description is "Login"
	// For "Sprint_24" (no description), the entire rest is the ID
	lastUnderscore := strings.LastIndex(rest, "_")
	if lastUnderscore == -1 {
		// No underscore found, entire rest is the ID
		return status, rest, "", nil
	}

	// Check if the part after the last underscore looks like a description (not purely numeric)
	// If it's purely numeric, it's part of the ID (e.g., "Sprint_24" -> ID="Sprint_24", no desc)
	// If it's not numeric, it's the description (e.g., "Sprint_24_Login" -> ID="Sprint_24", desc="Login")
	afterLastUnderscore := rest[lastUnderscore+1:]
	isNumeric := true
	for _, r := range afterLastUnderscore {
		if r < '0' || r > '9' {
			isNumeric = false
			break
		}
	}

	if isNumeric {
		// The numeric part is part of the ID, not a description
		// Entire rest is the ID
		return status, rest, "", nil
	}

	// Split at last underscore: ID is before, description is after
	sprintID := rest[:lastUnderscore]
	description := rest[lastUnderscore+1:]

	return status, sprintID, description, nil
}

// BuildFolderName constructs a sprint folder name from status, ID, and description.
// Format: {PREFIX}{Sprint_ID}_{Description}
func BuildFolderName(status core.SprintStatus, id string, desc string) string {
	if desc == "" {
		return status.Prefix() + id
	}
	return status.Prefix() + id + "_" + desc
}

// ExtractStatus extracts the sprint status from a folder name prefix.
// Returns StatusActive if no valid prefix is found (for backward compatibility).
func ExtractStatus(folderName string) core.SprintStatus {
	if len(folderName) == 0 {
		return core.StatusActive
	}

	firstChar := folderName[0]
	switch firstChar {
	case '!':
		return core.StatusActive
	case '+':
		return core.StatusReady
	case '@':
		return core.StatusPlanning
	case '~':
		return core.StatusArchived
	default:
		// No prefix found - default to Active for backward compatibility
		return core.StatusActive
	}
}

// ValidateSprintName validates that a sprint name/description doesn't contain status prefix characters.
// Returns an error if invalid characters are found.
func ValidateSprintName(name string) error {
	invalidChars := []string{"!", "+", "@", "~"}
	var found []string

	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			found = append(found, char)
		}
	}

	if len(found) > 0 {
		return fmt.Errorf("sprint name/description cannot contain status prefix characters: %s", strings.Join(found, ", "))
	}

	return nil
}
