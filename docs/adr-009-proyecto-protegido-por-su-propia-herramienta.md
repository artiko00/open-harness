# ADR-009: open-harness se protege con su propia herramienta

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

linelens es una herramienta de calidad de código que detecta archivos que superan un límite de líneas. El proyecto que la desarrolla (open-harness) puede usar linelens sobre sí mismo, creando una relación de auto-aplicación.

Esta práctica fue introducida desde el proyecto standalone (`linelens` v0.1.0, ADR-005) y se mantiene en el monorepo.

## Decisión

El hook `pre-commit` de lefthook ejecuta `linelens check --fail` sobre todo el repositorio antes de cada commit. Cualquier archivo que supere su límite configurado bloquea el commit.

Esto implica que **cada contribución al proyecto debe cumplir las mismas reglas que linelens impone a los proyectos que lo adoptan**.

## Consecuencias

- **Positivo:** credibilidad: la herramienta demuestra en su propio código que las reglas que propone son viables y no arbitrarias.
- **Positivo:** detecta regresiones inmediatamente; si un refactor hace crecer un archivo más allá del límite, el commit falla antes del push.
- **Positivo:** sirve como caso de prueba real y continuo del comportamiento de linelens en un proyecto Go activo.
- **Negativo:** cualquier cambio legítimo que supere el límite (ej. añadir una feature grande en un archivo existente) requiere un refactor previo o un ajuste explícito en `linelens.json`. Esto es intencional: fuerza a mantener archivos pequeños.
- **Invariante:** el archivo `tools/linelens/scanner_test.go` usa la regla `*_test.go` (300 líneas) porque los tests naturalmente tienden a ser más verbosos. Esta excepción está documentada en `linelens.json`.
