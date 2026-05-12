# @open_harness/secretlens

Secret and credential detector for any codebase. Scans source files for hardcoded AWS keys, GitHub tokens, PEM private keys, JWTs, and generic credential assignments. Single native binary, zero runtime dependencies.

Part of the [open-harness](https://github.com/artiko00/open-harness) monorepo. [Español abajo](#español).

## Install

```bash
npm install --save-dev @open_harness/secretlens
```

The right native binary for your platform (Linux x64, macOS arm64, macOS x64, Windows x64) is fetched automatically via `optionalDependencies`.

## Usage

```bash
npx secretlens check              # scan current directory
npx secretlens check --fail       # exit 1 if secrets found (git hooks / CI)
npx secretlens check --dir ./src  # scan a specific directory
npx secretlens check --no-color   # plain output for logs
npx secretlens init               # generate a default secretlens.json
npx secretlens version            # print version
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

## Integrations

```bash
# Husky pre-commit
npx secretlens check --fail
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
npm install --save-dev @open_harness/secretlens
```

El binario para tu plataforma se descarga automáticamente via `optionalDependencies`.

### Uso

```bash
npx secretlens check              # escanea el directorio actual
npx secretlens check --fail       # exit 1 si encuentra secretos (git hooks / CI)
npx secretlens check --dir ./src  # escanea un directorio específico
npx secretlens check --no-color   # salida sin colores
npx secretlens init               # genera un secretlens.json por defecto
npx secretlens version            # imprime la versión
```

### Patrones integrados

Los 8 patrones built-in cubren claves AWS, tokens GitHub (clásicos y fine-grained), claves privadas PEM, JWTs y asignaciones genéricas tipo `secret=…`, `password=…`, `api_key=…`, `token=…`, `bearer …`. Ver la tabla arriba para severidades exactas.

### Configuración

Colocá un `secretlens.json` en la raíz del repo (ver ejemplo arriba).

- `patterns: []` usa los 8 patrones built-in. Sobrescribí el array para agregar regexes propias.
- `allowlist` salta cualquier línea que contenga los strings indicados (case-insensitive) — útil para suprimir falsos positivos en docs o ejemplos.
- `exclude` ignora completamente los directorios que matcheen.

### Integraciones

Sirve con Husky, lefthook o GitHub Actions con los mismos snippets de la sección en inglés.

### Códigos de salida

| Código | Significado |
|---|---|
| `0` | No se detectaron secretos (o no se pasó `--fail`) |
| `1` | Hubo secretos con `--fail`, o error de configuración |

## License

MIT — see the [main repository](https://github.com/artiko00/open-harness).
