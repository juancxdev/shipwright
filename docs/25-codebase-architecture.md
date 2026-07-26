# Shipwright codebase architecture

Shipwright is moving from a single `internal/app/harness` god package toward a **modular hexagonal architecture with tactical DDD**.

## Problem

`internal/app/harness` accumulated lifecycle, design providers, executor generation, integrations, model policy, skill registry, project calibration, reviews, and templates in one package. That made the project hard to scale because every new feature naturally became another file in the same folder.

## Architecture style

Use DDD only for real domain concepts:

- Lifecycle, Phase, Gate, Approval
- AgentRole
- Evidence
- DesignProvider
- SkillPack
- Executor
- Integration
- ProjectProfile

Do not force DDD onto utility helpers such as formatting, time, path cleaning, or string functions.

## Package rule

`internal/app/harness` is a temporary façade. New reusable logic belongs under `internal/<module>`.

Internal modules follow this shape:

```txt
module/
  domain/       pure rules and concepts
  application/  use cases
  ports/        required interfaces
  adapters/     concrete IO/integrations, when applicable
```

## Current structure

```txt
cmd/                         CLI commands only
internal/app/harness/        CLI-facing application façade
internal/
  agents/
    domain/
  lifecycle/
    domain/
    application/
    ports/
  artifacts/
    domain/
    application/
    ports/
    fsadapter/
  contracts/
    application/
  design/
    domain/
    application/
    ports/
    providers/
      baselineweb/
      opendesign/
      stitch/
      openpencil/
      doconly/
  skillpacks/
    domain/
    application/
    sources/
      bundled/
      local/
      remote/
  executors/
    domain/
    application/
    opencode/
  integrations/
    domain/
    application/
    adapters/
      engram/
      opendesign/
      openpencil/
      stitch/
  config/
    application/
  platform/
    application/
  doctor/
    application/
  templates/
    application/
  projectprofile/
    domain/
    application/
    detectors/
  review/
    domain/
    application/
  skills/
    application/
  tdd/
    application/
  modelpolicy/
```

## Enforced boundary

Internal packages must not import `shipwright/internal/app/harness`. This prevents cycles and keeps the façade from leaking inward.

The rule is enforced by:

```txt
internal_architecture_test.go
```

## Migration strategy

1. Extract pure domain and application code first.
2. Keep current public names in `internal/app/harness` as compatibility wrappers while the CLI still imports the façade.
3. Move filesystem/MCP/OpenCode/Engram adapters after domain rules are stable.
4. Use package-level tests for pure domain/application behavior.
5. Keep `cmd/*` as CLI-only orchestration; it should call façade/application services, not implement business rules.


## Current façade thinning status

The following implementation areas have already been moved out of `internal/app/harness` and now expose compatibility wrappers only:

- `agents`
- `templates`
- `projectprofile`
- `platform`
- `config`
- `config validation`
- `integrations detection/state/health`
- `doctor reports/fixes`
- `skills registry/assignments/digests/packs`
- `contracts parsing, task generation, mock/backend compliance`
- `TDD policy and evidence assessment`

Remaining large legacy areas in `internal/app/harness` should be migrated next in bounded-context slices: OpenCode executor generation, lifecycle orchestration state, design evidence/provider adapters still living in the façade, and memory adapters.


## Public package rule

Do not create `pkg/common`. If code is shared only inside Shipwright, place it under `internal/<bounded-context>` or `internal/shared` when it is genuinely cross-cutting. Do not add `pkg/` unless Shipwright intentionally exposes a stable Go library API. Today this is a CLI/control-plane, so `pkg/` should remain absent.
