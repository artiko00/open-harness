# cli-contract

Contrato de línea de comandos de `scopelens 0.1.0`, alineado con el de los otros cuatro lenses
(linelens 0.2.1, dupelens 0.2.1, secretlens 0.2.1, testlens 0.2.5).

## ADDED Requirements

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
