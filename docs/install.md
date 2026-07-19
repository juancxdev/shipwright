# Installing Shipwright

Shipwright is distributed as a single CLI binary.

## Prerequisites

- Git
- OpenCode installed and configured
- Optional: Engram and OpenPencil MCP integrations

Install OpenCode first:

```bash
curl -fsSL https://opencode.ai/install | bash
```

On Windows, prefer WSL for the terminal workflow.


## Required GitHub Release

The curl installer downloads this script from `main`, but the script installs Shipwright by downloading prebuilt binaries from GitHub Releases.

That means this command only works after a tag/release exists and the release workflow has uploaded assets:

```bash
git tag v0.11.0
git push origin v0.11.0
```

Wait for the GitHub Actions release workflow to finish. The release must contain assets like:

```txt
shipwright-darwin-arm64.tar.gz
shipwright-darwin-amd64.tar.gz
shipwright-linux-arm64.tar.gz
shipwright-linux-amd64.tar.gz
shipwright-windows-amd64.zip
checksums.txt
latest.json
```

Then the installer works:

```bash
curl -fsSL https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.sh | bash
```

## macOS/Linux install

```bash
curl -fsSL https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.sh | bash
```

Useful overrides:

```bash
SHIPWRIGHT_REPO=juancxdev/shipwright \
SHIPWRIGHT_VERSION=v0.11.0 \
SHIPWRIGHT_INSTALL_DIR="$HOME/.local/bin" \
bash scripts/install.sh
```

## Windows PowerShell install

```powershell
iwr https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.ps1 -UseB | iex
```

If using WSL, use the macOS/Linux install command inside WSL instead.


### PATH handling

The installer installs `shipwright` into `$HOME/.local/bin` by default. If that directory is not in `PATH`, the installer tries to update your shell profile automatically:

- Bash: `~/.bashrc`
- Zsh: `~/.zshrc`
- Fish: `~/.config/fish/config.fish`
- Fallback: `~/.profile`

Because `curl | bash` runs in a child shell, it cannot update the current terminal process directly. After installation, either open a new terminal or run the `source ...` command printed by the installer.

Disable automatic profile changes with:

```bash
SHIPWRIGHT_NO_PATH_UPDATE=1 curl -fsSL https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.sh | bash
```

## Verify

```bash
shipwright version
shipwright help
```

## Start a Shipwright/OpenCode project

```bash
mkdir my-project
cd my-project
shipwright init
opencode
```

`shipwright init` defaults to OpenCode. These forms are equivalent:

```bash
shipwright init
shipwright init --ai opencode
shipwright init --executor opencode
```

Model overrides remain supported:

```bash
shipwright init \
  --reasoning-model opencode-go/deepseek-v4-flash \
  --fast-model opencode-go/deepseek-v4-flash
```

And regeneration remains supported:

```bash
shipwright executor generate opencode \
  --reasoning-model opencode-go/deepseek-v4-flash \
  --fast-model opencode-go/deepseek-v4-flash
```

## Build release artifacts locally

```bash
VERSION=v0.11.0 ./scripts/build-release.sh
```

Artifacts are written to `dist/` with SHA256 checksums.
