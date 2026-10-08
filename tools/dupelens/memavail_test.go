package main

import (
	"path/filepath"
	"testing"
)

func TestParseMeminfo(t *testing.T) {
	text := "MemTotal:       49213440 kB\nMemFree:        1000 kB\nMemAvailable:   28672000 kB\n"
	if got, ok := parseMeminfo(text); !ok || got != 28672000*1024 {
		t.Errorf("parseMeminfo = %d, %v; want %d", got, ok, 28672000*1024)
	}
	for _, bad := range []string{"", "MemTotal: 1 kB\n", "MemAvailable: muchos kB\n"} {
		if _, ok := parseMeminfo(bad); ok {
			t.Errorf("parseMeminfo(%q) debe fallar", bad)
		}
	}
}

func TestHeadroom(t *testing.T) {
	cases := []struct {
		max, cur string
		want     uint64
		ok       bool
	}{{"1073741824\n", "629145600\n", 1073741824 - 629145600, true}, {"max\n", "1\n", 0, false},
		{"", "", 0, false}, {"100", "200", 0, true}}
	for _, c := range cases {
		if got, ok := headroom(c.max, c.cur); got != c.want || ok != c.ok {
			t.Errorf("headroom(%q, %q) = %d, %v; want %d, %v", c.max, c.cur, got, ok, c.want, c.ok)
		}
	}
}

func TestCgroupLocation(t *testing.T) {
	cases := map[string]cgroupLoc{
		"0::/user.slice/app.scope\n": {"/sys/fs/cgroup", "/user.slice/app.scope", "memory.max", "memory.current"},
		"5:cpu:/x\n4:memory,blkio:/docker/abc\n0::/\n": {"/sys/fs/cgroup/memory", "/docker/abc",
			"memory.limit_in_bytes", "memory.usage_in_bytes"},
	}
	for in, want := range cases {
		if got, ok := cgroupLocation(in); !ok || got != want {
			t.Errorf("cgroupLocation(%q) = %+v, %v; want %+v", in, got, ok, want)
		}
	}
	if _, ok := cgroupLocation("5:cpu:/x\nbasura\n"); ok {
		t.Error("sin línea v2 ni controlador memory no hay cgroup de memoria")
	}
}

func TestCgroupHeadroom_takesTightestAncestor(t *testing.T) {
	fs := map[string]string{
		"/sys/fs/cgroup/a/b/memory.max": "max", "/sys/fs/cgroup/a/b/memory.current": "10",
		"/sys/fs/cgroup/a/memory.max": "1000", "/sys/fs/cgroup/a/memory.current": "600",
		"/sys/fs/cgroup/memory.max": "5000", "/sys/fs/cgroup/memory.current": "100",
	}
	read := func(p string) string { return fs[p] }
	if got, ok := cgroupHeadroom(read, "0::/a/b\n"); !ok || got != 400 {
		t.Errorf("cgroupHeadroom = %d, %v; want 400 (el margen del padre)", got, ok)
	}
	if _, ok := cgroupHeadroom(read, "0::/../fuera\n"); ok {
		t.Error("una ruta fuera de la jerarquía no debe dar margen")
	}
	if _, ok := cgroupHeadroom(read, ""); ok {
		t.Error("sin /proc/self/cgroup no hay margen")
	}
}

func TestMinAvailable(t *testing.T) {
	cases := []struct {
		sys          uint64
		sysOK        bool
		cg           uint64
		cgOK, wantOK bool
		want         uint64
	}{{8, true, 3, true, true, 3}, {8, true, 0, false, true, 8}, {0, false, 3, true, true, 3}, {0, false, 0, false, false, 0}}
	for _, c := range cases {
		if got, ok := minAvailable(c.sys, c.sysOK, c.cg, c.cgOK); got != c.want || ok != c.wantOK {
			t.Errorf("minAvailable(%+v) = %d, %v", c, got, ok)
		}
	}
}

func TestReadText(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "f", "hola")
	if readText(filepath.Join(dir, "f")) != "hola" || readText(filepath.Join(dir, "no")) != "" {
		t.Error("readText debe devolver el contenido o vacío si no existe")
	}
}

// availableMemory en la plataforma actual no debe fallar ni devolver 0 como
// valor válido.
func TestAvailableMemory_currentPlatform(t *testing.T) {
	if n, ok := availableMemory(); ok && n == 0 {
		t.Error("una memoria disponible válida no puede ser 0")
	}
}
