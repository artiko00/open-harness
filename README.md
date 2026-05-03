# open-harness

A monorepo of lightweight code quality tools — each one a single binary, zero runtime dependencies, works in any language ecosystem.

## Tools

| Tool | Description | Status |
|---|---|---|
| [linelens](tools/linelens/) | File length linter — detects files exceeding a line limit | `v0.1.0` |
| [dupelens](tools/dupelens/) | Code duplication detector (Rabin-Karp, language-agnostic) | `v0.1.0` |
| [secretlens](tools/secretlens/) | Secret and credential detector (AWS keys, GitHub tokens, JWT, PEM, etc.) | `v0.1.0` |
| bigo | Big O complexity analyzer | `planned` |

---

## linelens

Scans your project and reports files that exceed a configured line limit. Works with any language — Go, TypeScript, Python, Rust, or anything else.

### Install via npm

```bash
npm install --save-dev @open-harness/linelens
```

### Usage

```bash
# Scan current directory
linelens check

# Fail with exit code 1 if violations found (for CI / git hooks)
linelens check --fail

# Scan a specific directory
linelens check --dir ./src

# Override the line limit
linelens check --max 200

# Generate a default config file
linelens init
```

### Configuration (`linelens.json`)

```json
{
  "default": {
    "maxLines": 100
  },
  "rules": [
    { "pattern": "**/*_test.go",     "maxLines": 300 },
    { "pattern": "**/*.spec.*",      "maxLines": 300 },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": [
    "node_modules", "vendor", ".git", "dist", "build"
  ]
}
```

Pattern semantics follow `.gitignore` style: a pattern without `/` matches the filename in any directory.

### Husky integration

```bash
# .husky/pre-commit
npx linelens check --fail
```

### CI (GitHub Actions)

```yaml
- uses: actions/setup-node@v4
  with:
    node-version: 20
- run: npm ci
- run: npx linelens check --fail
```

---

## dupelens

Detects duplicated code blocks using **Rabin-Karp** rolling-hash fingerprinting over tokenized source. Language-agnostic (works on Go, TS, Python, Rust, etc.) — strings and comments are stripped before tokenization to avoid false positives. See [ADR-010](docs/adr-010-dupelens-rabin-karp-sobre-ast.md).

### Install via npm

```bash
npm install --save-dev @open-harness/dupelens
```

### Usage

```bash
# Scan current directory with defaults
dupelens check

# Fail with exit code 1 if duplicates found (CI / git hooks)
dupelens check --fail

# Override the token threshold
dupelens check --min-tokens 30

# JSON output for tooling integrations
dupelens check --format=json > report.json

# Verbose timings to stderr
dupelens check --verbose

# Generate default config
dupelens init
```

### Configuration (`dupelens.json`)

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

`minTokens` is the window size of the rolling hash. Higher values catch only larger duplications. `minLines` filters short matches (e.g., back-to-back identical imports).

### Output (console)

```
DUPLICATES (2 match(es) found in 87 files):

  src/auth.go:42-58  ↔  src/users.go:12-28  (35 tokens)
  | func validate(input string) error {
  | ...
  src/db.go:1-10  ↔  src/cache.go:1-10  (15 tokens)

SUMMARY: 2 match(es) across 87 files
Top duplicated files:
  - src/auth.go  (1 match(es))
```

### Output (JSON)

```json
{
  "scannedFiles": 87,
  "matchCount": 2,
  "matches": [
    { "fileA": "src/auth.go", "startLineA": 42, "endLineA": 58,
      "fileB": "src/users.go", "startLineB": 12, "endLineB": 28,
      "tokens": 35 }
  ],
  "summary": {
    "topDuplicatedFiles": [{ "file": "src/auth.go", "count": 1 }]
  }
}
```

### Limitations (v0.1.0)

- Detects only **literal** or near-literal duplication (token-by-token). Refactors with renamed variables are not flagged — that requires AST analysis ([ADR-010](docs/adr-010-dupelens-rabin-karp-sobre-ast.md) explains the trade-off).
- `--threshold` flag is not implemented; the algorithm is binary (match or not). See `[skip]` note in F-006.
- Per-rule `minTokens` override does not work cross-file because window sizes must be uniform. Skip via `rules` if you want per-pattern exclusion.

