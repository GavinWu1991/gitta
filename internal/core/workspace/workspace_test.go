package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectStructure_PrecedenceAndDefaults(t *testing.T) {
	t.Run("consolidated wins when present", func(t *testing.T) {
		repo := t.TempDir()
		requireDir(t, filepath.Join(repo, "tasks", "backlog"))
		requireDir(t, filepath.Join(repo, "tasks", "sprints"))

		structure, err := DetectStructure(context.Background(), repo)
		if err != nil {
			t.Fatalf("detect: %v", err)
		}
		if structure != Consolidated {
			t.Fatalf("expected Consolidated, got %v", structure)
		}
	})

	t.Run("legacy detected when only legacy present", func(t *testing.T) {
		repo := t.TempDir()
		requireDir(t, filepath.Join(repo, "backlog"))
		requireDir(t, filepath.Join(repo, "sprints"))

		structure, err := DetectStructure(context.Background(), repo)
		if err != nil {
			t.Fatalf("detect: %v", err)
		}
		if structure != Legacy {
			t.Fatalf("expected Legacy, got %v", structure)
		}
	})

	t.Run("default consolidated when empty repo", func(t *testing.T) {
		repo := t.TempDir()
		structure, err := DetectStructure(context.Background(), repo)
		if err != nil {
			t.Fatalf("detect: %v", err)
		}
		if structure != Consolidated {
			t.Fatalf("expected Consolidated default, got %v", structure)
		}
	})

	t.Run("accepts partial consolidated structure", func(t *testing.T) {
		repo := t.TempDir()
		requireDir(t, filepath.Join(repo, "tasks", "backlog"))

		structure, err := DetectStructure(context.Background(), repo)
		if err != nil {
			t.Fatalf("detect: %v", err)
		}
		if structure != Consolidated {
			t.Fatalf("expected Consolidated, got %v", structure)
		}
	})

	t.Run("context cancellation honored", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := DetectStructure(ctx, t.TempDir()); err == nil {
			t.Fatalf("expected context cancellation")
		}
	})
}

func TestResolvers(t *testing.T) {
	repo := "/repo"
	expect := func(got, want string) {
		if filepath.ToSlash(got) != want {
			t.Fatalf("unexpected path: %s want %s", got, want)
		}
	}
	expect(ResolveBacklogPath(repo, Consolidated), "/repo/tasks/backlog")
	expect(ResolveBacklogPath(repo, Legacy), "/repo/backlog")
	expect(ResolveSprintsPath(repo, Consolidated), "/repo/tasks/sprints")
	expect(ResolveSprintsPath(repo, Legacy), "/repo/sprints")
	expect(ResolveSprintPath(repo, "Sprint-01", Consolidated), "/repo/tasks/sprints/Sprint-01")
}

func requireDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
