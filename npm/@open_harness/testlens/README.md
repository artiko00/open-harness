# @open_harness/testlens

Test coverage detector. Finds source files that don't have a corresponding test file, across 9 languages. Single native binary, zero runtime dependencies.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

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
npx testlens check --fail                # exit 1 si hay archivos sin test
npx testlens init                        # genera la config por defecto
npx testlens version                     # imprime la versión
```

### Lenguajes soportados

Go, TypeScript, JavaScript, Python, Ruby, Rust, Java, Kotlin, C#. Ver la tabla arriba para las extensiones y patrones de naming de test que reconoce cada uno.

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
