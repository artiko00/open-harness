package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// cfgSkips es la config mínima usada por los tests de omisiones.
func cfgSkips() Config {
	return Config{Default: DefaultConfig{MinTokens: 5, MinLines: 1}}
}

// buscaSkip devuelve el motivo registrado para path, o "" si no aparece.
func buscaSkip(skips []pathmatch.Skip, path string) string {
	for _, s := range skips {
		if s.Path == path {
			return s.Reason
		}
	}
	return ""
}

// escribeConfigTmp deja un dupelens.json usable en dir y devuelve su ruta.
func escribeConfigTmp(t *testing.T, dir string) string {
	t.Helper()
	p := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(p, []byte(`{"default":{"minTokens":5,"minLines":1},"exclude":[".git"]}`), 0644); err != nil {
		t.Fatalf("escribiendo config: %v", err)
	}
	return p
}

// capturaStdout ejecuta fn con os.Stdout redirigido y devuelve lo impreso.
func capturaStdout(t *testing.T, fn func() int) (string, int) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	code := fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), code
}

func TestScan_FIFOSeOmiteComoNoRegular(t *testing.T) {
	tmpDir := t.TempDir()
	fifo := filepath.Join(tmpDir, "pipe.go")
	if err := syscall.Mkfifo(fifo, 0644); err != nil {
		t.Skipf("el sistema no soporta FIFO: %v", err)
	}
	_, _, skips, err := scan(tmpDir, cfgSkips(), 0)
	if err != nil {
		t.Fatalf("scan falló: %v", err)
	}
	if got := buscaSkip(skips, "pipe.go"); got != pathmatch.ReasonNotRegular {
		t.Errorf("motivo del FIFO = %q, quiere %q", got, pathmatch.ReasonNotRegular)
	}
}

func TestScan_ArchivoIlegibleSeOmiteComoReadError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod no bloquea como root")
	}
	tmpDir := t.TempDir()
	locked := filepath.Join(tmpDir, "locked.go")
	os.WriteFile(locked, []byte("package main"), 0644)
	os.Chmod(locked, 0000)
	defer os.Chmod(locked, 0644)

	_, _, skips, err := scan(tmpDir, cfgSkips(), 0)
	if err != nil {
		t.Fatalf("scan falló: %v", err)
	}
	if got := buscaSkip(skips, "locked.go"); got != pathmatch.ReasonReadError {
		t.Errorf("motivo del ilegible = %q, quiere %q", got, pathmatch.ReasonReadError)
	}
	if !pathmatch.AnyFailsGate(skips) {
		t.Error("un read error debe romper el gate")
	}
}

func TestScan_BinarioSeOmiteComoBinary(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "img.jpg"), []byte{0xFF, 0xD8, 0xFF}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "blob.dat"), []byte{'a', 0x00, 'b'}, 0644)

	_, _, skips, err := scan(tmpDir, cfgSkips(), 0)
	if err != nil {
		t.Fatalf("scan falló: %v", err)
	}
	if got := buscaSkip(skips, "img.jpg"); got != pathmatch.ReasonBinary {
		t.Errorf("motivo de img.jpg = %q, quiere %q", got, pathmatch.ReasonBinary)
	}
	if got := buscaSkip(skips, "blob.dat"); got != pathmatch.ReasonBinary {
		t.Errorf("motivo de blob.dat = %q, quiere %q", got, pathmatch.ReasonBinary)
	}
	if pathmatch.AnyFailsGate(skips) {
		t.Error("los binarios no deben romper el gate")
	}
}

func TestRunCheck_ReadErrorConFailDaExit1YSinOK(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod no bloquea como root")
	}
	tmpDir := t.TempDir()
	cfg := escribeConfigTmp(t, tmpDir)
	locked := filepath.Join(tmpDir, "locked.go")
	os.WriteFile(locked, []byte("package main"), 0644)
	os.Chmod(locked, 0000)
	defer os.Chmod(locked, 0644)

	out, code := capturaStdout(t, func() int {
		return runCheck([]string{"--dir", tmpDir, "--config", cfg, "--fail", "--no-color"})
	})
	if code != 1 {
		t.Errorf("exit code = %d, quiere 1", code)
	}
	if strings.Contains(out, "OK:") {
		t.Errorf("la salida no debe contener \"OK:\": %s", out)
	}
	if !strings.Contains(out, "SKIPPED") || !strings.Contains(out, pathmatch.ReasonReadError) {
		t.Errorf("la salida debe listar el read error bajo SKIPPED: %s", out)
	}
	if !strings.Contains(out, "1 skipped") {
		t.Errorf("el resumen debe contar los omitidos: %s", out)
	}
}

