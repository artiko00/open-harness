# Tasks — update-dupelens-import-noise

## Fase 1 · langsyntax.StripImports (compartido)

- [ ] 1.1 (red) Test de familias por extensión: cada una de las 25 extensiones de
  `pathmatch.CodeExtensions()` resuelve a su familia; una extensión desconocida no resuelve a
  ninguna y el fuente pasa intacto.
- [ ] 1.2 (red) Tests de descarte de una línea por lenguaje: JS/TS (`import … from`, `export … from`,
  `const x = require(…)`), Python (`import x`, `from x import y`), Go (`package`, `import "x"`),
  Ruby (`require`, `require_relative`), Rust (`use`, `extern crate`), JVM (`package`, `import`),
  PHP (`use`, `namespace`, `require_once`), C (`#include`, `using namespace`), C# (`using`,
  `global using`), Dart (`import`, `export`, `part`, `library`), Swift (`import`).
- [ ] 1.3 (red) Tests de continuación multilínea: `import {\n … \n} from …`, `import ( … )` de Go,
  `from x import ( … )` de Python; y de que el balance se cierra correctamente.
- [ ] 1.4 (red) Tests de falsos positivos: `using (var s = …)` de C# no se descarta; un `import`
  dentro de un identificador (`importedValue`) no se descarta; `import('./x')` dinámico no se
  descarta.
- [ ] 1.5 (red) Test de preservación de líneas: la cantidad de `\n` del output iguala la del input.
- [ ] 1.6 (green) Implementar `tools/_shared/langsyntax/importsyntax.go`: mapa extensión → familia
  y tabla de prefijos y pares (prefijo + substring requerido) por familia.
- [ ] 1.7 (green) Implementar `tools/_shared/langsyntax/imports.go`: `StripImports(src, ext)` con
  recorrido por líneas, `startsWithWord` (boundary + descarte de `prefijo(`) y balance de
  delimitadores para la continuación. Documentar que se aplica **después** de `StripComments`.
- [ ] 1.8 (refactor) Verificar que ambos archivos quedan bajo 100 líneas y que `langsyntax` mantiene
  100% de coverage.

## Fase 2 · dupelens: config y tokenizador

- [ ] 2.1 (red) Test de config: `ignoreImports` ausente resuelve a `true`; `false` explícito se
  respeta; `true` explícito se respeta.
- [ ] 2.2 (green) `config.go`: `IgnoreImports *bool` en `DefaultConfig` y default en
  `applyConfigDefaults`; actualizar `defaultConfigJSON` con la clave y su valor.
- [ ] 2.3 (red) Test de `tokenize` con y sin descarte de imports sobre el mismo fuente.
- [ ] 2.4 (green) `tokenizer.go`: `tokenize(src, ext string, stripImports bool)`, aplicando
  `StripImports` después de `StripComments`; propagar por `collect.go` y `scanner.go`.
- [ ] 2.5 (green) Actualizar los call sites de `tokenize` en los tests existentes.
- [ ] 2.6 (red/green) Test E2E: dos archivos `.ts` que solo comparten la cabecera de imports no
  producen match; con un bloque de lógica idéntico sí; la numeración de líneas del match es la del
  archivo original.

## Fase 3 · Desglose por tipo en el reporte

- [ ] 3.1 (red) Tests de consola: encabezado `DUPLICATES` y línea `SUMMARY` con `N exact · M renamed`.
- [ ] 3.2 (red) Test de JSON: `exactCount` y `renamedCount` en el objeto raíz.
- [ ] 3.3 (green) `classify.go`: `countByKind(matches)` reutilizando la lógica de `gateCount`.
- [ ] 3.4 (green) `reporter_console.go` y `reporter_json.go`: emitir el desglose.
- [ ] 3.5 (refactor) Ajustar los tests de contrato CLI que asertan el texto del encabezado o del
  resumen.

## Fase 4 · Filtro de baja entropía

- [ ] 4.1 (red) Tests unitarios de `lowEntropyWindow`: bloque de datos (todas las líneas con el
  mismo token inicial) → true; código real variado → false; ventana con menos líneas que el mínimo
  → false; ventana de una sola línea → false.
- [ ] 4.2 (green) `entropy.go`: `lowEntropyWindow(raw []Token, start, w int) bool` con umbral fijo
  documentado (mínimo de líneas y proporción de la moda).
- [ ] 4.3 (green) `normalize.go`: `fingerprintNormalized` recibe los tokens crudos y descarta las
  ventanas de baja entropía; `fingerprintCode` (pasada exact) queda intacta.
- [ ] 4.4 (red/green) Test E2E: dos arrays de datos análogos no producen match renamed; un `switch`
  copiado literal sigue reportándose como `exact`.

## Fase 5 · Documentación

- [ ] 5.1 `tutorial.go`: documentar `default.ignoreImports` y actualizar la sección de cambios.
- [ ] 5.2 `README.md`: sección dupelens — la clave nueva, el desglose por tipo y el filtro de
  entropía con su alcance (solo renamed).
- [ ] 5.3 `tools/dupelens/CHANGELOG.md` y `CHANGELOG.md` raíz.
- [ ] 5.4 `docs/UPGRADING.md`: nota de que el conteo de matches baja y cómo restituirlo.
- [ ] 5.5 Registrar F-022 en `.agent/feature-list.json` y actualizar `.agent/claude-progress.txt`.

## Fase 6 · Gates de calidad

- [ ] 6.1 `go test ./tools/...` en verde.
- [ ] 6.2 Coverage 100% en `langsyntax` y `dupelens` (ADR-011).
- [ ] 6.3 `linelens check --fail` sobre el repo.
- [ ] 6.4 `dupelens check --fail` sobre el repo.
- [ ] 6.5 `secretlens check --fail` y `testlens check --fail` sobre el repo.
