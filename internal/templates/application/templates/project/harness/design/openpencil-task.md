# OpenPencil Design Task

**Project:** {{project_name}}
**Request:** {{request}}

## Objective

Use the OpenPencil MCP tools to create a visual design for this project.

## MCP-first validation

Shipwright CLI detection can report `installed_no_active_canvas` even when OpenPencil desktop is open, because only the MCP client can verify the live editor/canvas. Do **not** treat that status as failure.

- First try the OpenCode MCP tools from the `open-pencil` server.
- Expected OpenCode tool pattern: `open-pencil_*` because MCP tools are prefixed with the server name.
- If another MCP server named `pencil` is also connected, do **not** use it for Shipwright OpenPencil work; it can be bound to another app host.
- Only fall back to doc-only design if no `open-pencil_*` MCP tool is visible or the editor-state check fails.

## Steps

1. Call `open-pencil_get_editor_state` to verify OpenPencil is connected and an editor/canvas is reachable.
2. If no active canvas exists but the MCP responds, create or open one with `open-pencil_batch_design` or use the existing canvas.
3. Read `.harness/artifacts/design/ux-brief.md` for product context and design requirements.
4. Read `.harness/artifacts/design/user-flows.md` for the flows to design.
5. Create the design file at `.harness/artifacts/design/openpencil/app.pen`.
6. Create responsive frames for every key screen: mobile 390×844, tablet 768×1024, desktop 1440×1024.
7. Design mobile-first, then adapt tablet and desktop. Do not simply scale the same layout.
8. Export wireframes to `.harness/artifacts/design/openpencil/exports/` using `open-pencil_export_nodes`.
9. Take screenshots with `open-pencil_get_screenshot` and inspect them for overflow, clipping, hidden text, tiny tap targets, and horizontal scroll.
10. Create `.harness/artifacts/design/responsive-qa.md` with the responsive/accessibility checks and findings.
11. Create `.harness/artifacts/design/prototype.md` describing the visual design and how it maps to user flows.
12. Run `shipwright design status` to verify all artifacts are in place.

## Responsive Layout Contract

- Use an 8px spacing grid and keep outer safe margins: 16px mobile, 24px tablet, 32px desktop.
- No component may extend outside its frame/canvas.
- Avoid fixed-width containers that exceed the frame width.
- Primary actions must remain visible without horizontal scrolling.
- Interactive targets must be at least 44×44px.
- Body text should be at least 16px with readable line-height.
- Color contrast must target WCAG AA: 4.5:1 normal text, 3:1 large text/UI components.
- Every screen must include loading, empty, error, and success states where relevant.

## Rules

- **NEVER** read `.pen` files directly with filesystem tools.
- **ONLY** manipulate `.pen` files via OpenPencil MCP tools from the `open-pencil` server (`open-pencil_*`).
- **DO NOT** mark design complete if any screenshot shows overflowing components, clipped text, or content outside the canvas.
- The `.pen` file is NOT considered approved just because it exists.
- Human approval via `shipwright approve ux-design` is still **mandatory**.

## Required artifacts after completion

- `.harness/artifacts/design/openpencil/app.pen` — the design file
- `.harness/artifacts/design/openpencil/exports/` — exported wireframes/prototypes
- `.harness/artifacts/design/responsive-qa.md` — responsive/accessibility validation
- `.harness/artifacts/design/prototype.md` — description of the visual design

## After completion

Run: `shipwright next` to advance to UX_APPROVAL.
Then the user must approve: `shipwright approve ux-design`
