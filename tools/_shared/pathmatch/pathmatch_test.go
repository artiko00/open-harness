package pathmatch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchGlobDobleAsteriscoCualquierProfundidad(t *testing.T) {
	casos := []struct {
		pattern string
		relPath string
		quiere  bool
	}{
		{"**/migrations/**", "db/migrations/a.sql", true},
		{"**/migrations/**", "db/migrations/2024/b.sql", true},
		{"**/migrations/**", "db/otro/a.sql", false},
		{"**/*.spec.ts", "src/a/b/c.spec.ts", true},
		{"src/**/*.ts", "src/a/b/c.ts", true},
		{"src/*.ts", "src/a.ts", true},
		{"src/*.ts", "src/nested/b.ts", false},
		{"*.go", "main.go", true},
		{"*.go", "cmd/main.go", true},
		{"node_modules", "node_modules", true},
		{"src", "src/a.ts", false},
		{"out/*", "src/a.ts", false},
	}
	for _, c := range casos {
		if got := MatchGlob(c.pattern, c.relPath); got != c.quiere {
			t.Errorf("MatchGlob(%q, %q) = %v, quiere %v", c.pattern, c.relPath, got, c.quiere)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	excludes := []string{"out", "**/migrations/**"}
	if IsExcluded("src/routes/handler.ts", excludes) {
		t.Error("no debe excluir src/routes/handler.ts con exclude 'out'")
	}
	if IsExcluded("layouts/page.ts", excludes) {
		t.Error("no debe excluir layouts/page.ts con exclude 'out'")
	}
	if !IsExcluded("out/bundle.js", excludes) {
		t.Error("debe excluir out/bundle.js")
	}
	if !IsExcluded("db/migrations/2024/b.sql", excludes) {
		t.Error("debe excluir por glob **/migrations/**")
	}
	if IsExcluded("readme.md", nil) {
		t.Error("sin excludes no excluye nada")
	}
}

func TestIsBinaryPath(t *testing.T) {
	if !IsBinaryPath("foto.PNG") {
		t.Error(".PNG debe ser binario por extensión")
	}
	if IsBinaryPath("main.go") {
		t.Error(".go no es binario")
	}
}

func escribir(t *testing.T, nombre string, datos []byte) string {
	t.Helper()
	ruta := filepath.Join(t.TempDir(), nombre)
	if err := os.WriteFile(ruta, datos, 0o644); err != nil {
		t.Fatal(err)
	}
	return ruta
}

func TestIsBinaryContent(t *testing.T) {
	utf16le := escribir(t, "le.env", []byte{0xFF, 0xFE, 'K', 0x00, '=', 0x00})
	if IsBinaryContent(utf16le) {
		t.Error("UTF-16 LE debe tratarse como texto")
	}
	utf16be := escribir(t, "be.env", []byte{0xFE, 0xFF, 0x00, 'K', 0x00, '='})
	if IsBinaryContent(utf16be) {
		t.Error("UTF-16 BE debe tratarse como texto")
	}
	utf8bom := escribir(t, "bom.txt", []byte{0xEF, 0xBB, 0xBF, 'h', 'i'})
	if IsBinaryContent(utf8bom) {
		t.Error("UTF-8 BOM debe tratarse como texto")
	}
	latin1 := escribir(t, "latin1.txt", []byte{'c', 'a', 'f', 0xE9})
	if IsBinaryContent(latin1) {
		t.Error("Latin-1 sin NUL es texto")
	}
	elf := escribir(t, "bin", []byte{0x7F, 'E', 'L', 'F', 0x00, 0x01, 0x00, 0x00})
	if !IsBinaryContent(elf) {
		t.Error("ELF con NUL debe ser binario")
	}
	if IsBinaryContent(filepath.Join(t.TempDir(), "no-existe")) {
		t.Error("archivo inexistente no es binario")
	}
}

func TestCodeExtensions(t *testing.T) {
	ext := CodeExtensions()
	for _, e := range []string{".go", ".ts", ".tsx", ".py", ".rs", ".dart"} {
		if !ext[e] {
			t.Errorf("CodeExtensions debe incluir %q", e)
		}
	}
	if ext[".png"] {
		t.Error("CodeExtensions no debe incluir .png")
	}
}
