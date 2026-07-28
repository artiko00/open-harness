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

