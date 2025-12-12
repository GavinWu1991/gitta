# `gitta list`

Display current Sprint tasks in a formatted table. Use `--all` to include backlog tasks.

## Usage

```bash
gitta list [flags]
```

## Flags

- `--all` (bool, default `false`): Include backlog tasks in addition to Sprint tasks.
- `--status` ([]string, optional): Filter by status (can specify multiple: `--status todo --status doing`)
  - Valid values: `todo`, `doing`, `review`, `done`
- `--priority` ([]string, optional): Filter by priority
  - Valid values: `low`, `medium`, `high`, `critical`
- `--assignee` ([]string, optional): Filter by assignee
- `--tag` ([]string, optional): Filter by tags (story must have any tag)
- `--sort` (string, optional): Sort field (id, title, status, priority, created_at) (default: "id")
- `--json` (bool, optional): Output JSON instead of formatted table

## Behavior

- Default: lists tasks from the current Sprint (lexicographically latest Sprint directory under `sprints/`).
- `--all`: lists Sprint + backlog tasks together, with sections for `Sprint` and `Backlog`.
- Filter flags: When any filter flag is specified, lists all stories (Sprint + backlog) and applies filters.
  - Multiple values within a field use OR logic (e.g., `--status todo --status doing` matches stories with status "todo" OR "doing")
  - Multiple filter fields use AND logic (e.g., `--status todo --priority high` matches stories with status "todo" AND priority "high")
- Status is derived from Git branch state when not explicitly set in frontmatter.
- Empty states print friendly messages (`No Sprint tasks found.` or `No tasks found.`).

## Output

- Formatted table with columns: ID, Title, Status, Assignee, Priority.
- Status colors: Todo (gray), Doing (yellow), Review (blue), Done (green).
- Rounded borders and aligned columns; long fields are truncated with ellipsis.
- Sections for Sprint/Backlog when `--all` is used.

## Examples

### Current Sprint only

```bash
$ gitta list
┌──────────┬──────────────────────────────────────┬────────────┬────────────────────┬──────────┐
│ ID       │ Title                                │ Status     │ Assignee           │ Priority │
├──────────┼──────────────────────────────────────┼────────────┼────────────────────┼──────────┤
│ US-001   │ Sprint task                          │ Doing      │ alice              │ High     │
└──────────┴──────────────────────────────────────┴────────────┴────────────────────┴──────────┘
```

### Sprint + Backlog

```bash
$ gitta list --all
Sprint
┌──────────┬──────────────────────────────────────┬────────────┬────────────────────┬──────────┐
│ ID       │ Title                                │ Status     │ Assignee           │ Priority │
├──────────┼──────────────────────────────────────┼────────────┼────────────────────┼──────────┤
│ US-001   │ Sprint task                          │ Doing      │ alice              │ High     │
└──────────┴──────────────────────────────────────┴────────────┴────────────────────┴──────────┘

Backlog
┌──────────┬──────────────────────────────────────┬────────────┬────────────────────┬──────────┐
│ ID       │ Title                                │ Status     │ Assignee           │ Priority │
├──────────┼──────────────────────────────────────┼────────────┼────────────────────┼──────────┤
│ BL-001   │ Backlog task                         │ Todo       │                    │ Medium   │
└──────────┴──────────────────────────────────────┴────────────┴────────────────────┴──────────┘
```

### Empty states

- No Sprint tasks: `No Sprint tasks found.`
- No tasks with `--all`: `No tasks found.`

## Exit Codes

- `0`: Success
- `1`: Error (directory scan failed, parse errors)
- `2`: Validation error (invalid filter values)

## Error Messages

- `"failed to scan directory: {error}"`: Directory scan error
- `"failed to parse story {file}: {error}"`: Story file parse error (non-fatal, continues)
- `"invalid filter value: {field}={value}"`: Invalid enum value in filter
- `"invalid status: {value} (valid: todo, doing, review, done)"`: Invalid status value
- `"invalid priority: {value} (valid: low, medium, high, critical)"`: Invalid priority value
- `"invalid assignee: {value} (must be alphanumeric with hyphens/underscores)"`: Invalid assignee format
- `"invalid tag: {value} (must be alphanumeric with hyphens/underscores)"`: Invalid tag format

## Notes

- Story files must be Markdown with YAML frontmatter.
- Sprint detection uses lexicographic order of `sprints/` subdirectories.
- Backlog directory is optional; missing backlog is treated as empty.
- Colors require a color-capable terminal; output degrades gracefully otherwise.
