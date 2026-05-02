# open-harness

A monorepo of lightweight code quality tools — each one a single binary, zero runtime dependencies, works in any language ecosystem.

## Tools

| Tool | Description | Status |
|---|---|---|
| [linelens](tools/linelens/) | File length linter — detects files exceeding a line limit | `v0.1.0` |
| [testlens](tools/testlens/) | Test coverage detector — finds source files without tests | `v0.1.0` |
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
| TypeScript | `.ts`, `.tsx` | `*.test.ts`, `*.spec.ts` |
| JavaScript | `.js`, `.jsx` | `*.test.js`, `*.spec.js` |
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
│   ├── linelens/      ← File length linter
│   ├── testlens/      ← Test coverage detector
│   └── secretlens/    ← Secret and credential detector
├── npm/               ← npm package source
├── docs/adr/          ← Architecture Decision Records (ADR-001 …)
├── scripts/
│   ├── build.sh       ← compile all tools
│   └── build-npm.sh   ← cross-compile + update npm packages
├── .agent/            ← agent harness (feature list)
├── go.work             ← Go workspace
├── lefthook.yml        ← git hooks
└── linelens.json       ← lint config for this repo
```

## Development

**Prerequisites:** Go 1.22+

```bash
# Clone
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

All non-obvious decisions are documented as ADRs in [`docs/`](docs/):

- [ADR-001](docs/adr/adr-001-testlens.md) — testlens: multi-language test coverage detection
- [ADR-002](docs/adr-002-cero-dependencias.md) — Zero external dependencies
- [ADR-003](docs/adr-adr-003-config-json.md) — JSON config format

## License

MIT
