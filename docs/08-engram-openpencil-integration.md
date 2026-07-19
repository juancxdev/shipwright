# 08 — Integración real con Engram y OpenPencil

## Estado verificado

### Engram

Engram está disponible como MCP y debe usarse desde el MVP para memoria histórica.

Uso recomendado:

- Guardar decisiones arquitectónicas.
- Guardar descubrimientos no obvios.
- Guardar bugs resueltos.
- Guardar convenciones.
- Guardar resúmenes de sesión.

No usar Engram para:

- logs brutos,
- estado operativo efímero,
- outputs completos de tests,
- artefactos temporales,
- progreso minuto a minuto.

### OpenPencil

OpenPencil está previsto como MCP para `.pen` design files.

Configuración MCP esperada:

```json
{
  "mcpServers": {
    "openpencil": {
      "type": "stdio",
      "command": "node",
      "args": [
        "/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs",
        "--stdio"
      ],
      "env": {}
    }
  }
}
```

Nota: el MCP puede estar instalado pero no conectado si la app/canvas no está activo. El harness debe detectar esta condición y degradar a modo `design-doc-only`.

## Decisión arquitectónica

El MVP ya debe incluir puertos de integración para Engram y OpenPencil, pero con comportamiento seguro:

```txt
Engram disponible -> guardar memoria histórica real
Engram no disponible -> escribir fallback en progress/decisions.md

OpenPencil disponible -> crear/editar/probar diseño .pen
OpenPencil no disponible -> crear documentos UX Markdown y marcar bloqueo opcional
```

## Puerto Engram

### Interface conceptual

```txt
MemoryPort
  save_decision(title, content, topic_key)
  save_discovery(title, content, topic_key)
  save_bugfix(title, content, topic_key)
  search(query)
  session_summary(content)
```

### Eventos que deben guardar memoria

| Evento | Engram type | Topic key sugerido |
|---|---|---|
| Alcance aprobado | decision | project/scope |
| Tecnología elegida | decision | architecture/technology-stack |
| Arquitectura definida | architecture | architecture/system |
| Contrato API definido | architecture | architecture/api-contract |
| Diseño UX aprobado | decision | design/ux-approval |
| Riesgo crítico descubierto | discovery | project/risks |
| Bug resuelto | bugfix | bugfix/<area> |
| Convención establecida | pattern | conventions/<area> |
| Cierre de sesión | session_summary | session |

### Formato obligatorio

```txt
**What**: ...
**Why**: ...
**Where**: ...
**Learned**: ...
```

## Puerto OpenPencil

### Interface conceptual

```txt
DesignPort
  get_state()
  create_design(file_path, brief)
  update_design(file_path, instructions)
  validate_layout(file_path)
  export_design(file_path, output_dir)
```

### Flujo UX con OpenPencil

```txt
UX_DECISION
  -> si requiere UI:
      verificar OpenPencil
      crear design/ux-brief.md
      crear/actualizar design/openpencil/app.pen
      exportar wireframes a design/openpencil/exports/
      crear design/design-approval.md
      pedir aprobación humana
```

### Gate UX actualizado

Para avanzar de `UX_APPROVAL` a `TECHNICAL_DESIGN` se requiere:

- `design/ux-brief.md`
- `design/user-flows.md`
- `design/design-decisions.md`
- uno de:
  - `design/openpencil/app.pen` + exports,
  - `design/wireframes.md` si OpenPencil no está disponible,
- `.harness/approvals/ux-design.json`

## Modo degradado

Si OpenPencil no conecta:

1. No se debe bloquear todo el proyecto automáticamente.
2. El UX Designer debe crear `design/wireframes.md` y `design/prototype.md` en Markdown.
3. El harness debe registrar en `progress/current.md`:

```txt
OpenPencil unavailable: design generated in doc-only mode.
```

4. El usuario puede decidir si acepta seguir o espera diseño visual real.

## Reglas de seguridad

- No leer archivos `.pen` directamente con herramientas de filesystem.
- Los `.pen` se manipulan solo mediante MCP OpenPencil.
- No considerar un diseño aprobado solo porque existe un `.pen`.
- La aprobación humana sigue siendo obligatoria.

## Cambios al MVP

La Fase 1 ahora debe incluir:

- `MemoryPort` con implementación Engram real y fallback local.
- `DesignPort` con detección OpenPencil y fallback doc-only.
- Comando `shipwright integrations status`.
- Comando `shipwright design start`.
- Registro de disponibilidad en `.harness/integrations.json`.

Ejemplo:

```json
{
  "engram": {
    "enabled": true,
    "mode": "mcp"
  },
  "openpencil": {
    "enabled": true,
    "mode": "mcp",
    "status": "unavailable_no_active_canvas",
    "fallback": "design-doc-only"
  }
}
```
