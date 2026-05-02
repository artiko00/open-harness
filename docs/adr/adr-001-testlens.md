# ADR-001: testlens - Test Coverage Detection

**Date:** 2026-05-02
**Status:** Proposed
**Deciders:** Jassen Castillo (artiko00)

---

## Context

open-harness needs a tool to detect source files missing corresponding test files. This is a cross-language problem — the tool must handle multiple file extensions and naming conventions.

## Decision

Create `testlens` as a new tool in `tools/testlens/`.

### Multi-language support

The tool maps source extensions to test extensions per language:

| Language | Source extensions | Test extensions |
|-----------|-------------------|-----------------|
| Go | `.go` | `_test.go` |
| TypeScript/JS | `.ts`, `.tsx`, `.js`, `.jsx` | `.test.ts`, `.test.tsx`, `.test.js`, `.test.jsx`, `.spec.ts`, `.spec.tsx`, `.spec.js`, `.spec.jsx` |
| Python | `.py` | `_test.py`, `test_.py` |
| Ruby | `.rb` | `_spec.rb`, `_test.rb` |
| Rust | `.rs` | `mod.rs` or `*_test.rs` |
| Java/Kotlin | `.java`, `.kt` | `*Test.java`, `*Test.kt` |
| C# | `.cs` | `*Tests.cs` |

### File matching rules

- Test file is adjacent to source (same directory) OR in `test/`, `tests/`, `__tests__/` subdirectory
- Matching by base name: `src/foo.ts` → `src/foo.test.ts`
- Skips: `node_modules/`, `vendor/`, `.git/`, `dist/`, `build/`, `*.generated.*`

### Output

Default: human-readable list
```
src/utils/parser.ts - no test found
src/api/client.ts - has test: src/api/client.test.ts
src/config.ts - no test found
```

Exit code: 0 if all covered, 1 if violations found (for CI integration).

## Consequences

### Positive
- Generic enough for most codebases
- Simple regex-based matching, no AST needed
- Useful immediately for improving test coverage

### Negative
- Can't detect if test actually covers the source (quality vs quantity)
- Naming convention assumptions may not match all teams

## Alternatives considered

- AST-based detection (rejected: complexity, language-specific)
- Config-driven extension mapping (rejected: too flexible, adds friction)
- Inline annotations in source (rejected: requires modifying code)