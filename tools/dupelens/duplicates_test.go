package main

import (
	"sort"
	"testing"
)

// tf describe un archivo de prueba por su nombre y su secuencia de tokens
// (un token por línea, empezando en la línea 1).
type tf struct {
	name   string
	tokens []string
}

// buildFiles arma files + fingerprints crudos y normalizados a partir de tf,
// reproduciendo lo que hace scan() sin tocar el filesystem.
func buildFiles(windowSize int, in ...tf) ([]fileData, []Fingerprint, []Fingerprint) {
	var files []fileData
	var raw, norm []Fingerprint
	for _, f := range in {
		toks := make([]Token, len(f.tokens))
		for i, v := range f.tokens {
			toks[i] = Token{Value: v, Line: i + 1}
		}
		ntoks := normalizeTokens(toks)
		id := len(files)
		files = append(files, fileData{name: f.name, raw: toks, norm: ntoks})
		raw = append(raw, fingerprint(toks, id, windowSize)...)
		norm = append(norm, fingerprintNormalized(ntoks, id, windowSize)...)
	}
	return files, raw, norm
}

func TestFindDuplicates_emptyInputReturnsNil(t *testing.T) {
	if got := findDuplicates(nil, nil, nil, 5, 1, 0); len(got) != 0 {
		t.Errorf("expected no matches for nil input, got %d", len(got))
	}
}

func TestFindDuplicates_twoFilesSameTokens_oneExactMatch(t *testing.T) {
	files, raw, norm := buildFiles(3,
		tf{"a.go", []string{"foo", "bar", "baz", "qux"}},
		tf{"b.go", []string{"foo", "bar", "baz", "qux"}})
	got := findDuplicates(files, raw, norm, 3, 1, 0)
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(got), got)
	}
	if got[0].Kind != "exact" {
		t.Errorf("Kind = %q; want exact", got[0].Kind)
	}
	if got[0].Tokens != 4 {
		t.Errorf("Tokens = %d; want 4 (2 windows merged)", got[0].Tokens)
	}
}

func TestFindDuplicates_hashCollisionWithDifferentTokens_noMatch(t *testing.T) {
	files := []fileData{
		{name: "a.go", raw: tok([]string{"x", "y", "z"}, []int{1, 2, 3})},
		{name: "b.go", raw: tok([]string{"a", "b", "c"}, []int{1, 2, 3})},
	}
	rawFps := []Fingerprint{
		{Hash: 100, FileID: 0, StartIdx: 0, StartLine: 1, EndLine: 3},
		{Hash: 100, FileID: 1, StartIdx: 0, StartLine: 1, EndLine: 3},
	}
	if got := findDuplicates(files, rawFps, nil, 3, 1, 0); len(got) != 0 {
		t.Errorf("colisión de hash con tokens distintos no debe matchear, got %d", len(got))
	}
}

func TestFindDuplicates_selfPositionGuard(t *testing.T) {
	files := []fileData{{name: "a.go", raw: tok([]string{"x", "y", "z"}, []int{1, 2, 3})}}
	rawFps := []Fingerprint{
		{Hash: 1, FileID: 0, StartIdx: 0, StartLine: 1, EndLine: 3},
		{Hash: 1, FileID: 0, StartIdx: 0, StartLine: 1, EndLine: 3},
	}
	if got := findDuplicates(files, rawFps, nil, 3, 1, 0); len(got) != 0 {
		t.Errorf("mismo fileID+startIdx no debe matchear consigo mismo, got %d", len(got))
	}
}

func TestFindDuplicates_overlappingMerged(t *testing.T) {
	files, raw, norm := buildFiles(3,
		tf{"a.go", []string{"p", "q", "r", "s", "t"}},
		tf{"b.go", []string{"p", "q", "r", "s", "t"}})
	got := findDuplicates(files, raw, norm, 3, 1, 0)
	if len(got) != 1 {
		t.Fatalf("expected 1 merged match, got %d", len(got))
	}
	if got[0].StartLineA != 1 || got[0].EndLineA != 5 {
		t.Errorf("merged A range = %d-%d; want 1-5", got[0].StartLineA, got[0].EndLineA)
	}
	if got[0].Tokens != 5 {
		t.Errorf("merged Tokens = %d; want 5", got[0].Tokens)
	}
}

func TestFindDuplicates_minLinesFilters(t *testing.T) {
	files, raw, norm := buildFiles(3,
		tf{"a.go", []string{"x", "y", "z"}},
		tf{"b.go", []string{"x", "y", "z"}})
	if got := findDuplicates(files, raw, norm, 3, 5, 0); len(got) != 0 {
		t.Errorf("match de 3 líneas debe filtrarse con minLines=5, got %d", len(got))
	}
	if got := findDuplicates(files, raw, norm, 3, 3, 0); len(got) != 1 {
		t.Errorf("match de 3 líneas debe pasar minLines=3, got %d", len(got))
	}
}

