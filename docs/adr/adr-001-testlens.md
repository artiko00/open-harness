# ADR-001: testlens - Test Coverage Detection

**Date:** 2026-05-02
**Status:** Accepted
**Deciders:** Jassen Castillo (artiko00)

---

## Context

open-harness needs a tool to detect source files missing corresponding test files. This is a cross-language problem — the tool must handle multiple file extensions and naming conventions.

## Decision

Create `testlens` as a new tool in `tools/testlens/`.

### Multi-language support

The tool maps source extensions to test extensions per language:

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

### File matching rules

- Test file is adjacent to source (same directory)
- Matching by base name: `src/foo.ts` → `src/foo.test.ts`
- Skips: `node_modules/`, `vendor/`, `.git/`, `dist/`, `build/`

### Output

```
  src/utils/parser.ts - no test found
  src/api/client.ts - has test: src/api/client.test.ts

2 file(s) without tests
```

Exit code: 0 if all covered, 1 if violations found (for CI integration).

### Implementation

```
tools/testlens/
├── main.go         # Entry point, commands: check, init, version
├── check.go        # CLI flag parsing, config struct
├── coverage.go     # Core scanning logic, filepath.Walk
├── detect.go       # Language auto-detection from file extensions
├── language.go     # Language mapping definitions
├── matcher.go      # Test candidate generation
├── scanner.go      # File discovery utilities
├── reporter.go     # Output formatting
├── file.go         # fileExists helper
└── scanner_test.go # 9 unit tests
```

### Usage

```bash
# Auto-detect language
testlens check

# Specific language
testlens check --lang go
testlens check --lang typescript --dir ./src

# CI mode
testlens check --fail
```

## Consequences

### Positive
- Generic enough for most codebases
- Simple regex-based matching, no AST needed
- Useful immediately for improving test coverage
- All files under 100 lines (linelens compliant)

### Negative
- Can't detect if test actually covers the source (quality vs quantity)
- Naming convention assumptions may not match all teams

## Status history

- 2026-05-02: Proposed
- 2026-05-02: Accepted and implemented (epic/testlens merged)