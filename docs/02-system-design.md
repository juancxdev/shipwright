# 02 — Diseño Técnico del Harness

## Arquitectura general

```txt
CLI / Agent Entry Point
        |
        v
Harness Orchestrator
        |
        +-- State Store
        +-- Artifact Store
        +-- Agent Registry
        +-- Gate Engine
        +-- Progress Log
        +-- Integration Ports
```

## Componentes

### 1. Harness Orchestrator

Responsable de:

- leer estado actual,
- decidir siguiente fase,
- invocar rol/agente correcto,
- validar precondiciones,
- crear artefactos,
- bloquear avances sin aprobación,
- registrar progreso.

### 2. State Store

MVP: JSON local.

```txt
.harness/state.json
```

Ejemplo:

```json
{
  "project_id": "demo-billing-system",
  "current_phase": "DISCOVERY",
  "status": "waiting_user_input",
  "approvals": {
    "scope": false,
    "ux_design": false,
    "technical_plan": false,
    "final_acceptance": false
  },
  "active_change_request": null,
  "created_at": "2026-07-15T00:00:00Z",
  "updated_at": "2026-07-15T00:00:00Z"
}
```

### 3. Artifact Store

MVP: carpetas Markdown/JSON dentro del repo.

```txt
product/
project/
design/
architecture/
contracts/
backlog/
sdd/
progress/
knowledge/
```

### 4. Agent Registry

Define roles disponibles y sus responsabilidades.

```txt
.harness/agents/
  product-owner.md
  project-manager.md
  technical-lead.md
  ui-ux-designer.md
  frontend-engineer.md
  backend-engineer.md
  qa-security-reviewer.md
```

### 5. Gate Engine

Valida si una fase puede avanzar.

Ejemplo:

```txt
BACKLOG_READY requires:
- product/scope.md exists
- project/project-plan.md exists
- architecture/system-architecture.md exists
- contracts/openapi.yaml exists when API exists
- design/design-approval.md exists when UI is required
```

### 6. Progress Log

Registra eventos del ciclo.

```txt
progress/current.md
progress/history.md
progress/reviews.md
progress/decisions.md
```

### 7. Integration Ports

Interfaces futuras:

```txt
integrations/
  engram.md
  openpencil.md
  jira.md
  github.md
  ci.md
```

## Persistencia recomendada

### MVP

- JSON para estado.
- Markdown para documentos.
- Logs en Markdown.

### v0.2

- SQLite local-first para runs, eventos, approvals y artefactos indexados.

### v0.3+

- Engram para memoria histórica.
- OKF para knowledge base.
- Jira/Linear/GitHub Issues para backlog externo.
- OpenPencil para diseño UI/UX.

## Separación de responsabilidades de almacenamiento

| Tipo | Storage MVP | Futuro |
|---|---|---|
| Estado actual | `.harness/state.json` | SQLite |
| Documentos del proyecto | Markdown | Markdown + DB index |
| Memoria histórica | `progress/decisions.md` | Engram |
| Knowledge reusable | `knowledge/*.md` | OKF |
| Backlog | Markdown | Jira/Linear/GitHub Issues |
| Diseño UI/UX | Markdown + exports | OpenPencil |
| Evidencias | `reports/` | CI artifacts |

## Regla crítica

El orchestrator no debe confiar en narrativa del agente. Debe confiar en:

- archivos existentes,
- estado explícito,
- approvals,
- checks,
- evidencias.
