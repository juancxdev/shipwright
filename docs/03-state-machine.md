# 03 — State Machine

## Estados principales

```txt
INTAKE
DISCOVERY
PRODUCT_CONTEXT_READY
TECHNICAL_SCOPE_DRAFT
SCOPE_REVIEW
SCOPE_APPROVED
PROJECT_PLANNING
UX_DECISION
UX_DESIGN
UX_APPROVAL
TECHNICAL_DESIGN
BACKLOG_READY
IMPLEMENTATION
INTEGRATION
QA_SECURITY_REVIEW
TECH_LEAD_REVIEW
USER_ACCEPTANCE
CLOSED
CHANGE_REQUEST
```

## Transiciones

| From | To | Trigger | Gate |
|---|---|---|---|
| INTAKE | DISCOVERY | Nueva petición | Solicitud registrada |
| DISCOVERY | PRODUCT_CONTEXT_READY | PO sin más dudas | `product/context.md` |
| PRODUCT_CONTEXT_READY | TECHNICAL_SCOPE_DRAFT | TL analiza contexto | `architecture/options.md` |
| TECHNICAL_SCOPE_DRAFT | SCOPE_REVIEW | PO prepara explicación | `product/scope.md` |
| SCOPE_REVIEW | DISCOVERY | Usuario pide cambios | Feedback registrado |
| SCOPE_REVIEW | SCOPE_APPROVED | Usuario aprueba | `approvals/scope.json` |
| SCOPE_APPROVED | PROJECT_PLANNING | PM genera plan | `project/project-plan.md` |
| PROJECT_PLANNING | UX_DECISION | Evaluar necesidad UI | `project/delivery-plan.md` |
| UX_DECISION | UX_DESIGN | Requiere UI | `design/ux-brief.md` |
| UX_DECISION | TECHNICAL_DESIGN | No requiere UI | Decisión registrada |
| UX_DESIGN | UX_APPROVAL | Diseño listo | `design/prototype.md` |
| UX_APPROVAL | UX_DESIGN | Usuario rechaza | Feedback registrado |
| UX_APPROVAL | TECHNICAL_DESIGN | Usuario aprueba | `approvals/ux-design.json` |
| TECHNICAL_DESIGN | BACKLOG_READY | TL crea docs y backlog | Arquitectura + contratos + backlog |
| BACKLOG_READY | IMPLEMENTATION | Gate técnico aprobado | `approvals/technical-plan.json` |
| IMPLEMENTATION | INTEGRATION | FE/BE completan tareas | Evidencia por módulo |
| INTEGRATION | QA_SECURITY_REVIEW | Integración candidata | Tests/contratos |
| QA_SECURITY_REVIEW | IMPLEMENTATION | Fallas críticas | Review report |
| QA_SECURITY_REVIEW | TECH_LEAD_REVIEW | Pasa QA/security | Approval técnico preliminar |
| TECH_LEAD_REVIEW | IMPLEMENTATION | TL rechaza | Feedback técnico |
| TECH_LEAD_REVIEW | USER_ACCEPTANCE | TL aprueba | `approvals/tech-lead.json` |
| USER_ACCEPTANCE | CHANGE_REQUEST | Usuario pide cambios | CR creado |
| USER_ACCEPTANCE | CLOSED | Usuario acepta | `project/acceptance-report.md` |
| CHANGE_REQUEST | DISCOVERY | Cambio grande | Nueva discovery parcial |
| CHANGE_REQUEST | TECHNICAL_DESIGN | Cambio técnico claro | Impact assessment |
| CHANGE_REQUEST | BACKLOG_READY | Cambio menor | Backlog actualizado |

## Estados bloqueantes

El harness debe detenerse y pedir intervención humana en:

- `DISCOVERY` si faltan respuestas.
- `SCOPE_REVIEW` si el usuario no aprobó.
- `UX_APPROVAL` si el usuario no aprobó diseño.
- `USER_ACCEPTANCE` si no hay aceptación final.
- `CHANGE_REQUEST` si no se decide impacto.

## Ejemplo de approval

```json
{
  "approval_id": "scope-001",
  "phase": "SCOPE_REVIEW",
  "approved_by": "user",
  "approved_at": "2026-07-15T00:00:00Z",
  "artifact_refs": [
    "product/scope.md",
    "architecture/options.md"
  ],
  "notes": "Usuario aprueba alcance inicial sin integración con SUNAT real para MVP."
}
```
