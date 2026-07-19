# Shipwright + OpenCode Integration

Shipwright can bootstrap a target project so OpenCode can work as an executor while Shipwright remains the lifecycle authority.

## Mental model

```txt
Shipwright      = lifecycle, state machine, gates, roles, contracts, evidence
OpenCode  = AI coding executor that reads Shipwright-generated instructions
```

Do not let OpenCode invent the workflow. It should read Shipwright state and execute the active role.

## Bootstrap a project

From a target project directory:

```bash
shipwright init --executor opencode
```

This creates the normal Shipwright project structure and OpenCode-supported files:

```txt
AGENTS.md
.opencode/
  opencode.json
  agents/
    product-owner.md
    project-manager.md
    technical-lead.md
    ui-ux-designer.md
    frontend-engineer.md
    backend-engineer.md
    qa-security-reviewer.md
  commands/
    shipwright-status.md
    shipwright-active-agent.md
    shipwright-next.md
    shipwright-doctor.md
  skills/
    _shared/agent-common.md
    product-owner/SKILL.md
    project-manager/SKILL.md
    technical-lead/SKILL.md
    ui-ux-designer/SKILL.md
    frontend-engineer/SKILL.md
    backend-engineer/SKILL.md
    qa-security-reviewer/SKILL.md
```

## Expected OpenCode usage

After initialization, the normal user flow happens inside OpenCode:

```bash
opencode
```

Then the user can type a product request directly:

```txt
Crear una plataforma de facturación simple con clientes, productos, facturas en borrador y dashboard básico.
```

Shipwright configures OpenCode with:

- `.harness/project-profile.md` as calibrated repo context;
- `.harness/skill-registry.md` as reusable skill index;
- `.harness/skill-digests.md` as compact role-specific skill rules;
- `default_agent: "shipwright-orchestrator"`
- `AGENTS.md` autopilot instructions
- a project-local CLI wrapper at `.harness/bin/shipwright`
- Shipwright role subagents under `.opencode/agents/`
- slash commands under `.opencode/commands/`

The orchestrator should:

1. Read `.harness/project-profile.md`, `.harness/tdd-policy.md`, `.harness/skill-registry.md`, and `.harness/skill-digests.md` if present.
2. Run `.harness/bin/shipwright status`.
3. If the project is in `INTAKE`, run `.harness/bin/shipwright start "<user request>"`.
4. Read the active agent with `.harness/bin/shipwright agents active`.
5. Delegate to the matching role.
6. Ask the user questions in chat when the role needs more context.
7. Generate artifacts itself through the active role; do not ask the user to manually fill files.
8. Stop at approval gates.

## Add OpenCode to an existing Shipwright project

```bash
shipwright executor generate opencode
shipwright executor status opencode
```

## TDD evidence gate

Before OpenCode advances from implementation to integration, it must run:

```bash
.harness/bin/shipwright tdd status
```

If `.harness/tdd-policy.md` is `strict`, frontend/backend progress or `reports/tdd-report.md` must include executed test evidence. OpenCode must not claim implementation is complete without that evidence.

## OpenCode workflow

Inside OpenCode, use Shipwright-aware commands:

```txt
/shipwright-status
/shipwright-active-agent
/shipwright-doctor
/shipwright-next
```

Recommended flow:

1. `/shipwright-status`
2. `/shipwright-active-agent`
3. Invoke or follow the matching Shipwright role agent.
4. Make only the changes allowed by the active Shipwright phase.
5. Return evidence and the next Shipwright command.

## Generated OpenCode files

### AGENTS.md

Project-level rules. OpenCode reads this as project context.

### .opencode/opencode.json

Project OpenCode config colocated with the OpenCode executor assets. It uses the official schema, `../AGENTS.md` as instruction source, registered Shipwright agents, registered Shipwright commands, permissions, tool access, and default model assignments.

Generated defaults:

```txt
shipwright-orchestrator      anthropic/claude-sonnet-4-20250514
technical-lead         anthropic/claude-sonnet-4-20250514
frontend-engineer      anthropic/claude-sonnet-4-20250514
backend-engineer       anthropic/claude-sonnet-4-20250514
qa-security-reviewer   anthropic/claude-sonnet-4-20250514
product-owner          anthropic/claude-haiku-4-20250514
project-manager        anthropic/claude-haiku-4-20250514
ui-ux-designer         anthropic/claude-haiku-4-20250514
```

Edit `.opencode/opencode.json` if your OpenCode provider uses different model IDs.

