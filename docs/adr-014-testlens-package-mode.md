## ADR-014: testlens — package mode vs file mode

**Estado:** Aceptado
**Fecha:** 2026-05-02

## Contexto

testlens detecta archivos fuente sin un test asociado. La heurística inicial mapea cada archivo fuente a un test con un patrón derivado del lenguaje:

- Python: `foo.py` → `test_foo.py` o `foo_test.py`
- TypeScript: `foo.ts` → `foo.test.ts` o `foo.spec.ts`
- Java: `Foo.java` → `FooTest.java`

Este modelo *file-based* funciona bien cuando el ecosistema asocia tests por archivo. Pero rompe en Go: el toolchain ejecuta tests **por paquete (directorio)**, no por archivo. Un solo `*_test.go` puede cubrir todos los `.go` del paquete (incluso múltiples archivos fuente) porque todos comparten el mismo `package <name>`.

Aplicar la regla file-based a Go genera **falsos positivos masivos**: en este repo testlens reportaba 21 violaciones contra `tools/`, todas de archivos como `binary.go`, `scanner.go`, `merge.go` que **sí están cubiertos** por `main_test.go`, `coverage_extra_test.go`, etc. — pero con nombres distintos.

El resultado: testlens no podía usarse para auto-validar el repo en `lefthook` porque el ruido ahogaba la señal.

## Decisión

**Agregar un atributo `packageBased bool` al `languageMapping` y derivar dos modos de chequeo:**

- **File mode** (default, `packageBased: false`): cada archivo fuente requiere un test con nombre derivado en el mismo directorio. Aplica a Python, TypeScript, JavaScript, Ruby, Rust, Java, Kotlin, C#.
- **Package mode** (`packageBased: true`): un directorio que contenga **al menos un** archivo de test (matching `lang.testSuffixes` + `lang.extensions`) cubre **todos** los archivos fuente del directorio. Aplica a Go.

`checkCoverage` decide el modo una vez al inicio y delega:

```go
if lang.packageBased {
    return checkCoveragePackage(cfg, lang)
}
```

`checkCoveragePackage` recolecta directorios con archivos fuente y reporta una violación por **directorio sin tests**, no por archivo.

Adicionalmente: `testdata/` se agrega al skip list por defecto, alineándose con la convención del propio toolchain Go (`go build` ignora `testdata/` automáticamente).

## Consecuencias

**Positivo:**

- testlens deja de generar falsos positivos en repos Go con tests por paquete
- El propio repo open-harness ahora puede auto-protegerse con `testlens check --lang go --dir tools/ --fail` en `lefthook` pre-commit
- El criterio `packageBased` es declarativo: agregar un nuevo lenguaje package-based es flippear un bool

**Negativo:**

- Granularidad menor para Go: si un paquete tiene 10 archivos y el test cubre solo 2 funcionalidades, testlens no lo detecta. Eso es un trabajo para herramientas de cobertura real (`go test -cover`), no para testlens
- Dos rutas de código en `coverage.go` (`checkCoverage` per-file + `checkCoveragePackage`) — duplicación controlada del walk pattern, justificable porque la lógica de detección es genuinamente distinta

**Neutral:**

- Rust queda como **file mode** por ahora aunque también soporta tests inline (`#[cfg(test)]`). En la práctica los tests inline no son detectables sin parsear, y la convención `*_test.rs` o `tests/` external es la más común. Si en el futuro un proyecto Rust real reporta falsos positivos, se reevalúa
- ADR-011 (cobertura 100% como estándar) se mantiene: el modo package se cubre con tests dedicados (`package_mode_test.go`), no rebaja la barra

## Criterio de elegibilidad para `packageBased: true`

Un lenguaje es package-based si **el toolchain estándar ejecuta tests agrupados por directorio y el archivo de test no necesita match de nombre con el archivo fuente**. Bajo este criterio:

- ✅ Go: `go test ./pkg/foo/` corre todos los `*_test.go` del directorio
- ❌ Python: `pytest test_foo.py` ejecuta tests específicos; el match por nombre es la convención
- ❌ Java: cada `Foo.java` espera un `FooTest.java` por convención JUnit
- 🟡 Rust: ambiguo — cargo test corre por crate (similar a paquete) pero la comunidad usa también el patrón file-based; default file mode hasta tener señal contraria
