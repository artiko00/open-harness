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

## 2. Workflow obligatorio: TDD (Test-Driven Development)

**Toda feature nueva o cambio de comportamiento sigue el ciclo Red → Green → Refactor.** Sin excepciones.

### 2.1 Red (test que falla primero)

1. Antes de escribir código de producción, **escribe el test**
2. El test debe **fallar** al correrlo: `go test ./...` produce `FAIL`
3. El test debe ser específico, no genérico — verifica un comportamiento concreto, no "que algo no crashee"
4. Solo cuando el test falla por la razón esperada (no por error de compilación tonto), continúas

### 2.2 Green (mínimo código para pasar)

1. Escribe **lo mínimo** para que el test pase
2. No optimices, no anticipes futuras features, no agregues "while we're here"
3. Re-corre `go test ./...` hasta ver `PASS`
4. Si el test pasa "por accidente" o el comportamiento real es distinto a lo esperado, **mejora el test** antes de continuar

### 2.3 Refactor (limpiar con red de seguridad)

1. Con tests en verde, refactoriza para legibilidad y SOLID
2. Re-corre tests después de cada cambio significativo
3. Si introduces nuevo comportamiento durante el refactor, **vuelve al Red** — agrega un test antes

### 2.4 Reglas duras

- **Nunca** commitees código sin sus tests asociados
- **Nunca** marques un step de feature-list como `[done]` si los tests asociados no pasan de forma contundente
- Un test que pasa con assertion genérica (ej: `assert.True(true)`) o sin verificar comportamiento real **es un falso positivo y equivale a no tener test**
- Antes de cada commit con código funcional: ejecutar `go test ./...` en el tool afectado y confirmar PASS

Detalle completo en [ADR-011](docs/adr-011-tdd-como-estandar.md).

---

## 3. Comandos esenciales

### Bootstrap del entorno
```bash
bash .agent/init.sh
```
Hace: `go work sync` + `go build` + `go test` por cada tool. Falla rápido si algo está roto.

### Test de un tool específico
```bash
cd tools/<nombre>
go test ./... -v
```

### Build de todos los tools
```bash
bash scripts/build.sh
```

### Validar regla de líneas (auto-protección)
```bash
./tools/linelens/linelens check --fail
```

### Hooks de git (lefthook)
```bash
lefthook install   # primera vez
```
- `pre-commit`: corre `linelens check --fail`
- `pre-push`: corre `go test ./...` por cada tool en paralelo

---

## 4. Estructura del repo

```
open-harness/
├── tools/
│   ├── linelens/        ← v0.1.0 (file length linter)
│   └── dupelens/        ← v0.1.0-scaffold (duplicate detector)
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

## 5. Cómo agregar una nueva feature

1. **Lee** [`.agent/feature-list.json`](.agent/feature-list.json) para entender qué está planeado
2. **Identifica** la feature ID (F-XXX). Si no existe, agrégala antes de codear
3. **Crea branch**: `epic/<nombre>` para épicas, `feat/<id>-descripción` para features acotadas
4. **Sigue TDD** por cada step de la feature (sección 2)
5. **Mantén archivos ≤ 100 líneas** — si un archivo crece, parte la responsabilidad
6. **Documenta decisiones no obvias** en un ADR nuevo (`docs/adr-NNN-titulo.md`)
7. **Actualiza el feature-list** marcando steps como `[done]` solo cuando tests pasen
8. **Commit atómicos** con mensaje claro: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
9. **No pushees a main directo** — abre PR desde tu branch

---

## 6. Cómo agregar un nuevo tool al monorepo

1. `mkdir tools/<name>` con `go.mod` (module path: `github.com/jassencastillo/open-harness/tools/<name>`)
2. Agregar al `go.work` con `use ./tools/<name>`
3. Registrar en `open-harness.json`
4. Agregar entradas al `.gitignore` para los binarios (`<name>` y `<name>.exe`)
5. Implementar **siguiendo TDD** desde el primer commit
6. Cumplir filosofía: un solo binario, cero deps, regla de 100 líneas
7. Crear ADR si la implementación toma decisiones no obvias

---

## 7. Convenciones de código Go

- Solo stdlib, **prohibido** agregar dependencias externas sin ADR justificándolo
- `package main` en cada tool (CLI standalone)
- Nombres en inglés para identificadores; comentarios pueden estar en español o inglés
- Errores: retornar `error`, no `panic` salvo en `main`
- CLI con `flag` package estándar — no `cobra`, no `urfave/cli`

---

## 8. Convenciones de tests

- Archivos `*_test.go` en el mismo package que el código bajo test
- Función `TestXxx(t *testing.T)` con nombres descriptivos: `TestTokenize_skipsPunctuation` no `TestTokenize1`
- Table-driven cuando hay varios casos similares
- Helpers de test van en `testhelpers_test.go` para no contaminar la build
- Cada función pública debe tener al menos un test
- **Cobertura mínima objetivo**: 70% por package — preferir tests significativos sobre cobertura cosmética

---

## 9. Antes de cualquier commit

Checklist obligatorio:
- [ ] `go test ./...` pasa en el tool tocado (sin skips, sin warnings ignorados)
- [ ] `linelens check --fail` pasa sobre el repo
- [ ] El commit no incluye binarios compilados
- [ ] El commit no incluye co-author de IA salvo que el usuario lo pida explícitamente
- [ ] Si tocaste decisiones arquitectónicas, hay un ADR nuevo o actualizado
- [ ] Si avanzaste en una feature, actualizaste `[done]` en `feature-list.json`

---

## 10. Cosas que NO se hacen

- ❌ Commitear sin tests
- ❌ Saltar el ciclo TDD "porque es trivial"
- ❌ Agregar dependencias externas sin ADR
- ❌ Push directo a `main` (todo va por PR)
- ❌ Skipear hooks con `--no-verify` salvo emergencia justificada
- ❌ Marcar steps `[done]` con tests débiles o inexistentes
- ❌ Crear archivos `>100 líneas` sin justificación documentada en una excepción del config
