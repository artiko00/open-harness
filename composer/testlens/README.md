# open-harness/testlens (Composer)

File length linter for any language. Reports files that exceed a configured line limit. Single native binary, zero runtime dependencies.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

## Install

```bash
composer require --dev open-harness/testlens
```

On install, a Composer post-install hook downloads the native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) from GitHub Releases and verifies its SHA256 checksum.

## Usage

```bash
vendor/bin/testlens check               # scan current directory
vendor/bin/testlens check --fail        # exit 1 on violations (CI / git hooks)
vendor/bin/testlens check --dir ./src   # scan a specific directory
vendor/bin/testlens check --max 200     # override the line limit
vendor/bin/testlens check --no-color    # plain output for logs
vendor/bin/testlens init                # generate a default testlens.json
vendor/bin/testlens version             # print version
```

## Configuration

Place a `testlens.json` at the repo root:

```json
{
  "default": { "maxLines": 100 },
  "rules": [
    { "pattern": "**/*_test.go",     "maxLines": 300 },
    { "pattern": "**/*.spec.*",      "maxLines": 300 },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": ["node_modules", "vendor", ".git", "dist"]
}
```

Pattern semantics follow `.gitignore` style. The first matching `rules` entry wins; if no rule matches, `default.maxLines` applies.

### Alternative: configure inside `composer.json`

If you prefer not to keep a separate `testlens.json`, add a `testlens` key in your `package.json` with the same shape:

```json
{
  "name": "my-project",
  "testlens": {
    "default": { "maxLines": 100 },
    "rules": [{ "pattern": "**/*_test.go", "maxLines": 300 }],
    "exclude": ["node_modules", "dist"]
  }
}
```

Precedence: `--config <path>` > `testlens.json` > `package.json` key > built-in defaults. CLI flags (`--max`, `--no-color`, etc.) always win.

## Integrations

```bash
# Husky pre-commit
vendor/bin/testlens check --fail
```

```yaml
# lefthook.yml
pre-commit:
  commands:
    testlens:
      run: vendor/bin/testlens check --fail --no-color
```

```yaml
# GitHub Actions
- name: Run testlens
  run: npx @open_harness/testlens check --fail
```

## Why a line limit?

Large files concentrate too many responsibilities and become hard to read, test, and refactor. A soft cap (e.g. 100 lines, with exceptions for tests) keeps modules focused and forces responsibility-split decisions early — when they are cheap.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | No violations (or `--fail` not passed) |
| `1` | Violations found and `--fail` was passed, or config error |

---

## Español

Linter de longitud de archivos, agnóstico al lenguaje. Reporta los archivos que superan un límite de líneas configurable. Un solo binario nativo, cero dependencias en tiempo de ejecución.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

### Instalación

```bash
composer require --dev open-harness/testlens
```

Tras instalar, un hook post-install de Composer descarga el binario nativo para tu plataforma (Linux x64, macOS arm64, macOS x64, Windows x64) desde GitHub Releases y verifica su checksum SHA256.

### Uso

```bash
vendor/bin/testlens check               # escanea el directorio actual
vendor/bin/testlens check --fail        # exit 1 si hay violaciones (CI / git hooks)
vendor/bin/testlens check --dir ./src   # escanea un directorio específico
vendor/bin/testlens check --max 200     # sobrescribe el límite de líneas
vendor/bin/testlens check --no-color    # salida sin colores
vendor/bin/testlens init                # genera un testlens.json por defecto
vendor/bin/testlens version             # imprime la versión
```

### Configuración

Colocá un `testlens.json` en la raíz del repo. Ver ejemplo arriba. La semántica de patrones sigue el estilo `.gitignore`. La primera regla coincidente en `rules` gana; si ninguna coincide, aplica `default.maxLines`.

#### Alternativa: configurar dentro de `composer.json`

Si preferís no tener un `testlens.json` separado, agregá una key `testlens` en tu `package.json` con la misma forma del archivo dedicado. Precedencia: `--config <path>` > `testlens.json` > key en `package.json` > defaults del binario. Los flags CLI (`--max`, `--no-color`, etc.) siempre ganan.

### Integraciones

Mismos snippets que arriba — sirven con Husky (`.husky/pre-commit`), lefthook (`lefthook.yml`) o GitHub Actions.

### Por qué un límite de líneas

Los archivos grandes concentran demasiadas responsabilidades y son difíciles de leer, testear y refactorizar. Un tope blando (por ejemplo 100 líneas, con excepciones para tests) mantiene los módulos enfocados y obliga a tomar decisiones de partición temprano — cuando son baratas.

### Códigos de salida

| Código | Significado |
|---|---|
| `0` | Sin violaciones (o no se pasó `--fail`) |
| `1` | Hubo violaciones con `--fail`, o error de configuración |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
