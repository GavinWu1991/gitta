# `gitta sprint` - Sprint Management Commands

Manage time-bounded work periods (sprints) for organizing tasks.

## Overview

Sprint management enables organizing tasks by time periods. Use `sprint start` to create and activate a new sprint, `sprint close` to close a sprint and rollover unfinished tasks, and `sprint burndown` to visualize sprint progress over time.

## Commands

### `gitta sprint start`

Creates a new sprint directory, sets up the current sprint link, and calculates the end date.

**Usage:**
```bash
gitta sprint start [sprint-name] [flags]
```

**Arguments:**
- `sprint-name` (optional): Sprint name (e.g., "Sprint-01"). If not provided, auto-generates next sequential name.

**Flags:**
- `--duration, -d` (string): Sprint duration in days or weeks
  - Format: `<number><unit>` where unit is `w` (weeks) or `d` (days)
  - Examples: `--duration 2w`, `--duration 14d`
  - Default: `2w` (2 weeks)
- `--start-date` (string): Sprint start date (ISO 8601 format: YYYY-MM-DD)
  - Default: Today's date
- `--json`: Output result as JSON instead of human-readable format

**Examples:**
```bash
# Create sprint with default settings (2 weeks, starts today)
gitta sprint start

# Create sprint with custom name and duration
gitta sprint start Sprint-02 --duration 3w

# Create sprint starting on specific date
gitta sprint start Sprint-03 --start-date 2025-02-01 --duration 14d

# JSON output
gitta sprint start --json
```

**Output Format:**

Human-readable (default):
```
✓ Created sprint: Sprint-01
✓ Start date: 2025-01-27
✓ End date: 2025-02-10
✓ Duration: 2 weeks
✓ Current sprint link updated
```

JSON (`--json` flag):
```json
{
  "name": "Sprint-01",
  "start_date": "2025-01-27T00:00:00Z",
  "end_date": "2025-02-10T00:00:00Z",
  "duration": "2w",
  "directory_path": "sprints/Sprint-01",
  "created_at": "2025-01-27T12:00:00Z",
  "updated_at": "2025-01-27T12:00:00Z"
}
```

**Exit Codes:**
- `0`: Success
- `1`: Error (sprint creation failed, symlink creation failed, etc.)
- `2`: Validation error (invalid sprint name, invalid duration format, sprint already exists)

**Error Messages:**
- `Error: sprint "Sprint-01" already exists` - Sprint directory with that name already exists
- `Error: invalid sprint name format` - Sprint name doesn't match required pattern
- `Error: invalid duration format` - Duration string is malformed
- `Error: failed to create current sprint link` - Symlink/junction creation failed (with fallback info)

### `gitta sprint close`

Closes the current sprint, identifies unfinished tasks, and provides interactive selection for rollover.

**Usage:**
```bash
gitta sprint close [target-sprint] [flags]
```

**Arguments:**
- `target-sprint` (optional): Target sprint name for rollover. If not provided, prompts user to select or create.

**Flags:**
- `--target-sprint, -t` (string): Target sprint name for rollover (skips interactive selection)
- `--all`: Rollover all unfinished tasks without prompting (non-interactive mode)
- `--skip`: Skip rollover, just close the sprint (no task movement)
- `--json`: Output result as JSON instead of human-readable format

**Examples:**
```bash
# Interactive close with TUI selection
gitta sprint close

# Close and rollover to specific sprint (non-interactive)
gitta sprint close --target-sprint Sprint-02

# Close and rollover all unfinished tasks (non-interactive)
gitta sprint close --target-sprint Sprint-02 --all

# Close without rollover
gitta sprint close --skip
```

**Status:** ✅ Implemented

### `gitta sprint burndown`

Generates a burndown chart showing sprint progress over time, reconstructed from Git history.

**Usage:**
```bash
gitta sprint burndown [sprint-name] [flags]
```

**Arguments:**
- `sprint-name` (optional): Sprint name to analyze. If not provided, uses current sprint.

**Flags:**
- `--sprint, -s` (string): Sprint name to analyze (alternative to positional argument)
- `--format` (string): Output format
  - Values: `ascii` (default), `json`, `csv`
- `--points-only`: Show only story points (hide task count)
- `--tasks-only`: Show only task count (hide story points)
- `--json`: Output as JSON (same as `--format json`)

**Examples:**
```bash
# Burndown for current sprint (ASCII chart)
gitta sprint burndown

# Burndown for specific sprint
gitta sprint burndown Sprint-01

# JSON output
gitta sprint burndown --format json

# CSV output
gitta sprint burndown --format csv
```

**Status:** ✅ Implemented

## Current Sprint Link

The current sprint link (`sprints/.current-sprint`) provides fast lookup of the active sprint:

- **Unix-like systems**: Symbolic link (`os.Symlink`)
- **Windows**: Junction link (`syscall.CreateSymbolicLink` with `SYMBOLIC_LINK_FLAG_DIRECTORY`)
- **Fallback**: Text file (`.current-sprint.txt`) containing sprint directory path

The link is automatically updated when you run `gitta sprint start`.

## Sprint Directory Structure

```
sprints/
├── .current-sprint          # Symlink/junction/text → Sprint-02
├── Sprint-01/               # Closed sprint
│   ├── US-001.md
│   └── US-002.md
├── Sprint-02/               # Current sprint
│   ├── US-003.md
│   └── US-004.md
└── Sprint-03/               # Future sprint
```

## See Also

- [Quick Start Guide](../../specs/001-sprint-management/quickstart.md)
- [Sprint Service Contract](../../specs/001-sprint-management/contracts/sprint-service.md)
- [CLI Contract](../../specs/001-sprint-management/contracts/cli.md)
