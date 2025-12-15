# `gitta sprint` - Sprint Management Commands

Manage time-bounded work periods (sprints) for organizing tasks.

## Overview

Sprint management enables organizing tasks by time periods. Use `sprint start` to create and activate a new sprint, `sprint close` to close a sprint and rollover unfinished tasks, and `sprint burndown` to visualize sprint progress over time.

## Commands

### `gitta sprint start`

Creates a new sprint directory, sets up the current sprint link, and calculates the end date.
Alternatively, activates an existing sprint in Ready or Planning status.

**Usage:**
```bash
gitta sprint start [sprint-id] [flags]
```

**Arguments:**
- `sprint-id` (optional): 
  - If not provided: Creates a new sprint with auto-generated name
  - If provided: Activates existing sprint in Ready or Planning status (e.g., "24", "Sprint_24", "Sprint_24_Login")

**Flags:**
- `--duration, -d` (string): Sprint duration in days or weeks (only for new sprint creation)
  - Format: `<number><unit>` where unit is `w` (weeks) or `d` (days)
  - Examples: `--duration 2w`, `--duration 14d`
  - Default: `2w` (2 weeks)
- `--start-date` (string): Sprint start date (ISO 8601 format: YYYY-MM-DD) (only for new sprint creation)
  - Default: Today's date
- `--dry-run`: Show what would be done without making changes
- `--json`: Output result as JSON instead of human-readable format

**Examples:**
```bash
# Create new sprint with default settings (2 weeks, starts today)
gitta sprint start

# Create new sprint with custom name and duration
gitta sprint start --duration 3w

# Activate existing sprint (partial match)
gitta sprint start 24

# Activate existing sprint (full match)
gitta sprint start Sprint_24

# Dry run to see what would happen
gitta sprint start 24 --dry-run

# JSON output
gitta sprint start 24 --json
```

**When activating existing sprint:**
- Automatically archives any currently active sprint
- Transitions target sprint from Ready/Planning → Active
- Updates Current link to point to newly activated sprint

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

### `gitta sprint plan`

Creates a new sprint in Planning status for future work.

**Usage:**
```bash
gitta sprint plan <name> [flags]
```

**Arguments:**
- `name` (required): Sprint description/name (e.g., "Dashboard Redesign", "Payment Integration")

**Flags:**
- `--id` (string): Specify sprint ID manually (default: auto-generate next sequential number)
- `--json`: Output result as JSON instead of human-readable format

**Examples:**
```bash
# Create planning sprint with auto-generated ID
gitta sprint plan "Dashboard Redesign"

# Create planning sprint with specific ID
gitta sprint plan "Payment" --id Sprint_25

# JSON output
gitta sprint plan "Login Feature" --json
```

**Output Format:**

Human-readable (default):
```
Created planning sprint: @Sprint_27_Dashboard_Redesign
Sprint will appear in the Planning section.
```

JSON (`--json` flag):
```json
{
  "sprint": {
    "id": "Sprint_27",
    "name": "@Sprint_27_Dashboard_Redesign",
    "path": "sprints/@Sprint_27_Dashboard_Redesign",
    "status": "planning",
    "description": "Dashboard Redesign"
  }
}
```

**Exit Codes:**
- `0`: Success
- `1`: Error (sprint creation failed, duplicate ID, invalid name)
- `2`: Validation error (empty description, invalid characters in name)

**Error Messages:**
- `Error: sprint description cannot be empty` - Name argument is required
- `Error: sprint name/description cannot contain status prefix characters: !, +, @, ~` - Invalid characters in description
- `Error: sprint with ID "Sprint_25" already exists` - Duplicate sprint ID

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

### `gitta doctor`

Detects and repairs inconsistencies between visual indicators (folder name prefixes) and authoritative status files (`.gitta/status`).

**Usage:**
```bash
gitta doctor [flags]
```

**Flags:**
- `--fix`: Automatically repair detected inconsistencies (default: report only)
- `--sprint` (string): Check specific sprint only (default: check all sprints)
- `--json`: Output result as JSON instead of human-readable format

