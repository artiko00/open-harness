# change-scope Specification

## Purpose
TBD - created by archiving change add-scopelens. Update Purpose after archive.
## Requirements
### Requirement: Conteo acumulado de la rama, no del commit

`scopelens` SHALL contar la unión de los archivos del diff acumulado entre la base de merge y `HEAD`
con los archivos del índice (staged). El conteo MUST ser una unión de conjuntos: un archivo presente
en ambos diffs cuenta una sola vez.

El tool SHALL invocar `git diff` con `--diff-filter=ACMRD` (incluye borrados) y con `-M` (un rename
cuenta como un archivo, no como dos).

#### Scenario: Cinco commits pequeños superan el presupuesto

- **GIVEN** una rama con 5 commits que tocan 4 archivos distintos cada uno (20 en total)
- **AND** `maxFiles` configurado en 15
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el conteo reportado es 20
- **AND** el exit code es 1

#### Scenario: Un archivo tocado dos veces cuenta una vez

- **GIVEN** una rama donde `src/a.ts` fue modificado en un commit previo
- **AND** `src/a.ts` está además staged con nuevos cambios
- **WHEN** se ejecuta `scopelens check`
- **THEN** `src/a.ts` aparece una sola vez en el reporte
- **AND** el conteo total lo suma una sola vez

#### Scenario: Un rename cuenta como un archivo

- **GIVEN** una rama donde `src/viejo.py` fue renombrado a `src/nuevo.py` sin otros cambios
- **WHEN** se ejecuta `scopelens check`
- **THEN** el conteo total es 1

#### Scenario: Un archivo borrado cuenta

- **GIVEN** una rama donde `src/obsoleto.go` fue eliminado
- **WHEN** se ejecuta `scopelens check`
- **THEN** `src/obsoleto.go` aparece en el reporte
- **AND** el conteo total lo incluye

### Requirement: Umbral binario con exit code diferenciado

`scopelens` SHALL comparar el conteo de archivos contables contra `maxFiles`. Con `--fail`, un
conteo estrictamente mayor a `maxFiles` MUST producir `exit 1`. Un conteo menor o igual MUST producir
`exit 0`.

El exit code 1 SHALL usarse exclusivamente para "el gate midió y el presupuesto se excedió", nunca
para errores operativos.

#### Scenario: Conteo exactamente en el límite pasa

- **GIVEN** una rama que toca exactamente 15 archivos contables
- **AND** `maxFiles` configurado en 15
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el exit code es 0
- **AND** la salida contiene `OK`

#### Scenario: Sin --fail reporta pero no bloquea

- **GIVEN** una rama que toca 18 archivos contables con `maxFiles` en 15
- **WHEN** se ejecuta `scopelens check` sin `--fail`
- **THEN** la salida reporta `FAIL: 18 files (max 15)`
- **AND** el exit code es 0

### Requirement: Clasificación en source, test y excluded

`scopelens` SHALL clasificar cada archivo tocado en exactamente una de tres categorías: `excluded`
(matchea un patrón de `exclude`), `test` (matchea un layout de test conocido) o `source`.

Los archivos `excluded` MUST NOT contar contra el presupuesto. Los archivos `test` SHALL contar,
salvo que `--exclude-tests` esté activo. La clasificación SHALL derivarse únicamente de la ruta,
sin leer el contenido del archivo.

El reporte SHALL mostrar el desglose por categoría, y cada archivo `excluded` SHALL mostrar el motivo
de su exclusión.

#### Scenario: Lockfile regenerado no consume presupuesto

- **GIVEN** una rama que toca 15 archivos `.ts` y además `pnpm-lock.yaml`
- **AND** `maxFiles` configurado en 15
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el conteo contable es 15
- **AND** `pnpm-lock.yaml` aparece bajo `excluded` con motivo `lockfile`
- **AND** el exit code es 0

#### Scenario: Tests de los tres ecosistemas se reconocen

- **GIVEN** una rama que toca `src/a.test.ts`, `tests/test_b.py` y `pkg/c_test.go`
- **WHEN** se ejecuta `scopelens check`
- **THEN** los tres archivos aparecen bajo la categoría `test`

#### Scenario: --exclude-tests descuenta los tests del presupuesto

- **GIVEN** una rama que toca 10 archivos `source` y 8 archivos `test`
- **AND** `maxFiles` configurado en 15
- **WHEN** se ejecuta `scopelens check --fail --exclude-tests`
- **THEN** el conteo contable es 10
- **AND** el exit code es 0

#### Scenario: Sin --exclude-tests los tests cuentan

- **GIVEN** la misma rama de 10 `source` + 8 `test` con `maxFiles` en 15
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el conteo contable es 18
- **AND** el exit code es 1

### Requirement: Modo staged-only

Con `--staged-only`, `scopelens` SHALL contar únicamente los archivos del índice, ignorando el diff
acumulado de la rama.

#### Scenario: staged-only ignora commits previos de la rama

- **GIVEN** una rama con 12 archivos ya commiteados y 4 archivos staged
- **AND** `maxFiles` configurado en 15
- **WHEN** se ejecuta `scopelens check --fail --staged-only`
- **THEN** el conteo reportado es 4
- **AND** el exit code es 0

### Requirement: Salida determinista

Dos ejecuciones de `scopelens check` sobre el mismo estado del repositorio SHALL producir salida
byte-idéntica. El orden de los archivos dentro de cada categoría MUST ser lexicográfico y explícito,
sin depender del orden de iteración de un `map`.

#### Scenario: Veinte corridas idénticas

- **GIVEN** una rama con 18 archivos tocados repartidos en las tres categorías
- **WHEN** se ejecuta `scopelens check --no-color` 20 veces consecutivas
- **THEN** las 20 salidas son byte-idénticas entre sí

