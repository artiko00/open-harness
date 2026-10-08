## ADR-012: Rabin-Karp sobre AST para `dupelens`

**Estado:** Aceptado
**Fecha:** 2026-05-02

## Contexto

Se necesita un detector de código duplicado dentro del monorepo open-harness, que pueda integrarse en pre-commit hooks junto a `linelens`. La herramienta debe:

- Funcionar en cualquier lenguaje (Go, JS, TS, Python, Rust, etc.) — coherente con la filosofía de [ADR-001](adr-001-go-sobre-node.md)
- Mantener el principio de **cero dependencias** ([ADR-002](adr-002-cero-dependencias.md))
- Ejecutar rápido en pre-commit
- Producir resultados accionables (par de archivos + rangos de líneas duplicados)

Existen dos familias principales de algoritmos para detección de duplicados:

| Familia | Ejemplos | Cómo funciona |
|---|---|---|
| **Token-based (fingerprinting)** | jscpd, PMD CPD | Tokeniza el código y aplica rolling hash sobre ventanas de N tokens. Encuentra duplicados literales o casi literales. |
| **AST-based (estructural)** | jsinspect, Semgrep | Parsea el código a AST y compara subárboles. Detecta duplicados estructurales aunque cambien nombres de variables. |

## Decisión

Se eligió **Rabin-Karp sobre tokens** (token-based fingerprinting).

| Criterio | Token-based (Rabin-Karp) | AST-based |
|---|---|---|
| Cero dependencias | ✅ stdlib pura | ❌ requiere parser por lenguaje |
| Language-agnostic | ✅ tokenizador genérico | ❌ uno por lenguaje |
| Velocidad | ✅ ~O(n) lineal | ⚠️ parser + comparación de árboles |
| Detecta refactor con renames | ❌ literal/casi literal | ✅ estructural |
| Líneas de código | ~250-350 | ~1500+ por lenguaje soportado |
| Encaja con la filosofía del repo | ✅ | ❌ |

El tokenizador genérico (split por espacios y puntuación + filtro de tokens triviales) cubre el 90% de los casos prácticos sin parsear gramática específica. Lo que se pierde — duplicados estructurales con variables renombradas — se cubre con `bigo` (planificado, [F-001](.agent/feature-list.json)) o agregando un detector AST específico por lenguaje en una iteración futura.

## Consecuencias

**Positivo:**
- Implementación en stdlib Go pura (~250-350 líneas, alineado con la regla de 100 líneas por archivo aplicada al monorepo)
- Soporta cualquier lenguaje sin escribir parser
- Velocidad comparable a `jscpd` (mismo algoritmo) sin overhead de runtime Node
- Extensible: el tokenizador puede mejorarse por lenguaje sin cambiar el core

**Negativo:**
- No detecta duplicados con renames de variables (limitación inherente del enfoque token-based)
- Falsos positivos posibles si los thresholds (`minTokens`, `minLines`) están mal calibrados
- Tokenizador agnóstico es menos preciso que un lexer real (ej. no distingue strings de identifiers)

**Neutral:**
- Si en el futuro se requiere detección estructural, se puede agregar un comando `dupelens check --mode ast` que use parsers específicos por lenguaje, sin romper el core token-based.
- La decisión es alineada con cómo `jscpd` aborda el problema — referencia probada en producción con 150+ lenguajes.

## Extensión: normalización de identificadores (clones Type-2) sin parser por lenguaje

El "Negativo" original decía que el enfoque token-based no detecta duplicados con renames de variables. Esa limitación se levanta parcialmente **sin introducir un parser por lenguaje**, aplicando una normalización léxica sobre el mismo stream de tokens que ya produce el tokenizador.

### Qué es un clon Type-2

Taxonomía estándar de clones:

- **Type-1 (exacto):** copia literal, salvo espacios y comentarios.
- **Type-2 (renamed):** misma estructura, pero con identificadores, tipos y literales renombrados (`total := a + b` vs `sum := x + y`).

Rabin-Karp sobre el stream crudo solo cubre Type-1. Para cubrir Type-2 se agrega una **segunda pasada normalizada** sobre el mismo stream.

### Normalización léxica (`normalizeTokens`)

Cada token se reescribe a una forma canónica antes de fingerprintar:

