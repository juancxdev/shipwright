# 05 — Artifacts and Gates

## Estructura propuesta

```txt
.harness/
  state.json
  agents/
  approvals/
  runs/

product/
  discovery.md
  context.md
  scope.md
  open-questions.md
  assumptions.md

project/
  project-charter.md
  stakeholders.md
  project-plan.md
  risk-register.md
  communication-plan.md
  change-management.md
  status-report.md
  acceptance-report.md

design/
  ux-brief.md
  user-flows.md
  wireframes.md
  prototype.md
  design-decisions.md
  design-approval.md
  openpencil/

architecture/
  system-architecture.md
  frontend-architecture.md
  backend-architecture.md
  data-model.md
  security-model.md
  technology-options.md

contracts/
  openapi.yaml
  events.md
  integration-contracts.md

backlog/
  epics.md
  user-stories.md
  frontend-tasks.md
  backend-tasks.md
  qa-tasks.md

sdd/
  proposal.md
  spec.md
  design.md
  tasks.md
  verification.md

knowledge/
  index.md
  domain/
  architecture/
  decisions/

progress/
  current.md
  history.md
  frontend.md
  backend.md
  reviews.md
  decisions.md

reports/
  qa-report.md
  security-review.md
  contract-test-report.md
```

## Gates principales

### Gate 1 — Discovery completo

Requiere:

- `product/context.md`
- `product/assumptions.md`
- `product/open-questions.md` sin preguntas críticas pendientes

### Gate 2 — Scope approval

Requiere:

- `product/scope.md`
- `architecture/technology-options.md`
- explicación del PO al usuario
- aprobación explícita en `.harness/approvals/scope.json`

### Gate 3 — Project planning

Requiere:

- `project/project-charter.md`
- `project/project-plan.md`
- `project/risk-register.md`
- `project/change-management.md`

### Gate 4 — UX approval cuando aplica

Requiere:

- `design/ux-brief.md`
- `design/user-flows.md`
- `design/prototype.md` o `design/wireframes.md`
- `.harness/approvals/ux-design.json`

### Gate 5 — Technical plan approval

Requiere:

- `architecture/system-architecture.md`
- `architecture/frontend-architecture.md` si hay frontend
- `architecture/backend-architecture.md` si hay backend
- `architecture/data-model.md` si hay persistencia
- `contracts/openapi.yaml` si hay API
- `backlog/epics.md`
- `backlog/user-stories.md`
- `backlog/frontend-tasks.md`
- `backlog/backend-tasks.md`

### Gate 6 — Implementation review

Requiere:

- tareas completadas,
- tests/evidencias,
- reportes FE/BE,
- QA/security review.

### Gate 7 — User acceptance

Requiere:

- demo o descripción de entrega,
- `project/acceptance-report.md`,
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
