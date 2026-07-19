# Shipwright — Step-by-step Usage Guide

This guide explains how to use the core Shipwright CLI flow.

Shipwright is not just a code generator. It is a project-local software delivery harness: it creates roles, gates, documents, contracts, and evidence so AI agents can work more like a real software team.

## 1. Install Shipwright

### macOS/Linux

```bash
curl -fsSL https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.sh | bash
```

After installation, open a new terminal or source the profile file printed by the installer.

Check:

```bash
shipwright version
```

### Windows PowerShell

```powershell
iwr https://raw.githubusercontent.com/juancxdev/shipwright/main/scripts/install.ps1 -useb | iex
```

Check:

```powershell
shipwright version
```

## 2. Install OpenCode

Shipwright works best when OpenCode is the AI executor.

Install OpenCode following its official instructions, then check:

```bash
opencode --version
```

## 3. Create or enter a project

For a new project:

```bash
mkdir billing-mvp
cd billing-mvp
```

For an existing project:

```bash
cd your-existing-project
```

## 4. Initialize Shipwright

Basic init:

```bash
shipwright init
```

This defaults to OpenCode and creates:

```txt
.harness/
.opencode/
product/
project/
design/
architecture/
contracts/
backlog/
sdd/
knowledge/
progress/
reports/
AGENTS.md
```

Shipwright also calibrates the current repository and writes:

```txt
.harness/project-profile.md
.harness/tdd-policy.md
.harness/skill-registry.md
.harness/skill-digests.md
```

## 5. Choose OpenCode models

You can choose default role models during init:

```bash
shipwright init \
  --reasoning-model openai/gpt-5.5 \
  --fast-model opencode-go/deepseek-v4-flash
```

You can override individual agents:

```bash
shipwright init \
  --agent-model technical-lead=openai/gpt-5.5 \
  --agent-model product-owner=opencode-go/deepseek-v4-flash
```

Regenerate OpenCode config later:

```bash
shipwright executor generate opencode \
  --reasoning-model openai/gpt-5.5 \
  --fast-model opencode-go/deepseek-v4-flash
```

## 6. Check project health

Run:

```bash
shipwright doctor
```

If safe automatic fixes are available:

```bash
shipwright doctor --fix
```

Check executor status:

```bash
shipwright executor status opencode
```

## 7. Optional integrations

### Engram memory

Detect:

```bash
shipwright integrations detect
```

Enable:

```bash
shipwright integrations enable engram
```

Engram stores decisions, handoffs, discoveries, and session context.

### OpenPencil design

Detect:

```bash
shipwright integrations detect
```

Enable:

```bash
shipwright integrations enable openpencil
shipwright executor generate opencode
```

Inside OpenCode, verify MCP:

```bash
opencode mcp list
```

Expected server:

```txt
open-pencil connected
```

## 8. Start work from the CLI

You can start from the CLI:

```bash
shipwright start "Create a simple invoicing MVP: customers, services, draft invoices, and dashboard. No legal tax integration yet."
```

Then check status:

```bash
shipwright status
```

See active agent:

```bash
shipwright agents active
```

## 9. Recommended: work from OpenCode

After `shipwright init`, open OpenCode in the project:

```bash
opencode
```

Then type your product request naturally:

```txt
Create a simple invoicing MVP: customers, services, draft invoices, and dashboard. No legal tax integration yet.
```

The generated OpenCode config makes `shipwright-orchestrator` the default agent. It should:

1. inspect Shipwright state,
2. start intake if needed,
3. delegate to Product Owner,
4. ask questions in chat,
5. write artifacts,
6. advance safe internal phases,
7. stop at approval gates.

## 10. Understand the lifecycle

Shipwright moves through phases:

```txt
INTAKE
DISCOVERY
SCOPE_REVIEW
PROJECT_PLANNING
UX_DESIGN / UX_APPROVAL
TECHNICAL_DESIGN
BACKLOG_READY
IMPLEMENTATION
INTEGRATION
QA_SECURITY_REVIEW
TECH_LEAD_REVIEW
USER_ACCEPTANCE
CLOSED
```

Important gates:

```bash
shipwright approve scope
shipwright approve ux-design
shipwright approve technical-plan
shipwright approve tech-lead
shipwright approve final-acceptance
```

Request a change:

```bash
shipwright request-change "Change invoice workflow before implementation"
```

## 11. Contract-first frontend/backend

Validate contract:

```bash
shipwright contract validate
```

Generate frontend/backend tasks:

```bash
shipwright contract generate-tasks
```

Check frontend mocks:

```bash
shipwright contract check-mocks
```

Check backend compliance:

```bash
shipwright contract check-compliance
```

Rules:

- frontend must keep mock mode and real API mode,
- mocks must align with contract,
- backend implements against `contracts/openapi.yaml`,
- contract changes require approval/change request.

## 12. TDD evidence gate

Check TDD policy:

```bash
shipwright tdd status
```

If mode is `strict`, Shipwright blocks implementation → integration until test evidence exists in:

```txt
progress/frontend.md
progress/backend.md
reports/tdd-report.md
```

Example evidence:

```md
## TDD evidence:
Red: failing test added for invoice total calculation.
Green: implementation added.
Refactor: extracted total calculator.
Command: go test ./... PASS.
```

## 13. QA/security review

Start review evidence:

```bash
shipwright review start
```

Check review status:

```bash
shipwright review status
```

Critical findings block user acceptance. Medium findings require explicit decision. Low findings can pass with registration.

## 14. Visual Studio / Hub

The browser Studio/Hub is now a separate project in the Shipwright ecosystem.

Use the core `shipwright` CLI for lifecycle orchestration and executor setup. Use the sibling `shipwright-studio` project for the browser lobby, project rooms, visual assets, and Docker Studio runtime.

```txt
shipwright-ecosystem/
├── shipwright/         # core CLI/harness
└── shipwright-studio/  # visual Studio/Hub
```

See:

```txt
../shipwright-studio/README.md
```
