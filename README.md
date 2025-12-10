# Gitta - Git Task Assistant

Lightweight, Git-native task management that treats your repository as the source of truth. Track Sprint and backlog stories as Markdown, and let branch state reflect progress automatically.

[English](README.md) | [中文](README.zh-CN.md)

[![CI](https://github.com/gavin/gitta/actions/workflows/ci.yml/badge.svg)](https://github.com/gavin/gitta/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-brightgreen.svg)](go.mod)

---

## Table of Contents

- [Gitta - Git Task Assistant](#gitta---git-task-assistant)
  - [Table of Contents](#table-of-contents)
  - [What is Gitta?](#what-is-gitta)
    - [Features](#features)
  - [Quick Start](#quick-start)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Build](#build)
    - [First Commands](#first-commands)
  - [Available Commands](#available-commands)
    - [Quick Examples](#quick-examples)
  - [Common Workflows](#common-workflows)
    - [Getting Started (Install → List → Start → Verify)](#getting-started-install--list--start--verify)
    - [Daily Flow (Pull → List → Start/Continue → Review)](#daily-flow-pull--list--startcontinue--review)
    - [Sprint Planning (Sprint vs Backlog)](#sprint-planning-sprint-vs-backlog)
  - [Architecture](#architecture)
  - [Development](#development)
    - [Project Structure](#project-structure)
    - [Tests](#tests)
    - [Adding New Commands](#adding-new-commands)
  - [Contributing](#contributing)
  - [Documentation](#documentation)
  - [Support](#support)
  - [License](#license)

---

## What is Gitta?

Gitta is a Git Task Assistant that stores tasks as Markdown files inside your repo. It derives status from Git branches, so your Sprint and backlog stay in sync with your actual work. No servers, no extra services—just Git.

### Features

- **Zero Infrastructure**: Nothing to provision; works in any Git repo.
- **Git-Native**: Tasks live as Markdown with YAML frontmatter.
- **Branch-Aware**: Branch state drives task status automatically.
- **CLI-First**: Fast command-line workflow; future TUI planned.
- **Offline-First**: Works entirely offline after setup.

---

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, for development)

### Installation

```bash
# Clone the repository
git clone https://github.com/gavin/gitta.git
cd gitta

# Install dependencies
go mod tidy

# Verify installation
make verify  # Run all checks
```

### Build

```bash
# Build the binary
go build -o gitta ./cmd/gitta

# Verify it works
./gitta --help
./gitta version
```

### First Commands

```bash
# List current Sprint tasks
gitta list

# List all tasks (Sprint + Backlog)
gitta list --all

# Start working on a task
gitta start US-001

# Check version
gitta version
```

---

## Available Commands

| Command | Description | Basic Usage | Docs |
|---------|-------------|-------------|------|
| `gitta list` | Show current Sprint tasks; `--all` includes backlog | `gitta list [--all]` | [docs/cli/list.md](docs/cli/list.md) |
| `gitta start` | Create/check out feature branch for a task, optionally set assignee | `gitta start <task-id|file-path> [--assignee <name>]` | [docs/cli/start.md](docs/cli/start.md) |
| `gitta version` | Report build metadata (semver, commit, build date, Go version) | `gitta version [--json]` | [docs/cli/version.md](docs/cli/version.md) |

### Quick Examples

```bash
# Sprint tasks only
gitta list

# Sprint + backlog
gitta list --all

# Start by task ID
gitta start US-001

# Start by file path
gitta start sprints/Sprint-01/US-001.md

# Check version
gitta version --json
```

---

## Common Workflows

### Getting Started (Install → List → Start → Verify)
1) Install and build (see Quick Start)  
2) View Sprint tasks: `gitta list`  
3) Start a task: `gitta start US-001`  
4) Verify branch and status: check Git branch and task frontmatter

### Daily Flow (Pull → List → Start/Continue → Review)
1) Update repo: `git pull`  
2) View Sprint: `gitta list`  
3) Start or continue a task: `gitta start <task-id>`  
4) Commit/push as you progress; use branches to reflect status

### Sprint Planning (Sprint vs Backlog)
1) List Sprint only: `gitta list`  
2) Review Sprint + backlog: `gitta list --all`  
3) Move tasks between Sprint/backlog by editing Markdown locations; rerun `gitta list --all` to verify

---

## Architecture

Hexagonal (ports-and-adapters) structure:
- **Domain**: `internal/core`, `internal/services`
- **Adapters**: `cmd/` (CLI), `infra/` (Git, filesystem), `ui/` (future TUI)
- **Shared**: `pkg/` utilities

See [docs/architecture.md](docs/architecture.md) for details.

---

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

### Tests

```bash
go test ./...
make verify  # Tests + linting
```

### Adding New Commands

1) Create command file: `cmd/gitta/<command>.go`  
2) Register in `cmd/gitta/root.go`  
3) Implement service in `internal/services/`  
4) Document in `docs/cli/<command>.md`

See `cmd/README.md` for command adapter guidance.

---

## Contributing

- Set up and verify: `go mod tidy && make verify`
- Follow hexagonal boundaries (no business logic in `cmd/`)
- Table-driven tests for non-trivial logic; add integration tests for CLI flows
- Open PRs referencing relevant specs/plans; describe which sections you touched
- Architecture reference: [docs/architecture.md](docs/architecture.md)
- Command docs reference: [cmd/README.md](cmd/README.md)

---

## Documentation

- [Architecture Guide](docs/architecture.md)
- [CLI Reference](docs/cli/)
- [Quickstart Guide](docs/quickstart.md)

---

## Support

- Issues: open on GitHub with repro steps and CLI output
- Troubleshooting: re-run `gitta list --all` to confirm task locations and statuses

---

## License

Licensed under the [MIT License](LICENSE).

