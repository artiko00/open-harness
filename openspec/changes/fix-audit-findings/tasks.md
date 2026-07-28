# Tasks: fix-audit-findings

Cada fase sigue rojo → verde → refactor. Ninguna tarea se marca `[done]` con tests en rojo o
cobertura por debajo del 100% (ADR-011). Los fixtures citados provienen de la auditoría y se
versionan como `testdata/` del tool correspondiente.

## 1. Fase 1 — pathmatch (módulo compartido)

- [x] 1.1 Crear `tools/_shared/pathmatch/go.mod` (patrón de `tomlmin`) y sumarlo a `go.work`
- [x] 1.2 [red] Test: `exclude` con `out` NO excluye `src/routes/` ni `layouts/`
- [x] 1.3 [red] Test: `**/migrations/**` matchea `db/migrations/2024/b.sql` (profundidad > 1)
- [x] 1.4 [red] Test: `src/*.ts` NO matchea `src/nested/b.ts` (el `*` simple no cruza `/`)
- [x] 1.5 [red] Test: un `.env` en UTF-16 NO se clasifica como binario; un ELF sí
- [x] 1.6 [green] Implementar `matchGlob`/`isExcluded` por segmento con `**` a cualquier profundidad
- [x] 1.7 [green] Implementar `isBinaryPath`/`isBinaryContent` con UTF-16 reconocido como texto
- [x] 1.8 [green] Exponer la lista compartida de extensiones de código (usada por linelens y dupelens)
- [x] 1.9 [refactor] Migrar linelens a `pathmatch` y borrar su `matcher.go`/`binary.go`
- [x] 1.10 [refactor] Migrar dupelens a `pathmatch` y borrar sus copias
- [x] 1.11 [refactor] Migrar secretlens a `pathmatch` y borrar sus copias
- [x] 1.12 [refactor] Migrar testlens a `pathmatch` y borrar `shouldSkipDir` (`strings.Contains`)
- [x] 1.13 Verificar 100% de cobertura en `pathmatch` y en los cuatro tools

## 2. Fase 2 — file-traversal (dejar de fallar en verde)

- [x] 2.1 [red] Test: `timeout 6 <tool> check` con un FIFO en el árbol termina sin exit 124, en los cuatro
- [x] 2.2 [red] Test: archivo sin permisos → aparece bajo `SKIPPED`, exit 1 con `--fail`, sin `OK:`
- [x] 2.3 [red] Test: archivo de 1,6 MB en una línea → `SKIPPED` con motivo de buffer, no desaparece
- [x] 2.4 [red] Test: un secreto dentro de un archivo omitido NO produce `OK: no secrets detected`
- [x] 2.5 [red] Test: binario y FIFO omitidos NO alteran el exit code con `--fail`
- [x] 2.6 [green] Filtrar por `d.Type().IsRegular()` en el walk de `pathmatch`
- [x] 2.7 [green] Introducir el estado `Skipped{Path, Reason}` en el resultado de escaneo
- [x] 2.8 [green] Subir el buffer de lectura a 4 MB y reportar el desborde como omisión
- [x] 2.9 [green] Propagar los omitidos al reporte y al exit code según la tabla de D2
- [x] 2.10 [red+green] testlens: la detección de lenguaje respeta `exclude` (fixture con `node_modules/*.rb`)

## 3. Fase 3 — cli-contract

