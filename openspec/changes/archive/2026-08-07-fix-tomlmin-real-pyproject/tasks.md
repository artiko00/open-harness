# Tasks: fix-tomlmin-real-pyproject

Rojo → verde → refactor (ADR-013). 100% coverage en `tomlmin` (ADR-011), archivos < 100 líneas (ADR-005).

## 1. Banco de pruebas de regresión

- [x] 1.1 [red] `testdata/`: cuatro `pyproject.toml` de referencia con estilos reales (poetry, setuptools, hatch, uv), cada uno con su `[tool.linelens]` válido
- [x] 1.2 [red] Test golden: extraer `tool.linelens` de los cuatro no produce error y devuelve la config esperada
- [x] 1.3 [red] Test de no-regresión: los casos ya soportados hoy (arrays de una línea, inline tables, `[[array of tables]]`, comentarios) siguen dando el mismo resultado

## 2. Líneas lógicas (`scan.go`)

- [x] 2.1 [red] Test: `dependencies = [` abierto en 4 líneas se agrupa en una sola línea lógica
- [x] 2.2 [red] Test: el número de línea reportado en los errores es el de la **primera** línea física de la asignación
- [x] 2.3 [red] Test: `#` dentro de `"…"`, `'…'` y `"""…"""` NO se trata como comentario; fuera sí, incluso dentro de un array multilínea
- [x] 2.4 [red] Test: un `]` o `[tool.x]` dentro de un string multilínea no abre ni cierra nada
- [x] 2.5 [red] Test: delimitador sin cerrar al final del documento → error, no cuelgue ni panic
- [x] 2.6 [green] `scan.go`: scanner de estados (normal / basic / literal / multilínea / comentario) que emite líneas lógicas con su línea física de origen
- [x] 2.7 [refactor] `stripTrailingComment` de `assignment.go` queda absorbido por el scanner y se elimina

## 3. Valores (`value.go`)

- [x] 3.1 [red] Test: literal string `'a\nb'` no interpreta escapes; `'…'` sin cerrar → error
- [x] 3.2 [red] Test: `"""…"""` y `'''…'''` multilínea; el salto de línea inmediato tras la apertura se descarta
- [x] 3.3 [red] Test: `1_000` → 1000; `0x1f` → 31; `0o17` → 15; `0b101` → 5; `0x` solo → error
- [x] 3.4 [red] Test: `2026-01-01`, `2026-01-01T10:00:00Z` y `10:32:00` se devuelven como string
- [x] 3.5 [red] Test: token irreconocible (`@foo`) → error que lo nombra
- [x] 3.6 [green] `value.go` + `numeric.go`: literal/multilínea, enteros con `_` y prefijos, fecha→string
- [x] 3.7 [green] Los saltos de línea cuentan como espacio en blanco en todos los puntos de skip del parser

## 4. Arrays e inline tables (`composite.go`)

- [x] 4.1 [red] Test: `["a", "b",]` con coma final → `["a","b"]`
- [x] 4.2 [red] Test: `[ "a", , "b" ]` (coma doble) sigue siendo error
- [x] 4.3 [red] Test: `{ a = 1, }` (coma final en inline table) sigue siendo error
- [x] 4.4 [red] Test: array de inline tables repartido en varias líneas
- [x] 4.5 [green] `composite.go`: coma final permitida solo en arrays

## 5. Claves (`assignment.go`)

- [x] 5.1 [red] Test: `default.maxLines = 42` crea la tabla intermedia `default`
- [x] 5.2 [red] Test: `a.b = 1` y `a.c = 2` conviven en la misma tabla `a`
- [x] 5.3 [red] Test: `"a.b" = 1` produce la clave literal `a.b`, sin anidar; `'a.b' = 1` idem
- [x] 5.4 [red] Test: `"a.b".c = 1` anida bajo la clave literal `a.b`
- [x] 5.5 [red] Test: clave vacía, comilla sin cerrar y segmento vacío (`a..b`) → error
- [x] 5.6 [green] `keys.go`: split de clave respetando comillas + materialización de tablas intermedias

## 6. Tolerancia fuera de la sección objetivo (`document.go`)

- [x] 6.1 [red] Test: sintaxis inválida en `[tool.poetry.dependencies]` no impide extraer `tool.dupelens`
- [x] 6.2 [red] Test: la misma sintaxis inválida **dentro** de `[tool.dupelens]` sí produce error con línea y clave
- [x] 6.3 [red] Test: el error alcanza también a las sub-tablas (`[tool.linelens.default]`, `[[tool.linelens.rules]]`)
- [x] 6.4 [red] Test: `sectionPath=""` (tabla raíz) es estricto solo para las asignaciones previas al primer header
- [x] 6.5 [red] Test: header de tabla malformado → error en cualquier posición, aunque esté fuera de la sección
- [x] 6.6 [green] `parseDocument` recibe `sectionPath` y aplica la asimetría estricto-dentro / tolerante-fuera
- [x] 6.7 [red] Test: la sección objetivo ausente sigue devolviendo `(nil, false, nil)` aunque el resto del documento tenga basura

## 7. Verificación cruzada de los cinco tools

- [x] 7.1 `go build ./...` y `go test ./...` en los cinco módulos + los tres `_shared`
- [x] 7.2 Coverage 100% en `tomlmin` (`go test -cover`)
- [x] 7.3 Test de integración en `configload`: `Pyproject[T]` sobre un `pyproject.toml` real con `dependencies` multilínea
- [x] 7.4 Smoke manual: repo Python temporal con `pyproject.toml` estilo poetry + `[tool.linelens]`; los cinco binarios corren sin abortar
- [x] 7.5 Los cinco tools corridos sobre el propio repo siguen en verde (linelens < 100 líneas incluido)

## 8. Documentación

- [x] 8.1 ADR-018: reemplazar la tabla del subset ("NO soporta multi-line strings/arrays, dotted keys") por el subset ampliado + la regla de tolerancia; nota de por qué la asimetría
- [x] 8.2 Resuelto como sección de ADR-018 ("Severidad asimétrica") en vez de un ADR nuevo: la decisión es un matiz del parser ya decidido allí, no una decisión independiente
- [x] 8.3 `docs/CONFIGURATION.md`: ejemplo de `pyproject.toml` real con `dependencies` multilínea y `[tool.<name>]`
- [x] 8.4 CHANGELOG de los cinco tools con la entrada del fix
- [x] 8.5 `.agent/feature-list.json`: entrada F-023

## 9. Release

- [x] 9.1 Bump: linelens 0.3.3, secretlens 0.3.3, testlens 0.3.3, dupelens 0.4.1, scopelens 0.2.1 (patch: corrección sin cambio de contrato)
- [x] 9.2 `scripts/build-npm.sh` por tool (4 plataformas + meta) y `scripts/build-pypi.sh`
- [x] 9.3 Meta-paquetes: `@open_harness/open-harness` y `open-harness-suite` con las dependencias actualizadas
- [x] 9.4 `scripts/check-versions.sh` en verde
- [x] 9.5 Merge `develop` → `main`, tags
- [x] 9.6 Publicar en npm (25 paquetes + meta) y PyPI (5 tools + meta)
- [x] 9.7 Verificar desde los registries: instalar en un repo Python limpio con `pyproject.toml` multilínea y confirmar que ya no aborta
