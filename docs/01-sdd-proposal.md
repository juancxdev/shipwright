# 01 — SDD Proposal: Agentic Software Delivery Harness MVP

## Change name

`agentic-sdlc-harness-mvp`

## Intent

Construir un MVP de harness que permita orquestar agentes de IA mediante un ciclo de vida profesional de software, con roles, artefactos, gates de aprobación, estado persistente y flujo incremental.

## Why now

El desarrollo asistido por IA falla cuando se usa como chatbot o generador directo de código. La oportunidad está en envolver al modelo con un harness que controle contexto, fases, responsabilidades, verificación y memoria.

## Scope MVP

### In scope

- Inicializar estructura `.harness/` en un proyecto.
- Mantener estado local del ciclo de vida.
- Definir fases y transiciones.
- Generar documentos base por fase.
- Definir agentes/roles como archivos de instrucciones.
- Soportar gates manuales de aprobación.
- Permitir continuar desde el estado actual.
- Registrar progreso histórico.
- Separar producto, proyecto, arquitectura, contratos, backlog, SDD y progreso.
- Preparar integración futura con OpenPencil y Engram.

### Out of scope inicial

- Integración real con Jira/Linear.
- Integración obligatoria con OpenPencil.
- Ejecución automática de frontend/backend.
- CI/CD completo.
- Multiusuario remoto.
- Base de datos remota.
- Autenticación y permisos enterprise.
- Generación completa de aplicaciones end-to-end.

## Target users

- Desarrolladores que usan agentes IA.
- Tech leads que quieren controlar calidad y contexto.
- Equipos pequeños que necesitan estructura sin burocracia pesada.
- Creadores de herramientas agentic software delivery.

## Main scenario

1. Usuario solicita: “crea un sistema de facturación electrónica”.
2. Harness clasifica la solicitud como intención de producto, no tarea implementable.
3. Product Owner agent genera preguntas de discovery.
4. Usuario responde.
5. Product Owner genera documento de contexto y alcance inicial.
6. Technical Lead genera opciones técnicas, riesgos y alcance técnico.
7. Product Owner presenta alcance al usuario.
8. Usuario aprueba o pide cambios.
9. Project Manager genera PMBOK-lite plan.
10. UI/UX agent se activa si aplica.
11. Technical Lead genera arquitectura, contratos y backlog.
12. Frontend y Backend trabajan en paralelo por contrato.
13. QA/Security revisa.
14. Technical Lead presenta al usuario.
15. Usuario acepta o abre change request.

## Success criteria

- El harness impide implementación antes de aprobación mínima.
- Cada fase produce artefactos verificables.
- El estado puede recuperarse entre sesiones.
- Los roles tienen responsabilidades claras.
- El flujo permite iteración con usuario.
- La arquitectura deja puntos de extensión para Engram, OpenPencil, Jira y CI.

## Risks

| Risk | Impact | Mitigation |
|---|---:|---|
| Sobrecargar el MVP con demasiadas integraciones | Alto | Empezar local-first con archivos |
| Convertir PMBOK en burocracia | Medio | PMBOK-lite y tailoring |
| Agentes inventan contexto | Alto | Gates y preguntas obligatorias |
| Frontend/backend se desalinean | Alto | Contract-first |
| Estado se vuelve inconsistente | Medio | State machine explícita |
| Documentos se vuelven basura | Medio | Ownership y criterios de calidad |

## Non-goals

El MVP no busca reemplazar herramientas de gestión, diseño o memoria. Busca definir el núcleo orquestador y el contrato de lifecycle.
