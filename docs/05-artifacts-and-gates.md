# 05 — Artifacts and Gates

## Estructura propuesta

```txt
.harness/
  state.json
  agents/
  approvals/
  runs/

.harness/artifacts/product/
  discovery.md
  context.md
  scope.md
  open-questions.md
  assumptions.md

.harness/artifacts/project/
  project-charter.md
  stakeholders.md
  project-plan.md
  risk-register.md
  communication-plan.md
  change-management.md
  status-report.md
  acceptance-report.md

.harness/artifacts/design/
  ux-brief.md
  user-flows.md
  wireframes.md
  prototype.md
  design-decisions.md
  design-approval.md
  openpencil/

.harness/artifacts/architecture/
  system-architecture.md
  frontend-architecture.md
  backend-architecture.md
  data-model.md
  security-model.md
  technology-options.md

.harness/artifacts/contracts/
  openapi.yaml
  events.md
  integration-contracts.md

.harness/artifacts/backlog/
  epics.md
  user-stories.md
  frontend-tasks.md
  backend-tasks.md
  qa-tasks.md

.harness/artifacts/sdd/
  proposal.md
  spec.md
  design.md
  tasks.md
  verification.md

.harness/artifacts/knowledge/
  index.md
  domain/
  .harness/artifacts/architecture/
  decisions/

.harness/artifacts/progress/
  current.md
  history.md
  frontend.md
  backend.md
  reviews.md
  decisions.md

.harness/artifacts/reports/
  qa-report.md
  security-review.md
  contract-test-report.md
```

## Gates principales

### Gate 1 — Discovery completo

Requiere:

- `.harness/artifacts/product/context.md`
- `.harness/artifacts/product/assumptions.md`
- `.harness/artifacts/product/open-questions.md` sin preguntas críticas pendientes

### Gate 2 — Scope approval

Requiere:

- `.harness/artifacts/product/scope.md`
- `.harness/artifacts/architecture/technology-options.md`
- explicación del PO al usuario
- aprobación explícita en `.harness/approvals/scope.json`

### Gate 3 — Project planning

Requiere:

- `.harness/artifacts/project/project-charter.md`
- `.harness/artifacts/project/project-plan.md`
- `.harness/artifacts/project/risk-register.md`
- `.harness/artifacts/project/change-management.md`

### Gate 4 — UX approval cuando aplica

Requiere:

- `.harness/artifacts/design/ux-brief.md`
- `.harness/artifacts/design/user-flows.md`
- `.harness/artifacts/design/prototype.md` o `.harness/artifacts/design/wireframes.md`
- `.harness/approvals/ux-design.json`

### Gate 5 — Technical plan approval

Requiere:

- `.harness/artifacts/architecture/system-architecture.md`
- `.harness/artifacts/architecture/frontend-architecture.md` si hay frontend
- `.harness/artifacts/architecture/backend-architecture.md` si hay backend
- `.harness/artifacts/architecture/data-model.md` si hay persistencia
- `.harness/artifacts/contracts/openapi.yaml` si hay API
- `.harness/artifacts/backlog/epics.md`
- `.harness/artifacts/backlog/user-stories.md`
- `.harness/artifacts/backlog/frontend-tasks.md`
- `.harness/artifacts/backlog/backend-tasks.md`

### Gate 6 — Implementation review

Requiere:

- tareas completadas,
- tests/evidencias,
- reportes FE/BE,
- QA/security review.

### Gate 7 — User acceptance

Requiere:

- demo o descripción de entrega,
- `.harness/artifacts/project/acceptance-report.md`,
- aprobación del usuario o change request.

## Change Request Template

```md
# CR-0001 — Título

## Solicitud

## Motivo

## Impacto funcional

## Impacto técnico

## Impacto en alcance

## Impacto en tiempo/esfuerzo

## Riesgos

## Decisión

- [ ] aprobado
- [ ] rechazado
- [ ] postergado

## Aprobado por
```
