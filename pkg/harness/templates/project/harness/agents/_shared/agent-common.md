# Shipwright Agent — Common Protocol

Boilerplate identical across all Shipwright shipwright agents. Every agent MUST load this
alongside their agent-specific SKILL.md.

Executor boundary: every Shipwright agent is an EXECUTOR, not an orchestrator. Do the
phase work yourself. Do NOT launch sub-agents, do NOT delegate, and do NOT bounce
work back unless the agent skill explicitly says to stop and report a blocker.

## A. Context Loading

Before starting phase work:

1. Run: shipwright status — read the current phase, active agent, and approvals
2. Run: shipwright agents active — confirm YOU are the active agent for this phase
3. Read .harness/state.json — check requires_ui, active_change_request, block_reason
4. Read progress/current.md — understand what happened before you
5. Read progress/handoffs.md — understand who handed off to you and why

If any of these fail or are missing, proceed with what you have — the harness
guarantees state.json exists (checked at init).

## B. Artifact Reading

Before writing ANY artifact:

1. Check if the artifact already exists (shipwright scaffold may have created a placeholder)
2. If it exists AND has the placeholder banner (> **PLACEHOLDER**), replace it entirely
3. If it exists AND has real content, READ it first and UPDATE rather than overwrite
4. If it doesn't exist, create it with the format specified in your agent skill

CRITICAL: Never destroy existing real content. If a previous agent wrote something
useful, preserve it. Use Read tool before Write/Edit.

## C. Artifact Writing

Every artifact you write MUST:

1. Start with a clear title (# Title)
2. Include the project name where relevant
3. Use concrete content, not placeholders — replace ALL (pending) markers
4. Follow the exact format specified in your agent skill
5. Be written to the path listed in your "Can Modify" section — NEVER write outside it

After writing each artifact, the harness will validate its existence on the next
"shipwright next" or "shipwright run" call. The harness checks file EXISTENCE, not
content quality — content quality is YOUR responsibility.

## D. Handoff Protocol

When you complete your phase work:

1. Write all output artifacts to the paths in "Can Modify"
2. Log your completion by running: shipwright next (or shipwright run)
   - The harness logs the handoff to progress/handoffs.md automatically
   - The harness logs the phase transition to progress/history.md
   - The harness saves memory events (Engram or fallback) automatically
3. If the next phase requires user approval (gate), the harness will STOP and
   display the gate name. Do NOT attempt to approve it yourself.
4. If you are blocked (missing input, unclear requirements), write what you have
   and set block_reason in state.json or report to the orchestrator.

CRITICAL: Every handoff goes to file. No phone-game. The next agent reads
progress/handoffs.md and progress/current.md to understand context — they do
NOT talk to you directly.

## E. Return Envelope

Every agent MUST return a structured envelope to the orchestrator:

- status: success | partial | blocked
- summary: 1-3 sentence summary of what was done
- artifacts: list of artifact paths written
- next_recommended: the next action (shipwright next, shipwright approve <gate>, etc.)
- risks: risks discovered, or "None"
- blocked_reason: (only if status=blocked) what is missing and what is needed

Example:

**Status**: success
**Summary**: Product context written with 5 assumptions and 3 open questions. Scope drafted with 4 in-scope items.
**Artifacts**: product/context.md, product/assumptions.md, product/open-questions.md, product/scope.md
**Next**: shipwright approve scope (user must approve scope)
**Risks**: None

## F. Gate Awareness

The harness has 5 approval gates. Know which ones affect you:

| Gate | Phase | Who approves | Your role |
|------|-------|-------------|-----------|
| scope | SCOPE_REVIEW | User | Present scope, wait for approval |
| ux-design | UX_APPROVAL | User | Present design, wait for approval |
| technical-plan | BACKLOG_READY | User | Present architecture+backlog, wait for approval |
| tech-lead | TECH_LEAD_REVIEW | User (on TL recommendation) | Review implementation, recommend approve/reject |
| final-acceptance | USER_ACCEPTANCE | User | Prepare acceptance report, wait for sign-off |

You CANNOT self-approve any gate. You CANNOT skip a gate. The harness enforces
this mechanically — but you should also respect it conceptually.

## G. Memory Protocol

The harness automatically saves memory events on:
- Gate approvals (decision type)
- Phase transitions to TECHNICAL_DESIGN (architecture type)
- Phase transitions to CLOSED (session_summary type)
- Change requests (discovery type)

If Engram is enabled, events queue in .harness/memory-queue.json for AI sync.
If Engram is disabled, events write to progress/decisions.md (fallback).

You do NOT need to manually save memory — the harness does it. But if you
discover something non-obvious during your work (a risk, a pattern, a gotcha),
you MAY request the orchestrator to save it via the memory system.
