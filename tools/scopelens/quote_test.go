package main

import (
	"reflect"
	"testing"
)

func TestParseNameOnlyVacio(t *testing.T) {
	if got := parseNameOnly([]byte("")); got != nil {
		t.Fatalf("esperado nil, got %v", got)
	}
	if got := parseNameOnly([]byte("\n\n")); got != nil {
		t.Fatalf("líneas vacías deben ignorarse, got %v", got)
	}
}

func TestParseNameOnlyOctal(t *testing.T) {
	got := parseNameOnly([]byte("\"caf\\303\\251.ts\"\n"))
	want := []string{"café.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("octal mal decodificado: got %v want %v", got, want)
	}
}

func TestParseNameOnlyEspacios(t *testing.T) {
	got := parseNameOnly([]byte("docs/guía de uso.md\n"))
	if len(got) != 1 || got[0] != "docs/guía de uso.md" {
		t.Fatalf("ruta con espacios mal parseada: %v", got)
	}
}

func TestParseNameOnlyVarias(t *testing.T) {
	got := parseNameOnly([]byte("a.go\nb.go\n"))
	want := []string{"a.go", "b.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestUnquotePathEscapes(t *testing.T) {
	cases := map[string]string{
		`a\tb`:   "a\tb",
		`a\nb`:   "a\nb",
		`a\rb`:   "a\rb",
		`a\"b`:   `a"b`,
		`a\\b`:   `a\b`,
		`plain`:  "plain",
		`end\`:   `end\`, // backslash final sin escape
		`\101BC`: "ABC",  // octal \101 = 'A'
	}
	for in, want := range cases {
		if got := unquotePath(in); got != want {
			t.Errorf("unquotePath(%q) = %q, want %q", in, got, want)
		}
	}
}
