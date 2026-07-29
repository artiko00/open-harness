# config-loading

Carga de configuración de `scopelens 0.1.0` desde el archivo central de cada ecosistema, siguiendo la
cadena ya establecida en ADR-018 para los otros cuatro lenses.

## ADDED Requirements

### Requirement: Cadena de configuración multi-ecosistema

`scopelens` SHALL resolver su configuración en este orden de precedencia, deteniéndose en la primera
fuente presente:

1. Flags de CLI
2. `scopelens.json` en la raíz
3. `pyproject.toml` → `[tool.scopelens]`
4. `package.json` → clave `"scopelens"`
5. `composer.json` → `extra.open-harness.scopelens`
6. Defaults compilados en el binario

Las claves soportadas SHALL ser `maxFiles` (int), `base` (string), `excludeTests` (bool) y `exclude`
(array de globs). Un flag de CLI presente MUST ganar sobre cualquier valor de archivo.

#### Scenario: pyproject.toml en un proyecto Python

- **GIVEN** un repo con `pyproject.toml` conteniendo `[tool.scopelens]` y `maxFiles = 10`
- **AND** sin `scopelens.json`
- **WHEN** se ejecuta `scopelens check --fail` sobre una rama de 12 archivos
- **THEN** el límite aplicado es 10
- **AND** el exit code es 1

#### Scenario: package.json en un proyecto Node

- **GIVEN** un repo con `package.json` conteniendo `"scopelens": { "maxFiles": 20 }`
- **WHEN** se ejecuta `scopelens check --fail` sobre una rama de 18 archivos
- **THEN** el límite aplicado es 20
- **AND** el exit code es 0

#### Scenario: composer.json en un proyecto PHP

- **GIVEN** un repo con `composer.json` conteniendo `extra.open-harness.scopelens.maxFiles = 8`
- **WHEN** se ejecuta `scopelens check`
- **THEN** el límite aplicado es 8

#### Scenario: scopelens.json gana sobre pyproject.toml

- **GIVEN** un repo con `scopelens.json` (`maxFiles: 15`) y `pyproject.toml` (`maxFiles = 10`)
- **WHEN** se ejecuta `scopelens check`
- **THEN** el límite aplicado es 15

#### Scenario: el flag gana sobre todo archivo

- **GIVEN** un repo con `scopelens.json` conteniendo `maxFiles: 15`
- **WHEN** se ejecuta `scopelens check --max-files 5`
- **THEN** el límite aplicado es 5

#### Scenario: sin ninguna fuente, default 15

- **GIVEN** un repo sin `scopelens.json`, sin `pyproject.toml`, sin `package.json` y sin `composer.json`
- **WHEN** se ejecuta `scopelens check`
- **THEN** el límite aplicado es 15

### Requirement: Config inválida falla ruidosa

Un archivo de configuración presente pero con sintaxis inválida, o con `maxFiles` negativo o no
numérico, SHALL producir `exit 2` con el nombre del archivo y la clave ofensora. `scopelens`
MUST NOT caer silenciosamente a los defaults en ese caso.

#### Scenario: JSON malformado

- **GIVEN** un `scopelens.json` con una coma sobrante
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el exit code es 2
- **AND** el mensaje nombra `scopelens.json`

#### Scenario: maxFiles negativo

- **GIVEN** un `scopelens.json` con `"maxFiles": -3`
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el exit code es 2
- **AND** el mensaje nombra la clave `maxFiles`

### Requirement: Exclusiones por defecto de los tres ecosistemas

Cuando la config no define `exclude`, `scopelens` SHALL aplicar un conjunto de exclusiones por
defecto que cubre los tres ecosistemas simultáneamente, sin requerir detección del lenguaje del
proyecto:

- Comunes: `.git/**`, `node_modules/**`, `vendor/**`, `dist/**`, `build/**`, `coverage/**`
- JS/TS: `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `.next/**`, `.nuxt/**`, `out/**`,
  `**/__snapshots__/**`
- Python: `poetry.lock`, `Pipfile.lock`, `uv.lock`, `**/__pycache__/**`, `*.egg-info/**`, `.venv/**`
- Go: `go.sum`, `**/*.pb.go`, `**/zz_generated*.go`

Un `exclude` definido por el usuario MUST reemplazar la lista completa, no sumarse a ella.

El matching de globs SHALL delegarse en `tools/_shared/pathmatch`. `scopelens` MUST NOT implementar
su propia lógica de globs.

#### Scenario: un repo Go no se rompe por las exclusiones de Node

- **GIVEN** un repo Go puro que toca 15 archivos `.go` y `go.sum`
- **WHEN** se ejecuta `scopelens check --fail` con `maxFiles` en 15
- **THEN** `go.sum` aparece bajo `excluded`
- **AND** el exit code es 0

#### Scenario: exclude del usuario reemplaza los defaults

- **GIVEN** un `scopelens.json` con `"exclude": ["docs/**"]`
- **AND** una rama que toca `go.sum` y `docs/api.md`
- **WHEN** se ejecuta `scopelens check`
- **THEN** `docs/api.md` aparece bajo `excluded`
- **AND** `go.sum` cuenta contra el presupuesto
