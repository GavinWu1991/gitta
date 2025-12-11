#!/usr/bin/env bash
# Remote one-shot installer + initializer for gitta
# Usage:
#   curl -sSf https://raw.githubusercontent.com/GavinWu1991/gitta/main/scripts/remote-init.sh | bash
#   curl -sSf https://raw.githubusercontent.com/GavinWu1991/gitta/main/scripts/remote-init.sh | bash -s -- --force --example-sprint Sprint-02

set -euo pipefail

log() { printf "[%s] %s\n" "gitta-remote" "$*"; }
fail() { log "ERROR: $*" >&2; exit 1; }

ensure_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "Missing required command: $1"
}

detect_os_arch() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch_raw="$(uname -m | tr '[:upper:]' '[:lower:]')"
  case "$arch_raw" in
    x86_64|amd64) arch="amd64";;
    arm64|aarch64) arch="arm64";;
    *) fail "Unsupported architecture: $arch_raw";;
  esac
  case "$os" in
    linux) os="linux";;
    darwin) os="darwin";;
    msys*|mingw*|cygwin*) os="windows";;
    *) fail "Unsupported OS: $os";;
  esac
  echo "$os" "$arch"
}

download() {
  local url="$1" dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$dest"
  else
    fail "Need curl or wget to download artifacts"
  fi
}

install_gitta() {
  local os="$1" arch="$2" tmpdir="$3"
  local base="https://github.com/GavinWu1991/gitta/releases/latest/download"
  local ext archive_url archive_path

  if [[ "$os" == "windows" ]]; then
    ext="zip"
  else
    ext="tar.gz"
  fi

  archive_url="$base/gitta-${os}-${arch}.${ext}"
  archive_path="$tmpdir/gitta.${ext}"

  log "Downloading gitta binary for ${os}/${arch}..."
  download "$archive_url" "$archive_path"

  local install_dir="/usr/local/bin"
  local fallback_dir="$(pwd)"

  if [[ "$os" == "windows" ]]; then
    ensure_cmd unzip
    unzip -o "$archive_path" -d "$tmpdir"
  else
    ensure_cmd tar
    tar -xzf "$archive_path" -C "$tmpdir"
  fi

  if install -m 0755 "$tmpdir/gitta" "$install_dir" 2>/dev/null; then
    log "Installed gitta to $install_dir"
    GITTA_BIN="$install_dir/gitta"
  else
    log "No permission to write $install_dir, installing to $fallback_dir"
    cp "$tmpdir/gitta" "$fallback_dir/"
    GITTA_BIN="$fallback_dir/gitta"
    log "Add to PATH if needed: export PATH=\"$fallback_dir:$PATH\""
  fi

  if ! command -v gitta >/dev/null 2>&1; then
    log "gitta not found in PATH; ensure install dir is on PATH"
  else
    gitta version || true
  fi
}

main() {
  local os arch
  read os arch < <(detect_os_arch)
  local tmpdir
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  install_gitta "$os" "$arch" "$tmpdir"

  if [[ ! -d .git ]]; then
    fail "Not a Git repository. Please run inside your project root (contains .git)."
  fi

  local gitta_bin="${GITTA_BIN:-$(command -v gitta || true)}"
  if [[ -z "$gitta_bin" ]]; then
    fail "gitta binary not found after install. Ensure it is on PATH."
  fi

  log "Running gitta init with args: $*"
  "$gitta_bin" init "$@"

  log "Done. You can now run: gitta list"
}

main "$@"
