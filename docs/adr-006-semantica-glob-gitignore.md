# ADR-006: Semántica de patrones glob estilo .gitignore

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

Al implementar el matching de patrones surgió la pregunta: ¿`*.go` debe coincidir solo con archivos en el directorio raíz, o con cualquier archivo `.go` en cualquier subdirectorio?

Comportamientos posibles:
1. **Shell glob:** `*.go` solo coincide en el directorio actual → `src/main.go` NO coincide
2. **gitignore:** `*.go` (sin `/`) coincide con cualquier archivo llamado `*.go` → `src/main.go` SÍ coincide
3. **ESLint/tooling moderno:** igual que gitignore

## Decisión

Se adoptó la **semántica gitignore**:

- Patrón **sin `/`** → se aplica al nombre del archivo en cualquier directorio
  - `*.go` → coincide con `main.go`, `src/main.go`, `pkg/util/math.go`
- Patrón **con `/`** → se aplica al path relativo completo
  - `src/*.go` → coincide con `src/main.go` pero NO con `pkg/main.go`
- Patrón con `**` → cualquier nivel de subdirectorios
  - `**/*.spec.ts` → coincide con `foo.spec.ts`, `src/foo.spec.ts`, `a/b/c.spec.ts`

## Razonamiento

Los usuarios que configuran reglas como `{ "pattern": "*.go", "maxLines": 200 }` esperan que aplique a TODOS los archivos Go del proyecto, no solo a los de la raíz. La semántica gitignore es la más intuitiva para este caso de uso.

El comportamiento se descubrió durante los tests: un test inicial esperaba que `*.go` NO coincidiera con `src/main.go`. Al revisar la intención real del usuario, se corrigió el test para reflejar la semántica gitignore.

## Consecuencias

- **Positivo:** comportamiento intuitivo para la mayoría de los casos de uso.
- **Positivo:** consistente con `.gitignore`, `.eslintignore`, y otros archivos de configuración del ecosistema.
- **Negativo:** puede sorprender a usuarios que esperan semántica shell.
- **Documentado en:** `matcher.go` y los tests en `matcher_test.go`.
