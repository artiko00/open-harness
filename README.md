# open-harness

A monorepo of lightweight code quality tools — each one a single binary, zero runtime dependencies, works in any language ecosystem.

## Tools

| Tool | Description | Status |
|---|---|---|
| [linelens](tools/linelens/) | File length linter — detects files exceeding a line limit | `v0.3.2` |
| [dupelens](tools/dupelens/) | Code duplication detector (Rabin-Karp, language-agnostic) | `v0.4.0` |
| [secretlens](tools/secretlens/) | Secret and credential detector (AWS keys, GitHub tokens, JWT, PEM, etc.) | `v0.3.2` |
| [testlens](tools/testlens/) | Test coverage detector — finds source files without tests (multi-language) | `v0.3.2` |
| [scopelens](tools/scopelens/) | Per-PR file- and line-budget gate — counts the branch-vs-base diff at local pre-commit | `v0.2.0` |
| bigo | Big O complexity analyzer | `planned` |

---

## Migration / breaking changes (0.3.0)

The `0.2.x → 0.3.0` bump is designed to be backward-compatible — existing configs, flags and hooks keep working. The main thing to plan for is that the tools now analyze **more accurately** and may surface findings that were previously missed (especially `secretlens`, whose recall went from ~25% to ~100%). Run each tool once without `--fail` before enforcing it in CI.

The one behaviour change to know about:

- **secretlens — custom patterns are now additive.** Entries in `patterns` are **added on top of** the built-in patterns instead of replacing them. Previously a non-empty `patterns` array silently disabled all built-ins (a security footgun). To keep the old "only my patterns" behaviour, set `"disableDefaultPatterns": true`.

Compatibility-preserving refinements: unknown config keys now print a **warning** (instead of being silently dropped) but do not fail; the `secretlens` allowlist matches the detected **value** rather than the whole line, and still includes `"example"` by default.

**See [docs/UPGRADING.md](docs/UPGRADING.md) for the full upgrade guide**, including how to triage newly-surfaced findings in `secretlens`, `testlens` and `linelens`.

Changelogs are per package: the suite [CHANGELOG.md](CHANGELOG.md) plus one per tool ([linelens](tools/linelens/CHANGELOG.md), [dupelens](tools/dupelens/CHANGELOG.md), [secretlens](tools/secretlens/CHANGELOG.md), [testlens](tools/testlens/CHANGELOG.md), [scopelens](tools/scopelens/CHANGELOG.md)).

---

## Configuration

All five tools share a single, per-tool configuration model: an optional
`<tool>.json` at the repo root, or your existing `pyproject.toml` /
`package.json` / `composer.json`. **[docs/CONFIGURATION.md](docs/CONFIGURATION.md)**
is the central reference — every config key with its default and description, the
multi-ecosystem precedence chain, and per-ecosystem examples.

Two shortcuts:

- **`open-harness init`** (meta-package) creates all five `<tool>.json` config
  files at the repo root in one run, delegating to each tool's own `init`. It does
  not silently overwrite — it reports which files it created and which already
  existed. Each tool still has its individual `<tool> init`.
- **`<tool> --tutorial`** prints each tool's configuration guide right in the
  terminal (keys, defaults, flags, and the relevant 0.3.0 behavior changes).

---

## linelens

Scans your project and reports files that exceed a configured line limit. Works with any language.

### Usage

```bash
linelens check               # scan current directory
linelens check --fail        # exit 1 if violations found (CI / git hooks)
linelens check --dir ./src   # scan a specific directory
linelens check --max 200     # override the line limit
linelens check --config ci.json  # use a specific config file (must exist)
linelens check --no-color    # plain output for logs
linelens check --format json # machine-readable output (console | json)
linelens init                # generate a default linelens.json
linelens init --output custom.json  # write the config to a different file
linelens --tutorial          # print the in-terminal configuration guide
```