func TestFindDuplicates_minTokensFilters(t *testing.T) {
	files, raw, norm := buildFiles(3,
		tf{"a.go", []string{"x", "y", "z"}},
		tf{"b.go", []string{"x", "y", "z"}})
	if got := findDuplicates(files, raw, norm, 3, 1, 5); len(got) != 0 {
		t.Errorf("match de 3 tokens debe filtrarse con minTokens=5, got %d", len(got))
	}
	if got := findDuplicates(files, raw, norm, 3, 1, 3); len(got) != 1 {
		t.Errorf("match de 3 tokens debe pasar minTokens=3, got %d", len(got))
	}
}

func TestFindDuplicates_intraFileDuplicate(t *testing.T) {
	files, raw, norm := buildFiles(3,
		tf{"a.go", []string{"p", "q", "r", "s", "t", "p", "q", "r", "s", "t"}})
	got := findDuplicates(files, raw, norm, 3, 1, 0)
	if len(got) == 0 {
		t.Fatal("intra-file dup esperaba ≥1 match, got 0")
	}
	if got[0].FileA != got[0].FileB {
		t.Errorf("intra-file: FileA y FileB deben ser iguales")
	}
}

func TestFindDuplicates_threeFilesSame_threePairs(t *testing.T) {
	files, raw, norm := buildFiles(3,
		tf{"a.go", []string{"x", "y", "z", "w"}},
		tf{"b.go", []string{"x", "y", "z", "w"}},
		tf{"c.go", []string{"x", "y", "z", "w"}})
	got := findDuplicates(files, raw, norm, 3, 1, 0)
	if len(got) != 3 {
		t.Fatalf("expected 3 pairwise matches, got %d", len(got))
	}
	pairs := map[string]bool{}
	for _, m := range got {
		pairs[m.FileA+"|"+m.FileB] = true
	}
	for _, w := range []string{"a.go|b.go", "a.go|c.go", "b.go|c.go"} {
		if !pairs[w] {
			t.Errorf("falta el par %q", w)
		}
	}
}

func TestFindDuplicates_sortedDeterministic(t *testing.T) {
	files, raw, norm := buildFiles(3,
		tf{"z.go", []string{"x", "y", "z", "w"}},
		tf{"a.go", []string{"x", "y", "z", "w"}},
		tf{"m.go", []string{"x", "y", "z", "w"}})
	got := findDuplicates(files, raw, norm, 3, 1, 0)
	var keys []string
	for _, m := range got {
		keys = append(keys, m.FileA+"|"+m.FileB)
	}
	if !sort.StringsAreSorted(keys) {
		t.Errorf("matches no ordenados: %v", keys)
	}
}

func TestFindDuplicates_renamedDetection(t *testing.T) {
	// Misma estructura (return ID ID end), identificadores distintos → renamed.
	files, raw, norm := buildFiles(4,
		tf{"a.go", []string{"return", "alpha", "beta", "end"}},
		tf{"b.go", []string{"return", "gamma", "delta", "end"}})
	got := findDuplicates(files, raw, norm, 4, 1, 0)
	if len(got) != 1 {
		t.Fatalf("expected 1 renamed match, got %d: %+v", len(got), got)
	}
	if got[0].Kind != "renamed" {
		t.Errorf("Kind = %q; want renamed", got[0].Kind)
	}
}

func TestFindDuplicates_exactNotDoubleReportedAsRenamed(t *testing.T) {
	files, raw, norm := buildFiles(4,
		tf{"a.go", []string{"return", "alpha", "beta", "end"}},
		tf{"b.go", []string{"return", "alpha", "beta", "end"}})
	got := findDuplicates(files, raw, norm, 4, 1, 0)
	if len(got) != 1 {
		t.Fatalf("copia literal debe dar 1 match, got %d: %+v", len(got), got)
	}
	if got[0].Kind != "exact" {
		t.Errorf("Kind = %q; want exact", got[0].Kind)
	}
}

func TestFindDuplicates_monotoneWindowsIgnored(t *testing.T) {
	// Bloque repetitivo de identificadores distintos: crudo no matchea (valores
	// distintos) y normalizado colapsa a "ID ID ID …" (monótono) → sin ruido.
	files, raw, norm := buildFiles(3,
		tf{"a.go", []string{"aa", "bb", "cc", "dd", "ee", "ff"}})
	if got := findDuplicates(files, raw, norm, 3, 1, 0); len(got) != 0 {
		t.Errorf("bloque de IDs distintos no debe generar matches, got %d: %+v", len(got), got)
	}
}

func TestSameTokens_outOfRangeReturnsFalse(t *testing.T) {
	a := tok([]string{"x", "y"}, []int{1, 2})
	if sameTokens(a, 0, a, 0, 5) {
		t.Error("rango fuera de slice debe devolver false")
	}
}
