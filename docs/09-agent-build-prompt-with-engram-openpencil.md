# 09 — Prompt para implementar MVP con Engram y OpenPencil desde el inicio

```txt
Actuá como Senior Software Architect e implementá el MVP del Agentic Software Delivery Harness siguiendo estos documentos:

- README.md
- 00-vision.md
- 01-sdd-proposal.md
- 02-system-design.md
- 03-state-machine.md
- 04-agent-roles.md
- 05-artifacts-and-gates.md
- 06-implementation-roadmap.md
- 08-engram-openpencil-integration.md

Cambio de alcance respecto al MVP inicial:

El MVP SÍ debe incluir puertos reales para Engram y OpenPencil desde el inicio.

Pero NO debe implementar todavía:

- Jira/Linear/GitHub Issues,
- CI/CD completo,
- Mongo/Postgres,
- generación completa FE/BE,
- multiusuario remoto.

Implementá:

1. CLI local:
   - shipwright init
   - shipwright start "<request>"
   - shipwright status
   - shipwright next
   - shipwright approve <gate>
   - shipwright request-change
   - shipwright integrations status
   - shipwright design start

2. Estado local:
   - .harness/state.json
   - .harness/integrations.json
   - .harness/approvals/
   - progress/history.md
   - progress/current.md

3. Engram integration:
   - Crear MemoryPort.
   - Si Engram MCP está disponible, guardar decisiones importantes.
   - Si no está disponible, fallback a progress/decisions.md.
   - No guardar logs brutos ni estado efímero.

4. OpenPencil integration:
   - Crear DesignPort.
   - Verificar disponibilidad de OpenPencil/canvas.
   - Si está disponible, usarlo para diseño .pen.
   - Si no está disponible, usar modo design-doc-only.
   - No leer archivos .pen con filesystem.

5. Gates:
   - No pasar de DISCOVERY sin contexto.
   - No pasar de SCOPE_REVIEW sin approval.
   - Si requiere UI, no pasar de UX_APPROVAL sin approval.
   - No pasar a IMPLEMENTATION sin technical-plan approval.

6. Reglas:
   - No avanzar por narrativa del agente.
   - Avanzar solo por estado + artefactos + approvals.
   - Todo handoff debe escribirse en archivo.
   - Todo evento relevante debe guardarse en Engram o fallback local.

Primero explorá el repositorio.
Luego proponé una implementación incremental.
Después implementá solo el núcleo CLI + state + integrations status + MemoryPort + DesignPort fallback.
No implementes todavía generación real de aplicaciones frontend/backend.
```
