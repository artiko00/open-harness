# Fix audit findings across the four lenses

Feature ID: **F-018** (`.agent/feature-list.json`)
Affected tools: **linelens · dupelens · secretlens · testlens** (los cuatro)
Risk: **medium**

## Why

Una auditoría adversarial (dos agentes independientes + verificación manual, commit `59a1371`)
reprodujo 18 defectos en los cuatro tools. El problema de fondo no es la precisión del análisis:
es que **los gates fallan en verde**. Un usuario recibe `exit 0` y un mensaje `OK:` en escenarios
donde la herramienta no analizó nada.

Los cuatro casos que anulan el gate por completo:

- `testlens check` sin `--lang` reporta el 100% de falsos positivos y **elige el lenguaje al azar**
  (itera un `map` de Go): sobre este mismo repo alterna entre 8 y 68 violaciones entre corridas idénticas.
- `testlens --lang <typo>` devuelve `exit 0` para siempre; un `touch x_test.go` vacío marca cubierto
  un paquete con `panic()` sin testear.
- `secretlens` no detecta **ningún** secreto en un `.env` sin comillas, y un comentario `# example`
  al final de la línea suprime el hallazgo (la allowlist se aplica a la línea completa, no al match).
- Los cuatro se cuelgan indefinidamente ante un FIFO en el árbol y descartan en silencio cualquier
  archivo con una línea > 1 MB o con error de lectura.

Una quality gate que no puede distinguir "analicé y está limpio" de "no analicé nada" produce falsa
seguridad, que es peor que no tener la herramienta. Por eso ahora.

## What Changes

### Corrección de gates que fallan en verde

- **BREAKING** `secretlens`: los `patterns` custom se **suman** a los 8 built-in en vez de reemplazarlos;
  opt-out explícito vía `disableDefaultPatterns`.
- **BREAKING** `secretlens`: la allowlist se evalúa contra el **valor detectado**, no contra la línea;
  se quita `example` del default.
- **BREAKING** `testlens`: `--lang` con un valor no soportado es error (`exit 1`), no un pase.
- `testlens`: el modo `auto` devuelve el `languageMapping` real (con `testSuffixes`, `testDirs`,
  `mirrors`, `packageBased`) y ordena la detección por conteo descendente sobre un recorrido determinista.
- `testlens`: la detección respeta `exclude` (hoy recorre `node_modules` y deja que las dependencias
  de terceros decidan el lenguaje del proyecto).
- `testlens`: un archivo de test cuenta como cobertura solo si contiene al menos una función de test.
- Los cuatro: los archivos saltados (FIFO/no regulares, línea > 1 MB, error de lectura, binarios)
  se reportan como `SKIPPED` y cuentan para el exit code, en vez de desaparecer.

### Unificación transversal

- Nuevo módulo `tools/_shared/pathmatch/` con `matchGlob`/`isExcluded`/`isBinary*`, hoy replicados
  byte a byte en tres tools y reimplementados peor en el cuarto.
- `testlens` pasa a usar ese motor: su `exclude` actual usa `strings.Contains` sobre el path completo,
  de modo que el default `"out"` excluye silenciosamente `src/routes/` y `layouts/`.
- `**/dir/**` pasa a matchear a cualquier profundidad (hoy solo un nivel, y es el patrón que shipean
  `linelens init`, `dupelens init`, el README y los `.json` del repo).
- `--format json` en los cuatro; `--format` inválido es error.
- `--config <ruta>` explícita e inexistente es error, no fallback silencioso a defaults.
- Claves de config desconocidas producen error (hoy `testlens.json` del repo declara `skip` y
  `languages`, que no existen: el archivo entero es inerte).
- Formato de reporte homogéneo (`TOOL (N …)` + `SUMMARY:`), diagnóstico siempre a stderr.

### Calidad del análisis

- `secretlens`: regla `KEY=VALUE` sin comillas obligatorias, entropía de Shannon sobre el valor,
  y prefijos de proveedor (`sk_live_`, `xoxb-`, `AIza`, `sk-proj-`, `glpat-`, `npm_`, `SG.`,
  URIs con credenciales). Recall medido hoy: 25%.
