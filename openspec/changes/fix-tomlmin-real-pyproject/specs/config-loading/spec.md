# config-loading

Lectura de `pyproject.toml` por los cinco tools mediante el parser compartido `tools/_shared/tomlmin`.
Hoy el parser recorre el documento línea por línea y aborta ante cualquier sintaxis fuera de su
subset, esté o no dentro de la tabla `[tool.<name>]` que le interesa: un `dependencies = [`
multilínea en `[project]` —la forma canónica de PEP 621— tumba la carga completa de configuración.

## ADDED Requirements

### Requirement: Arrays multilínea y trailing comma

El parser SHALL aceptar arrays cuyo contenido se extiende por varias líneas físicas, acumulando la
asignación hasta que los delimitadores `[` `]` `{` `}` queden balanceados fuera de strings. SHALL
aceptar también una coma final antes del `]` de cierre, tal como permite TOML 1.0.

Las inline tables `{ … }` MUST NOT aceptar coma final, y SHALL seguir rechazándose con error, igual
que en TOML 1.0.

Los comentarios `#` que aparezcan dentro de un array multilínea SHALL descartarse línea a línea sin
afectar el valor.

#### Scenario: dependencies multilínea de PEP 621

- **GIVEN** un `pyproject.toml` con `[project]` y `dependencies = [` abierto en varias líneas
- **AND** una tabla `[tool.linelens]` con `maxLines = 200`
- **WHEN** `linelens` carga la configuración
- **THEN** la carga no produce error
- **AND** el límite aplicado es 200

#### Scenario: exclude multilínea con coma final en la propia sección

- **GIVEN** un `pyproject.toml` con `[tool.linelens]` y `exclude = [` en varias líneas terminando en `"vendor",` seguido de `]`
- **WHEN** `linelens` carga la configuración
- **THEN** `exclude` vale `["…", "vendor"]` sin error

#### Scenario: array de inline tables multilínea

- **GIVEN** un `pyproject.toml` con `authors = [` y una inline table `{ name = "x" }` por línea
- **WHEN** cualquiera de los cinco tools carga la configuración
- **THEN** la carga no produce error

#### Scenario: coma final en inline table sigue siendo inválida

- **GIVEN** un `pyproject.toml` con `[tool.linelens]` y `default = { maxLines = 10, }`
- **WHEN** `linelens` carga la configuración
- **THEN** la carga produce un error que nombra la clave `default`

### Requirement: Formas léxicas adicionales del subset

El parser SHALL reconocer, además de las formas ya soportadas:

- **Literal strings** `'…'`, sin procesar secuencias de escape.
- **Strings multilínea** `"""…"""` y `'''…'''`, descartando el salto de línea inmediatamente
  posterior al delimitador de apertura.
- **Enteros** con separador `_` (`1_000`) y con prefijo `0x`, `0o` o `0b`.
- **Fechas y datetimes** RFC 3339, que SHALL representarse como **string** sin interpretación
  semántica.
- **Dotted keys** (`a.b = 1`), que SHALL materializar las tablas intermedias.
- **Quoted keys** (`"a.b" = 1`, `'a.b' = 1`), cuyo contenido SHALL tomarse literal, sin dividir por
  el punto.

#### Scenario: literal string en la sección del tool

- **GIVEN** un `pyproject.toml` con `[tool.linelens]` y `exclude = ['vendor/**']`
- **WHEN** `linelens` carga la configuración
- **THEN** `exclude` vale `["vendor/**"]`

#### Scenario: dotted key dentro de la sección del tool

- **GIVEN** un `pyproject.toml` con `[tool.linelens]` y `default.maxLines = 42`
- **WHEN** `linelens` carga la configuración
- **THEN** el límite aplicado es 42

#### Scenario: quoted key no se divide por el punto

- **GIVEN** un `pyproject.toml` con `[tool.mypy]` y `"module.sub" = 1`
- **WHEN** se extrae la tabla `tool.mypy`
- **THEN** la clave resultante es `module.sub` y no una tabla anidada

#### Scenario: fecha como string

- **GIVEN** un `pyproject.toml` con `[tool.linelens]` y `since = 2026-01-01`
- **WHEN** se extrae la tabla `tool.linelens`
- **THEN** `since` vale la cadena `2026-01-01`

### Requirement: Sintaxis no soportada fuera de la sección objetivo se ignora

`ExtractAsJSON(toml, sectionPath)` MUST NOT fallar por sintaxis que no pertenezca a `sectionPath`
ni a sus sub-tablas: una asignación no parseable fuera de esa sección SHALL descartarse y el parseo
SHALL continuar con el resto del documento.

Dentro de `sectionPath` (y de sus sub-tablas) el comportamiento SHALL ser el contrario: una
asignación no parseable SHALL producir error, con el número de línea y el nombre de la clave, para
que el usuario se entere de que su propia configuración está mal.

Un encabezado de tabla malformado SHALL seguir siendo error en cualquier posición del documento,
porque impide determinar a qué sección pertenecen las líneas siguientes.

#### Scenario: TOML exótico en una sección ajena no rompe la carga

- **GIVEN** un `pyproject.toml` con `[tool.poetry.dependencies]` que usa sintaxis fuera del subset
- **AND** una tabla `[tool.dupelens]` válida
- **WHEN** `dupelens` carga la configuración
- **THEN** la carga no produce error
- **AND** los valores de `[tool.dupelens]` se aplican

#### Scenario: sintaxis inválida en la sección del propio tool falla ruidosa

- **GIVEN** un `pyproject.toml` con `[tool.dupelens]` y una asignación no parseable
- **WHEN** `dupelens` carga la configuración
- **THEN** la carga produce error
- **AND** el mensaje nombra la línea y la clave ofensora

#### Scenario: encabezado de tabla malformado sigue siendo error

- **GIVEN** un `pyproject.toml` con un encabezado `[tool.linelens` sin cerrar
- **WHEN** `linelens` carga la configuración
- **THEN** la carga produce error

#### Scenario: pyproject.toml reales de proyectos Python

- **GIVEN** los `pyproject.toml` de referencia del banco de pruebas (estilo poetry, setuptools, hatch y uv)
- **WHEN** se extrae `tool.linelens` de cada uno
- **THEN** ninguna extracción produce error
