# AGENTS.md

open-harness is a Go monorepo for developer tools. Each tool is a single binary with zero runtime dependencies. Currently ships three CLI tools:

- **linelens** — file length linter
- **testlens** — multi-language test coverage detector  
- **secretlens** — secret and credential detector

## Commands

- Build single tool: `go build -o tools/<tool>/<tool> ./tools/<tool>`
- Build all tools: `./scripts/build.sh`
- Test all tools: `go test ./tools/linelens && go test ./tools/testlens && go test ./tools/secretlens`
- Lint: `go run ./tools/linelens/. check --fail --no-color`

## Tech stack

- Runtime: Go 1.22
- Language: Go (no framework)
- Package manager: Go modules + Go workspace (go.work)
- Testing: Go stdlib (`testing` package)
- 100% statement coverage required on every tool (ADR-011)

## Project structure

```
tools/
├── linelens/      # File length linter
│   ├── main.go    # Entry point, commands: check, init, version
│   ├── scanner.go # File discovery and line counting
│   ├── matcher.go # Pattern matching against config rules
│   ├── reporter.go# Output formatting
│   ├── config.go  # Config file loading (linelens.json)
│   ├── help.go    # Help text
│   ├── binary.go  # Binary file detection
│   └── *_test.go  # Unit tests (100% coverage)
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
└── secretlens/    # Secret and credential detector
    ├── main.go    # Entry (osExit injection + run() dispatcher)
    ├── scanner.go # File discovery and secret scanning
    ├── engine.go  # Pattern compilation, scanFile, allowlist
    ├── matcher.go # Glob pattern matching
    ├── reporter.go# Console output with severity colors
    ├── patterns.go# 8 built-in regex patterns
    ├── config.go  # secretlens.json loading
    ├── binary.go  # Binary file detection
    ├── help.go    # Usage text
    └── *_test.go  # Unit tests (100% coverage)
scripts/
├── build.sh       # Build all tools to their directories
└── build-npm.sh   # Build for npm distribution
docs/adr/          # Architecture Decision Records
.agent/            # Agent config and feature backlog
```

## Code style

- Exports: PascalCase for types/functions, camelCase for variables
- Files: max 100 lines (default), 300 lines for tests
- Error handling: wrap with `fmt.Errorf("context: %w", err)`
- No external dependencies in tools (pure stdlib)
- 100% statement coverage required on every tool

### CLI pattern for testable main.go

```go
var osExit = os.Exit // inject for testing

func main() {
    osExit(run(os.Args[1:]))
}

func run(args []string) int {
    // return exit code
}
```

- Use `flag.ContinueOnError` (not `ExitOnError`) in subcommands so flag errors are testable
- Never call `os.Exit` directly in business logic — use `osExit` variable

## Testing

- Framework: Go testing (`go test`)
- Location: co-located with source (`*_test.go`)
- Naming: `func TestX(t *testing.T)` pattern
- Run related: `go test ./tools/<tool>/... -v -run TestName`
- Coverage: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`

## Git workflow

- Branch: `feat/`, `fix/`, `refactor/`, `epic/` for multi-step features
- Commit: `feat:`, `fix:`, `test:`, `docs:`, `chore:` prefix
- GPG signing: all commits signed with artiko00 GPG key

## Boundaries

### Always do
- Run `linelens check` before committing new code
- Maintain 100% statement coverage — verify with `go test -coverprofile`
- Use `run(args []string) int` pattern in any new tool's `main.go`

### Ask first
- Before adding external dependencies to any tool
- Before modifying the `go.work` workspace structure
- Before adding new tools (check `.agent/feature-list.json` for roadmap)

### Never do
- Never commit compiled binaries (excluded in `.gitignore`)
- Never use `flag.ExitOnError` in subcommands — use `flag.ContinueOnError`
- Never call `os.Exit` directly in business logic — use `osExit` variable
- Never add dependencies without reviewing the stdlib goal

## Features (backlog)

See `.agent/feature-list.json` for details:

| ID | Feature | Status |
|---|---|---|
| F-001 | `bigo` — Big O complexity analyzer | planned |
| F-002 | GitHub Actions CI (matrix: linux/mac/win) | planned |
| F-003 | npm publish `@open-harness/linelens` | planned |
| F-004 | GitHub Actions release workflow | planned |
| F-005 | `linelens report --format=json` | planned |
| F-006 | `testlens` — test coverage detector | done |