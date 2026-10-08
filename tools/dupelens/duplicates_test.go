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

// buildFiles interna los archivos de prueba en un vocabulario compartido,
// reproduciendo lo que hace scan() sin tocar el filesystem.
func buildFiles(in ...tf) ([]fileData, *vocab) {
	v := newVocab()
	var files []fileData
	for _, f := range in {
		files = append(files, fdOf(v, f.name, f.tokens...))
	}
	return files, v
}

func find(t *testing.T, w, minLines, minTokens int, in ...tf) []Match {
	t.Helper()
	files, v := buildFiles(in...)
	got, err := findDuplicates(files, v, w, minLines, minTokens, memGuard{})
	if err != nil {
		t.Fatalf("findDuplicates: %v", err)
	}
	return got
}

func TestFindDuplicates_emptyInputReturnsNil(t *testing.T) {
	if got := find(t, 5, 1, 0); len(got) != 0 {
		t.Errorf("expected no matches for nil input, got %d", len(got))
	}
}

func TestFindDuplicates_twoFilesSameTokens_oneExactMatch(t *testing.T) {
	got := find(t, 3, 1, 0,
		tf{"a.go", []string{"foo", "bar", "baz", "qux"}},
		tf{"b.go", []string{"foo", "bar", "baz", "qux"}})
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(got), got)
	}
	if got[0].Kind != "exact" || got[0].Tokens != 4 {
		t.Errorf("Kind=%q Tokens=%d; want exact/4", got[0].Kind, got[0].Tokens)
	}
}

// Ventanas con el mismo hash pero tokens distintos: la verificación literal
// debe descartarlas. Se fuerza la colisión igualando el hash de cada token.
func TestFindDuplicates_hashCollisionWithDifferentTokens_noMatch(t *testing.T) {
	files, v := buildFiles(
		tf{"a.go", []string{"xx", "yy", "zz"}},
		tf{"b.go", []string{"aa", "bb", "cc"}})
	for _, p := range [][2]string{{"aa", "xx"}, {"bb", "yy"}, {"cc", "zz"}} {
		v.hash[v.id(p[0])] = v.hash[v.id(p[1])]
		v.hash[v.norm[v.id(p[0])]] = v.hash[v.norm[v.id(p[1])]]
	}
	got, err := findDuplicates(files, v, 3, 1, 0, memGuard{})
	if err != nil || len(got) != 0 {
		t.Errorf("colisión de hash con tokens distintos no debe matchear, got %d (err %v)", len(got), err)
	}
}

// Una ventana que el prefiltro deja pasar sin par real (falso positivo del
// bitset) forma un grupo de un solo miembro: no produce semillas.
func TestAnchorSeeds_singletonGroupProducesNoSeed(t *testing.T) {
	files, v := buildFiles(tf{"a.go", []string{"xx", "yy", "zz"}})
	seeds, err := anchorSeeds(files, v, []rec{{h: 7, f: 0, i: 0}}, 3, false, memGuard{})
	if err != nil || len(seeds) != 0 {
		t.Errorf("grupo de un miembro: seeds=%d err=%v; want 0/nil", len(seeds), err)
	}
}

func TestCandidates_uniqueWindowsAreNotMaterialized(t *testing.T) {
	files, v := buildFiles(
		tf{"a.go", []string{"p1", "p2", "p3", "u1", "u2"}},
		tf{"b.go", []string{"p1", "p2", "p3", "w1", "w2"}})
	recs, err := candidates(files, v, 3, false, memGuard{})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Errorf("solo la ventana repetida (p1 p2 p3, una por archivo) debe materializarse; got %d", len(recs))
	}
}

func TestFindDuplicates_overlappingMerged(t *testing.T) {
	got := find(t, 3, 1, 0,
		tf{"a.go", []string{"p", "q", "r", "s", "t"}},
		tf{"b.go", []string{"p", "q", "r", "s", "t"}})
	if len(got) != 1 {
		t.Fatalf("expected 1 merged match, got %d", len(got))
	}
	if got[0].StartLineA != 1 || got[0].EndLineA != 5 || got[0].Tokens != 5 {
		t.Errorf("merged = %d-%d/%d tokens; want 1-5/5", got[0].StartLineA, got[0].EndLineA, got[0].Tokens)
	}
}

