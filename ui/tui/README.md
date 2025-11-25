# Terminal UI Adapter (`ui/tui`)

**Purpose**: Bubble Tea-based terminal user interface adapter for interactive workflows (future feature).

**Responsibilities**:
- Render interactive TUI views (kanban board, story list)
- Handle keyboard input and navigation
- Delegate data fetching to domain services via interfaces
- Format output for terminal display

**Allowed Dependencies**:
- `internal/core` (domain interfaces)
- `internal/services` (service implementations)
- `pkg/config`, `pkg/logging` (shared utilities)
- Bubble Tea, Lipgloss libraries

**Forbidden Dependencies**:
- `infra/` (use via service interfaces)
- `cmd/` (separate adapter layer)

**Example**: `ui/tui/board.go` renders a kanban board by calling `services.TaskService.List()` and formatting results with Lipgloss.

