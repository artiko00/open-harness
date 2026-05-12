# @open_harness/secretlens

Secret and credential detector for any codebase. Scans source files for hardcoded AWS keys, GitHub tokens, PEM private keys, JWTs, and generic credential assignments. Single native binary, zero runtime dependencies, works in any language ecosystem.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo.

## Install

```bash
npm install --save-dev @open_harness/secretlens
```

The right native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) is downloaded automatically via `optionalDependencies`.

## Usage

```bash
npx secretlens check              # scan current directory
npx secretlens check --fail       # exit 1 if secrets found (git hooks / CI)
npx secretlens check --dir ./src  # scan a specific directory
npx secretlens check --no-color   # plain output for logs
npx secretlens init               # generate a default secretlens.json
npx secretlens version            # print version
```

## Built-in patterns

| Pattern | Severity |
|---|---|
| AWS Access Key ID (`AKIA…`) | critical |
| AWS Secret Access Key | critical |
| GitHub Personal Access Token (`ghp_…`) | critical |
| GitHub Fine-Grained Token (`github_pat_…`) | critical |
| PEM Private Key (`-----BEGIN … PRIVATE KEY`) | critical |
| JWT Token | high |
| Generic `secret/password/api_key` assignment | high |
| Generic `token/bearer` assignment | medium |

## Configuration

Place a `secretlens.json` at the repo root:

```json
{
  "patterns": [],
  "allowlist": ["example", "placeholder", "your_key_here", "changeme"],
  "exclude": ["node_modules", "vendor", ".git", "dist"]
}
```

- `patterns: []` uses the 8 built-in patterns. Override the array to add custom regexes.
- `allowlist` skips any line containing the listed strings (case-insensitive) — useful to suppress false positives in docs or examples.
- `exclude` skips matching directories entirely.

## Husky / lefthook / CI integration

```bash
# .husky/pre-commit
npx secretlens check --fail
```

```yaml
# .github/workflows/security.yml
- name: Scan for hardcoded secrets
  run: npx @open_harness/secretlens check --fail
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | No secrets detected (or `--fail` not passed) |
| `1` | Secrets found and `--fail` was passed, or config error |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