**Common flags** (shared by all tools): `--config <file>` (defaults to `<tool>.json`; when passed explicitly the file must exist), `--no-color`, `--format console|json`, and `--dir <path>`. Every `init` accepts `--output <file>` to choose the generated config path. Every tool also has `--tutorial`, which prints a static configuration guide (each config key with its default and an example) to stdout — see [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for the same content in one place.

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
secretlens check --config ci.json  # use a specific config file (must exist)
secretlens check --no-color   # plain output for logs
secretlens check --format json  # machine-readable output (console | json)
secretlens init               # generate a default secretlens.json
secretlens init --output custom.json  # write the config to a different file
```

**Common flags** (shared by all tools): `--config <file>`, `--no-color`, `--format console|json`, `--dir <path>`, and `--output <file>` on `init`. See the [linelens](#linelens) section for details.

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
  "disableDefaultPatterns": false,
  "allowlist": ["example", "placeholder", "your_key_here", "changeme"],
  "exclude": ["node_modules", "vendor", ".git", "dist"]
}
```

Custom `patterns` are **additive** — they run alongside the built-in patterns. Set `"disableDefaultPatterns": true` to run only your own patterns. (This changed in 0.3.0 — see [Migration / breaking changes](#migration--breaking-changes-030).)

The `allowlist` suppresses a finding when the **detected secret value** (not the whole line, case-insensitive) matches one of the terms — useful for placeholders in documentation or example files. Note that `"example"` is no longer part of the default allowlist; add it explicitly if you need it.

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

# Use a specific config file (must exist), plain output, or JSON
testlens check --config ci.json
testlens check --no-color
testlens check --format json

# Generate a default config (optionally to a custom path)
testlens init
testlens init --output custom.json
```

**Common flags** (shared by all tools): `--config <file>`, `--no-color`, `--format console|json`, `--dir <path>`, and `--output <file>` on `init`. See the [linelens](#linelens) section for details.

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
| Dart | `.dart` | `*_test.dart` |

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
npm install --save-dev @open_harness/dupelens
```

### Usage

```bash
# Scan current directory with defaults
dupelens check

# Fail with exit code 1 if duplicates found (CI / git hooks)
dupelens check --fail

# Choose which clone kinds break --fail: exact | renamed | all
dupelens check --fail --fail-on renamed

# Override the token threshold
dupelens check --min-tokens 30

# Use a specific config file (must exist), or plain output
dupelens check --config ci.json
dupelens check --no-color

# JSON output for tooling integrations
dupelens check --format=json > report.json

# Verbose timings to stderr
dupelens check --verbose

# Generate default config (optionally to a custom path)
dupelens init
dupelens init --output custom.json
```

**Common flags** (shared by all tools): `--config <file>`, `--no-color`, `--format console|json`, `--dir <path>`, and `--output <file>` on `init`. See the [linelens](#linelens) section for details.

`--fail-on` selects which clone kinds trip the `--fail` gate: `exact` (default, only token-for-token clones), `renamed` (also alpha-renamed clones), or `all`.

### Configuration (`dupelens.json`)

```json
{
  "default": {
    "minTokens": 50,
    "minLines": 5,
    "windowSize": 0,
    "ignoreImports": true
  },
  "rules": [
    { "pattern": "**/*_test.go",     "skip": true },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": ["node_modules", "vendor", ".git", "dist", "build"]
}
```

`minTokens` is the **report threshold** — matches shorter than this are dropped. `windowSize` is the **detection window** of the rolling hash and is now independent of `minTokens` (`0` falls back to the built-in default), so you can lower `minTokens` to reduce noise without changing how blocks are hashed. `minLines` filters short matches.

`ignoreImports` (default `true`) drops import declarations before tokenizing, the same way comments and string contents are already dropped: they are mandatory module-access syntax, not logic. This matters in modular codebases — a NestJS file opens with 5–15 `import { X } from 'Y';` lines, and once identifiers are normalized for renamed-clone detection, every file's header collapses to the same token stream, so any two files match. Set it to `false` to restore the pre-0.4.0 behaviour.

Recognition is per language family, by file extension, with no parser involved — JS/TS (`import`, `export … from`, `require(…)`), Python (`import`, `from … import`), Go (`package`, `import ( … )`), Ruby, Rust, JVM, PHP, C/C++/ObjC, C#, Dart and Swift. Multi-line declarations are dropped whole. Executable statements that merely start with the same word — C#'s `using (var s = …)`, JS's dynamic `import('./x')` — are left alone.

### Output (console)

```
DUPLICATES (2 match(es) (1 exact · 1 renamed) found in 87 files):

  src/auth.go:42-58  ↔  src/users.go:12-28  (35 tokens, exact)
  | func validate(input string) error {
  | ...
  src/db.go:1-10  ↔  src/cache.go:1-10  (15 tokens, renamed)

SUMMARY: 2 match(es) (1 exact · 1 renamed) across 87 files
Top duplicated files:
  - src/auth.go  (1 match(es))
```

The `exact · renamed` breakdown tells you at a glance whether `--fail` will trip: by default the gate only counts `exact`, so a long list of `renamed` findings alongside a green gate is expected, not a contradiction.

### Output (JSON)

```json
{
  "scannedFiles": 87,
  "matchCount": 2,
  "exactCount": 1,
  "renamedCount": 1,
  "matches": [
    { "fileA": "src/auth.go", "startLineA": 42, "endLineA": 58,
      "fileB": "src/users.go", "startLineB": 12, "endLineB": 28,
      "tokens": 35, "kind": "exact" }
  ],
  "summary": {
    "topDuplicatedFiles": [{ "file": "src/auth.go", "count": 1 }]
  }
}
```

### Low-entropy filter

Embedded data blocks — seed arrays, literal tables, constant lists — are structurally identical line
after line, so under identifier normalization they collide with any other block of the same shape. The
`renamed` pass therefore drops windows where at least 75% of the lines start with the same token (3
lines minimum). The `exact` pass keeps them: a byte-identical twenty-`case` `switch` is a genuine
finding, and since `--fail` defaults to `exact` only, the gate loses no detection power.

The filter keys on the *first token of each line*, so a data block whose lines start with distinct
identifiers (`entry_1(…)`, `entry_2(…)`) is not covered — exclude those paths in `dupelens.json`.

### Limitations (v0.4.0)

- Detects contiguous **exact** and **alpha-renamed** clones (see `--fail-on`). Structural refactors — reordered statements, inserted or deleted lines (gapped/Type-3 clones), and behaviourally-equivalent rewrites (Type-4) — are still not detected; that requires AST analysis ([ADR-012](docs/adr-012-dupelens-rabin-karp-sobre-ast.md) explains the trade-off).

---

## scopelens

Enforces a **per-PR file budget** ("no more than 15 files touched"). The unit of the policy is the **PR, not the commit**: scopelens counts the *union* of everything already committed on the branch (`git diff <merge-base>...HEAD`) and everything staged (`git diff --cached`). Five commits of 4 files each add up to 20 and fail.

It runs **locally, over `git`, with no network and no token**, and aborts the `git commit` **before the PR even exists**. That is the key difference from the CI tools it replaces (see the comparison below).

### Why not a CI size-labeler?

| Tool | Where it runs | What it does | Can block | Reads the diff via |
|---|---|---|---|---|
| `pr-size-labeler` | GitHub Actions, on the open PR | Labels `size/xs…xl` from diff + file count | **No** — only labels; blocking needs branch protection on the label | GitHub API |
| `Danger JS` | CI, needs `DANGER_GITHUB_API_TOKEN` + Node | Programmable rule in `dangerfile.ts`; only `fail()` breaks the build | Yes, but only in CI | GitHub API |
| **scopelens** | **Local pre-commit** | Counts the branch-vs-base diff and aborts the commit | **Yes, before the PR exists** | **`git` (local)** |

Both CI tools share three limitations that matter for a real gate: they **arrive late** (feedback minutes after the push, with the work already pushed), they **read the diff through the GitHub API** — which truncates to 300 files (HTTP 406) and the file list to 3000, so the budget breaks exactly on the large PRs the gate is meant to catch — and they **require network, a token and an open PR**, none of which exist at `git commit` time. They also only work inside GitHub Actions + Node; a Python or Go team installing from PyPI or `go install` will not stand up a Node runtime just to count files.

### Usage

```bash
scopelens check                       # measure the current branch vs its base
scopelens check --fail                # exit 1 if the budget is exceeded (git hooks / CI)
scopelens check --max-files 20        # override the budget for this run
scopelens check --base develop        # compare against an explicit base ref
scopelens check --staged-only         # count only the index (staged changes)
scopelens check --exclude-tests       # discount test files from the budget
scopelens check --dir ./repo          # run against a specific repository directory
scopelens check --no-color            # plain output for logs
scopelens init                        # generate a default scopelens.json
scopelens init --output custom.json   # write the config to a different file
```

Note: scopelens does **not** expose `--format json` or `--config` — the report is a single fixed console format, and configuration is resolved through the standard chain (see below).

### Exit codes

Unlike the other four tools (which only use `0`/`1`), scopelens adds **exit code 2** for "could not measure" — it never invents a count when it lacks the information to trust one.

| Code | Meaning |
|---|---|
| `0` | Measured and within budget (or over budget without `--fail`) |
| `1` | Measured and over budget (with `--fail`) |
| `2` | **Could not measure**: `git` missing from `PATH`, cwd is not a repo, shallow clone (`merge-base` unresolvable), base ref not found, invalid config, or a usage error |

### Configuration (`scopelens.json`)

```json
{
  "maxFiles": 15,
  "base": "",
  "excludeTests": false,
  "exclude": [
    ".git/**", "node_modules/**", "vendor/**", "dist/**", "build/**",
    "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
    "poetry.lock", "Pipfile.lock", "uv.lock",
    "go.sum", "**/*.pb.go", "**/zz_generated*.go"
  ]
}
```

- `maxFiles` — the budget (default `15`). `0` falls back to the default.
- `base` — base ref to diff against; empty means auto-discover (`origin/HEAD` → `main` → `master`). The `--base` flag overrides it.
- `excludeTests` — discount test files from the count. Either the config field **or** the `--exclude-tests` flag turns it on.
- `exclude` — glob patterns (`.gitignore` style) for paths that are **not review surface** and must not consume budget. The defaults already cover regenerated lockfiles and generated code across JS/TS, Python and Go; setting your own replaces the defaults.

Configuration is resolved through the same per-tool chain as the other tools (`scopelens.json` → `pyproject.toml [tool.scopelens]` → `package.json "scopelens"` → `composer.json extra.open-harness.scopelens` → compiled defaults). See [Config sources by ecosystem](#config-sources-by-ecosystem).

### lefthook integration

```yaml
pre-commit:
  commands:
    scopelens: { run: tools/scopelens/scopelens check --fail --no-color }
```

Exit `2` (could-not-measure) also blocks the commit, so a broken measurement is never silently treated as a pass. To bypass the gate deliberately, use `git commit --no-verify`, as with the other lenses.

---

## Usage in Node/TS/JS projects

Install all four tools at once via the meta-package:

```bash
npm install --save-dev @open_harness/open-harness
```

Or install individual tools:

```bash
npm install --save-dev @open_harness/linelens @open_harness/dupelens \
  @open_harness/secretlens @open_harness/testlens
```

Then scaffold every config file at once:

```bash
open-harness init   # creates linelens.json, dupelens.json, secretlens.json,
                    # testlens.json and scopelens.json at the repo root
```

`open-harness init` never overwrites silently: it reports which files it created
and which already existed. Each tool still has its own `<tool> init`.

### Configure everything from `package.json`

Each tool reads its configuration from a dedicated key inside your `package.json`:

```json
{
  "name": "my-project",
  "scripts": {
    "lint:lines":   "linelens   check --fail",
    "lint:dupes":   "dupelens   check --fail",
    "lint:secrets": "secretlens check --fail",
    "lint:tests":   "testlens   check --fail",
    "lint":         "npm run lint:lines && npm run lint:dupes && npm run lint:secrets && npm run lint:tests"
  },
  "linelens":   { "default": { "maxLines": 100 } },
  "dupelens":   { "default": { "minTokens": 50, "minLines": 5 } },
  "secretlens": { "allowlist": ["example", "placeholder"] },
  "testlens":   { "language": "typescript", "exclude": ["node_modules", "dist"] }
}
```

### Config sources by ecosystem

Each tool resolves its configuration through the following chain, stopping at the first match:

| # | Source | When to use |
|---|---|---|
| 1 | CLI flags (`--max`, `--min-tokens`, …) | Always win, useful for ad-hoc overrides |
| 2 | `<tool>.json` at repo root (`linelens.json`, `dupelens.json`, …) | Polyglot projects with no central manifest |
| 3 | `pyproject.toml` → `[tool.<name>]` | Python projects (PEP 518 idiom) |
| 4 | `package.json` → `"<name>": { ... }` | Node / TypeScript projects |
| 5 | `composer.json` → `"extra": { "open-harness": { "<name>": ... } }` | PHP projects |
| 6 | Built-in defaults | Fallback when nothing else matches |

Mix freely — the chain is **per-tool**. You can have `linelens` configured in `pyproject.toml`, `dupelens` in `linelens.json`, and `secretlens` in `package.json` simultaneously. The precedence above is documented in [ADR-018](docs/adr-018-config-multi-ecosistema.md) (with [ADR-014](docs/adr-014-config-en-package-json.md) covering the original `package.json` fallback).

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
  run: npm install -g @open_harness/open-harness

- run: linelens   check --fail
- run: dupelens   check --fail
- run: secretlens check --fail
- run: testlens   check --fail --lang typescript --dir src/
```

> Configuration for all four lints lives inside `package.json` under the `linelens`, `dupelens`, `secretlens` and `testlens` keys — no extra `*.json` files needed.

---

## Combined CI / git hooks

Run all four quality tools as gates in a single workflow:

```yaml
# GitHub Actions (via npx, no install step)
- run: npx @open_harness/linelens check --fail
- run: npx @open_harness/dupelens check --fail
- run: npx @open_harness/secretlens check --fail
- run: npx @open_harness/testlens check --fail
```

Via lefthook (this repo uses this pattern — see `lefthook.yml`):

```yaml
pre-commit:
  commands:
    linelens:   { run: tools/linelens/linelens check --fail --no-color }
    dupelens:   { run: tools/dupelens/dupelens check --fail --no-color }
    secretlens: { run: tools/secretlens/secretlens check --fail --no-color }
    testlens:   { run: tools/testlens/testlens check --lang go --dir tools/ --fail }

pre-push:
  parallel: true
  commands:
    test-linelens:   { run: cd tools/linelens && go test ./... }
    test-dupelens:   { run: cd tools/dupelens && go test ./... }
    test-secretlens: { run: cd tools/secretlens && go test ./... }
    test-testlens:   { run: cd tools/testlens && go test ./... }
```

The `pre-commit` hook runs all four lints; `pre-push` runs the test suite of each of the four tools in parallel.

---

## Repository structure

```
open-harness/
├── tools/
│   ├── linelens/          ← v0.3.2 (file length linter, 100% coverage)
│   ├── dupelens/          ← v0.4.0 (duplicate detector, Rabin-Karp, 100% coverage)
│   ├── secretlens/        ← v0.3.2 (secret/credential detector, 100% coverage)
│   └── testlens/          ← v0.3.2 (test coverage detector, multi-language, 100% coverage)
├── npm/
│   ├── open-harness/      ← meta-package (instala los 4 tools)
│   └── @open_harness/
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
├── lefthook.yml           ← git hooks (pre-commit: 4 tools, pre-push: tests x4)
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