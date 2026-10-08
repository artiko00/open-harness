package main

import (
	"math/big"
	"testing"
)

// fdOf arma un fileData de un token por línea, interno en v.
func fdOf(v *vocab, name string, tokens ...string) fileData {
	f := fileData{name: name}
	for i, s := range tokens {
		f.ids = append(f.ids, v.id(s))
		f.lines = append(f.lines, uint32(i+1))
	}
	return f
}

type window struct {
	h     uint64
	start int
	mono  bool
}

func windowsOf(v *vocab, f fileData, w int, norm bool) []window {
	var out []window
	eachWindow(v, f, w, norm, func(h uint64, start int, mono bool) {
		out = append(out, window{h, start, mono})
	})
	return out
}

func TestMulMod_matchesBigIntReference(t *testing.T) {
	m := new(big.Int).SetUint64(m61)
	for _, c := range [][2]uint64{{0, 5}, {m61 - 1, m61 - 1}, {1 << 60, 3}, {123456789123, 987654321987}} {
		want := new(big.Int).Mul(new(big.Int).SetUint64(c[0]), new(big.Int).SetUint64(c[1]))
		want.Mod(want, m)
		if got := mulMod(c[0], c[1]); got != want.Uint64() {
			t.Errorf("mulMod(%d, %d) = %d; want %d", c[0], c[1], got, want.Uint64())
		}
	}
}

func TestHashToken_deterministicAndBelowModulus(t *testing.T) {
	if hashToken("function") != hashToken("function") {
		t.Error("hashToken no es determinístico")
	}
	if hashToken("function") == hashToken("FUNCTION") {
		t.Error("hashToken debe distinguir mayúsculas")
	}
	if h := hashToken("un_identificador_muy_largo_para_forzar_varias_vueltas"); h >= m61 {
		t.Errorf("hashToken = %d no está reducido módulo 2^61-1", h)
	}
}

func TestEachWindow_noWindowsWhenShortOrInvalid(t *testing.T) {
	v := newVocab()
	f := fdOf(v, "a", "a1", "b1")
	for _, w := range []int{0, -1, 3} {
		if got := windowsOf(v, f, w, false); len(got) != 0 {
			t.Errorf("w=%d: se esperaban 0 ventanas, got %d", w, len(got))
		}
	}
}

func TestEachWindow_slidingStartsAndRollingEqualsNaive(t *testing.T) {
	v := newVocab()
	f := fdOf(v, "a", "xx", "yy", "zz", "ww", "uu", "vv")
	got := windowsOf(v, f, 3, false)
	if len(got) != 4 {
		t.Fatalf("se esperaban 4 ventanas, got %d", len(got))
	}
	for i, win := range got {
		var naive uint64
		for j := 0; j < 3; j++ {
			naive = addMod(mulMod(naive, rkBase), v.hash[f.ids[i+j]])
		}
		if win.start != i || win.h != naive {
			t.Errorf("ventana %d: start=%d hash=%d; want start=%d hash=%d", i, win.start, win.h, i, naive)
		}
	}
}

func TestEachWindow_sameSequenceSameHashDifferentSequenceDifferentHash(t *testing.T) {
	v := newVocab()
	a := windowsOf(v, fdOf(v, "a", "foo", "bar", "baz"), 3, false)
	b := windowsOf(v, fdOf(v, "b", "foo", "bar", "baz"), 3, false)
	c := windowsOf(v, fdOf(v, "c", "foo", "bar", "qux"), 3, false)
	if a[0].h != b[0].h {
		t.Error("secuencias idénticas deben dar el mismo hash")
	}
	if a[0].h == c[0].h {
		t.Error("secuencias distintas no deben dar el mismo hash")
	}
}

func TestEachWindow_normalizedViewUsesNormalizedIds(t *testing.T) {
	v := newVocab()
	a := windowsOf(v, fdOf(v, "a", "return", "alpha", "beta"), 3, true)
	b := windowsOf(v, fdOf(v, "b", "return", "gamma", "delta"), 3, true)
	if a[0].h != b[0].h {
		t.Error("en la vista normalizada, identificadores distintos deben dar el mismo hash")
	}
}

func TestEachWindow_flagsMonotoneWindows(t *testing.T) {
	v := newVocab()
	got := windowsOf(v, fdOf(v, "a", "aa", "aa", "aa", "bb"), 3, false)
	if !got[0].mono || got[1].mono {
		t.Errorf("mono = %v/%v; want true/false", got[0].mono, got[1].mono)
	}
}

func TestAddMod_wrapsAtModulus(t *testing.T) {
	if got := addMod(m61-1, 1); got != 0 {
		t.Errorf("addMod(m61-1, 1) = %d; want 0", got)
	}
	if got := addMod(m61-1, 5); got != 4 {
		t.Errorf("addMod(m61-1, 5) = %d; want 4", got)
	}
}