- [x] 3.1 [red] Test: `--format json` produce JSON parseable en los cuatro tools
- [x] 3.2 [red] Test: `--format=xml` → stderr lista los válidos, exit 1
- [x] 3.3 [red] Test: `--max -5` y `--min-tokens 0` → error, exit 1 (hoy se ignoran)
- [x] 3.4 [red] Test: `--dir /does/not/exist` → mismo mensaje y exit 1 en los cuatro
- [x] 3.5 [red] Test: sin hallazgos, la última línea empieza con `OK:` en los cuatro
- [x] 3.6 [red] Test: con hallazgos, la última línea empieza con `SUMMARY:` en los cuatro
- [x] 3.7 [red] Test: el aviso de detección de testlens no aparece en stdout
- [x] 3.8 [green] Portar `ReportOpts{Format}` de dupelens a linelens, secretlens y testlens
- [x] 3.9 [green] Validar valores enumerados de flags con `fs.Visit` para distinguir default de explícito
- [x] 3.10 [green] Validar `--dir` antes del walk en los cuatro, con el mensaje de dupelens
- [x] 3.11 [green] Unificar encabezado + `SUMMARY:` + `OK:` en el reporte de los cuatro
- [x] 3.12 [green] Mover todo diagnóstico a stderr (`coverage.go:51` en testlens)
- [skip] 3.13 [refactor] Extraer el reporter común si las cuatro implementaciones convergen
      (skip: los reporters no convergieron — contenido de dominio propio por tool; la única
      duplicación real es la validación de flags ~8 líneas/tool, por debajo del umbral que
      justificaría un módulo compartido nuevo fuera del design)

## 4. Fase 4 — config-loading

- [x] 4.1 [red] Test: `--config /nope/nope.json` explícito → exit 1 en los cuatro
- [x] 4.2 [red] Test: ausencia del config por defecto sigue siendo válida (exit 0)
- [x] 4.3 [red] Test: `testlens.json` con `skip` → error nombrando la clave y sugiriendo `exclude`
- [x] 4.4 [red] Test: `linelens.json` con `excludes` → error nombrando la clave
- [x] 4.5 [red] Test: `pyproject.toml` aporta `maxLines` y `package.json` aporta `exclude` (merge por campo)
- [x] 4.6 [green] Distinguir `--config` explícito de default con `fs.Visit` y fallar si no existe
- [x] 4.7 [green] `DisallowUnknownFields` en los cuatro `loadConfig`, con mensaje accionable
- [x] 4.8 [green] Implementar el merge por campo en `config_chain.go` (arrays: gana el primero entero)
- [x] 4.9 [refactor] Separar en `config_merge.go` si algún tool supera las 100 líneas
- [x] 4.10 Corregir `testlens.json` de la raíz: `skip` → `exclude`, borrar `languages`

## 5. Fase 5 — secretlens (recall del 25% al 90%)

- [x] 5.1 Versionar el fixture de auditoría (20 secretos + señuelos) como `testdata/`
- [x] 5.2 [red] Test de tabla: recall ≥ 18/20 y precisión ≥ 80% sobre el fixture
- [x] 5.3 [red] Test: `API_KEY=valor` sin comillas se detecta; `DEBUG=true` no
- [x] 5.4 [red] Test: `sk_live_`, `xoxb-`, `AIza`, `sk-proj-`, `glpat-`, `npm_`, `SG.` se detectan
- [x] 5.5 [red] Test: `postgres://admin:pass@host/db` se detecta
- [x] 5.6 [red] Test: `AWS_KEY = "AKIA..."  # see example above` SÍ se reporta, exit 1
- [x] 5.7 [red] Test: `API_KEY=your_key_here` NO se reporta
- [x] 5.8 [red] Test: `postgres://u:SuperSecretPass99@example.com/db` SÍ se reporta
- [x] 5.9 [red] Test: un pattern custom NO desactiva los built-in; `disableDefaultPatterns` sí
- [x] 5.10 [red] Test: JSON pretty-printed con clave y valor en líneas consecutivas se detecta
- [x] 5.11 [green] Hacer opcionales las comillas en la regla genérica de asignación
- [x] 5.12 [green] Implementar entropía de Shannon sobre el valor capturado, con `minEntropy`
- [x] 5.13 [green] Agregar los patterns de proveedor y de URI con credenciales
- [x] 5.14 [green] Mover la evaluación del allowlist al valor capturado; quitar `example` del default
- [x] 5.15 [green] `append(defaultPatterns(), cfg.Patterns...)` + `disableDefaultPatterns`
- [x] 5.16 [green] Ventana de una línea para asignaciones partidas
- [x] 5.17 Calibrar `minEntropy` contra el fixture y registrar el valor elegido en ADR-021

