# Changelog

All notable changes to `dupelens` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.1] - 2026-08-07

### Fixed

- **`pyproject.toml` real-world parsing**: the shared TOML parser aborted on
  valid files — a multi-line `dependencies = [` (the canonical PEP 621 form)
  produced `unexpected token in array` and took the whole config load down with
  it, even when the offending syntax lived in a section such as `[project]` or
  `[tool.ruff]`. The subset now covers multi-line arrays, trailing commas,
  literal and multi-line strings, `_`-separated and `0x`/`0o`/`0b` integers,
  RFC 3339 dates (read as strings), dotted keys and quoted keys. Unrecognised
  syntax **outside** the tool's own `[tool.<tool>]` table is skipped instead of
  failing; inside it, errors still surface with line and key. See
  [ADR-018](../../docs/adr-018-config-multi-ecosistema.md).

## [0.4.0] - 2026-07-31

Import declarations no longer count as duplicated code (F-022). Reported by a user
whose NestJS monorepo produced 26 matches across 77 files, **all of them `renamed`
and none `exact`** — every match was the import header.

### Added

- **`default.ignoreImports`** (default **`true`**): drops import declarations before
  tokenizing, alongside the comments and string contents that were already dropped.
  Recognition is per language family, by extension, without a parser: JS/TS
  (`import`, `export … from`, `require(…)`), Python, Go, Ruby, Rust, JVM, PHP,
  C/C++/ObjC, C#, Dart, Swift. Multi-line declarations are dropped whole; executable
  statements that merely start with the same word (C#'s `using (…)`, JS's dynamic
  `import('./x')`) are left alone. Set to `false` for the previous behaviour.
- **Kind breakdown in the report.** The console header and `SUMMARY` line now show
  `N exact · M renamed`; the JSON gains `exactCount` and `renamedCount`. Since
  `--fail` only counts `exact` by default, the total alone never said whether the
  gate would trip.

### Changed

- **Low-entropy filter on the `renamed` pass.** Windows where at least 75% of the
  lines start with the same token (3 lines minimum) are dropped: embedded data
  blocks — seed arrays, literal tables — are structurally identical line after line
  and collide with any block of the same shape. The `exact` pass is untouched, so a
  byte-identical repetitive block is still reported and the default gate loses no
  detection power.

### Impact

Existing projects will see **fewer matches**, all of them in the `renamed` bucket.
No match that broke the default `--fail` gate stops breaking it.

## [0.3.2] - 2026-07-29

Coordinated suite release (meta-package packaging fix). No changes to `dupelens` itself.

## [0.3.1] - 2026-07-28

### Added

- **`--tutorial`** flag: prints a static configuration guide to stdout — every
  config key with its default and an example, plus the tool's flags. Exit `0`;
  `--no-color` strips ANSI. Onboarding release (F-020).

## [0.3.0] - 2026-07-27

Part of the `fix-audit-findings` release (adversarial audit F-018). See
[docs/UPGRADING.md](../../docs/UPGRADING.md) and
[ADR-020](../../docs/adr-020-modulos-compartidos-y-duplicacion-estructural.md).

### Added

- **Renamed-clone detection.** In addition to token-for-token `exact` clones,
  `dupelens` now detects alpha-renamed clones (same structure, different
  identifiers) and labels them `renamed`.
- **`--fail-on exact|renamed|all`** (default `exact`) selects which clone kinds
  trip the `--fail` gate.
- **`windowSize`** config key (in `default`, `0` = built-in default of 25): the
  detection window of the rolling hash, now independent of `minTokens`.
- Unified CLI contract: `--format console|json`, `--config <path>`, `--no-color`,
  and `--output <file>` on `init`.

### Changed

- **`minTokens` is now the report threshold only**; detection uses the separate
  `windowSize`. Lowering `minTokens` to reduce noise no longer changes how blocks
  are hashed.
- **`dupelens.json` returns to the honest default `minTokens: 50`** (was `200`,
  which silently hid 38 real duplicate blocks over this repo). The accepted
  structural duplication of the per-tool CLI skeleton is now marked with explicit,
  enumerated `skip` rules instead (see ADR-020).
- **Strict config loading:** an unknown config key now prints a warning to stderr
  and continues, instead of being silently dropped.

### Fixed

- Skipped files are now reported and can break `--fail`; duplication no longer
  "passes green" by omission.
- Shared path-matching, binary detection and config-chain loading were extracted
  to `tools/_shared/*` (`pathmatch`, `langsyntax`, `configload`), removing
  byte-for-byte duplication flagged by the audit.

### BREAKING

- None for the default `--fail` behavior: it still counts only `exact` clones, so
  literal-duplication gates are unchanged. `renamed` clones are new and
  informational unless you opt in with `--fail-on renamed|all`.
