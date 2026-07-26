---
name: canvas-generate-design
description: Create or update complete UI views in a Figma-like design canvas using frames, components, tokens, variables, auto-layout-like structure, responsive variants, exports, and fidelity checks. Use when generating screens/views in OpenPencil, Figma, or any canvas design tool from product requirements, existing web pages, screenshots, or route inventories.
---

# Canvas Generate Design

Use this skill to create structured UI views in a design canvas. The goal is not to draw approximate boxes; the goal is to build maintainable canvas design: frames, layout hierarchy, components, tokens, variants, responsive views, and evidence.

## Core principle

Build like a designer using a design system, not like a screenshot tracer.

Prefer in this order:

1. Existing design system components/variables/styles.
2. Project design tokens from code or prior artifacts.
3. Reusable local components created in the canvas.
4. Primitives only when no component/token exists.

## Figma-inspired canvas discipline

Apply this discipline to OpenPencil, Figma, or any Figma-like canvas:

1. **Inspect first** — before writing nodes, inspect the current canvas/page, existing frames, components, variables/tokens, styles, and naming conventions when the tool supports it. Match the file; do not impose a new convention blindly.
2. **Component-first** — for buttons, cards, nav items, form fields, badges, pricing cards, repeated sections, and icons, reuse or create components/instances before duplicating primitives. A primitives-only screen is acceptable only for explicit low-fidelity wireframes or documented tool limitation.
3. **Token-first** — extract colors, typography, spacing, radius, shadows, and breakpoints from code/design artifacts before hardcoding values. Preserve original CSS variable names or token names where available.
4. **Incremental construction** — create wrapper frames first, then build one section or component family at a time. Validate after each major section before continuing.
5. **Readback required** — every create/update step must return or record affected frame/node IDs when the tool exposes IDs. Use those IDs for validation, exports, fixes, and handoff.
6. **Evidence beats intent** — exported/screenshot evidence is the authority. If exports differ from the source screenshot or route, fix the canvas before saying the design is ready.

## Required artifacts

Create or update:

- `.harness/artifacts/design/canvas-plan.md` — target views, frames, components, tokens, and tool strategy.
- `.harness/artifacts/design/component-inventory.md` — reusable components and source mapping.
- `.harness/artifacts/design/token-inventory.md` — color, type, spacing, radius, shadows, breakpoints.
- `.harness/artifacts/design/responsive-qa.md` — viewport validation and overflow checks.

If recreating existing UI, also require:

- `.harness/artifacts/design/route-inventory.md`
- `.harness/artifacts/design/fidelity-report.md`

## Workflow

### 1. Determine source mode

Classify the task:

- `greenfield` — create UI from requirements/scope.
- `existing-ui-baseline` — recreate current UI from app/routes/screenshots.
- `redesign` — modify an approved baseline.
- `component-system` — create components/tokens without full screens.

If the mode is `existing-ui-baseline`, follow `existing-web-to-openpencil` or equivalent fidelity workflow before redesigning.

### 2. Discover design system inputs

Inspect available sources:

- `.harness/artifacts/design/*`
- `.harness/project-profile.md`
- `.harness/skill-assignments.md`
- Code tokens: CSS variables, Tailwind config, theme files, design tokens JSON, component libraries.
- Canvas components/variables/styles when the MCP tool supports them.
- Existing canvas frames/screens that already use the project's design language.

Write `.harness/artifacts/design/token-inventory.md` and `.harness/artifacts/design/component-inventory.md` before building large screens.

Token inventory must include source evidence: CSS variable, Tailwind/theme path, existing canvas variable/style, or explicit approximation.
Component inventory must include source mapping: canvas component/instance or source component path, props/variants/states, and whether it was reused or recreated.

If the canvas already has components/styles, inspect those first. Only create new local components/tokens when no equivalent exists or when the user asked for a new visual system.

### 3. Plan frames and variants

Write `.harness/artifacts/design/canvas-plan.md` with:

- Views/screens to create.
- Viewport frames: mobile 390×844, tablet 768×1024, desktop 1440×1024 or route-specific full-height frames.
- Components to reuse/create.
- Token strategy.
- Interaction states required.
- Export/evidence plan.

Do not create only desktop unless the user explicitly requests desktop-only.

### 4. Build structured frames

For each view:

- Create a top-level frame named `{View} / {Viewport}`.
- Create the page/view wrapper before section children; do not create many top-level orphan sections and try to reparent later.
- Use nested frames/groups that reflect semantic sections: header, hero, feature grid, pricing, footer, etc.
- Use auto-layout-like structure when the tool supports it.
- Keep hierarchy clean: avoid hundreds of unrelated primitives at page root.
- Use real copy/content when available.
- Use components for repeated UI.
- Use variables/tokens for color, type, spacing, radius, and effects.
- Keep section order and responsive behavior explicit.
- After each major section, inspect/read back bounds and hierarchy when possible; fix overflow, off-canvas nodes, clipped text, or wrong stacking before adding the next section.

### 5. Validate canvas quality

Before claiming completion:

- Export or screenshot each frame.
- Inspect for clipping, overlap, off-canvas elements, unreadable text, incorrect stacking, bad spacing, and broken responsive behavior.
- Confirm touch targets are at least 44×44 for mobile.
- Confirm contrast is WCAG AA where possible.
- Update `.harness/artifacts/design/responsive-qa.md`.

### 6. Fidelity gate for existing UI

When recreating an existing UI:

- Compare source evidence against canvas export.
- Compare route coverage, viewport coverage, section order, content/copy, imagery/icons, typography, colors, spacing rhythm, and responsive behavior.
- Missing route/view means fail.
- Missing section means fail.
- Material layout/content mismatch means fail.
- Primitive-only approximations of component-rich UI mean fail unless explicitly marked as low-fidelity.
- No rendered evidence means conditional-pass at best.
- Do not proceed to redesign while fidelity status is fail.

### 7. Completion response

Report:

- Frames created/updated.
- Components/tokens created or reused.
- Exports/evidence files.
- Fidelity status when applicable.
- Known gaps/blockers.
- Whether redesign may proceed.

Never say “pixel-perfect” unless a visual diff or explicit manual comparison supports it.
