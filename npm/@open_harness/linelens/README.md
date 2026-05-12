# @open_harness/linelens

File length linter for any language. Scans your project and reports files that exceed a configured line limit. Single native binary, zero runtime dependencies, works with any language ecosystem.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo.

## Install

```bash
npm install --save-dev @open_harness/linelens
```

The right native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) is downloaded automatically via `optionalDependencies`.

## Usage

```bash
npx linelens check               # scan current directory
npx linelens check --fail        # exit 1 on violations (CI / git hooks)
npx linelens check --dir ./src   # scan a specific directory
npx linelens check --max 200     # override the line limit
npx linelens check --no-color    # plain output for logs
npx linelens init                # generate a default linelens.json
npx linelens version             # print version
```

## Configuration

Place a `linelens.json` at the repo root:

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

Pattern semantics follow `.gitignore` style. The first matching `rules` entry wins; if no rule matches, `default.maxLines` applies.

## Husky integration

```bash
# .husky/pre-commit
npx linelens check --fail
```

## lefthook integration

```yaml
# lefthook.yml
pre-commit:
  commands:
    linelens:
      run: npx linelens check --fail --no-color
```

## GitHub Actions

```yaml
- name: Run linelens
  run: npx @open_harness/linelens check --fail
```

## Why a line limit?

Large files concentrate too many responsibilities and become hard to read, test, and refactor. Enforcing a soft cap (e.g. 100 lines per file with exceptions for tests) keeps modules focused and pushes responsibility split decisions early in the development cycle, when they are cheap.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | No violations (or `--fail` not passed) |
| `1` | Violations found and `--fail` was passed, or config error |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
