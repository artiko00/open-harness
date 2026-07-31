# Upgrading to 0.3.0

Version 0.3.0 unifies the four tools (`linelens`, `dupelens`, `secretlens`, `testlens`) and fixes a
large batch of audit findings. **Most existing setups keep working unchanged.** This guide covers the
few behaviors that change and how to handle them.

The design goal was backward compatibility: config files, flags, and hook commands from 0.2.x continue
to work. What changed falls into two buckets — a couple of small compatibility-preserving refinements,
and the fact that the tools now analyze *better* and may surface issues that were previously hidden.

---

## TL;DR

| Area | What changed | Do you need to act? |
|---|---|---|
| Config files | Unknown keys now print a **warning** (were silently ignored) | No — just informational |
| `secretlens` | Detects **many more** real secrets (recall went from ~25% to ~100%) | Maybe — run before enforcing in CI |
| `secretlens` | Custom `patterns` now **add to** the built-ins instead of replacing them | Only if you relied on replacement |
| `testlens` | A test file must **contain a test** to count as coverage | Maybe — empty test files no longer pass |
| `linelens` | Counts **lines of code** (not physical lines) + optional nesting metric | Maybe — different files may be flagged |
| `dupelens` | `--fail` counts **exact** clones by default; `renamed` are reported but not failing | Only if you want renamed clones to fail CI |

If your CI runs any tool with `--fail`, **run 0.3.0 once without `--fail`** (or read its output) before
enforcing, so newly-surfaced findings don't break your pipeline unexpectedly.

---

## Fully backward-compatible refinements

These changed internally but should not break existing configs:

- **Unknown config keys are a warning, not an error.** In 0.2.x a typo like `"excludes"` (instead of
  `"exclude"`) was silently dropped — your setting simply didn't apply. In 0.3.0 the tool prints
  `warning: config "<file>": clave desconocida "excludes" (ignorada)` to stderr and continues with the
  keys it understood. Exit code is unaffected. Fix the key to make the setting take effect.

- **`secretlens` allowlist is matched against the detected value, not the whole line.** The default
  allowlist still includes `example`, `placeholder`, `your_key_here`, etc. A placeholder *value* like
  `KEY=example_value` is still suppressed. What no longer suppresses a finding is an unrelated mention
  elsewhere on the line — e.g. `AWS_KEY="AKIA..." # see example above` is now correctly reported.

- **`dupelens` `minTokens` is now the report threshold only**; detection uses a separate `windowSize`
  (default 25). Existing `dupelens.json` files with a `minTokens` value keep behaving as a report
  threshold. No change required.

---

## Behavior changes that may surface new findings

None of these are bugs — the tools got more accurate. But a repository that "passed" on 0.2.x may now
report issues that were previously missed. Plan for triage before enforcing `--fail` in CI.

### secretlens: much higher recall

0.2.x missed most secrets (no entropy analysis, quotes required around values, no provider prefixes).
0.3.0 detects unquoted `KEY=value` assignments, Shannon-entropy-filtered generic secrets, and provider
formats (`sk_live_`, `xoxb-`, `AIza`, `sk-proj-`, `glpat-`, `npm_`, `SG.`, `postgres://user:pass@…`,
etc.), and decodes UTF-16 files.

**Impact:** real secrets that 0.2.x silently missed will now be reported. This is the point — but if
your repo has such secrets, `secretlens check --fail` will start failing.

**What to do:** run `secretlens check` (without `--fail`) on 0.3.0, review the findings, rotate/remove
any real secrets, and add genuine false positives to `allowlist` or `exclude`. Then re-enable `--fail`.

### secretlens: custom patterns are additive (**breaking if you relied on replacement**)

In 0.2.x, defining any custom `patterns` **silently disabled all eight built-in rules** — a security
footgun the audit flagged. In 0.3.0 custom patterns are **added to** the built-ins.

**Impact:** if you added a custom pattern expecting *only* it to run, you now also get the built-in
detections (AWS keys, GitHub tokens, PEM, JWT, …).

**What to do:** this is almost always what you want. If you truly need only your custom patterns, set
`"disableDefaultPatterns": true` in `secretlens.json` to restore the old replace-behavior.

### testlens: content verification

