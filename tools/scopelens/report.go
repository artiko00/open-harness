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
	Max          int
	ExcludeTests bool
	StagedOnly   bool
}

// buildReport clasifica cada archivo tocado y calcula el conteo contable. Los
// archivos ya vienen ordenados lexicográficamente desde measure, así que cada
// categoría preserva ese orden (salida determinista).
func buildReport(res scanResult, exclude []string, maxFiles int, excludeTests, stagedOnly bool) report {
	r := report{scanResult: res, Max: maxFiles, ExcludeTests: excludeTests, StagedOnly: stagedOnly}
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
	if !excludeTests {
		r.Countable += len(r.Test)
	}
	return r
}

func (r report) exceeded() bool {
	return r.Countable > r.Max
}
