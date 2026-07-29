package main

type excludedFile struct {
	Path   string
	Reason string
}

// report es el resultado clasificado listo para imprimir.
type report struct {
	scanResult
	Source       []string
	Test         []string
	Excluded     []excludedFile
	Countable    int
	Lines        int
	Max          int
	MaxLines     int
	Mode         string
	LineMetric   string
	ExcludeTests bool
	StagedOnly   bool
}

// buildReport clasifica cada archivo tocado y calcula el conteo contable de
// archivos y de líneas. Los archivos ya vienen ordenados lexicográficamente
// desde measure, así que cada categoría preserva ese orden (salida determinista).
func buildReport(res scanResult, exclude []string, maxFiles, maxLines int, mode, lineMetric string, excludeTests, stagedOnly bool) report {
	r := report{scanResult: res, Max: maxFiles, MaxLines: maxLines, Mode: mode,
		LineMetric: lineMetric, ExcludeTests: excludeTests, StagedOnly: stagedOnly}
	for _, f := range res.Files {
		switch classify(f, exclude) {
		case catExcluded:
			r.Excluded = append(r.Excluded, excludedFile{Path: f, Reason: excludeReason(f)})
		case catTest:
			r.Test = append(r.Test, f)
		default:
			r.Source = append(r.Source, f)
		}
	}
	r.Countable = len(r.Source)
	r.Lines = countLinesFor(res.Churn, r.Source, lineMetric)
	if !excludeTests {
		r.Countable += len(r.Test)
		r.Lines += countLinesFor(res.Churn, r.Test, lineMetric)
	}
	return r
}

// countLinesFor suma el churn de los archivos dados según la métrica: "added"
// cuenta sólo agregadas, cualquier otra (default "changed") agregadas + borradas.
func countLinesFor(churn map[string]lineStat, files []string, lineMetric string) int {
	total := 0
	for _, f := range files {
		st := churn[f]
		if lineMetric == "added" {
			total += st.Added
		} else {
			total += st.Added + st.Deleted
		}
	}
	return total
}

// exceeded combina los presupuestos. El de archivos siempre está activo; el de
// líneas sólo si MaxLines > 0. Con ambos activos, Mode decide: "and" exige que
// se excedan los dos; "or" (default) que se exceda cualquiera.
func (r report) exceeded() bool {
	filesOver := r.Countable > r.Max
	if r.MaxLines <= 0 {
		return filesOver
	}
	linesOver := r.Lines > r.MaxLines
	if r.Mode == "and" {
		return filesOver && linesOver
	}
	return filesOver || linesOver
}