## 6. Fase 6 — dupelens (memoria y clones renombrados)

- [x] 6.1 Versionar los fixtures de clones (`exact`, `renamed`, CRUD generado) como `testdata/`
- [x] 6.2 [red] Test largo: 24 MB de fuente bajo 512 MB de RSS (`runtime.ReadMemStats`)
- [x] 6.3 [red] Test: dos funciones idénticas con identificadores distintos se reportan como `renamed`
- [x] 6.4 [red] Test: copia literal se reporta como `exact`
- [x] 6.5 [red] Test: hallazgo `renamed` NO rompe `--fail` por defecto; con `--fail-on=all` sí
- [x] 6.6 [red] Test: colisión de hash con contenido distinto NO se reporta (verificación literal viva)
- [x] 6.7 [red] Test: `data.csv`/`fixtures.json`/lockfiles NO se reportan
- [x] 6.8 [red] Test: `#[derive(Debug)]` en Rust y `color: #fff` en CSS no se tratan como comentario
- [x] 6.9 [green] Reemplazar `Fingerprint.Window []string` por `{Hash, FileID, StartIdx}`
- [x] 6.10 [green] Verificación literal por índice sobre el slice de tokens del archivo
- [x] 6.11 [green] Segundo conjunto de fingerprints con identificadores normalizados a `ID`/`NUM`
- [x] 6.12 [green] Etiquetar hallazgos `exact`/`renamed` y agregar `--fail-on=exact|renamed|all`
- [x] 6.13 [green] Separar `windowSize` (detección) de `minTokens` (reporte)
- [x] 6.14 [green] Filtrar por extensiones de código usando la lista de `pathmatch`
- [x] 6.15 [green] Comentarios sensibles al lenguaje en `strip.go` (`#` no es comentario universal)
- [x] 6.16 [green] Snippet con cada fragmento bajo su propia ruta
- [x] 6.17 Bajar `dupelens.json` a `minTokens` default y verificar 0 duplicados en `tools/`

## 7. Fase 7 — linelens (líneas de código)

- [x] 7.1 Versionar los fixtures (licencia de 380 líneas, `i18n.json`, `nightmare.go`) como `testdata/`
- [x] 7.2 [red] Test: archivo de 382 líneas con 380 de cabecera de licencia NO viola con `maxLines` 100
- [x] 7.3 [red] Test: archivo con 250 líneas de código sí viola
- [x] 7.4 [red] Test: el reporte muestra líneas de código y total físico
- [x] 7.5 [red] Test: `i18n.json` y `schema.sql` NO se reportan (filtro de extensiones)
- [x] 7.6 [red] Test: archivo de 83 líneas con 80 anidamientos viola el umbral de anidamiento
- [x] 7.7 [green] Contar líneas de código reutilizando el stripper de comentarios compartido
- [x] 7.8 [green] Aplicar el filtro de extensiones de código y de archivos generados
- [x] 7.9 [green] Calcular profundidad máxima de anidamiento y su umbral configurable
- [x] 7.10 [green] Reportar ambas métricas en consola y en JSON
- [x] 7.11 Documentar en el README el criterio de conteo frente a `wc -l`

## 8. Fase 8 — testlens (el gate más roto)

