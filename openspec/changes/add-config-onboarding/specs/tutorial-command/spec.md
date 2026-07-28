# tutorial-command

Cada uno de los cinco lenses (linelens 0.3.0, dupelens 0.3.0, secretlens 0.3.0, testlens 0.3.0,
scopelens 0.1.0) expone `--tutorial`: una guía impresa que explica cómo configurarlo, sin salir de
la terminal.

## ADDED Requirements

### Requirement: Flag `--tutorial` en cada tool

Cada tool SHALL aceptar `--tutorial` (y `tutorial`) como comando de nivel superior que imprime a
stdout una guía de configuración estática y devuelve exit 0. La guía SHALL nombrar cada clave de
configuración que ese tool acepta, con su valor por defecto y un ejemplo, y listar sus flags.

#### Scenario: El tutorial se imprime y sale 0

- **WHEN** se ejecuta `<tool> --tutorial` para cada uno de los cinco tools
- **THEN** stdout contiene la guía de configuración
- **AND** el exit code es 0

#### Scenario: El tutorial nombra las claves reales de la config

- **GIVEN** el tipo `Config` que el tool deserializa
- **WHEN** se ejecuta `<tool> --tutorial`
- **THEN** la guía menciona cada clave de configuración que el tool acepta
- **AND** menciona el valor por defecto de cada una

#### Scenario: `--no-color` no emite secuencias ANSI

- **WHEN** se ejecuta `<tool> --tutorial --no-color`
- **THEN** la salida no contiene ninguna secuencia de escape ANSI

#### Scenario: Los flags se documentan en la guía

- **WHEN** se ejecuta `<tool> --tutorial`
- **THEN** la guía lista los flags del tool (`--config`, `--fail`, `--format`, y los propios del tool)
