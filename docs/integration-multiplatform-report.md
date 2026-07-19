# Shipwright — Reporte de Integraciones Multiplataforma

## Fecha

2026-07-16

## Objetivo del análisis

Revisar si Shipwright puede detectar y operar correctamente sus integraciones principales — **Engram** y **OpenPencil** — en distintos escenarios:

- macOS con ambas integraciones instaladas.
- macOS con solo una integración instalada.
- Windows sin integraciones.
- Windows con una integración disponible.
- Linux sin integraciones.
- Linux con una integración disponible.
- Entornos CI/headless.

## Resumen ejecutivo

Shipwright ya tiene una buena base de arquitectura:

- `Integrations` model.
- `Engram` enabled/disabled.
- `OpenPencil` enabled/disabled.
- fallback local para memoria.
- fallback `doc-only` para diseño.
- `shipwright integrations status`.
- `shipwright integrations detect`.

Pero la detección actual **NO es multiplataforma todavía**.

El mayor riesgo actual es que Shipwright detecta integraciones con supuestos locales de la máquina del desarrollador:

- `detectEngram()` siempre retorna `true`.
- `detectOpenPencil()` solo revisa `/Applications/OpenPencil.app`.
- `detectOpenPencilCanvas()` siempre retorna `false`.
- No se detecta sistema operativo.
- No se detectan binarios en `PATH`.
- No se detectan MCP servers configurados.
- No se distingue entre instalado, configurado, disponible y conectado.

En criollo: **en tu Mac puede parecer que todo está bien, pero en Windows/Linux Shipwright puede mentir o degradar mal**.

## Estado actual verificado

### Archivos relevantes

- `cmd/integrations.go`
- `internal/harness/integrations.go`
- `internal/harness/memory.go`
- `internal/harness/engram_adapter.go`
- `internal/harness/design.go`
- `internal/harness/openpencil_adapter.go`
- `internal/harness/doc_only_adapter.go`
- `docs/troubleshooting.md`

### Comportamiento actual

#### Engram

```go
func detectEngram() bool {
    return true
}
```

Problema: esto genera falso positivo en cualquier plataforma. En Windows/Linux sin Engram instalado, `shipwright integrations detect` diría que Engram está disponible.

#### OpenPencil

```go
func detectOpenPencil() bool {
    info, err := os.Stat("/Applications/OpenPencil.app")
    return err == nil && info.IsDir()
}
```

Problema: solo sirve para macOS con instalación estándar en `/Applications`. No sirve para:

- Windows.
- Linux.
- macOS con app en otra ruta.
- OpenPencil portable.
- OpenPencil instalado pero no configurado como MCP.

#### Canvas activo

```go
func detectOpenPencilCanvas() bool {
    return false
}
```

Problema: nunca detecta canvas activo. Aunque OpenPencil esté funcionando, Shipwright siempre cae en `installed_no_active_canvas`.

## Problema conceptual principal

Shipwright mezcla cuatro conceptos que deberían estar separados:

| Concepto | Pregunta | Ejemplo |
|---|---|---|
| Installed | ¿Existe el binario/app? | `engram` en PATH, OpenPencil app instalada |
| Configured | ¿Está configurado para este agente/MCP? | config Codex/Claude/OpenCode contiene MCP |
| Available | ¿Se puede ejecutar/conectar ahora? | health check responde, MCP server inicia |
| Active | ¿Hay sesión/canvas/contexto operativo? | OpenPencil tiene canvas activo |

Hoy Shipwright básicamente trata “detectado” como “usable”. Eso no alcanza.

## Escenarios de riesgo

### Escenario 1 — Windows sin Engram ni OpenPencil

Resultado esperado:

- Engram: disabled/unavailable.
- Memory: fallback local a `progress/decisions.md`.
- OpenPencil: unavailable.
- Design: fallback `doc-only`.
- El harness debe seguir funcionando.

Riesgo actual:

- `detectEngram()` retorna `true`, falso positivo.
- OpenPencil no se detecta porque solo revisa `/Applications`.

### Escenario 2 — Linux con Engram en PATH, sin OpenPencil

Resultado esperado:

- Engram: installed/available si `engram` responde.
- OpenPencil: unavailable.
- Design: doc-only.

Riesgo actual:

- Shipwright no verifica `PATH` ni health check.
- No distingue entre Engram instalado y Engram operativo.

### Escenario 3 — macOS con OpenPencil instalado pero sin canvas activo

