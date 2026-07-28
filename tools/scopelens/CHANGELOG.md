# Changelog

All notable changes to `scopelens` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-07-28

Initial release (F-019). scopelens is the fifth tool in the suite. See
[ADR-023](../../docs/adr-023-scopelens-dependencia-git.md).

### Added

- **Per-PR file budget.** Counts the *union* of everything already committed on
  the branch (`git diff <merge-base>...HEAD`) and everything staged
  (`git diff --cached`), and fails if the diff exceeds `maxFiles` (default `15`).
  The unit of the policy is the PR, not the commit: five commits of 4 files each
  add up to 20 and fail.
- **Local pre-commit gate over `git`** — no network, no token, no open PR
  required. Aborts the `git commit` before the PR even exists.
- **Exit code 2 = "could not measure"** — distinct from over-budget. Returned when
  `git` is missing from `PATH`, the cwd is not a repo, a shallow clone makes
  `merge-base` unresolvable, the base ref is not found, the config is invalid, or
  on a usage error. A missing measurement is never silently treated as a pass.
- **Config keys** (`scopelens.json`): `maxFiles` (int, default `15`; negative is
  an error), `base` (base ref, empty = auto-discover `origin/HEAD` → `main` →
  `master`), `excludeTests` (bool, default `false`), and `exclude` (glob list,
  `.gitignore` style, atomic — replaces the built-in defaults when set).
- **Flags:** `--fail`, `--max-files <int>`, `--base <ref>`, `--staged-only`,
  `--exclude-tests`, `--dir <path>`, `--no-color`, plus `init` (with `--output`)
  and `--tutorial`.
- Multi-ecosystem config chain: `scopelens.json` → `pyproject.toml
  [tool.scopelens]` → `package.json "scopelens"` → `composer.json
  extra.open-harness.scopelens` → compiled defaults.

### Notes

- `git` is invoked as an external system binary via `os/exec` with
  `--no-optional-locks` and a `context.Context` timeout, so scopelens only ever
  reads and never interferes with a concurrent `git`. This keeps ADR-002 (zero Go
  runtime dependencies) intact — see [ADR-023](../../docs/adr-023-scopelens-dependencia-git.md).
- `scopelens` does **not** expose `--format json` or `--config`: the report is a
  single fixed console format, and config resolves through the standard chain.
