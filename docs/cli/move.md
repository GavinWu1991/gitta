# `gitta story move`

Move story file to different directory atomically.

## Usage

```bash
gitta story move <story-id> [flags]
```

## Flags

- `--to` (string, **required**): Target directory path
- `--force` (bool, optional): Overwrite existing file at destination (default: false)
- `--json` (bool, optional): Output JSON instead of human-readable format

## Arguments

- `<story-id>` (required): Story ID to move (e.g., "US-001")

## Behavior

1. Find story file by ID
2. Read story file
3. Create target directory if needed
4. Write to temp file at destination
5. Rename temp file to final name (atomic)
6. Remove source file
7. Output success message

## Output Format

**Human-readable** (default):
```
Moved US-001: /old/path/US-001.md -> /new/path/US-001.md
```

**JSON** (`--json`):
```json
{
  "id": "US-001",
  "from": "/old/path/US-001.md",
  "to": "/new/path/US-001.md"
}
```

## Exit Codes

- `0`: Success
- `1`: Error (story not found, file read/write error, move failed)
- `2`: Validation error (invalid target path, file exists without --force)

## Error Messages

- `"story not found: {id}"`: No story file found with given ID
- `"target file exists: {path} (use --force to overwrite)"`: File already exists at destination
- `"invalid target path: {path}"`: Path validation failed (path traversal, etc.)
- `"failed to move story: {error}"`: Move operation failed (original preserved)

## Examples

### Move to Sprint Directory

```bash
gitta story move US-001 --to sprints/2025-01/
```

### Move to Backlog

```bash
gitta story move US-001 --to backlog/
```

### Force Overwrite

```bash
gitta story move US-001 --to sprints/2025-01/ --force
```

### JSON Output

```bash
gitta story move US-001 --to sprints/2025-01/ --json
```

## Notes

- Move operations are atomic (temp file + rename pattern)
- Original file is preserved if move fails
- Target directory is created automatically if it doesn't exist
- Path traversal attempts (e.g., `../`) are rejected for security
- Story lookup searches: Current Sprint -> other Sprints -> backlog
