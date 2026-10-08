//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// memoryStatusEx refleja MEMORYSTATUSEX de la API de Windows.
type memoryStatusEx struct {
	length, memoryLoad                                 uint32
	totalPhys, availPhys, totalPageFile, availPageFile uint64
	totalVirtual, availVirtual, availExtendedVirtual   uint64
}

// availableMemory devuelve la memoria física disponible (ullAvailPhys).
func availableMemory() (uint64, bool) {
	st := memoryStatusEx{}
	st.length = uint32(unsafe.Sizeof(st))
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	if r, _, _ := proc.Call(uintptr(unsafe.Pointer(&st))); r == 0 {
		return 0, false
	}
	return st.availPhys, true
}
