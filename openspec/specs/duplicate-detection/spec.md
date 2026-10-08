# duplicate-detection Specification

## Purpose

Define cómo dupelens detecta código duplicado: qué cuenta como código, la detección de clones exact y renamed sobre ventanas de tokens, cómo se arman y reportan los hallazgos, y el presupuesto de memoria dentro del cual corre cada escaneo.

## Requirements

### Requirement: Techo de memoria proporcional al fuente

El fingerprinting NO MUST retener una copia de la ventana de tokens por cada posición. El consumo de
memoria residente SHALL crecer de forma proporcional al tamaño del fuente, no al producto del
tamaño por el tamaño de ventana.

El consumo de memoria NO MUST crecer con el cuadrado del número de copias de un mismo bloque: un
bloque repetido k veces SHALL costar memoria y tiempo proporcionales a k.

#### Scenario: 24 MB de fuente bajo 512 MB de RSS

- **GIVEN** un árbol con 24 MB de código en 1600 archivos
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el RSS máximo del proceso se mantiene por debajo de 512 MB

#### Scenario: Muchas copias del mismo paquete no disparan la memoria

- **GIVEN** un árbol con 32 copias idénticas de un paquete de 0,5 MB (16 MB en total)
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el RSS máximo del proceso se mantiene por debajo de 1 GiB
- **AND** el exit code no es `2`

#### Scenario: Duplicar las copias duplica el costo, no lo cuadruplica

- **GIVEN** dos árboles con 16 y 32 copias idénticas del mismo paquete
- **WHEN** se ejecuta `dupelens check --dir .` sobre cada uno
- **THEN** el RSS máximo del segundo es menor que 2,5 veces el del primero

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

### Requirement: Hallazgos anclados a la primera ocurrencia

Cuando un mismo bloque aparece en más de dos ubicaciones, cada ocurrencia SHALL reportarse una sola
vez, emparejada con la primera ocurrencia en orden canónico (ruta del archivo y, dentro del mismo
archivo, posición). Los pares entre ocurrencias no canónicas NO MUST reportarse: quedan implícitos
por transitividad.

El resultado de `--fail` (exit `0` o `1`) SHALL ser el mismo que si se reportaran todos los pares.

#### Scenario: Tres copias idénticas producen dos hallazgos

- **GIVEN** los archivos `a.py`, `b.py` y `c.py` con el mismo bloque byte-idéntico
- **WHEN** se ejecuta `dupelens check --dir . --format=json`
- **THEN** se reportan exactamente los pares `a.py`-`b.py` y `a.py`-`c.py`
- **AND** `exactCount` es 2

#### Scenario: Un bloque compartido solo por las copias no canónicas se reporta

- **GIVEN** `a.py`, `b.py` y `c.py` con un bloque común X
- **AND** un segundo bloque Y presente solo en `b.py` y `c.py`
- **WHEN** se ejecuta `dupelens check --dir . --format=json`
- **THEN** el bloque Y se reporta como el par `b.py`-`c.py`

#### Scenario: El gate no cambia con el anclaje

- **GIVEN** un árbol cuyos únicos hallazgos son tres copias `exact` del mismo bloque
- **WHEN** se ejecuta `dupelens check --dir . --fail`
- **THEN** el exit code es 1

#### Scenario: Un cambio de ancla a mitad del bloque no lo fragmenta

- **GIVEN** `m.py` y `n.py` con el mismo bloque de 10 tokens
- **AND** `a.py` con solo los primeros 5 tokens de ese bloque
- **WHEN** se ejecuta la detección con ventana 3 y `minTokens` 10
- **THEN** el par `m.py`-`n.py` se reporta con el bloque completo

### Requirement: El conteo de tokens refleja el tramo duplicado

El conteo de tokens de un hallazgo SHALL ser el número de tokens que abarca el bloque duplicado en
el primer archivo del par. NO MUST sumar un token por cada par de ventanas fusionado: cuando se
fusionan coincidencias de alineaciones distintas, el conteo no puede superar el tramo que cubren.
Un bloque cuyo tramo real es menor que `minTokens` NO MUST reportarse.

#### Scenario: Código periódico no infla el conteo

- **GIVEN** un archivo con la secuencia `p q r s` repetida cuatro veces, un token por línea
- **WHEN** se ejecuta la detección con ventana 3
- **THEN** ningún hallazgo declara más tokens que líneas abarca

#### Scenario: Un bloque real por debajo del umbral no se reporta

- **GIVEN** dos firmas de función que comparten 31 tokens contiguos y difieren después
- **WHEN** se ejecuta `dupelens check --dir .` con `minTokens` 50
- **THEN** el par no se reporta

### Requirement: Presupuesto de memoria configurable

dupelens SHALL aplicar un presupuesto de memoria a cada ejecución de `check`. El presupuesto SHALL
poder expresarse como valor absoluto (entero positivo seguido de `MiB` o `GiB`) o como porcentaje entero entre
1 y 100 seguido de `%`, calculado sobre la memoria disponible al arrancar.

