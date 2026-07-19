# 13 — Distribution & Installer

Phase 10 packages Shipwright as a global CLI.

## Goals

- Single-command install for macOS/Linux.
- PowerShell install for Windows.
- Cross-platform release artifacts.
- Checksums for verification.
- `latest.json` release manifest.
- Keep OpenCode as the default executor.

## Commands

Recommended project bootstrap:

```bash
shipwright init
```

Explicit equivalent:

```bash
shipwright init --ai opencode
shipwright init --executor opencode
```

Backward-compatible model overrides:

```bash
shipwright init --executor opencode \
  --reasoning-model opencode-go/deepseek-v4-flash \
  --fast-model opencode-go/deepseek-v4-flash

shipwright executor generate opencode \
  --reasoning-model opencode-go/deepseek-v4-flash \
  --fast-model opencode-go/deepseek-v4-flash
```

## Release artifacts

`scripts/build-release.sh` creates:

- `shipwright-<version>-darwin-amd64.tar.gz`
- `shipwright-<version>-darwin-arm64.tar.gz`
- `shipwright-<version>-linux-amd64.tar.gz`
- `shipwright-<version>-linux-arm64.tar.gz`
- `shipwright-<version>-windows-amd64.zip`
- versionless aliases for latest installers
- `checksums.txt`
- `latest.json`

## Installer variables

macOS/Linux:

```bash
SHIPWRIGHT_REPO=juancxdev/shipwright
SHIPWRIGHT_VERSION=latest
SHIPWRIGHT_INSTALL_DIR=$HOME/.local/bin
SHIPWRIGHT_BIN_NAME=shipwright
```

Windows PowerShell:

```powershell
$env:SHIPWRIGHT_REPO="juancxdev/shipwright"
$env:SHIPWRIGHT_VERSION="latest"
$env:SHIPWRIGHT_INSTALL_DIR="$HOME\.shipwright\bin"
```

## Post-install verification

```bash
shipwright version
shipwright doctor
```

`shipwright doctor` must still be run inside an initialized Shipwright project. Use `shipwright init` first for new projects.

## Creating the first release

Commit all distribution files, then create and push a semantic tag:

```bash
git add .
git commit -m "feat: add Shipwright distribution installer"
git tag v0.11.0
git push origin main --tags
```

The GitHub Actions release workflow publishes the release assets. The curl installer will fail until those assets exist.
