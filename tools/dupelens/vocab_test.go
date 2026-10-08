package main

import "testing"

func TestVocab_sameStringSameId(t *testing.T) {
	v := newVocab()
	if v.id("foo") != v.id("foo") {
		t.Error("el mismo string debe recibir el mismo id")
	}
	if v.id("foo") == v.id("bar") {
		t.Error("strings distintos deben recibir ids distintos")
	}
}

// La tabla de normalización debe producir la misma forma que normalizeValue:
// keyword intacta, literal numérico a NUM, identificador a ID, el resto tal cual.
func TestVocab_normalizedFormMatchesNormalizeValue(t *testing.T) {
	v := newVocab()
	for _, s := range []string{"func", "42", "3.14", "foo", "foo_1", "a+b", "NUM", "ID"} {
		if got, want := v.norm[v.id(s)], v.id(normalizeValue(s)); got != want {
			t.Errorf("norm(%q) = id %d; want id de %q (%d)", s, got, normalizeValue(s), want)
		}
	}
}

func TestVocab_atSelectsView(t *testing.T) {
	v := newVocab()
	f := fdOf(v, "a", "foo")
	if v.at(f, 0, false) != v.id("foo") || v.at(f, 0, true) != v.id("ID") {
		t.Error("at debe devolver el id crudo o el normalizado según la vista")
	}
}
