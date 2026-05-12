# @open_harness/testlens

Test coverage detector. Finds source files in your project that don't have a corresponding test file, across 9 languages. Single native binary, zero runtime dependencies, works in any ecosystem.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo.

## Install

```bash
npm install --save-dev @open_harness/testlens
```

The right native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) is downloaded automatically via `optionalDependencies`.

## Usage

```bash
npx testlens check                       # auto-detect language and scan
npx testlens check --lang typescript     # force a specific language
npx testlens check --dir ./src           # scan a specific directory
npx testlens check --fail                # exit 1 if files without tests are found
npx testlens init                        # generate a default config
npx testlens version                     # print version
```

## Supported languages

| Language | Source extensions | Test patterns |
|---|---|---|
| Go | `.go` | `*_test.go` |
| TypeScript | `.ts`, `.tsx` | `*.test.ts`, `*.spec.ts`, `test_*.ts` |
| JavaScript | `.js`, `.jsx` | `*.test.js`, `*.spec.js`, `test_*.js` |
| Python | `.py` | `*_test.py`, `test_*.py` |
| Ruby | `.rb` | `*_spec.rb`, `*_test.rb` |
| Rust | `.rs` | `*_test.rs` |
| Java | `.java` | `*Test.java` |
| Kotlin | `.kt`, `.kts` | `*Test.kt` |
| C# | `.cs` | `*Tests.cs` |

## Husky / lefthook / CI integration

```bash
# .husky/pre-commit
npx testlens check --fail
```

```yaml
# .github/workflows/quality.yml
- name: Detect source files without tests
  run: npx @open_harness/testlens check --fail --lang typescript --dir src/
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | All source files have tests (or `--fail` not passed) |
| `1` | Files without tests found and `--fail` was passed, or config error |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
