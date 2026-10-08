//go:build linux

package main

// availableMemory combina MemAvailable del sistema con el margen del cgroup
// propio (contenedores, CI, scopes de systemd): el menor de ambos.
func availableMemory() (uint64, bool) {
	sys, sysOK := parseMeminfo(readText("/proc/meminfo"))
	cg, cgOK := cgroupHeadroom(readText, readText("/proc/self/cgroup"))
	return minAvailable(sys, sysOK, cg, cgOK)
}
