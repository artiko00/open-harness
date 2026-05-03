# ADR-009: open-harness se protege con sus propias herramientas

**Estado:** Aceptado
**Fecha:** 2026-05-02
**Última actualización:** 2026-05-02 (extendido a múltiples tools tras release de dupelens v0.1.0)

## Contexto

open-harness desarrolla herramientas de calidad de código. El proyecto puede usar esas herramientas sobre sí mismo, creando una relación de auto-aplicación. Esto fue introducido inicialmente con linelens (v0.1.0, ADR-005) y se extendió a múltiples tools cuando llegó dupelens.

## Decisión

El hook `pre-commit` de lefthook ejecuta cada tool de calidad sobre todo el repositorio antes de cada commit:

| Tool | Comando |
|---|---|
| linelens | `tools/linelens/linelens check --fail --no-color` |
| dupelens | `tools/dupelens/dupelens check --fail --no-color` |

(secretlens también es candidato a esta lista — pendiente de decisión sobre si correr en cada commit o solo en CI por su costo de regex multi-pattern.)

Cualquier violación de cualquiera de estos tools bloquea el commit. Esto implica que **cada contribución al proyecto debe cumplir las mismas reglas que las herramientas imponen a los proyectos que las adoptan**.

## Consecuencias

**Positivo:**

- Credibilidad: las herramientas demuestran en su propio código que las reglas que proponen son viables y no arbitrarias.
- Detecta regresiones inmediatamente: si un refactor hace crecer un archivo más allá del límite, o introduce código duplicado, el commit falla antes del push.
- Funciona como caso de prueba real y continuo del comportamiento de cada tool en un proyecto Go activo.
- Cuando se incorpora un tool nuevo al monorepo, agregarlo al `pre-commit` es trivial — la infraestructura ya existe.

**Negativo:**

- Cualquier cambio legítimo que supere los thresholds requiere un refactor previo o un ajuste explícito en el config root del tool. Esto es intencional: fuerza calidad antes de comodidad.
- El `dupelens.json` root del repo usa `minTokens=200` en vez del default 50, porque actualmente hay duplicación inter-tool real (matcher.go ~166 tokens, binary.go ~87 tokens entre los 3 tools). Esa deuda está acknowledged en F-007 ("extraer helpers compartidos a `tools/_shared/`"). Cuando F-007 cierre, bajaremos el threshold y el detector se vuelve útil para uso real.

**Invariantes:**

- Los archivos `*_test.go` usan la regla `maxLines: 300` en `linelens.json` (los tests naturalmente son más verbosos). Excepción documentada.
- `.agent/feature-list.json` usa `maxLines: 500` en `linelens.json` (es un registry que crece con features, no código). Excepción documentada.
- Los `*_test.go` y `migrations/**` se saltan en `dupelens.json` (test boilerplate y migraciones suelen tener patrones repetidos por diseño).