Resultado esperado:

- OpenPencil: installed.
- Canvas: unavailable.
- Design: fallback doc-only o task file para conectar OpenPencil.

Riesgo actual:

- Detecta app si está en `/Applications`, pero canvas siempre false.
- No intenta MCP handshake real.

### Escenario 4 — macOS con OpenPencil instalado en ruta personalizada

Resultado esperado:

- Permitir ruta configurable.

Riesgo actual:

- Falso negativo porque solo revisa `/Applications/OpenPencil.app`.

### Escenario 5 — CI/headless

Resultado esperado:

- No intentar abrir GUI.
- Engram opcional.
- OpenPencil doc-only.
- Comandos deben ser deterministas.

Riesgo actual:

- No hay modo CI explícito.
- No hay reporte de capacidades por plataforma.

## Qué falta implementar

## 1. Detector multiplataforma de sistema

Crear un detector interno:

```txt
internal/harness/platform.go
```

Debe exponer:

```go
type PlatformInfo struct {
    OS          string // darwin, windows, linux
    Arch        string
    IsCI        bool
    HomeDir     string
    PathEntries []string
}
```

Usar:

- `runtime.GOOS`
- `runtime.GOARCH`
- `os.UserHomeDir()`
- env vars típicas de CI: `CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, etc.

## 2. Detector real de Engram

Reemplazar:

```go
func detectEngram() bool { return true }
```

Por un detector con estados:

```go
type ToolDetection struct {
    Name       string
    Installed  bool
    Configured bool
    Available  bool
    Version    string
    Path       string
    Status     string
    Reason     string
}
```

Estrategia:

1. Buscar binario en `PATH`:
   - `exec.LookPath("engram")`
   - en Windows también considerar `engram.exe`.
2. Si existe, intentar:
   - `engram --version`
   - o comando equivalente si existe.
3. Si Engram usa server local, agregar health check configurable:
   - `http://localhost:7437/health`
4. Si no existe, fallback local.

Estados sugeridos:

```txt
not_installed
installed_not_running
available
configured_unverified
unavailable_fallback_local
```

## 3. Detector real de OpenPencil

No debe depender solo de `/Applications`.

Estrategia por plataforma:

### macOS

Buscar:

- `/Applications/OpenPencil.app`
- `$HOME/Applications/OpenPencil.app`
- ruta configurable en env:
  - `OPENPENCIL_APP_PATH`
  - `OPENPENCIL_MCP_SERVER`

### Windows

Buscar ruta configurable:

- `OPENPENCIL_MCP_SERVER`
- `OPENPENCIL_APP_PATH`

Posibles ubicaciones si aplica:

- `%LOCALAPPDATA%`
- `%PROGRAMFILES%`

Pero NO hardcodear demasiado sin confirmar instalador real.

### Linux

Buscar ruta configurable:

- `OPENPENCIL_MCP_SERVER`
- AppImage/binario si existe.

### MCP server

El detector debería priorizar el MCP server path:

```txt
/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs
```

Pero de forma configurable.

Estados sugeridos:

```txt
not_installed
mcp_server_found
mcp_server_missing
app_found_no_mcp
available_no_canvas
available_with_canvas
unavailable_doc_only
```

## 4. Configuración portable de integraciones

Ampliar `.harness/integrations.json`.

Actual:

```json
{
  "engram": {
    "enabled": false,
    "mode": "mcp",
    "status": "not_configured",
    "fallback": "progress/decisions.md"
  },
  "openpencil": {
    "enabled": false,
    "mode": "mcp",
    "status": "not_configured",
    "fallback": "design-doc-only"
  }
}
```

Propuesto:

```json
{
  "platform": {
    "os": "darwin",
    "arch": "arm64",
    "ci": false
  },
  "engram": {
    "enabled": false,
    "mode": "mcp",
    "status": "not_installed",
    "fallback": "progress/decisions.md",
    "binary_path": "",
    "health_url": "http://localhost:7437/health",
    "version": "",
    "last_detected_at": ""
  },
  "openpencil": {
    "enabled": false,
    "mode": "mcp",
    "status": "not_installed",
    "fallback": "design-doc-only",
    "app_path": "",
    "mcp_server_path": "",
    "canvas_active": false,
    "last_detected_at": ""
  }
}
```

## 5. No habilitar integración solo porque fue detectada

Regla importante:

- `detect` solo detecta.
- `enable` habilita.
- `status` muestra realidad actual.

