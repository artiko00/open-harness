# AGENTS.md

Instrucciones para agentes (Claude Code, Codex, Cursor, etc.) que trabajen sobre `open-harness`.

> Este archivo es la **fuente de verdad** del workflow del proyecto. Si algo en este documento contradice una respuesta intuitiva del agente, gana este documento.

---

## 1. Filosofía del proyecto

Monorepo de 5 herramientas de calidad de código. Cada tool:

- Es un **binario único Go**, **cero dependencias** runtime ([ADR-001](docs/adr-001-go-sobre-node.md), [ADR-002](docs/adr-002-cero-dependencias.md))
- Es **language-agnostic** (debe servir en cualquier ecosistema vía wrapper npm)
- Cumple la **regla de 100 líneas por archivo** ([ADR-005](docs/adr-005-regla-100-lineas-aplicada-al-proyecto.md)) — el gate mide **líneas de código**, no líneas físicas
- Mantiene **alta cobertura de tests** ([ADR-011](docs/adr-011-cobertura-100-como-estandar.md))
- Está documentado por sus decisiones en `docs/adr-*.md`

Si tu cambio rompe alguna de estas premisas, **detente y abre un ADR antes de continuar**.

---

## 2. Tech stack

- Runtime: Go 1.22+
- Language: Go (no framework, sin deps externas en tools)
- Package manager: Go modules + Go workspace (`go.work`)
- Testing: Go stdlib (`testing` package)
- Linting: linelens + dupelens + secretlens + testlens + scopelens (auto-protección activa)
- Hooks: lefthook (`pre-commit` + `pre-push`)

### 2.1 Módulos compartidos (`tools/_shared/`)

Los cinco binarios comparten cuatro módulos, cada uno con su propio `go.mod` y
enlazado por el Go workspace ([ADR-020](docs/adr-020-modulos-compartidos-y-duplicacion-estructural.md)):

| Módulo | Responsabilidad |
|---|---|
| `tomlmin` | Parser TOML del subset de `pyproject.toml` ([ADR-018](docs/adr-018-config-multi-ecosistema.md)) |
| `configload` | Cadena de configuración (`pyproject.toml` → `package.json` → `composer.json`) |
| `pathmatch` | Semántica de globs estilo gitignore ([ADR-006](docs/adr-006-semantica-glob-gitignore.md)) |
| `langsyntax` | Sintaxis por familia de lenguaje (imports, comentarios) |

**Un cambio acá se publica en los cinco tools a la vez.** Antes de tocarlos:

- Corré los tests de los cinco tools **y** de los cuatro `_shared`, no solo el módulo que editaste.
- Presupuestá el release completo (bump ×5 + npm ×26 + PyPI ×6), no solo el parche.
- `go build ./...` desde la raíz **falla**: es un workspace con un módulo por directorio. Hay que entrar a cada uno.

---

## 3. Workflow obligatorio: TDD (Test-Driven Development)

**Toda feature nueva o cambio de comportamiento sigue el ciclo Red → Green → Refactor.** Sin excepciones.

### 3.1 Red (test que falla primero)

1. Antes de escribir código de producción, **escribe el test**
2. El test debe **fallar** al correrlo: `go test ./...` produce `FAIL`
3. El test debe ser específico, no genérico — verifica un comportamiento concreto

### 3.2 Green (mínimo código para pasar)

1. Escribe **lo mínimo** para que el test pase
2. No optimices, no anticipes futuras features
3. Re-corre `go test ./...` hasta ver `PASS`

### 3.3 Refactor

1. Con tests en verde, refactoriza para legibilidad y SOLID
2. Re-corre tests después de cada cambio

### 3.4 Reglas duras

- **Nunca** commitees código sin sus tests asociados
- **Nunca** marques un step como `[done]` si los tests no pasan
- Tests con assertions genéricas (`assert.True(true)`) **cuentan como no tener test**
- Cobertura objetivo: **100% statement coverage** ([ADR-011](docs/adr-011-cobertura-100-como-estandar.md))

