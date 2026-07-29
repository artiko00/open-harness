# config-loading Specification

## Purpose
TBD - created by archiving change fix-audit-findings. Update Purpose after archive.
## Requirements
### Requirement: Ruta de configuración explícita inexistente es un error

Cuando el usuario pasa `--config <ruta>` explícitamente y esa ruta no existe o no se puede leer, el
tool SHALL escribir un error en stderr y devolver exit 1. NO MUST caer a la cadena de configuración
ni a los defaults compilados.

La ausencia del archivo de configuración por defecto SHALL seguir siendo válida y continuar con la
cadena de configuración.

#### Scenario: Ruta explícita inexistente

- **WHEN** se ejecuta `<tool> check --config /nope/nope.json --dir .` para cada uno de los cuatro tools
- **THEN** stderr contiene `config file "/nope/nope.json" not found`
- **AND** el exit code es 1

#### Scenario: Ausencia del archivo por defecto sigue siendo válida

- **GIVEN** un directorio sin `linelens.json` ni `package.json`
- **WHEN** se ejecuta `linelens check --dir .`
- **THEN** se aplican los defaults compilados
- **AND** el exit code es 0

### Requirement: Claves de configuración desconocidas se avisan sin fallar

Al decodificar un archivo de configuración, los cuatro tools SHALL detectar las claves que no
correspondan a ningún campo conocido y emitir una advertencia a stderr que nombre la clave, pero
NO MUST descartar la configuración ni cambiar el exit code por ese motivo: los campos conocidos SHALL
cargarse igual. Este comportamiento preserva la retrocompatibilidad con versiones previas, que
ignoraban las claves desconocidas en silencio, a la vez que las hace visibles.

#### Scenario: Clave inexistente en testlens

- **GIVEN** un `testlens.json` que declara `"skip": ["legacy"]`
- **WHEN** se ejecuta `testlens check --dir .`
- **THEN** stderr advierte que `skip` es una clave desconocida y sugiere `exclude`
- **AND** el exit code no es alterado por la clave desconocida

#### Scenario: Typo en una clave de linelens

- **GIVEN** un `linelens.json` que declara `"excludes"` y un `default.maxLines` válido
- **WHEN** se ejecuta `linelens check --dir .`
- **THEN** stderr advierte sobre la clave desconocida `excludes`
- **AND** el valor de `maxLines` sí se aplica (los campos conocidos se cargan)

#### Scenario: La configuración del propio repo es válida

- **WHEN** se ejecuta cada uno de los cuatro tools desde la raíz del repositorio
- **THEN** ninguno emite advertencias de claves desconocidas
- **AND** ningún archivo de configuración del repositorio contiene claves inertes

### Requirement: Precedencia de la cadena de configuración

La cadena de configuración SHALL resolverse por campo: para cada campo, gana el valor del primer
archivo de la cadena que lo defina; los campos no definidos SHALL tomarse del siguiente archivo, y
solo al agotarse la cadena SHALL usarse el default compilado.

Este comportamiento SHALL quedar documentado en ADR-018 y en el README con un ejemplo.

#### Scenario: Campos complementarios entre dos archivos

- **GIVEN** un `pyproject.toml` con `[tool.linelens.default] maxLines = 10`
- **AND** un `package.json` con `{"linelens": {"exclude": ["legacy"]}}`
- **AND** un archivo `legacy/old.py` de 50 líneas
- **WHEN** se ejecuta `linelens check --dir .`
- **THEN** el límite aplicado es 10
- **AND** `legacy/old.py` queda excluido y no se reporta

#### Scenario: El primer archivo gana en un campo compartido

- **GIVEN** un `pyproject.toml` y un `package.json` que definen ambos `maxLines`
- **WHEN** se ejecuta `linelens check --dir .`
- **THEN** se aplica el valor de `pyproject.toml`

### Requirement: Umbrales del repositorio sin hallazgos ocultos

Los archivos de configuración versionados en el repositorio NO MUST elevar umbrales por encima del
default con el efecto de ocultar hallazgos reales. `dupelens.json` SHALL volver al `minTokens` por
defecto una vez absorbidas las copias del motor de path matching.

#### Scenario: El repo pasa con el umbral por defecto

- **WHEN** se ejecuta `dupelens check --dir tools --min-tokens 50 --no-color`
- **THEN** no se reportan duplicados
- **AND** `dupelens.json` no declara un `minTokens` mayor al default

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