Pero si el usuario habilita algo que luego no está disponible:

- Engram debe caer a fallback local.
- OpenPencil debe caer a doc-only.
- Status debe decir claramente: enabled but unavailable, fallback active.

## 6. Tests multiplataforma

Agregar tests table-driven para detección.

Ejemplos:

```go
func TestDetectEngramAcrossPlatforms(t *testing.T)
func TestDetectOpenPencilAcrossPlatforms(t *testing.T)
func TestIntegrationFallbackWhenEngramUnavailable(t *testing.T)
func TestIntegrationFallbackWhenOpenPencilUnavailable(t *testing.T)
```

Para que sea testeable, no usar directamente `os.Stat`, `exec.LookPath`, `runtime.GOOS` en funciones rígidas. Crear interfaces:

```go
type SystemProbe interface {
    GOOS() string
    GOARCH() string
    LookPath(binary string) (string, error)
    Stat(path string) (os.FileInfo, error)
    Getenv(key string) string
    UserHomeDir() (string, error)
}
```

Y un `RealSystemProbe` para runtime real.

## 7. Modo doctor

Recomendación fuerte:

```txt
shipwright doctor
```

Debe revisar:

- state.json válido.
- integrations.json válido.
- Engram instalado/configurado/disponible.
- OpenPencil instalado/configurado/disponible.
- fallback activo si corresponde.
- permisos de escritura.
- estructura de carpetas.
- reportes/gates críticos.

Esto no es feature grande si se limita a diagnóstico. Es hardening real.

## 8. Documentación de instalación por plataforma

Agregar:

```txt
docs/integrations.md
docs/install-macos.md
docs/install-windows.md
docs/install-linux.md
```

Mínimo deben explicar:

- qué pasa si no instalo Engram,
- qué pasa si no instalo OpenPencil,
- cómo configurar rutas manualmente,
- cómo correr en modo fallback,
- cómo verificar con `shipwright integrations detect/status`.

## Recomendación de roadmap

### Fase 8.1 — Integration hardening

- Crear `PlatformInfo`.
- Crear `SystemProbe`.
- Reemplazar detectores hardcodeados.
- Agregar tests table-driven.

### Fase 8.2 — Portable config

- Extender `.harness/integrations.json`.
- Agregar rutas configurables.
- Agregar env vars:
  - `ENGRAM_BINARY`
  - `ENGRAM_HEALTH_URL`
  - `OPENPENCIL_APP_PATH`
  - `OPENPENCIL_MCP_SERVER`

### Fase 8.3 — Doctor command

- `shipwright doctor`
- salida clara por plataforma.
- recomendaciones accionables.

### Fase 8.4 — Docs multiplataforma

- documentación macOS/Windows/Linux.
- troubleshooting de integraciones.

## Prioridad de fixes

| Prioridad | Item | Por qué |
|---|---|---|
| P0 | `detectEngram()` no puede retornar siempre true | Falso positivo crítico |
| P0 | `detectOpenPencil()` no puede depender solo de `/Applications` | No multiplataforma |
| P0 | Estados: installed/configured/available/active separados | Evita diagnósticos falsos |
| P1 | SystemProbe testeable | Permite tests Windows/Linux desde macOS |
| P1 | Env vars para rutas | Permite instalaciones custom |
| P1 | Doctor command | Diagnóstico claro para usuarios |
| P2 | Docs por plataforma | Reduce soporte manual |
| P2 | Health check real de Engram | Confirma disponibilidad, no solo instalación |
| P2 | MCP handshake real de OpenPencil | Confirma canvas/conexión real |

## Conclusión

Shipwright ya tiene fallback conceptual correcto:

- sin Engram -> `progress/decisions.md`
- sin OpenPencil -> `design-doc-only`

Pero todavía le falta **detección portable y confiable**.

La próxima implementación NO debería agregar más fases de delivery. Debería enfocarse en:

```txt
PlatformInfo + SystemProbe + DetectionResult + Doctor
```

Eso haría que Shipwright deje de depender de “mi Mac tiene todo instalado” y pueda correr con dignidad en Windows, Linux, macOS y CI.

---

## Actualización — Fase 8.1 implementada

Se implementó la primera parte del hardening multiplataforma.

### Implementado

- `internal/harness/platform.go`
  - `PlatformInfo`
  - `SystemProbe`
  - `RealSystemProbe`
  - detección de OS/arch/CI/home/PATH

