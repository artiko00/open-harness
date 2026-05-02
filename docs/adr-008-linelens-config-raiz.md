# ADR-008: linelens.json en la raíz del monorepo

**Estado:** Aceptado  
**Fecha:** 2026-05-02

## Contexto

Al reorganizar el proyecto en monorepo, cada tool tiene su propio `linelens.json` dentro de su directorio (`tools/linelens/linelens.json`). Sin embargo, al ejecutar `linelens check` desde la raíz del repositorio (como lo hace lefthook), linelens busca el archivo de configuración en el directorio de trabajo actual.

Sin un `linelens.json` en la raíz, linelens cae en su configuración por defecto: 100 líneas máximo sin ninguna regla adicional. Esto causaba que `tools/linelens/scanner_test.go` (132 líneas) reportara una violación falsa, ya que los archivos `*_test.go` deberían permitir hasta 300 líneas.

## Decisión

Se colocó un **`linelens.json` en la raíz del monorepo** que define las reglas aplicables a todo el proyecto.

La configuración de cada tool individual (`tools/linelens/linelens.json`) se mantiene para cuando se ejecuta linelens desde dentro del directorio del tool, pero la configuración raíz es la fuente de verdad para los hooks y CI.

## Consecuencias

- **Positivo:** una sola ejecución de `linelens check` desde la raíz cubre todo el monorepo con las reglas correctas.
- **Positivo:** los hooks de lefthook son simples; no necesitan conocer la estructura interna de `tools/`.
- **Negativo:** existe duplicación entre `linelens.json` raíz y `tools/linelens/linelens.json`. Si se actualiza una, hay que actualizar la otra.
- **Decisión futura:** cuando haya más de un tool con configuraciones distintas, evaluar si conviene una sola config raíz o un flag `--config` por tool en los hooks.
