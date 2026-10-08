## MODIFIED Requirements

### Requirement: Contrato de exit codes

Los cuatro tools SHALL usar: 0 cuando no hay violaciones o cuando no se pasó `--fail`; 1 cuando
`--fail` está presente y hay violaciones; 1 ante comando desconocido, flag inválido, `--dir`
inaccesible o `--config` explícita inexistente.

`dupelens` SHALL usar además el exit code 2 cuando el escaneo excede su presupuesto de memoria: la
medición no se completó, así que NO MUST terminar con 0 aunque no se haya pasado `--fail`. Los
errores de uso y de configuración de `dupelens` SHALL seguir terminando con 1.

#### Scenario: Sin --fail las violaciones no rompen el build

- **GIVEN** un árbol con violaciones
- **WHEN** se ejecuta `<tool> check --dir .` sin `--fail`
- **THEN** el exit code es 0

#### Scenario: Con --fail las violaciones rompen el build

- **GIVEN** el mismo árbol con violaciones
- **WHEN** se ejecuta `<tool> check --dir . --fail`
- **THEN** el exit code es 1

#### Scenario: dupelens distingue la medición fallida de las violaciones

- **GIVEN** un árbol cuyo escaneo excede el presupuesto de memoria de `dupelens`
- **WHEN** se ejecuta `dupelens check --dir .` con o sin `--fail`
- **THEN** el exit code es 2

#### Scenario: Los errores de uso de dupelens siguen siendo 1

- **WHEN** se ejecuta `dupelens check --max-memory 0MiB`
- **THEN** el exit code es 1
