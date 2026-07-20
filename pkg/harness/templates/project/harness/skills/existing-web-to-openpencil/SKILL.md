---
name: existing-web-to-openpencil
description: Reverse-engineer an existing web UI into an OpenPencil prototype. Use when asked to analyze an existing website, Astro/React/Next/Vue/Svelte frontend, landing page, app screen, screenshot, or live local route and recreate its current UI in OpenPencil before making design changes.
---

# Existing Web to OpenPencil

Use this skill to convert an existing web interface into a structured OpenPencil design that can be reviewed, modified, and handed off to frontend.

## Operating mode

Prefer evidence over invention:

1. Inventory all candidate routes/views before choosing frames.
2. Inspect the existing implementation and rendered output before drawing.
3. Create a visual inventory before drawing.
4. Recreate the current UI in OpenPencil before proposing changes.
5. Validate the canvas against source evidence before claiming completion.
6. If OpenPencil is unavailable, produce doc-only artifacts and clearly mark the fallback.

Do not claim a faithful baseline from code inspection alone. A baseline is only faithful when it is compared against rendered evidence: browser screenshot, supplied screenshot, production screenshot, or an explicitly documented fallback when rendering is impossible.

## Required inputs

Accept any of these sources:

- Existing frontend project files.
- A local route that can be launched by the project dev command.
- A production/staging URL provided by the user.
- Screenshots supplied by the user.
- A specific page/component path such as `src/pages/index.astro` or `src/components/Header.astro`.

If the target page is ambiguous, do not assume a single page. Build a route/view inventory, identify which routes are active in the app, and ask the user which subset to recreate unless the request clearly says "all public views" or "entire site".

## Artifact contract

Write or update these files:

- `.harness/artifacts/design/route-inventory.md` — all discovered routes/views, source files, status, and whether each is included in the baseline.
- `.harness/artifacts/design/reverse-engineering.md` — source analysis, inspected routes/files, assumptions, gaps.
- `.harness/artifacts/design/visual-inventory.md` — layout, sections, components, typography, color, spacing, imagery, responsive states.
- `.harness/artifacts/design/fidelity-report.md` — source evidence vs OpenPencil export comparison, missing sections, mismatches, and approval status.
- `.harness/artifacts/design/openpencil/design-task.md` — OpenPencil execution plan and target canvas/frame list.
- `.harness/artifacts/design/openpencil/exports/` — OpenPencil exports when available.
- `.harness/artifacts/design/source-screenshots/` — source/browser screenshots when screenshots can be captured or supplied.
- `.harness/artifacts/design/ux-brief.md`, `.harness/artifacts/design/wireframes.md`, or `.harness/artifacts/design/prototype.md` only when they need to reflect the recreated UI.

## Workflow

### 1. Inventory routes/views before drawing

Read the project profile first when available:

- `.harness/project-profile.md`
- `.harness/project-profile.json`
- `.harness/skill-assignments.md`

For Astro projects, inspect:

- `astro.config.*`
- `src/pages/**/*.astro`
- `src/layouts/**/*.astro`
- `src/components/**/*.{astro,tsx,jsx,ts,js}`
- `src/styles/**/*`
- `public/**/*`
- `package.json`

For other frameworks, inspect equivalent routes, layouts, components, styles, and public assets.

Write `.harness/artifacts/design/route-inventory.md` before creating OpenPencil frames. Include a table with:

- Route/view name.
- URL/path.
- Source file(s).
- Served by current app: yes/no/unknown.
- Included in baseline: yes/no.
- Reason if excluded.

For Astro, route inventory must include every discovered route in `src/pages`. Also call out root-level `index.html` or standalone prototypes separately as "not served by Astro" unless config/build proves otherwise. Do not merge standalone prototypes into the Astro baseline without user approval.

Default coverage rule:

- If the user says "landing actual", baseline the active landing route plus all visible in-page states/sections for that route.
- If the user says "sitio", "todas las vistas", "portfolio", "web actual", or gives no explicit route, inventory all public routes and ask before excluding any route.
- If there is a mismatch between served Astro pages and standalone HTML/prototypes, stop and ask whether to baseline the served app, the standalone concept, or both.

### 2. Capture rendered source evidence

Use the strongest available source of truth:

1. If a local dev command exists, ask to run it or use the already-running URL.
2. If a production/staging URL is provided, use that.
3. If browser access is unavailable, ask the user for screenshots.
4. If no rendered evidence is available, mark the baseline as `approximation-from-code` in `.harness/artifacts/design/fidelity-report.md` and do not call it faithful.

For each included route, capture or reference source evidence for desktop and mobile at minimum; tablet is recommended. Store source screenshots under `.harness/artifacts/design/source-screenshots/` when possible.

### 3. Build the visual inventory

Capture:

