# open-harness-scopelens

File length linter for any language. Reports files that exceed a configured line limit. Single native binary, zero runtime dependencies.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

> **Same tool, other ecosystems**: also available on **npm** ([`@open_harness/scopelens`](https://www.npmjs.com/package/@open_harness/scopelens)) and on **Packagist** (`open-harness/scopelens`). Identical binary, identical config; pick the registry that matches your stack.

## Install

```bash
pip install open-harness-scopelens
```

pip picks the right native wheel for your platform automatically (Linux x86_64, macOS arm64, macOS x86_64, Windows x86_64). Each wheel embeds the Go binary — no runtime deps.

## Usage

```bash
scopelens check               # scan current directory
scopelens check --fail        # exit 1 on violations (CI / git hooks)
scopelens check --dir ./src   # scan a specific directory
scopelens check --max 200     # override the line limit
scopelens check --no-color    # plain output for logs
scopelens init                # generate a default scopelens.json
scopelens version             # print version
```

## Configuration

Place a `scopelens.json` at the repo root:

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

### Alternative: configure inside `pyproject.toml` or the dedicated `scopelens.json`

If you prefer not to keep a separate `scopelens.json`, add a `scopelens` key in your `package.json` with the same shape:

```json
{
  "name": "my-project",
  "scopelens": {
    "default": { "maxLines": 100 },
    "rules": [{ "pattern": "**/*_test.go", "maxLines": 300 }],
    "exclude": ["node_modules", "dist"]
  }
}
```

Precedence: `--config <path>` > `scopelens.json` > `package.json` key > built-in defaults. CLI flags (`--max`, `--no-color`, etc.) always win.

## Integrations

```bash
# Husky pre-commit
scopelens check --fail
```

```yaml
# lefthook.yml
pre-commit:
  commands:
    scopelens:
      run: scopelens check --fail --no-color
```

```yaml
# GitHub Actions
- name: Run scopelens
  run: npx @open_harness/scopelens check --fail
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
pip install open-harness-scopelens
```

pip elige automáticamente la wheel nativa correcta para tu plataforma (Linux x86_64, macOS arm64, macOS x86_64, Windows x86_64). Cada wheel embebe el binario Go — sin deps en runtime.

### Uso

```bash
scopelens check               # escanea el directorio actual
scopelens check --fail        # exit 1 si hay violaciones (CI / git hooks)
scopelens check --dir ./src   # escanea un directorio específico
scopelens check --max 200     # sobrescribe el límite de líneas
scopelens check --no-color    # salida sin colores
scopelens init                # genera un scopelens.json por defecto
scopelens version             # imprime la versión
```

### Configuración

Colocá un `scopelens.json` en la raíz del repo. Ver ejemplo arriba. La semántica de patrones sigue el estilo `.gitignore`. La primera regla coincidente en `rules` gana; si ninguna coincide, aplica `default.maxLines`.

#### Alternativa: configurar dentro de `pyproject.toml` o `scopelens.json`

Si preferís no tener un `scopelens.json` separado, agregá una key `scopelens` en tu `package.json` con la misma forma del archivo dedicado. Precedencia: `--config <path>` > `scopelens.json` > key en `package.json` > defaults del binario. Los flags CLI (`--max`, `--no-color`, etc.) siempre ganan.

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
