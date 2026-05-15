# @open_harness/open-harness

The full **open-harness** suite in a single package. Installs all four code-quality linters at once: `linelens`, `dupelens`, `secretlens`, `testlens`.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

> **Same suite, other ecosystems**: the binaries shipped here are also published on **PyPI** ([`open-harness`](https://pypi.org/project/open-harness/)) for Python projects and on **Packagist** (`open-harness/open-harness`) for PHP/Composer projects. Pick the registry that matches your stack — the underlying tools and config formats are identical.

## Install

```bash
npm install --save-dev @open_harness/open-harness
```

npm resolves the dependency tree and pulls the right native binary for your platform for each tool (Linux x64, macOS arm64, macOS x64, Windows x64).

## What you get

| Tool | Purpose |
|---|---|
| **linelens**   | File length linter — flags files over a configurable line limit. |
| **dupelens**   | Code duplication detector (Rabin-Karp, language-agnostic). |
| **secretlens** | Secret and credential detector (AWS keys, GitHub tokens, JWT, PEM, …). |
| **testlens**   | Test coverage detector — finds source files without tests, 9 languages. |

After install, all four binaries are available via `npx`:

```bash
npx linelens   check
npx dupelens   check
npx secretlens check
npx testlens   check --lang typescript
```

## Configure everything from `package.json`

Each tool reads its config from a dedicated key inside `package.json`:

```json
{
  "name": "my-project",
  "linelens":   { "default": { "maxLines": 100 } },
  "dupelens":   { "default": { "minTokens": 50, "minLines": 5 } },
  "secretlens": { "allowlist": ["example", "placeholder"] },
  "testlens":   { "language": "typescript", "exclude": ["node_modules", "dist"] }
}
```

You can mix-and-match: keep dedicated files (`linelens.json`, etc.) for tools you want to override and use `package.json` keys for the rest. Precedence per tool: `--config <path>` > dedicated `*.json` > `package.json` key > built-in defaults.

## Run all four as a single CI / git-hook gate

```yaml
# GitHub Actions
- run: npx linelens   check --fail
- run: npx dupelens   check --fail
- run: npx secretlens check --fail
- run: npx testlens   check --fail --lang typescript
```

```yaml
# lefthook.yml
pre-commit:
  commands:
    linelens:   { run: npx linelens   check --fail --no-color }
    dupelens:   { run: npx dupelens   check --fail --no-color }
    secretlens: { run: npx secretlens check --fail --no-color }
    testlens:   { run: npx testlens   check --fail }
```

```bash
# .husky/pre-commit
npx linelens   check --fail
npx dupelens   check --fail
npx secretlens check --fail
npx testlens   check --fail
```

## Tool-level docs

Each tool ships in three ecosystems with identical behavior. Pick the registry that matches your project's stack:

| Tool | npm (Node/TS) | PyPI (Python) | Packagist (PHP) |
|---|---|---|---|
| linelens   | [@open_harness/linelens](https://www.npmjs.com/package/@open_harness/linelens)     | [open-harness-linelens](https://pypi.org/project/open-harness-linelens/)     | [open-harness/linelens](https://packagist.org/packages/open-harness/linelens)     |
| dupelens   | [@open_harness/dupelens](https://www.npmjs.com/package/@open_harness/dupelens)     | [open-harness-dupelens](https://pypi.org/project/open-harness-dupelens/)     | [open-harness/dupelens](https://packagist.org/packages/open-harness/dupelens)     |
| secretlens | [@open_harness/secretlens](https://www.npmjs.com/package/@open_harness/secretlens) | [open-harness-secretlens](https://pypi.org/project/open-harness-secretlens/) | [open-harness/secretlens](https://packagist.org/packages/open-harness/secretlens) |
| testlens   | [@open_harness/testlens](https://www.npmjs.com/package/@open_harness/testlens)     | [open-harness-testlens](https://pypi.org/project/open-harness-testlens/)     | [open-harness/testlens](https://packagist.org/packages/open-harness/testlens)     |

---

## Español

La suite completa de **open-harness** en un solo paquete. Instala los cuatro linters de calidad de código de una sola línea: `linelens`, `dupelens`, `secretlens`, `testlens`.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

### Instalación

```bash
npm install --save-dev @open_harness/open-harness
```

npm resuelve el árbol de dependencias y descarga el binario nativo correcto para tu plataforma por cada tool (Linux x64, macOS arm64, macOS x64, Windows x64).

### Qué incluye

| Tool | Propósito |
|---|---|
| **linelens**   | Linter de longitud — marca archivos que superan un límite de líneas. |
| **dupelens**   | Detector de duplicación (Rabin-Karp, agnóstico al lenguaje). |
| **secretlens** | Detector de secretos y credenciales (AWS, GitHub tokens, JWT, PEM, …). |
| **testlens**   | Detector de cobertura — encuentra archivos sin test, 9 lenguajes. |

Después del install, los cuatro binarios están disponibles via `npx`.

### Configurá todo desde `package.json`

Cada tool lee su config desde una key dedicada dentro de `package.json` (ver ejemplo arriba). Podés combinar: archivos dedicados (`linelens.json`, etc.) para los tools que quieras y keys en `package.json` para el resto. Precedencia por tool: `--config <path>` > archivo dedicado `*.json` > key en `package.json` > defaults del binario.

### Integraciones

Sirve con Husky, lefthook o GitHub Actions con los snippets de la sección en inglés.

### Docs por tool

Cada tool tiene su propio README bilingüe con flags completos, formato de configuración y ejemplos de integración. Los links arriba apuntan a cada página de npm.

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