- Page purpose and primary user action.
- Route and viewport targets.
- Layout grid, section order, content hierarchy.
- Components and repeated patterns.
- Typography scale, font families, weights, line heights.
- Color tokens, gradients, borders, radius, elevation, shadows.
- Spacing rhythm and container widths.
- Images, icons, decorative assets.
- Interactive states: hover, active, disabled, focus, empty, error, loading.
- Responsive behavior for mobile, tablet, desktop.
- Accessibility issues discovered during inspection.

### 4. Recreate current UI before changing it

Do not jump directly into redesign.

Create OpenPencil frames that represent the current state for every included route/view:

- Frame names must include route and viewport, e.g. `Home / Desktop`, `About / Mobile`.
- Use real content from the rendered/source page; do not replace important copy with generic lorem ipsum.
- Preserve section order, visual hierarchy, spacing rhythm, colors, imagery, and responsive behavior before improving anything.
- Do not omit sections because they are long; create taller frames instead.
- Mark approximations explicitly in `.harness/artifacts/design/reverse-engineering.md` and `.harness/artifacts/design/fidelity-report.md`.

### 5. Use OpenPencil correctly

If OpenPencil is enabled:

1. Use the `open-pencil` MCP server, not unrelated `pencil` servers bound to another IDE.
2. Start with an available OpenPencil state/canvas/snapshot tool. Prefer `open-pencil_get_editor_state` when present, but accept equivalent tools such as `open-pencil_get_state`, `open-pencil_get_canvas`, `open-pencil_snapshot_layout`, or similarly named `open-pencil_*` state tools.
3. Use an available OpenPencil design/edit tool to create/update frames. Prefer `open-pencil_batch_design` when present.
4. Use an available OpenPencil export tool to export results when possible. Prefer `open-pencil_export_nodes` when present.
5. Use an available screenshot/snapshot tool and inspect it. Prefer `open-pencil_get_screenshot` or `open-pencil_snapshot_layout` when present.
6. Fix overflow, clipped content, bad alignment, and non-responsive frames before completion.
7. Only fall back to doc-only mode if no usable `open-pencil_*` tools are visible or all available state/design calls fail.

Never read or edit `.pen` files directly with filesystem tools.

### 6. Fidelity QA gate

Before claiming the baseline is ready, compare source evidence against OpenPencil exports.

Write `.harness/artifacts/design/fidelity-report.md` with:

- Coverage matrix: route × viewport × source evidence × OpenPencil frame × export.
- Missing routes/views.
- Missing or reordered sections.
- Text/content mismatches.
- Color/typography/spacing mismatches.
- Image/icon/asset mismatches.
- Responsive mismatches.
- Final status: `pass`, `conditional-pass`, or `fail`.

Rules:

- If any included route has no OpenPencil frame, status is `fail`.
- If any requested route/view was excluded without user approval, status is `fail`.
- If rendered source evidence was unavailable, status cannot be `pass`; use `conditional-pass` at best and explain why.
- If the baseline is materially different from the current page, do not proceed to redesign.

### 7. Design modification pass

Only after the current UI is recreated:

1. Ask or confirm the design change goal.
2. Create alternatives only when the user asks for options or the change is ambiguous.
3. Keep a before/after note in `.harness/artifacts/design/reverse-engineering.md`.
4. Update OpenPencil frames.
5. Run canvas QA again.

### 8. Handoff to frontend

When the user approves the design direction, provide:

- A short implementation summary.
- Component mapping: design frame/component → source file/component.
- Token changes needed.
- Responsive rules.
- Risks where implementation may diverge from the design.

## Astro-specific guidance

Astro projects often mix server-rendered `.astro` files with framework islands.

When analyzing Astro:

- Treat `src/pages` as route ownership.
- Treat `src/layouts` as global visual shell.
- Treat `.astro` components as first-class UI components.
- Identify islands by `client:*` directives.
- Preserve static/server-rendered behavior; do not assume everything is a React app.
- Check global styles imported from layouts and page frontmatter.
- Check content collections or markdown-driven sections if the route uses them.

## Completion checklist

Before saying the OpenPencil recreation is ready:

- `.harness/artifacts/design/route-inventory.md` exists and lists all discovered routes/views.
- Source routes/files inspected.
- Rendered source evidence captured or fallback documented.
- `.harness/artifacts/design/reverse-engineering.md` exists.
- `.harness/artifacts/design/visual-inventory.md` exists.
- `.harness/artifacts/design/fidelity-report.md` exists.
- OpenPencil was attempted when enabled.
- Current UI recreated before redesign.
- Every included route has desktop and mobile frames; tablet is included when requested or useful.
- OpenPencil export inspected against source evidence.
- Known gaps and approximations documented.
- Fidelity status is not `fail`.
