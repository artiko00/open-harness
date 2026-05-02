# ADR-007: lefthook como gestor de git hooks

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

Se necesita un sistema de git hooks que:
- Bloquee commits que rompan las reglas de calidad del proyecto (linelens)
- Bloquee pushes con tests fallidos
- Funcione sin instalar runtimes adicionales en el entorno de CI/CD
- Sea fácil de activar para cualquier colaborador con un solo comando

Las alternativas evaluadas fueron Husky, pre-commit y git hooks nativos.

## Decisión

Se eligió **lefthook**.

| Criterio | Husky | pre-commit | Git hooks nativos | lefthook |
|---|---|---|---|---|
| Runtime requerido | Node.js | Python | ninguno | ninguno |
| Config compartida en repo | sí | sí | no (`.git/` no se commitea) | sí |
| Ejecución paralela | no | no | manual | nativa |
| Velocidad | lenta | media | depende | muy rápida |
| Multi-lenguaje | no | sí | sí | sí |

Husky fue descartado porque requiere Node.js como runtime, lo que añade una dependencia externa en un proyecto Go. pre-commit fue descartado por requerir Python. Los hooks nativos no son portables entre colaboradores (`.git/hooks/` no se incluye en el repositorio).

lefthook es un binario único sin dependencias, se instala con `brew install lefthook` o descargando el binario, y su configuración vive en `lefthook.yml` dentro del repositorio.

## Consecuencias

- **Positivo:** activación en un comando (`lefthook install`); compatible con cualquier OS.
- **Positivo:** ejecución paralela nativa reduce la latencia del pre-push cuando hay múltiples tools.
- **Positivo:** consistente con la filosofía del proyecto: binarios sin dependencias de runtime.
- **Negativo:** es una herramienta adicional que cada colaborador debe instalar manualmente (no se instala automáticamente como Husky con `npm install`).
- **Mitigación:** documentar `lefthook install` como paso obligatorio en el onboarding del repo.
