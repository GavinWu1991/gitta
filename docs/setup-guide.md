# Gitta Setup Guide

This guide walks you through using gitta in an existing Git repository: downloading prebuilt binaries, verifying checksums, initializing directory structure, and running the release pipeline with quality gates.

## 1. Get & Install

### 1.1 Prebuilt binaries (recommended)

1) Open the GitHub Releases page and pick a version.  
2) Download the archive for your platform:  
   - macOS: `gitta-<version>-darwin-amd64.tar.gz` (Intel) or `darwin-arm64.tar.gz` (Apple Silicon)  
   - Linux: `gitta-<version>-linux-amd64.tar.gz` or `linux-arm64.tar.gz`  
   - Windows: `gitta-<version>-windows-amd64.zip` or `windows-arm64.zip`  
3) Verify checksum (recommended):  
   ```bash
   shasum -a 256 gitta-<version>-<platform>-<arch>.tar.gz
   # Compare with checksums.txt on the release page
   ```
4) Extract and add to PATH:  
   ```bash
   tar -xzf gitta-<version>-darwin-amd64.tar.gz   # macOS/Linux
   unzip gitta-<version>-windows-amd64.zip        # Windows
   sudo mv gitta /usr/local/bin/                  # optional
   gitta --help
   ```

### 1.2 Build from source

```bash
git clone https://github.com/GavinWu1991/gitta.git
cd gitta
go mod tidy
go build -o gitta ./cmd/gitta
./gitta --help
```

### 1.3 One-line install + init

```bash
curl -sSf https://raw.githubusercontent.com/GavinWu1991/gitta/main/scripts/remote-init.sh | bash
# Force re-init or custom sprint name:
curl -sSf https://raw.githubusercontent.com/GavinWu1991/gitta/main/scripts/remote-init.sh | bash -s -- --force --example-sprint Sprint-02
```

## 2. Initialize an existing project (init script)

> Works in any Git repo, creates gitta directories and example tasks.

```bash
# Assume the script is available
./scripts/init.sh

# If already initialized and need to rebuild
./scripts/init.sh --force                    # backup existing dirs then rebuild
./scripts/init.sh --example-sprint Sprint-02 # custom sprint name
```

Script behavior:
- Creates `sprints/<Sprint-01>/` and `backlog/` directories
- Writes example tasks (US-001, US-002) following gitta format
- If directories already exist:
  - Default: skip and warn
  - With `--force`: backup then recreate
- Must be run at Git repo root (checks `.git/`)

After completion, run:
```bash
gitta list       # view example tasks
gitta list --all # view Sprint + backlog
```

## 3. Release & rollback (maintainers)

### 3.1 Create a release
```bash
# Ensure clean workspace and passing tests
make verify

# Tag (semantic version)
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions will:
# - Run quality gate (make verify)
# - Invoke GoReleaser to build 6 platform/arch artifacts
# - Generate checksums.txt and attach to Release
```

### 3.2 Quick rollback
```bash
# Delete remote tag and release (delete Release in GitHub UI)
git push origin :refs/tags/v1.0.0
# To re-release, fix issues and push the same or higher tag
```

## 4. Quality gates & troubleshooting

- GitHub Actions `quality` job runs `make verify` (gofmt/goimports/vet/staticcheck/govulncheck/tests/architecture checks).
- On failure: release job is skipped; check Actions logs for the failing step.
- Reproduce locally: `make verify`, or `make release-snapshot` for a local GoReleaser dry run (no publish).

## 5. FAQ

- **Not a Git repo**: run `git init` at project root before the script.  
- **Permission denied**: ensure write permission; rerun with appropriate permissions if needed.  
- **Checksum mismatch**: re-download and ensure platform/arch match checksums.txt.  
- **Platform not supported**: build from source per README steps.  

## 6. Release notes & changelog (reference)

- Release notes are derived from Git commits (GoReleaser changelog) and attached automatically.
- To add manual notes, edit the Release page on GitHub.
- Suggested Release notes template (paste into Release body):  
  ```
  ## Summary
  - Features: ...
  - Fixes: ...
  - Docs/DevEx: ...
  - Breaking changes: none/yes (describe migration)

  ## Verification
  - Artifacts: macOS amd64/arm64, Linux amd64/arm64, Windows amd64/arm64
  - Checksums: checksums.txt attached
  - Quality gate: make verify passed
  ```

## 7. Directory layout (after init)

```
<repo>/
├── sprints/
│   └── Sprint-01/
│       └── US-001.md   # example sprint task
└── backlog/
    └── US-002.md       # example backlog task
```

After these steps, you can use the usual gitta commands:
```bash
gitta list         # current Sprint
gitta list --all   # Sprint + backlog
gitta start <id>   # create/switch branch for a task
```
