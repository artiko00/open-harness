# ADR-002: Cero dependencias externas

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

El matching de patrones glob con soporte `**` (doble asterisco) no está en la librería estándar de Go. La librería `github.com/bmatcuk/doublestar/v4` resuelve esto de forma robusta.

Alternativas evaluadas:
1. Usar `doublestar/v4` para glob completo
2. Implementar matching manual limitado
3. Usar solo `path/filepath.Match` (sin `**`)

## Decisión

Se optó por **cero dependencias externas** e implementación manual del matching `**`.

La razón principal: esta herramienta está diseñada para ejecutarse como parte del toolchain de cualquier proyecto. Añadir una dependencia externa significa que `go.sum` crece, `go mod download` es necesario antes del primer build, y existe riesgo de supply chain (aunque mínimo).

El caso de uso real de `**` en este proyecto es siempre `**/pattern` (prefijo `**` más sufijo). Este patrón se cubre con una implementación de ~20 líneas en `matcher.go`.

## Consecuencias

- **Positivo:** `go build` funciona sin `go mod download`, incluso offline.
- **Positivo:** `go.sum` no existe (no hay nada que verificar), auditoría de dependencias trivial.
- **Positivo:** el binario no arrastra código de terceros.
- **Negativo:** el glob engine no soporta todos los casos de `doublestar` (ej: `a/**/b/**/c`).
- **Negativo:** cualquier caso de uso avanzado requiere código nuevo en `matcher.go`.
