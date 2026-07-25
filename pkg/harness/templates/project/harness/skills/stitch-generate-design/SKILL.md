---
name: stitch-generate-design
description: Generate high-fidelity UI screens with Google Stitch SDK/MCP. Use when Shipwright design provider is Stitch, when OpenPencil is not desired, or when the user wants AI-generated responsive UI, variants, screenshots, HTML exports, DESIGN.md design systems, or Stitch-first design evidence.
---

# Stitch Generate Design

Use Google Stitch as the primary design provider for Shipwright UI/UX work.

Do not use OpenPencil unless the user explicitly asks for OpenPencil. Stitch is preferred when the goal is high-fidelity UI generation, variants, screenshots, HTML export, and design-to-development evidence.

## Preconditions

Stitch requires credentials:

- `STITCH_API_KEY`, or
- `STITCH_ACCESS_TOKEN` plus `GOOGLE_CLOUD_PROJECT`.

Shipwright stores local project credentials in `.harness/secrets.local.env` when configured through `shipwright init`. Do not print secret values.

Shipwright exposes Stitch to OpenCode through a generated MCP proxy under `.opencode/mcp`:

- MCP server name: `stitch`
- Expected tools: `stitch_*`
- Install command when dependencies are missing: `npm install --prefix .opencode/mcp`

Do not treat a missing standalone `stitch` CLI as the primary blocker. The normal path is MCP tools or the generated `@google/stitch-sdk` proxy. If credentials exist but `stitch_*` tools are absent, report that the MCP dependencies need installation/restart, not that credentials are missing.

If credentials are unavailable, report `blocked` for visual Stitch generation and continue only with doc-only artifacts after explaining the limitation.

## Required artifacts

Create or update:

- `.harness/artifacts/design/stitch/DESIGN.md` — design system/context passed to Stitch.
- `.harness/artifacts/design/stitch/design-task.md` — task plan and prompts.
- `.harness/artifacts/design/stitch/stitch-report.md` — Stitch project IDs, screen IDs, prompts, variants, exports, and limitations.
- `.harness/artifacts/design/stitch/exports/` — screenshots/images exported from Stitch.
- `.harness/artifacts/design/stitch/html/` — HTML exports when available.
- `.harness/artifacts/design/prototype.md` — final chosen design, screen flow, links to exports.
- `.harness/artifacts/design/responsive-qa.md` — mobile/tablet/desktop QA from exported screenshots.
- `.harness/artifacts/design/design-decisions.md` — selected direction and rejected variants.
- `.harness/artifacts/design/code-component-map.md` — when reusable UI components exist.

For existing UI baselines, also require:

- `.harness/artifacts/design/route-inventory.md`
- `.harness/artifacts/design/source-screenshots/`
- `.harness/artifacts/design/fidelity-report.md`

## Workflow

### 1. Load Shipwright design context

Read:

- `.harness/artifacts/product/context.md`
- `.harness/artifacts/product/scope.md`
- `.harness/artifacts/project/delivery-plan.md` when available
- `.harness/project-profile.md`
- `.harness/artifacts/design/stitch/DESIGN.md`

If recreating an existing UI, capture rendered route screenshots before Stitch generation. Do not rely only on source code.

### 2. Prepare Stitch prompt

The prompt must include:

- product goal and target users;
- required screens/routes;
- visual direction and brand constraints;
- responsive targets: mobile, tablet, desktop;
- accessibility expectations;
- source screenshots or route evidence when recreating existing UI;
- explicit instruction to preserve section order/content for baselines;
- required output: screenshots plus HTML.

### 3. Generate screens

Use Stitch SDK/MCP when available:

- create or reuse a Stitch project;
- generate screens for the required device types;
- generate variants only when exploration is useful;
- record project ID and screen IDs in `stitch-report.md`;
- export screenshot/image evidence for every target screen;
- export HTML when available.

Do not treat exported HTML as production code automatically. It is design evidence and implementation reference unless the frontend engineer explicitly adopts it.

### 4. Validate fidelity and responsiveness

For every generated screen:

- inspect screenshots for clipping, overlap, unreadable text, broken spacing, incorrect stacking, and generic/template-looking UI;
- verify mobile/tablet/desktop are adapted, not merely scaled;
- verify touch targets and contrast where possible;
- if baseline from existing UI, compare source screenshots vs Stitch screenshots and write `fidelity-report.md`.

A Stitch design can be visually impressive and still fail if it does not match the requested/current UI. Evidence wins over aesthetics.

### 5. Handoff

Update `prototype.md` with:

- selected Stitch direction;
- screen IDs and screenshot paths;
- HTML export paths;
- interaction flow;
- states: loading, empty, error, success;
- responsive behavior;
- known limitations.

Run `design-code-component-map` when reusable UI components exist so frontend tasks can reference `DCC-*` mappings.

## Completion rules

Do not claim Stitch design is complete if:

- credentials were missing;
- no screenshots were exported;
- no responsive QA was written;
- generated UI materially diverges from an existing UI baseline;
- HTML exports are presented as production code without frontend review;
- no `stitch-report.md` exists.

Return `partial` if Stitch generated useful screens but some routes/viewports are missing or fidelity is not verified. Return `blocked` if Stitch credentials or tools are unavailable.
