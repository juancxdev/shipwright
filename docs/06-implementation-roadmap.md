# 06 — Implementation Roadmap

## Fase 0 — Document-only MVP

Objetivo: validar el flujo con archivos, sin automatización compleja.

Tareas:

- Crear estructura de carpetas.
- Crear templates de documentos.
- Crear `state.json` manual.
- Definir roles en `.harness/agents/`.
- Definir gates en `.harness/gates.json`.

Resultado:

- El harness puede operarse manualmente con un agente siguiendo instrucciones.

## Fase 1 — CLI local

Comandos propuestos:

```txt
shipwright init
shipwright start "crea un sistema de facturación electrónica"
shipwright status
shipwright next
shipwright approve scope
shipwright approve ux-design
shipwright approve technical-plan
shipwright request-change
```

Tareas:

- Leer/escribir `state.json`.
- Crear artefactos desde templates.
- Validar gates mínimos.
- Registrar history.

## Fase 2 — Agent instructions

Tareas:

- Crear instrucciones formales para PO, PM, TL, UX, FE, BE, QA.
- Crear handoff templates.
- Crear formato de reportes.
- Crear reglas anti “teléfono descompuesto”: todo resultado va a archivo.

## Fase 3 — SDD integration

Tareas:

- Generar `sdd/proposal.md` desde scope aprobado.
- Generar `sdd/spec.md` desde proposal.
- Generar `sdd/design.md` desde arquitectura.
- Generar `sdd/tasks.md` desde backlog.
- Generar `sdd/verification.md` desde QA criteria.

Inspiración: enfoque Gentle-AI donde SDD es opcional para trabajo sustancial, con artifacts persistibles y gates por fase.

## Fase 4 — OpenPencil integration

Tareas:

- Detectar si OpenPencil MCP está disponible.
- Permitir crear/actualizar diseño en `design/openpencil/`.
- Exportar wireframes/prototipos.
- Bloquear backlog frontend final hasta aprobación UX cuando aplica.

## Fase 5 — Engram integration

Tareas:

- Guardar decisiones, bugs, patrones y aprendizajes.
- No guardar logs brutos ni estado operativo efímero.
- Buscar memoria antes de iniciar proyectos similares.

## Fase 6 — External backlog integration

Opciones:

- Jira,
- Linear,
- GitHub Issues,
- archivos locales.

Regla:

El backend de backlog debe ser intercambiable. El harness no debe depender de una sola herramienta.

## Fase 7 — Verification and review hardening

Tareas:

- Checks automatizados.
- Contract tests.
- Security review.
- Review receipts.
- Evidencia adjunta por tarea.

## Recomendación de tecnología inicial

Para MVP:

- Go si querés binario portable tipo Gentle-AI/Engram.
- TypeScript si querés iterar rápido con tooling Node/MCP.
- Python si querés prototipar muy rápido.

Mi recomendación arquitectónica:

- **Go** para CLI final.
- **Markdown/JSON** al inicio.
- **SQLite** en v0.2.

Tradeoff:

| Opción | Pros | Contras |
|---|---|---|
| Go | Portable, robusto, ideal CLI | Más ceremonia inicial |
| TypeScript | Buen ecosistema MCP/web | Dependencias Node |
| Python | Rápido para prototipo | Distribución menos limpia |

## Primera implementación recomendada

Implementar solo:

- `shipwright init`,
- `shipwright start`,
- `shipwright status`,
- `shipwright approve`,
- templates,
- state machine básica.

Nada más. Si esa base funciona, recién ahí se agregan agentes reales.
