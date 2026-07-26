# Shipwright internal architecture

`internal/app/harness` is the CLI-facing application façade used by `cmd/*` and existing tests; `internal/` is the real control-plane core. Shipwright currently exposes no public Go package API.
New domain logic should live in `internal/<bounded-context>` instead of being added directly to the façade.

## Architecture style

Shipwright uses **modular hexagonal architecture with tactical DDD**:

- `domain/` — pure concepts and rules.
- `application/` — use cases that coordinate domain objects through ports.
- `ports/` — interfaces required by application services.
- `adapters/`, `providers/`, `sources/` — concrete integrations such as filesystem, OpenCode, MCP providers, and skill sources.

## Current modules

```txt
internal/
  agents/             agent role definitions and phase-to-agent mapping
  lifecycle/          phases, gates, approvals, transitions
  artifacts/          artifact layout plus filesystem adapter
  contracts/          OpenAPI parsing, FE/BE task generation, mock/backend compliance
  design/             design domain, evidence gates, provider ports/providers
  skillpacks/         skill pack domain, install/update use-case boundary, sources
  skills/             bundled/project skill registry, assignment, digest, and import use cases
  executors/          executor domain/application and OpenCode adapter boundary
  integrations/       integration detection, health, state, and provider adapters
  config/             portable config defaults, env merge, validation
  platform/           OS/arch/system probing
  doctor/             doctor reports and fix use cases
  templates/          artifact scaffold templates
  projectprofile/     project profile domain/application/detectors
  review/             review findings and assessment boundary
  tdd/                TDD policy and evidence assessment
  modelpolicy/        model-tier policy and agent-to-model resolution
```

## Compatibility rule

`internal/app/harness` may re-export internal domain/application types to keep CLI migration incremental, but internal packages must never import `shipwright/internal/app/harness`.

This rule is enforced by `internal_architecture_test.go` at the module root.

## Migration rule

When moving a subsystem:

1. Move pure domain types/rules first.
2. Add tests in the internal domain/application package.
3. Keep a façade wrapper in `internal/app/harness`.
4. Move side effects/adapters last.
5. Run `go test ./...` after each bounded extraction.


## Public package rule

Do not create `pkg/common`. If code is shared only inside Shipwright, place it under `internal/<bounded-context>` or `internal/shared` when it is genuinely cross-cutting. Do not add `pkg/` unless Shipwright intentionally exposes a stable Go library API. Today this is a CLI/control-plane, so `pkg/` should remain absent.
