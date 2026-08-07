# tomlmin: leer pyproject.toml del mundo real sin abortar

Feature ID: **F-023** (`.agent/feature-list.json`)
Affected tools: **linelens, dupelens, secretlens, testlens, scopelens** (los cinco, vía `tools/_shared/tomlmin`)
Risk: **medium** (toca el parser compartido por los cinco binarios; mitigado por 100% coverage y golden tests contra `pyproject.toml` reales)

## Why

Un usuario reporta que las cinco herramientas abortan con

```
tomlmin: line 7 (in "project"): for key "dependencies": unexpected token in array near ""
```

contra un `pyproject.toml` perfectamente válido, que `tomllib` de la stdlib de Python lee sin
queja. La causa es que `parseDocument` recorre el documento **línea por línea**, y
`dependencies = [` abierto en varias líneas es la forma canónica de declarar dependencias en
cualquier proyecto Python (PEP 621).

Al reproducirlo aparece un problema más ancho: `ExtractAsJSON` parsea **todo** el documento aunque
solo le interese la tabla `[tool.<name>]`, así que cualquier sintaxis fuera del subset de ADR-018
—esté donde esté— aborta la carga entera. Doce casos habituales de un `pyproject.toml` real, nueve
fallan hoy:

| Caso | Dónde aparece típicamente | Hoy |
|---|---|---|
| `dependencies = [` multilínea | `[project]` (PEP 621) | ❌ `unexpected token in array` |
| `exclude = ["a", "b",]` (trailing comma) | cualquier tabla | ❌ `unexpected token in array` |
| `authors = [\n  { name = "x" },\n]` | `[project]` (PEP 621) | ❌ |
| `select = 'E'` (literal string) | `[tool.ruff]` | ❌ `unrecognized value` |
| `description = """…"""` | `[project]` | ❌ `trailing garbage` |
| `black.line-length = 88` (dotted key) | `[tool]` | ❌ error explícito |
| `"module.*" = …` (quoted key) | `[tool.mypy]`, `[tool.poetry.dependencies]` | ❌ error explícito |
| `d = 2026-01-01` (date) | metadata varia | ❌ `invalid number` |
| `x = 0x1f`, `x = 1_000` | poco frecuente | ❌ `trailing garbage` |

El efecto neto es que la promesa de ADR-018 —"config en el archivo central del ecosistema"— no se
cumple en Python: cualquiera que instale la suite en un repo Python real se topa con el fallo de
entrada y debe esquivarlo con un `<tool>.json` dedicado.

## What Changes

Dos correcciones complementarias en `tools/_shared/tomlmin/`. La primera amplía el subset; la
segunda ataca la clase entera de bugs para que ningún `pyproject.toml` ajeno vuelva a romper la
carga.

### 1. Subset ampliado

- **Líneas lógicas**: una asignación se acumula hasta que los delimitadores `[` `]` `{` `}` quedan
  balanceados fuera de strings. Habilita arrays multilínea y arrays de inline tables multilínea.
- **Trailing comma** en arrays (TOML 1.0 la permite; en inline tables no, y se sigue rechazando).
- **Literal strings** `'…'` y strings multilínea `"""…"""` / `'''…'''`.
- **Enteros** con `_` separador y prefijos `0x` / `0o` / `0b`.
- **Fechas y datetimes** RFC 3339 se leen como **string**, sin interpretarlos.
- **Dotted keys** (`a.b = 1`) y **quoted keys** (`"a.b" = 1`, `'a.b' = 1`).

### 2. Tolerancia fuera de la sección objetivo

`ExtractAsJSON(toml, sectionPath)` deja de abortar por sintaxis que no pertenece a `sectionPath`:

- Una línea que no se puede parsear **fuera** de `sectionPath` (y de sus sub-tablas) se **ignora**.
- Una línea que no se puede parsear **dentro** de `sectionPath` sigue siendo **error duro**, con el
  mismo mensaje que hoy: la config del usuario está mal y debe enterarse.

Esta asimetría es deliberada. El subset ampliado reduce los casos ignorados a casi cero, pero la
tolerancia garantiza que un TOML exótico en `[tool.poetry]` o `[build-system]` nunca vuelva a
tumbar un tool que solo quería leer `[tool.linelens]`.

### 3. Documentación y release

- ADR-018 actualizado: su tabla "NO soporta … multi-line strings/arrays, dotted keys" queda
  obsoleta y se reemplaza por el subset ampliado + la regla de tolerancia.
- Los cinco tools publican una versión nueva en npm y PyPI, más el meta-paquete.

## Non-Goals

- **No** se convierte `tomlmin` en un parser TOML 1.0 completo. Sigue siendo un subset; lo que
  cambia es que lo no soportado deja de ser fatal fuera de la sección de interés.
- **No** se agrega una dependencia externa de TOML (ADR-002 intacto).
- **No** cambia la precedencia de la cadena de configuración de ADR-018 ni ningún `Config` struct.

## Impact

- `tools/_shared/tomlmin/`: `document.go`, `assignment.go`, `value.go`, `composite.go` y un módulo
  nuevo para líneas lógicas. Archivos < 100 líneas (ADR-005), 100% coverage (ADR-011).
- `tools/_shared/configload/pyproject.go`: sin cambios de API.
- Los cinco tools: sin cambios de código; heredan el fix por el Go workspace.
- Retrocompatible: todo `pyproject.toml` que hoy carga bien sigue cargando idéntico.
