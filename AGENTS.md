# AGENTS.md

Instrucciones para agentes (Claude Code, Codex, Cursor, etc.) que trabajen sobre `open-harness`.

> Este archivo es la **fuente de verdad** del workflow del proyecto. Si algo en este documento contradice una respuesta intuitiva del agente, gana este documento.

---

## 1. Filosofía del proyecto

Monorepo de 4 herramientas de calidad de código. Cada tool:

- Es un **binario único Go**, **cero dependencias** runtime ([ADR-001](docs/adr-001-go-sobre-node.md), [ADR-002](docs/adr-002-cero-dependencias.md))
- Es **language-agnostic** (debe servir en cualquier ecosistema vía wrapper npm)
- Cumple la **regla de 100 líneas por archivo** ([ADR-005](docs/adr-005-regla-100-lineas-aplicada-al-proyecto.md))
- Mantiene **alta cobertura de tests** ([ADR-011](docs/adr-011-cobertura-100-como-estandar.md))
- Está documentado por sus decisiones en `docs/adr-*.md`

Si tu cambio rompe alguna de estas premisas, **detente y abre un ADR antes de continuar**.

---

## 2. Tech stack

- Runtime: Go 1.22+
- Language: Go (no framework, sin deps externas en tools)
- Package manager: Go modules + Go workspace (`go.work`)
- Testing: Go stdlib (`testing` package)
- Linting: linelens + dupelens + secretlens + testlens (auto-protección activa)
- Hooks: lefthook (`pre-commit` + `pre-push`)

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

---

## 5. Comandos esenciales

| Acción | Comando |
|---|---|
| Bootstrap entorno | `bash .agent/init.sh` |
| Build single tool | `go build -o tools/<name>/<name> ./tools/<name>` |
| Build all tools | `bash scripts/build.sh` |
| Test single tool | `cd tools/<name> && go test ./... -v` |
| Test all | `go test ./tools/...` |
| Coverage | `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out` |
| Lint repo | `./tools/linelens/linelens check --fail` |
| Install hooks | `lefthook install` |

---

## 6. Estructura del repo

```
open-harness/
├── tools/
│   ├── linelens/        ← v0.1.0 (file length linter)
│   ├── dupelens/        ← v0.1.0 (duplicate detector, Rabin-Karp)
│   ├── secretlens/      ← v0.1.0 (secret/credential detector)
│   └── testlens/        ← v0.1.0 (test coverage detector, multi-language)
├── npm/@open_harness/   ← wrappers npm por plataforma
├── docs/                ← ADRs (decisiones arquitectónicas)
├── scripts/             ← build.sh, build-npm.sh, bench-vs-jscpd.sh
├── .agent/              ← harness: feature-list, progress, init
├── go.work              ← Go workspace (4 tools)
├── lefthook.yml         ← git hooks (pre-commit: 3 tools, pre-push: tests x4)
├── linelens.json        ← linelens config for this repo
├── dupelens.json        ← dupelens config for this repo
└── secretlens.json      ← secretlens config for this repo
```

---

## 7. Cómo agregar una nueva feature

1. **Lee** [`.agent/feature-list.json`](.agent/feature-list.json) para entender qué está planeado
2. **Identifica** la feature ID (F-XXX). Si no existe, agrégala antes de codear
3. **Crea branch**: `epic/<nombre>` para épicas, `feat/<id>-descripción` para features acotadas
4. **Sigue TDD** por cada step de la feature (sección 3)
5. **Mantén archivos ≤ 100 líneas** — si un archivo crece, parte la responsabilidad
6. **Documenta decisiones no obvias** en un ADR nuevo (`docs/adr-NNN-titulo.md`)
7. **Actualiza el feature-list** marcando steps como `[done]` solo cuando tests pasen
8. **Commit atómicos**: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
9. **No pushees a main directo** — abre PR desde tu branch

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
- `go test ./...` antes de commitear
- `linelens check --fail` antes de cada commit (lefthook lo hace automático)
- Mantener 100% coverage en código nuevo

### Ask first
- Antes de agregar dependencias externas
- Antes de modificar `go.work`
- Antes de agregar nuevos tools (revisar feature-list)

### Never do
- Commitear binarios compilados (excluidos en `.gitignore`)
- Saltar fallos de checks en CI
- `flag.ExitOnError` en subcomandos (usar `ContinueOnError`)
- `os.Exit` directo en business logic (usar `osExit` variable)
- Push directo a `main` — todo va por PR
- Skipear hooks con `--no-verify` salvo emergencia justificada
