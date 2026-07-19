# 07 — Prompt para pedir implementación a un agente

Usá este prompt cuando quieras pedirle a un agente que empiece a implementar el MVP.

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

Reglas obligatorias:

1. No implementes todo el roadmap.
2. Implementá solo Fase 0 y Fase 1.
3. No agregues integraciones reales con OpenPencil, Engram, Jira, GitHub Issues ni CI todavía.
4. Usá almacenamiento local con Markdown y JSON.
5. El harness debe impedir avanzar de fase si faltan artefactos o approvals.
6. Todo avance debe registrarse en progress/history.md.
7. No declares nada como terminado sin evidencia.
8. Si falta contexto, no inventes: registrá una pregunta o bloqueo.

MVP a implementar:

- shipwright init
- shipwright start "<request>"
- shipwright status
- shipwright next
- shipwright approve <gate>
- shipwright request-change

Estructura esperada:

.harness/
  state.json
  agents/
  approvals/
  gates.json
  templates/

product/
project/
design/
architecture/
contracts/
backlog/
sdd/
knowledge/
progress/
reports/

Criterios de aceptación:

- Puedo inicializar un proyecto vacío.
- Puedo registrar una petición inicial.
- El sistema entra en DISCOVERY, no en implementación.
- El sistema crea artefactos base.
- El sistema muestra estado actual.
- El sistema bloquea gates sin approvals.
- El sistema registra history.
- El diseño permite agregar OpenPencil, Engram y Jira después sin reescribir el núcleo.

Primero explorá el repositorio y proponé el plan de implementación incremental. Después implementá solo la primera tanda.
```
