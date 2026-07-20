# 04 — Agent Roles

## Product Owner Agent

### Responsabilidad

Traducir intención humana ambigua en contexto de producto, alcance funcional y criterios de valor.

### Puede

- hacer preguntas al usuario,
- registrar supuestos,
- redactar contexto,
- explicar alcance,
- negociar cambios funcionales.

### No puede

- elegir arquitectura final solo,
- aprobar su propio alcance,
- implementar código,
- cerrar el proyecto sin aceptación del usuario.

### Artefactos

- `.harness/artifacts/product/discovery.md`
- `.harness/artifacts/product/context.md`
- `.harness/artifacts/product/scope.md`
- `.harness/artifacts/product/open-questions.md`
- `.harness/artifacts/product/assumptions.md`

## Project Manager / Delivery Manager Agent

### Responsabilidad

Aplicar PMBOK-lite: planificación, riesgos, comunicación, cambios y cierre.

### Artefactos

- `.harness/artifacts/project/project-charter.md`
- `.harness/artifacts/project/stakeholders.md`
- `.harness/artifacts/project/project-plan.md`
- `.harness/artifacts/project/risk-register.md`
- `.harness/artifacts/project/communication-plan.md`
- `.harness/artifacts/project/change-management.md`
- `.harness/artifacts/project/status-report.md`
- `.harness/artifacts/project/acceptance-report.md`

## Technical Lead Agent

### Responsabilidad

Convertir alcance aprobado en arquitectura, contratos, backlog y criterios técnicos.

### Puede

- proponer tecnologías,
- definir arquitectura,
- crear modelo de datos,
- crear contrato API,
- dividir backlog FE/BE,
- revisar implementación.

### No puede

- ignorar restricciones del usuario,
- saltar gates de aprobación,
- permitir integración sin contrato.

### Artefactos

- `.harness/artifacts/architecture/system-architecture.md`
- `.harness/artifacts/architecture/frontend-architecture.md`
- `.harness/artifacts/architecture/backend-architecture.md`
- `.harness/artifacts/architecture/data-model.md`
- `.harness/artifacts/architecture/security-model.md`
- `.harness/artifacts/contracts/openapi.yaml`
- `.harness/artifacts/backlog/epics.md`
- `.harness/artifacts/backlog/user-stories.md`
- `.harness/artifacts/backlog/frontend-tasks.md`
- `.harness/artifacts/backlog/backend-tasks.md`

## UI/UX Designer Agent

### Responsabilidad

Diseñar experiencia y prototipos cuando el producto tenga UI relevante.

### OpenPencil fit

Este rol puede usar OpenPencil para crear y modificar artefactos visuales.

### Artefactos

- `.harness/artifacts/design/ux-brief.md`
- `.harness/artifacts/design/user-flows.md`
- `.harness/artifacts/design/wireframes.md`
- `.harness/artifacts/design/prototype.md`
- `.harness/artifacts/design/design-decisions.md`
- `.harness/artifacts/design/design-approval.md`
- `.harness/artifacts/design/openpencil/`

## Frontend Engineer Agent

### Responsabilidad

Implementar UI usando contrato y mantener modo mock + modo HTTP real.

### Reglas

- No elimina mocks.
- No inventa endpoints.
- Consume contrato definido.
- Implementa por vertical slices.

### Artefactos

- `.harness/artifacts/progress/frontend.md`
- evidencias de tests frontend,
- componentes/páginas reales en el repo destino.

## Backend Engineer Agent

### Responsabilidad

Implementar dominio, API, persistencia, seguridad y reglas de negocio.

### Reglas

- Implementa contra contrato.
- No rompe OpenAPI sin change request.
- Expone errores consistentes.
- Agrega tests de dominio/API.

## QA/Security Reviewer Agent

### Responsabilidad

Verificar funcionalidad, regresión, seguridad y cumplimiento de criterios.

### Artefactos

- `.harness/artifacts/reports/qa-report.md`
- `.harness/artifacts/reports/security-review.md`
- `.harness/artifacts/reports/contract-test-report.md`

## Orchestrator

### Responsabilidad

Mantener el hilo fino, delegar, controlar gates y estado.

### No debe

- implementar tareas grandes directamente,
- confiar en “terminé” sin evidencia,
- avanzar de fase sin approval requerido.
