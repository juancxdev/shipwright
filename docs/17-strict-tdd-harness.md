# Phase 14 — Strict TDD Harness

Shipwright should not treat implementation as complete just because files exist. If project calibration detects a reliable test command, the harness can enforce a strict TDD/test-evidence gate before integration.

## What this phase adds

Generated policy files:

```txt
.harness/tdd-policy.json
.harness/tdd-policy.md
```

CLI commands:

```bash
shipwright tdd refresh
shipwright tdd status
shipwright tdd policy
```

Lifecycle enforcement:

```txt
IMPLEMENTATION -> INTEGRATION
```

is blocked when:

- `.harness/tdd-policy.md` says mode is `strict`, and
- implementation progress has no TDD/test evidence.

## Policy modes

| Mode | Meaning | Blocks integration? |
|---|---|---|
| `strict` | A reliable test command was detected. Agents must record executed test evidence before integration. | Yes |
| `suggested` | Stack exists, but test command is not reliable enough for enforcement. | No |
| `none` | Greenfield/no test capability detected. | No |

## Evidence sources

Strict mode accepts evidence in either role progress files or the central TDD report:

```txt
progress/frontend.md
progress/backend.md
reports/tdd-report.md
```

Evidence should include markers such as:

```md
## TDD evidence:
Red: failing test added for invoice total calculation.
Green: implementation added.
Refactor: extracted money formatter.
Command: go test ./... PASS.
```

or:

```md
## Test evidence:
Command: pnpm test PASS.
```

When a test command was detected in `.harness/project-profile.md`, agents should use that command in the evidence.

## OpenCode behavior

Generated OpenCode instructions now tell the orchestrator and implementation agents to read:

```txt
.harness/tdd-policy.md
```

Before advancing from implementation to integration, OpenCode should run:

```bash
.harness/bin/shipwright tdd status
```

If the policy is strict and evidence is missing, the agent must not claim the implementation is complete.

## Why this exists

Without this gate, an AI agent can create files and say “done” even when no test was executed. Strict TDD mode turns project calibration into an operational quality gate: detected test capability becomes required evidence.
