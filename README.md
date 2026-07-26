<p align="center">
  <h1 align="center">Shipwright</h1>
</p>

<p align="center">
  An agentic software delivery harness for turning vague product ideas into scoped, designed, specified, reviewed software work.
</p>

<p align="center">
  <a href="https://github.com/juancxdev/shipwright/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/juancxdev/shipwright?style=flat-square" /></a>
  <a href="https://github.com/juancxdev/shipwright/actions/workflows/release.yml"><img alt="Release workflow" src="https://img.shields.io/github/actions/workflow/status/juancxdev/shipwright/release.yml?style=flat-square" /></a>
  <a href="https://github.com/juancxdev/shipwright/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/github/license/juancxdev/shipwright?style=flat-square" /></a>
</p>

---

## What is Shipwright?

Shipwright is a local CLI that prepares your project so an AI coding agent can work like a real software team instead of jumping directly from a vague prompt to code.

It creates a controlled delivery lifecycle with roles, documents, gates, approvals, contracts, QA evidence, memory, and optional design tooling.

Instead of this:

```txt
"Build me an invoicing system" -> generated code
```

Shipwright pushes the workflow toward this:

```txt
Product discovery
-> scope review
-> project planning
-> UX/UI design when needed
-> technical architecture
-> contract-first FE/BE backlog
-> implementation
-> QA/security/contract evidence
-> user acceptance
-> change management
```

## Why?

AI agents are powerful, but software delivery is not only code generation.

Real teams need:

- product discovery,
- scope control,
- technical planning,
- UX validation,
- architecture decisions,
- contracts between frontend and backend,
- QA/security evidence,
- approvals,
- and change management.

Shipwright packages that lifecycle into a project-local harness that OpenCode or another AI executor can follow.

## How it works

Shipwright first calibrates the current repository, then creates a project structure and executor configuration:

```txt
.harness/          state, config, project profile, agents, approvals, integrations
.opencode/         OpenCode agents, commands, skills, config
.harness/artifacts/product/           context, assumptions, scope, open questions
.harness/artifacts/project/           PMBOK-lite project plan, risks, delivery plan
.harness/artifacts/design/            UX brief, flows, prototype, responsive QA
.harness/artifacts/architecture/      system architecture, technology options
.harness/artifacts/contracts/         OpenAPI and integration contracts
.harness/artifacts/backlog/           epics, stories, frontend/backend tasks
.harness/artifacts/sdd/               proposal, spec, design, tasks
.harness/artifacts/reports/           QA, security, contract and review evidence
.harness/artifacts/progress/          handoffs, decisions, history
```

The generated OpenCode setup makes `shipwright-orchestrator` the default agent. You can then open OpenCode and type the product request directly.

## Installation

### 1. Install OpenCode

```bash
curl -fsSL https://opencode.ai/install | bash
```

### 2. Install Shipwright

macOS/Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.sh | bash
```

Windows PowerShell:

```powershell
iwr https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.ps1 -UseB | iex
```

> On Windows, WSL is recommended for the terminal workflow.

### Installation directory

The install script uses:

1. `$SHIPWRIGHT_INSTALL_DIR` when provided.
2. `$HOME/.local/bin` by default on macOS/Linux.
3. `$HOME\.shipwright\bin` by default on Windows.

Example:

```bash
SHIPWRIGHT_INSTALL_DIR=/usr/local/bin \
  curl -fsSL https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.sh | bash
