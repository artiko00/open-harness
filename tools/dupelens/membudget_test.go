package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

// withAvailable fija la memoria disponible que ve resolveBudget.
func withAvailable(t *testing.T, n uint64, ok bool) {
	t.Helper()
	orig := memAvailable
	memAvailable = func() (uint64, bool) { return n, ok }
	t.Cleanup(func() { memAvailable = orig })
}

func TestParseMemValue(t *testing.T) {
	valid := map[string][2]uint64{"512MiB": {512 * mib, 0}, "2GiB": {2 * gib, 0}, "40%": {0, 40}, "100%": {0, 100}}
	for in, want := range valid {
		b, p, err := parseMemValue(in)
		if err != nil || b != want[0] || p != want[1] {
			t.Errorf("parseMemValue(%q) = %d, %d, %v; want %d, %d", in, b, p, err, want[0], want[1])
		}
	}
	for _, in := range []string{"", "0MiB", "0%", "101%", "mucho", "2GB", "-1GiB", "1.5GiB", "%"} {
		if _, _, err := parseMemValue(in); err == nil || !strings.Contains(err.Error(), "<n>MiB") {
			t.Errorf("parseMemValue(%q) debe fallar explicando el formato; err = %v", in, err)
		}
	}
}

func TestResolveBudget_precedence(t *testing.T) {
	withAvailable(t, 28*gib, true)
	t.Setenv("DUPELENS_MAX_MEMORY", "2GiB")
	cases := []struct {
		flag    string
		flagSet bool
		cfg     string
		want    budget
	}{
		{"512MiB", true, "3GiB", budget{512 * mib, "--max-memory"}},
		{"", false, "3GiB", budget{2 * gib, "DUPELENS_MAX_MEMORY"}},
	}
	for _, c := range cases {
		got, err := resolveBudget(c.flag, c.flagSet, c.cfg, os.Stderr)
		if err != nil || got != c.want {
			t.Errorf("resolveBudget(%q, %v, %q) = %+v, %v; want %+v", c.flag, c.flagSet, c.cfg, got, err, c.want)
		}
	}
	t.Setenv("DUPELENS_MAX_MEMORY", "")
	if got, _ := resolveBudget("", false, "768MiB", os.Stderr); got != (budget{768 * mib, "maxMemory (config)"}) {
		t.Errorf("sin flag ni env, la config debe ganar; got %+v", got)
	}
	if got, _ := resolveBudget("", false, "", os.Stderr); got.bytes != gib || !strings.HasPrefix(got.source, "auto") {
		t.Errorf("sin nada explícito rige auto; got %+v", got)
	}
}

func TestResolveBudget_auto(t *testing.T) {
	cases := []struct {
		avail uint64
		ok    bool
		want  uint64
	}{{28 * gib, true, gib}, {2 * gib, true, 512 * mib}, {0, false, gib}}
	for _, c := range cases {
		withAvailable(t, c.avail, c.ok)
		if got, _ := resolveBudget("", false, "", os.Stderr); got.bytes != c.want {
			t.Errorf("auto con %d disponibles (ok=%v) = %d; want %d", c.avail, c.ok, got.bytes, c.want)
		}
	}
}

func TestResolveBudget_percent(t *testing.T) {
	withAvailable(t, 8*gib, true)
	if got, _ := resolveBudget("40%", true, "", os.Stderr); got.bytes != 8*gib*40/100 {
		t.Errorf("40%% de 8 GiB = %d; want %d (sin el tope de 1 GiB)", got.bytes, 8*gib*40/100)
	}
	withAvailable(t, 0, false)
	var warn strings.Builder
	got, _ := resolveBudget("40%", true, "", &warn)
	if got.bytes != gib || !strings.Contains(warn.String(), "1 GiB") {
		t.Errorf("40%% indeterminable = %d, aviso %q; want 1 GiB con aviso", got.bytes, warn.String())
	}
}

func TestResolveBudget_invalidValueNamesSource(t *testing.T) {
	if _, err := resolveBudget("", false, "mucho", os.Stderr); err == nil || !strings.Contains(err.Error(), "maxMemory") {
		t.Errorf("un valor inválido en config debe nombrar la clave; err = %v", err)
	}
}

func TestFormatBytes(t *testing.T) {
	for n, want := range map[uint64]string{64 * mib: "64 MiB", gib: "1 GiB", 3 * gib: "3 GiB", gib + 512*mib: "1536 MiB"} {
		if got := formatBytes(n); got != want {
			t.Errorf("formatBytes(%d) = %q; want %q", n, got, want)
		}
	}
}

// La clave maxMemory se lee desde cualquier archivo de la cadena de config.
func TestLoadConfig_maxMemoryFromEveryFormat(t *testing.T) {
	files := map[string]string{
		"dupelens.json":  `{"maxMemory": "768MiB"}`,
		"pyproject.toml": "[tool.dupelens]\nmaxMemory = \"768MiB\"\n",
		"package.json":   `{"dupelens": {"maxMemory": "768MiB"}}`,
		"composer.json":  `{"extra": {"open-harness": {"dupelens": {"maxMemory": "768MiB"}}}}`,
	}
	for name, content := range files {
		dir := t.TempDir()
		writeTempFile(t, dir, name, content)
		cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
		if err != nil || cfg.MaxMemory != "768MiB" {
			t.Errorf("%s: MaxMemory = %q, err = %v; want 768MiB", name, cfg.MaxMemory, err)
		}
	}
}

func TestRunCheck_invalidMaxMemoryExits1(t *testing.T) {
	_, stderr, code := captureRunCheck(t, []string{"--dir", t.TempDir(), "--max-memory", "mucho"})
	if code != 1 || !strings.Contains(stderr, "<n>MiB") {
		t.Errorf("exit = %d, stderr = %q; want 1 con el formato válido", code, stderr)
	}
}

func TestRunCheck_overBudgetExits2WithoutPartialReport(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "a.go", blockX("x"))
	fakeMemUsed(t, 1<<40)
	for _, args := range [][]string{{"--fail"}, {}} {
		stdout, stderr, code := captureRunCheck(t, append([]string{"--dir", dir, "--max-memory", "64MiB"}, args...))
		if code != 2 || stdout != "" {
			t.Errorf("args %v: exit = %d, stdout = %q; want 2 y stdout vacío", args, code, stdout)
		}
		for _, want := range []string{"64 MiB", "from --max-memory", "DUPELENS_MAX_MEMORY", "maxMemory", "exclude"} {
			if !strings.Contains(stderr, want) {
				t.Errorf("stderr no menciona %q: %q", want, stderr)
			}
		}
	}
}

func TestApplyBudget_setsAndRestoresSoftLimit(t *testing.T) {
	prev := debug.SetMemoryLimit(-1)
	restore := applyBudget(budget{bytes: 1000 * mib})
	if got := debug.SetMemoryLimit(-1); got != int64(900*mib) {
		t.Errorf("límite blando = %d; want 90%% del presupuesto (%d)", got, 900*mib)
	}
	restore()
	if got := debug.SetMemoryLimit(-1); got != prev {
		t.Errorf("restore dejó %d; want %d", got, prev)
	}
}
