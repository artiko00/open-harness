package main

import (
	"sort"
	"testing"
)

func TestFindDuplicates_emptyInputReturnsNil(t *testing.T) {
	if got := findDuplicates(nil, 5, 1); len(got) != 0 {
		t.Errorf("expected no matches for nil input, got %d", len(got))
	}
}

func TestFindDuplicates_singleFingerprintProducesNoMatch(t *testing.T) {
	perFile := map[string][]Fingerprint{
		"a.go": {{Hash: 100, StartLine: 1, EndLine: 5, Window: []string{"foo", "bar", "baz"}}},
	}
	if got := findDuplicates(perFile, 3, 1); len(got) != 0 {
		t.Errorf("single fp must produce no match, got %d", len(got))
	}
}

func TestFindDuplicates_twoFilesSameHashAndWindow_oneMatch(t *testing.T) {
	perFile := map[string][]Fingerprint{
		"a.go": {{Hash: 100, StartLine: 1, EndLine: 5, Window: []string{"x", "y", "z"}}},
		"b.go": {{Hash: 100, StartLine: 10, EndLine: 14, Window: []string{"x", "y", "z"}}},
	}
	got := findDuplicates(perFile, 3, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}
	m := got[0]
	if m.Tokens != 3 {
		t.Errorf("Tokens = %d; want 3", m.Tokens)
	}
}

func TestFindDuplicates_hashCollisionWithDifferentWindow_noMatch(t *testing.T) {
	// Mismo hash, contenido del Window distinto = colisión Rabin-Karp.
	// findDuplicates debe verificar Window literal y descartar.
	perFile := map[string][]Fingerprint{
		"a.go": {{Hash: 100, StartLine: 1, EndLine: 3, Window: []string{"x", "y", "z"}}},
		"b.go": {{Hash: 100, StartLine: 1, EndLine: 3, Window: []string{"a", "b", "c"}}},
	}
	if got := findDuplicates(perFile, 3, 1); len(got) != 0 {
		t.Errorf("hash collision with different windows must NOT produce match, got %d", len(got))
	}
}

func TestFindDuplicates_overlappingFingerprintsMergedToOne(t *testing.T) {
	// 3 ventanas consecutivas que matchean entre 2 archivos.
	// La detección honesta colapsa los 3 fp-matches en 1 Match con rango extendido.
	perFile := map[string][]Fingerprint{
		"a.go": {
			{Hash: 1, StartLine: 1, EndLine: 3, Window: []string{"x", "y", "z"}},
			{Hash: 2, StartLine: 2, EndLine: 4, Window: []string{"y", "z", "w"}},
			{Hash: 3, StartLine: 3, EndLine: 5, Window: []string{"z", "w", "v"}},
		},
		"b.go": {
			{Hash: 1, StartLine: 10, EndLine: 12, Window: []string{"x", "y", "z"}},
			{Hash: 2, StartLine: 11, EndLine: 13, Window: []string{"y", "z", "w"}},
			{Hash: 3, StartLine: 12, EndLine: 14, Window: []string{"z", "w", "v"}},
		},
	}
	got := findDuplicates(perFile, 3, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 merged match, got %d", len(got))
	}
	m := got[0]
	if m.StartLineA != 1 || m.EndLineA != 5 {
		t.Errorf("merged A range = %d-%d; want 1-5", m.StartLineA, m.EndLineA)
	}
	if m.StartLineB != 10 || m.EndLineB != 14 {
		t.Errorf("merged B range = %d-%d; want 10-14", m.StartLineB, m.EndLineB)
	}
	if m.Tokens != 5 {
		t.Errorf("merged Tokens = %d; want 5 (3 fps merged with windowSize 3)", m.Tokens)
	}
}

func TestFindDuplicates_minLinesFiltersShortMatches(t *testing.T) {
	perFile := map[string][]Fingerprint{
		"a.go": {{Hash: 1, StartLine: 1, EndLine: 2, Window: []string{"x", "y", "z"}}},
		"b.go": {{Hash: 1, StartLine: 5, EndLine: 6, Window: []string{"x", "y", "z"}}},
	}
	if got := findDuplicates(perFile, 3, 5); len(got) != 0 {
		t.Errorf("match with 2 lines must be filtered by minLines=5")
	}
	if got := findDuplicates(perFile, 3, 2); len(got) != 1 {
		t.Errorf("match with 2 lines must pass minLines=2")
	}
}

func TestFindDuplicates_intraFileDuplicateProducesMatch(t *testing.T) {
	// Función copy-pasted dentro del mismo archivo.
	perFile := map[string][]Fingerprint{
		"a.go": {
			{Hash: 1, StartLine: 1, EndLine: 3, Window: []string{"x", "y", "z"}},
			{Hash: 1, StartLine: 100, EndLine: 102, Window: []string{"x", "y", "z"}},
		},
	}
	got := findDuplicates(perFile, 3, 1)
	if len(got) != 1 {
		t.Fatalf("intra-file dup expected 1 match, got %d", len(got))
	}
	if got[0].FileA != got[0].FileB {
		t.Errorf("intra-file match: FileA y FileB deben ser iguales")
	}
}

func TestFindDuplicates_threeFilesSameHash_threePairs(t *testing.T) {
	perFile := map[string][]Fingerprint{
		"a.go": {{Hash: 1, StartLine: 1, EndLine: 3, Window: []string{"x", "y", "z"}}},
		"b.go": {{Hash: 1, StartLine: 1, EndLine: 3, Window: []string{"x", "y", "z"}}},
		"c.go": {{Hash: 1, StartLine: 1, EndLine: 3, Window: []string{"x", "y", "z"}}},
	}
	got := findDuplicates(perFile, 3, 1)
	// 3 archivos con misma ventana → C(3,2) = 3 pares
	if len(got) != 3 {
		t.Fatalf("expected 3 pairwise matches, got %d", len(got))
	}
	// Verificar todos los pares posibles
	pairs := make(map[string]bool)
	for _, m := range got {
		a, b := m.FileA, m.FileB
		if a > b {
			a, b = b, a
		}
		pairs[a+"|"+b] = true
	}
	want := []string{"a.go|b.go", "a.go|c.go", "b.go|c.go"}
	for _, w := range want {
		if !pairs[w] {
			t.Errorf("expected pair %q in matches", w)
		}
	}
}

func TestFindDuplicates_resultsSortedDeterministic(t *testing.T) {
	// Output debe ser estable para tests/diffs reproducibles.
	perFile := map[string][]Fingerprint{
		"z.go": {{Hash: 1, StartLine: 5, EndLine: 7, Window: []string{"x", "y", "z"}}},
		"a.go": {{Hash: 1, StartLine: 1, EndLine: 3, Window: []string{"x", "y", "z"}}},
		"m.go": {{Hash: 1, StartLine: 9, EndLine: 11, Window: []string{"x", "y", "z"}}},
	}
	got := findDuplicates(perFile, 3, 1)
	files := []string{}
	for _, m := range got {
		files = append(files, m.FileA+"|"+m.FileB)
	}
	if !sort.StringsAreSorted(files) {
		t.Errorf("matches no están ordenados deterministicamente: %v", files)
	}
}
