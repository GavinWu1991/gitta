package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/gavin/gitta/internal/core"
)

const (
	// lockTimeout is the maximum time to wait for a file lock (5 seconds).
	lockTimeout = 5 * time.Second
	// maxRetries is the maximum number of retry attempts for ID generation.
	maxRetries = 3
	// retryBaseDelay is the base delay for exponential backoff (100ms).
	retryBaseDelay = 100 * time.Millisecond
)

var (
	// prefixPattern validates that prefix is exactly 2 uppercase letters (to match validator pattern).
	prefixPattern = regexp.MustCompile(`^[A-Z]{2}$`)
)

// IDCounter implements core.IDGenerator using a file-based counter with advisory file locking.
// It stores counters in a JSON file at `.gitta/id-counters.json` and uses file locking
// to ensure thread-safe concurrent access.
type IDCounter struct {
	repoPath string
	mu       sync.Mutex // Protects the counter file path
}

// counterFile represents the structure of the counter file.
type counterFile struct {
	Counters map[string]int `json:"counters"`
}

// NewIDCounter creates a new IDCounter instance for the given repository path.
func NewIDCounter(repoPath string) *IDCounter {
	return &IDCounter{
		repoPath: repoPath,
	}
}

// GenerateNextID generates the next unique ID for the given prefix.
// It uses file locking to ensure thread-safe counter increments.
func (c *IDCounter) GenerateNextID(ctx context.Context, prefix string) (string, error) {
	// Validate context
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context cancelled: %w", err)
	}

	// Validate prefix format
	if !prefixPattern.MatchString(prefix) {
		return "", core.ErrInvalidPrefix
	}

	// Retry logic with exponential backoff
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := retryBaseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(delay):
			}
		}

		id, err := c.generateIDWithLock(ctx, prefix)
		if err == nil {
			return id, nil
		}

		// If it's a lock timeout, retry
		if err == core.ErrLockTimeout {
			lastErr = err
			continue
		}

		// For other errors, return immediately
		return "", err
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// generateIDWithLock performs the actual ID generation with file locking.
func (c *IDCounter) generateIDWithLock(ctx context.Context, prefix string) (string, error) {
	counterPath := filepath.Join(c.repoPath, ".gitta", "id-counters.json")

	// Create .gitta directory if it doesn't exist
	gittaDir := filepath.Dir(counterPath)
	if err := os.MkdirAll(gittaDir, 0755); err != nil {
		return "", &core.IOError{
			Operation: "create",
			FilePath:  gittaDir,
			Cause:     err,
		}
	}

	// Open or create counter file
	file, err := os.OpenFile(counterPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return "", &core.IOError{
			Operation: "open",
			FilePath:  counterPath,
			Cause:     err,
		}
	}
	defer file.Close()

	// Acquire exclusive lock with timeout using lock file
	lockFile := counterPath + ".lock"
	lockCtx, cancel := context.WithTimeout(ctx, lockTimeout)
	defer cancel()

	lockAcquired := make(chan error, 1)
	go func() {
		lockAcquired <- c.acquireLockFile(lockFile)
	}()

	select {
	case <-lockCtx.Done():
		return "", core.ErrLockTimeout
	case err := <-lockAcquired:
		if err != nil {
			return "", fmt.Errorf("failed to acquire lock: %w", err)
		}
	}
	defer func() {
		// Release lock by removing lock file
		os.Remove(lockFile)
	}()

	// Read current counters
	counters, err := c.readCounters(file)
	if err != nil {
		return "", err
	}

	// Increment counter for prefix
	current := counters[prefix]
	newValue := current + 1
	counters[prefix] = newValue

	// Write back to file
	if err := c.writeCounters(file, counters); err != nil {
		return "", err
	}

	// Release lock when file is closed
	// On Unix: closing the file releases the Flock
	// On Windows: closing the file releases the LockFileEx

	// Generate ID (format: PREFIX-NUMBER, e.g., STORY-1, STORY-2)
	id := fmt.Sprintf("%s-%d", prefix, newValue)

	// Check for conflicts (verify ID doesn't already exist)
	// This is a simple check - in production, you might want to scan for existing stories
	// For now, we'll rely on the counter being unique

	return id, nil
}

// acquireLockFile acquires an exclusive lock using a lock file.
// This is a simple cross-platform approach that works on all systems.
// The lock file is created exclusively - if it exists, another process has the lock.
func (c *IDCounter) acquireLockFile(lockFile string) error {
	// Retry loop with small delay
	maxAttempts := 50 // 50 attempts * 100ms = 5 seconds max
	for i := 0; i < maxAttempts; i++ {
		// Try to create lock file exclusively
		file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			// Lock acquired
			file.Close()
			return nil
		}

		// If file exists, another process has the lock - wait and retry
		if os.IsExist(err) {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Other error
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	return core.ErrLockTimeout
}

// readCounters reads the counter file and returns the counters map.
func (c *IDCounter) readCounters(file *os.File) (map[string]int, error) {
	// Seek to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek: %w", err)
	}

	// Read file
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if stat.Size() == 0 {
		// Empty file, return empty map
		return make(map[string]int), nil
	}

	var cf counterFile
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cf); err != nil {
		return nil, core.ErrCounterCorrupted
	}

	if cf.Counters == nil {
		cf.Counters = make(map[string]int)
	}

	return cf.Counters, nil
}

// writeCounters writes the counters map to the file.
func (c *IDCounter) writeCounters(file *os.File, counters map[string]int) error {
	// Truncate file
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate: %w", err)
	}

	// Seek to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	// Write JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	cf := counterFile{Counters: counters}
	if err := encoder.Encode(&cf); err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	// Sync to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	return nil
}
