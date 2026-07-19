# Shipwright Troubleshooting

## Philosophy

Shipwright must fail closed. If evidence, state, gates, or integrations are uncertain, the harness blocks and asks for explicit repair instead of pretending the project is done.

## state.json problems

### Symptom: `cannot parse .harness/state.json`

Cause: `.harness/state.json` is invalid JSON.

Recovery path:

1. Back up the broken file.
2. Recover a minimal safe state.
3. Inspect the backup before continuing.

Programmatic recovery is available through `RecoverCorruptState(projectName)` in `internal/harness`.

Expected behavior:

- corrupt state is backed up as `.harness/state.json.corrupt.<timestamp>.bak`
- new state starts in `INTAKE`
- status is `blocked`
- transition audit records `recover-state`

### Symptom: invalid phase/status or missing approvals

Shipwright performs semantic repair on load:

- invalid `current_phase` -> `INTAKE`
- invalid `status` -> `ready`
- missing `project_id` -> regenerated
- missing approval gates -> initialized to `false`
- missing timestamps -> initialized

If semantic repair cannot produce a valid state, loading fails.

## Transition audit

Every successful or blocked transition writes JSONL to:

```txt
.harness/transition-audit.jsonl
```

This file is append-only and intended for debugging lifecycle decisions.

Events include:

- `start`
- `next`
- `approve`
- `request-change`
- `recover-state`

## Review gate blocks

### Critical/high findings

Any finding containing `CRITICAL` or `HIGH` in a bullet/list line blocks progress.

Fix:

- return to implementation, or
- create a change request with mitigation plan.

### Medium findings

Medium findings block unless they include an explicit decision marker, for example:

```md
- MEDIUM: Rate limiting missing — Decision: accepted for MVP and tracked for hardening.
```

Accepted decision markers include:

- `Decision:`
- `accepted`
- `approved`
- `deferred`
- `mitigated`
- Spanish equivalents like `decisión`, `aceptado`, `aprobado`, `postergado`, `mitigado`

### Low findings

Low findings do not block if evidence exists and no higher severity blockers remain.

## Evidence requirements

Review reports must include evidence markers:

- `Contract evidence:`
- `Test evidence:`
- `Security evidence:`
- `Command output:`
- `Verification:`

Placeholder text blocks progress. Replace generated placeholders with real evidence.

## Contract-first checks

In `IMPLEMENTATION`, run:

```txt
shipwright contract check-mocks
shipwright contract check-compliance
```

Current limitation: Phase 6/7 checks are MVP-level and mostly document-backed through `progress/*.md` and reports. Real source-code contract tests should be added in future hardening once FE/BE repos exist.

## OpenPencil integration

If OpenPencil is enabled but Shipwright reports `installed_no_active_canvas`, do not assume the desktop app is closed. Shipwright CLI can detect the MCP command/path, but it cannot verify the live canvas without an MCP client handshake.

Correct validation:

```bash
opencode mcp list
```

Then, inside OpenCode, the `ui-ux-designer` should try the `open-pencil` MCP tools before falling back:

1. Try `open-pencil_get_editor_state`.
2. If the editor-state call succeeds, continue with OpenPencil.
3. If a separate MCP server named `pencil` is connected, do not use it for Shipwright OpenPencil work; it can belong to another desktop host such as Antigravity.
4. If no `open-pencil_*` MCP tool is visible, restart OpenCode after regenerating `.opencode/opencode.json`.
5. Only use doc-only mode after the `open-pencil` MCP tool call fails or no `open-pencil_*` tools are registered.

Rules:

- Never read `.pen` files directly with filesystem tools.
- Use OpenPencil MCP tools only.
- Human UX approval is still required.

## Engram integration

If Engram is enabled but unavailable, memory must fall back to local decisions log:

```txt
progress/decisions.md
```

Do not store raw logs or ephemeral progress in Engram. Store decisions, discoveries, bug fixes, patterns, and session summaries.
