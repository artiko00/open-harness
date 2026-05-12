# @open_harness/linelens

File length linter for any language. Reports files that exceed a configured line limit. Single native binary, zero runtime dependencies.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

## Install

```bash
npm install --save-dev @open_harness/linelens
```

The right native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) is fetched automatically via `optionalDependencies`.

## Usage

```bash
npx linelens check               # scan current directory
npx linelens check --fail        # exit 1 on violations (CI / git hooks)
npx linelens check --dir ./src   # scan a specific directory
npx linelens check --max 200     # override the line limit
npx linelens check --no-color    # plain output for logs
npx linelens init                # generate a default linelens.json
npx linelens version             # print version
```

## Configuration

Place a `linelens.json` at the repo root:

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

### Alternative: configure inside `package.json`

If you prefer not to keep a separate `linelens.json`, add a `linelens` key in your `package.json` with the same shape:

```json
{
  "name": "my-project",
  "linelens": {
    "default": { "maxLines": 100 },
    "rules": [{ "pattern": "**/*_test.go", "maxLines": 300 }],
    "exclude": ["node_modules", "dist"]
  }
}
```

Precedence: `--config <path>` > `linelens.json` > `package.json` key > built-in defaults. CLI flags (`--max`, `--no-color`, etc.) always win.

## Integrations

```bash
# Husky pre-commit
npx linelens check --fail
```

```yaml
# lefthook.yml
pre-commit:
  commands:
    linelens:
      run: npx linelens check --fail --no-color
```

```yaml
# GitHub Actions
- name: Run linelens
  run: npx @open_harness/linelens check --fail
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
npm install --save-dev @open_harness/linelens
```

El binario nativo correcto para tu plataforma (Linux x64, macOS arm64, macOS x64, Windows x64) se descarga automáticamente via `optionalDependencies`.

### Uso

```bash
npx linelens check               # escanea el directorio actual
npx linelens check --fail        # exit 1 si hay violaciones (CI / git hooks)
npx linelens check --dir ./src   # escanea un directorio específico
npx linelens check --max 200     # sobrescribe el límite de líneas
npx linelens check --no-color    # salida sin colores
npx linelens init                # genera un linelens.json por defecto
npx linelens version             # imprime la versión
```

### Configuración

Colocá un `linelens.json` en la raíz del repo. Ver ejemplo arriba. La semántica de patrones sigue el estilo `.gitignore`. La primera regla coincidente en `rules` gana; si ninguna coincide, aplica `default.maxLines`.

#### Alternativa: configurar dentro de `package.json`

Si preferís no tener un `linelens.json` separado, agregá una key `linelens` en tu `package.json` con la misma forma del archivo dedicado. Precedencia: `--config <path>` > `linelens.json` > key en `package.json` > defaults del binario. Los flags CLI (`--max`, `--no-color`, etc.) siempre ganan.

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