- `internal/harness/integration_detection.go`
  - `DetectionResult`
  - `DetectEngram(probe)`
  - `DetectOpenPencil(probe)`
  - detección por `PATH`
  - detección por variables de entorno
  - candidatos de OpenPencil por plataforma

- `.harness/integrations.json` extendido mediante `Integrations.ApplyDetection(...)`
  - platform metadata
  - binary path
  - MCP server path
  - reason
  - last_detected_at

- `cmd/integrations.go`
  - `shipwright integrations detect` ya no usa `detectEngram() true`
  - ya no depende solo de `/Applications/OpenPencil.app`
  - muestra reason/path/status/fallback

- Tests table-driven multiplataforma:
  - Linux sin Engram
  - Linux con Engram en PATH
  - Windows con `engram.exe`
  - env override de Engram
  - macOS OpenPencil MCP path
  - Linux OpenPencil MCP path
  - Windows `OPENPENCIL_MCP_SERVER`
  - fallback cuando OpenPencil no está

### Variables soportadas

```txt
ENGRAM_BINARY
OPENPENCIL_APP_PATH
OPENPENCIL_MCP_SERVER
OPENPENCIL_CANVAS_ACTIVE
```

`OPENPENCIL_CANVAS_ACTIVE` es una señal temporal/testeable para representar canvas activo. En una fase posterior debería reemplazarse por un handshake MCP real.

### Pendiente

- Fase 8.2: portable config más completa.
- Fase 8.3: `shipwright doctor`.
- Health check real de Engram (`ENGRAM_HEALTH_URL`).
- Handshake real con OpenPencil MCP/canvas.
- Documentación específica de instalación macOS/Windows/Linux.

### Ajuste adicional de Fase 8.1

Se agregó separación explícita de tipo de path en `DetectionResult`:

```txt
binary
app
mcp_server
```

Esto evita mezclar el path de la aplicación OpenPencil con el path del MCP server. Por ejemplo, `OPENPENCIL_APP_PATH=/Applications/OpenPencil.app` ahora se guarda como `openpencil.app_path`, mientras que `OPENPENCIL_MCP_SERVER=.../mcp-server.cjs` se guarda como `openpencil.mcp_server_path`.

Este detalle es importante para soporte real multiplataforma porque una instalación puede existir sin que el servidor MCP/canvas esté disponible.

---

## Actualización — Fase 8.2 implementada

Se agregó una capa de configuración portable separada del estado de detección.

### Nuevo modelo

```txt
.harness/config.json        -> configuración deseada y portable
.harness/integrations.json  -> estado detectado / resultado operativo
```

Esto separa dos responsabilidades que antes estaban mezcladas:

- **Config:** qué paths/modos quiero usar en este proyecto.
- **Detection state:** qué encontró realmente Shipwright en esta máquina.

### Archivo portable

`shipwright init` ahora crea:

```txt
.harness/config.json
```

El archivo soporta:

- defaults cross-platform
- overrides por sistema operativo
- variables de entorno con prioridad final

Ejemplo conceptual:

```json
{
  "version": "1",
  "artifact_root": ".",
  "integrations": {
    "engram": {
      "mode": "mcp",
      "binary_path": "",
      "health_url": "http://localhost:7437/health",
      "fallback": "progress/decisions.md"
    },
    "openpencil": {
      "mode": "mcp",
      "app_path": "",
      "mcp_server_path": "",
      "fallback": "design-doc-only"
    }
  },
  "platform_overrides": {
    "windows": {
      "engram": {
        "binary_path": "C:\\Tools\\engram.exe"
      },
      "openpencil": {
        "mcp_server_path": "C:\\Tools\\OpenPencil\\mcp-server.cjs"
      }
    },
    "darwin": {
      "openpencil": {
        "app_path": "/Applications/OpenPencil.app"
      }
    },
    "linux": {
      "openpencil": {
        "mcp_server_path": "/opt/OpenPencil/resources/mcp-server.cjs"
      }
    }
  }
}
```

### Orden de precedencia

```txt
1. Defaults internos
2. .harness/config.json
3. platform_overrides[GOOS]
4. Variables de entorno
```

Las variables de entorno ganan siempre porque son útiles para CI, máquinas efímeras y secretos/config local.

### Nuevos comandos

```txt
shipwright config show
shipwright config init
shipwright config env
```

- `show`: muestra la config efectiva ya resuelta.
- `init`: crea `.harness/config.json` en proyectos antiguos.
- `env`: lista overrides soportados.