func TestRunCheck_BinarioConFailNoCambiaExitCode(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := escribeConfigTmp(t, tmpDir)
	os.WriteFile(filepath.Join(tmpDir, "img.jpg"), []byte{0xFF, 0xD8, 0xFF}, 0644)

	out, code := capturaStdout(t, func() int {
		return runCheck([]string{"--dir", tmpDir, "--config", cfg, "--fail", "--no-color"})
	})
	if code != 0 {
		t.Errorf("exit code = %d, quiere 0", code)
	}
	if !strings.Contains(out, "SKIPPED") || !strings.Contains(out, pathmatch.ReasonBinary) {
		t.Errorf("la salida debe listar el binario bajo SKIPPED: %s", out)
	}
	if !strings.Contains(out, "OK:") || !strings.Contains(out, "(1 skipped)") {
		t.Errorf("la salida debe ser OK con el conteo de omitidos: %s", out)
	}
}

func TestReportConsole_SinMatchesConSkipQueRompeGate_NoImprimeOK(t *testing.T) {
	var buf bytes.Buffer
	skips := []pathmatch.Skip{{Path: "a.go", Reason: pathmatch.ReasonReadError}}
	reportConsole(nil, ReportOpts{ScannedCount: 3, NoColor: true, Skips: skips}, &buf)
	out := buf.String()
	if strings.Contains(out, "OK:") {
		t.Errorf("no debe imprimir OK: %s", out)
	}
	if !strings.Contains(out, "SUMMARY:") || !strings.Contains(out, "1 skipped") {
		t.Errorf("debe imprimir SUMMARY con el conteo: %s", out)
	}
	if !strings.Contains(out, "SKIPPED (1 file(s) not analyzed):") {
		t.Errorf("encabezado SKIPPED incorrecto: %s", out)
	}
	if !strings.Contains(out, "  a.go    read error") {
		t.Errorf("entrada SKIPPED incorrecta: %s", out)
	}
}

func TestReportConsole_ConMatchesYSkips_SummaryEsLaUltimaLinea(t *testing.T) {
	var buf bytes.Buffer
	skips := []pathmatch.Skip{{Path: "x.png", Reason: pathmatch.ReasonBinary}}
	reportConsole(sampleMatches(), ReportOpts{ScannedCount: 10, NoColor: true, Skips: skips}, &buf)
	lineas := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	ultima := lineas[len(lineas)-1]
	if !strings.HasPrefix(ultima, "SUMMARY:") || !strings.HasSuffix(ultima, ", 1 skipped") {
		t.Errorf("última línea = %q", ultima)
	}
}

func TestReportConsole_SinSkips_NoImprimeSeccionNiConteo(t *testing.T) {
	var buf bytes.Buffer
	reportConsole(nil, ReportOpts{ScannedCount: 3, NoColor: true}, &buf)
	out := buf.String()
	if strings.Contains(out, "SKIPPED") || strings.Contains(out, "skipped") {
		t.Errorf("sin omitidos no debe mencionarlos: %s", out)
	}
	if !strings.Contains(out, "OK:") {
		t.Errorf("debe imprimir OK: %s", out)
	}
}

func TestReportJSON_IncluyeSkipped(t *testing.T) {
	var buf bytes.Buffer
	skips := []pathmatch.Skip{{Path: "a.go", Reason: pathmatch.ReasonReadError}}
	reportJSON(nil, ReportOpts{ScannedCount: 1, Skips: skips}, &buf)

	var parsed struct {
		Skipped []struct {
			Path   string `json:"path"`
			Reason string `json:"reason"`
		} `json:"skipped"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if len(parsed.Skipped) != 1 || parsed.Skipped[0].Path != "a.go" ||
		parsed.Skipped[0].Reason != pathmatch.ReasonReadError {
		t.Errorf("skipped incorrecto: %+v", parsed.Skipped)
	}
}

func TestReportJSON_SinSkips_ArrayVacio(t *testing.T) {
	var buf bytes.Buffer
	reportJSON(nil, ReportOpts{}, &buf)
	if !strings.Contains(buf.String(), `"skipped": []`) {
		t.Errorf("skipped debe serializarse como array vacío: %s", buf.String())
	}
}
