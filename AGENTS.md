# AGENTS.md

open-harness is a Go monorepo for developer tools. Currently ships a single CLI tool (linelens) with architecture ready for more.

## Commands

- Install: `go install ./tools/linelens` (from repo root)
- Build all tools: `./scripts/build.sh`
- Build single tool: `go build -o tools/linelens/linelens ./tools/linelens`
- Test: `go test ./tools/...`
- Lint: `linelens check --fail`
- Lint fix: manually review files exceeding line limits

## Tech stack

- Runtime: Go 1.22
- Language: Go (no framework)
- Package manager: Go modules + Go workspace (go.work)
- Testing: Go stdlib (`testing` package)
- Linting: linelens (custom, builtin)

## Project structure

tools/
├── linelens/      # CLI file length linter (only tool for now)
│   ├── main.go    # Entry point, flag parsing, commands
│   ├── scanner.go # File discovery and line counting
│   ├── matcher.go # Pattern matching against config rules
│   ├── reporter.go# Output formatting (console, JSON)
│   ├── config.go  # Config file loading (linelens.json)
│   ├── help.go    # Help text
│   ├── binary.go  # Binary path resolution
│   └── *_test.go  # Unit tests
scripts/
├── build.sh       # Build all tools to their directories
└── build-npm.sh   # Build for npm distribution
npm/               # npm package source (future publish)
.agent/            # Agent config and feature backlog
docs/              # Documentation

## Code style

- Exports: PascalCase for types/functions, camelCase for variables
- Files: max 100 lines (default), 300 lines for tests
- Error handling: wrap with `fmt.Errorf("context: %w", err)`
- No external dependencies in tools (pure stdlib)

## Testing

- Framework: Go testing (`go test`)
- Location: co-located with source (`*_test.go`)
- Naming: `func TestX(t *testing.T)` pattern
- Run related: `go test ./tools/linelens/... -v -run TestMatcher`

## Git workflow

- Branch: `feat/`, `fix/`, `refactor/`
- Commit: `feat: description` / `fix: description`
- No hooks currently configured

## Boundaries

### Always do
- Run `linelens check` before committing new code
- Run `go test ./tools/...` after modifying source
- Build with `./scripts/build.sh` to verify all tools compile

### Ask first
- Before adding external dependencies to any tool
- Before modifying the go.work workspace structure
- Before adding new tools (check feature-list.json for roadmap)

### Never do
- Never commit compiled binaries (they're excluded in .gitignore)
- Never skip linelens failures in CI
- Never add dependencies to tools/ without reviewing impact on stdlib goal

## Features (backlog)

See `.agent/feature-list.json` for planned additions. Currently:
- F-001: `bigo` - Big O complexity estimator (not implemented)
- F-002: GitHub Actions CI
- F-003: npm publish workflow
- F-004: Release automation
- F-005: JSON report output for linelens