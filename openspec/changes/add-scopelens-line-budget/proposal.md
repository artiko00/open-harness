# scopelens: presupuesto de líneas además de archivos

Feature ID: **F-021** (`.agent/feature-list.json`)
Affected tools: **scopelens**
Risk: **low** (aditivo; el presupuesto de líneas viene deshabilitado por defecto)

## Why

`scopelens` mide el alcance de un PR por cantidad de archivos, pero un cambio de **un solo archivo
con 11.000 líneas** pasa el gate de archivos y sin embargo es una superficie de review enorme. Falta
poder acotar también las líneas del diff, y dejar que cada equipo elija cómo combinar ambos límites.

## What Changes

- **`maxLines`** (nuevo, default `0` = deshabilitado): presupuesto de líneas cambiadas del diff
  acumulado (rama vs base + staged), de los archivos contables (excluye lockfiles/generados; los
  tests cuentan salvo `--exclude-tests`).
- **`mode`** (`"or"` default | `"and"`): cómo se combinan los presupuestos cuando **ambos** están
  activos. `or` falla si se excede cualquiera; `and` falla sólo si se exceden ambos. Con un solo
  presupuesto activo, `mode` es irrelevante.
- **`lineMetric`** (`"changed"` default = agregadas + borradas | `"added"` = sólo agregadas): qué
  cuenta como línea. Lo decide quien configura.
- Flags equivalentes: `--max-lines`, `--mode`, `--line-metric`.
- El reporte muestra **ambos** conteos (archivos y líneas) y su límite; el estado `OK`/`FAIL` refleja
  la combinación elegida.

## Capabilities

### Modified Capabilities

- `change-scope`: agrega el presupuesto de líneas y la combinación configurable de presupuestos.

## Scope

### In Scope
- `maxLines`, `mode`, `lineMetric` en config y como flags; parseo de `git diff --numstat`; el gate
  combinado; el reporte con ambas métricas; validación (valores inválidos → exit 2); 100% coverage.

### Out of Scope
- Líneas de código excluyendo comentarios/blancos (se cuenta el diff crudo; una refinación futura).
- Presupuestos por categoría (source vs test) separados.

## Impact

| Área | Impacto | Detalle |
|---|---|---|
| `tools/scopelens/config.go` | Modified | `MaxLines`, `Mode`, `LineMetric` + validación |
| `tools/scopelens/scanner.go` | Modified | numstat por archivo (churn) además del name-only |
| `tools/scopelens/report.go` | Modified | conteo de líneas contables + gate combinado |
| `tools/scopelens/reporter.go` | Modified | estado y SUMMARY con archivos y líneas |
| `tools/scopelens/check_cmd.go` | Modified | flags `--max-lines`/`--mode`/`--line-metric` |
| Dependencias / coverage | Sin cambios | stdlib + `git`; 100% (ADR-011) |

## Rollback Plan

Aditivo: revertir el commit. Con `maxLines: 0` (default) el comportamiento es idéntico al actual —
sólo gate de archivos.

## Success Criteria

- [ ] `maxLines` cuenta las líneas cambiadas de los archivos contables (add+del o added según `lineMetric`).
- [ ] `mode` combina los presupuestos (`or`/`and`) cuando ambos están activos.
- [ ] Valores inválidos de `maxLines`/`mode`/`lineMetric` → exit 2 nombrando la causa.
- [ ] El reporte muestra archivos y líneas; con `maxLines: 0` el gate es sólo por archivos (retrocompat).
- [ ] 100% coverage; scopelens pasa su propio gate.
