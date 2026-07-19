# Template Assets

Shipwright separates engine code from project templates.

## Rule

Static project-facing files must live under:

```txt
internal/harness/templates/project/
```

The Go code should orchestrate, validate, copy, and render templates. It should not contain long markdown prompts, agent skills, or generated project documents as large string constants when those files can be maintained as templates.

## Current template groups

```txt
internal/harness/templates/project/harness/agents/
```

This directory contains the agent skill files copied into the user's project during `shipwright init` and OpenCode executor generation.

```txt
internal/harness/templates/project/harness/agents/_shared/agent-common.md
```

This is the shared protocol loaded by all generated agent skills.

## Why

Keeping these files as templates makes Shipwright easier to maintain:

- non-Go contributors can edit prompts and markdown files;
- template diffs are readable in pull requests;
- the CLI binary remains portable because templates are embedded at compile time;
- generated project files behave like copied scaffolding, not hidden Go literals;
- future executors can reuse the same template assets.

## What belongs here

Good candidates:

- agent `SKILL.md` files;
- executor bootstrap markdown;
- scaffolded project documents;
- default configuration templates;
- prompt packs and reusable role instructions.

Bad candidates:

- runtime state;
- user-generated project artifacts;
- compiled assets;
- code that needs tests, branching, or platform-specific logic.

## Implementation note

Templates are embedded with Go `embed.FS`, then written into user projects by the harness engine. This keeps installs simple: the user only needs the Shipwright binary.
