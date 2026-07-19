# 14 — Project Calibration Harness

Phase 11 adds the missing “step zero” before an AI agent starts making technical decisions.

`shipwright init` now calibrates the current repository and persists a project profile that OpenCode agents must read before proposing architecture, tasks, tests, or implementation.

## Why this exists

A harness should not only create folders. It should understand where it is standing.

Without calibration, an agent can easily:

- invent a stack,
- run the wrong package manager,
- skip tests because it does not know the test command,
- assume frontend/backend boundaries that do not exist,
- apply strict TDD where no test runner exists,
- or treat a mature repo like a greenfield project.

The Project Calibration Harness prevents that by turning repo facts into explicit artifacts.

## Generated files

```txt
.harness/project-profile.json   machine-readable calibration
.harness/project-profile.md     human/agent-readable calibration summary
```

OpenCode instructions reference `.harness/project-profile.md`, so agents should use the detected stack and commands before making assumptions.

## What Shipwright detects

Current detection includes:

- languages: JavaScript, TypeScript, Go, Python, Rust, Java, PHP, Dart;
- package managers/build systems: npm, pnpm, yarn, bun, Go modules, Cargo, Maven, Gradle, Composer, Pub;
- frontend/backend hints;
- monorepo hints: `apps/`, `packages/`, Node workspaces, Turbo, Nx;
- repository state hints: Git, CI, Docker, Docker Compose;
- existing delivery artifacts: README, docs, OpenAPI, OpenCode config, AGENTS.md;
- commands:
  - test,
  - build,
  - lint,
  - dev;
- TDD capability:
  - `strict` when a reliable test command exists,
  - `suggested` when stack exists but no test command is detected,
  - `none` for greenfield/no-stack projects.

## Example profile summary

```md
# Project Calibration Profile

**Project:** billing-api
**Existing project:** yes

## Detected stack

- Language: `Go`
- build-system `Go modules` — go.mod

## Commands

- Test: `go test ./...` (go.mod, high)
- Build: `go build ./...` (go.mod, medium)

## TDD capability

- Supported: `yes`
- Recommended mode: `strict`
```

## Agent rules

Agents must:

1. read `.harness/project-profile.md` before technical planning or implementation;
2. prefer detected commands over invented commands;
3. avoid strict TDD unless the profile says test capability is present;
4. ask or create explicit tasks when commands are missing;
5. use warnings as planning input, not as fatal errors by default.

## Limitations

Calibration is intentionally conservative. It detects common project markers but does not execute build/test commands during `init`.

That is deliberate: `init` should be safe and fast. Real execution evidence still belongs to QA/review phases.
