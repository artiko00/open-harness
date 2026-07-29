# cli-contract Specification

## Purpose
TBD - created by archiving change fix-audit-findings. Update Purpose after archive.
## Requirements
### Requirement: Salida en JSON en los cuatro tools

Los cuatro tools SHALL aceptar `--format <console|json>`, con `console` por defecto. La salida JSON
SHALL escribirse en stdout sin códigos de color ANSI y SHALL ser parseable sin preprocesamiento.

#### Scenario: JSON parseable en cada tool

- **WHEN** se ejecuta `<tool> check --dir . --format json` para cada uno de los cuatro tools
- **THEN** la salida completa de stdout es un documento JSON válido
- **AND** incluye el conteo de archivos escaneados, los hallazgos y los archivos omitidos

#### Scenario: El formato consola sigue siendo el default

- **WHEN** se ejecuta `<tool> check --dir .` sin `--format`
- **THEN** la salida es el reporte de consola legible

### Requirement: Valores de flag inválidos son errores

Un valor no soportado en cualquier flag enumerado SHALL producir un mensaje en stderr que liste los
valores válidos y un exit code 1. NO MUST caer silenciosamente a un default.

#### Scenario: Formato desconocido

- **WHEN** se ejecuta `dupelens check --format=xml --dir .`
- **THEN** stderr indica que `xml` no es válido y lista `console, json`
- **AND** el exit code es 1

#### Scenario: Lenguaje desconocido

- **WHEN** se ejecuta `testlens check --lang typscript --dir . --fail`
- **THEN** stderr indica que `typscript` no es un lenguaje soportado y lista los soportados
- **AND** el exit code es 1
- **AND** la salida no afirma que todos los archivos tienen tests

#### Scenario: Override numérico no positivo

- **WHEN** se ejecuta `linelens check --max -5 --dir .`
- **THEN** stderr indica que `--max` debe ser mayor que 0
- **AND** el exit code es 1

#### Scenario: Override numérico en cero

- **WHEN** se ejecuta `dupelens check --min-tokens 0 --dir .`
- **THEN** stderr indica que `--min-tokens` debe ser mayor que 0
- **AND** el exit code es 1

### Requirement: Validación del directorio de entrada

Los cuatro tools SHALL validar que el valor de `--dir` existe y es un directorio antes de iniciar el
recorrido, con un mensaje de error homogéneo y exit code 1.

#### Scenario: Directorio inexistente

- **WHEN** se ejecuta `<tool> check --dir /does/not/exist` para cada uno de los cuatro tools
- **THEN** stderr contiene `directory "/does/not/exist" not accessible`
- **AND** el exit code es 1
- **AND** ningún tool emite antes un mensaje de diagnóstico sobre detección de lenguaje

### Requirement: Formato de reporte homogéneo

En formato consola, los cuatro tools SHALL usar la misma estructura: un encabezado
`<HALLAZGOS> (N …)` cuando hay violaciones, una sección `SKIPPED` cuando hay omitidos, y una línea
final que empieza con `SUMMARY:`. Cuando no hay hallazgos, la línea final SHALL empezar con `OK:`.

#### Scenario: testlens adopta el formato común

- **GIVEN** un proyecto Go con dos archivos fuente sin test
- **WHEN** se ejecuta `testlens check --lang go --dir . --no-color`
- **THEN** la salida incluye un encabezado con el conteo de archivos sin test
- **AND** la última línea empieza con `SUMMARY:`

#### Scenario: Mensaje de éxito homogéneo

- **GIVEN** un proyecto sin hallazgos ni omitidos
- **WHEN** se ejecuta `<tool> check --dir . --no-color` para cada uno de los cuatro tools
- **THEN** la última línea de cada uno empieza con `OK:`

### Requirement: Diagnóstico en stderr

Todo mensaje que no sea parte del reporte —advertencias, avisos de detección, errores— SHALL
escribirse en stderr. Stdout SHALL contener únicamente el reporte.

#### Scenario: Aviso de detección no contamina stdout

- **GIVEN** un directorio donde no se puede inferir el lenguaje
- **WHEN** se ejecuta `testlens check --dir ./empty --no-color 2>/dev/null`
- **THEN** stdout no contiene el aviso sobre detección de lenguaje

