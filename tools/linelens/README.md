# linelens

Linter de tamaño de archivos para cualquier lenguaje. Mide **líneas de código**
(no líneas físicas) y la **profundidad de anidamiento**, para no penalizar
catálogos i18n ni cabeceras de licencia y sí detectar archivos cortos pero
densos.

## Qué cuenta

linelens sólo analiza archivos con **extensión de código**
(`pathmatch.CodeExtensions`: `.go`, `.ts`, `.py`, `.rs`, `.java`, ...). Los
datos (`i18n.json`, `schema.sql`, ...) y los archivos generados
(`*.pb.go`, `*_gen.go`, `*.g.dart`, `*-lock.json`) se ignoran por defecto; la
lista de exclusión es configurable.

Por cada archivo reporta dos métricas:

- **líneas de código** (`lines`): se elimina el contenido de comentarios y
  strings con `langsyntax.StripComments` y se cuentan las líneas **no vacías**
  del resultado. Comentarios y líneas en blanco no suman.
- **líneas físicas** (`physical`): el total de líneas del archivo, sólo
  informativo.

### líneas de código vs `wc -l`

| entrada                     | `wc -l` | físico (linelens) | código (linelens) |
|-----------------------------|:-------:|:-----------------:|:-----------------:|
| `a\nb\n`                    |    2    |         2         |         2         |
| `a\nb` (sin salto final)    |    1    |         2         |         2         |
| 380 comentarios + 2 código  |   382   |        382        |         2         |

`wc -l` cuenta **saltos de línea**: una última línea sin `\n` final no la
cuenta. El **físico** de linelens cuenta esa última línea igual (como un
editor), así que difiere de `wc -l` exactamente en la última línea sin salto
final. El **código** ignora comentarios y blancos, por eso una cabecera de
licencia de 380 líneas sobre 2 de código cuenta como 2.

## Anidamiento

`maxNesting` mide la **profundidad máxima de bloques** por balance de llaves
`{ }` sobre el stream ya despojado de comentarios/strings. **No es complejidad
ciclomática**: sólo cuán hondo llegan los bloques anidados. Con `0`
(por defecto) la métrica queda desactivada.

```json
{
  "default": { "maxLines": 100, "maxNesting": 5 }
}
```

Un archivo de 83 líneas con 80 `if` anidados viola con `maxNesting: 5`; un
archivo plano largo (miles de líneas, poca anidación) no.

## Uso

```bash
linelens check                      # escanea el directorio actual
linelens check --fail               # exit 1 si hay violaciones (git hooks)
linelens check --max 200 --dir ./src
linelens check --format json
linelens init                       # crea linelens.json por defecto
```

El contrato de salida (`OK:` / `SUMMARY:`, sección `SKIPPED`, `--format json`,
`--config` estricto y merge por campo) es el de las fases 2–4.
