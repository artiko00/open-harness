# open-harness

The full **open-harness** suite in a single `pip install`. Installs all four code-quality linters at once: `linelens`, `dupelens`, `secretlens`, `testlens`.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

## Install

```bash
pip install open-harness
```

pip resolves the dependency tree and downloads the right native wheel per platform for each tool (Linux x86_64, macOS arm64, macOS x86_64, Windows x86_64).

## What you get

| Tool | Purpose |
|---|---|
| **linelens**   | File length linter — flags files over a configurable line limit. |
| **dupelens**   | Code duplication detector (Rabin-Karp, language-agnostic). |
| **secretlens** | Secret and credential detector (AWS keys, GitHub tokens, JWT, PEM, …). |
| **testlens**   | Test coverage detector — finds source files without tests, 9 languages. |

After install, all four binaries live in your venv's `bin/`:

```bash
linelens   check
dupelens   check
secretlens check
testlens   check --lang python
```

## Configure from `pyproject.toml`

Each tool will read its config from `[tool.<name>]` inside your `pyproject.toml` (planned, see roadmap). Today, configuration goes in `<tool>.json` files at the repo root:

```json
// linelens.json
{ "default": { "maxLines": 100 } }
```

Precedence per tool: `--config <path>` > dedicated `*.json` > `pyproject.toml` table > built-in defaults.

## Run all four as a single CI gate

```yaml
# GitHub Actions / GitLab CI / etc.
- pip install open-harness
- linelens   check --fail
- dupelens   check --fail
- secretlens check --fail
- testlens   check --fail --lang python
```

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: linelens
        name: linelens
        entry: linelens check --fail --no-color
        language: system
        pass_filenames: false
      - id: dupelens
        name: dupelens
        entry: dupelens check --fail --no-color
        language: system
        pass_filenames: false
      - id: secretlens
        name: secretlens
        entry: secretlens check --fail --no-color
        language: system
        pass_filenames: false
      - id: testlens
        name: testlens
        entry: testlens check --fail --lang python
        language: system
        pass_filenames: false
```

## Per-tool docs

Each tool has its own PyPI page with full flags, config shape, and examples:

- [open-harness-linelens](https://pypi.org/project/open-harness-linelens/)
- [open-harness-dupelens](https://pypi.org/project/open-harness-dupelens/)
- [open-harness-secretlens](https://pypi.org/project/open-harness-secretlens/)
- [open-harness-testlens](https://pypi.org/project/open-harness-testlens/)

---

## Español

La suite completa de **open-harness** en un solo `pip install`. Instala los cuatro linters de calidad de código: `linelens`, `dupelens`, `secretlens`, `testlens`.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

### Instalación

```bash
pip install open-harness
```

pip descarga automáticamente la wheel nativa correcta para tu plataforma por cada tool (Linux x86_64, macOS arm64, macOS x86_64, Windows x86_64).

### Qué incluye

Los 4 binarios disponibles en el `bin/` de tu venv tras el install: `linelens`, `dupelens`, `secretlens`, `testlens`.

### Configurá desde `pyproject.toml`

Cada tool leerá su config desde `[tool.<name>]` dentro de tu `pyproject.toml` (roadmap). Hoy la configuración va en archivos `<tool>.json` en la raíz del repo.

### CI / pre-commit

Sirve con GitHub Actions, GitLab CI o el framework `pre-commit` (snippets arriba).

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
