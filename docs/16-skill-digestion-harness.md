# 16 — Skill Digestion Harness

Phase 13 turns the skill registry into compact, role-specific rules.

The registry answers:

> What skills exist?

The digest answers:

> Which skills matter for this agent right now, and what compact rules should it follow?

## Generated files

```txt
.harness/skill-digests.json   machine-readable compact skill rules
.harness/skill-digests.md     human/agent-readable compact skill rules
```

`shipwright init`, `shipwright executor generate ...`, and `shipwright skills refresh` update both registry and digests.

## Commands

```bash
shipwright skills digest                 # show digest summary for all agents
shipwright skills digest frontend-engineer # show compact rules for one agent
shipwright skills refresh                # refresh registry and digests
```

## Why digestion exists

Do not pass every full skill file to every subagent.

That creates noisy context, increases token cost, and can make the subagent follow unrelated rules.

Skill digestion keeps delegation small:

1. scan skill files into `.harness/skill-registry.*`,
2. match skills to Shipwright roles,
3. produce compact rules per agent,
4. let agents load full skill files only when deeper detail is needed.

## Matching rules

Current matching uses:

- direct role skill name/path matches;
- inferred capability tags:
  - `testing`,
  - `frontend`,
  - `backend`,
  - `design`,
  - `go`,
  - `typescript`,
  - `docs`;
- role-specific mappings:
  - `frontend-engineer` → frontend, TypeScript, testing;
  - `backend-engineer` → backend, Go, testing;
  - `ui-ux-designer` → design, frontend, docs;
  - `technical-lead` → frontend, backend, testing, Go, TypeScript, docs;
  - `qa-security-reviewer` → testing, backend, frontend.

## Project profile integration

Digests also include compact rules from `.harness/project-profile.md` / `.harness/project-profile.json`, such as:

- detected stack,
- detected test command,
- strict/suggested/no TDD mode.

That means implementation agents get project-aware rules like:

```txt
Use detected test command for evidence: `pnpm test`.
Strict TDD is available; create/adjust failing tests before implementation when changing behavior.
```

## Agent rules

Agents must:

1. prefer `.harness/skill-digests.md` over loading every skill file;
2. load full skill files only when the digest marks them relevant and extra detail is needed;
3. record a fallback if a required skill is missing;
4. refresh digests after adding or regenerating skills.


## Skill assignments

`shipwright skills refresh` also writes stack-aware assignments:

```txt
.harness/skill-assignments.json
.harness/skill-assignments.md
```

Assignments use `.harness/project-profile.json` as their source of truth. Technology detection stays centralized in `CalibrateProject()` / `shipwright init`; assignment only maps calibrated stack signals and combos to Shipwright agents. Installed skills are linked through the local skill registry; missing recommendations are treated as explicit gaps.

For greenfield work, `.harness/project-profile.json` can also contain `planned_stacks`. Shipwright derives these from Technical Lead artifacts (`.harness/artifacts/architecture/technology-options.md`, `.harness/artifacts/architecture/system-architecture.md`, `.harness/artifacts/project/delivery-plan.md`, `.harness/artifacts/sdd/design.md`, `.harness/artifacts/sdd/tasks.md`). `shipwright next` refreshes planned stack assignments automatically after technical-planning transitions, and `shipwright skills refresh|assign|digest` also re-read planned stack artifacts.

## Curated UI/UX skill pack

OpenCode bootstrap installs Shipwright-managed UI/UX skills into `.opencode/skills/`:

- `frontend-design` — product-aware web UI design
- `stitch-generate-design` — Google Stitch high-fidelity UI generation, variants, screenshots, HTML exports, and DESIGN.md workflow
- `existing-web-to-openpencil` — legacy name, but now used as the evidence-first existing web baseline workflow across Stitch/OpenPencil/canvas providers
- `canvas-generate-design` — Figma-like canvas/screen generation rules
- `openpencil-generate-design` — optional OpenPencil-specific MCP workflow when explicitly requested
- `accessibility` — WCAG-minded review rules
- `responsive-layout-systems` — mobile/tablet/desktop layout rules
- `design-system-tokens` — colors, typography, spacing, radius, elevation
- `interaction-design-patterns` — flow states, feedback, errors, destructive actions
- `openpencil-canvas-qa` — optional OpenPencil frame/export/overflow validation
- `design-code-component-map` — Figma Code Connect-inspired design ↔ code component traceability
- `visual-handoff-to-frontend` — design-to-frontend implementation handoff

When a detected or planned stack suggests UI/frontend work, assignments add these skills under `frontend-ui-quality` so designer, frontend, and QA roles get stronger interface guidance.

The mapping is declarative. Skill pack manifests live in `pkg/harness/templates/project/harness/skill-packs/`; Go code loads these manifests and evaluates them against `.harness/project-profile.json`.

## Stitch-first design provider

Shipwright now treats Google Stitch as the primary high-fidelity design provider. OpenPencil remains optional and should be used only when the user explicitly asks for it or disables Stitch.

Stitch artifacts live under:

```txt
.harness/artifacts/design/stitch/
  DESIGN.md
  design-task.md
  stitch-report.md
  exports/
  html/
  screens/
```

The UI/UX Designer must use Stitch credentials (`STITCH_API_KEY`, or `STITCH_ACCESS_TOKEN` + `GOOGLE_CLOUD_PROJECT`) to generate responsive screens, screenshot evidence, HTML exports when available, and `stitch-report.md`. Generated HTML is design evidence and implementation reference; it is not production frontend code until the Frontend Engineer accepts it.

