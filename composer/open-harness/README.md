# open-harness/open-harness (Composer)

The full **open-harness** suite in a single `composer require`. Installs all four code-quality linters at once: `linelens`, `dupelens`, `secretlens`, `testlens`.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

## Install

```bash
composer require --dev open-harness/open-harness
```

This is a Composer `metapackage` — no source code of its own, just declares dependencies on the 4 per-tool packages. Each tool's post-install hook then downloads the matching native binary from GitHub Releases.

## What you get

| Tool | Purpose |
|---|---|
| **linelens**   | File length linter — flags files over a configurable line limit. |
| **dupelens**   | Code duplication detector (Rabin-Karp, language-agnostic). |
| **secretlens** | Secret and credential detector (AWS keys, GitHub tokens, JWT, PEM, …). |
| **testlens**   | Test coverage detector — finds source files without tests, 9 languages. |

After install, the four binaries are available in `vendor/bin/`:

```bash
vendor/bin/linelens   check
vendor/bin/dupelens   check
vendor/bin/secretlens check
vendor/bin/testlens   check --lang php
```

## Configure from `composer.json`

Each tool reads its configuration from a dedicated key inside `composer.json` under `extra.open-harness` (planned). Today, configuration goes in `<tool>.json` files at the repo root.

```json
{
  "extra": {
    "open-harness": {
      "linelens":   { "default": { "maxLines": 100 } },
      "dupelens":   { "default": { "minTokens": 50, "minLines": 5 } },
      "secretlens": { "allowlist": ["example", "placeholder"] },
      "testlens":   { "language": "php", "exclude": ["vendor", "var"] }
    }
  }
}
```

Precedence per tool: `--config <path>` > dedicated `*.json` > `composer.json` extra > built-in defaults.

## Run as a CI / pre-commit gate

```yaml
# GitHub Actions
- run: composer install --no-progress --prefer-dist
- run: vendor/bin/linelens   check --fail
- run: vendor/bin/dupelens   check --fail
- run: vendor/bin/secretlens check --fail
- run: vendor/bin/testlens   check --fail --lang php
```

```yaml
# .grumphp.yml (with GrumPHP)
grumphp:
  tasks:
    open-harness-lines:
      metadata:
        task: shell
      command: 'vendor/bin/linelens check --fail'
```

## Per-tool docs

- [open-harness/linelens](https://packagist.org/packages/open-harness/linelens)
- [open-harness/dupelens](https://packagist.org/packages/open-harness/dupelens)
- [open-harness/secretlens](https://packagist.org/packages/open-harness/secretlens)
- [open-harness/testlens](https://packagist.org/packages/open-harness/testlens)

---

## Español

La suite completa de **open-harness** en un solo `composer require`. Instala los cuatro linters de calidad de código.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

### Instalación

```bash
composer require --dev open-harness/open-harness
```

Es un `metapackage` de Composer — no tiene código propio, solo declara dependencias a los 4 paquetes individuales. Cada tool tiene su hook post-install que descarga el binario nativo correspondiente desde GitHub Releases.

### Qué incluye

Los 4 binarios accesibles desde `vendor/bin/` tras el install: `linelens`, `dupelens`, `secretlens`, `testlens`.

### Configurá desde `composer.json`

Cada tool leerá su config desde `extra.open-harness.<tool>` (roadmap). Hoy la configuración va en archivos `<tool>.json` en la raíz del repo.

### CI / pre-commit

Sirve con GitHub Actions, GitLab CI o GrumPHP (snippets arriba).

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
