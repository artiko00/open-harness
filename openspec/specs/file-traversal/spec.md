# file-traversal Specification

## Purpose
TBD - created by archiving change fix-audit-findings. Update Purpose after archive.
## Requirements
### Requirement: Solo se abren archivos regulares

El recorrido SHALL abrir únicamente archivos regulares. Named pipes (FIFO), sockets, dispositivos y
otros archivos no regulares SHALL omitirse sin intentar leerlos.

#### Scenario: Un FIFO no cuelga el proceso

- **GIVEN** un árbol que contiene `a/pipe.go` creado con `mkfifo`
- **WHEN** se ejecuta `timeout 6 <tool> check --dir .` para cada uno de los cuatro tools
- **THEN** cada proceso termina antes del timeout
- **AND** ningún exit code es 124

#### Scenario: El FIFO se reporta como omitido

- **GIVEN** el mismo árbol con un FIFO
- **WHEN** se ejecuta `<tool> check --dir .`
- **THEN** el FIFO aparece en la lista de archivos omitidos con el motivo `not a regular file`

### Requirement: Los archivos omitidos se reportan y afectan el exit code

Cuando un archivo no se puede analizar —por ser no regular, por exceder el límite de línea, por
error de lectura o por ser binario— el tool SHALL listarlo con su motivo bajo un encabezado
`SKIPPED`. El resumen SHALL incluir el conteo de archivos omitidos.

Con `--fail`, los archivos omitidos por error de lectura o por exceder el límite de línea SHALL
producir exit 1. Los omitidos por ser binarios o no regulares NO MUST alterar el exit code.

Ningún tool SHALL imprimir un mensaje `OK:` si omitió archivos por error o por límite de línea.

#### Scenario: Error de lectura no se confunde con archivo limpio

- **GIVEN** un archivo sin permisos de lectura en el árbol
- **WHEN** se ejecuta `secretlens check --dir . --fail`
- **THEN** el archivo aparece bajo `SKIPPED` con el motivo del error
- **AND** el exit code es 1
- **AND** la salida no contiene `OK: no secrets detected`

#### Scenario: Binarios omitidos no rompen el gate

- **GIVEN** un árbol con un ejecutable ELF y ningún otro hallazgo
- **WHEN** se ejecuta `linelens check --dir . --fail`
- **THEN** el binario aparece en el conteo de omitidos
- **AND** el exit code es 0

### Requirement: Líneas que exceden el límite del buffer

Un archivo que contenga una línea mayor al límite del buffer de lectura NO MUST descartarse en
silencio. El tool SHALL reportarlo bajo `SKIPPED` con el motivo `line exceeds buffer limit`.

#### Scenario: Bundle minificado de una línea

- **GIVEN** un archivo `vendor.min.js` de 1,6 MB en una sola línea
- **WHEN** se ejecuta `linelens check --dir .`
- **THEN** el archivo aparece bajo `SKIPPED`
- **AND** el conteo de archivos escaneados refleja el total real del directorio

#### Scenario: Un secreto en un archivo omitido no produce OK

- **GIVEN** un `bundle.min.js` de una línea que contiene un token `ghp_` real
- **WHEN** se ejecuta `secretlens check --dir . --fail`
- **THEN** la salida no afirma `OK: no secrets detected`
- **AND** el exit code es 1

### Requirement: Detección de lenguaje limitada al alcance analizado

Cuando un tool infiere el lenguaje del proyecto a partir del árbol, la inferencia SHALL respetar la
configuración de `exclude`. Los archivos de directorios excluidos NO MUST influir en la detección.

#### Scenario: node_modules no decide el lenguaje del proyecto

- **GIVEN** un árbol con 10 archivos `.rb` bajo `node_modules/` y 7 archivos `.ts` en `src/`
- **WHEN** se ejecuta `testlens check --dir .`
- **THEN** el lenguaje detectado es TypeScript y no se emite el aviso de fallback a todas las extensiones
- **AND** ningún archivo bajo `node_modules/` aparece en el reporte

#### Scenario: las dependencias no producen un falso negativo

- **GIVEN** un árbol con 10 archivos `.rb` bajo `node_modules/` y un único `src/app.ts` sin test
- **WHEN** se ejecuta `testlens check --dir .`
- **THEN** `src/app.ts` se reporta como archivo sin test
- **AND** la salida no afirma que todos los archivos fuente tienen tests

> El umbral que hoy exige más de cinco archivos para inferir un lenguaje hace que un proyecto
> pequeño caiga al conjunto completo de extensiones. Eliminarlo corresponde a
> `test-coverage-detection`; este requisito solo exige que las rutas excluidas no participen
> de la inferencia.