Existing UI baselines are still evidence-first:

1. route inventory,
2. rendered screenshots/evidence per route and viewport,
3. visual inventory of sections/tokens/components/assets,
4. Stitch screen generation,
5. screenshot/HTML export validation,
6. fidelity report.

A Stitch design must not pass when sections are missing, route coverage is partial, screenshots materially diverge from the source route, or no source screenshot/rendered evidence was inspected.

## Canvas generation skills

Shipwright includes Figma-inspired canvas/design skills:

- `stitch-generate-design` — Stitch-specific high-fidelity generation workflow.
- `canvas-generate-design` — tool-agnostic rules for creating full screens/views with frames, components, tokens, responsive variants, exports, and fidelity gates.
- `openpencil-generate-design` — optional OpenPencil-specific workflow for available `open-pencil_*` tools.

These skills are intended to prevent primitive-only or pretty-but-wrong designs. Agents should build with reusable components, token inventories, clean hierarchy, responsive screens, and evidence before claiming a view is ready.

The workflow is inspired by Figma design-system skills but adapted to Stitch-first Shipwright:

- inspect existing app/design evidence before generation;
- reuse or create components/instances for repeated UI;
- extract tokens from code, CSS variables, themes, Tailwind config, or design-system artifacts;
- generate responsive mobile/tablet/desktop outputs;
- export screenshots and HTML evidence;
- validate exports against source screenshots after major steps;
- fix mismatches before proceeding to redesign or approval.

OpenPencil save remains a gate only when OpenPencil is explicitly used: agents must export visual evidence, attempt to save `.harness/artifacts/design/openpencil/app.pen`, retry once on timeout, write `.harness/artifacts/design/openpencil/save-status.md`, and only claim `.pen` persistence after verification.

## Optional provider import

Shipwright does not invoke autoskills during `shipwright init`. If a user already ran autoskills or any compatible tool that writes `.agents/skills`, they can opt in:

```bash
shipwright skills providers
shipwright skills import autoskills
```

`import autoskills` copies compatible skill directories into `.opencode/skills`, then refreshes skill registry, assignments, and digests.


## Existing web to OpenPencil

Shipwright includes the curated `existing-web-to-openpencil` skill for existing frontend projects. It is assigned through the `frontend-ui-quality` pack when the calibrated or planned stack suggests frontend/UI work, including Astro. The UI/UX Designer must inventory all discovered routes/views, inspect current routes/components/styles and rendered evidence, write `.harness/artifacts/design/route-inventory.md`, `.harness/artifacts/design/reverse-engineering.md`, `.harness/artifacts/design/visual-inventory.md`, and `.harness/artifacts/design/fidelity-report.md`, recreate the current UI in OpenPencil before redesigning, and compare exports against source evidence before claiming completion. If requested routes are missing or source evidence was unavailable, the fidelity report cannot pass.

The baseline flow is intentionally evidence-first:

1. route inventory,
2. rendered screenshots/evidence per route and viewport,
3. visual inventory of sections/tokens/components/assets,
4. OpenPencil frame creation,
5. export/readback validation,
6. fidelity report.

An existing UI baseline must not be treated as complete when sections are missing, route coverage is partial, the OpenPencil export visibly diverges from the source route, or no source screenshot/rendered evidence was inspected.


## Canvas generation skills

Shipwright includes two Figma-inspired canvas skills:

- `canvas-generate-design` — tool-agnostic rules for creating full screens/views with frames, components, tokens, responsive variants, exports, and fidelity gates.
- `openpencil-generate-design` — OpenPencil-specific workflow for available `open-pencil_*` tools such as page/tree reads, render/update operations, exports, save handling, and QA evidence.

These skills are intended to prevent primitive-only canvas drawings. Agents should build with reusable components, token inventories, clean hierarchy, responsive frames, and evidence before claiming a view is ready.

The workflow is inspired by Figma design-system skills but adapted to OpenPencil:

- inspect the current canvas/page before drawing;
- reuse or create components/instances for repeated UI;
- extract tokens from code, CSS variables, themes, Tailwind config, or existing canvas styles;
- create wrapper frames first and build section-by-section;
- record frame/node IDs when tools expose them;
- validate with page tree/bounds plus exports/screenshots after major steps;
- fix mismatches before proceeding to redesign or approval.


OpenPencil save is a gate: agents must export visual evidence, attempt to save `.harness/artifacts/design/openpencil/app.pen`, retry once on timeout, write `.harness/artifacts/design/openpencil/save-status.md`, and only claim `.pen` persistence after verification. If save times out, exports are temporary evidence and the save failure is a blocker unless the user explicitly accepts export-only work.

## Design ↔ code component mapping

Shipwright includes `design-code-component-map`, inspired by Figma Code Connect but portable across OpenPencil and other canvas tools. The skill writes `.harness/artifacts/design/code-component-map.md` and, when practical, `.harness/artifacts/design/code-component-map.json`.

The goal is traceability, not decoration: each reusable design element should map to an existing source component, a missing-code task, a missing-design task, or an explicit ambiguous candidate requiring user confirmation. Frontend tasks should reference stable `DCC-*` mapping IDs so implementation does not recreate components blindly.

Official Figma Code Connect remains Figma-specific and should only be used when the user explicitly works in Figma with the required MCP tools, published library components, permissions, and plan support. Otherwise Shipwright uses its local mapping artifact.