```

### PATH handling

If `$HOME/.local/bin` is not active in your current shell, the installer tries to update your shell profile automatically:

- Bash: `~/.bashrc`
- Zsh: `~/.zshrc`
- Fish: `~/.config/fish/config.fish`
- Fallback: `~/.profile`

Because `curl | bash` runs in a child shell, it cannot update the current terminal process directly. After installation, open a new terminal or run the `source ...` command printed by the installer.

Disable automatic profile changes with:

```bash
SHIPWRIGHT_NO_PATH_UPDATE=1 curl -fsSL https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.sh | bash
```

## First use

Create a new project folder:

```bash
mkdir my-product
cd my-product
shipwright init
```

`shipwright init` defaults to OpenCode, calibrates the current repository, then clears the terminal and opens an optional interactive integration checklist in interactive terminals. Move with `↑/↓` or `j/k`, toggle with `Space`, and continue with `Enter`.

```txt
> [X] Engram      — memoria persistente y decisiones del proyecto
  [X] Stitch      — diseño high-fidelity principal
  [ ] OpenDesign  — diseño/canvas MCP local
  [ ] OpenPencil  — canvas local opcional
```

If you enable Stitch and paste an API key, Shipwright stores it in:

```txt
.harness/secrets.local.env
```

That file is ignored by `.harness/.gitignore` and is intended only for your local machine. For scripts/CI, skip prompts:

```bash
shipwright init --no-interactive
```

`shipwright init` generates:

- `AGENTS.md`
- `.opencode/opencode.json`
- `.opencode/agents/*.md`
- `.opencode/commands/*.md`
- `.opencode/skills/*/SKILL.md`
- `.harness/project-profile.json` and `.harness/project-profile.md`
- `.harness/skill-registry.json` and `.harness/skill-registry.md`
- `.harness/skill-digests.json` and `.harness/skill-digests.md`
- `.harness/tdd-policy.json` and `.harness/tdd-policy.md`
- `.harness/artifacts/*` lifecycle documents

Then open OpenCode:

```bash
opencode
```

Now type a product request:

```txt
Create a simple invoicing platform: customers, services, draft invoices, and a basic dashboard. Do not include legal tax integration yet.
```

The orchestrator should start the Shipwright lifecycle automatically.

## Common commands

```bash
shipwright init                          # Bootstrap current folder for OpenCode
shipwright init --no-interactive         # Bootstrap without prompts for scripts/CI
shipwright status                        # Show current phase, gates, integrations, next action
shipwright doctor                        # Diagnose config/integrations/fallbacks
shipwright integrations detect           # Detect Engram/Stitch/OpenDesign/OpenPencil availability
shipwright integrations enable engram    # Enable Engram memory integration
shipwright integrations enable stitch    # Enable Stitch design integration
shipwright integrations configure opendesign --help # Configure OpenDesign MCP for OpenCode
shipwright integrations enable openpencil # Optional: enable OpenPencil only if explicitly needed
shipwright executor status opencode      # Verify OpenCode files
shipwright executor generate opencode    # Regenerate OpenCode executor files
shipwright skills refresh                  # Refresh reusable skill index + assignments
shipwright skills status                   # Show indexed skills
shipwright skills assign                   # Detect project stack and assign recommended skills
shipwright skills digest frontend-engineer  # Show compact skill rules for one role
shipwright tdd status                    # Show TDD mode/evidence gate
shipwright agents active                 # Show current active role
shipwright next                          # Advance when artifacts/gates are satisfied
shipwright approve scope                 # Approve scope gate
shipwright approve ux-design             # Approve UX gate
shipwright approve technical-plan        # Approve technical plan gate
```


## Project calibration

`shipwright init` is not only a scaffold command. It now detects the current repository before configuring agents.

It writes:

```txt
.harness/project-profile.json
.harness/project-profile.md
```

The profile captures detected stack, package manager, test/build/lint/dev commands, repository structure, CI/Docker hints, existing artifacts, and recommended TDD mode. OpenCode agents are instructed to read this profile before making technical assumptions.


## Strict TDD policy

When project calibration detects a reliable test command, Shipwright writes:

```txt
.harness/tdd-policy.json
.harness/tdd-policy.md
```

If the policy mode is `strict`, Shipwright blocks `IMPLEMENTATION → INTEGRATION` until implementation progress contains real test evidence. Check it with:

```bash
shipwright tdd status
```

Accepted evidence can live in:

```txt
.harness/artifacts/progress/frontend.md
.harness/artifacts/progress/backend.md
.harness/artifacts/reports/tdd-report.md
```

This prevents agents from saying “done” without executed tests.

## OpenCode configuration

Shipwright defaults to OpenCode:

```bash
shipwright init
```

These are equivalent:

```bash
shipwright init --ai opencode
shipwright init --executor opencode
```

Choose models at bootstrap time:

```bash
shipwright init \
  --reasoning-model openai/gpt-5.5 \
  --fast-model opencode-go/deepseek-v4-flash
```

Override models per role:

```bash
shipwright init \
  --agent-model technical-lead=openai/gpt-5.5 \
  --agent-model product-owner=opencode-go/deepseek-v4-flash
```

Regenerate OpenCode configuration later:

```bash
shipwright executor generate opencode \
  --reasoning-model opencode-go/deepseek-v4-flash \
  --fast-model opencode-go/deepseek-v4-flash
```

Generated OpenCode assets live in `.opencode/`.


## Skill registry

`shipwright init` also indexes reusable agent skills after generating executor assets.

It writes:

```txt
.harness/skill-registry.json
.harness/skill-registry.md
```

The registry lets the orchestrator discover available skills without loading every skill file into context. Refresh it after adding or regenerating skills:

```bash
shipwright skills refresh
shipwright skills status
shipwright skills digest frontend-engineer
```


### Skill assignment

Shipwright can detect the project stack and recommend skills per agent:

```bash
shipwright skills assign
shipwright skills assign --json
```

This writes:

```txt
.harness/skill-assignments.json
.harness/skill-assignments.md
```

The assignment harness is inspired by the `autoskills` reference, but it does not rescan the project independently. It consumes `.harness/project-profile.json` generated by `shipwright init` / `CalibrateProject()`, then maps calibrated technologies and combos to role-specific skills. Missing recommended skills are reported as gaps instead of being silently assumed.

For greenfield projects, Shipwright also watches Technical Lead artifacts such as `.harness/artifacts/architecture/technology-options.md` and `.harness/artifacts/architecture/system-architecture.md`. When `shipwright next` advances through technical planning phases, it extracts the planned stack into `.harness/project-profile.json`, refreshes `.harness/skill-assignments.*`, and refreshes `.harness/skill-digests.*` so OpenCode agents use the stack the TL chose even before code exists.

Shipwright also installs a small local curated UI/UX skill pack into `.opencode/skills/` during OpenCode bootstrap:

- `frontend-design`
- `stitch-generate-design`
- `existing-web-to-openpencil`
- `canvas-generate-design`
- `openpencil-generate-design`
- `accessibility`
- `responsive-layout-systems`
- `design-system-tokens`
- `interaction-design-patterns`
- `openpencil-canvas-qa`
- `design-code-component-map`
- `visual-handoff-to-frontend`

These are generic lifecycle skills, not remote downloads. They give `ui-ux-designer`, `frontend-engineer`, and `qa-security-reviewer` stronger defaults for interface quality, Stitch-first high-fidelity generation, Figma-like canvas discipline, existing-web reverse engineering, responsive behavior, accessibility, optional OpenPencil QA, design-code mapping, and design-to-frontend handoff.

Stitch is the primary design provider. It uses `STITCH_API_KEY`, or `STITCH_ACCESS_TOKEN` + `GOOGLE_CLOUD_PROJECT`, and stores evidence under `.harness/artifacts/design/stitch/`: `DESIGN.md`, `design-task.md`, `stitch-report.md`, screenshot exports, and HTML exports when available.

OpenPencil is optional. If explicitly used, save remains a gate: agents must export visual evidence, attempt to save `.harness/artifacts/design/openpencil/app.pen`, retry once on timeout, write `.harness/artifacts/design/openpencil/save-status.md`, and only claim `.pen` persistence after verification.

For existing web projects, including Astro sites, the baseline workflow tells the UI/UX Designer to inventory all routes/views first; inspect routes, layouts, components, styles, rendered screenshots, and assets; write `.harness/artifacts/design/route-inventory.md`, `.harness/artifacts/design/reverse-engineering.md`, `.harness/artifacts/design/visual-inventory.md`, and `.harness/artifacts/design/fidelity-report.md`; generate the current UI with Stitch; compare Stitch exports against source evidence; then apply requested design changes only after the baseline fidelity gate passes.

Skill pack rules live as declarative manifests in `internal/skills/application/templates/project/harness/skill-packs/`. Shipwright does not execute external installers during `init`; if you already used autoskills or another provider that writes `.agents/skills`, import those skills explicitly:

```bash
shipwright skills providers
shipwright skills import autoskills
```

The import copies `.agents/skills/*` into `.opencode/skills/*`, then refreshes registry, assignments, and digests.

## Shipwright Studio

Shipwright Studio is now a separate project in the Shipwright ecosystem.

Use the core CLI for lifecycle orchestration and executor setup. Use `shipwright-studio` for the browser Hub/lobby and visual project rooms.

## Optional integrations

Shipwright works without optional integrations. Missing integrations fall back safely.

### Engram memory

Engram stores decisions, discoveries, bug fixes, and session summaries.

```bash
shipwright integrations detect
shipwright integrations enable engram
shipwright doctor
```

If Engram is unavailable, Shipwright falls back to:

```txt
.harness/artifacts/progress/decisions.md
```

### Stitch design

Stitch is the primary high-fidelity design provider. The easiest path is the `shipwright init` wizard. Manual setup also works:

```bash
export STITCH_API_KEY="..."
shipwright integrations detect
shipwright integrations enable stitch
shipwright executor generate opencode
npm install --prefix .opencode/mcp
```

Alternative OAuth-style setup uses `STITCH_ACCESS_TOKEN` plus `GOOGLE_CLOUD_PROJECT`. You can also let the init wizard store `STITCH_API_KEY` in `.harness/secrets.local.env` for this project only.

Shipwright generates a local OpenCode MCP proxy at `.opencode/mcp/stitch-proxy.mjs` using `@google/stitch-sdk`. After installing dependencies, restart OpenCode and verify:

```bash
opencode mcp list
```

You should see a connected server named `stitch`; the UI/UX Designer can then use `stitch_*` tools. Stitch evidence is written under `.harness/artifacts/design/stitch/`.

### OpenDesign design MCP (optional)

If you run OpenDesign locally, do not edit `.opencode/opencode.json` by hand. Configure it through Shipwright once, then regenerate OpenCode assets:

```bash
shipwright integrations configure opendesign \
  --command "/opt/homebrew/Cellar/node@24/24.15.0/bin/node" \
  --arg "/Users/developer/Documents/TOOLS/APPs/open-design/apps/daemon/dist/cli.js" \
  --arg mcp \
  --data-dir "/Users/developer/Documents/TOOLS/APPs/open-design/.od" \
  --ipc-path "/tmp/open-design/ipc/default/daemon.sock"

shipwright executor generate opencode
opencode mcp list
```

Shipwright writes `.opencode/mcp/open-design.sh` and registers an `open-design` MCP server in `.opencode/opencode.json`. The UI/UX Designer can then use exact OpenDesign artifact tools such as `open-design_get_active_context`, `open-design_list_projects`, and `open-design_create_artifact` when OpenCode reports the server as connected.

OpenDesign is treated as an artifact provider, not a Figma/OpenPencil canvas. Generated HTML artifacts must include a sidecar manifest named `<entry>.artifact.json`; otherwise OpenDesign can reject imports with `ARTIFACT_MANIFEST_REQUIRED`. Shipwright generates the `opendesign-generate-artifact` skill so the UI/UX Designer knows this contract.

### OpenPencil design (optional)

Use OpenPencil only when you explicitly want a local OpenPencil canvas. Recommended MCP setup:

```bash
npm install -g @open-pencil/mcp
which openpencil-mcp
shipwright integrations detect
shipwright integrations enable openpencil
shipwright executor generate opencode
```

OpenCode should show:

```bash
opencode mcp list
```

Expected server:

```txt
open-pencil connected
```

The UI/UX Designer uses `open-pencil_*` tools and must produce responsive evidence:

```txt
.harness/artifacts/design/responsive-qa.md
```

## Lifecycle phases

| Phase | Main role | Output |
|---|---|---|
| Discovery | Product Owner | product context, assumptions, open questions |
| Scope review | Product Owner + user | approved product scope |
| Planning | Project Manager | project plan, risks, delivery plan |
| UX/UI design | UI/UX Designer | UX brief, flows, prototype, responsive QA |
| Technical design | Technical Lead | architecture, contracts, SDD artifacts |
| Backlog | Technical Lead | epics, stories, FE/BE tasks |
| Implementation | Frontend + Backend | working implementation |
| QA/security review | QA/Security Reviewer | reports and findings |
| Acceptance | User | final approval or change request |

## Contract-first frontend/backend

Shipwright coordinates frontend and backend through:

```txt
.harness/artifacts/contracts/openapi.yaml
.harness/artifacts/backlog/frontend-tasks.md
.harness/artifacts/backlog/backend-tasks.md
```

Rules:

- frontend keeps mock mode and real API mode,
- mocks must stay aligned with the contract,
- backend implements against the contract,
- contract changes require technical approval/change request.

## Design quality gate

UX work is not considered ready without responsive evidence.

Required before UX approval:

```txt
.harness/artifacts/design/ux-brief.md
.harness/artifacts/design/user-flows.md
.harness/artifacts/design/prototype.md
.harness/artifacts/design/responsive-qa.md
```

The UI/UX Designer must check:

- mobile `390x844`,
- tablet `768x1024`,
- desktop `1440x1024`,
- no overflow outside the canvas,
- no clipped content,
- no horizontal scroll,
- touch targets at least `44x44`,
- readable text and WCAG AA contrast targets.

## Troubleshooting

Run:

```bash
shipwright doctor
```

Machine-readable output:

```bash
shipwright doctor --json
```

Safe config repair:

```bash
shipwright doctor --fix
```

If `shipwright` is installed but your shell cannot find it:

```bash
$HOME/.local/bin/shipwright version
source ~/.bashrc # or ~/.zshrc, depending on your shell
```

## Release/install development

Build release artifacts locally:

```bash
VERSION=v0.11.0 ./scripts/build-release.sh
```

Create a release tag:

```bash
git tag v0.11.0
git push origin v0.11.0
```

The release workflow publishes:

```txt
shipwright-darwin-arm64.tar.gz
shipwright-darwin-amd64.tar.gz
shipwright-linux-arm64.tar.gz
shipwright-linux-amd64.tar.gz
shipwright-windows-amd64.zip
checksums.txt
latest.json
```

## License

Shipwright source code is licensed under the [Apache License 2.0](LICENSE).

Third-party notices for the core CLI are documented in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md). Shipwright Studio assets are documented in the separate `shipwright-studio` project.

## Documentation

- [`docs/install.md`](docs/install.md)
- [`docs/12-opencode-integration.md`](docs/12-opencode-integration.md)
- [`docs/13-distribution-installer.md`](docs/13-distribution-installer.md)
- [`docs/14-project-calibration-harness.md`](docs/14-project-calibration-harness.md)
- [`docs/15-skill-registry-harness.md`](docs/15-skill-registry-harness.md)
- [`docs/16-skill-digestion-harness.md`](docs/16-skill-digestion-harness.md)
- [`docs/17-strict-tdd-harness.md`](docs/17-strict-tdd-harness.md)
- [`docs/22-usage-guide.md`](docs/22-usage-guide.md)
- [`docs/24-template-assets.md`](docs/24-template-assets.md)
- [`docs/troubleshooting.md`](docs/troubleshooting.md)
- [`docs/00-vision.md`](docs/00-vision.md)

## Status

Shipwright is currently an early agentic delivery harness. The OpenCode-first flow is the recommended path.
