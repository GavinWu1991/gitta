# Git Infrastructure Adapter (`infra/git`)

**Purpose**: Adapter for Git operations using go-git library, implementing repository interfaces from `internal/core`.

**Responsibilities**:
- Implement `core.StoryRepository` or similar Git-backed interfaces
- Wrap go-git API calls
- Handle Git-specific errors and edge cases
- Provide Git metadata (branches, commits) to domain services

**Allowed Dependencies**:
- `internal/core` (interfaces to implement)
- `pkg/logging` (for structured logging)
- go-git library, standard library

**Forbidden Dependencies**:
- `internal/services` (services depend on infra, not vice versa)
- `cmd/`, `ui/` (adapter layers)

**Example**: `infra/git/repository.go` implements `core.StoryRepository` using go-git's `Repository` type.

