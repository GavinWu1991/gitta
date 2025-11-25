# Gitta Architecture

## Overview

Gitta follows **Hexagonal Architecture** (Ports and Adapters) principles to maintain clear boundaries between business logic and infrastructure concerns. This document describes the architectural layers, dependency rules, and development guidelines.

## Architectural Layers

### 1. Domain Layer (`internal/core`)

**Purpose**: Define business contracts and value objects without implementation details.

**Contents**:
- Interfaces for repositories (`StoryRepository`, `TaskRepository`)
- Domain value objects (`Story`, `TaskStatus`, `Priority`, `Status`)
- Business rule contracts
- Parser interfaces (`StoryParser`) for reading/writing story files

**Rules**:
- ✅ Can depend on: Standard library only
- ❌ Cannot depend on: Any adapter layer (`cmd/`, `infra/`, `ui/`), external libraries

**Example**: `internal/core/story.go` defines `StoryRepository` interface.

### 2. Domain Services (`internal/services`)

**Purpose**: Implement domain interfaces, orchestrating business workflows.

**Contents**:
- Service implementations (`TaskService`, `StoryService`)
- Business logic and validation
- Coordination between repositories

**Rules**:
- ✅ Can depend on: `internal/core` (interfaces), `pkg/` (shared utilities), standard library
- ❌ Cannot depend on: Adapter layers (`cmd/`, `infra/`, `ui/`)

**Example**: `internal/services/task_service.go` implements `core.TaskService` and uses `core.StoryRepository` (injected via constructor).

### 3. Adapter Layers

#### CLI Adapters (`cmd/`)

**Purpose**: Command-line interface using Cobra.

**Rules**:
- ✅ Can depend on: `internal/core`, `internal/services`, `pkg/`, Cobra
- ❌ Cannot depend on: `infra/` (use via service interfaces), `ui/`

#### Infrastructure Adapters (`infra/`)

**Purpose**: External system integrations (Git, filesystem).

**Rules**:
- ✅ Can depend on: `internal/core` (interfaces), `pkg/logging`, external libraries (go-git, etc.)
- ❌ Cannot depend on: `internal/services`, `cmd/`, `ui/`

**Subdirectories**:
- `infra/git/`: Git operations via go-git
- `infra/filesystem/`: Filesystem operations for Markdown files
  - `markdown_parser.go`: Implements `core.StoryParser` interface for reading/writing Markdown story files with YAML frontmatter

#### UI Adapters (`ui/`)

**Purpose**: Terminal user interface using Bubble Tea (future).

**Rules**:
- ✅ Can depend on: `internal/core`, `internal/services`, `pkg/`, Bubble Tea
- ❌ Cannot depend on: `infra/` (use via service interfaces)

### 4. Shared Packages (`pkg/`)

**Purpose**: Reusable utilities accessible to all layers.

**Contents**:
- `pkg/config/`: Configuration management (Viper)
- `pkg/logging/`: Structured logging (slog)

**Rules**:
- ✅ Can depend on: External libraries (Viper, slog), standard library
- ❌ Cannot depend on: Domain packages (`internal/core`, `internal/services`), adapter packages

## Dependency Flow

```
┌─────────────┐
│    cmd/     │ ──┐
│    ui/      │ ──┼──► internal/services ──► internal/core
│   infra/    │ ──┘         │
└─────────────┘              │
                              ▼
                          pkg/config
                          pkg/logging
```

**Key Principle**: Dependencies flow inward. Adapters depend on services, services depend on core interfaces, but core never depends on adapters.

## Context Propagation

All services and adapters MUST accept `context.Context` as the first parameter:

```go
// ✅ Correct
func (s *TaskService) List(ctx context.Context) ([]Story, error)

// ❌ Wrong
func (s *TaskService) List() ([]Story, error)
```

**Rationale**: Enables cancellation, timeouts, and request tracing across layers.

## Concurrency Guidelines

1. **Goroutines**: Always receive `context.Context` and respect cancellation
2. **Channels**: Use `select` with timeout/heartbeat guards
3. **Worker Pools**: Use bounded semaphores for filesystem scanning
4. **Shared State**: Protect via immutable copies or `sync` primitives

**Example**: Scanner uses worker pool with `context.WithCancel`:

```go
func ScanStories(ctx context.Context, dir string) ([]Story, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Use bounded worker pool
    sem := make(chan struct{}, 10)
    // ... scan logic
}
```

## Architecture Guard

The `tools/check-architecture.go` script enforces import rules:

```bash
go run ./tools/check-architecture.go
```

**Violations**:
- Adapter importing domain implementation (should use interfaces)
- Domain importing adapter packages

**CI Integration**: Runs automatically via `make verify` and CI workflows.

## Adding New Features

### 1. Add Domain Interface (`internal/core`)

```go
package core

type MyRepository interface {
    Get(ctx context.Context, id string) (*MyEntity, error)
}
```

### 2. Implement Service (`internal/services`)

```go
package services

type MyService struct {
    repo core.MyRepository
}

func NewMyService(repo core.MyRepository) *MyService {
    return &MyService{repo: repo}
}
```

### 3. Create Infrastructure Adapter (`infra/`)

```go
package infra

type myRepository struct {
    // implementation
}

func (r *myRepository) Get(ctx context.Context, id string) (*core.MyEntity, error) {
    // implementation
}
```

### 4. Wire CLI Command (`cmd/gitta`)

```go
var myCmd = &cobra.Command{
    Use: "mycommand",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()
        repo := infra.NewMyRepository()
        svc := services.NewMyService(repo)
        return svc.DoSomething(ctx, args)
    },
}
```

## Testing Strategy

- **Unit Tests**: Test services with mock repositories (interfaces from `internal/core`)
- **Integration Tests**: Test adapters with real infrastructure (`tests/integration/`)
- **Architecture Tests**: Verify import rules (`tools/check-architecture_test.go`)

## References

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- Constitution: `.specify/memory/constitution.md`
- Command Registration: `cmd/README.md`

