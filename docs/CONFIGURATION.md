# Configuration guide

Central reference for configuring the five open-harness lenses —
[`linelens`](#linelens), [`dupelens`](#dupelens), [`secretlens`](#secretlens),
[`testlens`](#testlens) and [`scopelens`](#scopelens).

Every tool is a single native binary with zero runtime dependencies and works in
any language ecosystem. Configuration is **optional**: run with no config and each
tool uses its compiled defaults. To customize, drop a `<tool>.json` at the repo
root, or use your existing manifest (`pyproject.toml`, `package.json`,
`composer.json`) — see [Config precedence](#config-precedence).

Two shortcuts help you get started:

- **`open-harness init`** — the meta-package command that creates all five
  `<tool>.json` files at the repo root in a single run, delegating to each tool's
  own `init`. It does not silently overwrite: it reports which files it created and
  which already existed. Each tool still has its individual `<tool> init`
  (with `--output <file>` to choose the path).
- **`<tool> --tutorial`** — every tool prints a static, in-terminal guide of its
  own config keys (with defaults and examples), its flags, and the relevant 0.3.0
  behavior changes. Exit `0`; `--no-color` strips ANSI. Same content as this
  document, available where you are working.

---

## Config precedence

Each tool resolves its configuration through the following chain, **per tool**,
stopping at the first source that provides a given field:

| # | Source | When to use |
|---|---|---|
| 1 | **CLI flags** (`--max`, `--min-tokens`, `--max-files`, …) | Always win; ad-hoc overrides |
| 2 | **`<tool>.json`** at the repo root (`linelens.json`, `dupelens.json`, `secretlens.json`, `testlens.json`, `scopelens.json`) | Polyglot projects with no central manifest |
| 3 | **`pyproject.toml`** → `[tool.<tool>]` | Python projects (PEP 518 idiom) |
| 4 | **`package.json`** → `"<tool>": { … }` | Node / TypeScript projects |
| 5 | **`composer.json`** → `"extra": { "open-harness": { "<tool>": … } }` | PHP projects |
| 6 | **Compiled defaults** | Fallback when nothing else matches |

Notes:

- The chain is **per field**: for each config field, the first source in the chain
  that defines it wins; anything left unset is filled from the compiled defaults.
- Arrays (`exclude`, `allowlist`, …) are **atomic**: if you define one, it
  **replaces** the built-in list entirely — it is not merged element by element.
  The one exception is `secretlens` `patterns`, which are **additive** (see below).
- The chain is per-tool, so you can configure `linelens` in `pyproject.toml`,
  `dupelens` in `dupelens.json`, and `secretlens` in `package.json` at the same
  time. See [ADR-018](adr-018-config-multi-ecosistema.md) and
  [ADR-014](adr-014-config-en-package-json.md).
- **Strict loading (0.3.0):** an unknown config key prints
  `warning: config "<file>": clave desconocida "<key>" (ignorada)` to stderr and
  the tool continues. Exit code is unaffected — but the misspelled setting does not
  take effect, so fix the key.

---

## linelens

Reports files that exceed a configured line limit. Counts **lines of code**
(comments and blank lines excluded) as of 0.3.0.

Default config file: `linelens.json`.

| Key | Type | Default | Description |
|---|---|---|---|
| `default.maxLines` | int | `100` | Max lines of code per file. `0` falls back to the default. |
| `default.maxNesting` | int | `0` | Max block-nesting depth to flag. `0` = disabled. |
| `rules[].pattern` | string | — | Glob (`.gitignore` style) selecting files for this rule. |
| `rules[].maxLines` | int | — | Per-pattern line limit overriding `default.maxLines`. |
| `rules[].skip` | bool | `false` | If `true`, matching files are skipped entirely. |
| `exclude` | []string | `node_modules`, `vendor`, `.git`, `dist`, `build`, `coverage`, `__pycache__`, `target`, `.next`, `.nuxt`, `out`, `.cache`, `*.pb.go`, `*_gen.go`, `*.g.dart`, `*-lock.json` | Paths/globs not scanned. Setting it replaces the defaults. |

```json
{
  "default": { "maxLines": 100, "maxNesting": 0 },
  "rules": [
    { "pattern": "**/*_test.go", "maxLines": 300 },
    { "pattern": "**/*.spec.*",  "maxLines": 300 },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": ["node_modules", "vendor", ".git", "dist"]
}
```

---

## dupelens

Detects duplicated code blocks with Rabin-Karp fingerprinting over tokenized
source. Language-agnostic.

Default config file: `dupelens.json`.

| Key | Type | Default | Description |
|---|---|---|---|
| `default.minTokens` | int | `50` | **Report threshold** — matches shorter than this are dropped. `0` falls back to the default. |
| `default.minLines` | int | `5` | Minimum lines for a duplicate block (filters short back-to-back matches). `0` falls back to the default. |
| `default.windowSize` | int | `0` | Rolling-hash **detection window**, independent of `minTokens`. `0` uses the internal default (`25`). |
| `default.ignoreImports` | bool | `true` | Drops import declarations before tokenizing (`import`/`require`/`use`/`#include`/… per language). They are mandatory module-access syntax, not logic: normalized, every file's header collapses to the same token stream. `false` restores the pre-0.4.0 counting. |
| `rules[].pattern` | string | — | Glob selecting files for this rule. |
| `rules[].minTokens` | int | — | Per-pattern token threshold. |
| `rules[].skip` | bool | `false` | If `true`, matching files are skipped. |
| `exclude` | []string | `node_modules`, `vendor`, `.git`, `dist`, `build`, `coverage`, `__pycache__`, `target`, `.next`, `.nuxt`, `out`, `.cache` | Directories not scanned. Setting it replaces the defaults. |

```json
{
  "default": { "minTokens": 50, "minLines": 5, "windowSize": 0, "ignoreImports": true },
  "rules": [
    { "pattern": "**/*_test.go", "skip": true },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": ["node_modules", "vendor", ".git", "dist", "build"]
}
```

The `--fail-on exact|renamed|all` flag (default `exact`) selects which clone kinds
break `--fail`. See [ADR-012](adr-012-dupelens-rabin-karp-sobre-ast.md) and
[ADR-020](adr-020-modulos-compartidos-y-duplicacion-estructural.md).

---

## secretlens

Scans for hardcoded secrets and credentials (AWS keys, GitHub tokens, PEM, JWT,
provider-prefix tokens, connection URIs, and generic `KEY=VALUE` assignments).

Default config file: `secretlens.json`.

| Key | Type | Default | Description |
|---|---|---|---|
| `patterns` | []object | `[]` | Custom detection rules, **added on top of** the built-ins (see note). Each: `{ "name", "pattern", "severity", "entropyGate" }`. |
| `allowlist` | []string | `example`, `placeholder`, `your_key_here`, `changeme`, `xxxx`, `****` | Substrings that mark a **detected value** as a false positive (case-insensitive). Setting it replaces the defaults. |
| `exclude` | []string | `node_modules`, `vendor`, `.git`, `dist`, `build`, `coverage`, `__pycache__`, `target`, `.next`, `out`, `.cache`, `*.lock`, `go.sum` | Paths/globs not scanned. Setting it replaces the defaults. |
| `minEntropy` | float64 | `3.0` | Shannon-entropy threshold (bits/char) the captured value must exceed for the **generic** `KEY=VALUE` rules. `0` falls back to the default. |
| `disableDefaultPatterns` | bool | `false` | If `true`, the built-in patterns are **not** prepended — only your `patterns` run. |

**`patterns` are additive (0.3.0).** Custom entries run alongside the built-ins.
To run *only* your own patterns, set `"disableDefaultPatterns": true`. The
`entropyGate` field marks a custom rule as generic (its captured value is subject
to `minEntropy`); strong-prefix rules leave it `false` and are never gated.
See [ADR-021](adr-021-secretlens-entropia.md).

```json
{
  "patterns": [
    { "name": "corp-token", "pattern": "CORP-[A-Z0-9]{20}", "severity": "high", "entropyGate": true }
  ],
  "disableDefaultPatterns": false,
  "allowlist": ["example", "placeholder", "your_key_here", "changeme"],
  "minEntropy": 3.0,
  "exclude": ["node_modules", "vendor", ".git", "dist"]
}
```

---

## testlens

Finds source files without an associated test, across multiple languages.

Default config file: `testlens.json`.

| Key | Type | Default | Description |
|---|---|---|---|
| `language` | string | `"auto"` | Language to analyze, or `"auto"` to detect. Supported: `go`, `typescript`, `javascript`, `python`, `ruby`, `rust`, `java`, `kotlin`, `csharp`, `dart`. |
| `exclude` | []string | `node_modules`, `.git`, `vendor`, `dist`, `build`, `coverage`, `__pycache__`, `target`, `.next`, `.nuxt`, `out`, `.cache`, `testdata` | Directories/patterns skipped during the scan. Setting it replaces the defaults. |
| `notest` | []string | `__init__.py`, `conftest.py`, `settings.py`, `*_pb2.py`, `*.pb.go`, `*_gen.go`, `*.g.dart`, `main.go`, `doc.go` | Globs (over the file base name) of sources that do not require their own test. `migrations/` directories are always exempt. Setting it replaces the defaults. |

```json
{
  "language": "auto",
  "exclude": ["node_modules", ".git", "dist", "testdata"],
  "notest": ["main.go", "*_gen.go"]
}
```

The `--lang <language>` flag overrides `language`. Go uses **package mode** (a
directory with any `*_test.go` covers all its sources); the per-file ecosystems use
**file mode**. See [ADR-022](adr-022-testlens-package-mode.md).

---

## scopelens

Enforces a per-PR file and (optionally) line budget by counting the
branch-vs-base diff over `git`.

Default config file: `scopelens.json`. Note: `scopelens` does **not** expose
`--format json` or `--config`; the report is a single fixed console format, and
configuration is resolved through the chain above.

| Key | Type | Default | Description |
|---|---|---|---|
| `maxFiles` | int | `15` | The file budget: max countable files in the diff. `0` falls back to the default; **negative is an error** (exit 2). Always on. |
| `maxLines` | int | `0` | The line budget: max churn of the diff. `0` (or absent) **disables** it — the gate then counts files only (backward compatible). Negative is an error (exit 2). |
| `mode` | string | `"or"` | How the file and line budgets combine when `maxLines > 0`: `"or"` fails if **either** is exceeded, `"and"` only if **both** are. Any other value is an error. |
| `lineMetric` | string | `"changed"` | What `maxLines` counts: `"changed"` = added + deleted, `"added"` = only added lines. Any other value is an error. |
| `base` | string | `""` | Base ref to diff against. Empty auto-discovers `origin/HEAD` → `main` → `master`. The `--base` flag overrides it. |
| `excludeTests` | bool | `false` | Discount test files from the count (files **and** lines). The config field **or** the `--exclude-tests` flag turns it on (OR). |
| `exclude` | []string | see below | Globs (`.gitignore` style) for paths that are not review surface and must not consume budget. **Atomic** — setting it replaces the defaults. |

Each config key has a matching flag that overrides it: `--max-files`,
`--max-lines`, `--mode`, `--line-metric`, `--base`, `--exclude-tests`.

Default `exclude` covers common vendor/build dirs and regenerated lockfiles and
generated code across JS/TS, Python and Go: `.git/**`, `node_modules/**`,
`vendor/**`, `dist/**`, `build/**`, `coverage/**`, `package-lock.json`,
`pnpm-lock.yaml`, `yarn.lock`, `.next/**`, `.nuxt/**`, `out/**`,
`**/__snapshots__/**`, `poetry.lock`, `Pipfile.lock`, `uv.lock`,
`**/__pycache__/**`, `*.egg-info/**`, `.venv/**`, `go.sum`, `**/*.pb.go`,
`**/zz_generated*.go`.

```json
{
  "maxFiles": 15,
  "maxLines": 0,
  "mode": "or",
  "lineMetric": "changed",
  "base": "",
  "excludeTests": false,
  "exclude": [
    ".git/**", "node_modules/**", "vendor/**", "dist/**", "build/**",
    "go.sum", "**/*.pb.go", "**/zz_generated*.go"
  ]
}
```

**Exit codes:** `0` = within budget (or over without `--fail`), `1` = over budget
(with `--fail`), `2` = **could not measure** (git missing, not a repo, shallow
clone, base ref unresolvable, invalid config —including a bad `mode`/`lineMetric`—
or a usage error). See
[ADR-023](adr-023-scopelens-dependencia-git.md).

---

## Per-ecosystem examples

The examples below configure every tool from a single manifest. Remember the chain
is per-tool — mix and match freely.

### `pyproject.toml` (Python)

The tables below live alongside whatever else your `pyproject.toml` already
holds — `[project]`, `[tool.poetry]`, `[tool.ruff]` and friends:

```toml
[project]
name = "my-project"
dependencies = [
    "requests>=2.0",
    "pydantic>=2.0",
]

[tool.linelens.default]
maxLines = 100

[tool.linelens]
exclude = [
    "vendor/**",
    "build/**",
]

[tool.testlens]
language = "python"
exclude = ["node_modules", ".git", "dist", ".venv"]

[tool.scopelens]
maxFiles = 20
```

Multi-line arrays, trailing commas, `'literal strings'`, `"""multi-line
strings"""`, dotted keys (`default.maxLines = 100`) and quoted keys are all
understood. Syntax the parser does not recognise **outside** your `[tool.<tool>]`
tables is skipped rather than treated as a failure, so an exotic
`[tool.poetry]` section never stops a tool from reading its own config. Inside
your own tables the opposite holds: a broken value is reported with its line
number instead of silently falling back to defaults. See
[ADR-018](adr-018-config-multi-ecosistema.md) for the supported subset.

### `package.json` (Node / TypeScript)

```json
{
  "name": "my-project",
  "linelens":   { "default": { "maxLines": 100 } },
  "dupelens":   { "default": { "minTokens": 50, "minLines": 5 } },
  "secretlens": { "allowlist": ["example", "placeholder"] },
  "testlens":   { "language": "typescript", "exclude": ["node_modules", "dist"] },
  "scopelens":  { "maxFiles": 25 }
}
```

### `composer.json` (PHP)

```json
{
  "extra": {
    "open-harness": {
      "linelens":  { "default": { "maxLines": 120 } },
      "scopelens": { "maxFiles": 15 }
    }
  }
}
```
