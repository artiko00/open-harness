package main

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func TestShannonEntropy_Empty(t *testing.T) {
	if e := shannonEntropy(""); e != 0 {
		t.Errorf("entropía de \"\" = %v, want 0", e)
	}
}

func TestShannonEntropy_UniformVsRandom(t *testing.T) {
	if e := shannonEntropy("aaaaaaaa"); e != 0 {
		t.Errorf("entropía uniforme = %v, want 0", e)
	}
	if e := shannonEntropy("aB3xK9mQ2pL7wZ4tR8n"); e < 3.0 {
		t.Errorf("entropía de secreto = %v, want >= 3.0", e)
	}
}

func TestDecodeContent_Passthrough(t *testing.T) {
	in := []byte("API_KEY=abc")
	if got := decodeContent(in); string(got) != "API_KEY=abc" {
		t.Errorf("passthrough alterado: %q", got)
	}
	if got := decodeContent([]byte{0xFF}); len(got) != 1 {
		t.Errorf("dato corto debe devolverse igual, got %q", got)
	}
}

func utf16BOM(s string, big bool) []byte {
	u := utf16.Encode([]rune(s))
	var b []byte
	if big {
		b = []byte{0xFE, 0xFF}
		for _, c := range u {
			b = append(b, byte(c>>8), byte(c))
		}
		return b
	}
	b = []byte{0xFF, 0xFE}
	for _, c := range u {
		b = append(b, byte(c), byte(c>>8))
	}
	return b
}

func TestDecodeContent_UTF16(t *testing.T) {
	if got := decodeContent(utf16BOM("hola", false)); string(got) != "hola" {
		t.Errorf("UTF-16 LE = %q, want hola", got)
	}
	if got := decodeContent(utf16BOM("hola", true)); string(got) != "hola" {
		t.Errorf("UTF-16 BE = %q, want hola", got)
	}
}

// TAREA 5.18 — un secreto en un .env UTF-16 (LE y BE) debe detectarse.
func TestScanFile_UTF16Detecta(t *testing.T) {
	compiled, _ := compilePatterns(defaultPatterns())
	cfg := auditConfig()
	for _, big := range []bool{false, true} {
		dir := t.TempDir()
		path := filepath.Join(dir, "secrets.env")
		os.WriteFile(path, utf16BOM("API_KEY=aB3xK9mQ2pL7wZ4tR8n\n", big), 0644)
		findings, reason := scanFile(path, "secrets.env", compiled, cfg)
		if reason != "" {
			t.Fatalf("big=%v motivo inesperado %q", big, reason)
		}
		if len(findings) == 0 {
			t.Errorf("big=%v: esperaba detectar el secreto UTF-16", big)
		}
	}
}

func TestIsDanglingKey(t *testing.T) {
	if !isDanglingKey(`  "private_key":`) {
		t.Error("clave partida debe ser dangling")
	}
	if isDanglingKey(`API_KEY=aB3xK9mQ2pL7wZ4tR8n`) {
		t.Error("asignación completa NO es dangling")
	}
}

func TestExtractValue(t *testing.T) {
	if v := extractValue([]string{"AKIA0000000000000000"}); v != "AKIA0000000000000000" {
		t.Errorf("sin grupos debe usar match completo, got %q", v)
	}
	if v := extractValue([]string{"k=v", "", "valor"}); v != "valor" {
		t.Errorf("debe tomar el último grupo no vacío, got %q", v)
	}
}

// TAREA 5.3 — KEY=VALUE sin comillas se detecta; DEBUG=true no.
func TestScanFile_UnquotedAssignment(t *testing.T) {
	compiled, _ := compilePatterns(defaultPatterns())
	cfg := auditConfig()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.env")
	os.WriteFile(path, []byte("API_KEY=aB3xK9mQ2pL7wZ4tR8n\nDEBUG=true\n"), 0644)
	findings, _ := scanFile(path, "app.env", compiled, cfg)
	if len(findings) != 1 {
		t.Fatalf("esperaba 1 hallazgo (solo API_KEY), got %d", len(findings))
	}
}

// TAREA 5.12 — un valor de baja entropía se descarta en reglas genéricas.
func TestScanFile_EntropyGateFilters(t *testing.T) {
	compiled, _ := compilePatterns(defaultPatterns())
	cfg := auditConfig()
	dir := t.TempDir()
	path := filepath.Join(dir, "low.env")
	os.WriteFile(path, []byte("api_key=aaaaaaaa\n"), 0644)
	findings, _ := scanFile(path, "low.env", compiled, cfg)
	if len(findings) != 0 {
		t.Errorf("valor de entropía 0 debe filtrarse, got %d", len(findings))
	}
}

// TAREA 5.6 — AKIA con comentario "example" SÍ se reporta (allowlist por valor).
func TestRunCheck_ValorReportadoPeseAExampleEnLinea(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "aws.tf"),
		[]byte(`AWS_KEY = "AKIAQNB7Q7AIIVSMBGPF"  # see example above`+"\n"), 0644)
	var code int
	capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1 (el valor no está en allowlist)", code)
	}
}

// TAREA 5.7 — API_KEY=your_key_here NO se reporta (allowlist por valor).
func TestRunCheck_PlaceholderNoReportado(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.env"), []byte("API_KEY=your_key_here\n"), 0644)
	var code int
	capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
}

// TAREA 5.8 — URI con credenciales SÍ se reporta pese al host example.com.
func TestRunCheck_URIConCredencialesReportada(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "db.env"),
		[]byte("DB=postgres://u:SuperSecretPass99@example.com/db\n"), 0644)
	var code int
	capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
}

// TAREA 5.9 — pattern custom NO desactiva built-in; disableDefaultPatterns:true sí.
func TestLoadConfig_PatternsAditivosYDisable(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "additive.json")
	os.WriteFile(custom, []byte(`{"patterns":[{"name":"c","pattern":"FOO_\\d+","severity":"low"}]}`), 0644)
	cfg, err := loadConfig(custom)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Patterns) != len(defaultPatterns())+1 {
		t.Errorf("aditivo: esperaba defaults+1, got %d", len(cfg.Patterns))
	}

	only := filepath.Join(dir, "only.json")
	os.WriteFile(only, []byte(`{"disableDefaultPatterns":true,"patterns":[{"name":"c","pattern":"FOO_\\d+","severity":"low"}]}`), 0644)
	cfg2, err := loadConfig(only)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg2.Patterns) != 1 {
		t.Errorf("disable: esperaba solo 1 pattern custom, got %d", len(cfg2.Patterns))
	}
}
