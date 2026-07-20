---
name: openpencil-generate-design
description: Create or update OpenPencil canvas views using available open-pencil_* MCP tools. Use for OpenPencil-specific frame creation, node updates, layout, components, variables, exports, screenshots, save handling, and QA evidence.
---

# OpenPencil Generate Design

Use this skill with `canvas-generate-design` whenever the target design tool is OpenPencil.

## Tool selection

Use the `open-pencil` MCP server only. Do not use a separate `pencil` server unless the user explicitly says Shipwright is configured for that server.

Tool names can vary. Prefer capabilities, not exact names.

### Read/state tools

Prefer available tools such as:

- `open-pencil_get_current_page`
- `open-pencil_list_pages`
- `open-pencil_get_page_tree`
- `open-pencil_get_selection`
- `open-pencil_page_bounds`
- `open-pencil_describe`
- `open-pencil_node_tree`
- `open-pencil_node_bounds`
- `open-pencil_viewport_get`

### Create/update tools

Prefer available tools such as:

- `open-pencil_render`
- `open-pencil_create_shape`
- `open-pencil_update_node`
- `open-pencil_batch_update`
- `open-pencil_set_layout`
- `open-pencil_set_layout_child`
- `open-pencil_set_fill`
- `open-pencil_set_font`
- `open-pencil_set_text`
- `open-pencil_create_component`
- `open-pencil_create_instance`
- `open-pencil_import_svg`
- `open-pencil_insert_icon`

### Export/save tools

Prefer available tools such as:

- `open-pencil_export_image`
- `open-pencil_export_svg`
- `open-pencil_export_pdf`
- `open-pencil_save_file`
- `open-pencil_open_file`
- `open-pencil_new_document`

## Required OpenPencil workflow

1. Verify connection with a read-only tool such as `open-pencil_get_current_page` or `open-pencil_get_page_tree`.
2. Inspect the current page/canvas before drawing: pages, page tree, existing frames, selection, bounds, components, variables, and reusable nodes when those tools exist.
3. If the page is empty and the task is to create a baseline, create frames only after route/source inventory is complete.
4. Create a wrapper frame for each required view/viewport before section nodes.
5. Build incrementally: one view/viewport or one major section at a time.
6. After every major create/update call, read back page tree/node bounds and record affected node/frame IDs in `.harness/artifacts/design/openpencil/design-task.md`.
7. Use layout/grouping tools to keep section hierarchy clean.
8. Use components/instances for repeated elements when practical; a primitive-only result must be documented as low-fidelity or tool-limited.
9. Export each completed frame to `.harness/artifacts/design/openpencil/exports/`.
10. Inspect exports against source evidence or design intent. Fix visible mismatches before continuing.
11. Save the OpenPencil document using the save protocol below.
12. If save still times out or fails after the protocol, do not claim `.pen` persistence. Report the save failure as a blocker and keep exports plus recovery metadata as temporary evidence.

## Save protocol

OpenPencil save must be treated as a real gate, not a best-effort note.

Target file:

- `.harness/artifacts/design/openpencil/app.pen`

Before creating many nodes:

1. If no document is open, create or open `.harness/artifacts/design/openpencil/app.pen` with the available `open-pencil_new_document` or `open-pencil_open_file` tool.
2. After initial document creation/open, run a small save check before generating a large canvas.
3. If the save check fails, stop and report the blocker before doing expensive design work.

After frame generation:

1. Export PNG/SVG evidence first so visual work is not lost if save hangs.
2. Call `open-pencil_save_file` with explicit target path when the tool supports a path argument.
3. If save times out, wait briefly and retry once.
4. After a successful save response, verify persistence by checking one of:
   - the tool response includes a saved path;
   - `open-pencil_open_file` can reopen `.harness/artifacts/design/openpencil/app.pen`;
   - a filesystem-visible `.pen` exists at `.harness/artifacts/design/openpencil/app.pen` when the environment exposes it.
5. Write `.harness/artifacts/design/openpencil/save-status.md` with attempted path, attempts, result, error text, exports available, and whether `.pen` persistence is verified.

If save fails after retry:

- Do not say the `.pen` file is saved.
- Mark save status as `failed` or `unverified`.
- Keep exports under `.harness/artifacts/design/openpencil/exports/`.
- Write enough recovery data in `.harness/artifacts/design/openpencil/save-status.md` to recreate the frames: frame names, frame IDs, export paths, and failed error.
- Tell the user this is a blocker before moving to redesign or approval.

Completion language:

- Use “OpenPencil canvas active, exports persisted, save unverified” when exports exist but `.pen` save failed/timed out.
- Use “OpenPencil document saved and verified” only after verification.

## Frame naming

Use consistent names:

- `{Route or View} / Desktop`
- `{Route or View} / Tablet`
- `{Route or View} / Mobile`
- `{Component} / Variants`

For full-page landings, allow taller frames. Do not compress content just to fit 1024 height.

## Quality rules

- Do not declare done if page bounds are 0×0 after creation.
- Do not declare done if exported images do not exist.
- Do not declare done if a requested route/view has no frame.
- Do not declare done if frame/node IDs were not recorded when tools returned IDs.
- Do not declare done if the exported frame visibly diverges from the source screenshot for an existing UI baseline.
- Do not declare done if repeated UI is represented only as unrelated primitives without documenting why components/instances could not be used.
- Do not declare done if save failed without reporting it.
- Do not declare `.pen` persistence unless save verification passed.
- Do not proceed to redesign approval while save status is `failed` unless the user explicitly accepts export-only evidence.
- Do not delete or overwrite existing user-created frames unless explicitly instructed.

## Handoff

Update:

- `.harness/artifacts/design/canvas-plan.md`
- `.harness/artifacts/design/responsive-qa.md`
- `.harness/artifacts/design/fidelity-report.md` when recreating existing UI
- `.harness/artifacts/design/openpencil/design-task.md`

Return frame IDs, export paths, save status from `.harness/artifacts/design/openpencil/save-status.md`, and blockers.
