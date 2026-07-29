# Add scopelens: presupuesto de archivos por PR

Feature ID: **F-019** (`.agent/feature-list.json`)
Affected tools: **scopelens** (nuevo) · `_shared/pathmatch` (consumo) · lefthook · npm/PyPI/Packagist
Risk: **medium** (primer tool del monorepo que depende de un binario externo: `git`)

## Why

Los equipos que fijan un techo de archivos por PR ("no más de 15") hoy sólo pueden hacerlo cumplir
**después** de abrir el PR. Las dos herramientas de referencia del mercado lo confirman:

| Herramienta | Dónde corre | Qué hace | Puede bloquear |
|---|---|---|---|
| `pr-size-labeler` (cbrgm, Bedrock, actions-ecosystem) | GitHub Actions, sobre el PR | Etiqueta `size/xs…xl` según `diff` + `files` | **No.** Sólo etiqueta; el bloqueo requiere branch protection sobre la label |
| `Danger JS` | CI, con `DANGER_GITHUB_API_TOKEN` | Regla programable en `dangerfile.ts`; sólo `fail()` rompe el build | Sí, pero en CI |

Ambas comparten tres limitaciones que importan para un gate real:

1. **Llegan tarde.** El feedback aparece minutos después del push, con el trabajo ya empujado y el PR
   ya abierto. Reorganizar 40 archivos en tres PRs a esa altura cuesta mucho más que no haberlos
   commiteado juntos.
2. **Leen el diff por la API de GitHub, no por git.** GitHub trunca el diff a 300 archivos (HTTP 406
   `diff exceeded the maximum number of files`) y la lista a 3000. Hay bugs abiertos donde
   `danger.git.modified_files` devuelve 898 archivos mientras `git diff` local muestra 2549
   (danger-js #1249, #1432). El techo se rompe exactamente en los PRs grandes, que son los únicos
   que el gate necesita atrapar.
3. **Requieren red, token y un PR abierto.** Nada de eso existe en el momento del `git commit`.

Y una limitación propia del ecosistema del repo: ninguna de las dos sirve fuera de GitHub Actions +
Node. Un equipo Python o Go que instala open-harness desde PyPI o `go install` no va a montar un
runtime de Node y un token de GitHub para contar archivos.

`scopelens` corre local, sobre `git`, sin red ni token, y **aborta el commit** antes de que el PR
exista. Es el mismo contrato que ya tienen los otros cuatro lenses: `check --fail --no-color` en
`pre-commit`, `exit 1` cuando el gate no pasa.

## What Changes

### Núcleo: medición del alcance contra la base, no contra el commit

La unidad de la política es el **PR**, no el commit. Cinco commits de 4 archivos suman 20 y deben
fallar. `scopelens` cuenta la unión de:

- `git diff --name-only --diff-filter=ACMRD -M <merge-base>...HEAD` — lo ya commiteado en la rama
- `git diff --name-only --diff-filter=ACMRD -M --cached` — lo que está por commitearse

### El gate no falla en verde (lección de F-018)

`scopelens` no puede inventar un conteo cuando le falta información. Cada una de estas condiciones
produce `exit 2` con mensaje accionable, nunca `exit 0`:

- el cwd no es un repo git
- el binario `git` no está en `PATH`
- el clon es shallow (`actions/checkout` con `fetch-depth: 1` rompe `merge-base`)
- la rama base no existe localmente ni como remota

### Multi-ecosistema (JS/TS · Python · Go)

El conteo es agnóstico del lenguaje, pero tres cosas no lo son y son el grueso del trabajo:

- **Cadena de config (ADR-018)**: `scopelens.json` → `pyproject.toml [tool.scopelens]` →
  `package.json "scopelens"` → `composer.json extra.open-harness.scopelens` → defaults compilados.
- **Exclusiones por defecto por ecosistema**: un `pnpm-lock.yaml`, un `poetry.lock` o un `go.sum`
  regenerados no son superficie de review y no deben consumir presupuesto.
- **Clasificación source/test por ecosistema**: un PR de 8 fuentes + 8 tests son 16 archivos y
  bloquearía con un techo de 15, castigando justamente al PR que sí trae tests. `scopelens` reconoce
  los layouts de test de los tres ecosistemas (reutilizando el criterio de testlens F-015) y permite
  descontarlos del presupuesto.

### Distribución

npm `@open_harness/scopelens` (4 plataformas + meta), PyPI y Packagist, en el mismo tren que los
otros cuatro (F-012, F-013).

## In Scope

- Nuevo tool `tools/scopelens/` con `check` y `version`, patrón `osExit` + `run([]string) int`
- Adaptador `git` sobre `os/exec` con timeout y errores tipados
- Descubrimiento de base: `--base`, config, `origin/<default>`, `<default>` local
- Categorías `source` / `test` / `excluded` en el reporte, con desglose
- Cadena de config multi-ecosistema (reutiliza `tomlmin` y el patrón de `config_chain.go`)
- Defaults de exclusión para JS/TS, Python, Go
- Flags: `--max-files`, `--base`, `--staged-only`, `--exclude-tests`, `--fail`, `--no-color`, `--dir`
- ADR nuevo: dependencia de `git` como binario externo
- Entrada en `lefthook.yml`, `open-harness.json`, `scripts/build-npm.sh`
- 100% de cobertura de statements (ADR-011), TDD estricto (ADR-013)

## Out of Scope

- Umbral por líneas de diff (`+/-`). El pedido es archivos; agregar líneas es otra feature.
- Etiquetado de PRs en GitHub, comentarios, o cualquier llamada de red.
- Integración con plataformas (GitLab, Bitbucket) más allá de lo que `git` ya resuelve.
- Categorías de tamaño `xs/s/m/l/xl`. Aquí hay un único umbral binario: pasa o no pasa.
- Historial o telemetría de tamaños de PR a lo largo del tiempo.
- Bypass propio. Si hay que saltarse el gate se usa `git commit --no-verify`, como con los otros lenses.
