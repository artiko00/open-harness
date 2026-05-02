# ADR-003: JSON como formato del archivo de configuración

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

La herramienta necesita un archivo de configuración para definir:
- Límite de líneas por defecto
- Reglas específicas por patrón de archivo
- Directorios y patrones excluidos

Formatos evaluados: **JSON**, YAML, TOML, HCL.

## Decisión

Se eligió **JSON** (`linelens.json`).

| Criterio | JSON | YAML | TOML |
|---|---|---|---|
| Dependencia externa | ninguna | requiere librería | requiere librería |
| Familiaridad en proyectos web | alta | media | baja |
| Soporte en editores (validación, autocompletado) | amplio | amplio | limitado |
| Comentarios | no | sí | sí |
| Verbosidad | media | baja | media |

JSON es el único formato que Go parsea con la librería estándar (`encoding/json`), lo que mantiene el objetivo de cero dependencias externas. Además es el formato de configuración más reconocido en proyectos JavaScript/TypeScript donde esta herramienta se usará frecuentemente.

La ausencia de comentarios en JSON es la principal desventaja, pero se compensa generando un archivo de ejemplo bien estructurado con `linelens init`.

## Consecuencias

- **Positivo:** cero dependencias adicionales para parsear el config.
- **Positivo:** familiaridad inmediata para desarrolladores de cualquier stack.
- **Positivo:** validación y autocompletado disponibles si se publica un JSON Schema.
- **Negativo:** no soporta comentarios, lo que dificulta documentar el porqué de cada regla.
- **Futuro posible:** agregar soporte YAML/TOML como opt-in si hay demanda, sin romper JSON.