---

## Combined CI / git hooks

Run all three quality tools as gates in a single workflow:

```yaml
# GitHub Actions
- run: npx @open-harness/linelens check --fail
- run: npx @open-harness/dupelens check --fail
- run: npx @open-harness/secretlens check --fail
```

Or via lefthook (this repo uses this pattern — see `lefthook.yml`):

```yaml
pre-commit:
  commands:
    linelens:   { run: tools/linelens/linelens   check --fail --no-color }
    dupelens:   { run: tools/dupelens/dupelens   check --fail --no-color }
    secretlens: { run: tools/secretlens/secretlens check --fail --no-color }
```

---

## Repository structure

```
open-harness/
├── tools/
│   ├── linelens/          ← v0.1.0 (file length linter)
│   ├── dupelens/          ← v0.1.0 (duplicate detector)
│   └── secretlens/        ← v0.1.0 (secret/credential detector)
├── npm/
│   └── @open-harness/
│       ├── linelens/      ← npm wrapper (JS)
│       ├── linelens-{linux-x64,darwin-arm64,darwin-x64,win32-x64}/
│       ├── dupelens/      ← npm wrapper (JS)
│       └── dupelens-{linux-x64,darwin-arm64,darwin-x64,win32-x64}/
├── docs/                  ← Architecture Decision Records (ADR-001 … ADR-011)
├── scripts/
│   ├── build.sh           ← compile all tools
│   ├── build-npm.sh       ← cross-compile + npm packages (acepta <tool> como arg)
│   └── bench-vs-jscpd.sh  ← manual perf comparison dupelens vs jscpd
├── .agent/                ← agent harness (feature list, session log, init script)
├── AGENTS.md              ← agent instructions (TDD workflow, conventions)
├── go.work                ← Go workspace (3 tools)
├── lefthook.yml           ← git hooks (pre-commit: linelens + dupelens, pre-push: tests x3)
├── linelens.json          ← linelens config for this repo
└── dupelens.json          ← dupelens config for this repo
```

## Development

**Prerequisites:** Go 1.22+, [lefthook](https://github.com/evilmartians/lefthook)

```bash
# Clone
git clone git@github.com:artiko00/open-harness.git
cd open-harness

# Activate git hooks
brew install lefthook
lefthook install

# Run tests
cd tools/linelens && go test ./...

# Build all tools
bash scripts/build.sh

# Build npm packages for linelens
bash scripts/build-npm.sh linelens
```

## Architecture decisions

All non-obvious decisions are documented as ADRs in [`docs/`](docs/):

- [ADR-001](docs/adr-001-go-sobre-node.md) — Go over Node.js
- [ADR-002](docs/adr-002-cero-dependencias.md) — Zero external dependencies
- [ADR-003](docs/adr-003-config-json.md) — JSON config format
- [ADR-004](docs/adr-004-deteccion-binarios.md) — Binary file detection
- [ADR-005](docs/adr-005-regla-100-lineas-aplicada-al-proyecto.md) — The project enforces its own rules
- [ADR-006](docs/adr-006-semantica-glob-gitignore.md) — Glob pattern semantics
- [ADR-007](docs/adr-007-lefthook-sobre-alternativas.md) — lefthook over Husky / pre-commit
- [ADR-008](docs/adr-008-linelens-config-raiz.md) — Root-level linelens.json in monorepo
- [ADR-009](docs/adr-009-proyecto-protegido-por-su-propia-herramienta.md) — Repo protected by its own tool
- [ADR-010](docs/adr-010-dupelens-rabin-karp-sobre-ast.md) — Rabin-Karp over AST for dupelens
- [ADR-011](docs/adr-011-tdd-como-estandar.md) — TDD as project standard

## For agents

If you are an AI agent (Claude Code, Codex, Cursor, etc.) working on this repo, read [`AGENTS.md`](AGENTS.md) first — it is the source of truth for workflow, TDD requirements, and conventions.

## License

MIT
