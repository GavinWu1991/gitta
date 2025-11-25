# Gitta - Git Task Assistant

A lightweight task management tool that stores tasks as Markdown files in your Git repository. Gitta uses branch state to track task progress automatically, following the "Git is the Database" philosophy.

## Features

- **Zero Infrastructure**: No servers, databases, or external services required
- **Git-Native**: Tasks stored as Markdown files in your repository
- **Branch-Aware**: Automatically tracks task status based on Git branch state
- **CLI & TUI**: Command-line interface with future interactive TUI support
- **Offline-First**: Works completely offline after initial setup

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, for development)

### Installation

```bash
git clone https://github.com/gavin/gitta.git
cd gitta
go mod tidy
make verify  # Run all checks
```

### Build

```bash
go build -o gitta ./cmd/gitta
./gitta --help
./gitta version
```

See [docs/quickstart.md](docs/quickstart.md) for detailed setup instructions.

## Architecture

Gitta follows **Hexagonal Architecture** principles:

- **Domain Layer** (`internal/core`, `internal/services`): Business logic
- **Adapter Layers** (`cmd/`, `infra/`, `ui/`): External interfaces
- **Shared Packages** (`pkg/`): Reusable utilities

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

## Development

### Project Structure

```
cmd/gitta/          # CLI commands (Cobra)
internal/           # Domain logic
  core/             # Interfaces
  services/         # Implementations
infra/              # Infrastructure adapters (Git, filesystem)
pkg/                # Shared utilities
tools/              # Development tools
docs/               # Documentation
```

### Running Tests

```bash
go test ./...
make verify  # Includes tests + linting
```

### Adding New Commands

1. Create command file: `cmd/gitta/<command>.go`
2. Register in `cmd/gitta/root.go`
3. Implement service in `internal/services/`
4. Update `docs/cli/<command>.md`

See `cmd/README.md` for detailed command registration guidelines.

## Documentation

- [Architecture Guide](docs/architecture.md)
- [CLI Reference](docs/cli/)
- [Quickstart Guide](specs/001-go-init-skeleton/quickstart.md)

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]

