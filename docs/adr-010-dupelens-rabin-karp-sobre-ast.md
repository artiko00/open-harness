## ADR-010: Rabin-Karp sobre AST para `dupelens`

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