- [x] 8.1 [refactor] Borrar `reporter.go` completo y `scanner.go:20-58` (producción muerta)
- [x] 8.2 Versionar los fixtures (Go, TS, Python, Ruby, Rust, Dart, monorepo) como `testdata/`
- [x] 8.3 [red] Test: 20 corridas de `testlens check` sobre el mismo árbol dan salida idéntica
- [x] 8.4 [red] Test: monorepo con 8 `.go` y 3 `.ts` detecta Go de forma estable
- [x] 8.5 [red] Test: proyecto Go de 3 archivos se detecta (hoy el umbral de >5 lo impide)
- [x] 8.6 [red] Test: `check` sin flags da el mismo resultado que `check --lang go`
- [x] 8.7 [red] Test: los seis fixtures de lenguaje dan 0 falsos positivos sin `--lang`
- [x] 8.8 [red] Test: `--lang typscript` → error, exit 1, sin afirmar cobertura
- [x] 8.9 [red] Test: `zzz_test.go` con solo `package svc` NO cubre el paquete, exit 1
- [x] 8.10 [red] Test: `tests/test_calc.py` no se reporta a sí mismo y sí cubre `app/calc.py`
- [x] 8.11 [red] Test: `__init__.py`, `settings.py`, `*_pb2.py` y migraciones NO se reportan
- [x] 8.12 [green] `detectLanguageFromFiles` con orden fijo, conteo máximo y sin umbral de 5
- [x] 8.13 [green] `getLanguageMapping` devuelve el mapping real en modo `auto`
- [x] 8.14 [green] Validar `--lang` con `supportedLanguages()` y fallar con la lista de soportados
- [x] 8.15 [green] Verificar contenido del candidato buscando el marcador de test del lenguaje
- [x] 8.16 [green] Reconocer el prefijo `test_` además de los sufijos en `matcher.go`
- [x] 8.17 [green] Exclusiones por defecto de archivos que no llevan test, configurables
- [x] 8.18 [green] Agregar el mirror `app/` ↔ `tests/` para Python

## 9. Fase 9 — metadatos, ADRs y documentación

- [x] 9.1 [green] Script `scripts/check-versions.sh` que compara `main.go` contra las demás fuentes
- [x] 9.2 Sincronizar `open-harness.json` con las versiones reales de cada `main.go`
- [x] 9.3 Sincronizar la tabla y el árbol de estructura del README, y AGENTS.md
- [x] 9.4 Documentar Dart en la tabla de lenguajes de testlens
- [x] 9.5 Documentar `--config`, `--no-color`, `--format` y `--output` en los cuatro tools
- [x] 9.6 Corregir "3 tools" → 4 en README y AGENTS.md (`lefthook.yml` tiene 4)
- [x] 9.7 Alinear los ejemplos de hooks del README con `lefthook.yml`
- [x] 9.8 Escribir ADR-020: módulo compartido `pathmatch`
- [x] 9.9 Escribir ADR-021: entropía como filtro en secretlens
- [x] 9.10 Extender ADR-018 con la semántica verificable del merge por campo
- [x] 9.11 Extender ADR-012 con la normalización de identificadores
- [x] 9.12 Renumerar `adr-014-testlens-package-mode.md` a ADR-022 (hoy colisiona con ADR-014)
- [x] 9.13 Registrar F-016 y F-017 en `.agent/feature-list.json` (citados en commits, ausentes del archivo)
- [x] 9.14 Nota de migración en el README para los tres cambios BREAKING (D5, D6, D12)

## 10. Gates de calidad

- [x] 10.1 `go test ./...` verde en los cinco módulos
- [x] 10.2 `go tool cover -func` al 100% de statements en los cuatro tools y en `pathmatch`
- [x] 10.3 `linelens check --fail` sobre el repo
- [x] 10.4 `dupelens check --fail` sobre el repo con `minTokens` default
- [x] 10.5 `secretlens check --fail` sobre el repo
- [x] 10.6 `testlens check --fail` sobre el repo
- [x] 10.7 `lefthook run pre-commit` y `lefthook run pre-push`
- [x] 10.8 `openspec validate fix-audit-findings --strict`
- [x] 10.9 Reproducir los 18 comandos de la auditoría y confirmar el comportamiento corregido
- [x] 10.10 `scripts/check-versions.sh` en verde y bump de versión de los cuatro tools
- [x] 10.11 Actualizar `.agent/claude-progress.txt`
