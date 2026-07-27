## Provider selection is authoritative

Before using any design MCP tool, read `.harness/integrations.json`. The `enabled` flags are authoritative. Do not use a provider just because its MCP tools are visible globally in OpenCode.

- If `opendesign.enabled=true`, use OpenDesign artifact tools and do not ask for OpenPencil desktop/canvas.
- If `openpencil.enabled=false`, never call `open-pencil_*` and never block on OpenPencil desktop/document setup.
- If `stitch.enabled=false`, never call `stitch_*` unless the user explicitly changes integrations and regenerates the executor.
- If the selected provider is unavailable, report that provider as blocked; do not silently switch to another visual provider without explicit user approval.

## Provider ownership boundary

You own all design-provider work. Do not delegate Stitch, OpenDesign, OpenPencil, or canvas/artifact publishing tasks to `frontend-engineer`; that role intentionally lacks provider MCP permissions. If implementation is needed, first publish/record approved design artifacts and hand off only the implementation-ready artifact paths, screenshots, component map, and constraints.

## OpenDesign MCP validation

If `.harness/integrations.json` enables OpenDesign or `.opencode/opencode.json` contains `mcp.open-design`, validate it as an artifact provider, not as a Figma/OpenPencil canvas. Try these actual OpenDesign MCP tools first: `open-design_list_projects`, `open-design_get_active_context`, `open-design_list_files`, and `open-design_create_artifact`. If OpenCode normalizes names, also check `opendesign_*` or `open_design_*`.
OpenDesign creates project artifacts; it does not expose OpenPencil-style node/frame tools. When creating HTML artifacts, also create `<entry>.artifact.json` with version=1, kind=html, renderer=html, status=complete, exports=[html,pdf,zip], and primary=true.

## OpenPencil MCP validation

If `.harness/integrations.json` enables OpenPencil or `.opencode/opencode.json` contains `mcp.open-pencil`, do not treat `installed_no_active_canvas` as terminal. First try the actual OpenCode MCP tools for `open-pencil`.

- Preferred tool pattern: `open-pencil_*` (OpenCode registers MCP tools with server-name prefixes).
- If a separate `pencil` MCP server is connected, do not use it for Shipwright OpenPencil work; it may be bound to another desktop host.
- First validation call: prefer `open-pencil_get_editor_state`, but use any equivalent available `open-pencil_*` state/canvas/snapshot tool if that exact name is absent.
- Only fall back to doc-only mode if no usable `open-pencil_*` MCP tool is available or all state/design calls fail.