La memoria disponible SHALL ser la del sistema y, si el proceso corre en un cgroup con límite de
memoria, el menor entre esa cifra y el margen restante del cgroup.

Sin configuración explícita SHALL aplicarse `auto`: el menor entre 1 GiB y el 25 % de la memoria
disponible. Si la memoria disponible no puede determinarse, `auto` SHALL usar 1 GiB y un porcentaje
explícito SHALL caer a 1 GiB con un aviso en stderr.

El presupuesto SHALL resolverse con esta precedencia: flag `--max-memory`, variable de entorno
`DUPELENS_MAX_MEMORY`, clave `maxMemory` de la cadena de configuración, `auto`. Un valor explícito
SHALL aplicarse tal cual, sin combinarse con el default.

Un valor con formato inválido SHALL ser un error con exit code 1, igual que el resto de los flags y
claves inválidos.

#### Scenario: Default automático en una máquina grande

- **GIVEN** una máquina con 28 GiB de memoria disponible y ninguna configuración de presupuesto
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el presupuesto aplicado es 1 GiB

#### Scenario: Default automático con poca memoria disponible

- **GIVEN** una máquina con 2 GiB de memoria disponible y ninguna configuración de presupuesto
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el presupuesto aplicado es 512 MiB

#### Scenario: El límite del cgroup acota el default

- **GIVEN** un cgroup con `memory.max` de 1 GiB y 600 MiB ya en uso
- **WHEN** se ejecuta `dupelens check --dir .` dentro de ese cgroup
- **THEN** el presupuesto aplicado es el 25 % de los 424 MiB restantes

#### Scenario: Porcentaje explícito

- **GIVEN** una máquina con 8 GiB de memoria disponible
- **WHEN** se ejecuta `dupelens check --dir . --max-memory 40%`
- **THEN** el presupuesto aplicado es el 40 % de 8 GiB, sin el tope de 1 GiB

#### Scenario: El flag gana sobre la variable de entorno y la config

- **GIVEN** `DUPELENS_MAX_MEMORY=2GiB` en el entorno
- **AND** un `pyproject.toml` con `[tool.dupelens]` y `maxMemory = "3GiB"`
- **WHEN** se ejecuta `dupelens check --dir . --max-memory 512MiB`
- **THEN** el presupuesto aplicado es 512 MiB

#### Scenario: La variable de entorno gana sobre la config

- **GIVEN** `DUPELENS_MAX_MEMORY=2GiB` en el entorno
- **AND** un `dupelens.json` con `"maxMemory": "3GiB"`
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el presupuesto aplicado es 2 GiB

#### Scenario: La clave se lee desde cualquier archivo de la cadena

- **GIVEN** un `package.json` con `{"dupelens": {"maxMemory": "768MiB"}}` y ningún otro archivo de config
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el presupuesto aplicado es 768 MiB

#### Scenario: Valor inválido

- **WHEN** se ejecuta `dupelens check --dir . --max-memory mucho`
- **THEN** stderr explica el formato válido (`<n>MiB`, `<n>GiB` o `<n>%`)
- **AND** el exit code es 1

#### Scenario: Memoria disponible indeterminable

- **GIVEN** una plataforma donde la memoria disponible no puede determinarse
- **WHEN** se ejecuta `dupelens check --dir . --max-memory 40%`
- **THEN** stderr avisa que se aplica 1 GiB
- **AND** el presupuesto aplicado es 1 GiB

### Requirement: Exceder el presupuesto de memoria es una medición fallida

Cuando el consumo de memoria del escaneo supera el presupuesto, dupelens SHALL detener el análisis
y terminar con exit code `2` ("no se pudo medir"), con o sin `--fail`. Stdout NO MUST contener un
reporte parcial. Stderr SHALL nombrar el presupuesto aplicado, de dónde salió (flag, variable de
entorno, config o `auto`) y cómo subirlo o reducir el alcance del escaneo.

El proceso NO MUST superar el presupuesto en más de lo que se asigne entre dos puntos de control
consecutivos.

#### Scenario: Escaneo que no entra en el presupuesto

- **GIVEN** un árbol cuyo escaneo requiere más de 64 MiB
- **WHEN** se ejecuta `dupelens check --dir . --max-memory 64MiB --fail`
- **THEN** el exit code es 2
- **AND** stdout está vacío
- **AND** stderr menciona `64 MiB`, que el valor vino de `--max-memory` y las opciones `--max-memory`,
  `DUPELENS_MAX_MEMORY`, `maxMemory` y `exclude`

#### Scenario: Sin --fail el exceso también es exit 2

- **GIVEN** el mismo árbol
- **WHEN** se ejecuta `dupelens check --dir . --max-memory 64MiB`
- **THEN** el exit code es 2

#### Scenario: Un escaneo dentro del presupuesto no cambia

- **GIVEN** un árbol cuyo escaneo requiere menos que el presupuesto
- **WHEN** se ejecuta `dupelens check --dir .`
- **THEN** el reporte y el exit code son los mismos que sin presupuesto
