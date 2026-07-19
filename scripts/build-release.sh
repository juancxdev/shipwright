#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${DIST_DIR:-$ROOT_DIR/dist}"
VERSION="${VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"
REPO="${SHIPWRIGHT_REPO:-${LOOM_REPO:-juancxdev/shipwright}}"

mkdir -p "$DIST_DIR"
rm -f "$DIST_DIR"/shipwright-* "$DIST_DIR"/loom-* "$DIST_DIR"/checksums.txt "$DIST_DIR"/latest.json

build_one() {
  local goos="$1"
  local goarch="$2"
  local ext=""
  if [[ "$goos" == "windows" ]]; then
    ext=".exe"
  fi

  local name="shipwright-${VERSION}-${goos}-${goarch}"
  local work="$DIST_DIR/$name"
  mkdir -p "$work"

  echo "==> building $name"
  (cd "$ROOT_DIR" && GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X shipwright/cmd.Version=$VERSION" -o "$work/shipwright$ext" .)

  cp "$ROOT_DIR/README.md" "$work/README.md"
  cp "$ROOT_DIR/LICENSE" "$work/LICENSE" 2>/dev/null || true

  if [[ "$goos" == "windows" ]]; then
    if command -v zip >/dev/null 2>&1; then
      (cd "$DIST_DIR" && zip -qr "$name.zip" "$name")
      cp "$DIST_DIR/$name.zip" "$DIST_DIR/shipwright-${goos}-${goarch}.zip"
      rm -rf "$work"
    else
      (cd "$DIST_DIR" && tar -czf "$name.tar.gz" "$name")
      cp "$DIST_DIR/$name.tar.gz" "$DIST_DIR/shipwright-${goos}-${goarch}.tar.gz"
      rm -rf "$work"
    fi
  else
    (cd "$DIST_DIR" && tar -czf "$name.tar.gz" "$name")
    cp "$DIST_DIR/$name.tar.gz" "$DIST_DIR/shipwright-${goos}-${goarch}.tar.gz"
    rm -rf "$work"
  fi
}

build_one darwin amd64
build_one darwin arm64
build_one linux amd64
build_one linux arm64
build_one windows amd64

(
  cd "$DIST_DIR"
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 shipwright-* > checksums.txt
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum shipwright-* > checksums.txt
  else
    echo "warning: no sha256 tool found; checksums.txt not generated" >&2
  fi
)

python3 - "$DIST_DIR" "$VERSION" "$REPO" <<'PY'
import json
import pathlib
import sys

dist = pathlib.Path(sys.argv[1])
version = sys.argv[2]
repo = sys.argv[3]
checksums = {}
checksums_file = dist / "checksums.txt"
if checksums_file.exists():
    for line in checksums_file.read_text().splitlines():
        parts = line.split()
        if len(parts) >= 2:
            checksums[parts[-1]] = parts[0]
assets = []
for path in sorted(dist.glob("shipwright-*")):
    if path.is_file():
        assets.append({
            "name": path.name,
            "sha256": checksums.get(path.name, ""),
            "url": f"https://github.com/{repo}/releases/download/{version}/{path.name}",
        })
manifest = {
    "name": "Shipwright",
    "version": version,
    "repo": repo,
    "assets": assets,
}
(dist / "latest.json").write_text(json.dumps(manifest, indent=2) + "\n")
PY

echo "Release artifacts written to $DIST_DIR"
