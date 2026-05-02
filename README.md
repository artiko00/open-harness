# open-harness

A monorepo of lightweight code quality tools — each one a single binary, zero runtime dependencies, works in any language ecosystem.

## Tools

| Tool | Description | Status |
|---|---|---|
| [linelens](tools/linelens/) | File length linter — detects files exceeding a line limit | `v0.1.0` |
| [testlens](tools/testlens/) | Test coverage detector — finds source files without tests | `v0.1.0` |
| [secretlens](tools/secretlens/) | Secret detector — finds hardcoded credentials | `v0.1.0` |
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

## Repository structure

```
open-harness/
├── tools/
│   ├── linelens/      ← File length linter (100% coverage)
│   ├── testlens/      ← Test coverage detector
│   └── secretlens/    ← Secret detector (100% coverage)
├── docs/adr/          ← Architecture Decision Records (ADR-001 …)
├── scripts/
│   ├── build.sh       ← compile all tools
│   └── build-npm.sh   ← cross-compile + update npm packages
├── .agent/            ← agent harness (feature list)
├── go.work             ← Go workspace
├── lefthook.yml        ← hooks: pre-commit linelens, pre-push go test
└── linelens.json       ← lint config for this repo
```

## Development

**Prerequisites:** Go 1.22+

```bash
git clone git@github.com:artiko00/open-harness.git
cd open-harness

# Run tests for all tools
go test ./tools/linelens && go test ./tools/testlens && go test ./tools/secretlens

# Build all tools
bash scripts/build.sh

# Lint this repo with its own tool
./tools/linelens/linelens check --fail
```

## Architecture decisions

| ADR | Decision |
|---|---|
| [ADR-001](docs/adr/adr-001-testlens.md) | testlens: multi-language test coverage detection |
| [ADR-002](docs/adr/adr-002-cero-dependencias.md) | Zero external dependencies |
| [ADR-003](docs/adr/adr-003-config-json.md) | JSON config format |
| [ADR-004](docs/adr/adr-004-deteccion-binarios.md) | Binary file detection |
| [ADR-005](docs/adr/adr-005-regla-100-lineas-aplicada-al-proyecto.md) | Project enforces its own rules |
| [ADR-006](docs/adr/adr-006-semantica-glob-gitignore.md) | Glob pattern semantics |
| [ADR-007](docs/adr/adr-007-lefthook-sobre-alternativas.md) | lefthook over Husky |
| [ADR-008](docs/adr/adr-008-linelens-config-raiz.md) | Root-level linelens.json in monorepo |
| [ADR-009](docs/adr/adr-009-proyecto-protegido-por-su-propia-herramienta.md) | Repo protected by its own tool |
| [ADR-010](docs/adr-010-secretlens-diseno-detector.md) | secretlens design: regex, allowlist, severity |
| [ADR-011](docs/adr-011-cobertura-100-como-estandar.md) | 100% test coverage as project standard |

## License

MIT