# config-loading

Carga de configuración de linelens 0.2.1, dupelens 0.2.1, secretlens 0.2.1 y testlens 0.2.5.
Hoy `--config` con una ruta inexistente cae a defaults en silencio, las claves desconocidas se
descartan sin aviso —el `testlens.json` del propio repo declara `skip` y `languages`, que no
existen, y por eso es inerte— y ADR-018 describe un merge por campo que la implementación no hace.

## ADDED Requirements

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
