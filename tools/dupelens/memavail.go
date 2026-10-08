package main

import (
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

// Funciones puras y portables para leer la memoria disponible en Linux. Los
// adaptadores por plataforma (memavail_*.go) solo leen y delegan aquí, para que
// toda la lógica quede testeada en cualquier SO.

// readText devuelve el contenido de un archivo o "" si no se puede leer.
func readText(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}

// parseMeminfo extrae MemAvailable (en kB) de /proc/meminfo y lo da en bytes.
func parseMeminfo(text string) (uint64, bool) {
	for _, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 && f[0] == "MemAvailable:" {
			n, err := strconv.ParseUint(f[1], 10, 64)
			return n * 1024, err == nil
		}
	}
	return 0, false
}

// cgroupLoc ubica los archivos de límite y uso del cgroup de memoria propio.
type cgroupLoc struct{ base, rel, maxFile, curFile string }

// cgroupLocation interpreta /proc/self/cgroup. Prefiere el controlador memory
// de v1 (en modo híbrido la línea "0::" no tiene ese controlador) y si no, v2.
func cgroupLocation(selfCgroup string) (cgroupLoc, bool) {
	v2, found := cgroupLoc{}, false
	for _, line := range strings.Split(selfCgroup, "\n") {
		parts := strings.SplitN(line, ":", 3)
		switch {
		case len(parts) != 3:
		case slices.Contains(strings.Split(parts[1], ","), "memory"):
			return cgroupLoc{"/sys/fs/cgroup/memory", parts[2], "memory.limit_in_bytes", "memory.usage_in_bytes"}, true
		case parts[0] == "0" && parts[1] == "":
			v2, found = cgroupLoc{"/sys/fs/cgroup", parts[2], "memory.max", "memory.current"}, true
		}
	}
	return v2, found
}

// cgroupHeadroom devuelve el menor margen (límite − uso) entre el cgroup propio
// y sus ancestros: un límite puede vivir en cualquier nivel (slice, scope,
// contenedor). Una ruta fuera de la jerarquía montada no da margen.
func cgroupHeadroom(read func(string) string, selfCgroup string) (uint64, bool) {
	loc, ok := cgroupLocation(selfCgroup)
	if !ok {
		return 0, false
	}
	best, found := uint64(0), false
	for dir := path.Join(loc.base, loc.rel); strings.HasPrefix(dir, loc.base); dir = path.Dir(dir) {
		if h, ok := headroom(read(dir+"/"+loc.maxFile), read(dir+"/"+loc.curFile)); ok && (!found || h < best) {
			best, found = h, true
		}
		if dir == loc.base {
			break
		}
	}
	return best, found
}

// headroom calcula límite − uso. "max" o un archivo ausente = sin límite.
func headroom(maxText, curText string) (uint64, bool) {
	limit, err1 := strconv.ParseUint(strings.TrimSpace(maxText), 10, 64)
	used, err2 := strconv.ParseUint(strings.TrimSpace(curText), 10, 64)
	if err1 != nil || err2 != nil {
		return 0, false
	}
	return limit - min(used, limit), true
}

// minAvailable combina la memoria del sistema y el margen del cgroup: el menor
// de los que se pudieron determinar.
func minAvailable(sys uint64, sysOK bool, cg uint64, cgOK bool) (uint64, bool) {
	switch {
	case sysOK && cgOK:
		return min(sys, cg), true
	case sysOK:
		return sys, true
	}
	return cg, cgOK
}
