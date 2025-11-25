# Filesystem Infrastructure Adapter (`infra/filesystem`)

**Purpose**: Adapter for filesystem operations (reading/writing Markdown files, backlog access), implementing repository interfaces from `internal/core`.

**Responsibilities**:
- Implement `core.StoryRepository` or similar filesystem-backed interfaces
- Read/write Markdown files from `backlog/` and `sprints/` directories
- Handle filesystem errors and edge cases
- Provide file metadata (modification time, size) to domain services

**Allowed Dependencies**:
- `internal/core` (interfaces to implement)
- `pkg/logging` (for structured logging)
- Standard library (`os`, `path/filepath`, `io`)

**Forbidden Dependencies**:
- `internal/services` (services depend on infra, not vice versa)
- `cmd/`, `ui/` (adapter layers)

**Example**: `infra/filesystem/repository.go` implements `core.StoryRepository` using standard library file operations.

