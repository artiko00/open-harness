# ADR-001: Go sobre Node.js como lenguaje del CLI

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

Se necesita una herramienta CLI que:
- Funcione en Linux, macOS y Windows sin instalar runtimes adicionales
- Se ejecute rápido dentro de un git hook (pre-commit), donde cada milisegundo afecta la experiencia del desarrollador
- Sea fácil de distribuir: el usuario descarga un archivo y listo

Node.js era la alternativa natural dado que Husky (el sistema de hooks) ya vive en ese ecosistema.

## Decisión

Se eligió **Go**.

| Criterio | Node.js | Go |
|---|---|---|
| Startup (cold) | ~150ms | ~5ms |
| Distribución | requiere runtime | binario único |
| Compilación cruzada | no nativa | `GOOS/GOARCH` trivial |
| Integración Husky | `.js` nativo | wrapper de una línea |

La diferencia de startup importa: un pre-commit hook que tarda 150ms se siente lento cuando el desarrollador hace commits frecuentes. Go produce un binario que arranca en ~5ms.

## Consecuencias

- **Positivo:** binario único por plataforma, cero dependencias de runtime, startup ~30x más rápido que Node.
- **Positivo:** compilación cruzada con un solo comando (`GOOS=linux go build`).
- **Negativo:** requiere Go instalado para compilar (no para ejecutar).
- **Negativo:** los desarrolladores JavaScript no pueden contribuir fácilmente al código fuente.
- **Neutral:** proyectos consumidores pueden integrar linelens con Husky usando el wrapper npm (`@open_harness/linelens`). El propio repositorio open-harness usa lefthook en lugar de Husky — esa decisión está documentada en ADR-007.
