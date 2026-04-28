#!/bin/sh
# install.sh — one-liner installer for worklog
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jquiaios/worklog/main/install.sh | sh
#
# Environment variables:
#   WORKLOG_VERSION     — pin a specific version, e.g. "0.2.0" (default: latest)
#   WORKLOG_INSTALL_DIR — override install directory  (default: /usr/local/bin or ~/.local/bin)

set -eu

REPO="jquiaios/worklog"
BINARY="worklog"
GITHUB_BASE="https://github.com/${REPO}"
API_BASE="https://api.github.com/repos/${REPO}"

# --- colour helpers (only when stdout is a tty) ---
if [ -t 1 ]; then
  BOLD=$(printf '\033[1m')
  GREEN=$(printf '\033[32m')
  YELLOW=$(printf '\033[33m')
  RED=$(printf '\033[31m')
  RESET=$(printf '\033[0m')
else
  BOLD="" GREEN="" YELLOW="" RED="" RESET=""
fi

info()    { printf '%s\n' "${BOLD}==> $*${RESET}"; }
success() { printf '%s\n' "${GREEN}==> $*${RESET}"; }
warn()    { printf '%s\n' "${YELLOW}warning: $*${RESET}" >&2; }
error()   { printf '%s\n' "${RED}error: $*${RESET}" >&2; exit 1; }

# --- HTTP: curl with wget fallback ---
fetch() {       # fetch <url> [dest]  — dest omitted → stdout
  url=$1; dest=${2:-}
  if command -v curl >/dev/null 2>&1; then
    if [ -n "$dest" ]; then
      curl -fsSL --retry 3 -o "$dest" "$url"
    else
      curl -fsSL --retry 3 "$url"
    fi
  elif command -v wget >/dev/null 2>&1; then
    if [ -n "$dest" ]; then
      wget -q --tries=3 -O "$dest" "$url"
    else
      wget -q --tries=3 -O- "$url"
    fi
  else
    error "curl or wget is required but neither was found. Install one and retry."
  fi
}

# --- OS / arch ---
detect_os() {
  case "$(uname -s)" in
    Linux)  echo linux ;;
    Darwin) echo darwin ;;
    *)      error "Unsupported OS: $(uname -s). Download manually from ${GITHUB_BASE}/releases" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64)  echo amd64 ;;
    aarch64 | arm64) echo arm64 ;;
    *)               error "Unsupported architecture: $(uname -m). Download manually from ${GITHUB_BASE}/releases" ;;
  esac
}

# --- version resolution ---
resolve_version() {
  if [ -n "${WORKLOG_VERSION:-}" ]; then
    printf '%s' "${WORKLOG_VERSION#v}"   # strip leading 'v' if present
    return
  fi
  version=$(fetch "${API_BASE}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"v?([^"]+)".*/\1/')
  [ -n "$version" ] || error "Could not resolve latest version. Set WORKLOG_VERSION and retry."
  printf '%s' "$version"
}

# --- install directory ---
resolve_install_dir() {
  if [ -n "${WORKLOG_INSTALL_DIR:-}" ]; then
    printf '%s' "$WORKLOG_INSTALL_DIR"
    return
  fi
  if [ -w /usr/local/bin ] || [ "$(id -u)" -eq 0 ]; then
    printf '%s' /usr/local/bin
  else
    printf '%s' "${HOME}/.local/bin"
  fi
}

# --- asset discovery: query the release JSON and pattern-match by OS + arch ---
# This avoids hardcoding the goreleaser naming convention (capitalisation, separators, etc.)
find_asset_url() {
  version=$1 os=$2 arch=$3
  url=$(fetch "${API_BASE}/releases/tags/v${version}" \
    | grep '"browser_download_url"' \
    | grep -i "\".*${os}.*${arch}.*\.tar\.gz\"" \
    | sed -E 's/.*"(https:[^"]+)".*/\1/' \
    | head -1)
  [ -n "$url" ] || error "No matching release asset for ${os}/${arch} in v${version}.
  Check available assets at: ${GITHUB_BASE}/releases/tag/v${version}"
  printf '%s' "$url"
}

find_checksums_url() {
  version=$1
  url=$(fetch "${API_BASE}/releases/tags/v${version}" \
    | grep '"browser_download_url"' \
    | grep '".*checksums\.txt"' \
    | sed -E 's/.*"(https:[^"]+)".*/\1/' \
    | head -1)
  printf '%s' "$url"   # empty string if not found — caller skips verification
}

# --- checksum verification ---
verify_checksum() {
  archive_path=$1 checksums_file=$2
  archive_name=$(basename "$archive_path")

  expected=$(grep " ${archive_name}$" "$checksums_file" 2>/dev/null | awk '{print $1}')
  if [ -z "$expected" ]; then
    warn "No checksum entry found for ${archive_name} — skipping verification."
    return
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$archive_path" | awk '{print $1}')
  elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$archive_path" | awk '{print $1}')
  else
    warn "sha256sum / shasum not found — skipping checksum verification."
    return
  fi

  [ "$expected" = "$actual" ] || error "Checksum mismatch for ${archive_name}!
  expected: ${expected}
  actual:   ${actual}
  The download may be corrupt or tampered with. Try again."

  success "Checksum OK"
}

# --- main ---
main() {
  os=$(detect_os)
  arch=$(detect_arch)
  version=$(resolve_version)
  install_dir=$(resolve_install_dir)

  printf '\n'
  info "Installing worklog v${version} (${os}/${arch}) → ${install_dir}/${BINARY}"
  printf '\n'

  asset_url=$(find_asset_url "$version" "$os" "$arch")
  checksums_url=$(find_checksums_url "$version")
  archive=$(basename "$asset_url")

  # Temp workspace; cleaned up on exit, interrupt, or error
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT INT TERM

  info "Downloading ${archive}..."
  fetch "$asset_url" "${tmp}/${archive}"

  if [ -n "$checksums_url" ]; then
    info "Verifying checksum..."
    fetch "$checksums_url" "${tmp}/checksums.txt"
    verify_checksum "${tmp}/${archive}" "${tmp}/checksums.txt"
  else
    warn "No checksums.txt found for v${version} — skipping verification."
  fi

  info "Extracting..."
  mkdir -p "${tmp}/extracted"
  tar -xzf "${tmp}/${archive}" -C "${tmp}/extracted"

  binary_path=$(find "${tmp}/extracted" -maxdepth 2 -type f -name "$BINARY" | head -1)
  [ -n "$binary_path" ] || error "Could not find '${BINARY}' in the downloaded archive. Please file a bug: ${GITHUB_BASE}/issues"

  mkdir -p "$install_dir"
  if [ -w "$install_dir" ]; then
    mv "$binary_path" "${install_dir}/${BINARY}"
    chmod +x "${install_dir}/${BINARY}"
  else
    info "Requesting sudo to install to ${install_dir}..."
    sudo mv "$binary_path" "${install_dir}/${BINARY}"
    sudo chmod +x "${install_dir}/${BINARY}"
  fi

  printf '\n'
  success "worklog v${version} installed to ${install_dir}/${BINARY}"

  # Verify the binary runs and warn if install_dir is not in PATH
  if command -v "$BINARY" >/dev/null 2>&1; then
    installed_version=$("${install_dir}/${BINARY}" version 2>/dev/null || true)
    [ -n "$installed_version" ] && info "Installed: ${installed_version}"
  else
    printf '\n'
    warn "${install_dir} is not in your PATH."
    warn "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    printf '\n  export PATH="%s:$PATH"\n\n' "$install_dir"
  fi
}

main "$@"
