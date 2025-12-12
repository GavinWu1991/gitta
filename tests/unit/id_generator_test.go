package unit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
)

func TestIDCounter_GenerateNextID(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		wantPattern   string
		wantErr       error
		setupCounters map[string]int
	}{
		{
			name:          "first ID for prefix",
			prefix:        "US",
			wantPattern:   "US-1",
			setupCounters: nil,
		},
		{
			name:          "incremental IDs",
			prefix:        "US",
			wantPattern:   "US-2",
			setupCounters: map[string]int{"US": 1},
		},
		{
			name:          "different prefix",
			prefix:        "BG",
			wantPattern:   "BG-1",
			setupCounters: map[string]int{"US": 5},
		},
		{
			name:          "multiple prefixes",
			prefix:        "TS",
			wantPattern:   "TS-1",
			setupCounters: map[string]int{"US": 10, "BUG": 3},
		},
		{
			name:          "invalid prefix - too short",
			prefix:        "A",
			wantErr:       core.ErrInvalidPrefix,
			setupCounters: nil,
		},
		{
			name:          "invalid prefix - lowercase",
			prefix:        "story",
			wantErr:       core.ErrInvalidPrefix,
			setupCounters: nil,
		},
		{
			name:          "invalid prefix - numbers",
			prefix:        "ST1",
			wantErr:       core.ErrInvalidPrefix,
			setupCounters: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temporary directory
			tmpDir := t.TempDir()
			counter := filesystem.NewIDCounter(tmpDir)

			// Setup initial counters if needed
			if tt.setupCounters != nil {
				setupCounterFile(t, tmpDir, tt.setupCounters)
			}

			ctx := context.Background()
			got, err := counter.GenerateNextID(ctx, tt.prefix)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("GenerateNextID() expected error %v, got nil", tt.wantErr)
					return
				}
				if err != tt.wantErr && !isErrorType(err, tt.wantErr) {
					t.Errorf("GenerateNextID() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateNextID() error = %v, want nil", err)
				return
			}

			if got != tt.wantPattern {
				t.Errorf("GenerateNextID() = %v, want %v", got, tt.wantPattern)
			}

			// Verify counter was incremented
			if tt.setupCounters != nil {
				expected := tt.setupCounters[tt.prefix] + 1
				verifyCounterValue(t, tmpDir, tt.prefix, expected)
			} else {
				verifyCounterValue(t, tmpDir, tt.prefix, 1)
			}
		})
	}
}

func TestIDCounter_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	counter := filesystem.NewIDCounter(tmpDir)
	ctx := context.Background()

	// Generate 10 IDs concurrently
	const numGoroutines = 10
	ids := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			id, err := counter.GenerateNextID(ctx, "US")
			if err != nil {
				errors <- err
				return
			}
			ids <- id
		}()
	}

	// Collect all IDs
	collectedIDs := make(map[string]bool)
	for i := 0; i < numGoroutines; i++ {
		select {
		case id := <-ids:
			if collectedIDs[id] {
				t.Errorf("Duplicate ID generated: %s", id)
			}
			collectedIDs[id] = true
		case err := <-errors:
			t.Errorf("GenerateNextID() error = %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for ID generation")
		}
	}

	// Verify we got 10 unique IDs
	if len(collectedIDs) != numGoroutines {
		t.Errorf("Expected %d unique IDs, got %d", numGoroutines, len(collectedIDs))
	}

	// Verify counter value
	verifyCounterValue(t, tmpDir, "US", numGoroutines)
}

func TestIDCounter_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	counter := filesystem.NewIDCounter(tmpDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := counter.GenerateNextID(ctx, "US")
	if err == nil {
		t.Error("GenerateNextID() expected error on cancelled context, got nil")
	}
}

func setupCounterFile(t *testing.T, repoPath string, counters map[string]int) {
	t.Helper()
	counterPath := filepath.Join(repoPath, ".gitta", "id-counters.json")
	if err := os.MkdirAll(filepath.Dir(counterPath), 0755); err != nil {
		t.Fatalf("Failed to create .gitta directory: %v", err)
	}

	data := map[string]interface{}{
		"counters": counters,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal counters: %v", err)
	}

	if err := os.WriteFile(counterPath, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write counter file: %v", err)
	}
}

func verifyCounterValue(t *testing.T, repoPath, prefix string, expected int) {
	t.Helper()
	counterPath := filepath.Join(repoPath, ".gitta", "id-counters.json")
	data, err := os.ReadFile(counterPath)
	if err != nil {
		t.Fatalf("Failed to read counter file: %v", err)
	}

	var cf struct {
		Counters map[string]int `json:"counters"`
	}
	if err := json.Unmarshal(data, &cf); err != nil {
		t.Fatalf("Failed to unmarshal counter file: %v", err)
	}

	actual := cf.Counters[prefix]
	if actual != expected {
		t.Errorf("Counter value for %s = %d, want %d", prefix, actual, expected)
	}
}

func isErrorType(err, target error) bool {
	return err == target || err.Error() == target.Error()
}
