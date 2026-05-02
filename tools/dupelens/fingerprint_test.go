package main

import "testing"

func tok(values []string, lines []int) []Token {
	if len(values) != len(lines) {
		panic("test helper: values y lines deben tener mismo largo")
	}
	out := make([]Token, len(values))
	for i, v := range values {
		out[i] = Token{Value: v, Line: lines[i]}
	}
	return out
}

func TestFingerprint_returnsNilWhenTokensSmallerThanWindow(t *testing.T) {
	tokens := tok([]string{"a", "b"}, []int{1, 1})
	if got := fingerprint(tokens, 5); got != nil {
		t.Errorf("expected nil when len(tokens)<window, got %d fps", len(got))
	}
}

func TestFingerprint_returnsNilWhenWindowSizeZeroOrNegative(t *testing.T) {
	tokens := tok([]string{"a", "b", "c"}, []int{1, 1, 1})
	if got := fingerprint(tokens, 0); got != nil {
		t.Errorf("expected nil for windowSize=0, got %d", len(got))
	}
	if got := fingerprint(tokens, -1); got != nil {
		t.Errorf("expected nil for windowSize=-1, got %d", len(got))
	}
}

func TestFingerprint_emitsOneWindowWhenLenEqualsWindow(t *testing.T) {
	tokens := tok([]string{"a", "b", "c"}, []int{1, 2, 3})
	got := fingerprint(tokens, 3)
	if len(got) != 1 {
		t.Fatalf("expected 1 fp, got %d", len(got))
	}
	if got[0].StartLine != 1 || got[0].EndLine != 3 {
		t.Errorf("expected lines 1-3, got %d-%d", got[0].StartLine, got[0].EndLine)
	}
	if len(got[0].Window) != 3 {
		t.Errorf("expected window of 3 tokens, got %d", len(got[0].Window))
	}
}

func TestFingerprint_emitsSlidingWindows(t *testing.T) {
	tokens := tok([]string{"a", "b", "c", "d", "e"}, []int{1, 1, 2, 2, 3})
	got := fingerprint(tokens, 3)
	// len 5 - window 3 + 1 = 3 fingerprints
	if len(got) != 3 {
		t.Fatalf("expected 3 fps for sliding window, got %d", len(got))
	}
	wantLines := [][2]int{{1, 2}, {1, 2}, {2, 3}}
	for i, w := range wantLines {
		if got[i].StartLine != w[0] || got[i].EndLine != w[1] {
			t.Errorf("fp[%d]: want lines %d-%d, got %d-%d", i, w[0], w[1], got[i].StartLine, got[i].EndLine)
		}
	}
}

func TestFingerprint_identicalSequencesProduceSameHash(t *testing.T) {
	a := tok([]string{"foo", "bar", "baz"}, []int{1, 1, 1})
	b := tok([]string{"foo", "bar", "baz"}, []int{42, 42, 42})
	hashA := fingerprint(a, 3)[0].Hash
	hashB := fingerprint(b, 3)[0].Hash
	if hashA != hashB {
		t.Errorf("identical token sequences must produce same hash: %d != %d", hashA, hashB)
	}
}

func TestFingerprint_differentSequencesProduceDifferentHash(t *testing.T) {
	a := tok([]string{"foo", "bar", "baz"}, []int{1, 1, 1})
	b := tok([]string{"foo", "bar", "qux"}, []int{1, 1, 1})
	hashA := fingerprint(a, 3)[0].Hash
	hashB := fingerprint(b, 3)[0].Hash
	if hashA == hashB {
		t.Errorf("different sequences should have different hash, both got %d", hashA)
	}
}

func TestFingerprint_windowContentsPreserved(t *testing.T) {
	tokens := tok([]string{"foo", "bar", "baz", "qux"}, []int{1, 2, 3, 4})
	got := fingerprint(tokens, 3)
	wantWindows := [][]string{
		{"foo", "bar", "baz"},
		{"bar", "baz", "qux"},
	}
	for i, w := range wantWindows {
		if len(got[i].Window) != len(w) {
			t.Fatalf("fp[%d] window len mismatch", i)
			continue
		}
		for j := range w {
			if got[i].Window[j] != w[j] {
				t.Errorf("fp[%d].Window[%d] = %q; want %q", i, j, got[i].Window[j], w[j])
			}
		}
	}
}

func TestFingerprint_rollingHashConsistent_naivVsRolling(t *testing.T) {
	// Verifica que el hash producido por sliding (rolling) coincide con
	// hashear cada ventana desde cero: protege contra bugs aritméticos
	// en la sustracción del primer token al deslizar.
	tokens := tok([]string{"x", "y", "z", "w", "u", "v"}, []int{1, 1, 1, 2, 2, 2})
	rolling := fingerprint(tokens, 3)

	// Para cada ventana, calcular hash desde cero usando la misma estrategia
	// que la implementación interna (sum-acumulado de hashToken con base/mod).
	for i, fp := range rolling {
		var naive uint64
		for j := 0; j < 3; j++ {
			naive = (naive*rkBase + hashToken(tokens[i+j].Value)) % rkMod
		}
		if naive != fp.Hash {
			t.Errorf("fp[%d] rolling hash %d != naive %d", i, fp.Hash, naive)
		}
	}
}
