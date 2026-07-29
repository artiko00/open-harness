# Tasks: add-scopelens-line-budget

Rojo → verde → refactor (ADR-013). 100% coverage (ADR-011), archivos < 100 líneas (ADR-005).

## 1. Config

- [x] 1.1 [red] Test: `maxLines`, `mode`, `lineMetric` se leen de scopelens.json con defaults (0/or/changed)
- [x] 1.2 [red] Test: `maxLines` negativo → error; `mode` ∉ {or,and} → error; `lineMetric` ∉ {changed,added} → error
- [x] 1.3 [green] `Config`: campos `MaxLines`, `Mode`, `LineMetric` + validación en validateAndDefault

## 2. Conteo de líneas (numstat)

- [x] 2.1 [red] Test: `parseNumstat` parsea `added\tdeleted\tpath`, con `-` (binario) → 0
- [x] 2.2 [red] Test: churn por archivo = suma de los dos diffs (merge-base...HEAD y --cached)
- [x] 2.3 [green] `scanner.go`: numstat por spec, acumular churn (added, deleted) por archivo en scanResult
- [x] 2.4 [red] Test: `--staged-only` cuenta churn sólo del índice

## 3. Gate combinado y reporte

- [x] 3.1 [red] Test: líneas contables suman churn de source (+ test salvo excludeTests), excluidos = 0
- [x] 3.2 [red] Test: lineMetric changed = add+del; added = sólo added
- [x] 3.3 [red] Test: mode or falla ante cualquier exceso; and sólo ante ambos; un solo presupuesto → ese
- [x] 3.4 [red] Test: maxLines 0 → gate sólo por archivos (retrocompat)
- [x] 3.5 [green] `report.go`: conteo de líneas + `exceeded()` combinado por mode
- [x] 3.6 [green] `reporter.go`: estado y SUMMARY con archivos y líneas

## 4. CLI

- [x] 4.1 [red] Test: `--max-lines`, `--mode`, `--line-metric` overridean la config; inválidos → exit 2
- [x] 4.2 [green] `check_cmd.go`: flags nuevos con fs.Visit para distinguir explícito de default
- [x] 4.3 [red] Test end-to-end: 1 archivo de 11000 líneas con `--max-lines 5000 --fail` → exit 1

## 5. Cierre

- [x] 5.1 `cd tools/scopelens && go test ./... -cover` → 100%
- [x] 5.2 `--tutorial` de scopelens menciona maxLines/mode/lineMetric (test de F-020 lo exige por reflection)
- [x] 5.3 `scopelens check --fail` sobre el repo
- [x] 5.4 CHANGELOG de scopelens + `.agent/feature-list.json` (F-021)