Detalle completo en [ADR-013](docs/adr-013-tdd-como-estandar.md).

---

## 4. CLI pattern para `main.go` testeable

```go
var osExit = os.Exit // inject for testing

func main() {
    osExit(run(os.Args[1:]))
}

func run(args []string) int {
    // return exit code
}
```

- Usa `flag.ContinueOnError` (no `ExitOnError`) en subcomandos
- Nunca `os.Exit` directo en business logic — usar `osExit` variable

### 4.1 Exit codes

Los primeros 4 tools (linelens, dupelens, secretlens, testlens) usan sólo **`0`** (ok) y **`1`** (violaciones encontradas, con `--fail`).

**`scopelens` (v0.2.1) agrega el exit code `2`: "no se pudo medir".** El gate no puede inventar un conteo cuando le falta información, así que **nunca falla en verde**: cada condición que impide una medición confiable devuelve `2`, no `0`.

| Code | Semántica |
|---|---|
| `0` | Medido y dentro del presupuesto (o fuera, sin `--fail`) |
| `1` | Medido y excede el presupuesto (con `--fail`) |
| `2` | **No se pudo medir**: `git` ausente del `PATH`, el cwd no es un repo, clon shallow (`merge-base` no resoluble), rama base no encontrada, config inválida, o error de uso (flag desconocido, `--max-files` negativo, `--dir` inaccesible, subcomando desconocido) |

En `pre-commit`, el `2` también aborta el commit: una medición rota nunca se trata como aprobación silenciosa. El bypass deliberado sigue siendo `git commit --no-verify`.

---

## 5. Comandos esenciales

| Acción | Comando |
|---|---|
| Bootstrap entorno | `bash .agent/init.sh` |
| Build single tool | `go build -o tools/<name>/<name> ./tools/<name>` |
| Build all tools | `bash scripts/build.sh` |
| Test single tool | `cd tools/<name> && go test ./... -v` |
| Test all (5 tools + 4 `_shared`) | `for m in tools/*/ tools/_shared/*/; do (cd $m && go test ./...); done` |
| Coverage | `cd tools/<name> && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out` |
| Lint repo | `./tools/linelens/linelens check --fail` |
| Install hooks | `lefthook install` |
| Verificar versiones | `bash scripts/check-versions.sh` |

> `go test ./tools/...` y `go build ./...` desde la raíz **no funcionan**: el
> workspace tiene un módulo por directorio, así que los comandos se corren
> dentro de cada uno.

---

## 6. Estructura del repo

```
open-harness/
├── tools/
│   ├── linelens/        ← v0.3.3 (file length linter)
│   ├── dupelens/        ← v0.4.1 (duplicate detector, Rabin-Karp)
│   ├── secretlens/      ← v0.3.3 (secret/credential detector)
│   ├── testlens/        ← v0.3.3 (test coverage detector, multi-language)
│   └── scopelens/       ← v0.2.1 (per-PR file+line budget gate sobre git, exit 2 = no medible)
│   └── _shared/         ← módulos compartidos por los 5 binarios (sección 2.1)
│       ├── tomlmin/     ← parser TOML de pyproject.toml
│       ├── configload/  ← cadena de configuración multi-ecosistema
│       ├── pathmatch/   ← globs estilo gitignore
│       └── langsyntax/  ← sintaxis por familia de lenguaje
├── npm/@open_harness/   ← wrappers npm por plataforma (5 tools × 4 plataformas + meta)
├── pypi/                ← paquetes PyPI (5 tools + meta open-harness-suite)
├── docs/                ← ADRs (decisiones arquitectónicas) + RELEASE.md, CONFIGURATION.md, UPGRADING.md
├── scripts/             ← build.sh, build-npm.sh, build-pypi.sh, check-versions.sh
├── .agent/              ← harness: feature-list, claude-progress, init, .gitignore
├── .claude/             ← config de Claude Code (permisos compartidos)
├── openspec/            ← Spec-Driven Development (changes + specs promovidas)
├── go.work              ← Go workspace (5 tools + 4 módulos _shared)
├── lefthook.yml         ← git hooks (pre-commit: 5 gates, pre-push: tests de los 5 tools + _shared)
└── {linelens,dupelens,secretlens,testlens,scopelens}.json  ← configs de auto-protección
```