0.2.x counted a source file as covered if a test file merely *existed*, so an empty `x_test.go` passed.
0.3.0 requires the test file to contain at least one test marker (`func Test`, `it(`, `test(`,
`def test_`, `#[test]`).

**Impact:** packages whose "tests" were empty or markerless are now reported as untested.

**What to do:** add real tests, or exclude intentionally-untestable files (see the default exclusions
for `__init__.py`, `*_pb2.py`, `*.pb.go`, generated files, migrations, etc.).

Also note: language auto-detection is now **deterministic** (0.2.x could return different results
between runs). If your CI saw flaky testlens results, they should stabilize.

### linelens: counts lines of code

0.2.x counted physical lines, so a 400-line license header triggered a violation while a dense 90-line
file passed. 0.3.0 counts **lines of code** (comments and blanks excluded) and reports both the code
count and the physical total.

**Impact:** files that violated only because of comment/blank bulk no longer violate; genuinely large
code files still do. Data files (`.json`, `.sql`, …) and generated files are no longer scanned.

**What to do:** nothing required. Optionally set `"maxNesting"` in `linelens.json` to also flag files
with deep block nesting (0 = disabled, the default).

### dupelens: exact vs renamed clones

0.3.0 detects clones with renamed identifiers (`renamed`) in addition to literal copies (`exact`).
By default `--fail` counts only `exact` clones, so renamed clones are reported but do not fail CI.

**Impact:** `--fail` behavior is unchanged for literal duplication; renamed clones are new, informational.

**What to do:** to make renamed clones fail too, use `--fail-on=all` (or `--fail-on=renamed`).

---

## New flags and config keys (all optional)

- All tools: `--format console|json`, `--config <path>`, `--no-color`, and `--output` (for `init`).
- `secretlens`: `minEntropy` (default 3.0), `disableDefaultPatterns` (default false).
- `dupelens`: `windowSize`, `--fail-on=exact|renamed|all` (default `exact`).
- `linelens`: `maxNesting` (default 0 = disabled).

See the [README](../README.md) for the full flag and config reference.

---

# Upgrading to dupelens 0.4.0

`dupelens` 0.4.0 stops counting import declarations as duplicated code. Nothing to configure —
**existing setups keep working and report fewer matches**, all of them in the `renamed` bucket.

## dupelens 0.4.0

### Imports no longer count as code

The tokenizer already dropped comments and string contents. It now drops import declarations too:
`import … from`, `export … from`, `require(…)`, `from … import`, `use`, `using`, `#include`, `package`
and their multi-line forms, recognized per language family by file extension.

**Why:** with string contents already blanked, `import { UserService } from './user.service';` reduces
to `import from`, and once identifiers are normalized for renamed-clone detection it becomes
`import ID from` — the same token stream in **every file of the project**. A modular codebase opens
each file with 5–15 imports, so with the default 25-token window any two headers collide. The better
modularized the project, the more noise the tool produced.

**Impact:** fewer `renamed` matches. `exact` matches that consisted purely of import headers also
disappear — those were never real duplication.

**What to do:** nothing. To restore the previous counting, set:

```json
{ "default": { "ignoreImports": false } }
```

### The report breaks findings down by kind

The console header and `SUMMARY` line now read `N match(es) (X exact · Y renamed)`, and the JSON
output gains `exactCount` and `renamedCount` next to `matchCount`.

**Why:** `--fail` counts only `exact` findings by default, so a report listing 26 renamed matches with
a green gate looked like a contradiction until you read all 26 lines.

**Impact:** additive. The JSON schema gains two fields; existing consumers are unaffected.

### Repetitive data blocks no longer produce renamed clones

The `renamed` pass now drops windows where at least 75% of the lines start with the same token
(3 lines minimum) — embedded seed arrays, literal tables and constant lists, which are structurally
identical line after line and therefore collide with any block of the same shape.

The `exact` pass is **not** filtered: a byte-identical repetitive block is genuine duplication, and
since `--fail` defaults to `exact` only, the gate loses no detection power.

**Impact:** fewer `renamed` matches over data-heavy files.

**What to do:** nothing. The filter keys on the first token of each line, so a data block whose lines
start with distinct identifiers is not covered — exclude those paths in `dupelens.json` as before.
