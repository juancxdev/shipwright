# OpenDesign Artifact Task

Use OpenDesign as an artifact design provider. Do not treat OpenDesign as OpenPencil/Figma canvas; it exposes project/file/artifact tools.

## Role ownership

This task must be executed by `ui-ux-designer` only. Do not delegate it to `frontend-engineer`; Frontend intentionally does not have OpenDesign/Stitch/OpenPencil provider MCP permissions. If another role receives this task, stop and report a routing error back to the orchestrator.

## Request

{{request}}

## Required skills

Load and follow `.opencode/skills/opendesign-generate-artifact/SKILL.md`. For existing UI baselines, also follow `existing-web-to-openpencil` for route/screenshot/fidelity discipline, but publish through OpenDesign artifacts rather than OpenPencil frames.

## MCP tools

Try these exact OpenCode tool names first: `open-design_get_active_context`, `open-design_list_projects`, `open-design_list_files`, `open-design_create_artifact`. If absent, check `opendesign_*` and `open_design_*`.

## Required outputs

- `.harness/artifacts/design/opendesign/<entry>.html` or equivalent artifact entry.
- `.harness/artifacts/design/opendesign/<entry>.html.artifact.json` sidecar manifest.
- `.harness/artifacts/design/opendesign/opendesign-report.md` with MCP calls, active project/context, artifact ID/name, manifest path, and limitations.
- `.harness/artifacts/design/prototype.md`, `.harness/artifacts/design/responsive-qa.md`, and `.harness/artifacts/design/fidelity-report.md` for existing UI baselines.


## Daemon connection rules

- Read `.harness/config.json` and `.harness/integrations.json` before checking OpenDesign. If `opendesign.daemon_url` is configured, the generated `.opencode/mcp/open-design.sh` wrapper exports it as `OD_DAEMON_URL`.
- Do not start OpenDesign daemon processes from the agent unless the user explicitly asks. Prefer reporting the configured daemon URL/status and asking the user to start OpenDesign.
- Never run `cli.js daemon --ipc-path`; OpenDesign daemon does not support that flag. `--ipc-path` is a Shipwright configuration flag that maps to `OD_SIDECAR_IPC_PATH`.
- If OpenDesign MCP cannot connect, report `blocked` with: configured command, daemon URL, data dir, IPC path, and the exact `shipwright integrations configure opendesign --daemon-url ...` command to fix it.

## Guardrails

- If `ARTIFACT_MANIFEST_REQUIRED` appears, create `<entry>.artifact.json` and retry; do not mark pass without a manifest.
- If OpenDesign MCP publish/import fails, status is `partial` unless the user explicitly accepts local artifact fallback.
- Do not claim OpenDesign canvas/frame completion; OpenDesign completion means artifact creation/publish plus evidence.
