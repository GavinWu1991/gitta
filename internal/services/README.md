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