- `dupelens`: normalización de identificadores para detectar clones Type-2 (copy-paste con renombre),
  etiquetados `exact` / `renamed`.
- `dupelens`: eliminar `Fingerprint.Window []string` — hoy consume ~100× el tamaño del fuente
  (2.370 MB con 24 MB de código), lo que produce OOM en runners de CI.
- `dupelens`: separar `windowSize` (sensibilidad) de `minTokens` (reporte) y filtrar por extensión
  de código, para no competir con fixtures JSON, i18n y lockfiles.
- `linelens`: contar líneas de código (sin comentarios ni blancos) reutilizando el stripper de dupelens.

### Consistencia de metadatos

- Versión con fuente única de verdad: hoy diverge en 5 lugares (binarios 0.2.1/0.2.5,
  README 0.2.0 y 0.1.0, `open-harness.json` 0.1.0).
- README/AGENTS: Dart sin documentar, `--config`/`--no-color`/`--output` sin documentar,
  "3 tools" en pre-commit donde `lefthook.yml` tiene 4.
- ADR-018 describe un merge de config por campo que no existe: se implementa o se corrige el ADR.

## Capabilities

### New Capabilities

Ninguna capability existe todavía en `openspec/specs/` (el directorio está vacío), de modo que
todas las siguientes se crean en este change:

| Capability | Cubre |
|---|---|
| `path-matching` | glob por segmento, `**` a cualquier profundidad, exclusión, detección de binarios |
| `file-traversal` | solo archivos regulares, líneas largas, errores de lectura ruidosos, symlinks |
| `cli-contract` | flags comunes, `--format json`, validación de flags, exit codes, formato de reporte |
| `config-loading` | cadena de configs, claves desconocidas, `--config` inexistente, precedencia |
| `secret-detection` | reglas, entropía, allowlist por match, merge de patterns built-in |
| `duplicate-detection` | clones exact/renamed, techo de memoria, filtro por lenguaje |
| `test-coverage-detection` | detección de lenguaje determinista, mapeo test↔fuente, verificación de contenido |
| `line-metrics` | conteo de líneas de código vs líneas de archivo |
| `release-metadata` | versión con fuente única, sincronía docs ↔ binarios |

### Modified Capabilities

Ninguna (no hay specs previas que modificar).

## Scope

### In Scope

- Los 18 hallazgos de la auditoría, en los cuatro tools.
- El módulo compartido `tools/_shared/pathmatch/` y la migración de los cuatro tools a él.
- `--format json` para linelens, secretlens y testlens (dupelens ya lo tiene).
- Corrección de `testlens.json`, `open-harness.json`, README, AGENTS.md y ADR-018.
- ADRs nuevos para: módulo compartido de path matching, y entropía en secretlens.

### Out of Scope

- Paralelismo en el walk de los cuatro tools (medido: irrelevante salvo en dupelens, donde el
  cuello es memoria y no CPU).
- Salida SARIF / integración con GitHub Code Scanning (se habilita al tener `--format json`,
  pero se difiere).
- Resolución de imports en testlens para atribuir tests de integración a sus módulos.
- Winnowing en dupelens (el fix de memoria no lo requiere).
- Un quinto tool (`bigo`, F-001) y los canales de distribución PyPI/Packagist (F-012/F-013).

## Impact

| Área | Impacto | Detalle |
|---|---|---|
| `tools/_shared/pathmatch/` | New | Módulo nuevo con su propio `go.mod`; precedente: `tools/_shared/tomlmin` |
| `tools/testlens/{detect,coverage,matcher,scanner,config,language}.go` | Modified | El grueso del riesgo: cambia qué se reporta |
| `tools/testlens/reporter.go`, `scanner.go:20-58` | Removed | ~100 líneas de producción muertas cubiertas al 100% por tests |
| `tools/secretlens/{patterns,engine,config}.go` | Modified | Recall, allowlist y merge de patterns |
| `tools/dupelens/{fingerprint,scanner,strip}.go` | Modified | Memoria y clones Type-2 |
| `tools/linelens/{scanner,matcher}.go` | Modified | Conteo de líneas de código |
| `tools/*/matcher.go`, `binary.go`, `config_chain.go` | Removed | 3 copias byte-idénticas absorbidas por `pathmatch` |
| `testlens.json`, `open-harness.json`, `dupelens.json` | Modified | Configs rotas o con umbrales que ocultan hallazgos |
| `README.md`, `AGENTS.md`, `docs/adr-018` | Modified | Deriva documental y un ADR que describe algo inexistente |
| `docs/adr-020`, `docs/adr-021` | New | Módulo compartido; entropía en secretlens |
| Dependencias externas | Ninguna | Todo con stdlib (ADR-002 intacto) |

