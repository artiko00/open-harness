# open-harness

A monorepo of lightweight code quality tools — each one a single binary, zero runtime dependencies, works in any language ecosystem.

## Tools

| Tool | Description | Status |
|---|---|---|
| [linelens](tools/linelens/) | File length linter — detects files exceeding a line limit | `v0.1.0` |
| [dupelens](tools/dupelens/) | Code duplication detector (Rabin-Karp, language-agnostic) | `v0.1.0` |
| [secretlens](tools/secretlens/) | Secret and credential detector (AWS keys, GitHub tokens, JWT, PEM, etc.) | `v0.1.0` |
| [testlens](tools/testlens/) | Test coverage detector — finds source files without tests (multi-language) | `v0.1.0` |
| bigo | Big O complexity analyzer | `planned` |

---

## linelens

Scans your project and reports files that exceed a configured line limit. Works with any language.

### Usage

```bash
linelens check               # scan current directory
linelens check --fail        # exit 1 if violations found (CI / git hooks)
linelens check --dir ./src   # scan a specific directory
linelens check --max 200     # override the line limit
linelens init                # generate a default linelens.json
```

### Configuration (`linelens.json`)

```json
{
  "default": { "maxLines": 100 },
  "rules": [
    { "pattern": "**/*_test.go",     "maxLines": 300 },
    { "pattern": "**/*.spec.*",      "maxLines": 300 },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": ["node_modules", "vendor", ".git", "dist"]
}
```

Pattern semantics follow `.gitignore` style.

### Husky integration

```bash
# .husky/pre-commit
npx linelens check --fail
```

---

## secretlens

Scans your codebase for hardcoded secrets and credentials. Detects AWS keys, GitHub tokens, PEM keys, JWTs, and generic secret assignments.

### Usage

```bash
secretlens check              # scan current directory
secretlens check --fail       # exit 1 if secrets found (git hooks / CI)
secretlens check --dir ./src  # scan a specific directory
secretlens check --no-color   # plain output for logs
secretlens init               # generate a default secretlens.json
```

### Built-in patterns

| Pattern | Severity |
|---|---|
| AWS Access Key ID (`AKIA…`) | critical |
| AWS Secret Access Key | critical |
| GitHub Personal Access Token (`ghp_…`) | critical |
| GitHub Fine-Grained Token (`github_pat_…`) | critical |
| PEM Private Key (`-----BEGIN … PRIVATE KEY`) | critical |
| JWT Token | high |
| Generic secret/password/api_key assignment | high |
| Generic token/bearer assignment | medium |

### Configuration (`secretlens.json`)

```json
{
  "patterns": [],
  "allowlist": ["example", "placeholder", "your_key_here", "changeme"],
  "exclude": ["node_modules", "vendor", ".git", "dist"]
}
```

`patterns: []` uses the 8 built-in patterns. Override to add custom rules.

The `allowlist` skips any line containing those strings (case-insensitive) — useful to suppress false positives in documentation or example files.

---

## testlens

Finds source files that don't have corresponding test files. Supports multiple languages out of the box.

### Usage

```bash
# Auto-detect language and scan
testlens check

# Scan with specific language
testlens check --lang go
testlens check --lang typescript

# Scan specific directory
testlens check --dir ./src

# Exit with code 1 if violations found (for CI)
testlens check --fail

# Generate a default config
testlens init
```

### Supported languages

| Language | Source extensions | Test patterns |
|---|---|---|
| Go | `.go` | `*_test.go` |
| TypeScript | `.ts`, `.tsx` | `*.test.ts`, `*.spec.ts`, `test_*.ts` |
| JavaScript | `.js`, `.jsx` | `*.test.js`, `*.spec.js`, `test_*.js` |
| Python | `.py` | `*_test.py`, `test_*.py` |
| Ruby | `.rb` | `*_spec.rb`, `*_test.rb` |
| Rust | `.rs` | `*_test.rs` |
| Java | `.java` | `*Test.java` |
| Kotlin | `.kt`, `.kts` | `*Test.kt` |
| C# | `.cs` | `*Tests.cs` |

### CI (GitHub Actions)

```yaml
- name: Run testlens
  run: ./tools/testlens/testlens check --lang go --fail
```

---

## dupelens

Detects duplicated code blocks using **Rabin-Karp** rolling-hash fingerprinting over tokenized source. Language-agnostic (works on Go, TS, Python, Rust, etc.) — strings and comments are stripped before tokenization to avoid false positives. See [ADR-012](docs/adr-012-dupelens-rabin-karp-sobre-ast.md).

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

- Detects only **literal** or near-literal duplication (token-by-token). Refactors with renamed variables are not flagged — that requires AST analysis ([ADR-012](docs/adr-012-dupelens-rabin-karp-sobre-ast.md) explains the trade-off).
- `--threshold` flag is not implemented; the algorithm is binary (match or not). See `[skip]` note in F-006.
- Per-rule `minTokens` override does not work cross-file because window sizes must be uniform. Skip via `rules` if you want per-pattern exclusion.

---

## Usage in Node/TS/JS projects

Install all tools with a single command (once `open-harness` is published to npm):

```bash
npm install --save-dev open-harness
```

Or install individual tools:

```bash
npm install --save-dev @open-harness/linelens @open-harness/dupelens \
  @open-harness/secretlens @open-harness/testlens
```

Add to your `package.json`:

```json
{
  "scripts": {
    "lint:lines":   "linelens check --fail",
    "lint:dupes":   "dupelens check --fail",
    "lint:secrets": "secretlens check --fail",
    "lint:tests":   "testlens check --fail",
    "lint":         "npm run lint:lines && npm run lint:dupes && npm run lint:secrets && npm run lint:tests"
  }
}
```

