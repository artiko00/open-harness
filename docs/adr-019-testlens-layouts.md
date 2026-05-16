# ADR-019: testlens — layouts no co-ubicados

**Fecha:** 2026-05-13
**Estado:** Aceptada
**Aplica a:** testlens

## Contexto

testlens v0.2.1 implementa `testExists` así:

```go
sourceDir := filepath.Dir(sourcePath)
for _, candidate := range candidates {
    if fileExists(filepath.Join(sourceDir, candidate)) { return candidate }
}
```

Esto solo encuentra tests **co-ubicados** (en el mismo directorio que el source). Funciona para Go (convención canónica) y para proyectos JS/TS que adoptan el patrón "test al lado del source", pero genera **falsos positivos** en los layouts más comunes del ecosistema:

| Layout | Ejemplo |
|---|---|
| `__tests__/` subdir | `src/foo.ts` ↔ `src/__tests__/foo.test.ts` (Jest/Vitest default) |
| `tests/` paralelo top-level | `src/foo.py` ↔ `tests/test_foo.py` (pytest típico) |
| Maven mirror | `src/main/java/com/X/Bar.java` ↔ `src/test/java/com/X/BarTest.java` |
| `spec/` mirror | `lib/foo.rb` ↔ `spec/foo_spec.rb` (Rails) |

Un usuario que adopta testlens con un proyecto Vitest típico ve decenas de "no test found" cuando en realidad tiene cobertura completa — solo que en `__tests__/`.

## Decisión

Extender `languageMapping` con dos campos nuevos:

```go
type languageMapping struct {
    extensions   []string
    testSuffixes []string
    testPrefixes []string
    testDirs     []string         // ej: ["__tests__", "tests"]
    mirrors      [][2]string      // ej: [["src/main", "src/test"]]
    packageBased bool
}
```

### `testDirs` — subdirs adjacentes al source

Para cada `testDir` en la lista, testlens busca el candidate también en `<source-dir>/<testDir>/`.

Ejemplo: con `testDirs: ["__tests__"]`, source `src/auth/user.ts`, candidate `user.test.ts`:
- Probar `src/auth/user.test.ts` (co-ubicado, comportamiento existente)
- Probar `src/auth/__tests__/user.test.ts` (nuevo)

### `mirrors` — espejos de prefijo de path

Para cada par `[from, to]`, testlens probar el candidate reemplazando el prefijo `from/` por `to/` en el path completo del source.

Ejemplo Maven: `mirrors: [["src/main/java", "src/test/java"]]`, source `src/main/java/com/foo/Bar.java`, candidate `BarTest.java`:
- Probar `src/main/java/com/foo/BarTest.java` (co-ubicado, no encontrado)
- Probar `src/test/java/com/foo/BarTest.java` (mirrored, encontrado ✓)

Ejemplo Python: `mirrors: [["src", "tests"]]`, source `src/auth/user.py`, candidate `test_user.py`:
- Probar `src/auth/test_user.py`
- Probar `tests/auth/test_user.py` (mirrored)

### Configuración por lenguaje

| Lenguaje | testDirs | mirrors |
|---|---|---|
| Go | `[]` | `[]` |
| TypeScript | `["__tests__", "tests"]` | `[]` |
| JavaScript | `["__tests__", "tests"]` | `[]` |
| Python | `["tests"]` | `[["src", "tests"]]` |
| Ruby | `["spec"]` | `[["lib", "spec"]]` |
| Rust | `["tests"]` | `[]` |
| Java | `[]` | `[["src/main/java", "src/test/java"]]` |
| Kotlin | `[]` | `[["src/main/kotlin", "src/test/kotlin"]]` |
| C# | `[]` | `[]` |

Go queda sin `testDirs`/`mirrors` porque la convención canónica del lenguaje es co-ubicado.

## Alternativas descartadas

### Índice global por basename

"Indexar todos los tests por basename y matchear sin restricción de path". Cubre todo, pero produce falsos positivos cross-módulo: `src/admin/user.ts` se cuenta como cubierto por `src/auth/user.test.ts`.

### Templates de path configurables por usuario

Exponer `testDirs`/`mirrors` como config del usuario en JSON/TOML. **Postergado**: los defaults por lenguaje cubren el 95% de proyectos. Si aparece demanda real para un layout exótico, se agrega como F-016.

## Consecuencias

**Positivas:**
- Cubre los 5 layouts más comunes del ecosistema sin requerir config del usuario.
- Predecible: ningún match cross-módulo, ningún falso positivo nuevo.
- No-breaking: proyectos que hoy funcionan siguen funcionando idénticos.

**Negativas:**
- ~30 líneas extra en `matcher.go` para `applyMirror` + la búsqueda extendida.
- Mantenibilidad: cuando agreguemos un lenguaje nuevo hay que pensar su `testDirs`/`mirrors`. Mitigado por una tabla de defaults sensatos.

**Neutras:**
- **testlens no se transforma en coverage tool**. Sigue siendo un check estático "¿existe un archivo de test asociado?". Para métricas funcionales (líneas ejercidas, branches cubiertas) hay que usar Vitest/Jest/pytest-cov/coverage.py. El README de cada paquete documenta esta distinción explícitamente.

## Comparación con coverage tools reales

| Criterio | testlens | vitest --coverage / jest / pytest-cov |
|---|---|---|
| Costo | milisegundos (file walk) | segundos a minutos (compila + ejecuta + instrumenta) |
| Detecta archivos sin test | ✅ | parcial (sólo si se cargan en algún test) |
| Detecta líneas no ejecutadas | ❌ (out of scope) | ✅ |
| Detecta branches no testeadas | ❌ | ✅ |
| Funciona en pre-commit hook | ✅ | demasiado lento |
| Funciona en CI gate | ✅ | sí, pero como step separado más lento |

**testlens es un smoke test ultra-rápido de existencia de tests**, complementario a un coverage tool real. Usar ambos: testlens en pre-commit (instantáneo) + coverage tool en CI (mide ejecución real).
