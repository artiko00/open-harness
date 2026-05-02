# ADR-005: El proyecto se aplica su propia regla de 100 líneas

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

Durante el desarrollo, `scanner.go` alcanzó 125 líneas y `main.go` 106. La propia herramienta violaba la regla que impone a otros proyectos.

Se planteó la pregunta: ¿debe el proyecto ser excepción de su propia regla, o debe cumplirla?

## Decisión

El proyecto **cumple su propia regla**. Se refactorizó extrayendo responsabilidades:

- `scanner.go` → la detección de binarios se movió a `binary.go`
- `main.go` → el texto de ayuda se movió a `help.go`

Cada archivo quedó bajo 100 líneas. Al ejecutarse sobre sí mismo, linelens reporta todos los archivos dentro del límite (`OK: all files within limits`).

## Razonamiento

Una herramienta que impone restricciones pero no las cumple envía una señal contradictoria. Si los mantenedores del proyecto no siguen la regla, es difícil argumentar que otros deberían seguirla.

Además, el ejercicio de refactorizar el propio código para cumplir la regla validó que:
1. El umbral de 100 líneas es alcanzable sin sacrificar legibilidad
2. La extracción forzada resultó en una mejor separación de responsabilidades

## Consecuencias

- **Positivo:** credibilidad: la herramienta practica lo que predica.
- **Positivo:** el refactor descubrió que `binary.go` y `help.go` son cohesivos por sí solos.
- **Negativo:** el número de archivos fuente de `tools/linelens` aumentó de 7 a 9 para cumplir el umbral (se extrajeron `binary.go` y `help.go`).
- **Regla en `linelens.json`:** los archivos `*_test.go` tienen límite 300 líneas, reconociendo que los tests legítimamente requieren más líneas que el código de producción.
- **Actualización (monorepo):** al migrar a open-harness, la regla se extiende a todo el repositorio vía `linelens.json` en la raíz (ADR-008) y se aplica automáticamente en cada `git commit` mediante lefthook (ADR-007 y ADR-009).
