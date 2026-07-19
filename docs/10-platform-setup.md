# Shipwright Platform Setup — macOS, Windows, Linux, CI

This guide documents how Shipwright resolves configuration and integrations across platforms.

## Configuration model

Shipwright separates desired configuration from detected runtime state:

```txt
.harness/config.json        desired portable configuration
.harness/integrations.json  detected local integration state
```

Effective config precedence:

```txt
1. internal defaults
2. .harness/config.json
3. platform_overrides[GOOS]
4. environment variables
```

## Required baseline

Shipwright itself only requires the harness files and Go runtime while developing/testing the harness.

Engram and OpenPencil are optional integrations. If they are missing, Shipwright uses explicit fallbacks:

```txt
Engram missing      -> progress/decisions.md
OpenPencil missing  -> design-doc-only
```

## Environment variables

```txt
SHIPWRIGHT_HEALTH_TIMEOUT_MS   Health-check timeout in milliseconds
ENGRAM_BINARY            Absolute path to Engram binary
ENGRAM_HEALTH_URL        Engram health endpoint
OPENPENCIL_APP_PATH      Absolute path to OpenPencil app/install dir
OPENPENCIL_MCP_SERVER    Absolute path to legacy/bundled OpenPencil MCP JS server
OPENPENCIL_MCP_COMMAND   OpenPencil MCP command, usually openpencil-mcp
OPENPENCIL_CANVAS_ACTIVE Temporary signal until real MCP canvas handshake exists
```

## macOS

Recommended OpenPencil MCP setup:

```bash
npm install -g @open-pencil/mcp
which openpencil-mcp
```

Typical legacy/bundled OpenPencil locations:

```txt
/Applications/OpenPencil.app
/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs
~/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs
```

Example override:

```json
{
  "platform_overrides": {
    "darwin": {
      "openpencil": {
        "app_path": "/Applications/OpenPencil.app",
        "mcp_server_path": "/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs"
      }
    }
  }
}
```

## Windows

Prefer environment variables for per-machine paths:

```powershell
$env:ENGRAM_BINARY="C:\\Tools\\engram.exe"
$env:OPENPENCIL_MCP_SERVER="C:\\Tools\\OpenPencil\\mcp-server.cjs"
```

Example override:

```json
{
  "platform_overrides": {
    "windows": {
      "engram": {
        "binary_path": "C:\\Tools\\engram.exe"
      },
      "openpencil": {
        "mcp_server_path": "C:\\Tools\\OpenPencil\\mcp-server.cjs"
      }
    }
  }
}
```

## Linux

Typical OpenPencil candidates:

```txt
~/.local/share/OpenPencil/resources/mcp-server.cjs
/opt/OpenPencil/resources/mcp-server.cjs
```

Example env config:

```bash
export ENGRAM_BINARY=/usr/local/bin/engram
export OPENPENCIL_MCP_SERVER=/opt/OpenPencil/resources/mcp-server.cjs
```

## CI

Recommended CI behavior:

```bash
export SHIPWRIGHT_HEALTH_TIMEOUT_MS=500
shipwright config validate --json
shipwright doctor --json
```

CI should treat exit code `2` as a blocking config/runtime error. Missing optional integrations without config corruption should not fail the build.

## Operational commands

```txt
shipwright config show
shipwright config validate
shipwright config validate --json
shipwright config env
shipwright doctor
shipwright doctor --json
shipwright doctor --fix
shipwright integrations detect
shipwright integrations status
```

## Safety rules

- `doctor --fix` may create, normalize, or back up/recreate config.
- `doctor --fix` does not delete custom paths automatically.
- OpenPencil `.pen` files must never be read directly by filesystem tools.
- OpenPencil MCP health uses `node --check`, not `require(...)`, to avoid executing the server during diagnostics.
