# @open_harness/dupelens

Code duplication detector for any language. Uses **Rabin-Karp** rolling-hash fingerprinting over tokenized source — strings and comments are stripped before hashing to reduce false positives. Single native binary, zero runtime dependencies, works on any codebase (Go, TS, JS, Python, Rust, Java, etc.).

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo.

## Install

```bash
npm install --save-dev @open_harness/dupelens
```

The right native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) is downloaded automatically via `optionalDependencies`.

## Usage

```bash
npx dupelens check                  # scan current directory with defaults
npx dupelens check --fail           # exit 1 if duplicates found (CI / git hooks)
npx dupelens check --min-tokens 30  # override window size
npx dupelens check --format=json    # JSON output for tooling integrations
npx dupelens check --dir ./src      # scan a specific directory
npx dupelens check --verbose        # print timings to stderr
npx dupelens check --no-color       # plain console output
npx dupelens init                   # generate a default dupelens.json
npx dupelens version                # print version
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

## Husky / lefthook / CI integration

```bash
# .husky/pre-commit
npx dupelens check --fail
```

```yaml
# .github/workflows/quality.yml
- name: Run dupelens
  run: npx @open_harness/dupelens check --fail
```

## Why Rabin-Karp over AST?

- Zero dependencies: no language-specific parsers to ship per language.
- Language-agnostic: the same binary scans Go, TypeScript, Python, Rust, Java, etc.
- Fast: rolling hash detects matches in `O(n)` over the token stream.

The trade-off is documented in [ADR-012](https://github.com/artiko00/open-harness/blob/main/docs/adr-012-dupelens-rabin-karp-sobre-ast.md).

## Limitations (v0.1.0)

- Detects only **literal** or near-literal duplication (token-by-token). Refactors with renamed variables are not flagged — that requires AST analysis.
- The algorithm is binary (match or no match); there is no similarity threshold flag.
- Per-rule `minTokens` override does not work cross-file because window sizes must be uniform. Use `rules.skip` to exclude patterns entirely.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | No duplicates (or `--fail` not passed) |
| `1` | Duplicates found and `--fail` was passed, or config error |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
