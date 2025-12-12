# `gitta init`

Initialize the current Git repository with the gitta workspace structure and example tasks.

## Usage

```bash
gitta init [--force] [--example-sprint <name>]
```

## Flags

- `--force` (default: false): Backup existing gitta directories and recreate them.
- `--example-sprint <name>` (default: `Sprint-01`): Sprint folder name for example tasks; must be non-empty and filesystem-safe.
- `--help` / `-h`: Show help.

## Behavior

- Requires running inside a Git repository root (must contain `.git`).
- Creates `sprints/<name>/` and `backlog/` under the repo root.
- Copies bundled example tasks into the new folders:
  - `sprints/<name>/US-001.md`
  - `backlog/US-002.md`
- If `sprints/<name>/` or `backlog/` already exist and `--force` is not provided, exits with guidance to rerun with `--force`.
- With `--force`, existing gitta directories are moved to timestamped backups before recreation.
- Prints a summary of created directories/files and next steps (`gitta list`, `gitta list --all`).

## Exit Codes

- `0`: Initialization succeeded.
- `1`: Preconditions failed (not a Git repo, invalid sprint name, missing permissions) with a clear message.
- `2`: Flag or argument parsing errors.

## Examples

```bash
# Initialize with defaults
gitta init

# Initialize with a custom sprint folder
gitta init --example-sprint Sprint-02

# Recreate workspace, backing up existing folders first
gitta init --force

# Remote one-liner (install + init with flags)
curl -sSf https://raw.githubusercontent.com/GavinWu1991/gitta/main/scripts/remote-init.sh | bash -s -- --example-sprint Sprint-03
```
