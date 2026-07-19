#!/usr/bin/env bash
set -euo pipefail

REPO="${SHIPWRIGHT_REPO:-${LOOM_REPO:-juancxdev/shipwright}}"
VERSION="${SHIPWRIGHT_VERSION:-${LOOM_VERSION:-latest}}"
INSTALL_DIR="${SHIPWRIGHT_INSTALL_DIR:-${LOOM_INSTALL_DIR:-$HOME/.local/bin}}"
BIN_NAME="${SHIPWRIGHT_BIN_NAME:-${LOOM_BIN_NAME:-shipwright}}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

log() { printf '==> %s\n' "$*"; }
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

path_has_dir() {
  case ":$PATH:" in
    *":$1:"*) return 0 ;;
    *) return 1 ;;
  esac
}

profile_file_for_shell() {
  local shell_name
  shell_name="$(basename "${SHELL:-}")"
  case "$shell_name" in
    zsh) printf '%s\n' "$HOME/.zshrc" ;;
    fish) printf '%s\n' "$HOME/.config/fish/config.fish" ;;
    bash) printf '%s\n' "$HOME/.bashrc" ;;
    *)
      if [[ -f "$HOME/.bashrc" ]]; then
        printf '%s\n' "$HOME/.bashrc"
      elif [[ -f "$HOME/.zshrc" ]]; then
        printf '%s\n' "$HOME/.zshrc"
      else
        printf '%s\n' "$HOME/.profile"
      fi
      ;;
  esac
}

add_install_dir_to_profile() {
  if path_has_dir "$INSTALL_DIR"; then
    return 0
  fi
  if [[ "${SHIPWRIGHT_NO_PATH_UPDATE:-}" == "1" ]]; then
    return 1
  fi

  local profile_file shell_name line marker
  profile_file="$(profile_file_for_shell)"
  shell_name="$(basename "${SHELL:-sh}")"
  marker="# Shipwright installer"

  mkdir -p "$(dirname "$profile_file")"
  touch "$profile_file"

  if grep -Fq "$INSTALL_DIR" "$profile_file" 2>/dev/null; then
    printf '%s\n' "$profile_file"
    return 0
  fi

  if [[ "$shell_name" == "fish" ]]; then
    line="fish_add_path $INSTALL_DIR"
  else
    line="export PATH=\"$INSTALL_DIR:\$PATH\""
  fi

  {
    printf '\n%s\n' "$marker"
    printf '%s\n' "$line"
  } >> "$profile_file"

  printf '%s\n' "$profile_file"
}

need curl
need tar

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$OS" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  *) fail "unsupported OS: $OS" ;;
esac
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) fail "unsupported architecture: $ARCH" ;;
esac

if [[ "$VERSION" == "latest" ]]; then
  URL="https://github.com/${REPO}/releases/latest/download/shipwright-${OS}-${ARCH}.tar.gz"
else
  URL="https://github.com/${REPO}/releases/download/${VERSION}/shipwright-${VERSION}-${OS}-${ARCH}.tar.gz"
fi

ARCHIVE="$TMP_DIR/shipwright.tar.gz"
log "downloading $URL"
if ! curl -fsSL "$URL" -o "$ARCHIVE"; then
  cat >&2 <<MSG
error: download failed from:
  $URL

Shipwright installers download prebuilt binaries from GitHub Releases.
Make sure you have created and pushed a tag, for example:

  git tag v0.11.0
  git push origin v0.11.0

Then wait for the GitHub release workflow to publish assets like:
  shipwright-${OS}-${ARCH}.tar.gz

Overrides:
  SHIPWRIGHT_REPO=$REPO
  SHIPWRIGHT_VERSION=$VERSION
MSG
  exit 1
fi

mkdir -p "$TMP_DIR/extract" "$INSTALL_DIR"
tar -xzf "$ARCHIVE" -C "$TMP_DIR/extract"
BIN_PATH="$(find "$TMP_DIR/extract" -type f -name shipwright -perm -111 | head -n 1 || true)"
if [[ -z "$BIN_PATH" ]]; then
  BIN_PATH="$(find "$TMP_DIR/extract" -type f -name shipwright | head -n 1 || true)"
fi
[[ -n "$BIN_PATH" ]] || fail "shipwright binary not found in archive"

install -m 0755 "$BIN_PATH" "$INSTALL_DIR/$BIN_NAME"
log "installed $BIN_NAME to $INSTALL_DIR/$BIN_NAME"

UPDATED_PROFILE=""
if ! path_has_dir "$INSTALL_DIR"; then
  UPDATED_PROFILE="$(add_install_dir_to_profile || true)"
fi

if command -v "$BIN_NAME" >/dev/null 2>&1; then
  "$BIN_NAME" version || true
else
  "$INSTALL_DIR/$BIN_NAME" version || true
fi

if ! path_has_dir "$INSTALL_DIR"; then
  cat <<MSG

$BIN_NAME was installed, but $INSTALL_DIR is not active in this terminal session.
MSG
  if [[ -n "$UPDATED_PROFILE" ]]; then
    cat <<MSG
The installer added it to:
  $UPDATED_PROFILE

Apply it now with:
  source "$UPDATED_PROFILE"

Or open a new terminal.
MSG
  else
    cat <<MSG
Add this to your shell profile:
  export PATH="$INSTALL_DIR:\$PATH"

Or run Shipwright directly with:
  $INSTALL_DIR/$BIN_NAME
MSG
  fi
fi

cat <<MSG
Next steps:
  mkdir my-project && cd my-project
  $BIN_NAME init
  opencode
MSG
