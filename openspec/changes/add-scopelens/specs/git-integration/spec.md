# git-integration

Obtención de la lista de archivos tocados desde `git`, resolución de la rama base, y comportamiento
ante cada condición en la que el dato no puede determinarse. Capability nueva, provista por
`scopelens 0.1.0`. Es el primer punto del monorepo donde un tool depende de un binario externo.

## ADDED Requirements

### Requirement: Ningún error operativo degrada a exit 0

`scopelens` MUST NOT reportar un conteo cuando no pudo obtener la lista completa de archivos. Toda
condición que impida determinar el conteo con certeza SHALL producir `exit 2` y un mensaje en
`stderr` que nombre la causa y la acción correctiva.

`exit 2` SHALL distinguirse de `exit 1`: el primero significa "no pude medir", el segundo "medí y se
excedió el presupuesto".

#### Scenario: git ausente en PATH

- **GIVEN** un entorno donde `git` no está en `PATH`
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el exit code es 2
- **AND** `stderr` contiene el texto `git` y la acción de instalarlo
- **AND** `stdout` NO contiene `OK`

#### Scenario: el directorio no es un repositorio git

- **GIVEN** un directorio temporal sin `.git`
- **WHEN** se ejecuta `scopelens check --fail` dentro de ese directorio
- **THEN** el exit code es 2
- **AND** el mensaje indica que no es un repositorio git

#### Scenario: clon shallow

- **GIVEN** un repositorio clonado con `--depth=1`
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el exit code es 2
- **AND** el mensaje menciona `shallow` y sugiere `fetch-depth: 0`
- **AND** el conteo NO se reporta

#### Scenario: la rama base no existe

- **GIVEN** un repositorio sin rama `main` ni `master` ni `origin/HEAD`
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el exit code es 2
- **AND** el mensaje sugiere pasar `--base <ref>` explícitamente

#### Scenario: git devuelve un exit code distinto de cero

- **GIVEN** un `git diff` que termina con exit code 128
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el exit code es 2
- **AND** el mensaje incluye el `stderr` de git

### Requirement: Resolución de la rama base por precedencia

`scopelens` SHALL resolver la rama base en este orden, deteniéndose en la primera que exista:
`--base`, el valor `base` de la config, `origin/HEAD`, `main`, `master`.

Sobre la ref resuelta el tool SHALL calcular `git merge-base <base> HEAD` y usar ese commit como
punto de comparación, no la punta de la rama base.

#### Scenario: --base gana sobre la config

- **GIVEN** una config con `"base": "main"`
- **WHEN** se ejecuta `scopelens check --base develop`
- **THEN** la comparación se hace contra `develop`
- **AND** el encabezado del reporte nombra `develop`

#### Scenario: merge-base aísla los commits ajenos a la rama

- **GIVEN** una rama creada desde `main`, y 6 commits nuevos agregados a `main` después
- **AND** la rama toca 3 archivos propios
- **WHEN** se ejecuta `scopelens check`
- **THEN** el conteo es 3
- **AND** los archivos de los 6 commits de `main` NO aparecen en el reporte

### Requirement: HEAD sobre la rama base mide sólo el índice

Cuando `HEAD` está sobre la propia rama base, `merge-base` coincide con `HEAD` y el diff acumulado es
vacío. En ese caso `scopelens` SHALL contar únicamente los archivos staged, sin emitir error.

#### Scenario: commit directo sobre main

- **GIVEN** un repositorio con `HEAD` en `main` y 4 archivos staged
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el conteo es 4
- **AND** el exit code es 0

#### Scenario: repositorio sin commits

- **GIVEN** un repositorio recién inicializado con `git init` y 2 archivos staged
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el conteo es 2
- **AND** el exit code es 0

### Requirement: Invocación acotada y no intrusiva

`scopelens` SHALL invocar `git` con `--no-optional-locks` para no interferir con otros procesos git
en curso, y bajo un contexto con timeout. Un timeout SHALL tratarse como error operativo (`exit 2`),
no como conteo vacío.

#### Scenario: timeout de git

- **GIVEN** un `git` que no responde dentro del timeout configurado
- **WHEN** se ejecuta `scopelens check --fail`
- **THEN** el exit code es 2
- **AND** el mensaje menciona el timeout

### Requirement: Parseo robusto de rutas

`scopelens` SHALL interpretar correctamente la salida de `git diff --name-only`, incluyendo rutas con
espacios y rutas con quoting octal de git (por ejemplo `"caf\303\251.ts"`). Una salida vacía SHALL
interpretarse como cero archivos, no como error.

#### Scenario: ruta con caracteres no ASCII

- **GIVEN** una rama que toca `src/café.ts`
- **WHEN** se ejecuta `scopelens check`
- **THEN** el reporte muestra `src/café.ts` con los acentos correctos
- **AND** el conteo es 1

#### Scenario: ruta con espacios

- **GIVEN** una rama que toca `docs/guía de uso.md`
- **WHEN** se ejecuta `scopelens check`
- **THEN** el archivo se cuenta una sola vez con la ruta completa
