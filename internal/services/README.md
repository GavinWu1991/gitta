# Domain Services (`internal/services`)

**Purpose**: Implement domain interfaces defined in `internal/core`, orchestrating business workflows.

**Responsibilities**:
- Implement `internal/core` interfaces
- Coordinate between repositories and domain logic
- Enforce business rules and validation
- Accept `context.Context` for cancellation/timeouts

**Allowed Dependencies**:
- `internal/core` (interfaces and value objects)
- `pkg/config`, `pkg/logging` (shared utilities)
- Standard library

**Forbidden Dependencies**:
- `cmd/`, `ui/` (adapter layers)
- `infra/` (use via interfaces from `internal/core`)

**Example**: `internal/services/task_service.go` implements `core.TaskService` and uses `core.StoryRepository` (injected via constructor).

## StatusEngine

The `StatusEngine` service derives story status from Git branch state automatically.

### Features

- **Automatic Status Derivation**: Determines status (Todo, Doing, Review, Done) based on branch state
- **Configurable Branch Patterns**: Supports custom branch naming conventions (default: "feat/<story-id>")
- **Explicit Status Override**: Frontmatter status takes precedence over derived status
- **Batch Processing**: Efficiently processes multiple stories with shared branch list
- **Edge Case Handling**: Gracefully handles empty repos, detached HEAD, missing remotes

### Usage

```go
// Create engine
engine := services.NewStatusEngine()
gitRepo := git.NewRepository()

// Get branch list
branchList, err := gitRepo.GetBranchList(ctx, repoPath)
if err != nil {
    return err
}

// Derive status for single story
status, err := engine.DeriveStatus(ctx, story, branchList, repoPath)
if err != nil {
    return err
}

// Derive status for multiple stories (batch)
statuses, err := engine.DeriveStatusBatch(ctx, stories, branchList, repoPath)
if err != nil {
    return err
}
```

### Configuration

Configuration is loaded via Viper with the following keys:

- `branch.prefix`: Branch naming prefix pattern (default: `"feat/"`)
- `branch.case_sensitive`: Case sensitivity for matching (default: `true`)
- `branch.target_branches`: Target branches for merge check (default: `["main", "master"]`)

### Status Derivation Priority

1. **Explicit Status**: If story has explicit status in Frontmatter → use it
2. **Branch Existence**: If no matching branch → Todo
3. **Merge Status**: If branch merged into main/master → Done
4. **Remote Branch**: If branch on remote → Review, else → Doing

