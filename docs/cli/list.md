# `gitta list`

Display current Sprint tasks in a formatted table. Use `--all` to include backlog tasks.

## Usage

```bash
gitta list [--all]
```

## Flags

- `--all` (bool, default `false`): Include backlog tasks in addition to Sprint tasks.

## Behavior

- Default: lists tasks from the current Sprint (lexicographically latest Sprint directory under `sprints/`).
- `--all`: lists Sprint + backlog tasks together, with sections for `Sprint` and `Backlog`.
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

## Errors

- Not a git repository: `failed to derive task status: not a Git repository`
- Missing Sprint directory: `failed to find current Sprint: no Sprint directories found in sprints/`
- Permission or read errors bubble with context.

## Notes

- Story files must be Markdown with YAML frontmatter.
- Sprint detection uses lexicographic order of `sprints/` subdirectories.
- Backlog directory is optional; missing backlog is treated as empty.
- Colors require a color-capable terminal; output degrades gracefully otherwise.
