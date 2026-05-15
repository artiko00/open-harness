# open-harness-secretlens

Secret and credential detector for any codebase. Scans source files for hardcoded AWS keys, GitHub tokens, PEM private keys, JWTs, and generic credential assignments. Single native binary, zero runtime dependencies.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

> **Same tool, other ecosystems**: also available on **npm** ([`@open_harness/secretlens`](https://www.npmjs.com/package/@open_harness/secretlens)) and on **Packagist** (`open-harness/secretlens`). Identical binary, identical config; pick the registry that matches your stack.

## Install

```bash
pip install open-harness-secretlens
```

pip picks the right native wheel for your platform automatically (Linux x86_64, macOS arm64, macOS x86_64, Windows x86_64). Each wheel embeds the Go binary — no runtime deps.

## Usage

```bash
secretlens check              # scan current directory
secretlens check --fail       # exit 1 if secrets found (git hooks / CI)
secretlens check --dir ./src  # scan a specific directory
secretlens check --no-color   # plain output for logs
secretlens init               # generate a default secretlens.json
secretlens version            # print version
```

## Built-in patterns

| Pattern | Severity |
|---|---|
| AWS Access Key ID (`AKIA…`) | critical |
| AWS Secret Access Key | critical |
| GitHub Personal Access Token (`ghp_…`) | critical |
| GitHub Fine-Grained Token (`github_pat_…`) | critical |
| PEM Private Key (`-----BEGIN … PRIVATE KEY`) | critical |
| JWT Token | high |
| Generic `secret/password/api_key` assignment | high |
| Generic `token/bearer` assignment | medium |

## Configuration

Place a `secretlens.json` at the repo root:

```json
{
  "patterns": [],
  "allowlist": ["example", "placeholder", "your_key_here", "changeme"],
  "exclude": ["node_modules", "vendor", ".git", "dist"]
}
```

- `patterns: []` uses the 8 built-in patterns. Override the array to add custom regexes.
- `allowlist` skips any line containing the listed strings (case-insensitive) — useful to suppress false positives in docs or examples.
- `exclude` skips matching directories entirely.

### Alternative: configure inside `pyproject.toml` or the dedicated `secretlens.json`

If you prefer not to keep a separate `secretlens.json`, add a `secretlens` key in your `package.json` with the same shape:

```json
{
  "name": "my-project",
  "secretlens": {
    "allowlist": ["example", "your_key_here"],
    "exclude": ["node_modules", "dist"]
  }
}
```

Precedence: `--config <path>` > `secretlens.json` > `package.json` key > built-in defaults. CLI flags (`--no-color`, etc.) always win.

## Integrations

```bash
# Husky pre-commit
secretlens check --fail
```

```yaml
# GitHub Actions
- name: Scan for hardcoded secrets
  run: npx @open_harness/secretlens check --fail
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | No secrets detected (or `--fail` not passed) |
| `1` | Secrets found and `--fail` was passed, or config error |

---

## Español

Detector de secretos y credenciales para cualquier base de código. Escanea archivos buscando claves AWS, tokens de GitHub, claves privadas PEM, JWTs y asignaciones genéricas de credenciales hardcodeadas. Un solo binario nativo, cero dependencias.

Parte del monorepo [open-harness](https://github.com/artiko00/open-harness).

### Instalación

```bash
pip install open-harness-secretlens
```

pip descarga automáticamente la wheel nativa correcta para tu plataforma.

### Uso

```bash
secretlens check              # escanea el directorio actual
secretlens check --fail       # exit 1 si encuentra secretos (git hooks / CI)
secretlens check --dir ./src  # escanea un directorio específico
secretlens check --no-color   # salida sin colores
secretlens init               # genera un secretlens.json por defecto
secretlens version            # imprime la versión
```

### Patrones integrados

Los 8 patrones built-in cubren claves AWS, tokens GitHub (clásicos y fine-grained), claves privadas PEM, JWTs y asignaciones genéricas tipo `secret=…`, `password=…`, `api_key=…`, `token=…`, `bearer …`. Ver la tabla arriba para severidades exactas.

### Configuración

Colocá un `secretlens.json` en la raíz del repo (ver ejemplo arriba).

- `patterns: []` usa los 8 patrones built-in. Sobrescribí el array para agregar regexes propias.
- `allowlist` salta cualquier línea que contenga los strings indicados (case-insensitive) — útil para suprimir falsos positivos en docs o ejemplos.
- `exclude` ignora completamente los directorios que matcheen.

#### Alternativa: configurar dentro de `pyproject.toml` o `secretlens.json`

Si preferís no tener un `secretlens.json` separado, agregá una key `secretlens` en tu `package.json` con la misma forma del archivo dedicado. Precedencia: `--config <path>` > `secretlens.json` > key en `package.json` > defaults. Los flags CLI siempre ganan.

### Integraciones

Sirve con Husky, lefthook o GitHub Actions con los mismos snippets de la sección en inglés.

### Códigos de salida

| Código | Significado |
|---|---|
| `0` | No se detectaron secretos (o no se pasó `--fail`) |
| `1` | Hubo secretos con `--fail`, o error de configuración |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
