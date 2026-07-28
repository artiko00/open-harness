# Changelog

All notable changes to `secretlens` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.1] - 2026-07-28

### Added

- **`--tutorial`** flag: prints a static configuration guide to stdout — every
  config key with its default and an example, plus the tool's flags. Exit `0`;
  `--no-color` strips ANSI. Onboarding release (F-020).

## [0.3.0] - 2026-07-27

Part of the `fix-audit-findings` release (adversarial audit F-018). See
[docs/UPGRADING.md](../../docs/UPGRADING.md) and
[ADR-021](../../docs/adr-021-secretlens-entropia.md).

### Added

- **Much higher recall — from ~25% to ~100%** on the audit fixture. `secretlens`
  now detects unquoted `KEY=value` assignments (quotes are no longer required),
  and decodes UTF-16 files before scanning.
- **Provider-prefix patterns.** New built-in rules for Stripe (`sk_live_`), Slack
  (`xoxb-`/`xoxp-…` and `hooks.slack.com/services/…` webhooks), Google
  (`AIza…`), OpenAI (`sk-proj-…`), GitLab (`glpat-…`), npm (`npm_…`), SendGrid
  (`SG.…`), and credentials embedded in connection URIs
  (`postgres://user:pass@…`, `mysql://`, `mongodb+srv://`, `redis://`, `amqp://`).
- **Shannon-entropy filter** for the generic `KEY=VALUE` rules (`minEntropy`,
  default `3.0` bits/char). Applied only to generic rules via a per-pattern
  `entropyGate`; strong-prefix rules (AWS, GitHub, PEM, JWT, providers) are never
  gated, preserving their recall.
- **`disableDefaultPatterns`** config key (default `false`) to opt into
  "only my custom patterns".
- **`entropyGate`** field on custom `patterns` entries.
- Unified CLI contract: `--format console|json`, `--config <path>`, `--no-color`,
  and `--output <file>` on `init`.

### Changed

- **Allowlist is matched against the detected value, not the whole line.** A
  placeholder *value* (`KEY=example_value`) is still suppressed, but an unrelated
  mention elsewhere on the line (`AWS_KEY="AKIA..." # see example above`) is now
  correctly reported.
- **Strict config loading:** an unknown config key now prints a warning to stderr
  and continues, instead of being silently dropped.

### Fixed

- Closed the recall gap where 0.2.x silently missed most real secrets (no entropy
  analysis, quotes required, no provider prefixes).

### BREAKING

- **Custom `patterns` are now additive.** In 0.2.x, defining any custom `patterns`
  **silently disabled all built-in rules** (the audit flagged this as a security
  footgun). In 0.3.0 custom patterns are **added on top of** the built-ins. If you
  truly need only your own patterns, set `"disableDefaultPatterns": true`.
- **Higher recall surfaces new findings.** Real secrets that 0.2.x missed will now
  be reported, so `secretlens check --fail` may start failing on a repo that
  "passed" before. Run once without `--fail`, triage, rotate/remove real secrets,
  and add genuine false positives to `allowlist`/`exclude` before enforcing.
