# Configuration Package (`pkg/config`)

**Purpose**: Shared configuration management using Viper, accessible to adapters and services.

**Responsibilities**:
- Load configuration from files, environment variables, flags
- Provide typed configuration accessors
- Validate configuration values
- Support multiple config sources (YAML, env vars, defaults)

**Allowed Dependencies**:
- Viper library
- Standard library

**Forbidden Dependencies**:
- Domain packages (`internal/core`, `internal/services`)
- Adapter packages (`cmd/`, `infra/`, `ui/`)

**Example**: `pkg/config/config.go` exports `Load()` function that returns a `Config` struct usable by any layer.

