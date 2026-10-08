//go:build darwin

package main

import (
	"encoding/binary"
	"syscall"
)

// availableMemory devuelve la RAM total (hw.memsize): macOS no expone la
// memoria disponible sin cgo. syscall.Sysctl recorta el NUL final del valor
// binario, así que se rellena a 8 bytes antes de decodificarlo.
func availableMemory() (uint64, bool) {
	s, err := syscall.Sysctl("hw.memsize")
	if err != nil {
		return 0, false
	}
	b := make([]byte, 8)
	copy(b, s)
	return binary.LittleEndian.Uint64(b), true
}
