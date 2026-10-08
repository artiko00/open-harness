## ADDED Requirements

### Requirement: Los paquetes npm de plataforma no exponen comandos

Los paquetes npm de plataforma (`@open_harness/<tool>-<os>-<cpu>`) NO MUST declarar el campo `bin`.
El único paquete que expone el comando `<tool>` SHALL ser el wrapper `@open_harness/<tool>` (y el meta,
que lo re-exporta). Si una plataforma declara un bin con el mismo nombre, npm borra el link del
wrapper al revertir las dependencias opcionales de las otras plataformas.

El script de verificación de versiones SHALL fallar si algún paquete de plataforma declara `bin`.

#### Scenario: El comando queda disponible tras instalar el wrapper

- **GIVEN** un proyecto sin dependencias
- **WHEN** se ejecuta `npm install --save-dev @open_harness/dupelens`
- **THEN** existe `node_modules/.bin/dupelens`
- **AND** `npx dupelens version` imprime la versión publicada

#### Scenario: El meta deja disponibles los cinco comandos

- **GIVEN** un proyecto sin dependencias
- **WHEN** se ejecuta `npm install --save-dev @open_harness/open-harness`
- **THEN** `npx linelens version`, `npx dupelens version`, `npx secretlens version`,
  `npx testlens version` y `npx scopelens version` imprimen sus versiones

#### Scenario: El gate detecta un bin en una plataforma

- **GIVEN** `npm/@open_harness/dupelens-linux-x64/package.json` con un campo `bin`
- **WHEN** se ejecuta `bash scripts/check-versions.sh`
- **THEN** el script nombra ese paquete
- **AND** el exit code es 1
