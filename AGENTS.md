# AGENTS.md

open-harness is a Go monorepo for developer tools. Currently ships three CLI tools: linelens, testlens, and secretlens.

## Commands

- Install all tools: `go install ./tools/...` (from repo root)
- Build all tools: `./scripts/build.sh`
- Build single tool: `go build -o tools/<tool>/<tool> ./tools/<tool>`
- Test: `go test ./tools/linelens && go test ./tools/testlens && go test ./tools/secretlens`
- Lint: `linelens check --fail` (from repo root)

## Tech stack

- Runtime: Go 1.22
- Language: Go (no framework)
- Package manager: Go modules + Go workspace (go.work)
- Testing: Go stdlib (`testing` package)
- Linting: linelens (custom, builtin)

## Project structure

tools/
├── linelens/      # File length linter
│   ├── main.go    # Entry point, commands: check, init, version
│   ├── scanner.go # File discovery and line counting
│   ├── matcher.go # Pattern matching against config rules
│   ├── reporter.go# Output formatting
│   ├── config.go  # Config file loading (linelens.json)
│   ├── help.go    # Help text
│   ├── binary.go  # Binary path resolution
│   └── *_test.go  # Unit tests
├── testlens/      # Test coverage detector (multi-language)
│   ├── main.go    # Entry point, commands: check, init, version
│   ├── check.go   # CLI flag parsing, config struct
│   ├── coverage.go# Core scanning logic, filepath.Walk
│   ├── detect.go  # Language auto-detection
│   ├── language.go# Language mapping definitions
│   ├── matcher.go # Test candidate generation
│   ├── scanner.go # File discovery utilities
│   ├── reporter.go# Output formatting
│   ├── file.go    # fileExists helper
│   └── scanner_test.go # Unit tests
├── secretlens/    # Secret and credential detector
scripts/
├── build.sh       # Build all tools to their directories
└── build-npm.sh   # Build for npm distribution
npm/               # npm package source (future publish)
docs/adr/          # Architecture Decision Records
.agent/            # Agent config and feature backlog

## Code style

- Exports: PascalCase for types/functions, camelCase for variables
- Files: max 100 lines (default), 300 lines for tests
- Error handling: wrap with `fmt.Errorf("context: %w", err)`
- No external dependencies in tools (pure stdlib)

## Testing

- Framework: Go testing (`go test`)
- Location: co-located with source (`*_test.go`)
- Naming: `func TestX(t *testing.T)` pattern
- Run related: `go test ./tools/<tool>/... -v -run TestName`

## Git workflow

- Branch: `feat/`, `fix/`, `refactor/`, `epic/` for multi-step features
- Commit: `feat: description` / `fix: description`
- GPG signing: all commits signed with artiko00 GPG key
- No hooks currently configured

## Boundaries

### Always do
- Run `linelens check --fail` before committing new code
- Run `go test ./tools/<tool>` after modifying source
- Build with `./scripts/build.sh` to verify all tools compile

### Ask first
- Before adding external dependencies to any tool
- Before modifying the go.work workspace structure
- Before adding new tools (check feature-list.json for roadmap)

### Never do
- Never commit compiled binaries (excluded in .gitignore)
- Never skip linelens failures in CI
- Never add dependencies to tools without reviewing impact on stdlib goal

## Tools

### linelens
File length linter for any language. Detects files exceeding configured line limits.

### testlens
Test coverage detector. Finds source files without corresponding test files. Supports 9 languages.

### secretlens
Secret and credential detector. Scans codebases for hardcoded secrets, API keys, passwords.

## Features (backlog)

See `.agent/feature-list.json` for planned additions:
- F-001: `bigo` - Big O complexity estimator (planned)
- F-002: GitHub Actions CI (planned)
- F-003: npm publish workflow (planned)
- F-004: Release automation (planned)
- F-005: linelens JSON report output (planned)