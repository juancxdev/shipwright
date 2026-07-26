---
name: opendesign-generate-artifact
description: Generate and publish OpenDesign artifacts through the OpenDesign MCP. Use when Shipwright design provider is OpenDesign, when the user explicitly asks for OpenDesign, or when creating HTML/React/Markdown/SVG design artifacts instead of OpenPencil/Figma canvas frames.
---

# OpenDesign Generate Artifact

OpenDesign is an **artifact provider**, not an OpenPencil/Figma-like canvas provider.

Use this skill when the design target is OpenDesign.

## Provider model

OpenDesign MCP exposes project/file/artifact tools. It does not expose node/frame canvas operations.

Expected OpenCode tool names usually use the `open-design` server prefix:

- `open-design_list_projects`
- `open-design_get_active_context`
- `open-design_get_artifact`
- `open-design_get_project`
- `open-design_get_file`
- `open-design_search_files`
- `open-design_list_files`
- `open-design_create_artifact`

If OpenCode normalizes the server name, also check equivalent `opendesign_*` or `open_design_*` tools before declaring the MCP unavailable.

## Required validation before fallback

Before falling back to doc-only or local-only HTML:

1. Read `.harness/config.json` and `.harness/integrations.json`.
2. Confirm OpenDesign is enabled/configured.
3. Check `.opencode/opencode.json` contains `mcp.open-design`.
4. Try a read-only OpenDesign MCP call:
   - first: `open-design_get_active_context`
   - second: `open-design_list_projects`
   - fallback names: `opendesign_get_active_context`, `open_design_get_active_context`.
5. If tools are unavailable, report `blocked` or `partial`; do not claim a complete OpenDesign handoff.

## Artifact manifest contract

Every OpenDesign HTML artifact must have a sidecar manifest next to it:

```txt
<entry>.artifact.json
```

For example:

```txt
.harness/artifacts/design/opendesign/sgc-open-design-artifact.html
.harness/artifacts/design/opendesign/sgc-open-design-artifact.html.artifact.json
```

Minimum valid HTML manifest:

```json
{
  "version": 1,
  "kind": "html",
  "title": "Project Design Prototype",
  "entry": "sgc-open-design-artifact.html",
  "renderer": "html",
  "status": "complete",
  "exports": ["html", "pdf", "zip"],
  "primary": true,
  "metadata": {
    "provider": "opendesign",
    "shipwright": true
  }
}
```

Allowed values known by OpenDesign:

- `kind`: `html`, `deck`, `react-component`, `markdown-document`, `svg`, `diagram`, `code-snippet`, `mini-app`, `design-system`
- `renderer`: `html`, `deck-html`, `react-component`, `markdown`, `svg`, `diagram`, `code`, `mini-app`, `design-system`
- `status`: `streaming`, `complete`, `error`
- `exports`: `html`, `pdf`, `zip`, `jsx`, `md`, `svg`, `txt`

## Evidence-first asset preservation

For existing apps, OpenDesign is an output provider. The source of truth is the rendered app baseline plus `.harness/artifacts/design/asset-manifest.json`.

Before calling `create_artifact`:

- read `.harness/artifacts/design/baseline/design-task.md`;
- read `.harness/artifacts/design/asset-manifest.json`;
- preserve logo, icons, images, fonts, colors, and meaningful copy;
- do not substitute a generic logo or generated brand mark;
- if a real asset cannot be embedded, document the exact missing asset and mark fidelity `fail` or `conditional-pass`, not `pass`.

## Existing web baseline workflow

When recreating an existing site/app:

1. Inventory all routes/views before designing.
2. Capture source evidence: route list, rendered screenshots, components, tokens, assets, responsive breakpoints.
3. Write:
   - `.harness/artifacts/design/route-inventory.md`
   - `.harness/artifacts/design/reverse-engineering.md`
   - `.harness/artifacts/design/visual-inventory.md`
   - `.harness/artifacts/design/token-inventory.md`
4. Generate the OpenDesign artifact as HTML/React/SVG/Markdown depending on the requested deliverable.
5. Generate the sidecar `.artifact.json` manifest.
6. Use `open-design_create_artifact` when available, passing the artifact name/content expected by the tool.
7. Write `.harness/artifacts/design/opendesign/opendesign-report.md` with project id/context, tool calls, artifacts, manifest path, limitations, and evidence.
8. Write `.harness/artifacts/design/fidelity-report.md` comparing source screenshots/route evidence against the generated artifact.

## Completion rules

Do not say OpenDesign is complete unless one of these is true:

- OpenDesign MCP `create_artifact` succeeded and report links the created artifact/project context.
- Or the user explicitly accepted local artifact fallback after being told OpenDesign MCP import/publish did not succeed.

If HTML exists but MCP publish/import failed, status is `partial`, not `pass`. If quota/tokens are exhausted, status is `blocked` and the UX gate cannot be approved.

If `ARTIFACT_MANIFEST_REQUIRED` appears, create the sidecar `<entry>.artifact.json` manifest and retry. Do not claim the error is undocumented after reading this skill.

## Output artifacts

Required when OpenDesign is used:

- `.harness/artifacts/design/opendesign/design-task.md`
- `.harness/artifacts/design/opendesign/<entry>.html` or equivalent artifact entry
- `.harness/artifacts/design/opendesign/<entry>.html.artifact.json`
- `.harness/artifacts/design/opendesign/opendesign-report.md`
- `.harness/artifacts/design/prototype.md`
- `.harness/artifacts/design/responsive-qa.md`
- `.harness/artifacts/design/fidelity-report.md` for existing UI baselines
