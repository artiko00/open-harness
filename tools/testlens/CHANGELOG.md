# Changelog

All notable changes to `testlens` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-07-27

Part of the `fix-audit-findings` release (adversarial audit F-018). See
[docs/UPGRADING.md](../../docs/UPGRADING.md) and
[ADR-022](../../docs/adr-022-testlens-package-mode.md).

### Added

- **Test-content verification.** A test file must now contain at least one test
  marker (`func Test`, `it(`, `test(`, `def test_`, `#[test]`) to count as
  coverage — an empty `x_test.go` no longer passes.
- **Package mode for Go.** A directory with at least one `*_test.go` covers all
  source files in that directory (Go runs tests per package, not per file),
  eliminating massive false positives. File mode remains the default for the
  per-file ecosystems (Python, TypeScript, JavaScript, Ruby, Rust, Java, Kotlin,
  C#).
- **`notest`** config key: glob patterns (over the file base name) of sources that
  do not require their own test. Defaults cover `__init__.py`, `conftest.py`,
  `settings.py`, `*_pb2.py`, `*.pb.go`, `*_gen.go`, `*.g.dart`, `main.go`,
  `doc.go`, plus `migrations/` directories.
- **`--tutorial`** flag: prints a static configuration guide to stdout; exit `0`,
  `--no-color` strips ANSI.
- Unified CLI contract: `--format console|json`, `--config <path>`, `--no-color`,
  and `--output <file>` on `init`.

### Changed

- **Language auto-detection is now deterministic** — 0.2.x could return different
  results between runs; flaky CI results should stabilize.
- `testdata/` is skipped by default, matching the Go toolchain convention.
- **Strict config loading:** an unknown config key now prints a warning to stderr
  and continues; the legacy `"skip"` key is called out with a hint to use
  `"exclude"`.

### Fixed

- Skipped files (binaries, etc.) are reported in the summary and under the JSON
  `skipped` key.

### BREAKING

- **Empty/markerless test files no longer count as coverage.** Packages whose
  "tests" were empty are now reported as untested. Add real tests, or list
  intentionally-untestable files under `notest`/`exclude`. Run once without
  `--fail` before enforcing in CI.
