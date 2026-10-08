# Changelog

All notable changes to `dupelens` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.1] - 2026-10-08

### Fixed

- **npm: `npx dupelens` works after a local install.** The four platform packages
  (`@open_harness/dupelens-<os>-<cpu>`) declared `"bin": { "dupelens": ... }`, the same
  name as the wrapper's command. With npm 10, the optional dependencies for the
  other platforms fail on `os`/`cpu` and, when npm rolls them back, it deletes
  `node_modules/.bin/dupelens` — the wrapper's link. `npx dupelens` then looked the
  package up in the registry and failed with a 404. The platform packages no
  longer declare `bin`; the wrapper never used it (it resolves the binary through
  `require.resolve`). No change to the binary or its behavior (F-025).

## [0.5.0] - 2026-10-08

Bounded memory on large repositories (F-024). Reported by users whose projects made
`dupelens` take all the RAM: reproduced on a 226 MB `site-packages`, where 0.4.1 grew
from 1.5 GB to 22 GB in two minutes until `systemd-oomd` killed the whole session.
0.5.0 scans the same tree in 388 MB. See
[ADR-024](../../docs/adr-024-dupelens-presupuesto-de-memoria.md) and
[UPGRADING](../../docs/UPGRADING.md#upgrading-to-dupelens-050).

### Added

- **Memory budget.** `--max-memory`, the `DUPELENS_MAX_MEMORY` environment variable
  and the `maxMemory` config key (in that order of precedence) accept `"512MiB"`,
  `"2GiB"` or `"40%"` of the available memory. The default, `auto`, is the lower of
  1 GiB and 25% of the available memory — the system's or, inside a cgroup with a
  limit, its remaining headroom (Linux `/proc/meminfo` and cgroup v1/v2, macOS
  `hw.memsize`, Windows `GlobalMemoryStatusEx`; stdlib only).
- **Exit code 2** when the scan exceeds the budget: no partial report, with or
  without `--fail`, and a message with the budget, its source and how to raise it.
  The GC soft limit is set to 90% of the budget and checkpoints read
  `runtime/metrics`, so the process stops cleanly instead of being OOM-killed.
- `--verbose` prints the memory budget in force and where it came from.

### Changed

- **No more quadratic axis.** 0.4.x enumerated every pair inside each hash bucket and
  allocated one match per window and pair before merging: a block of L windows
  copied k times cost L·k(k−1)/2 objects. Each occurrence is now paired only with
  the first one of its class (k−1 seeds), seeds are extended to the maximal clone of
  their diagonal and merged per pair. Measured: 32 copies of a 0.5 MB package went
  from >3 GB (OOM) to 69 MB; Ansible's generated `fortinet/fortios` (20 MB) from
  >3 GB to 45 MB.
- **Compact representation.** Interned tokens (id + line, 8 bytes per token), the
  normalized view derived through a table instead of a copy, sorted (hash, position)
  candidates instead of a map, and a two-bitset prefilter that never materializes
  unique windows. The Python standard library (11.3 MB) went from 258 MB to 25 MB.
- **61-bit rolling hash** (modulo 2^61−1) instead of ~30 bits, which produced millions
  of spurious collisions on large trees. Literal verification is unchanged.
- **BREAKING (counts):** with three or more copies of a block, each copy is reported
  against the first one (`a-b`, `a-c`; `b-c` is implied). The result of `--fail`
  does not change.
- **BREAKING (counts):** `tokens` is the real span of the duplicated block. 0.4.x
  added one token per merged pair of windows, even across different alignments, and
  reported blocks above `minTokens` that never were (31 real tokens reported as 71).

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
