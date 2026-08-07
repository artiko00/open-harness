# Changelog — open-harness (suite)

All notable changes to the **open-harness** suite (the meta-package and the tools
it bundles) are documented in this file. Per-tool detail lives in each tool's own
changelog:
[linelens](tools/linelens/CHANGELOG.md) ·
[dupelens](tools/dupelens/CHANGELOG.md) ·
[secretlens](tools/secretlens/CHANGELOG.md) ·
[testlens](tools/testlens/CHANGELOG.md) ·
[scopelens](tools/scopelens/CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.5] - 2026-08-07

Suite-wide fix in the shared TOML parser: every tool is republished
(`linelens` 0.3.3, `dupelens` 0.4.1, `secretlens` 0.3.3, `testlens` 0.3.3,
`scopelens` 0.2.1).

### Fixed

- **`pyproject.toml` in real Python projects no longer aborts config loading**
  (F-023). Reported by a user whose valid `pyproject.toml` — one `tomllib` reads
  without complaint — made all five tools fail with `tomlmin: line 7 (in
  "project"): for key "dependencies": unexpected token in array`. The shared
  parser walked the document line by line, so a multi-line `dependencies = [`
  (the canonical PEP 621 form) broke the whole load, even though the syntax sat
  in a section no tool cares about.

  The TOML subset now covers multi-line arrays, trailing commas, literal and
  multi-line strings, `_`-separated and `0x`/`0o`/`0b` integers, RFC 3339 dates
  (read as strings), dotted keys and quoted keys. On top of that, unrecognised
  syntax **outside** the tool's own `[tool.<tool>]` table is skipped instead of
  aborting; inside that table errors still surface with line and key, so a
  broken config never falls back to defaults in silence. See
  [ADR-018](docs/adr-018-config-multi-ecosistema.md).

## [0.3.4] - 2026-07-31

Meta-package release pinning `dupelens` 0.4.0. `dupelens` now runs on its **own
version cycle**, like `scopelens`: `linelens`, `secretlens` and `testlens` stay at
0.3.2 and are not republished.

### Changed

- **`dupelens` 0.4.0**: import declarations no longer count as duplicated code
  (`ignoreImports`, default `true`); the report breaks findings down into
  `exact` / `renamed` in console and JSON; the `renamed` pass drops repetitive
  data blocks. See [dupelens/CHANGELOG.md](tools/dupelens/CHANGELOG.md) and
  [UPGRADING](docs/UPGRADING.md#dupelens-040).
- `scripts/check-versions.sh`: `SOLO_TOOLS` generalizes the own-version-cycle
  handling that existed only for `scopelens`, and now covers `dupelens` too.

## [0.3.2] - 2026-07-29

### Fixed

- **The meta-package now bundles `scopelens`**: added the `scopelens` bin
  shim and its `bin` entry so the command is exposed on install (0.3.1 shipped
  the dependency but not the shim), and listed it across the README.
- Corrected the README's registry claims: the all-in-one meta is **npm-only**;
  on PyPI install the per-tool `open-harness-<tool>` packages (Packagist is planned).

## [0.1.0-scopelens] - 2026-07-28

The suite gains its **fifth tool, `scopelens`** (F-019), released at `0.1.0`. It
is a per-PR file-budget gate that runs locally at pre-commit over `git`, counting
the branch-vs-base diff and aborting the commit before the PR exists — with a
dedicated **exit code 2** for "could not measure". The meta-package
(`open-harness 0.3.0`) now advertises five tools. See the
[scopelens changelog](tools/scopelens/CHANGELOG.md) and
[ADR-023](docs/adr-023-scopelens-dependencia-git.md).

## [0.3.1] - 2026-07-28

### Added

- **`open-harness init`**: new suite command that creates all five `<tool>.json`
  config files at the project root in one run, delegating to each tool's `init`
  (does not overwrite existing files).
- **`--tutorial`** on all five tools: prints a static, per-tool configuration
  guide in the terminal. Onboarding release (F-020).

## [0.3.0] - 2026-07-27

The `fix-audit-findings` release (adversarial audit F-018): the four original
tools were unified and a large batch of audit findings fixed. The bump is designed
to be backward compatible — existing configs, flags and hooks keep working — but
the tools now analyze **more accurately** and may surface findings that were
previously hidden. See [docs/UPGRADING.md](docs/UPGRADING.md).

### Added

- **Unified CLI contract** across the four tools: `--format console|json`,
  `--config <path>`, `--no-color`, and `--output <file>` on `init`.
- **Shared internal modules** in `tools/_shared/*` (`pathmatch`, `langsyntax`,
  `configload`), removing byte-for-byte duplication across tools (ADR-020).
- **secretlens:** Shannon-entropy filter (`minEntropy`), `disableDefaultPatterns`,
  and provider-prefix patterns (Stripe, Slack, Google, OpenAI, GitLab, npm,
  SendGrid, connection URIs).
- **testlens:** package mode for Go, `notest` config key.
- **linelens:** lines-of-code counting and optional `maxNesting`.
- **dupelens:** renamed-clone detection, `--fail-on exact|renamed|all`, and a
  `windowSize` decoupled from `minTokens`.

### Changed

- **Strict config loading everywhere:** an unknown config key now prints a warning
  to stderr and continues, instead of being silently dropped. Exit code unaffected.
- Tools stopped "passing green": skipped files are reported, and a gate that could
  not run is no longer silently treated as a pass.
- `dupelens.json` returned to the honest default `minTokens: 50` (was `200`).

### BREAKING

The three behavior changes to plan for (full detail in
[docs/UPGRADING.md](docs/UPGRADING.md)):

1. **secretlens — custom `patterns` are additive.** They now run alongside the
   built-ins instead of silently replacing them; set `"disableDefaultPatterns":
   true` to restore replace-behavior.
2. **secretlens — recall ~25% → ~100%.** Real secrets missed before are now
   reported; `--fail` may start failing on a repo that "passed" before.
3. **testlens — content verification.** A test file must contain a real test
   marker to count as coverage; empty test files no longer pass.

Run each tool once **without `--fail`** before enforcing in CI.
