---
name: design-code-component-map
description: Map design canvas components/frames to frontend code components. Use when connecting OpenPencil/Figma-like design artifacts to code, creating design-to-code traceability, preparing frontend handoff, validating component parity, or adapting Figma Code Connect concepts without requiring Figma Organization/Enterprise Code Connect.
---

# Design Code Component Map

Use this skill to establish traceability between design components/screens and frontend code components.

This is inspired by Figma Code Connect, but Shipwright treats it as a portable harness artifact. Official Figma Code Connect can be used only when the user explicitly works in Figma with the required MCP tools, permissions, published library components, and plan support. For OpenPencil, create a local mapping artifact instead of pretending the official Figma feature exists.

## Purpose

Design-to-code mapping answers:

- Which design frame/component corresponds to which source component?
- Which props, variants, states, tokens, and assets must the frontend preserve?
- Which design elements do not yet have a code implementation?
- Which code components are not represented in the design?

## Required artifacts

Create or update:

- `.harness/artifacts/design/code-component-map.md` — human-readable mapping.
- `.harness/artifacts/design/code-component-map.json` — machine-readable mapping when practical.
- `.harness/artifacts/design/component-inventory.md` — design components and source components.
- `.harness/artifacts/backlog/frontend-tasks.md` — frontend tasks must reference mapping IDs when implementation work follows.

If recreating an existing UI, also update:

- `.harness/artifacts/design/fidelity-report.md` — include unmapped or mismatched components.

## Workflow

### 1. Discover design-side components

Inspect available design sources:

- OpenPencil page tree, node tree, selected nodes, components, instances, node bounds, and exports when `open-pencil_*` tools are available.
- `.harness/artifacts/design/openpencil/design-task.md` for frame names/IDs.
- `.harness/artifacts/design/openpencil/exports/` for visual evidence.
- `.harness/artifacts/design/component-inventory.md`, `.harness/artifacts/design/token-inventory.md`, and `.harness/artifacts/design/visual-inventory.md`.
- Figma URL/node IDs only when the user explicitly provides Figma context.

For every reusable design element, capture:

- Design name.
- Frame/component/node ID when available.
- Route/screen/viewport where it appears.
- Variants and states: size, variant, disabled, hover, active, loading, error, empty, etc.
- Text/content props.
- Tokens: color, typography, spacing, radius, shadow.
- Asset/icon dependencies.

Do not map a full screen directly to a single source component unless the code actually has a page-level component matching it.

### 2. Scan code-side components

Search the codebase before asking the user for paths.

Common locations:

- `src/components/**`
- `components/**`
- `app/components/**`
- `src/pages/**`
- `src/layouts/**`
- `src/ui/**`
- `lib/ui/**`
- `packages/**/src/components/**`

Framework hints:

- Astro: `.astro` components, `src/pages`, `src/layouts`, islands with `client:*`, imported React/Vue/Svelte components.
- React/Next/Vite: `.tsx`, `.jsx`, props interfaces, component exports, stories/tests.
- Vue/Nuxt: `.vue`, props/emits, slots.
- Svelte/SvelteKit: `.svelte`, exported props, slots.
- Angular: component selectors, inputs/outputs, templates.

For each candidate component, compare:

- Name similarity.
- Purpose and visual role.
- Prop names and variant values.
- Slot/children structure.
- DOM/component hierarchy.
- CSS classes, Tailwind utilities, CSS modules, style objects, or token usage.
- Existing stories/tests/docs that confirm intended states.

Look beyond name matching. A component with matching props and structure is a stronger match than a component with only a similar name.

### 3. Classify mapping confidence

Use explicit confidence levels:

- `exact` — name, purpose, props/variants, and structure align.
- `strong` — purpose and most props/states align; small naming differences exist.
- `partial` — component exists but missing variants/states/tokens.
- `candidate` — plausible match but user confirmation required.
- `missing-code` — design component has no code implementation.
- `missing-design` — code component has no design representation.
- `blocked` — insufficient design/code evidence.

Do not silently choose between multiple plausible candidates. Present options and ask for confirmation before treating a candidate as accepted.

### 4. Write mapping artifact

`.harness/artifacts/design/code-component-map.md` format:

```md
# Design ↔ Code Component Map

## Summary

- Design source: OpenPencil | Figma | screenshots | doc-only
- Codebase stack: {framework/language}
- Mapping status: pass | partial | blocked

## Mapping table

| ID | Design element | Design ref | Code component | Code path | Props/states mapped | Confidence | Notes |
|----|----------------|------------|----------------|-----------|---------------------|------------|-------|
| DCC-001 | Button / Primary | node 1:23 | Button | src/components/Button.tsx | variant, size, disabled | exact | - |
| DCC-002 | Pricing card | Home/Desktop section Pricing | PricingCard | src/components/PricingCard.astro | plan, price, cta | strong | Missing highlighted state |

## Missing code components

- {Design element}: {why no code match was found, suggested next task}

## Missing design components

- {Code component}: {why it should be represented in design}

## Ambiguous mappings needing user decision

- {Design element}: candidates `{path A}`, `{path B}`, reason ambiguity exists

## Frontend implications

- Components to reuse.
- Components to create.
- Props/states to preserve.
- Token changes needed.
- Tests/stories recommended.
```

`.harness/artifacts/design/code-component-map.json` should use stable IDs:

```json
{
  "version": 1,
  "status": "partial",
  "mappings": [
    {
      "id": "DCC-001",
      "design_name": "Button / Primary",
      "design_ref": "node 1:23",
      "code_component": "Button",
      "code_path": "src/components/Button.tsx",
      "props_states": ["variant", "size", "disabled"],
      "confidence": "exact",
      "notes": []
    }
  ],
  "missing_code": [],
  "missing_design": [],
  "ambiguous": []
}
```

### 5. Official Figma Code Connect mode

Use official Figma Code Connect only when all conditions are true:

- The user explicitly asks for Figma Code Connect or provides a Figma component URL/selection.
- The required Figma MCP tools are available, such as `get_code_connect_suggestions` and `send_code_connect_mappings`.
- The component is published to a team library.
- The user has required Figma plan/permissions.
- The user confirms the proposed mappings.

When parsing Figma URLs, convert URL `node-id=1-2` to tool `nodeId=1:2`.

If official Code Connect is unavailable, produce the local Shipwright mapping artifact instead and explain the limitation.

### 6. Update frontend handoff

Before frontend implementation starts, make sure tasks reference mapping IDs:

- `DCC-001 Button -> src/components/Button.tsx`
- `DCC-002 PricingCard -> src/components/PricingCard.astro`

Frontend tasks should not recreate components from scratch when a mapped component exists. If a design component is `missing-code`, create an explicit frontend task to implement it. If a code component is `missing-design`, create a design follow-up task unless the component is internal-only.

## Completion rules

Do not claim design-code mapping is complete if:

- no codebase scan was performed;
- no design-side components/frames were inspected;
- ambiguous mappings were auto-selected without user confirmation;
- mapped code paths do not exist;
- mapped component names are not exported or clearly identifiable;
- official Figma Code Connect was claimed without using the required Figma Code Connect tools.

Report `partial` when useful mappings exist but some components are missing or ambiguous. Report `blocked` when neither design-side nor code-side evidence is sufficient.
