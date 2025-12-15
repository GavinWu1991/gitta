#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(pwd)"
TEMPLATE_DIR="$(cd "${SCRIPT_DIR}/../testdata/init" && pwd)"

SPRINT_NAME="Sprint-01"
FORCE=false

usage() {
  cat <<'EOF'
Usage: scripts/init.sh [--force] [--example-sprint <name>] [--help]

Options:
  --force                Overwrite existing gitta directories (backs up first)
  --example-sprint NAME  Sprint name for example tasks (default: Sprint-01)
  --help                 Show this message
EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --force)
      FORCE=true
      shift
      ;;
    --example-sprint)
      if [[ $# -lt 2 ]]; then
        echo "❌ Missing value for --example-sprint" >&2
        exit 2
      fi
      SPRINT_NAME="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "❌ Unknown option: $1" >&2
      usage
      exit 2
      ;;
  esac
done

# Preconditions: must be in a Git repo
if [[ ! -d "${REPO_ROOT}/.git" ]]; then
  echo "❌ Error: Not a Git repository. Please run this script from a repo root." >&2
  exit 1
fi

TASKS_DIR="${REPO_ROOT}/tasks"
SPRINT_DIR="${TASKS_DIR}/sprints/${SPRINT_NAME}"
BACKLOG_DIR="${TASKS_DIR}/backlog"

backup_if_exists() {
  local target="$1"
  if [[ -d "$target" ]]; then
    local backup="${target}.backup-$(date +%s)"
    echo "⚠️  Backing up existing $(basename "$target") to $(basename "$backup")"
    mv "$target" "$backup"
  fi
}

# Handle existing directories
if [[ -d "$SPRINT_DIR" || -d "$BACKLOG_DIR" ]]; then
  if [[ "$FORCE" == "false" ]]; then
    echo "⚠️  Gitta appears initialized (found sprints/ or backlog/)."
    echo "    Use --force to backup and recreate."
    exit 0
  fi
  backup_if_exists "$SPRINT_DIR"
  backup_if_exists "$BACKLOG_DIR"
fi

echo "📁 Creating gitta directories..."
mkdir -p "$SPRINT_DIR" "$BACKLOG_DIR"

echo "📝 Writing example tasks..."
cp "${TEMPLATE_DIR}/US-001.md" "${SPRINT_DIR}/US-001.md"
cp "${TEMPLATE_DIR}/US-002.md" "${BACKLOG_DIR}/US-002.md"

cat <<EOF
✅ Gitta initialized successfully!

Created directories:
  - sprints/${SPRINT_NAME}/
  - backlog/

Created example tasks:
  - sprints/${SPRINT_NAME}/US-001.md
  - backlog/US-002.md

Next steps:
  1) gitta list
  2) gitta list --all
  3) Edit example tasks or create new ones
EOF
