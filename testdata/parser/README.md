# Parser Test Data

This directory contains test Markdown files used for parser testing.

## Test Files

- `valid-story.md`: Complete example with all fields
- `empty-file.md`: Empty file edge case
- `no-frontmatter.md`: Body-only file (no frontmatter)
- `malformed-yaml.md`: Invalid YAML frontmatter
- `missing-fields.md`: Missing optional fields

## Purpose

These files are used by:
- Unit tests (`tests/unit/story_parser_test.go`)
- Integration tests (`tests/integration/parser_test.go`)
- Golden file tests for round-trip validation

