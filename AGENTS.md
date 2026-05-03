# AGENTS.md

Instrucciones para agentes (Claude Code, Codex, Cursor, etc.) que trabajen sobre `open-harness`.

> Este archivo es la **fuente de verdad** del workflow del proyecto. Si algo en este documento contradice una respuesta intuitiva del agente, gana este documento.

---

## 1. Filosofía del proyecto

Monorepo de herramientas de calidad de código. Cada tool:

- Es un **binario único Go**, **cero dependencias** runtime ([ADR-001](docs/adr-001-go-sobre-node.md), [ADR-002](docs/adr-002-cero-dependencias.md))
- Es **language-agnostic** (debe servir en cualquier ecosistema vía wrapper npm)
- Cumple la **regla de 100 líneas por archivo** ([ADR-005](docs/adr-005-regla-100-lineas-aplicada-al-proyecto.md))
- Está documentado por sus decisiones en `docs/adr-*.md`

Si tu cambio rompe alguna de estas premisas, **detente y abre un ADR antes de continuar**.

---

## 2. Tech stack

- Runtime: Go 1.22+
- Language: Go (no framework, sin deps externas en tools)
- Package manager: Go modules + Go workspace (`go.work`)
- Testing: Go stdlib (`testing` package)
- Linting: linelens (custom, builtin) — auto-protección activa
- Hooks: lefthook (`pre-commit` + `pre-push`)

---

## 3. Workflow obligatorio: TDD (Test-Driven Development)

**Toda feature nueva o cambio de comportamiento sigue el ciclo Red → Green → Refactor.** Sin excepciones.

### 3.1 Red (test que falla primero)

1. Antes de escribir código de producción, **escribe el test**
2. El test debe **fallar** al correrlo: `go test ./...` produce `FAIL`
3. El test debe ser específico, no genérico — verifica un comportamiento concreto, no "que algo no crashee"
4. Solo cuando el test falla por la razón esperada (no por error de compilación tonto), continúas

### 3.2 Green (mínimo código para pasar)

1. Escribe **lo mínimo** para que el test pase
2. No optimices, no anticipes futuras features, no agregues "while we're here"
3. Re-corre `go test ./...` hasta ver `PASS`
4. Si el test pasa "por accidente" o el comportamiento real es distinto a lo esperado, **mejora el test** antes de continuar

### 3.3 Refactor (limpiar con red de seguridad)

1. Con tests en verde, refactoriza para legibilidad y SOLID
2. Re-corre tests después de cada cambio significativo
3. Si introduces nuevo comportamiento durante el refactor, **vuelve al Red** — agrega un test antes

### 3.4 Reglas duras

- **Nunca** commitees código sin sus tests asociados
- **Nunca** marques un step de feature-list como `[done]` si los tests asociados no pasan de forma contundente
- Un test que pasa con assertion genérica (ej: `assert.True(true)`) o sin verificar comportamiento real **es un falso positivo y equivale a no tener test**
- Antes de cada commit con código funcional: ejecutar `go test ./...` en el tool afectado y confirmar PASS

Detalle completo en [ADR-011](docs/adr-011-tdd-como-estandar.md).

---

## 4. Comandos esenciales

| Acción | Comando |
|---|---|
| Bootstrap entorno | `bash .agent/init.sh` |
| Build single tool | `go build -o tools/<name>/<name> ./tools/<name>` |
| Build all tools | `bash scripts/build.sh` |
| Test single tool | `cd tools/<name> && go test ./... -v` |
| Test all | `go test ./tools/...` |
| Lint repo | `./tools/linelens/linelens check --fail` |
| Install hooks | `lefthook install` |

---

## 5. Estructura del repo

```
open-harness/
├── tools/
│   ├── linelens/        ← v0.1.0 (file length linter)
│   ├── dupelens/        ← v0.1.0 (duplicate detector, Rabin-Karp)
│   └── secretlens/      ← v0.1.0 (secret/credential detector)
├── npm/@open-harness/   ← wrappers npm por plataforma
├── docs/                ← ADRs (decisiones arquitectónicas)
├── scripts/             ← build.sh, build-npm.sh
├── .agent/              ← harness: feature-list, progress, init
├── go.work              ← Go workspace
├── lefthook.yml         ← git hooks
├── linelens.json        ← config del propio repo
└── open-harness.json    ← registry de tools
```

---

## 6. Cómo agregar una nueva feature

1. **Lee** [`.agent/feature-list.json`](.agent/feature-list.json) para entender qué está planeado
2. **Identifica** la feature ID (F-XXX). Si no existe, agrégala antes de codear
3. **Crea branch**: `epic/<nombre>` para épicas, `feat/<id>-descripción` para features acotadas
4. **Sigue TDD** por cada step de la feature (sección 3)
5. **Mantén archivos ≤ 100 líneas** — si un archivo crece, parte la responsabilidad
6. **Documenta decisiones no obvias** en un ADR nuevo (`docs/adr-NNN-titulo.md`)
7. **Actualiza el feature-list** marcando steps como `[done]` solo cuando tests pasen
8. **Commit atómicos** con mensaje claro: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
9. **No pushees a main directo** — abre PR desde tu branch

---

## 7. Cómo agregar un nuevo tool al monorepo

1. `mkdir tools/<name>` con `go.mod` (module path: `github.com/jassencastillo/open-harness/tools/<name>`)
2. Agregar al `go.work` con `use ./tools/<name>`
3. Registrar en `open-harness.json`
4. Agregar entradas al `.gitignore` para los binarios (`<name>` y `<name>.exe`)
5. Implementar **siguiendo TDD** desde el primer commit
6. Cumplir filosofía: un solo binario, cero deps, regla de 100 líneas
7. Crear ADR si la implementación toma decisiones no obvias

---

## 8. Convenciones de código Go

- Solo stdlib, **prohibido** agregar dependencias externas sin ADR justificándolo
- `package main` en cada tool (CLI standalone)
- Exports PascalCase, locales camelCase
- Errores: `fmt.Errorf("context: %w", err)` para envolver con contexto
- Errores fatales: retornar `error`, no `panic` salvo en `main`
- CLI con `flag` package estándar — no `cobra`, no `urfave/cli`
- Comentarios en español o inglés, identifiers en inglés

---

## 9. Convenciones de tests

- Archivos `*_test.go` co-localizados con el código bajo test
- `func TestXxx_describe(t *testing.T)` con nombres descriptivos
- Table-driven cuando hay varios casos similares
- Helpers de test van en `testhelpers_test.go` para no contaminar la build
- Cada función pública debe tener al menos un test
- **Cobertura mínima objetivo**: 70% por package — preferir tests significativos sobre cobertura cosmética

---

## 10. Boundaries

### Always do
- TDD para cada cambio de comportamiento (sección 3)
- `go test ./...` en el tool tocado antes de commitear
- `linelens check --fail` antes de cada commit (lefthook lo hace automático)
- Build con `bash scripts/build.sh` para verificar que todos los tools compilan

### Ask first
- Antes de agregar dependencias externas a cualquier tool
- Antes de modificar la estructura de `go.work`
- Antes de agregar nuevos tools (revisar `feature-list.json` para roadmap)
- Antes de marcar `[done]` un step que tenga tests débiles

### Never do
- Commitear binarios compilados (excluidos en `.gitignore`)
- Saltar fallos de `linelens check` en CI
- Agregar dependencias a `tools/` sin revisar impacto en goal de stdlib
- Push directo a `main` (todo va por PR)
- Skipear hooks con `--no-verify` salvo emergencia justificada
- Crear archivos `>100 líneas` sin justificación documentada
