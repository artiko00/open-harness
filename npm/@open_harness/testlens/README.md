# @open_harness/testlens

Test coverage detector. Finds source files that don't have a corresponding test file, across 9 languages. Single native binary, zero runtime dependencies.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

> **Same tool, other ecosystems**: also available on **PyPI** ([`open-harness-testlens`](https://pypi.org/project/open-harness-testlens/)) and on **Packagist** (`open-harness/testlens`). Identical binary, identical config; pick the registry that matches your stack.

## Install

```bash
npm install --save-dev @open_harness/testlens
```

The right native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) is fetched automatically via `optionalDependencies`.

## Usage

```bash
npx testlens check                       # auto-detect language and scan
npx testlens check --lang typescript     # force a specific language
npx testlens check --dir ./src           # scan a specific directory
npx testlens check --config my.json      # use a specific config file
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

## Configuration

Place a `testlens.json` at the repo root:

```json
{
  "language": "auto",
  "exclude": ["node_modules", ".git", "vendor", "dist", "build", "testdata"]
}
```

- `language` — `auto` (default), `go`, `typescript`, `javascript`, `python`, `ruby`, `rust`, `java`, `kotlin`, `csharp`.
- `exclude` — directories to skip during the scan.

### Alternative: configure inside `package.json`

If you prefer not to keep a separate `testlens.json`, add a `testlens` key in your `package.json` with the same shape:

```json
{
  "name": "my-project",
  "testlens": {
    "language": "typescript",
    "exclude": ["node_modules", "dist", "fixtures"]
  }
}
```

Precedence: `--config <path>` > `testlens.json` > `package.json` key > built-in defaults. CLI flags win **when passed explicitly** (`fs.Visit`); if you don't pass `--lang`, the value from the config is used.


## Supported test layouts

testlens recognises **colocated**, **adjacent** subdirs, **mirror trees**, and **centralised** test roots that mirror the source path:

| Pattern | Example |
|---|---|
| Colocated | `src/foo.ts` ↔ `src/foo.test.ts` |
| Adjacent subdir | `src/foo.ts` ↔ `src/__tests__/foo.test.ts` |
| **Centralised tests root** (F-016) | `src/services/foo.ts` ↔ `src/__tests__/services/foo.test.ts` |
| Mirror (Python) | `src/auth/user.py` ↔ `tests/auth/test_user.py` |
| Mirror (Maven) | `src/main/java/com/X/Bar.java` ↔ `src/test/java/com/X/BarTest.java` |

Per language defaults:

testlens looks for tests in these locations (per language):

| Language | Colocated | Subdir | Mirror |
|---|---|---|---|
| Go | `foo.go` ↔ `foo_test.go` | — | — |
| TypeScript / JavaScript | `foo.ts` ↔ `foo.test.ts` | `__tests__/foo.test.ts`, `tests/foo.spec.ts` | — |
| Python | `foo.py` ↔ `test_foo.py` | `tests/test_foo.py` | `src/foo.py` ↔ `tests/foo.py` |
| Ruby | `foo.rb` ↔ `foo_spec.rb` | `spec/foo_spec.rb` | `lib/foo.rb` ↔ `spec/foo.rb` |
| Java / Kotlin | — | — | `src/main/java/.../Bar.java` ↔ `src/test/java/.../BarTest.java` |
| Rust | `foo.rs` ↔ `foo_test.rs` | `tests/foo_test.rs` | — |

## testlens vs `vitest --coverage`, `jest`, `pytest-cov`

These have **different goals** — use both:

| | testlens | Real coverage tools |
|---|---|---|
| Speed | milliseconds (file walk) | seconds to minutes (compile + execute + instrument) |
| Detects files without tests | ✅ | partial (only files actually loaded by a test) |
| Detects untested lines | ❌ | ✅ |
| Detects untested branches | ❌ | ✅ |
| Works as pre-commit hook | ✅ | too slow |

testlens is a fast **smoke test of test-file existence**, intended as a pre-flight gate. For real functional coverage metrics, run Vitest, Jest, pytest-cov, etc. as a separate (slower) CI step.

## Why this exists

Coverage tools tell you which **lines** are tested. testlens tells you which **files** have no test at all — a different and complementary check. It surfaces orphan modules early, when adding a first test is cheapest. Combine both for a complete picture.

## Integrations

```bash
# Husky pre-commit
npx testlens check --fail
```

```yaml
# GitHub Actions
- name: Detect source files without tests
  run: npx @open_harness/testlens check --fail --lang typescript --dir src/
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | All source files have tests (or `--fail` not passed) |
| `1` | Files without tests found and `--fail` was passed, or config error |

---

## Español

Detector de cobertura de tests. Encuentra archivos fuente que no tienen un archivo de test correspondiente, en 9 lenguajes. Un solo binario nativo, cero dependencias.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

### Instalación

```bash
npm install --save-dev @open_harness/testlens
```

El binario para tu plataforma se descarga automáticamente via `optionalDependencies`.

### Uso

```bash
npx testlens check                       # autodetecta lenguaje y escanea
npx testlens check --lang typescript     # fuerza un lenguaje específico
npx testlens check --dir ./src           # escanea un directorio específico
npx testlens check --config my.json      # usa un archivo de config específico
npx testlens check --fail                # exit 1 si hay archivos sin test
npx testlens init                        # genera la config por defecto
npx testlens version                     # imprime la versión
```

### Lenguajes soportados

Go, TypeScript, JavaScript, Python, Ruby, Rust, Java, Kotlin, C#. Ver la tabla arriba para las extensiones y patrones de naming de test que reconoce cada uno.

### Configuración

Colocá un `testlens.json` en la raíz del repo (ver ejemplo arriba).

- `language` — `auto` (default), `go`, `typescript`, `javascript`, `python`, `ruby`, `rust`, `java`, `kotlin`, `csharp`.
- `exclude` — directorios a ignorar durante el escaneo.

#### Alternativa: configurar dentro de `package.json`

Si preferís no tener un `testlens.json` separado, agregá una key `testlens` en tu `package.json` con la misma forma. Precedencia: `--config <path>` > `testlens.json` > key en `package.json` > defaults. Los flags CLI ganan **solo cuando se pasan explícitamente**; si omitís `--lang`, se usa el valor de la config.


### Layouts de test soportados

testlens busca tests en estas ubicaciones (por lenguaje):

| Lenguaje | Co-ubicado | Subdir | Mirror |
|---|---|---|---|
| Go | `foo.go` ↔ `foo_test.go` | — | — |
| TypeScript / JavaScript | `foo.ts` ↔ `foo.test.ts` | `__tests__/foo.test.ts`, `tests/foo.spec.ts` | — |
| Python | `foo.py` ↔ `test_foo.py` | `tests/test_foo.py` | `src/foo.py` ↔ `tests/foo.py` |
| Ruby | `foo.rb` ↔ `foo_spec.rb` | `spec/foo_spec.rb` | `lib/foo.rb` ↔ `spec/foo.rb` |
| Java / Kotlin | — | — | `src/main/java/.../Bar.java` ↔ `src/test/java/.../BarTest.java` |
| Rust | `foo.rs` ↔ `foo_test.rs` | `tests/foo_test.rs` | — |

### testlens vs `vitest --coverage`, `jest`, `pytest-cov`

Tienen **finalidades distintas** — usá los dos:

testlens es un **smoke test ultra-rápido de existencia de archivos de test**, pensado como gate de pre-commit. Para métricas funcionales de cobertura (líneas/branches ejercidas) corré Vitest, Jest, pytest-cov, etc. como step separado de CI. testlens corre en milisegundos; un coverage tool real toma segundos a minutos.

### Por qué existe

Las herramientas tradicionales de cobertura te dicen qué **líneas** están testeadas. testlens te dice qué **archivos** no tienen ningún test — una verificación distinta y complementaria. Detecta módulos huérfanos temprano, cuando agregar el primer test es más barato. Combinalos para tener una vista completa.

### Integraciones

Sirve con Husky, lefthook o GitHub Actions usando los snippets de la sección en inglés.

### Códigos de salida

| Código | Significado |
|---|---|
| `0` | Todos los archivos fuente tienen tests (o no se pasó `--fail`) |
| `1` | Hay archivos sin test con `--fail`, o error de configuración |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