func TestFindDuplicates_minLinesAndMinTokensFilter(t *testing.T) {
	in := []tf{{"a.go", []string{"x", "y", "z"}}, {"b.go", []string{"x", "y", "z"}}}
	if got := find(t, 3, 5, 0, in...); len(got) != 0 {
		t.Errorf("match de 3 líneas debe filtrarse con minLines=5, got %d", len(got))
	}
	if got := find(t, 3, 3, 0, in...); len(got) != 1 {
		t.Errorf("match de 3 líneas debe pasar minLines=3, got %d", len(got))
	}
	if got := find(t, 3, 1, 5, in...); len(got) != 0 {
		t.Errorf("match de 3 tokens debe filtrarse con minTokens=5, got %d", len(got))
	}
	if got := find(t, 3, 1, 3, in...); len(got) != 1 {
		t.Errorf("match de 3 tokens debe pasar minTokens=3, got %d", len(got))
	}
}

func TestFindDuplicates_intraFileDuplicate(t *testing.T) {
	got := find(t, 3, 1, 0,
		tf{"a.go", []string{"p", "q", "r", "s", "t", "p", "q", "r", "s", "t"}})
	if len(got) == 0 || got[0].FileA != got[0].FileB {
		t.Fatalf("intra-file dup esperaba ≥1 match con FileA == FileB, got %+v", got)
	}
}

func TestFindDuplicates_threeFilesSame_anchoredToFirst(t *testing.T) {
	got := find(t, 3, 1, 0,
		tf{"c.go", []string{"x", "y", "z", "w"}},
		tf{"a.go", []string{"x", "y", "z", "w"}},
		tf{"b.go", []string{"x", "y", "z", "w"}})
	pairs := pairsOf(got)
	if len(got) != 2 || pairs["a.go|b.go"] != 1 || pairs["a.go|c.go"] != 1 {
		t.Fatalf("expected a-b and a-c anchored to a.go, got %v", pairs)
	}
}

func TestFindDuplicates_sortedDeterministic(t *testing.T) {
	got := find(t, 3, 1, 0,
		tf{"z.go", []string{"x", "y", "z", "w"}},
		tf{"a.go", []string{"x", "y", "z", "w"}},
		tf{"m.go", []string{"x", "y", "z", "w"}})
	var keys []string
	for _, m := range got {
		keys = append(keys, m.FileA+"|"+m.FileB)
	}
	if !sort.StringsAreSorted(keys) {
		t.Errorf("matches no ordenados: %v", keys)
	}
}

func TestFindDuplicates_renamedDetection(t *testing.T) {
	got := find(t, 4, 1, 0,
		tf{"a.go", []string{"return", "alpha", "beta", "end"}},
		tf{"b.go", []string{"return", "gamma", "delta", "end"}})
	if len(got) != 1 || got[0].Kind != "renamed" {
		t.Fatalf("expected 1 renamed match, got %+v", got)
	}
}

func TestFindDuplicates_exactNotDoubleReportedAsRenamed(t *testing.T) {
	got := find(t, 4, 1, 0,
		tf{"a.go", []string{"return", "alpha", "beta", "end"}},
		tf{"b.go", []string{"return", "alpha", "beta", "end"}})
	if len(got) != 1 || got[0].Kind != "exact" {
		t.Fatalf("copia literal debe dar 1 match exact, got %+v", got)
	}
}

func TestFindDuplicates_monotoneWindowsIgnored(t *testing.T) {
	// Crudo no matchea (valores distintos) y normalizado colapsa a "ID ID ID …"
	// (monótono) → sin ruido.
	if got := find(t, 3, 1, 0, tf{"a.go", []string{"aa", "bb", "cc", "dd", "ee", "ff"}}); len(got) != 0 {
		t.Errorf("bloque de IDs distintos no debe generar matches, got %+v", got)
	}
}