**Nota sobre `dupelens.json`**: hoy fija `minTokens: 200` (default 50). Con el default real, el propio
repo reporta 38 duplicados en `tools/`. Bajar el umbral es parte del change, y solo es viable después
de absorber las copias en `pathmatch`.

## Risks

| Riesgo | Prob. | Mitigación |
|---|---|---|
| Los cambios en testlens alteran qué se reporta y rompen el pre-commit del propio repo | Alta | Fase de testlens al final; correr `lefthook run pre-commit` tras cada fase |
| Entropía en secretlens introduce falsos positivos nuevos | Media | Umbral calibrado contra el fixture de auditoría; medir recall/precisión antes y después |
| Clones Type-2 en dupelens generan ruido en boilerplate | Media | Etiquetar `renamed` aparte de `exact`; que `--fail` cuente solo `exact` por defecto |
| Las 3 fases BREAKING rompen configs de usuarios existentes | Media | Documentar en README + nota de migración; el opt-out `disableDefaultPatterns` preserva el comportamiento viejo |
| El módulo compartido rompe el aislamiento de binarios (ADR-002) | Baja | `pathmatch` es stdlib-only y se compila estáticamente, igual que `tomlmin` |
| Regresión de cobertura por debajo del 100% (ADR-011) | Media | TDD estricto (ADR-013); `go tool cover -func` como gate por fase |

## Rollback Plan

Cada fase es un commit atómico e independiente, en este orden: `pathmatch` → `file-traversal` →
`cli-contract` → `config-loading` → `secretlens` → `dupelens` → `linelens` → `testlens` → metadatos.

- Revertir una fase concreta: `git revert <sha>` — ninguna fase posterior depende de otra salvo
  que las cuatro de tools dependen de `pathmatch`.
- Revertir todo: `git revert` del rango, o `git reset --hard 59a1371` si aún no se publicó.
- Los binarios publicados en npm no se tocan hasta que las 9 fases estén verdes; el bump de versión
  es la última tarea, de modo que un rollback previo al release no afecta a ningún usuario.
- Los fixtures de reproducción de la auditoría quedan versionados como `testdata/` en cada tool,
  así que una regresión futura se detecta con `go test`, no con una auditoría manual.

## Success Criteria

- [ ] Los 18 hallazgos tienen un test que falla antes del fix y pasa después.
- [ ] `testlens check` sin flags produce el mismo resultado en 20 corridas consecutivas.
- [ ] `testlens check` sin flags reporta 0 falsos positivos en fixtures de Go, TypeScript, Python, Ruby, Rust y Dart.
- [ ] `secretlens` detecta ≥ 90% de los secretos del fixture de auditoría (hoy: 25%).
- [ ] Ningún tool reporta `OK:` habiendo saltado archivos: los saltos se listan y afectan el exit code.
- [ ] Ningún tool se cuelga con un FIFO en el árbol (verificado con `timeout 6`).
- [ ] `dupelens` procesa 24 MB de fuente bajo 512 MB de RSS (hoy: 2.370 MB).
- [ ] `dupelens check --dir tools --min-tokens 50` reporta 0 duplicados tras absorber las copias.
- [ ] Los cuatro tools aceptan `--format json` y rechazan `--format <inválido>` con exit 1.
- [ ] `--config <ruta inexistente>` explícita devuelve exit 1 en los cuatro.
- [ ] La versión coincide en `main.go`, `open-harness.json`, README y AGENTS.md, verificado por script.
- [ ] Cobertura de statements sigue en 100% en los cuatro tools (ADR-011).
- [ ] `lefthook run pre-commit` y `lefthook run pre-push` pasan sobre el repo con los umbrales corregidos.
