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
