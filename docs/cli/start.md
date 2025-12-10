# Command: `gitta start`

## Description

Start working on a task by creating (if needed) and checking out its feature branch, optionally updating the task's `assignee` field.

## Usage

```bash
gitta start <task-id|file-path> [--assignee <name>]
```

## Arguments

- `<task-id|file-path>` (required): Task identifier (e.g., `US-001`) or path to a task Markdown file.

## Flags

- `--assignee <name>`: Explicitly set the assignee in the task frontmatter. If omitted, attempts to use Git `user.name`; skips update if unavailable.

## Behavior

1. Locate the task (ID search through Sprint/backlog or direct file path).
2. Construct branch name `<prefix><task-id>` (default prefix `feat/` from config).
3. Create branch if missing; checkout branch (requires clean working tree).
4. Optionally update `assignee` frontmatter (atomic write, preserves content).

## Examples

```bash
# Start by task ID
gitta start US-001

# Start by file path
gitta start sprints/Sprint-01/US-001.md

# Start and set assignee explicitly
gitta start US-001 --assignee alice
```

## Exit Codes

- `0`: Success (branch checked out; assignee updated if applicable)
- `1`: Error (task not found, Git checkout failed, validation error)

## Notes

- Requires a Git repository and at least one commit (to create new branches).
- Requires a clean working tree unless a force option is added in the future.
