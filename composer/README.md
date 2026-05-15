# composer/

Packagist (Composer / PHP) distribution layer for open-harness. 5 packages publish wrappers that download the same native binaries used by `npm/@open_harness/` and `pypi/`.

| Package | Packagist name | Version |
|---|---|---|
| `linelens/`     | `open-harness/linelens`     | 0.1.3 |
| `dupelens/`     | `open-harness/dupelens`     | 0.1.3 |
| `secretlens/`   | `open-harness/secretlens`   | 0.1.2 |
| `testlens/`     | `open-harness/testlens`     | 0.1.2 |
| `open-harness/` | `open-harness/open-harness` | 0.1.0 (metapackage) |

## Distribution model

Composer does not have a `manylinux`-style native wheel mechanism, so each per-tool package follows a different recipe than npm/PyPI:

1. The Composer package itself is **pure PHP** (≈ 200 lines: a `Platform` detector, a `BinaryDownloader` post-install script, an `Archive` extractor, and a `bin/<tool>` shim that delegates to the native binary).
2. After `composer install`, the post-install hook downloads the right tarball/zip for the host platform from the matching **GitHub Release** of `artiko00/open-harness`.
3. The hook verifies the SHA256 of the asset against the release's `checksums.txt` before extracting the binary into `vendor/bin/`.
4. The `bin/<tool>` shim that Composer registers in `vendor/bin/` finds the native binary and `passthru()`'s into it.

This means **F-013 depends on a published GitHub Release** with the asset names produced by F-011 (or any equivalent manual release flow): `open-harness-<tool>-<os>-<arch>.tar.gz` (Linux/macOS) and `.zip` (Windows), plus `checksums.txt`.

## Layout per tool

```
composer/linelens/
├── composer.json                 # name: open-harness/linelens
├── README.md
├── src/
│   ├── Platform.php              # OS/arch detection
│   ├── BinaryDownloader.php      # post-install hook
│   └── Archive.php               # tar.gz / zip extractor
└── bin/
    └── linelens                  # PHP shim → native binary
```

The `open-harness/open-harness` directory holds a `metapackage` `composer.json` with hard `==` requires on the four per-tool packages, so `composer require --dev open-harness/open-harness` pulls all four with one command.

## Publishing to Packagist

Composer / Packagist expects a Git repository **per package**. Because we live in a monorepo, each `composer/<tool>/` subdirectory needs a "split" mirror repository (e.g. `artiko00/open-harness-linelens-php`) that contains only that subdirectory's files. The standard tool for this is [`splitsh-lite`](https://github.com/splitsh/lite). Once the split repos exist, you submit each one once at https://packagist.org/packages/submit and Packagist auto-syncs on each tag.

For the first manual round you can also do a one-time split via:

```bash
splitsh-lite --prefix=composer/linelens \
  --target=refs/heads/main \
  --origin=refs/heads/main
git push split-linelens main
```

…repeated per tool. Add a GitHub webhook from each split repo to Packagist so future tags propagate automatically.

## Manual / first time setup

1. Create accounts on https://packagist.org and https://github.com.
2. Reserve the `open-harness` vendor namespace by submitting `linelens` first (the first submission claims the vendor).
3. For each tool, push the split repo and submit it on Packagist.
4. Tag a version (e.g. `v0.1.3`) on the monorepo; the splitsh-lite hook propagates the tag to each split repo; Packagist picks it up via webhook.

## Roadmap (F-013 in `.agent/feature-list.json`)

- [ ] `extra.open-harness.<tool>` in `composer.json` as an alternate config source (mirrors `package.json` keys on npm). Requires extending each tool's Go `config_pkg.go`.
- [ ] Smoke test in a fresh Laravel/Symfony scaffold: `composer require --dev open-harness/open-harness && vendor/bin/linelens version`.
- [ ] Subtree split script in `scripts/split-composer.sh`.
- [ ] ADR-017 documenting why post-install download from GitHub Releases instead of embedding binaries.