**Examples:**
```bash
# Check for inconsistencies (report only)
gitta doctor

# Check and automatically fix
gitta doctor --fix

# Check specific sprint
gitta doctor --sprint Sprint_24

# JSON output
gitta doctor --json
```

**Output Format:**

Human-readable (default, no issues):
```
Checking sprint status consistency...
✓ All sprints are consistent
✓ Current link points to active sprint
```

Human-readable (issues found, without --fix):
```
Checking sprint status consistency...
✗ Found 2 inconsistencies:

1. Sprint "Sprint_24_Login"
   - Folder name: !Sprint_24_Login (active)
   - Status file: ready
   → Should rename to: +Sprint_24_Login

2. Sprint "Sprint_25_Payment"
   - Folder name: Sprint_25_Payment (no prefix)
   - Status file: planning
   → Should rename to: @Sprint_25_Payment

Run with --fix to repair these issues.
```

Human-readable (with --fix):
```
Checking sprint status consistency...
✗ Found 2 inconsistencies, repairing...

1. Sprint "Sprint_24_Login"
   ✓ Renamed to: +Sprint_24_Login

2. Sprint "Sprint_25_Payment"
   ✓ Renamed to: @Sprint_25_Payment

✓ All inconsistencies repaired
```

JSON (`--json` flag):
```json
{
  "status": "ok",
  "sprints_checked": 10,
  "inconsistencies": [],
  "current_link_valid": true
}
```

**Exit Codes:**
- `0`: Success (no inconsistencies found, or all repaired with --fix)
- `1`: Error (filesystem error, permission error)
- `2`: Inconsistencies found (when --fix not used)

**Error Messages:**
- `Error: failed to detect inconsistencies: ...` - Error during scan
- `Error: failed to repair inconsistencies: ...` - Error during repair

**Status:** ✅ Implemented

## Sprint Status Management

Sprints use visual status indicators (folder name prefixes) that automatically sort in file managers:

- **`!` Active** - Currently active sprint (appears at top)
- **`+` Ready** - Prepared sprint ready to activate (appears after Active)
- **`@` Planning** - Future sprint in planning (appears in middle)
- **`~` Archived** - Completed sprint (appears at bottom)

Each sprint maintains a `.gitta/status` file containing the authoritative status. The `doctor` command ensures consistency between folder names and status files.

## Current Sprint Link

The current sprint link (`sprints/Current`) provides fast lookup of the active sprint:

- **Unix-like systems**: Symbolic link (`os.Symlink`)
- **Windows**: Junction link (`syscall.CreateSymbolicLink` with `SYMBOLIC_LINK_FLAG_DIRECTORY`)
- **Fallback**: Text file (`.current-sprint.txt`) containing sprint directory path

The link is automatically updated when you run `gitta sprint start`.

## Sprint Directory Structure

```
sprints/
├── Current -> !Sprint_24_Login          # Symlink/junction/text → Active sprint
├── !Sprint_24_Login                     # Active sprint (top)
│   ├── .gitta/
│   │   └── status                       # Contains: "active"
│   └── tasks/
├── +Sprint_25_Payment                   # Ready sprint
│   └── .gitta/
│       └── status                       # Contains: "ready"
├── @Sprint_26_Dashboard                 # Planning sprint
│   └── .gitta/
│       └── status                       # Contains: "planning"
└── ~Sprint_23_Onboarding               # Archived sprint (bottom)
    └── .gitta/
        └── status                       # Contains: "archived"
```

**Visual Organization:**
File managers automatically sort sprints by ASCII value of status prefix:
1. **Top**: `!` Active sprints (ASCII 33)
2. **After Active**: `+` Ready sprints (ASCII 43)
3. **Middle**: `@` Planning sprints (ASCII 64)
4. **Bottom**: `~` Archived sprints (ASCII 126)

## See Also

- [Quick Start Guide](../../specs/001-sprint-management/quickstart.md)
- [Sprint Service Contract](../../specs/001-sprint-management/contracts/sprint-service.md)
- [CLI Contract](../../specs/001-sprint-management/contracts/cli.md)
