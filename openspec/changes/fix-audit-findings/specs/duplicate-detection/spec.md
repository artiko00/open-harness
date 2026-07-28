# duplicate-detection

Motor de detección de duplicados de dupelens 0.2.1. Detecta clones literales con precisión, pero es
ciego al copy-paste con renombre de variables y retiene ~100× el tamaño del fuente en memoria
(2.370 MB con 24 MB de código), lo que produce OOM en runners de CI.

## ADDED Requirements

### Requirement: Techo de memoria proporcional al fuente

El fingerprinting NO MUST retener una copia de la ventana de tokens por cada posición. El consumo de
memoria residente SHALL crecer de forma proporcional al tamaño del fuente, no al producto del
tamaño por el tamaño de ventana.

#### Scenario: 24 MB de fuente bajo 512 MB de RSS

- **GIVEN** un árbol con 24 MB de código en 1600 archivos
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el RSS máximo del proceso se mantiene por debajo de 512 MB

#### Scenario: La verificación literal se conserva

- **GIVEN** dos bloques con el mismo hash pero contenido distinto
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el par no se reporta como duplicado

### Requirement: Detección de clones con identificadores renombrados

El motor SHALL detectar bloques estructuralmente idénticos cuyos identificadores difieren,
normalizando identificadores y literales numéricos antes de calcular el fingerprint. Cada hallazgo
SHALL etiquetarse como `exact` o `renamed`.

Por defecto, `--fail` SHALL considerar únicamente los hallazgos `exact`. Un flag SHALL permitir
incluir los `renamed` en el gate.

#### Scenario: Copy-paste con variables renombradas

- **GIVEN** dos funciones de 21 líneas idénticas en estructura, una con `orders/order/results` y otra con `invoices/inv/output`
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el par se reporta con la etiqueta `renamed`

#### Scenario: Copia literal se sigue etiquetando exact

- **GIVEN** dos archivos con un bloque byte-idéntico
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el par se reporta con la etiqueta `exact`

#### Scenario: Los renombrados no rompen el gate por defecto

- **GIVEN** un árbol cuyo único hallazgo es de tipo `renamed`
- **WHEN** se ejecuta `dupelens check --dir . --fail`
- **THEN** el exit code es 0
- **AND** el hallazgo aparece en el reporte

### Requirement: Sensibilidad separada del umbral de reporte

El tamaño de ventana del fingerprint SHALL configurarse de forma independiente del umbral mínimo de
tokens para reportar. Reducir el ruido NO MUST requerir volver ciego al detector.

#### Scenario: Umbral alto con ventana fina

- **GIVEN** una configuración con ventana de 50 tokens y umbral de reporte de 200
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** solo se reportan bloques de 200 tokens o más
- **AND** la detección de esos bloques usa ventanas de 50 tokens

### Requirement: Alcance limitado a archivos de código

El análisis SHALL restringirse por defecto a extensiones de código. Datos, fixtures, catálogos de
traducción y lockfiles NO MUST reportarse como duplicación de código. La lista SHALL ser configurable.

#### Scenario: Fixtures de datos no se reportan

- **GIVEN** un árbol con `data.csv`, `data2.csv`, `fixtures_a.json` y `fixtures_b.json`, con contenido repetido
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** no se reporta ningún duplicado

#### Scenario: Los lockfiles quedan fuera

- **GIVEN** un árbol con dos lockfiles de contenido similar
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** no se reporta ningún duplicado

### Requirement: Comentarios sensibles al lenguaje

El stripper de comentarios SHALL aplicar la sintaxis de comentarios del lenguaje del archivo. El
carácter `#` NO MUST tratarse como inicio de comentario en lenguajes donde no lo es.

#### Scenario: Atributos de Rust se preservan

- **GIVEN** un archivo `.rs` con `#[derive(Debug)]`
- **WHEN** se tokeniza el archivo
- **THEN** el atributo no se descarta como comentario

#### Scenario: Colores CSS se preservan

- **GIVEN** un archivo `.css` con `color: #fff;`
- **WHEN** se tokeniza el archivo
- **THEN** el valor no se descarta como comentario

### Requirement: Snippet legible en el reporte

El reporte de consola SHALL presentar los fragmentos de ambos archivos separados y etiquetados con
su ruta, de modo que no se lean como un bloque continuo.

#### Scenario: Fragmentos identificables

- **GIVEN** un duplicado entre `a.go` y `c.go`
- **WHEN** se ejecuta `dupelens check --dir . --no-color`
- **THEN** el reporte muestra el fragmento de cada archivo bajo su propia ruta
