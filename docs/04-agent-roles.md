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

- `product/discovery.md`
- `product/context.md`
- `product/scope.md`
- `product/open-questions.md`
- `product/assumptions.md`

## Project Manager / Delivery Manager Agent

### Responsabilidad

Aplicar PMBOK-lite: planificación, riesgos, comunicación, cambios y cierre.

### Artefactos

- `project/project-charter.md`
- `project/stakeholders.md`
- `project/project-plan.md`
- `project/risk-register.md`
- `project/communication-plan.md`
- `project/change-management.md`
- `project/status-report.md`
- `project/acceptance-report.md`

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

- `architecture/system-architecture.md`
- `architecture/frontend-architecture.md`
- `architecture/backend-architecture.md`
- `architecture/data-model.md`
- `architecture/security-model.md`
- `contracts/openapi.yaml`
- `backlog/epics.md`
- `backlog/user-stories.md`
- `backlog/frontend-tasks.md`
- `backlog/backend-tasks.md`

## UI/UX Designer Agent

### Responsabilidad

Diseñar experiencia y prototipos cuando el producto tenga UI relevante.

### OpenPencil fit

Este rol puede usar OpenPencil para crear y modificar artefactos visuales.

### Artefactos

- `design/ux-brief.md`
- `design/user-flows.md`
- `design/wireframes.md`
- `design/prototype.md`
- `design/design-decisions.md`
- `design/design-approval.md`
- `design/openpencil/`

## Frontend Engineer Agent

### Responsabilidad

Implementar UI usando contrato y mantener modo mock + modo HTTP real.

### Reglas

- No elimina mocks.
- No inventa endpoints.
- Consume contrato definido.
- Implementa por vertical slices.

### Artefactos

- `progress/frontend.md`
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

- `reports/qa-report.md`
- `reports/security-review.md`
- `reports/contract-test-report.md`

## Orchestrator

### Responsabilidad

Mantener el hilo fino, delegar, controlar gates y estado.

### No debe

- implementar tareas grandes directamente,
- confiar en “terminé” sin evidencia,
- avanzar de fase sin approval requerido.
