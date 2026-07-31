# dupelens: los imports dejan de contar como código duplicado

Feature ID: **F-022** (`.agent/feature-list.json`)
Affected tools: **dupelens** (vía `tools/_shared/langsyntax`)
Risk: **medium** (cambia el conteo de matches de instalaciones existentes; el gate por defecto no se debilita)

## Why

Reporte de un usuario sobre un monorepo NestJS: 26 matches en 77 archivos, **todos `renamed`, ninguno
`exact`**. La causa está en el tokenizador.

`tokenize()` (`tools/dupelens/tokenizer.go:24`) descarta comentarios y el contenido de los strings,
pero conserva las declaraciones de import. Con los strings ya vacíos, una cabecera como

```ts
import { UserService } from './user/user.service';
import { OrderRepo }   from './order/order.repo';
```

queda reducida a `import from import from`, y tras `normalizeTokens()` a `import ID from import ID
from` — **idéntica en todos los archivos del proyecto**. Una arquitectura modular abre cada archivo
con 5–15 imports, así que con `windowSize: 25` cualquier par de cabeceras colisiona. El resultado es
que cuanto mejor modularizado está el proyecto, más ruido produce la herramienta.

Los imports son la misma clase de cosa que los comentarios y los strings que el tokenizador ya
descarta: sintaxis obligatoria del lenguaje que no expresa lógica. Es una omisión del stripper.

Dos problemas secundarios que el mismo reporte expone:

- El encabezado dice `DUPLICATES (26 match(es) found in 77 files)` sin desglosar el tipo, que solo
  aparece en cada línea suelta. Con `--fail` en verde (el gate por defecto solo cuenta `exact`), la
  salida se lee como alarmante y hay que revisar los 26 renglones para descubrir que ninguno rompe
  nada.
- Los bloques de datos embebidos en código (arrays de seed, tablas literales) siguen produciendo
  matches `renamed` que ninguna regla de imports cubre: cada línea es estructuralmente igual a la
  anterior.

## What Changes

- **`default.ignoreImports`** (nuevo, default **`true`**): las declaraciones de import, include y
  re-export se descartan antes de tokenizar. Se apaga con `"ignoreImports": false`.
- **Reconocimiento por familia de lenguaje**, sin parser, sobre las 25 extensiones que ya cubre
  `pathmatch.CodeExtensions()`: JS/TS, Python, Go, Ruby, Rust, JVM, PHP, C/C++/ObjC, C#, Dart, Swift.
  Soporta declaraciones multilínea (`import {\n A,\n} from …`, `import ( … )` de Go, `from x import
  ( … )` de Python) por balance de delimitadores.
- **Desglose por tipo en el reporte**: el encabezado y el `SUMMARY` de consola muestran `N exact ·
  M renamed`; el JSON gana `exactCount` y `renamedCount` (aditivo, no rompe el schema).
- **Filtro de baja entropía** en la pasada `renamed`: se descarta una ventana cuando una proporción
  alta de sus líneas empieza por el mismo token. Ataca los bloques de datos. **No se aplica a la
  pasada `exact`**: allí la igualdad literal ya es señal genuina (un `switch` copiado byte a byte es
  un hallazgo que se quiere ver), y como `--fail` por defecto solo mira `exact`, el gate no pierde
  poder de detección en ningún caso.

## Capabilities

### Modified Capabilities

- `duplicate-detection`: el alcance del análisis excluye declaraciones de import; el reporte
  desglosa hallazgos por tipo; la pasada renamed descarta bloques de baja entropía.

## Scope

### In Scope
- `langsyntax.StripImports(src, ext)` compartido, aplicado después de `StripComments`.
- Clave `default.ignoreImports` con default `true`, en `init` y en `--tutorial`.
- Contadores por tipo en consola y JSON.
- Filtro de baja entropía sobre tokens crudos, aplicado solo a la pasada renamed.
- Tests TDD por lenguaje y por caso límite; 100% statement coverage (ADR-011).
- README, CHANGELOG y `docs/UPGRADING.md`.

### Out of Scope
- Flag CLI `--ignore-imports`: la clave de config alcanza, igual que `windowSize`, que tampoco tiene
  flag equivalente.
- Umbral de entropía configurable (YAGNI; el filtro se documenta con su umbral fijo, como
  `monotoneWindow`).
- Descartar declaraciones de tipo, anotaciones o boilerplate de framework (decoradores de NestJS,
  `@Injectable()`): expresan intención, no son sintaxis obligatoria de acceso a otro módulo.
- Lenguajes fuera de `pathmatch.CodeExtensions()` (CSS, Vue, Svelte): hoy no se escanean.

## Impact

| Área | Impacto | Detalle |
|---|---|---|
| `tools/_shared/langsyntax/imports.go` | New | `StripImports` + continuación multilínea |
| `tools/_shared/langsyntax/importsyntax.go` | New | familias por extensión y sus prefijos |
| `tools/dupelens/tokenizer.go` | Modified | `tokenize` recibe si debe descartar imports |
| `tools/dupelens/config.go` | Modified | `IgnoreImports *bool`, default `true` |
| `tools/dupelens/collect.go`, `scanner.go` | Modified | propagar la opción y los tokens crudos |
| `tools/dupelens/entropy.go` | New | filtro de baja entropía de la pasada renamed |
| `tools/dupelens/normalize.go` | Modified | la pasada normalizada aplica el filtro |
| `tools/dupelens/classify.go` | Modified | `countByKind` para el desglose |
| `tools/dupelens/reporter_console.go`, `reporter_json.go` | Modified | desglose por tipo |
| `tools/dupelens/init_cmd.go`, `tutorial.go` | Modified | documentar la clave nueva |
| Dependencias / coverage | Sin cambios | stdlib (ADR-002); 100% (ADR-011) |

## Rollback Plan

Revertir el commit. Para un usuario que ya actualizó y quiere el comportamiento anterior sin
degradar: `"default": { "ignoreImports": false }` restituye el conteo de imports; el desglose por
tipo y el filtro de entropía no tienen escape por config y solo eliminan hallazgos de la pasada
renamed.

## Success Criteria

- [ ] Dos archivos cuya única coincidencia es la cabecera de imports no producen ningún match.
- [ ] Un duplicado real de lógica se sigue reportando con `ignoreImports: true`.
- [ ] `"ignoreImports": false` restituye el conteo de tokens de import.
- [ ] Los imports multilínea de JS/TS, Go y Python se descartan completos, no solo su primera línea.
- [ ] La numeración de líneas de los matches no se corre al descartar imports.
- [ ] El encabezado y el `SUMMARY` de consola muestran el desglose `exact` / `renamed`.
- [ ] El JSON incluye `exactCount` y `renamedCount`.
- [ ] Un bloque de datos repetitivo no produce match renamed; un `switch` copiado literal sigue
      reportándose como `exact`.
- [ ] 100% coverage; dupelens pasa su propio gate sobre el repo.
