# Domain Contracts (`internal/core`)

**Purpose**: Define domain interfaces and value objects that represent business concepts without implementation details.

**Responsibilities**:
- Declare interfaces for repositories, services, and domain entities
- Define value objects (Story, TaskStatus, etc.)
- Encode business rules as interface contracts

**Allowed Dependencies**:
- Standard library only
- Domain value objects

**Forbidden Dependencies**:
- Any adapter layer (`cmd/`, `infra/`, `ui/`)
- External libraries (except standard library)

**Example**: `internal/core/story.go` defines `StoryRepository` interface; implementations live in `internal/services`.

