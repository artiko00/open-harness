# release-metadata Specification

## Purpose

Define los metadatos de release y distribución: la constante version de cada main.go como única fuente de verdad, el gate que verifica su sincronía en docs, npm y PyPI, la forma de los paquetes npm y la alineación de README, backlog y ADRs con lo implementado.

## Requirements

### Requirement: Fuente única de verdad para la versión

La constante `version` de `tools/<tool>/main.go` SHALL ser la única fuente de verdad. Todo otro
lugar que declare versiones —`open-harness.json`, `README.md`, `AGENTS.md`, `openspec/config.yaml`,
los `package.json` de npm— SHALL derivarse de ella.

Un script del repositorio SHALL verificar la coincidencia y SHALL formar parte del gate de release.

#### Scenario: Verificación de sincronía

- **WHEN** se ejecuta el script de verificación de versiones
- **THEN** reporta que todas las fuentes coinciden con la constante de cada `main.go`
- **AND** el exit code es 0

#### Scenario: Divergencia detectada

- **GIVEN** un `open-harness.json` con una versión distinta a la de `main.go`
- **WHEN** se ejecuta el script de verificación
- **THEN** el script nombra el archivo divergente
- **AND** el exit code es 1

### Requirement: Documentación alineada con las capacidades reales

El README SHALL documentar todos los lenguajes soportados por testlens, incluido Dart, y todos los
flags disponibles en cada tool, incluidos `--config`, `--no-color`, `--format` y `--output`.

#### Scenario: Dart documentado

- **WHEN** se consulta la tabla de lenguajes de testlens en el README
- **THEN** Dart figura como soportado

#### Scenario: Flags completos

- **GIVEN** los flags aceptados por cada tool
- **WHEN** se comparan con los documentados en el README
- **THEN** no hay flags implementados que falten en la documentación

#### Scenario: Los ejemplos coinciden con la configuración real

- **WHEN** se comparan los ejemplos de hooks del README con `lefthook.yml`
- **THEN** el número de tools y los flags coinciden

### Requirement: El backlog refleja el trabajo realizado

Toda feature implementada SHALL tener su entrada en `.agent/feature-list.json`. Las features
referenciadas en mensajes de commit SHALL existir en ese archivo.

#### Scenario: Sin IDs huérfanos

- **WHEN** se comparan los IDs de feature citados en el historial de commits con `.agent/feature-list.json`
- **THEN** todos los IDs citados existen en el archivo

### Requirement: Los ADR describen el comportamiento implementado

Un ADR NO MUST describir comportamiento que el código no implementa. ADR-018 SHALL alinearse con la
resolución de la cadena de configuración efectivamente implementada.

#### Scenario: ADR-018 verificable

- **GIVEN** el comportamiento descrito en ADR-018 sobre precedencia por campo
- **WHEN** se ejecuta el escenario descrito en el ADR
- **THEN** el resultado observado coincide con el documentado

#### Scenario: Los ADR nuevos quedan registrados

- **WHEN** se completa este change
- **THEN** existe un ADR para el módulo compartido de path matching
- **AND** existe un ADR para la detección por entropía en secretlens

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
