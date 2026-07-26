# Evidence-first Web Baseline Task

Shipwright requires baseline evidence before OpenDesign, Stitch, OpenPencil, or doc-only redesign output. Do not invent assets. Do not replace logos, icons, imagery, fonts, or copy without explicit user approval.

## Request

{{request}}

## Required outputs

- `.harness/artifacts/design/route-inventory.md` listing every discovered route/view and included/excluded status.
- `.harness/artifacts/design/source-screenshots/` with desktop, tablet, and mobile source screenshots for included routes.
- `.harness/artifacts/design/asset-manifest.json` with all logos, images, icons, fonts, and key brand assets; mark logo assets with `role: "logo"`.
- `.harness/artifacts/design/visual-inventory.md`, `.harness/artifacts/design/token-inventory.md`, `.harness/artifacts/design/component-inventory.md`, and `.harness/artifacts/design/code-component-map.md` when applicable.
- `.harness/artifacts/design/fidelity-report.md` with `Status: pass|conditional-pass|fail`, route coverage, asset preservation, screenshot comparison, and provider publish status.
{{routes_section}}
## Hard gates

- `baseline-captured`: route inventory and source screenshots exist.
- `assets-preserved`: `asset-manifest.json` includes real assets and generated design does not substitute them.
- `provider-published`: selected provider publish/import succeeded, or user explicitly accepted fallback.
- `fidelity-verified`: fidelity report is not fail/partial without explicit acceptance.
- `token/quota-ok`: provider did not fail due to token/quota exhaustion.
