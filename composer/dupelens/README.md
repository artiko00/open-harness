# open-harness/dupelens (Composer)

Code duplication detector. Uses **Rabin-Karp** rolling-hash fingerprinting over tokenized source — strings, comments and import declarations are stripped before hashing to reduce false positives. Detects **exact** (token-for-token) and **renamed** (same structure, different identifiers) clones. Language-agnostic (Go, TS, JS, Python, Rust, Java, etc.). Single native binary, zero runtime dependencies, with a **bounded memory budget**.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

> **Same tool, other ecosystems**: also available on **npm** ([`@open_harness/dupelens`](https://www.npmjs.com/package/@open_harness/dupelens)) and **PyPI** ([`open-harness-dupelens`](https://pypi.org/project/open-harness-dupelens/)). Identical binary, identical config; pick the registry that matches your stack.

## Install

```bash
composer require --dev open-harness/dupelens
```

On install, a Composer post-install hook downloads the native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) from GitHub Releases and verifies its SHA256 checksum.

## Usage

```bash
vendor/bin/dupelens check                     # scan current directory with defaults
vendor/bin/dupelens check --fail              # exit 1 if exact duplicates found (CI / git hooks)
vendor/bin/dupelens check --fail-on all       # which kinds break --fail: exact | renamed | all
vendor/bin/dupelens check --min-tokens 30     # override the report threshold
vendor/bin/dupelens check --max-memory 2GiB   # memory budget: <n>MiB, <n>GiB or <n>% of available memory
vendor/bin/dupelens check --format=json       # JSON output for tooling integrations
vendor/bin/dupelens check --dir ./src         # scan a specific directory
vendor/bin/dupelens check --verbose           # print timings and the memory budget to stderr
vendor/bin/dupelens check --no-color          # plain console output
vendor/bin/dupelens init                      # generate a default dupelens.json
vendor/bin/dupelens version                   # print version
```

## Configuration

Place a `dupelens.json` at the repo root:

```json
{
  "default": {
    "minTokens": 50,
    "minLines": 5,
    "windowSize": 0,
    "ignoreImports": true
  },
  "rules": [
    { "pattern": "**/*_test.go",     "skip": true },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": ["node_modules", "vendor", ".git", "dist", "build"],
  "maxMemory": "2GiB"
}
```

- `minTokens` — report threshold: blocks shorter than this are not reported. `tokens` is the real span of the duplicated block.
- `minLines` — filters short matches.
- `windowSize` — detection window of the rolling hash, independent of `minTokens` (`0` = built-in default, 25).
- `ignoreImports` — drops import declarations before tokenizing (default `true`).
- `rules` — per-pattern `skip`. The first matching entry wins.
- `maxMemory` — memory budget, see below (default: `auto`).

### Alternative: configure inside `composer.json`

If you prefer not to keep a separate `dupelens.json`, add an `extra.open-harness.dupelens` section in your `composer.json` with the same shape:

```json
{
  "name": "acme/my-project",
  "extra": {
    "open-harness": {
      "dupelens": {
        "default": { "minTokens": 50, "minLines": 5 },
        "exclude": ["vendor", "node_modules"],
        "maxMemory": "2GiB"
      }
    }
  }
}
```

Precedence: `--config <path>` > `dupelens.json` > the manifest chain (`pyproject.toml` > `package.json` > `composer.json`, merged field by field) > built-in defaults. CLI flags (`--min-tokens`, `--max-memory`, etc.) always win.

## Memory budget

Every `check` runs under a memory budget, so a large repository can no longer take the whole machine down. The default, `auto`, is **the lower of 1 GiB and 25% of the available memory** — the system's available memory or, inside a cgroup with a limit (containers, CI runners), its remaining headroom. Being a share of what is *available*, several instances running in parallel each get less as memory fills up.

| Value | Meaning |
|---|---|
| *(unset)* | `auto` = min(1 GiB, 25% of available memory) |
| `"512MiB"`, `"2GiB"` | absolute budget — raise or restrict it |
| `"40%"` | percentage (1–100) of the available memory at startup |

Precedence: `--max-memory` > `DUPELENS_MAX_MEMORY` environment variable > `maxMemory` config key > `auto`. If the scan exceeds the budget, dupelens stops with **exit code 2** and no partial report, with or without `--fail`.

Memory grows linearly with the source and with the number of copies of a block: 226 MB of Python source scans in under 400 MB. With three or more copies of the same block, each copy is reported against the first one (`a-b`, `a-c`; `b-c` is implied).

## Output (console)

```
DUPLICATES (2 match(es) (1 exact · 1 renamed) found in 87 files):

  src/auth.go:42-58  <->  src/users.go:12-28  (35 tokens, exact)
  | func validate(input string) error {
  | ...
  src/db.go:1-10  <->  src/cache.go:1-10  (15 tokens, renamed)

SUMMARY: 2 match(es) (1 exact · 1 renamed) across 87 files
Top duplicated files:
  - src/auth.go  (1 match(es))
```

## Output (JSON)

```json
{
  "scannedFiles": 87,
  "matchCount": 2,
  "exactCount": 1,
  "renamedCount": 1,
  "matches": [
    {
      "fileA": "src/auth.go", "startLineA": 42, "endLineA": 58,
      "fileB": "src/users.go", "startLineB": 12, "endLineB": 28,
      "tokens": 35, "kind": "exact"
    }
  ],
  "summary": {
    "topDuplicatedFiles": [{ "file": "src/auth.go", "count": 1 }]
  }
}
```

## Integrations

```bash
# Husky pre-commit
vendor/bin/dupelens check --fail
```

```yaml
# GitHub Actions
- name: Run dupelens
  run: vendor/bin/dupelens check --fail
```

## Why Rabin-Karp over AST?

- Zero dependencies: no language-specific parsers to ship per language.
- Language-agnostic: the same binary scans Go, TypeScript, Python, Rust, Java, etc.
- Fast: rolling hash detects matches in `O(n)` over the token stream.

The trade-off is documented in [ADR-012](https://github.com/artiko00/open-harness/blob/main/docs/adr-012-dupelens-rabin-karp-sobre-ast.md); the memory budget in [ADR-024](https://github.com/artiko00/open-harness/blob/main/docs/adr-024-dupelens-presupuesto-de-memoria.md).

## Limitations (v0.5.0)

- Detects contiguous **exact** and **renamed** clones. Reordered statements, inserted or deleted lines (gapped clones) and behaviourally-equivalent rewrites are not detected — that requires AST analysis.
- The algorithm is binary (match or no match); there is no similarity threshold flag.
- Per-rule `minTokens` override does not work cross-file because window sizes must be uniform. Use `rules.skip` to exclude patterns entirely.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | No duplicates (or `--fail` not passed) |
| `1` | Duplicates found and `--fail` was passed, or usage/config error |
| `2` | Memory budget exceeded: the scan could not be completed |

---

## Español

Detector de duplicación de código. Usa fingerprinting **Rabin-Karp** (hash rodante) sobre el código tokenizado — los strings, comentarios y declaraciones de import se eliminan antes del hashing para reducir falsos positivos. Detecta clones **exact** (token a token) y **renamed** (misma estructura, identificadores distintos). Agnóstico al lenguaje. Un solo binario nativo, cero dependencias, con **presupuesto de memoria acotado**.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

### Instalación

```bash
composer require --dev open-harness/dupelens
```

Tras instalar, un hook post-install de Composer descarga el binario nativo para tu plataforma desde GitHub Releases y verifica su checksum SHA256.

### Uso

```bash
vendor/bin/dupelens check                     # escanea con defaults
vendor/bin/dupelens check --fail              # exit 1 si hay duplicados exact (CI / git hooks)
vendor/bin/dupelens check --fail-on all       # qué rompe --fail: exact | renamed | all
vendor/bin/dupelens check --min-tokens 30     # cambia el umbral de reporte
vendor/bin/dupelens check --max-memory 2GiB   # presupuesto: <n>MiB, <n>GiB o <n>% de la memoria disponible
vendor/bin/dupelens check --format=json       # salida JSON para integraciones
vendor/bin/dupelens check --dir ./src         # escanea un directorio específico
vendor/bin/dupelens check --verbose           # imprime tiempos y el presupuesto en stderr
vendor/bin/dupelens check --no-color          # consola sin colores
vendor/bin/dupelens init                      # genera un dupelens.json por defecto
vendor/bin/dupelens version                   # imprime la versión
```

### Configuración

Colocá un `dupelens.json` en la raíz del repo (ver ejemplo arriba), o una sección `extra.open-harness.dupelens` en tu `composer.json` con la misma forma. Precedencia: `--config <path>` > `dupelens.json` > la cadena de manifiestos (`pyproject.toml` > `package.json` > `composer.json`, fusionados campo por campo) > defaults. Los flags CLI siempre ganan.

- `minTokens` — umbral de reporte: los bloques más cortos no se reportan. `tokens` es el tramo real del bloque duplicado.
- `minLines` — filtra matches cortos.
- `windowSize` — ventana de detección del hash rodante, independiente de `minTokens` (`0` = default interno, 25).
- `ignoreImports` — descarta las declaraciones de import antes de tokenizar (default `true`).
- `rules` — `skip` por patrón. Gana la primera regla coincidente.
- `maxMemory` — presupuesto de memoria (default: `auto`).

### Presupuesto de memoria

Cada `check` corre bajo un presupuesto de memoria. El default, `auto`, es **el menor entre 1 GiB y el 25 % de la memoria disponible** (la del sistema o, dentro de un cgroup con límite, su margen). Se cambia con `--max-memory`, la variable `DUPELENS_MAX_MEMORY` o la clave `maxMemory`, en ese orden de precedencia, con un valor absoluto (`"512MiB"`, `"2GiB"`) o un porcentaje (`"40%"`). Si el escaneo excede el presupuesto, termina con **exit code 2** y sin reporte parcial, con o sin `--fail`.

Con tres o más copias del mismo bloque, cada copia se reporta contra la primera (`a-b`, `a-c`; `b-c` queda implícito).

### Salida

Soporta consola coloreada y JSON estructurado. Ver ejemplos arriba.

### Integraciones

Sirve con Husky, lefthook o GitHub Actions usando los mismos snippets de la sección en inglés.

### Por qué Rabin-Karp en vez de AST

- Cero dependencias: no hay que enviar parsers por lenguaje.
- Agnóstico: el mismo binario escanea Go, TypeScript, Python, Rust, Java, etc.
- Rápido: el hash rodante detecta matches en `O(n)` sobre el stream de tokens.

El trade-off está documentado en [ADR-012](https://github.com/artiko00/open-harness/blob/main/docs/adr-012-dupelens-rabin-karp-sobre-ast.md); el presupuesto de memoria, en [ADR-024](https://github.com/artiko00/open-harness/blob/main/docs/adr-024-dupelens-presupuesto-de-memoria.md).

### Limitaciones (v0.5.0)

- Detecta clones contiguos **exact** y **renamed**. Sentencias reordenadas, líneas insertadas o borradas y reescrituras equivalentes no se detectan — eso requiere análisis AST.
- El algoritmo es binario (hay match o no hay); no existe un flag de umbral de similitud.
- El override de `minTokens` por regla no funciona entre archivos porque la ventana debe ser uniforme. Usá `rules.skip` para excluir patrones por completo.

### Códigos de salida

| Código | Significado |
|---|---|
| `0` | Sin duplicados (o no se pasó `--fail`) |
| `1` | Hay duplicados con `--fail`, o error de uso o configuración |
| `2` | Se excedió el presupuesto de memoria: no se pudo completar el escaneo |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
