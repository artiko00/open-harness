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
	if got := fingerprint(tokens, 0, 5); got != nil {
		t.Errorf("expected nil when len(tokens)<window, got %d fps", len(got))
	}
}

func TestFingerprint_returnsNilWhenWindowSizeZeroOrNegative(t *testing.T) {
	tokens := tok([]string{"a", "b", "c"}, []int{1, 1, 1})
	if got := fingerprint(tokens, 0, 0); got != nil {
		t.Errorf("expected nil for windowSize=0, got %d", len(got))
	}
	if got := fingerprint(tokens, 0, -1); got != nil {
		t.Errorf("expected nil for windowSize=-1, got %d", len(got))
	}
}

func TestFingerprint_emitsOneWindowWhenLenEqualsWindow(t *testing.T) {
	tokens := tok([]string{"a", "b", "c"}, []int{1, 2, 3})
	got := fingerprint(tokens, 7, 3)
	if len(got) != 1 {
		t.Fatalf("expected 1 fp, got %d", len(got))
	}
	if got[0].StartLine != 1 || got[0].EndLine != 3 {
		t.Errorf("expected lines 1-3, got %d-%d", got[0].StartLine, got[0].EndLine)
	}
	if got[0].FileID != 7 || got[0].StartIdx != 0 {
		t.Errorf("expected FileID=7 StartIdx=0, got %d/%d", got[0].FileID, got[0].StartIdx)
	}
}

func TestFingerprint_emitsSlidingWindows(t *testing.T) {
	tokens := tok([]string{"a", "b", "c", "d", "e"}, []int{1, 1, 2, 2, 3})
	got := fingerprint(tokens, 0, 3)
	if len(got) != 3 {
		t.Fatalf("expected 3 fps for sliding window, got %d", len(got))
	}
	wantLines := [][2]int{{1, 2}, {1, 2}, {2, 3}}
	for i, w := range wantLines {
		if got[i].StartLine != w[0] || got[i].EndLine != w[1] {
			t.Errorf("fp[%d]: want lines %d-%d, got %d-%d", i, w[0], w[1], got[i].StartLine, got[i].EndLine)
		}
		if got[i].StartIdx != i {
			t.Errorf("fp[%d]: StartIdx = %d; want %d", i, got[i].StartIdx, i)
		}
	}
}

func TestFingerprint_identicalSequencesProduceSameHash(t *testing.T) {
	a := tok([]string{"foo", "bar", "baz"}, []int{1, 1, 1})
	b := tok([]string{"foo", "bar", "baz"}, []int{42, 42, 42})
	hashA := fingerprint(a, 0, 3)[0].Hash
	hashB := fingerprint(b, 1, 3)[0].Hash
	if hashA != hashB {
		t.Errorf("identical token sequences must produce same hash: %d != %d", hashA, hashB)
	}
}

func TestFingerprint_differentSequencesProduceDifferentHash(t *testing.T) {
	a := tok([]string{"foo", "bar", "baz"}, []int{1, 1, 1})
	b := tok([]string{"foo", "bar", "qux"}, []int{1, 1, 1})
	hashA := fingerprint(a, 0, 3)[0].Hash
	hashB := fingerprint(b, 0, 3)[0].Hash
	if hashA == hashB {
		t.Errorf("different sequences should have different hash, both got %d", hashA)
	}
}

func TestFingerprint_rollingHashConsistent_naivVsRolling(t *testing.T) {
	tokens := tok([]string{"x", "y", "z", "w", "u", "v"}, []int{1, 1, 1, 2, 2, 2})
	rolling := fingerprint(tokens, 0, 3)
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
