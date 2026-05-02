# AGENTS.md

open-harness is a Go monorepo for developer tools. Each tool is a single binary with zero runtime dependencies. Currently ships two CLI tools: **linelens** (file length linter) and **secretlens** (secret detector).

## Commands

- Build single tool: `go build -o tools/linelens/linelens ./tools/linelens`
- Build all tools: `./scripts/build.sh`
- Test one tool: `cd tools/linelens && go test ./...`
- Test all tools: `cd tools/linelens && go test ./... && cd ../secretlens && go test ./...`
- Coverage report: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- Lint: `go run ./tools/linelens/. check --fail --no-color`

## Tech stack

- Runtime: Go 1.22
- Language: Go (no framework, no external dependencies)
- Package manager: Go modules + Go workspace (`go.work`)
- Testing: Go stdlib (`testing` package) — 100% statement coverage required
- Linting: linelens (builtin), lefthook (git hooks)

## Project structure

```
tools/
├── linelens/      # File length linter
│   ├── main.go    # Entry (osExit injection + run() dispatcher)
│   ├── scanner.go # File discovery and line counting
│   ├── matcher.go # Glob pattern matching (gitignore-style)
│   ├── reporter.go# Console output formatting
│   ├── config.go  # linelens.json loading
│   ├── binary.go  # Binary file detection
│   ├── help.go    # Usage text
│   └── *_test.go  # Unit tests (100% coverage)
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
```

## Code style

- Max file length: 100 lines (enforced by linelens on pre-commit)
- Test files: up to 300 lines (rule in `linelens.json`)
- Exports: PascalCase types/functions, camelCase variables
- No external dependencies in any tool (pure stdlib)
- `main()` delegates to `run(args []string) int` — never put logic in `main()`
- Error handling: wrap with `fmt.Errorf("context: %w", err)`

## Testing standards

- 100% statement coverage required on every tool (ADR-011)
- Use `var osExit = os.Exit` + `func run(args []string) int` pattern for testable CLIs
- Use `flag.ContinueOnError` (not `ExitOnError`) in subcommands so flag errors are testable
- Error branches requiring OS-level failure: use `chmod 0000` in tests with `defer` restore

## Git hooks (lefthook)

- `pre-commit`: `linelens check --fail` — blocks commit if any file exceeds line limit
- `pre-push` (parallel):
  - `cd tools/linelens && go test ./...`
  - `cd tools/secretlens && go test ./...`

## Git workflow

- Branch: `feat/`, `fix/`, `refactor/`, `epic/`
- Commit: `feat:`, `fix:`, `test:`, `docs:`, `chore:` prefix
- Push only after both test suites pass (lefthook enforces this)

## Boundaries

### Always do
- Run `linelens check` before committing
- Maintain 100% statement coverage — verify with `go test -coverprofile`
- Use `run(args []string) int` pattern in any new tool's `main.go`

### Ask first
- Before adding external dependencies to any tool
- Before modifying the `go.work` workspace structure
- Before adding new tools (see `.agent/feature-list.json` for roadmap)

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
