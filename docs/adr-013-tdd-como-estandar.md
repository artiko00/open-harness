## ADR-011: TDD como estándar del proyecto

**Estado:** Aceptado
**Fecha:** 2026-05-02

## Contexto

Hasta ahora el proyecto tenía tests, pero no un workflow obligatorio. Algunas funciones (`binary.go`, `config.go`, `reporter.go`, `help.go` en `linelens`) llegaron a producción sin tests propios. La cobertura era heterogénea y los tests existentes se escribieron *después* del código, no antes.

El problema con tests escritos a posteriori:

- Tienden a ser **tests del código que se escribió**, no del **comportamiento que se necesitaba**
- Pasan por construcción (escribís el test sabiendo cómo funciona el código → verifica lo que ya hiciste)
- No funcionan como **especificación ejecutable** ni como red de seguridad para refactor
- Generan falsos positivos: assertion genérica que pasa pero no verifica nada real

Esto choca con la filosofía del repo: cero dependencias, archivos de 100 líneas, auto-protección — todas decisiones que apuntan a confiabilidad. Tener tests débiles rompe el espíritu.

## Decisión

**Adoptar TDD (Test-Driven Development) como workflow obligatorio para todo cambio de comportamiento.**

Ciclo Red → Green → Refactor descrito en `AGENTS.md` sección 2:

1. **Red**: escribir test que falle por la razón correcta
2. **Green**: el mínimo código para que pase
3. **Refactor**: limpiar con tests en verde como red de seguridad

Reglas duras:

- Toda feature/cambio: test primero, código después
- Nunca commitear código sin tests asociados
- Tests con assertions genéricas (`assert.True(true)`, etc.) cuentan como **no tener test**
- Una feature solo se marca `[done]` en `feature-list.json` si sus tests pasan de forma contundente
- Cobertura objetivo mínima por package: 70% (no como métrica vanidosa, sino como signo de tests significativos)

Excepciones permitidas — sin TDD pero documentadas:

- Cambios de **solo documentación** (`docs:`, `chore: docs`)
- Cambios de **solo configuración** sin lógica (`chore:`)
- **Refactor puro** que no cambia comportamiento — pero los tests existentes deben seguir pasando como red

## Consecuencias

**Positivo:**

- Tests funcionan como especificación: leer `*_test.go` revela el contrato de la API
- Confianza para refactorizar — si los tests pasan, el comportamiento se mantiene
- Cobertura sube de forma natural sin perseguir el número
- Los falsos positivos (tests que pasan sin verificar nada) son evitados estructuralmente porque el test se escribió cuando NO existía la implementación

**Negativo:**

- Velocidad inicial de desarrollo aparenta bajar — escribir tests primero se siente más lento
- Curva de adopción: agentes y humanos acostumbrados a "código primero" tienen que reentrenar el reflejo
- Algunos casos (UI experimental, prototipos throwaway) son menos naturales para TDD — esos escenarios no aplican aquí

**Neutral:**

- El `.agent/init.sh` valida que todos los tools compilen y sus tests pasen al inicializar el entorno — refuerza el contrato pero no obliga TDD del lado del agente, solo verifica resultado
- `lefthook.yml` corre `go test` en `pre-push` — captura el caso de "código sin test" antes de que llegue al remote
- Esta decisión aplica al monorepo open-harness; cada tool nuevo (incluido `dupelens` en curso) la hereda

## Aplicación inmediata

- `AGENTS.md` documenta el workflow TDD como **obligatorio**
- `dupelens` (épica F-006 en curso) se implementa siguiendo TDD desde la fase 1 (core Rabin-Karp): cada step debe empezar con un test que falle
- `linelens` adopta TDD para cualquier cambio futuro; los archivos sin tests actuales (`binary.go`, `config.go`, `reporter.go`, `help.go`) se cubren cuando se modifiquen
