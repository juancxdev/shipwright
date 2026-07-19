# 00 — Visión del Agentic Software Delivery Harness

## Problema

Las herramientas actuales de IA para desarrollo suelen permitir que una persona escriba:

```txt
crea un sistema de facturación electrónica
```

Y el agente empieza a producir código demasiado pronto.

Eso genera varios problemas:

- alcance ambiguo,
- requisitos inventados,
- arquitectura accidental,
- frontend y backend desalineados,
- documentación posterior y no conductora,
- poca trazabilidad,
- revisiones superficiales,
- pérdida de contexto entre sesiones,
- cambios sin control.

## Hipótesis

Un agente de IA no debería actuar como “programador solitario”, sino como parte de un **sistema de delivery** con roles, fases, artefactos, aprobaciones y memoria.

## Visión

Crear un harness que simule un equipo profesional de software:

- Product Owner,
- Project Manager / Delivery Manager,
- Technical Lead,
- UI/UX Designer,
- Frontend Engineer,
- Backend Engineer,
- QA/Security Reviewer,
- User Acceptance Reviewer.

## Resultado esperado

El sistema debe convertir una petición ambigua en un flujo controlado:

```txt
intención humana
-> entendimiento de producto
-> alcance validado
-> diseño de solución
-> planificación
-> diseño UI/UX si aplica
-> backlog técnico
-> implementación
-> revisión
-> aprobación
-> evolución
```

## Principios

### 1. No hay código sin contexto suficiente

Si el Product Owner detecta ambigüedad, pregunta. No inventa.

### 2. No hay backlog final sin alcance aprobado

El usuario debe aprobar alcance funcional y restricciones relevantes.

### 3. No hay backlog frontend final sin gate UX cuando la UI importa

Si el producto tiene interfaz significativa, el diseño/prototipo debe validarse antes de cerrar backlog detallado.

### 4. Contract-first entre frontend y backend

Frontend y backend trabajan en paralelo a partir de contratos compartidos:

- OpenAPI,
- DTOs,
- eventos,
- modelos de datos,
- reglas de error.

### 5. Los mocks no se eliminan

Frontend debe conservar modo mock y modo HTTP real.

### 6. La memoria no es el estado operativo

Separar:

- estado operativo del run,
- documentos del proyecto,
- decisiones memorables,
- conocimiento reusable,
- evidencias de verificación.

### 7. PMBOK-lite, no burocracia infinita

Usar gobernanza mínima para:

- charter,
- alcance,
- stakeholders,
- riesgos,
- comunicación,
- cambios,
- aceptación.

### 8. El agente demuestra, no declara

No alcanza con “terminé”. Debe existir evidencia:

- tests,
- reportes,
- revisión,
- checklist,
- aceptación.
