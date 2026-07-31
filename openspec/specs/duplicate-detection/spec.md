# duplicate-detection Specification

## Purpose
TBD - created by archiving change fix-audit-findings. Update Purpose after archive.
## Requirements
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

### Requirement: Las declaraciones de import no cuentan como código

El tokenizador SHALL descartar las declaraciones de import, include y re-export antes de calcular
fingerprints, del mismo modo que ya descarta comentarios y contenido de strings: son sintaxis
obligatoria de acceso a otro módulo, no lógica. El comportamiento SHALL estar activo por defecto y
SHALL poder desactivarse con `default.ignoreImports: false`.

El reconocimiento SHALL ser por familia de lenguaje sobre la extensión del archivo, sin parsear la
gramática, y SHALL cubrir las declaraciones multilínea completas, no solo su primera línea. El
descarte NO MUST alterar la numeración de líneas del reporte.

#### Scenario: Dos cabeceras de imports no son un duplicado

- **GIVEN** dos archivos `.ts` cuya única coincidencia son 12 declaraciones `import … from …` de
  módulos distintos, con cuerpos de lógica sin relación
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** no se reporta ningún duplicado

#### Scenario: La lógica duplicada se sigue detectando

- **GIVEN** dos archivos `.ts` con cabeceras de imports distintas y un bloque de lógica idéntico de
  60 tokens
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el bloque de lógica se reporta como duplicado

#### Scenario: Import multilínea de JS/TS se descarta completo

- **GIVEN** un archivo `.ts` con `import {` seguido de doce identificadores en líneas propias y
  `} from './modulo';`
- **WHEN** se tokeniza el archivo
- **THEN** ninguno de los identificadores de la lista aparece entre los tokens

#### Scenario: Bloque import de Go se descarta completo

- **GIVEN** un archivo `.go` con `import (` seguido de ocho rutas y `)`
- **WHEN** se tokeniza el archivo
- **THEN** ningún token proviene del bloque de import

#### Scenario: from … import multilínea de Python se descarta completo

- **GIVEN** un archivo `.py` con `from paquete.modulo import (` seguido de nombres en líneas propias
  y `)`
- **WHEN** se tokeniza el archivo
- **THEN** ningún token proviene de la declaración

#### Scenario: La numeración de líneas no se corre

- **GIVEN** un archivo con 10 líneas de import y una función que empieza en la línea 12
- **WHEN** se reporta un duplicado de esa función
- **THEN** el rango de líneas del match empieza en 12

#### Scenario: El using de recurso de C# no es un import

- **GIVEN** un archivo `.cs` con `using (var stream = File.OpenRead(path))` dentro de un método
- **WHEN** se tokeniza el archivo
- **THEN** la línea se conserva como código

#### Scenario: ignoreImports false restituye el comportamiento anterior

- **GIVEN** una config con `"default": { "ignoreImports": false }`
- **WHEN** se tokeniza un archivo con declaraciones de import
- **THEN** los tokens de import aparecen en el análisis

### Requirement: Desglose de hallazgos por tipo en el reporte

El reporte SHALL indicar cuántos hallazgos son `exact` y cuántos `renamed` en su encabezado y en su
línea de resumen, de modo que se entienda sin leer el detalle por qué `--fail` pasa o falla. La
salida JSON SHALL exponer ambos conteos como campos propios.

#### Scenario: Encabezado y resumen desglosados

- **GIVEN** un árbol con 1 hallazgo `exact` y 3 `renamed`
- **WHEN** se ejecuta `dupelens check --dir . --no-color`
- **THEN** el encabezado `DUPLICATES` indica `1 exact` y `3 renamed`
- **AND** la línea `SUMMARY` indica el mismo desglose

#### Scenario: Conteos en el JSON

- **GIVEN** el mismo árbol
- **WHEN** se ejecuta `dupelens check --dir . --format=json`
- **THEN** el objeto raíz contiene `exactCount: 1` y `renamedCount: 3`

### Requirement: Los bloques de baja entropía no producen clones renombrados

La pasada de detección `renamed` SHALL descartar las ventanas cuyo bloque de origen sea repetitivo:
aquellas en las que una proporción alta de sus líneas comienza con el mismo token. La evaluación
SHALL hacerse sobre los tokens crudos, no sobre los normalizados, donde casi todo código real es de
baja entropía por construcción.

El filtro NO MUST aplicarse a la pasada `exact`: allí la igualdad literal ya es señal genuina, y el
gate por defecto de `--fail` depende exclusivamente de esa pasada.

#### Scenario: Un array de datos embebido no genera match renamed

- **GIVEN** dos archivos con arrays de objetos literales de la misma forma y valores distintos,
  una entrada por línea
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** no se reporta ningún hallazgo `renamed` entre esos bloques

#### Scenario: Un bloque literal repetitivo se sigue reportando como exact

- **GIVEN** dos archivos con un `switch` de veinte `case` byte-idéntico
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el par se reporta con la etiqueta `exact`

