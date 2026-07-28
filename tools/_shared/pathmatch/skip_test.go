package pathmatch

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// entradas lee el directorio y devuelve sus os.DirEntry indexados por nombre.
func entradas(t *testing.T, dir string) map[string]os.DirEntry {
	t.Helper()
	leidas, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	res := make(map[string]os.DirEntry, len(leidas))
	for _, e := range leidas {
		res[e.Name()] = e
	}
	return res
}

func TestIsRegularArchivoNormal(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	e := entradas(t, dir)
	if !IsRegular(e["a.go"]) {
		t.Error("un archivo regular debe dar IsRegular true")
	}
	if IsRegular(e["sub"]) {
		t.Error("un directorio no es un archivo regular")
	}
}

func TestIsRegularFIFO(t *testing.T) {
	dir := t.TempDir()
	pipe := filepath.Join(dir, "pipe.go")
	if err := syscall.Mkfifo(pipe, 0o644); err != nil {
		t.Skipf("no se pudo crear el FIFO: %v", err)
	}

	info, err := os.Lstat(pipe)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeNamedPipe == 0 {
		t.Fatal("el archivo creado no es un named pipe")
	}

	if IsRegular(entradas(t, dir)["pipe.go"]) {
		t.Error("un FIFO no debe considerarse archivo regular")
	}
}

func TestFailsGatePorMotivo(t *testing.T) {
	casos := []struct {
		motivo string
		quiere bool
	}{
		{ReasonReadError, true},
		{ReasonLineTooLong, true},
		{ReasonNotRegular, false},
		{ReasonBinary, false},
		{"motivo desconocido", false},
	}
	for _, c := range casos {
		if got := FailsGate(c.motivo); got != c.quiere {
			t.Errorf("FailsGate(%q) = %v, quiere %v", c.motivo, got, c.quiere)
		}
	}
}

func TestAnyFailsGate(t *testing.T) {
	if AnyFailsGate(nil) {
		t.Error("sin omisiones no se rompe el gate")
	}
	benignos := []Skip{
		{Path: "a/pipe.go", Reason: ReasonNotRegular},
		{Path: "b/logo.png", Reason: ReasonBinary},
	}
	if AnyFailsGate(benignos) {
		t.Error("binarios y no regulares no deben romper el gate")
	}
	mezcla := append(benignos, Skip{Path: "c/vendor.min.js", Reason: ReasonLineTooLong})
	if !AnyFailsGate(mezcla) {
		t.Error("una línea demasiado larga debe romper el gate")
	}
	if !AnyFailsGate([]Skip{{Path: "d/priv.env", Reason: ReasonReadError}}) {
		t.Error("un error de lectura debe romper el gate")
	}
}

func TestMotivosYBufferLimit(t *testing.T) {
	if ReasonNotRegular != "not a regular file" ||
		ReasonBinary != "binary" ||
		ReasonReadError != "read error" ||
		ReasonLineTooLong != "line exceeds buffer limit" {
		t.Error("los motivos canónicos cambiaron de texto")
	}
	if BufferLimit != 4*1024*1024 {
		t.Errorf("BufferLimit = %d, quiere 4 MB", BufferLimit)
	}
	s := Skip{Path: "a.go", Reason: ReasonBinary}
	if s.Path != "a.go" || s.Reason != ReasonBinary {
		t.Error("Skip debe exponer Path y Reason")
	}
}
