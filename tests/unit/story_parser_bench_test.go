package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavin/gitta/infra/filesystem"
	"github.com/gavin/gitta/internal/core"
)

func BenchmarkReadStory_100Files(b *testing.B) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	// Get path to test file
	testFile := filepath.Join("..", "..", "testdata", "parser", "valid-story.md")
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate reading 100 files
		for j := 0; j < 100; j++ {
			_, err := parser.ReadStory(ctx, absPath)
			if err != nil {
				b.Fatalf("ReadStory failed: %v", err)
			}
		}
	}
}

func BenchmarkReadStory_SingleFile(b *testing.B) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "parser", "valid-story.md")
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ReadStory(ctx, absPath)
		if err != nil {
			b.Fatalf("ReadStory failed: %v", err)
		}
	}
}

func BenchmarkWriteStory_SingleFile(b *testing.B) {
	parser := filesystem.NewMarkdownParser()
	ctx := context.Background()

	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench-story.md")

	assignee := "bench-user"
	story := &core.Story{
		ID:       "BN-001",
		Title:    "Benchmark Story",
		Assignee: &assignee,
		Priority: core.PriorityMedium,
		Status:   core.StatusTodo,
		Body:     "## Description\n\nThis is a benchmark story.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := parser.WriteStory(ctx, testFile, story)
		if err != nil {
			b.Fatalf("WriteStory failed: %v", err)
		}
		// Clean up for next iteration
		os.Remove(testFile)
	}
}
