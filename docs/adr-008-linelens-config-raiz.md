# ADR-008: configs root para los tools del monorepo

**Estado:** Aceptado
**Fecha:** 2026-05-02
**Última actualización:** 2026-05-02 (generalizado tras release de dupelens v0.1.0 — el patrón ahora aplica a múltiples tools)

## Contexto

Al reorganizar el proyecto en monorepo, cada tool tiene su propio archivo de config dentro de su directorio (ej: `tools/linelens/linelens.json`). Sin embargo, al ejecutar el tool desde la raíz del repositorio (como lo hace lefthook), busca el archivo de configuración en el directorio de trabajo actual.

Sin un config en la raíz, cada tool cae en su configuración por defecto, que puede no respetar las excepciones del proyecto. Ejemplos:

- `linelens` sin config root reportaba `tools/linelens/scanner_test.go` (132 líneas) como violación, porque los archivos `*_test.go` deberían permitir hasta 300.
- `dupelens` sin config root usaría `minTokens=50` (default), pero el repo tiene duplicación inter-tool intencional (~166 tokens en matcher.go) que dispararía el `--fail` en cada commit.

## Decisión

Se coloca **un archivo de config en la raíz por cada tool que el repo use sobre sí mismo**:

| Archivo root | Tool | Comentario |
|---|---|---|
| `linelens.json` | linelens | maxLines 100 por default, 300 para tests, 500 para `.agent/feature-list.json` |
| `dupelens.json` | dupelens | minTokens 200 (threshold pragmático mientras F-007 deduplica helpers) |
| `secretlens.json` | secretlens | excluye `**/*_test.go` y `**/testdata/**` (los tests de secretlens contienen secretos falsos como fixtures by design) |

Las configs específicas de cada tool (`tools/<name>/<name>.json`) se mantienen como **plantilla / ejemplo** para usuarios externos que adoptan el tool, pero la **configuración raíz es la fuente de verdad** para los hooks y CI del propio open-harness.

## Consecuencias

**Positivo:**

- Una sola ejecución de `<tool> check` desde la raíz cubre todo el monorepo con las reglas correctas.
- Los hooks de lefthook son simples; no necesitan conocer la estructura interna de `tools/`.
- El patrón es uniforme: todo tool nuevo que se incorpore al monorepo creará su `<tool>.json` root si necesita reglas distintas al default.

**Negativo:**

- Existe duplicación entre `linelens.json` raíz y `tools/linelens/linelens.json` (lo mismo aplica a dupelens). Si se actualiza una, hay que actualizar la otra.
- Esta duplicación es **acknowledged** y de bajo costo: las configs son cortas (~30 líneas) y cambian raramente.

**Decisión futura:**

- Cuando haya más de un tool con configuraciones que necesiten compartir reglas (ej: ambos saltan `node_modules`), evaluar si conviene un `open-harness.json` de top-level con reglas comunes y configs por tool que extiendan. Por ahora cada config root es independiente.
