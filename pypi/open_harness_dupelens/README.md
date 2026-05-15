# open-harness-dupelens

Code duplication detector. Uses **Rabin-Karp** rolling-hash fingerprinting over tokenized source — strings and comments are stripped before hashing to reduce false positives. Language-agnostic (Go, TS, JS, Python, Rust, Java, etc.). Single native binary, zero runtime dependencies.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

> **Same tool, other ecosystems**: also available on **npm** ([`@open_harness/dupelens`](https://www.npmjs.com/package/@open_harness/dupelens)) and on **Packagist** (`open-harness/dupelens`). Identical binary, identical config; pick the registry that matches your stack.

## Install

```bash
pip install open-harness-dupelens
```

pip picks the right native wheel for your platform automatically (Linux x86_64, macOS arm64, macOS x86_64, Windows x86_64). Each wheel embeds the Go binary — no runtime deps.

## Usage

```bash
dupelens check                  # scan current directory with defaults
dupelens check --fail           # exit 1 if duplicates found (CI / git hooks)
dupelens check --min-tokens 30  # override the rolling window size
dupelens check --format=json    # JSON output for tooling integrations
dupelens check --dir ./src      # scan a specific directory
dupelens check --verbose        # print timings to stderr
dupelens check --no-color       # plain console output
dupelens init                   # generate a default dupelens.json
dupelens version                # print version
```

## Configuration

Place a `dupelens.json` at the repo root:

```json
{
  "default": {
    "minTokens": 50,
    "minLines": 5
  },
  "rules": [
    { "pattern": "**/*_test.go",     "skip": true },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": ["node_modules", "vendor", ".git", "dist", "build"]
}
```

- `minTokens` — window size of the rolling hash. Higher values catch only larger duplications.
- `minLines` — filters short matches (e.g. back-to-back identical imports).
- `rules` — per-pattern `skip`. The first matching entry wins.

### Alternative: configure inside `pyproject.toml` or the dedicated `dupelens.json`

If you prefer not to keep a separate `dupelens.json`, add a `dupelens` key in your `package.json` with the same shape:

```json
{
  "name": "my-project",
  "dupelens": {
    "default": { "minTokens": 50, "minLines": 5 },
    "rules": [{ "pattern": "**/*_test.go", "skip": true }],
    "exclude": ["node_modules", "dist"]
  }
}
```

Precedence: `--config <path>` > `dupelens.json` > `package.json` key > built-in defaults. CLI flags (`--min-tokens`, `--format`, etc.) always win.

## Output (console)

```
DUPLICATES (2 match(es) found in 87 files):

  src/auth.go:42-58  <->  src/users.go:12-28  (35 tokens)
  | func validate(input string) error {
  | ...
  src/db.go:1-10  <->  src/cache.go:1-10  (15 tokens)

SUMMARY: 2 match(es) across 87 files
Top duplicated files:
  - src/auth.go  (1 match(es))
```

## Output (JSON)

```json
{
  "scannedFiles": 87,
  "matchCount": 2,
  "matches": [
    {
      "fileA": "src/auth.go", "startLineA": 42, "endLineA": 58,
      "fileB": "src/users.go", "startLineB": 12, "endLineB": 28,
      "tokens": 35
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
dupelens check --fail
```

```yaml
# GitHub Actions
- name: Run dupelens
  run: npx @open_harness/dupelens check --fail
```

## Why Rabin-Karp over AST?

- Zero dependencies: no language-specific parsers to ship per language.
- Language-agnostic: the same binary scans Go, TypeScript, Python, Rust, Java, etc.
- Fast: rolling hash detects matches in `O(n)` over the token stream.

The trade-off is documented in [ADR-012](https://github.com/artiko00/open-harness/blob/main/docs/adr-012-dupelens-rabin-karp-sobre-ast.md).

## Limitations (v0.2.0)

- Detects only **literal** or near-literal duplication (token-by-token). Refactors with renamed variables are not flagged — that requires AST analysis.
- The algorithm is binary (match or no match); there is no similarity threshold flag.
- Per-rule `minTokens` override does not work cross-file because window sizes must be uniform. Use `rules.skip` to exclude patterns entirely.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | No duplicates (or `--fail` not passed) |
| `1` | Duplicates found and `--fail` was passed, or config error |

---

## Español

Detector de duplicación de código. Usa fingerprinting **Rabin-Karp** (hash rodante) sobre el código tokenizado — los strings y comentarios se eliminan antes del hashing para reducir falsos positivos. Agnóstico al lenguaje (Go, TS, JS, Python, Rust, Java, etc.). Un solo binario nativo, cero dependencias.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

### Instalación

```bash
pip install open-harness-dupelens
```

pip descarga automáticamente la wheel nativa correcta para tu plataforma.

### Uso

```bash
dupelens check                  # escanea con defaults
dupelens check --fail           # exit 1 si hay duplicados (CI / git hooks)
dupelens check --min-tokens 30  # cambia el tamaño de ventana del hash rodante
dupelens check --format=json    # salida JSON para integraciones
dupelens check --dir ./src      # escanea un directorio específico
dupelens check --verbose        # imprime timings en stderr
dupelens check --no-color       # consola sin colores
dupelens init                   # genera un dupelens.json por defecto
dupelens version                # imprime la versión
```

### Configuración

Colocá un `dupelens.json` en la raíz del repo (ver ejemplo arriba).

- `minTokens` — tamaño de la ventana del hash rodante. Valores más altos detectan solo duplicaciones más grandes.
- `minLines` — filtra matches cortos (ej. imports idénticos consecutivos).
- `rules` — `skip` por patrón. Gana la primera regla coincidente.

#### Alternativa: configurar dentro de `pyproject.toml` o `dupelens.json`

Si preferís no tener un `dupelens.json` separado, agregá una key `dupelens` en tu `package.json` con la misma forma del archivo dedicado. Precedencia: `--config <path>` > `dupelens.json` > key en `package.json` > defaults. Los flags CLI (`--min-tokens`, `--format`, etc.) siempre ganan.

### Salida

Soporta consola coloreada y JSON estructurado. Ver ejemplos arriba.

### Integraciones

Sirve con Husky, lefthook o GitHub Actions usando los mismos snippets de la sección en inglés.

### Por qué Rabin-Karp en vez de AST

- Cero dependencias: no hay que enviar parsers por lenguaje.
- Agnóstico: el mismo binario escanea Go, TypeScript, Python, Rust, Java, etc.
- Rápido: el hash rodante detecta matches en `O(n)` sobre el stream de tokens.

El trade-off está documentado en [ADR-012](https://github.com/artiko00/open-harness/blob/main/docs/adr-012-dupelens-rabin-karp-sobre-ast.md).

### Limitaciones (v0.2.0)

- Solo detecta duplicación **literal** o cuasi-literal (token a token). Refactors con variables renombradas no se detectan — eso requiere análisis AST.
- El algoritmo es binario (hay match o no hay); no existe un flag de umbral de similitud.
- El override de `minTokens` por regla no funciona entre archivos porque la ventana debe ser uniforme. Usá `rules.skip` para excluir patrones por completo.

### Códigos de salida

| Código | Significado |
|---|---|
| `0` | Sin duplicados (o no se pasó `--fail`) |
| `1` | Hay duplicados con `--fail`, o error de configuración |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