### .opencode/agents/*.md

Role-specific OpenCode subagents. Each one points back to the matching Shipwright role skill and harness state.

### .opencode/commands/*.md

Convenience slash commands that ask OpenCode to inspect Shipwright status, active role, doctor output, and next gate safety.

### .opencode/skills/*/SKILL.md

Reusable Shipwright role instructions derived from the existing shipwright agents.


## Model selection

Shipwright does not force one model for every role. OpenCode models are stored in `.harness/config.json` under:

```json
{
  "executors": {
    "opencode": {
      "default_model": "anthropic/claude-sonnet-4-20250514",
      "reasoning_model": "anthropic/claude-sonnet-4-20250514",
      "fast_model": "anthropic/claude-haiku-4-20250514",
      "agent_models": {}
    }
  }
}
```

Reasoning-heavy roles use `reasoning_model` by default:

- `shipwright-orchestrator`
- `technical-lead`
- `frontend-engineer`
- `backend-engineer`
- `qa-security-reviewer`

Fast/documentary roles use `fast_model` by default:

- `product-owner`
- `project-manager`
- `ui-ux-designer`

Override at init time:

```bash
shipwright init --executor opencode \
  --reasoning-model openai/gpt-5.5 \
  --fast-model opencode-go/deepseek-v4-flash
```

Override a specific role:

```bash
shipwright executor generate opencode \
  --agent-model technical-lead=openai/gpt-5.5 \
  --agent-model product-owner=opencode-go/deepseek-v4-flash
```

Or use environment variables for machine-specific overrides:

```bash
export SHIPWRIGHT_OPENCODE_REASONING_MODEL=openai/gpt-5.5
export SHIPWRIGHT_OPENCODE_FAST_MODEL=opencode-go/deepseek-v4-flash
export SHIPWRIGHT_OPENCODE_AGENT_MODELS="technical-lead=openai/gpt-5.5,product-owner=opencode-go/deepseek-v4-flash"
```

After changing models, regenerate OpenCode files:

```bash
shipwright executor generate opencode
```

## OpenPencil MCP

Shipwright can add OpenPencil to `.opencode/opencode.json` under OpenCode's `mcp` config when it detects either:

- the current OpenPencil MCP command: `openpencil-mcp`
- a configured legacy JS server path: `OPENPENCIL_MCP_SERVER=/path/to/mcp-server.cjs`

Recommended setup:

```bash
npm install -g @open-pencil/mcp
which openpencil-mcp
shipwright integrations detect
shipwright integrations enable openpencil
shipwright executor generate opencode
```

The old screenshot-style config using `/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs` only works if that file exists. Some current OpenPencil app bundles do not include that JS file; they expect the npm MCP package instead.

### Validating OpenPencil inside OpenCode

Do not treat `installed_no_active_canvas` as a hard failure. That status means Shipwright found the MCP entry point but cannot verify the live OpenPencil editor from plain CLI detection.

The real validation must happen inside OpenCode:

```bash
opencode mcp list
```

Then ask the `ui-ux-designer` to use the `open-pencil` MCP tools. OpenCode registers MCP tools with the server name as prefix, so the expected tool pattern is:

- `open-pencil_*`

If another MCP server named `pencil` is also connected, do not use it for Shipwright OpenPencil work. It may belong to another desktop host (for example Antigravity) and fail even when `open-pencil` is healthy.

The first tool call should be `open-pencil_get_editor_state`. If that call succeeds, continue with OpenPencil even if `shipwright status` still says `installed_no_active_canvas`.

## Safety rules

- Shipwright is the source of truth.
- OpenCode is an executor, not the PM/orchestrator.
- OpenCode must not self-approve gates.
- OpenCode must not advance phases unless the user explicitly asks and Shipwright gates are satisfied.
- Frontend/backend work must respect contract-first rules.
- QA/security review must include evidence before user acceptance.

## Fallback

If OpenCode is not installed, the project still works with the generic `AGENTS.md` flow:

```bash
shipwright executor generate generic
```

## References

OpenCode supports the generated formats documented by the official docs:

- Project rules: `AGENTS.md`
- Project calibration: `.harness/project-profile.md`
- Skill registry: `.harness/skill-registry.md`
- Skill digests: `.harness/skill-digests.md`
- Project config: `.opencode/opencode.json`
- Project agents: `.opencode/agents/*.md`
- Project commands: `.opencode/commands/*.md`
- Project skills: `.opencode/skills/<name>/SKILL.md`
