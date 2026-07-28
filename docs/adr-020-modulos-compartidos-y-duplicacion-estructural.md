# ADR-020: Módulos compartidos y duplicación estructural aceptada

**Estado:** Aceptado

## Contexto

La auditoría adversarial (change `fix-audit-findings`) señaló que el motor de path matching,
la detección de binarios y los cargadores de configuración estaban copiados byte a byte entre los
cuatro tools, y que `dupelens.json` fijaba `minTokens: 200` — un umbral que ocultaba en silencio 38
bloques duplicados reales sobre el propio repositorio.

Al bajar el umbral al default (`minTokens: 50`) los duplicados quedan a la vista. Corresponde decidir
qué duplicación se elimina de raíz y cuál es una propiedad aceptada de la arquitectura.

## Decisión

### 1. Lógica genuinamente compartible → módulos en `tools/_shared/`

Se extrae a módulos compartidos (cada uno con su `go.mod`, en el `go.work`, sin dependencias externas
— ADR-002) toda la lógica que era idéntica salvo un parámetro:

- **`pathmatch`** — glob por segmento, exclusión, detección de binarios, filtro de archivos regulares.
- **`langsyntax`** — stripping de comentarios sensible al lenguaje (`StripComments(src, ext)`).
- **`configload`** — extracción genérica de la sección de config desde `package.json`,
  `pyproject.toml` y `composer.json` (`PackageJSON[T]`, `Composer[T]`, `Pyproject[T]`).

Precedente: `tomlmin`, avalado por ADR-018.

### 2. Esqueleto de CLI por tool → duplicación estructural aceptada

Cada tool es un binario nativo, independiente y sin dependencias de runtime compartidas (ADR-001,
ADR-002). Esa decisión implica que los cuatro comparten un **esqueleto estructural** con la misma
forma pero contenido específico de cada dominio:

- `main.go` / `check.go` / `check_cmd.go` / `init_cmd.go` — punto de entrada, `run()` y flags. Cada
  uno declara su propia `const version`, su conjunto de subcomandos y sus flags.
- `config.go` / `config_chain.go` — el tipo `Config` de cada tool (campos distintos) y el orquestador
  de la cadena, cuyo merge es específico de esos campos.
- `scanner.go` — el `WalkDir` (que ya delega en `pathmatch`) con el procesamiento por-archivo propio
  de cada tool (contar líneas / tokenizar / regex / mapear tests).
- `reporter_console.go` / `reporter_json.go` / `coverage*.go` — salida específica del dominio.

A `minTokens: 50` estos archivos producen ~28 coincidencias de 50–108 tokens. **No** se fuerzan a un
módulo compartido: hacerlo introduciría una capa de abstracción (registro de comandos, walker
genérico con callbacks, merge genérico) que contradice la simplicidad y la independencia de binarios
que ADR-001/002 buscan, a cambio de deduplicar un esqueleto que por diseño es paralelo.

En su lugar, `dupelens.json` los marca con reglas `skip` **explícitas y enumeradas**. Esto es
estrictamente más transparente que el `minTokens: 200` original: un lector ve exactamente qué
duplicación se acepta y por qué, mientras `dupelens` sigue vigilando la duplicación en la **lógica de
dominio** (fingerprinting, patrones de secretos, mapeo de tests, tokenización), que es donde un
copy-paste sí sería un defecto.

### 3. Shims de distribución (npm / PyPI) → excluidos como artefactos generados

Los lanzadores `npm/**/bin/*.js` y los `pypi/**/setup.py` / `pypi/**/__main__.py` son plantillas
generadas por `scripts/build-npm.sh` (y equivalentes), idénticas por-tool a propósito: solo ejecutan
el binario nativo correspondiente. No son código fuente bajo revisión de calidad, así que se marcan
`skip` igual que `node_modules` o `vendor`.

## Consecuencias

- `dupelens.json` vuelve a `minTokens: 50` (el default honesto) y el repo pasa `dupelens check --fail`.
- Las reglas `skip` del esqueleto documentan la duplicación aceptada; cualquier duplicación nueva en
  archivos de lógica de dominio sí rompe el gate.
- Si en el futuro el esqueleto crece o se estabiliza, extraer un walker/dispatcher compartido queda
  como opción abierta, no como deuda urgente.