---

## 7. Cómo agregar una nueva feature

1. **Lee** [`.agent/feature-list.json`](.agent/feature-list.json) para entender qué está planeado
2. **Identifica** la feature ID (F-XXX). Si no existe, agrégala antes de codear
3. **Trabajá en `develop`** — no se crean ramas feature (sección 7.1)
4. **Abrí un change de openspec** (`openspec new change <nombre>`) con proposal, delta specs y tasks antes de codear
5. **Sigue TDD** por cada step de la feature (sección 3)
6. **Mantén archivos ≤ 100 líneas de código** — si un archivo crece, parte la responsabilidad
7. **Documenta decisiones no obvias** en un ADR nuevo (`docs/adr-NNN-titulo.md`), o como sección de un ADR existente si es un matiz de una decisión ya tomada
8. **Actualiza el feature-list** marcando steps como `[done]` solo cuando tests pasen
9. **Commit atómicos**: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
10. **Archivá el change** al terminar (`openspec archive <nombre> -y`) para promover los delta specs a `openspec/specs/`

### 7.1 Flujo de ramas

**Solo `develop` y `main`.** El trabajo va directo en `develop`; cuando está
listo se mergea a `main`, que es desde donde se publica. No se crean ramas
`feat/`, `fix/` ni `epic/`.

---

## 8. Cómo agregar un nuevo tool al monorepo

1. `mkdir tools/<name>` con `go.mod` (module path: `github.com/artiko00/open-harness/tools/<name>`)
2. Agregar al `go.work` con `use ./tools/<name>`
3. Registrar en `open-harness.json`
4. Agregar entradas al `.gitignore` para los binarios (`<name>` y `<name>.exe`)
5. Implementar **siguiendo TDD** desde el primer commit, **100% coverage objetivo**
6. Cumplir filosofía: un solo binario, cero deps, regla de 100 líneas
7. Crear ADR si la implementación toma decisiones no obvias

---

## 9. Convenciones

**Go:**
- Solo stdlib, prohibido agregar deps externas sin ADR
- `package main` en cada tool (CLI standalone)
- `flag` package estándar — no `cobra`, no `urfave/cli`
- Errores: `fmt.Errorf("context: %w", err)` para envolver

**Tests:**
- Archivos `*_test.go` co-localizados
- Table-driven cuando hay varios casos similares
- Cobertura 100% statement requerida ([ADR-011](docs/adr-011-cobertura-100-como-estandar.md))

---

## 10. Boundaries

### Always do
- TDD para cada cambio de comportamiento (sección 3)
- `go test ./...` dentro de cada módulo tocado antes de commitear
- Quality gates de auto-protección (los 5 tools sobre sí mismos) antes de cada commit — lefthook lo hace automático
- Mantener 100% coverage en código nuevo
- **Una feature por sesión** — terminar (tests + lint + commit) antes de abrir otra
- Antes de marcar listo: actualizar `.agent/claude-progress.txt` con qué se hizo y próximos pasos

### Ask first
- Antes de agregar dependencias externas
- Antes de modificar `go.work`
- Antes de agregar nuevos tools (revisar feature-list)
- Antes de **ampliar el alcance de un módulo `_shared`**: implica republicar los cinco tools

### Never do
- Commitear binarios compilados (excluidos en `.gitignore`)
- Saltar fallos de checks en CI
- `flag.ExitOnError` en subcomandos (usar `ContinueOnError`)
- `os.Exit` directo en business logic (usar `osExit` variable)
- Crear ramas feature — el trabajo va en `develop` y de ahí a `main` (sección 7.1)
- Skipear hooks con `--no-verify`, **salvo el commit de release**: toca > `maxFiles`
  archivos y excede el gate de scopelens por construcción. Ahí se verifican los
  otros cuatro gates a mano primero y se deja constancia en el mensaje del commit
  ([docs/RELEASE.md](docs/RELEASE.md))

