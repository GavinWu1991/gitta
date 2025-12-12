# `gitta story status`

Update story status atomically.

## Usage

```bash
gitta story status <story-id> [flags]
```

## Flags

- `--status` (string, **required**): New status value (todo, doing, review, done)
- `--json` (bool, optional): Output JSON instead of human-readable format

## Arguments

- `<story-id>` (required): Story ID to update (e.g., "US-001")

## Behavior

1. Find story file by ID (scan directories)
2. Read story file
3. Update status and updated_at timestamp
4. Write atomically (temp file + rename)
5. Output success message

## Output Format

**Human-readable** (default):
```
Updated US-001: status -> doing
```

**JSON** (`--json`):
```json
{
  "id": "US-001",
  "status": "doing",
  "updated_at": "2025-01-27T10:30:00Z"
}
```

## Exit Codes

- `0`: Success
- `1`: Error (story not found, file read/write error)
- `2`: Validation error (invalid status value)

## Error Messages

- `"story not found: {id}"`: No story file found with given ID
- `"failed to read story: {error}"`: File read error
- `"failed to update story: {error}"`: File write error (original preserved)
- `"invalid status: {value} (valid: todo, doing, review, done)"`: Invalid status enum value
- `"--status is required"`: Status flag not provided

## Examples

### Update Status to Doing

```bash
gitta story status US-001 --status doing
```

### Update Status to Done

```bash
gitta story status US-001 --status done
```

### JSON Output

```bash
gitta story status US-001 --status review --json
```

## Notes

- Status updates are atomic (temp file + rename pattern)
- Original file is preserved if update fails
- Updated_at timestamp is automatically set to current time
- Story lookup searches: Current Sprint -> other Sprints -> backlog
