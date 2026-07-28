package main

import (
	"testing"
)

// fixtures mapea cada proyecto de testdata a su lenguaje esperado.
var fixtures = []struct {
	dir  string
	lang string
}{
	{"testdata/go_project", "go"},
	{"testdata/ts_project", "typescript"},
	{"testdata/python_project", "python"},
	{"testdata/ruby_project", "ruby"},
	{"testdata/rust_project", "rust"},
	{"testdata/dart_project", "dart"},
}

// 8.12 — la detección es determinista y acierta el lenguaje de cada fixture.
func TestFase8_DeteccionPorFixture(t *testing.T) {
	for _, f := range fixtures {
		if got := detectLanguage(f.dir, defaultConfig.Exclude); got != f.lang {
			t.Errorf("detectLanguage(%s) = %q, want %q", f.dir, got, f.lang)
		}
	}
}

// 8.7 — los 6 fixtures dan 0 falsos positivos SIN --lang.
func TestFase8_CeroFalsosPositivosAuto(t *testing.T) {
	for _, f := range fixtures {
		cov, err := checkCoverage(config{language: "auto", root: f.dir})
		if err != nil {
			t.Fatalf("%s: checkCoverage: %v", f.dir, err)
		}
		if len(cov.untested) != 0 {
			t.Errorf("%s: falsos positivos en auto: %v", f.dir, cov.untested)
		}
	}
}

// 8.6 — 'check' sin --lang da el MISMO resultado que 'check --lang <detectado>'.
func TestFase8_AutoIgualExplicito(t *testing.T) {
	for _, f := range fixtures {
		auto, err1 := checkCoverage(config{language: "auto", root: f.dir})
		expl, err2 := checkCoverage(config{language: f.lang, root: f.dir})
		if err1 != nil || err2 != nil {
			t.Fatalf("%s: err auto=%v expl=%v", f.dir, err1, err2)
		}
		if !equalSlice(auto.untested, expl.untested) || auto.scanned != expl.scanned {
			t.Errorf("%s: auto (%v/%d) != explícito (%v/%d)",
				f.dir, auto.untested, auto.scanned, expl.untested, expl.scanned)
		}
	}
}

// 8.3 — 20 corridas sobre el mismo árbol dan salida idéntica.
func TestFase8_Determinismo20Corridas(t *testing.T) {
	var base string
	for i := 0; i < 20; i++ {
		out := capturaSalida(t, func() {
			runCheck([]string{"--lang", "auto", "--dir", "testdata/monorepo",
				"--no-color", "--format", "json"})
		})
		if i == 0 {
			base = out
			continue
		}
		if out != base {
			t.Fatalf("corrida %d difiere:\n--- base ---\n%s\n--- got ---\n%s", i, base, out)
		}
	}
}