### Variables soportadas

```txt
ENGRAM_BINARY
ENGRAM_HEALTH_URL
OPENPENCIL_APP_PATH
OPENPENCIL_MCP_SERVER
OPENPENCIL_CANVAS_ACTIVE
```

### Integración con detection

`shipwright integrations detect` ahora usa la config efectiva antes de detectar:

```txt
config defaults + file + OS override + env -> detection -> integrations.json
```

Así Shipwright puede correr correctamente en:

- macOS con OpenPencil instalado como `.app`
- Windows con binarios en rutas custom
- Linux con instalaciones bajo `/opt` o `$HOME/.local`
- CI sin Engram/OpenPencil, usando fallback explícito

### Tests agregados

- Defaults si `.harness/config.json` no existe.
- Platform override por OS.
- Variables de entorno ganando sobre archivo y override.
- Detectores usando config sin depender de env vars.
- `integrations.ApplyPortableConfig(...)` propagando paths/config.

### Pendiente

- Fase 8.3: `shipwright doctor`.
- Validación semántica fuerte de `.harness/config.json`.
- Health check real de Engram usando `ENGRAM_HEALTH_URL`.
- Handshake real con OpenPencil MCP/canvas.

---

## Actualización — Fase 8.3 implementada

Se agregó `shipwright doctor` como diagnóstico accionable para plataforma, configuración portable e integraciones.

### Comandos nuevos

```txt
shipwright doctor
shipwright doctor --json
```

### Qué diagnostica

- plataforma detectada: OS, arch, CI
- existencia y carga de `.harness/config.json`
- disponibilidad de Engram
- disponibilidad de OpenPencil
- si OpenPencil tiene canvas activo
- fallback que Shipwright va a usar
- acciones recomendadas para corregir configuración

### Severidades

```txt
OK       -> funcionando
INFO     -> no bloquea, fallback esperado o dato informativo
WARNING  -> configuración parcial o rota, pero con fallback viable
ERROR    -> problema bloqueante, como config corrupta
```

### Exit code

- `0`: sin errores bloqueantes.
- `2`: errores bloqueantes encontrados.

Los warnings no bloquean porque Shipwright debe poder operar con fallback documental/local. Esto es intencional: no tener Engram u OpenPencil no debe romper el harness.

### Salida JSON

`shipwright doctor --json` devuelve un reporte machine-readable con:

```txt
platform
config_file
config_exists
config_loaded
engram
openpencil
checks
actions
summary
```

Esto deja preparada la base para CI o para una futura UI.

### Tests agregados

- proyecto viejo sin `.harness/config.json`
- paths configurados pero inexistentes
- integraciones disponibles con OpenPencil canvas activo
- config corrupta como error bloqueante

### Pendiente

- Health check real de Engram.
- Handshake real con OpenPencil MCP/canvas.
- Validación semántica más fuerte de `.harness/config.json`.
- Posible comando `shipwright doctor --fix` para crear config faltante o limpiar paths inválidos.

---

## Actualización — Fase 8.4 implementada

Se agregaron health checks reales al `shipwright doctor`.

### Nuevo modelo

```txt
detection -> existe / está configurado / está disponible
health    -> responde / es usable / falla con detalle
```

Esto evita mezclar dos preguntas distintas:

- **¿Está instalado/configurado?**
- **¿Está vivo y responde?**

### Engram health

Engram ahora se valida usando HTTP contra:

```txt
config.integrations.engram.health_url
ENGRAM_HEALTH_URL
```

Default:

```txt
http://localhost:7437/health
```

El health check usa timeout y considera sano cualquier status HTTP `2xx`.

### OpenPencil health

OpenPencil ahora valida el MCP server de forma segura con:

```txt
node --check <mcp-server.cjs>
```

Importante: NO se ejecuta `require(mcp-server.cjs)` porque eso podría iniciar el servidor MCP, abrir stdio o colgar el proceso. `node --check` valida que el archivo sea legible y parseable por Node sin ejecutarlo.

### Timeout configurable

Se agregó:

```txt
health.timeout_millis
SHIPWRIGHT_HEALTH_TIMEOUT_MS
```

Default:

```txt
1500
```

### Doctor extendido

`shipwright doctor` y `shipwright doctor --json` ahora incluyen:

```txt
engram_health
openpencil_health
```

con:

```txt
checked
healthy
status
endpoint
detail
latency_ms
suggestion
```

### Severidad

