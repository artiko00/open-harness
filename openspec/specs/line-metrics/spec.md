# line-metrics Specification

## Purpose
TBD - created by archiving change fix-audit-findings. Update Purpose after archive.
## Requirements
### Requirement: Conteo de líneas de código

linelens SHALL contar líneas de código, excluyendo líneas en blanco y líneas que solo contienen
comentarios, reutilizando el stripper de comentarios ya presente en el monorepo. El reporte SHALL
mostrar tanto las líneas de código como el total físico.

#### Scenario: Cabecera de licencia no genera violación

- **GIVEN** un archivo de 382 líneas donde 380 son una cabecera de copyright y 2 son código
- **WHEN** se ejecuta `linelens check --dir .` con `maxLines` en 100
- **THEN** el archivo no se reporta como violación

#### Scenario: Un archivo de código real sí se reporta

- **GIVEN** un archivo con 250 líneas de código y sin comentarios
- **WHEN** se ejecuta `linelens check --dir .` con `maxLines` en 100
- **THEN** el archivo se reporta como violación

#### Scenario: El reporte muestra ambas métricas

- **GIVEN** un archivo con 300 líneas físicas y 120 de código
- **WHEN** se ejecuta `linelens check --dir . --no-color`
- **THEN** el reporte indica las líneas de código y el total físico

### Requirement: Alcance limitado a archivos de código

El análisis SHALL restringirse por defecto a extensiones de código y SHALL excluir por defecto los
archivos generados (`*.pb.go`, `*_gen.go`, `*.g.dart`, `*-lock.json`) y los archivos de datos.
La lista SHALL ser configurable.

#### Scenario: Catálogo de traducción no se reporta

- **GIVEN** un archivo `i18n.json` de 502 líneas con claves de traducción
- **WHEN** se ejecuta `linelens check --dir .` con `maxLines` en 100
- **THEN** el archivo no se reporta

#### Scenario: Esquema SQL no se reporta

- **GIVEN** un `schema.sql` de 300 líneas de sentencias `CREATE TABLE`
- **WHEN** se ejecuta `linelens check --dir .` con `maxLines` en 100
- **THEN** el archivo no se reporta

### Requirement: Métrica de anidamiento

linelens SHALL reportar la profundidad máxima de anidamiento por archivo y SHALL permitir
configurar un umbral propio para esa métrica, de modo que un archivo corto pero denso se detecte.

#### Scenario: Archivo corto con anidamiento profundo

- **GIVEN** un archivo de 83 líneas con 80 condicionales anidados
- **WHEN** se ejecuta `linelens check --dir .` con umbral de anidamiento en 5
- **THEN** el archivo se reporta por exceder la profundidad de anidamiento

#### Scenario: Archivo plano y largo no dispara la métrica de anidamiento

- **GIVEN** un archivo de 200 líneas sin anidamiento
- **WHEN** se ejecuta `linelens check --dir .` con umbral de anidamiento en 5
- **THEN** el archivo no se reporta por anidamiento

### Requirement: Consistencia del conteo físico

El total físico reportado SHALL coincidir con el criterio de `wc -l` respecto de la última línea sin
salto de línea terminal, o la diferencia SHALL documentarse explícitamente en el README.

#### Scenario: Archivo sin salto de línea final

- **GIVEN** un archivo cuya última línea no termina en salto de línea
- **WHEN** se ejecuta `linelens check --dir . --format json`
- **THEN** el total físico reportado coincide con el criterio documentado

