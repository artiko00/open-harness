# Changelog

All notable changes to `linelens` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-07-27

Part of the `fix-audit-findings` release (adversarial audit F-018). See
[docs/UPGRADING.md](../../docs/UPGRADING.md) for the full upgrade guide.

### Added

- **Lines-of-code counting.** The report now shows both the code count (comments
  and blank lines excluded) and the physical total for each file.
- **`maxNesting` metric** (in `default`, default `0` = disabled): optionally flag
  files whose deepest block nesting exceeds the configured depth.
- **`--tutorial`** flag: prints a static configuration guide (every config key
  with its default and an example) to stdout; exit `0`, `--no-color` strips ANSI.
- Unified CLI contract shared with the other tools: `--format console|json`,
  `--config <path>`, `--no-color`, and `--output <file>` on `init`.

### Changed

- **Counts lines of code, not physical lines.** A 400-line license header no
  longer triggers a violation, while a dense 90-line code file still can.
- Data files (`.json`, `.sql`, …) and generated files are no longer scanned.
- **Strict config loading:** an unknown config key now prints
  `warning: config "<file>": clave desconocida "<key>" (ignorada)` to stderr and
  continues, instead of being silently dropped. Exit code is unaffected.

### Fixed

- Skipped files are now reported; a `skip` rule that gates no longer "passes
  green" (silent approval by omission) under `--fail`.

### BREAKING

- None. The `0.2.x → 0.3.0` bump is backward compatible for `linelens`: existing
  `linelens.json` files, flags and hooks keep working. Because the metric changed
  from physical lines to lines of code, **a different set of files may be flagged**
  — run once without `--fail` before enforcing in CI.
