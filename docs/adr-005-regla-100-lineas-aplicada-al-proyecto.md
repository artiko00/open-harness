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

Cada archivo quedó bajo 100 líneas. El comando `linelens check` ejecutado sobre el propio proyecto retorna `OK: all 11 files within limits`.

## Razonamiento

Una herramienta que impone restricciones pero no las cumple envía una señal contradictoria. Si los mantenedores del proyecto no siguen la regla, es difícil argumentar que otros deberían seguirla.

Además, el ejercicio de refactorizar el propio código para cumplir la regla validó que:
1. El umbral de 100 líneas es alcanzable sin sacrificar legibilidad
2. La extracción forzada resultó en una mejor separación de responsabilidades

## Consecuencias

- **Positivo:** credibilidad: la herramienta practica lo que predica.
- **Positivo:** el refactor descubrió que `binary.go` y `help.go` son cohesivos por sí solos.
- **Negativo:** el número de archivos aumentó de 9 a 11 para cumplir el umbral.
- **Regla en `linelens.json`:** los archivos `*_test.go` tienen límite 300 líneas, reconociendo que los tests legítimamente requieren más líneas que el código de producción.