With [lint-staged](https://github.com/okonet/lint-staged) + Husky pre-commit:

```json
{
  "lint-staged": {
    "**/*": ["linelens check --fail", "secretlens check --fail"]
  }
}
```

GitHub Actions CI:

```yaml
- name: Install open-harness
  run: npm install -g open-harness

- run: linelens check --fail
- run: dupelens check --fail
- run: secretlens check --fail
- run: testlens check --fail --lang typescript --dir src/
```

---

## Combined CI / git hooks

Run all four quality tools as gates in a single workflow:

```yaml
# GitHub Actions (via npx, no install step)
- run: npx @open-harness/linelens check --fail
- run: npx @open-harness/dupelens check --fail
- run: npx @open-harness/secretlens check --fail
- run: npx @open-harness/testlens check --fail
```

Via lefthook (this repo uses this pattern — see `lefthook.yml`):

```yaml
pre-commit:
  commands:
    linelens:   { run: linelens check --fail --no-color }
    dupelens:   { run: dupelens check --fail --no-color }
    secretlens: { run: secretlens check --fail --no-color }
    testlens:   { run: testlens check --fail --lang typescript --dir src/ }
```

---

## Repository structure

```
open-harness/
├── tools/
│   ├── linelens/          ← v0.1.0 (file length linter, 100% coverage)
│   ├── dupelens/          ← v0.1.0 (duplicate detector, Rabin-Karp, 100% coverage)
│   ├── secretlens/        ← v0.1.0 (secret/credential detector, 100% coverage)
│   └── testlens/          ← v0.1.0 (test coverage detector, multi-language, 100% coverage)
├── npm/
│   ├── open-harness/      ← meta-package (instala los 4 tools)
│   └── @open-harness/
│       ├── open-harness/  ← meta-package (scoped)
│       ├── linelens/      ← npm wrapper (JS)
│       ├── linelens-{linux-x64,darwin-arm64,darwin-x64,win32-x64}/
│       ├── dupelens/      ← npm wrapper (JS)
│       ├── dupelens-{linux-x64,darwin-arm64,darwin-x64,win32-x64}/
│       ├── secretlens/    ← npm wrapper (JS)
│       ├── secretlens-{linux-x64,darwin-arm64,darwin-x64,win32-x64}/
│       ├── testlens/      ← npm wrapper (JS)
│       └── testlens-{linux-x64,darwin-arm64,darwin-x64,win32-x64}/
├── docs/                  ← Architecture Decision Records (ADR-001 … ADR-013)
├── scripts/
│   ├── build.sh           ← compile all tools
│   ├── build-npm.sh       ← cross-compile + npm packages (acepta <tool> como arg)
│   └── bench-vs-jscpd.sh  ← manual perf comparison dupelens vs jscpd
├── .agent/                ← agent harness (feature list, session log, init script)
├── AGENTS.md              ← agent instructions (TDD workflow, conventions)
├── go.work                ← Go workspace (4 tools)
├── lefthook.yml           ← git hooks (pre-commit: 3 tools, pre-push: tests x4)
├── linelens.json          ← linelens config for this repo
├── dupelens.json          ← dupelens config for this repo
└── secretlens.json        ← secretlens config for this repo
```

## Development

**Prerequisites:** Go 1.22+

```bash
git clone git@github.com:artiko00/open-harness.git
cd open-harness

# Run tests for all tools
go test ./tools/linelens && go test ./tools/dupelens && go test ./tools/testlens && go test ./tools/secretlens

# Build all tools
bash scripts/build.sh

# Lint this repo with its own tool
./tools/linelens/linelens check --fail
```

## Architecture decisions

All non-obvious decisions are documented as ADRs in [`docs/`](docs/):

| ADR | Decision |
|---|---|
| [ADR-001](docs/adr-001-go-sobre-node.md) | Go over Node.js |
| [ADR-002](docs/adr-002-cero-dependencias.md) | Zero external dependencies |
| [ADR-003](docs/adr-003-config-json.md) | JSON config format |
| [ADR-004](docs/adr-004-deteccion-binarios.md) | Binary file detection |
| [ADR-005](docs/adr-005-regla-100-lineas-aplicada-al-proyecto.md) | Project enforces its own rules |
| [ADR-006](docs/adr-006-semantica-glob-gitignore.md) | Glob pattern semantics |
| [ADR-007](docs/adr-007-lefthook-sobre-alternativas.md) | lefthook over Husky |
| [ADR-008](docs/adr-008-linelens-config-raiz.md) | Root-level configs (linelens, dupelens, secretlens) |
| [ADR-009](docs/adr-009-proyecto-protegido-por-su-propia-herramienta.md) | Repo protected by its own tools |
| [ADR-010](docs/adr-010-secretlens-diseno-detector.md) | secretlens design: regex, allowlist, severity |
| [ADR-011](docs/adr-011-cobertura-100-como-estandar.md) | 100% test coverage as project standard |
| [ADR-012](docs/adr-012-dupelens-rabin-karp-sobre-ast.md) | Rabin-Karp over AST for dupelens |
| [ADR-013](docs/adr-013-tdd-como-estandar.md) | TDD as project standard |

## For agents

If you are an AI agent (Claude Code, Codex, Cursor, etc.) working on this repo, read [`AGENTS.md`](AGENTS.md) first — it is the source of truth for workflow, TDD requirements, and conventions.

## License

MIT