Health fallido en Engram/OpenPencil es `WARNING`, no `ERROR`, porque ambas integraciones tienen fallback explícito.

Sigue siendo `ERROR` sólo algo bloqueante para operar el harness, por ejemplo `.harness/config.json` corrupto.

### Tests agregados

- health checks sanos con mocks
- health checks fallidos como warning no bloqueante
- timeout configurable por `SHIPWRIGHT_HEALTH_TIMEOUT_MS`

### Pendiente

- Handshake MCP real a nivel protocolo cuando OpenPencil exponga una forma estable/no interactiva de consultar estado.
- Posible `shipwright doctor --fix`.
- Validación semántica avanzada de `.harness/config.json`.

---

## Actualización — Fase 8.5 implementada

Se agregó validación semántica de `.harness/config.json` y fixes seguros mediante `shipwright doctor --fix`.

### Validación semántica

`shipwright doctor` ahora valida:

- versión de config soportada
- `artifact_root` no vacío
- `health.timeout_millis` positivo
- modos soportados: `mcp` o `disabled`
- fallback de Engram no vacío
- fallback de OpenPencil no vacío
- `engram.health_url` como URL `http://` o `https://`
- paths con caracteres inválidos
- paths relativos como warning

### Severidad

```txt
ERROR   -> config semánticamente inválida y debe corregirse
WARNING -> config riesgosa pero no necesariamente bloqueante
```

Ejemplo: un path relativo es `WARNING`, no `ERROR`, porque puede ser intencional en algún entorno, pero Shipwright recomienda paths absolutos o variables de entorno.

### Nuevo comando

```txt
shipwright doctor --fix
shipwright doctor --json --fix
```

### Qué arregla `--fix`

Fixes seguros:

- crea `.harness/config.json` si falta
- si el JSON está corrupto:
  - crea backup `.harness/config.json.corrupt.<timestamp>.bak`
  - recrea config con defaults portables
- normaliza campos faltantes
- repara versión inválida
- repara timeout inválido
- repara modos inválidos
- repara health URL inválida de Engram
- repara fallbacks vacíos

### Qué NO arregla automáticamente

`--fix` NO borra paths custom como:

```txt
engram.binary_path
openpencil.app_path
openpencil.mcp_server_path
```

Motivo: esos paths expresan intención del usuario o del entorno. Borrarlos automáticamente sería peligroso. `doctor` los reporta y te dice qué acción tomar.

### JSON output

Con:

```txt
shipwright doctor --json --fix
```

la salida incluye:

```txt
fix
report
```

### Tests agregados

- config default válida
- errores semánticos detectados
- overrides Windows parciales válidos
- doctor reportando errores de config
- `doctor --fix` creando config faltante
- `doctor --fix` respaldando config corrupta
- `doctor --fix` normalizando config parcial
- `doctor --fix` reparando versión/modos/health URL inválidos

### Pendiente

- Posible `shipwright config validate` como comando explícito independiente.
- Posible modo `doctor --fix --aggressive` para limpiar paths inválidos, si el usuario lo pide explícitamente.
- Handshake MCP real con OpenPencil cuando exista una API estable no interactiva.

---

## Actualización — Fase 8.6 implementada

Se cerró el hardening final de la serie 8.x.

### Agregado

```txt
shipwright config validate
shipwright config validate --json
```

Este comando permite validar `.harness/config.json` sin correr todo `doctor`.

### Smoke tests

Se agregó un smoke test del CLI que verifica:

- usage/help incluye comandos nuevos
- `shipwright init` crea `.harness/state.json`
- `shipwright init` crea `.harness/config.json`
- `shipwright init` crea `.harness/integrations.json`
- `shipwright config validate` pasa en proyecto fresco
- `shipwright doctor --json` expone `summary` y health checks

### Documentación final

Se agregaron:

```txt
docs/10-platform-setup.md
docs/11-release-hardening-checklist.md
```

### Estado de cierre

La línea de hardening 8.1–8.6 deja listo Shipwright para operar mejor en:

- macOS
- Windows
- Linux
- CI
- máquinas sin Engram
- máquinas sin OpenPencil
- máquinas con config parcial
- máquinas con config corrupta recuperable

### Pendiente futuro

- Handshake real de OpenPencil MCP cuando exista API estable no interactiva.
- Perfil CI más estricto si el usuario quiere bloquear por integraciones faltantes.
- Posible `doctor --fix --aggressive` sólo con confirmación explícita.
