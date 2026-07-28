# secret-detection Specification

## Purpose
TBD - created by archiving change fix-audit-findings. Update Purpose after archive.
## Requirements
### Requirement: Asignaciones sin comillas

La regla genérica de asignación de secretos SHALL detectar valores tanto entrecomillados como sin
comillas, de modo que un archivo `.env` en su formato habitual produzca hallazgos.

#### Scenario: Par KEY=VALUE sin comillas

- **GIVEN** un archivo `.env` con la línea `API_KEY=aB3xK9mNoPqRsTuVwXyZ01`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** la línea se reporta como hallazgo

#### Scenario: El caso entrecomillado sigue funcionando

- **GIVEN** un archivo con `API_KEY="aB3xK9mNoPqRsTuVwXyZ01"`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** la línea se reporta como hallazgo

#### Scenario: Valor de baja entropía no genera ruido

- **GIVEN** un archivo `.env` con las líneas `DEBUG=true` y `LOG_LEVEL=verbose`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** ninguna de las dos se reporta

### Requirement: Filtro por entropía

El motor SHALL calcular la entropía de Shannon del valor detectado y SHALL descartar los valores
por debajo de un umbral configurable, cuyo default SHALL estar calibrado contra el fixture de
auditoría. El umbral SHALL ser configurable mediante la clave `minEntropy`.

#### Scenario: Valor de alta entropía se reporta

- **GIVEN** una línea `SECRET=Xk9mQ2vB7nR4tL8wZ1cF5jH3sD6gY0pA`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** la línea se reporta como hallazgo

#### Scenario: Palabra de diccionario no se reporta

- **GIVEN** una línea `PASSWORD_HINT=favoritecolorbluegreen`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** la línea no se reporta

### Requirement: Prefijos de proveedor

El conjunto de patterns built-in SHALL incluir los prefijos de credencial de los proveedores de
mayor impacto: `sk_live_`, `xoxb-`, `xoxp-`, `AIza`, `sk-proj-`, `glpat-`, `npm_`, `SG.`,
`hooks.slack.com/services/`, y URIs con credenciales embebidas (`postgres://`, `mongodb+srv://`,
`redis://`).

#### Scenario: Clave viva de Stripe

- **GIVEN** una línea `STRIPE_KEY=sk_live_<REDACTED-EN-DOC>`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** se reporta un hallazgo de severidad `critical`

#### Scenario: URI con contraseña embebida

- **GIVEN** una línea `DB=postgres://admin:Sup3rS3cr3tP4ss@db.prod.io:5432/main`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** se reporta un hallazgo

#### Scenario: Recall sobre el fixture de auditoría

- **GIVEN** el fixture de auditoría con 20 secretos reales
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** se detectan al menos 18 de los 20

### Requirement: Allowlist aplicada al valor detectado

La allowlist SHALL evaluarse contra el valor que produjo el match, NO MUST evaluarse contra la línea
completa. El término `example` SHALL retirarse de la allowlist por defecto.

#### Scenario: Un comentario no suprime el hallazgo

- **GIVEN** la línea `AWS_KEY = "AKIAIOSFODNN7EXAMPLE"  # see example above`
- **WHEN** se ejecuta `secretlens check --dir . --fail`
- **THEN** la clave de AWS se reporta
- **AND** el exit code es 1

#### Scenario: Un placeholder real sí se suprime

- **GIVEN** la línea `API_KEY=your_key_here`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** no se reporta ningún hallazgo

#### Scenario: Credencial contra un host de ejemplo

- **GIVEN** la línea `DB = "postgres://u:SuperSecretPass99@example.com/db"`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** la credencial se reporta

### Requirement: Patterns custom aditivos

Los patterns declarados por el usuario SHALL sumarse a los built-in. La desactivación de los
built-in SHALL requerir la clave explícita `disableDefaultPatterns: true`.

#### Scenario: Un pattern propio no desactiva los built-in

- **GIVEN** un `secretlens.json` con un único pattern custom
- **AND** un archivo con una AWS Access Key ID y un match del pattern custom
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** se reportan ambos hallazgos

#### Scenario: Opt-out explícito

- **GIVEN** un `secretlens.json` con `disableDefaultPatterns: true` y un pattern custom
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** solo se evalúa el pattern custom

### Requirement: Secretos repartidos en varias líneas

El motor SHALL detectar asignaciones donde la clave y el valor quedan en líneas distintas, como
ocurre en JSON formateado. El alcance SHALL limitarse a un salto de línea entre clave y valor, para
no penalizar el rendimiento del escaneo línea a línea.

#### Scenario: JSON pretty-printed

- **GIVEN** un JSON donde `"secretAccessKey":` y su valor están en líneas consecutivas
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** se reporta un hallazgo en la línea del valor

### Requirement: Decodificación de contenido UTF-16

secretlens SHALL decodificar el contenido de archivos en UTF-16 (BOM `FF FE` o `FE FF`) antes de
aplicar las reglas, de modo que un secreto en un `.env` generado en UTF-16 se detecte. El módulo
`path-matching` ya garantiza que esos archivos no se descarten como binarios; este requisito cubre
la decodificación del contenido.

#### Scenario: Secreto en un .env UTF-16

- **GIVEN** un archivo `.env` en UTF-16 con la línea `AWS=AKIAIOSFODNN7EXAMPLE`
- **WHEN** se ejecuta `secretlens check --dir .`
- **THEN** la clave de AWS se reporta como hallazgo