### Requirement: Contrato de exit codes

Los cuatro tools SHALL usar: 0 cuando no hay violaciones o cuando no se pasó `--fail`; 1 cuando
`--fail` está presente y hay violaciones; 1 ante comando desconocido, flag inválido, `--dir`
inaccesible o `--config` explícita inexistente.

#### Scenario: Sin --fail las violaciones no rompen el build

- **GIVEN** un árbol con violaciones
- **WHEN** se ejecuta `<tool> check --dir .` sin `--fail`
- **THEN** el exit code es 0

#### Scenario: Con --fail las violaciones rompen el build

- **GIVEN** el mismo árbol con violaciones
- **WHEN** se ejecuta `<tool> check --dir . --fail`
- **THEN** el exit code es 1

### Requirement: Subcomandos y flags

`scopelens` SHALL exponer los subcomandos `check`, `version` e `init`, y el flag `--help`.

`check` SHALL aceptar: `--max-files <int>`, `--base <ref>`, `--dir <path>`, `--staged-only`,
`--exclude-tests`, `--fail`, `--no-color`.

`main.go` MUST seguir el patrón del monorepo: variable `osExit`, función `run([]string) int` y
`flag.ContinueOnError`.

#### Scenario: version imprime la versión

- **WHEN** se ejecuta `scopelens version`
- **THEN** la salida contiene `scopelens` y la versión de `const version` en `main.go`
- **AND** el exit code es 0

#### Scenario: init genera la config por defecto

- **GIVEN** un directorio sin `scopelens.json`
- **WHEN** se ejecuta `scopelens init`
- **THEN** se crea `scopelens.json` con `maxFiles: 15` y la lista de `exclude` por defecto
- **AND** el exit code es 0

#### Scenario: flag desconocido no entra en pánico

- **WHEN** se ejecuta `scopelens check --inexistente`
- **THEN** el exit code es 2
- **AND** la salida muestra el uso del comando

#### Scenario: --no-color suprime los códigos ANSI

- **GIVEN** una rama que excede el presupuesto
- **WHEN** se ejecuta `scopelens check --no-color`
- **THEN** la salida NO contiene ninguna secuencia de escape ANSI

### Requirement: Semántica de exit codes

`scopelens` SHALL usar exactamente tres exit codes: `0` (medido y dentro del presupuesto, o medido
fuera del presupuesto sin `--fail`), `1` (medido, fuera del presupuesto, con `--fail`) y `2` (no se
pudo medir: error operativo, de git o de config).

#### Scenario: los tres códigos son distinguibles

- **WHEN** se ejecuta `scopelens check --fail` bajo el presupuesto
- **THEN** el exit code es 0
- **WHEN** se ejecuta sobre una rama que lo excede
- **THEN** el exit code es 1
- **WHEN** se ejecuta fuera de un repositorio git
- **THEN** el exit code es 2

### Requirement: Reporte con desglose y resumen

La salida de `check` SHALL incluir un encabezado con la rama, la base y el merge-base abreviado; el
listado de archivos agrupado por categoría; y una línea final `SUMMARY` con el conteo contable, el
conteo de excluidos y el límite aplicado.

#### Scenario: encabezado nombra la base efectiva

- **GIVEN** una rama `feature/x` comparada contra `origin/main`
- **WHEN** se ejecuta `scopelens check --no-color`
- **THEN** el encabezado contiene `feature/x`, `origin/main` y el hash abreviado del merge-base

#### Scenario: línea SUMMARY siempre presente

- **GIVEN** cualquier ejecución exitosa de `check`
- **THEN** la última línea de la salida comienza con `SUMMARY:`
- **AND** contiene el conteo contable, el de excluidos y el límite

### Requirement: Integración con lefthook

`scopelens` SHALL incorporarse a `lefthook.yml` en `pre-commit` con
`tools/scopelens/scopelens check --fail --no-color`, junto a los otros cuatro lenses.

#### Scenario: el hook aborta el commit

- **GIVEN** un repo cuyo pre-commit incluye `scopelens check --fail --no-color`
- **AND** una rama con 18 archivos contables y `maxFiles` en 15
- **WHEN** se ejecuta `git commit`
- **THEN** el commit se aborta
- **AND** la salida del hook muestra el desglose de scopelens

