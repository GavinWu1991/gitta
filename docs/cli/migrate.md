# `gitta migrate`

Migrate legacy gitta workspace (`backlog/`, `sprints/` in repo root) to consolidated `tasks/backlog/` and `tasks/sprints/`.

## Usage

```bash
gitta migrate [--force] [--dry-run] [--json]
```

## Flags

- `--force`: Overwrite existing `tasks/` targets (creates timestamped backups).
- `--dry-run`: Simulate migration without making changes.
- `--json`: Output JSON result.

## Behavior

1) Detect workspace structure. If already consolidated, exits successfully with message.  
2) If legacy structure, checks for conflicts in `tasks/backlog/` or `tasks/sprints/`.  
3) Without `--force`, conflicts abort with guidance. With `--force`, conflicting targets are backed up before overwrite.  
4) Moves `backlog/` → `tasks/backlog/` and `sprints/` → `tasks/sprints/` (preserves Git history via rename).  
5) Reports moved directories and backups.

## Examples

```bash
# Migrate legacy repo
gitta migrate

# Dry run
gitta migrate --dry-run

# Force overwrite with backups
gitta migrate --force

# JSON output
gitta migrate --json
```

## Exit Codes

- `0`: Success (migrated or already consolidated)
- `1`: Error (conflict without --force, not a git repo, missing legacy dirs)
- `2`: Flag/argument parsing errors

## Notes

- Requires a Git repository (`.git` present).
- Legacy structure must contain both `backlog/` and `sprints/`.
- Backups are created only when `--force` and conflicts are present.  
