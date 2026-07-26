# Shipwright Project Instructions

This project is managed by Shipwright. Shipwright is the source of truth for lifecycle, gates, roles, contracts, reviews, and evidence.

## Mandatory workflow

1. Read `.harness/communication-policy.md`, `.harness/project-profile.md`, `.harness/tdd-policy.md`, `.harness/skill-registry.md`, `.harness/skill-assignments.md`, and `.harness/skill-digests.md` before making technical assumptions or choosing response style.
2. Run `shipwright status` before making changes.
3. Run `shipwright agents active` to identify the active role.
4. Run `shipwright agents run <agent-name>` and follow that role's instructions.
5. Do not advance phases manually; use `shipwright next`.
6. Do not approve gates yourself; user approvals must use `shipwright approve <gate>`.
7. Do not mark work finished without evidence in `.harness/artifacts/reports/` when the phase requires it.
8. If `.harness/tdd-policy.md` says `strict`, record executed test evidence before integration.
9. Run `shipwright doctor` if environment or integration behavior is unclear.

## Contract-first rules

- Backend must implement `.harness/artifacts/contracts/openapi.yaml`.
- Frontend must preserve mock mode and real API mode.
- Contract changes require explicit technical approval/change request.

## Communication policy

- Follow `.harness/communication-policy.md` for response language, tone, and dialect.
- Project-local communication policy overrides global/personal assistant style instructions for this project.

## Safety

- Shipwright orchestrates. The AI coding agent executes.
- If scope, target, or phase is ambiguous, stop and ask. Do not guess.