---

## 11. Quality gates (auto-protección)

El repo aplica sus 5 tools sobre sí mismo. Antes de mergear a `main` todos deben pasar:

```bash
./tools/linelens/linelens     check --fail   # archivos ≤ límites (linelens.json)
./tools/dupelens/dupelens     check --fail   # cero duplicación significativa (dupelens.json)
./tools/secretlens/secretlens check --fail   # cero secretos hardcodeados (secretlens.json)
./tools/testlens/testlens     check --fail   # cero archivos Go sin tests
./tools/scopelens/scopelens   check --fail   # presupuesto de archivos y líneas del diff
```

lefthook `pre-commit` automatiza los 5 gates; `pre-push` re-corre además la
batería de tests de los 5 tools y de los módulos `_shared`. Cualquier excepción a
estos gates va vía ADR — con la única salvedad ya documentada del commit de
release, que excede el presupuesto de scopelens por construcción.

---

> **Proceso de release** (bump, sincronización de manifiestos npm/PyPI, publish,
> tags y flujo de ramas): ver [docs/RELEASE.md](docs/RELEASE.md). El gate único
> antes de publicar es `bash scripts/check-versions.sh`, que verifica que la
> versión coincida en `main.go`, `open-harness.json`, README/AGENTS y **todos los
> manifiestos de distribución** (npm y PyPI).

## 12. Errores recurrentes a evitar

| Error | Síntoma | Solución |
|---|---|---|
| Modificar `feature-list.json` sin terminar la feature | `steps` marcados `[done]` con tests rojos | Re-correr `go test`, revertir el `[done]` si falla |
| `npm publish` del wrapper antes que los 4 paquetes de plataforma | Usuarios instalan y reciben "platform not supported" | Publicar siempre `linelens-linux-x64`, `…-darwin-arm64`, `…-darwin-x64`, `…-win32-x64` antes del wrapper |
| Bumpear versión sin re-compilar binarios | `package.json` dice 0.1.X pero `bin/` tiene el binario viejo | Correr `scripts/build-npm.sh <tool>` antes de cada publish |
| Cambiar `module path` en `go.mod` sin actualizar wrappers | Tests rompen porque imports cruzados no resuelven | Usar `grep -r 'github.com/' tools/` antes y después del rename |
| Olvidar `--access public` en `npm publish` de scope | npm tira `402 Payment Required` | El script imprime el comando completo; copiar y pegar |
| Agregar campo nuevo al `Config` struct sin actualizar todos los formatos | `linelens.json` lo lee pero `pyproject.toml` no, y los tests pasan porque cada fuente solo testea su propio happy-path | Cuando toques un `Config`, ejecutá los 8+8+8+7 tests de `config_pyproject_test.go`, `config_pkg_test.go`, `config_composer_test.go` y `config_test.go` por tool antes de commitear |
| Testear un parser de manifiestos solo con fixtures que uno mismo escribe | El subset soporta lo que el fixture usa, y el archivo real del usuario —lleno de secciones ajenas— rompe (F-023: `dependencies` multilínea de PEP 621 tumbaba los 5 tools) | Golden tests con manifiestos **reales** de cada herramienta del ecosistema (`tools/_shared/tomlmin/testdata/`: poetry, setuptools, hatch, uv), no solo la tabla `[tool.<name>]` |
| Tratar como fatal la sintaxis que está fuera de la sección que te interesa | Un `[tool.poetry]` exótico deja sin config a un tool que solo quería leer `[tool.linelens]` | Estricto dentro de la sección pedida, descarte silencioso fuera ([ADR-018](docs/adr-018-config-multi-ecosistema.md)) |
| Dar por verificado un fix de config porque el tool imprime `OK` | El `OK` puede venir de haber caído a los defaults tras fallar la carga | Usar un valor de config que **cambie el resultado** (un `maxLines` bajo que produzca violación) y comparar contra el binario publicado en `npm/@open_harness/<tool>-linux-x64/bin/` |
