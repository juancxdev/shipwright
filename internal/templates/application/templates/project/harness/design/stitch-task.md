# Stitch Design Task

Use Google Stitch as the primary design provider for this project. Do not use OpenPencil unless the user explicitly asks for it.

## Request

{{request}}

## Credentials

Use Stitch only when `STITCH_API_KEY` is set, or `STITCH_ACCESS_TOKEN` + `GOOGLE_CLOUD_PROJECT` are set. If credentials are unavailable, report blocked and continue with doc-only artifacts only after explaining the limitation.

## Workflow

1. Read `.harness/artifacts/product/context.md`, `.harness/artifacts/product/scope.md`, and `.harness/artifacts/design/stitch/DESIGN.md`.
2. If recreating an existing UI, capture source route screenshots first and write `.harness/artifacts/design/route-inventory.md`.
3. Use Stitch SDK/MCP to create or update a Stitch project and generate mobile/tablet/desktop screens.
4. Generate variants only when useful; choose one recommended direction and document alternatives.
5. Export screenshots to `.harness/artifacts/design/stitch/exports/`.
6. Export HTML to `.harness/artifacts/design/stitch/html/` when available.
7. Write `.harness/artifacts/design/stitch/stitch-report.md` with project ID, screen IDs, prompts, exports, and known limitations.
8. Write or update `.harness/artifacts/design/prototype.md`, `.harness/artifacts/design/responsive-qa.md`, `.harness/artifacts/design/design-decisions.md`, and `.harness/artifacts/design/code-component-map.md` when components exist.
9. For existing UI baselines, compare Stitch screenshots against source screenshots in `.harness/artifacts/design/fidelity-report.md`; do not pass if the design materially diverges.

## Guardrails

- Stitch is the design provider; OpenPencil is not required.
- Do not claim fidelity without source screenshot vs Stitch screenshot comparison.
- Do not claim implementation readiness without code-component-map for reusable UI.
- Do not commit generated HTML as production frontend unless the frontend engineer explicitly accepts it as source.
