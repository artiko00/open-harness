# release-metadata

Metadatos de versión y documentación del monorepo. Hoy la versión diverge en cinco lugares: los
binarios reportan linelens 0.2.1, dupelens 0.2.1, secretlens 0.2.1 y testlens 0.2.5, mientras el
README declara 0.2.0 en su tabla y 0.1.0 en su árbol de estructura, `open-harness.json` declara
0.1.0 para los cuatro y AGENTS.md declara 0.2.0.

## ADDED Requirements

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
