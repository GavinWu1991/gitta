# `gitta story create`

Create a new story file from a template, generate unique ID, and open editor.

## Usage

```bash
gitta story create [flags]
```

## Flags

- `--title` (string, **required**): Story title/name
- `--prefix` (string, optional, default: "US"): ID prefix (e.g., "US", "BUG", "TS")
- `--template` (string, optional): Path to custom template file (default: built-in template)
- `--editor` (string, optional): Editor command to use (default: `$EDITOR` env var, fallback: "vi")
- `--status` (string, optional): Initial status (default: "todo")
  - Valid values: `todo`, `doing`, `review`, `done`
- `--priority` (string, optional): Initial priority (default: "medium")
  - Valid values: `low`, `medium`, `high`, `critical`
- `--assignee` (string, optional): Initial assignee
- `--tag` ([]string, optional): Initial tags (can be specified multiple times)
- `--json` (bool, optional): Output JSON instead of human-readable format

## Arguments

None (all input via flags)

## Behavior

1. Generate unique ID using prefix (e.g., "US-1", "US-2")
2. Load template (built-in or custom)
3. Create story file with frontmatter populated
4. Launch editor with story file (if `$EDITOR` is set or `--editor` is specified)
5. Wait for editor to exit
6. Validate story file after edit
7. Output success message or error

## Output Format

**Human-readable** (default):
```
Created story US-1: My Story Title
File: /path/to/backlog/US-1.md
Opened in editor: vim
```

**JSON** (`--json`):
```json
{
  "id": "US-1",
  "file": "/path/to/backlog/US-1.md",
  "title": "My Story Title",
  "editor": "vim"
}
```

## Exit Codes

- `0`: Success (story created, editor launched if specified)
- `1`: Error (ID generation failed, template error, editor launch failed)
- `2`: Validation error (invalid flag values, story validation failed after edit)

## Error Messages

- `"failed to generate ID: {error}"`: ID generation failed (lock timeout, counter corruption)
- `"template file not found: {path}"`: Template file missing
- `"failed to create story file: {error}"`: File write error
- `"failed to launch editor: {error}"`: Editor command failed
- `"story validation failed: {errors}"`: Story file invalid after edit
- `"--title is required"`: Title flag not provided

## Examples

### Basic Creation

```bash
# Create story with default prefix
gitta story create --title "Implement user authentication"
```

### With Custom Prefix

```bash
# Create bug with custom prefix
gitta story create --title "Fix login bug" --prefix BG --priority high
```

### With Initial Metadata

```bash
# Create story with initial status, assignee, and tags
gitta story create \
  --title "Add dark mode" \
  --status doing \
  --assignee alice \
  --tag frontend \
  --tag ui
```

### Custom Template

```bash
# Use custom template
gitta story create --title "New feature" --template ./templates/custom.md
```

### JSON Output

```bash
# Get machine-readable output
gitta story create --title "Test Story" --json
```

## Notes

- Story files are created in the `backlog/` directory by default
- If `backlog/` doesn't exist, stories are created in the current directory
- Editor integration respects the `$EDITOR` environment variable
- If editor fails to launch, the story file is still created (user can edit manually)
- ID generation is thread-safe and handles concurrent creation
