# path-matching

Motor compartido de coincidencia de rutas para los cuatro tools (linelens 0.2.1, dupelens 0.2.1,
secretlens 0.2.1, testlens 0.2.5). Hoy `matchGlob`/`isExcluded` están replicados byte a byte en tres
tools y reimplementados de forma divergente en testlens.

## ADDED Requirements

### Requirement: Módulo compartido de path matching

El monorepo SHALL exponer un único paquete `tools/_shared/pathmatch` con las funciones de
coincidencia de globs, exclusión de rutas y detección de binarios. Los cuatro tools SHALL consumir
ese paquete y NO MUST mantener copias propias de esa lógica.

El paquete SHALL depender exclusivamente de la stdlib de Go, preservando ADR-002.

#### Scenario: Sin copias duplicadas del motor

- **WHEN** se ejecuta `dupelens check --dir tools --min-tokens 50`
- **THEN** no se reporta ningún duplicado originado en `matcher.go`, `binary.go` o `config_chain.go`
- **AND** el conteo de duplicados en `tools/` es 0

#### Scenario: Comportamiento idéntico entre tools

- **GIVEN** un árbol con `src/routes/big.ts` y `layouts/page.ts`
- **WHEN** se corre `check --dir .` con cada uno de los cuatro tools
- **THEN** los cuatro consideran ambos archivos dentro del alcance del análisis

### Requirement: Exclusión por segmento de ruta

La exclusión SHALL comparar segmentos completos de la ruta. Una entrada de `exclude` NO MUST
coincidir por subcadena arbitraria dentro de un nombre de directorio o archivo.

#### Scenario: Un directorio cuyo nombre contiene el término excluido no se excluye

- **GIVEN** un `exclude` que contiene `out`
- **AND** un árbol con `src/routes/handler.ts` y `layouts/page.ts`
- **WHEN** se ejecuta `testlens check --lang typescript --dir .`
- **THEN** `src/routes/handler.ts` y `layouts/page.ts` se analizan
- **AND** solo se excluyen directorios cuyo nombre completo es `out`

#### Scenario: El directorio exactamente nombrado sí se excluye

- **GIVEN** un `exclude` que contiene `out`
- **AND** un árbol con `out/bundle.js`
- **WHEN** se ejecuta cualquier tool con `check --dir .`
- **THEN** `out/bundle.js` no se analiza

### Requirement: Glob de doble asterisco a cualquier profundidad

Un patrón que contenga `**` SHALL coincidir con cero o más segmentos de ruta. En particular
`**/dir/**` SHALL coincidir con cualquier archivo bajo `dir/`, sin importar la profundidad.

#### Scenario: Coincidencia en profundidad mayor a uno

- **GIVEN** una regla `{"pattern": "**/migrations/**", "skip": true}`
- **AND** los archivos `db/migrations/a.sql` y `db/migrations/2024/b.sql`, ambos de 150 líneas
- **WHEN** se ejecuta `linelens check --dir .` con `maxLines` en 100
- **THEN** ninguno de los dos archivos se reporta como violación

#### Scenario: El patrón shipeado por init funciona

- **WHEN** se ejecuta `linelens init` o `dupelens init`
- **AND** se aplica el archivo generado sobre un árbol con migraciones anidadas
- **THEN** todas las migraciones quedan excluidas a cualquier profundidad

#### Scenario: Prefijo sin doble asterisco no cruza separadores

- **GIVEN** una regla con patrón `src/*.ts`
- **AND** los archivos `src/a.ts` y `src/nested/b.ts`
- **WHEN** se evalúa la regla
- **THEN** solo `src/a.ts` coincide

### Requirement: Detección de contenido binario

El paquete SHALL exponer la detección de binarios usada por los tools. La heurística NO MUST
clasificar como binario un archivo de texto codificado en UTF-16.

#### Scenario: UTF-16 se trata como texto

- **GIVEN** un archivo `.env` codificado en UTF-16 con un token real
- **WHEN** un tool recorre el árbol y evalúa el archivo con `IsBinaryContent`
- **THEN** el archivo se clasifica como texto y se incluye en el análisis, no se salta como binario
- **AND** la decodificación del contenido UTF-16 para reportar el token es responsabilidad del
  motor de detección (ver `secret-detection`), fuera del alcance de este módulo

#### Scenario: Un binario real se omite

- **GIVEN** un ejecutable ELF en el árbol
- **WHEN** se ejecuta cualquier tool con `check --dir .`
- **THEN** el archivo se clasifica como binario y no se analiza línea a línea
