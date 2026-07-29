# @open_harness/open-harness

The full **open-harness** suite in a single package. Installs all five code-quality tools at once: `linelens`, `dupelens`, `secretlens`, `testlens`, `scopelens`.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

> **Other ecosystems**: the individual tools are also published on **PyPI** as `open-harness-<tool>` (e.g. [`open-harness-linelens`](https://pypi.org/project/open-harness-linelens/), [`open-harness-scopelens`](https://pypi.org/project/open-harness-scopelens/)) for Python projects. This all-in-one meta-package is **npm-only** — on PyPI, install the per-tool packages you need. Packagist (PHP/Composer) is planned. The underlying binaries and config formats are identical across registries.

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
| **testlens**   | Test coverage detector — finds source files without tests, 10 languages. |
| **scopelens**  | Per-PR file- and line-budget gate — counts the branch-vs-base diff locally, at `pre-commit`. |

After install, all five binaries are available via `npx`:

```bash
npx linelens   check
npx dupelens   check
npx secretlens check
npx testlens   check --lang typescript
npx scopelens  check --max-files 15
```

Any tool prints a configuration guide with `--tutorial` (e.g. `npx secretlens --tutorial`), and `npx open-harness init` creates every tool's config file at your project root in one shot.

## Configure everything from `package.json`

Each tool reads its config from a dedicated key inside `package.json`:

```json
{
  "name": "my-project",
  "linelens":   { "default": { "maxLines": 100 } },
  "dupelens":   { "default": { "minTokens": 50, "minLines": 5 } },
  "secretlens": { "allowlist": ["example", "placeholder"] },
  "testlens":   { "language": "typescript", "exclude": ["node_modules", "dist"] },
  "scopelens":  { "maxFiles": 15 }
}
```

You can mix-and-match: keep dedicated files (`linelens.json`, etc.) for tools you want to override and use `package.json` keys for the rest. Precedence per tool: `--config <path>` > dedicated `*.json` > `package.json` key > built-in defaults. See [docs/CONFIGURATION.md](https://github.com/artiko00/open-harness/blob/main/docs/CONFIGURATION.md).

## Run them as a single CI / git-hook gate

```yaml
# GitHub Actions
- run: npx linelens   check --fail
- run: npx dupelens   check --fail
- run: npx secretlens check --fail
- run: npx testlens   check --fail --lang typescript
- run: npx scopelens  check --fail
```

```yaml
# lefthook.yml
pre-commit:
  commands:
    linelens:   { run: npx linelens   check --fail --no-color }
    dupelens:   { run: npx dupelens   check --fail --no-color }
    secretlens: { run: npx secretlens check --fail --no-color }
    testlens:   { run: npx testlens   check --fail }
    scopelens:  { run: npx scopelens  check --fail --no-color }
```

Note: `scopelens` needs `git` on `PATH` and a full (non-shallow) clone; it exits `2` (not `0`) when it can't measure, so a broken checkout fails loudly instead of passing silently.

## Tool-level docs

Each tool ships on npm and PyPI with identical behavior. Pick the registry that matches your project's stack:

| Tool | npm (Node/TS) | PyPI (Python) |
|---|---|---|
| linelens   | [@open_harness/linelens](https://www.npmjs.com/package/@open_harness/linelens)     | [open-harness-linelens](https://pypi.org/project/open-harness-linelens/)     |
| dupelens   | [@open_harness/dupelens](https://www.npmjs.com/package/@open_harness/dupelens)     | [open-harness-dupelens](https://pypi.org/project/open-harness-dupelens/)     |
| secretlens | [@open_harness/secretlens](https://www.npmjs.com/package/@open_harness/secretlens) | [open-harness-secretlens](https://pypi.org/project/open-harness-secretlens/) |
| testlens   | [@open_harness/testlens](https://www.npmjs.com/package/@open_harness/testlens)     | [open-harness-testlens](https://pypi.org/project/open-harness-testlens/)     |
| scopelens  | [@open_harness/scopelens](https://www.npmjs.com/package/@open_harness/scopelens)   | [open-harness-scopelens](https://pypi.org/project/open-harness-scopelens/)   |

---

## Español

La suite completa de **open-harness** en un solo paquete. Instala los cinco tools de calidad de código de una sola línea: `linelens`, `dupelens`, `secretlens`, `testlens`, `scopelens`.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

> **Otros ecosistemas**: los tools individuales están también en **PyPI** como `open-harness-<tool>`. Este meta-paquete todo-en-uno es **solo de npm** — en PyPI se instalan los paquetes por tool. Packagist (PHP/Composer) está planeado. Los binarios y formatos de config son idénticos entre registries.

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
| **testlens**   | Detector de cobertura — encuentra archivos sin test, 10 lenguajes. |
| **scopelens**  | Presupuesto de archivos por PR — cuenta el diff rama-vs-base local en `pre-commit`. |

Después del install, los cinco binarios están disponibles via `npx`. Cada tool imprime su guía de configuración con `--tutorial`, y `npx open-harness init` crea los cinco archivos de config en la raíz de una.

### Configurá todo desde `package.json`

Cada tool lee su config desde una key dedicada dentro de `package.json` (ver ejemplo arriba). Podés combinar: archivos dedicados (`linelens.json`, etc.) para los tools que quieras y keys en `package.json` para el resto. Precedencia por tool: `--config <path>` > archivo dedicado `*.json` > key en `package.json` > defaults del binario. Guía completa: [docs/CONFIGURATION.md](https://github.com/artiko00/open-harness/blob/main/docs/CONFIGURATION.md).

### Integraciones

Sirve con Husky, lefthook o GitHub Actions con los snippets de la sección en inglés. `scopelens` necesita `git` en el `PATH` y un clon completo (no shallow); sale con código `2` cuando no puede medir, en vez de pasar en verde.

### Docs por tool

Cada tool tiene su propio README bilingüe con flags completos, formato de configuración y ejemplos. Los links de arriba apuntan a cada página de npm y PyPI.

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
