# Command: `gitta version`

## Description

Reports build metadata for audit and debugging purposes. Outputs semantic version, commit SHA, build date, and Go runtime version.

## Usage

```bash
gitta version [--json]
```

## Flags

- `--json` (bool, default: `false`): Emit machine-readable JSON output instead of human-readable text.

## Global Flags

- `--log-level` (string, default: `info`): Set log level (debug|info|warn|error). Not used by this command but available for consistency.

## Output Formats

### Human-Readable (default)

```
gitta v0.1.0
commit: abc123def456
build date: 2024-01-15T10:30:00Z
go: go1.21.5
```

If commit is unavailable, shows `<unknown>`.

### JSON (`--json` flag)

```json
{
  "version": "0.1.0",
  "commit": "abc123def456",
  "buildDate": "2024-01-15T10:30:00Z",
  "goVersion": "go1.21.5"
}
```

**JSON Schema**:
- `version` (string): Semantic version (e.g., "0.1.0")
- `commit` (string): Git commit SHA (empty string if unavailable)
- `buildDate` (string): ISO 8601 timestamp of build time
- `goVersion` (string): Go runtime version (e.g., "go1.21.5")

## Exit Codes

- `0`: Success - version information emitted
- `1`: General error (invalid flag, JSON marshaling failure)
- `3`: Version metadata missing or corrupted (reserved per contract)

## Environment Variables

- `GITTA_OUTPUT=json`: Equivalent to `--json` flag (if supported in future)

## Build-Time Configuration

Version information is injected via ldflags during build:

```bash
go build -ldflags "-X github.com/gavin/gitta/cmd/gitta.buildVersion=0.1.0 \
  -X github.com/gavin/gitta/cmd/gitta.buildCommit=$(git rev-parse HEAD) \
  -X github.com/gavin/gitta/cmd/gitta.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o gitta ./cmd/gitta
```

## Migration Notes

- **v0.1.0 → v0.2.0**: No breaking changes expected. Adding fields to JSON output requires dual-mode period per constitution.
- **Future**: If removing fields, maintain backward compatibility for one MAJOR version cycle.

## Examples

```bash
# Human-readable output
$ gitta version
gitta v0.1.0
commit: abc123def456
build date: 2024-01-15T10:30:00Z
go: go1.21.5

# JSON output
$ gitta version --json
{
  "version": "0.1.0",
  "commit": "abc123def456",
  "buildDate": "2024-01-15T10:30:00Z",
  "goVersion": "go1.21.5"
}
```

## Constraints

- No network calls; all data sourced from build-time ldflags or runtime detection
- Adding/removing JSON fields requires migration note + dual-mode period per constitution
- Exit code 3 reserved for metadata corruption (not currently used)

