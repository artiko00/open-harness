# suite-init

El meta-paquete `open-harness` expone un comando `init` que deja el proyecto configurado de una sola
corrida, creando los archivos de configuración de los cinco lenses en la raíz.

## ADDED Requirements

### Requirement: `open-harness init` crea todos los archivos de config

El launcher del meta-paquete SHALL exponer `open-harness init`, que crea los cinco archivos de
configuración (`linelens.json`, `dupelens.json`, `secretlens.json`, `testlens.json`,
`scopelens.json`) en el directorio actual, delegando en el `init` de cada tool. Cada tool SHALL
conservar además su propio `init` individual.

#### Scenario: Un solo comando configura la suite

- **GIVEN** un directorio de proyecto sin archivos de configuración
- **WHEN** se ejecuta `open-harness init`
- **THEN** existen los cinco archivos `<tool>.json` en la raíz con sus defaults

#### Scenario: El init individual sigue disponible

- **WHEN** se ejecuta `<tool> init` para cualquiera de los cinco tools
- **THEN** se crea únicamente el `<tool>.json` de ese tool

#### Scenario: No sobrescribe sin aviso

- **GIVEN** un directorio donde ya existe `linelens.json`
- **WHEN** se ejecuta `open-harness init`
- **THEN** el comando no sobrescribe silenciosamente el archivo existente
- **AND** informa cuáles archivos creó y cuáles ya existían
