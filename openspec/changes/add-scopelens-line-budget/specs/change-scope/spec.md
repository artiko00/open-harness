# change-scope

Delta: scopelens agrega un presupuesto de líneas, combinable con el de archivos.

## ADDED Requirements

### Requirement: Presupuesto de líneas del diff

scopelens SHALL contar las líneas cambiadas del diff acumulado (rama vs base, unido con el índice)
de los archivos contables, y SHALL exponer un presupuesto `maxLines`. `maxLines: 0` (default) lo
deja deshabilitado. Los archivos excluidos NO MUST aportar líneas; los de test aportan salvo
`--exclude-tests`. La métrica SHALL ser configurable con `lineMetric`: `changed` (agregadas +
borradas, default) o `added` (sólo agregadas). Los archivos binarios SHALL contar 0 líneas.

#### Scenario: Un solo archivo enorme excede por líneas

- **GIVEN** un diff de un único archivo `.ts` con 11000 líneas agregadas
- **WHEN** se ejecuta `scopelens check --max-lines 5000 --fail`
- **THEN** el exit code es 1 y el reporte indica que se excedió el presupuesto de líneas
- **AND** el conteo de archivos (1) por sí solo no habría fallado

#### Scenario: lineMetric added ignora las borradas

- **GIVEN** un diff con 100 líneas agregadas y 9000 borradas en archivos contables
- **WHEN** `lineMetric` es `added` y `maxLines` es 5000
- **THEN** el conteo de líneas es 100 y no excede
- **AND** con `lineMetric` `changed` el conteo es 9100 y sí excede

#### Scenario: Los excluidos no aportan líneas

- **GIVEN** un diff con un `pnpm-lock.yaml` de 8000 líneas y `src/a.ts` de 20
- **WHEN** se cuenta con `maxLines` habilitado
- **THEN** el lockfile no aporta al conteo de líneas y el total contable es 20

### Requirement: Combinación configurable de presupuestos

Cuando `maxFiles` y `maxLines` están **ambos** habilitados (> 0), scopelens SHALL combinarlos según
`mode`: con `or` (default) el gate falla si se excede **cualquiera**; con `and` falla sólo si se
exceden **ambos**. Si sólo uno está habilitado, SHALL aplicar únicamente ese; si ninguno, NO MUST
fallar nunca por presupuesto.

#### Scenario: mode or falla ante cualquier exceso

- **GIVEN** `maxFiles 15` y `maxLines 5000` con `mode or`
- **WHEN** el diff tiene 3 archivos pero 8000 líneas
- **THEN** con `--fail` el exit code es 1

#### Scenario: mode and exige ambos excesos

- **GIVEN** `maxFiles 15` y `maxLines 5000` con `mode and`
- **WHEN** el diff tiene 3 archivos y 8000 líneas
- **THEN** con `--fail` el exit code es 0 (sólo un presupuesto excedido)
- **AND** con 20 archivos y 8000 líneas el exit code es 1 (ambos excedidos)

#### Scenario: Retrocompatibilidad con maxLines deshabilitado

- **GIVEN** `maxLines: 0` (default)
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el gate depende sólo de `maxFiles`, igual que antes

### Requirement: Validación de la configuración de líneas

Un `maxLines` negativo, un `mode` fuera de `{or, and}` o un `lineMetric` fuera de `{changed, added}`
SHALL producir exit 2 con un mensaje que nombre la causa; NO MUST degradar a un conteo silencioso.

#### Scenario: mode inválido

- **WHEN** se ejecuta `scopelens check --mode xor`
- **THEN** stderr indica que `mode` debe ser `or` o `and` y el exit code es 2
