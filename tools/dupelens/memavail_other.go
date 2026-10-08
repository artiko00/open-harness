//go:build !linux && !darwin && !windows

package main

// availableMemory: sin una fuente conocida en esta plataforma, la memoria
// disponible es indeterminable y el presupuesto cae a 1 GiB.
func availableMemory() (uint64, bool) { return 0, false }
