# Logging Package (`pkg/logging`)

**Purpose**: Shared structured logging helpers using `log/slog`, providing consistent log formatting across layers.

**Responsibilities**:
- Initialize structured logger with context support
- Provide log level configuration
- Format logs consistently (JSON in production, human-readable in dev)
- Support trace IDs for request correlation

**Allowed Dependencies**:
- `log/slog` (standard library)
- Standard library

**Forbidden Dependencies**:
- Domain packages (`internal/core`, `internal/services`)
- Adapter packages (`cmd/`, `infra/`, `ui/`)

**Example**: `pkg/logging/logger.go` exports `NewLogger()` that returns a configured `*slog.Logger` for use across the codebase.