- **Identificadores → `ID`** (empieza con letra o `_`, resto alfanumérico o `_`).
- **Literales numéricos → `NUM`** (empieza con dígito).
- **Keywords y operadores → se conservan tal cual.**

La distinción keyword vs identificador es una **lista de keywords** por familia sintáctica, no un lexer gramatical: no se parsea el lenguaje, solo se clasifica cada token contra un conjunto conocido. Es una heurística léxica O(n), coherente con la filosofía de cero parsers de este ADR. Así, `total := a + 1` y `sum := x + 2` colapsan ambos a `ID := ID + NUM`, produciendo el mismo fingerprint y detectándose como clon Type-2.

### Dos pasadas sobre el mismo stream

`dupelens` corre **dos** fingerprintings sobre idéntico stream de tokens:

1. **Pasada exacta:** tokens crudos → captura Type-1.
2. **Pasada renamed:** tokens normalizados (`ID`/`NUM`) → captura Type-2.

Se preserva la línea de cada token en ambas pasadas para reportar rangos accionables. La normalización aumenta la sensibilidad, por lo que se filtran las **ventanas monótonas** (todas iguales, p. ej. una tira de `ID ID ID …` tras normalizar tablas de datos o mapas de keywords): no aportan señal estructural y dispararían falsos positivos en contenido repetitivo. Ese filtro se aplica en ambas pasadas.

### Trade-off

- **Positivo:** se detectan refactors con renames (Type-2) sin escribir un parser por lenguaje; solo una lista de keywords y una clasificación léxica.
- **Negativo:** la clasificación es heurística; un token ambiguo entre keyword e identificador en un lenguaje no cubierto por la lista puede normalizarse distinto de lo ideal. El costo es a lo sumo un match Type-2 perdido o de más, nunca un fallo del core Type-1.
- **Neutral:** no cubre Type-3 (clones con líneas insertadas/borradas) ni Type-4 (equivalencia semántica); eso sigue requiriendo AST o análisis de flujo, fuera del alcance de este ADR.

## Extensión: semillas ancladas, extensión por diagonal y hash de 61 bits (0.5.0)

El Rabin-Karp de este ADR se conserva; cambia qué se hace con las ventanas coincidentes. Hasta 0.4.x,
`detect` enumeraba **todos los pares** de cada grupo de hash y creaba un hallazgo por ventana y por par
antes de fusionar. Un bloque de L ventanas copiado k veces costaba L·k(k−1)/2 objetos: 16 MB de código
generado superaban 3 GB. Ver [ADR-024](adr-024-dupelens-presupuesto-de-memoria.md) para el presupuesto
de memoria que acompaña este cambio.

- **Semillas ancladas.** Dentro de cada clase de ventanas idénticas (verificadas literalmente), cada
  ocurrencia se empareja solo con la primera en orden canónico (ruta, posición): k−1 semillas. Como
  cada ventana aporta como mucho una semilla, el total queda acotado por el número de ventanas.
- **Extensión por diagonal.** Cada semilla se extiende hacia atrás y hacia adelante sobre su diagonal
  (`idxB − idxA`) mientras los tokens coincidan y las ventanas pasen los filtros de ruido. Sin esto, un
  archivo anterior en el orden que comparte solo parte de un bloque cambia el ancla a mitad de camino
  y deja al par en pedazos por debajo de `minTokens`.
- **Conteo de tokens = tramo real.** La fusión por líneas entre corridas se conserva, pero el conteo
  es el tramo que abarcan en A, no `windowSize + pares − 1`, que en código periódico inflaba bloques
  que en ninguna alineación eran contiguos.
- **Hash módulo 2^61−1** (con `math/bits.Mul64`) en lugar de `1e9+7`: con cientos de millones de
  ventanas, ~30 bits producían millones de colisiones falsas. La verificación literal sigue igual.
- **Normalización por vocabulario.** `normalizeTokens` ya no copia el stream: cada token distinto se
  interna una vez con su hash y el id de su forma normalizada, y la vista `renamed` se deriva por tabla.

**Consecuencia en el reporte:** con tres o más copias se reportan los pares contra la primera (`a-b`,
`a-c`; `b-c` queda implícito). El resultado de `--fail` no cambia.
