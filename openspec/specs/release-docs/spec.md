# release-docs Specification

## Purpose
TBD - created by archiving change add-config-onboarding. Update Purpose after archive.
## Requirements
### Requirement: CHANGELOG por paquete

Cada uno de los cinco tools y el meta-paquete SHALL tener un `CHANGELOG.md` en formato *Keep a
Changelog*, con una entrada por versión publicada. Los cuatro tools originales y el meta SHALL
documentar la entrada `0.3.0` (los arreglos de la auditoría y los tres cambios de comportamiento);
scopelens SHALL documentar la entrada `0.1.0`.

#### Scenario: Cada paquete tiene su changelog

- **WHEN** se inspecciona el directorio de cada uno de los cinco tools y del meta
- **THEN** cada uno contiene un `CHANGELOG.md` con la versión actual y su descripción de cambios

#### Scenario: Los BREAKING quedan marcados

- **GIVEN** el `CHANGELOG.md` de secretlens 0.3.0
- **WHEN** se lee su entrada
- **THEN** el cambio de patterns aditivos aparece marcado como breaking, con la nota de opt-out

### Requirement: Guía central de configuración

El repositorio SHALL tener `docs/CONFIGURATION.md` que documente la configuración de los cinco tools:
las claves de cada `<tool>.json`, sus defaults, la precedencia de la cadena multi-ecosistema
(config JSON → pyproject → package.json → composer → defaults) y ejemplos. El README SHALL enlazarla.

#### Scenario: La guía cubre los cinco tools

- **WHEN** se lee `docs/CONFIGURATION.md`
- **THEN** documenta las claves de configuración de linelens, dupelens, secretlens, testlens y scopelens

#### Scenario: El README enlaza la guía

- **WHEN** se lee el README
- **THEN** contiene un enlace a `docs/CONFIGURATION.md`

