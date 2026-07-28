package pathmatch

import "os"

// Skip describe un archivo que no se pudo analizar.
type Skip struct {
	Path   string
	Reason string
}

// Motivos canónicos de omisión.
const (
	ReasonNotRegular  = "not a regular file"
	ReasonBinary      = "binary"
	ReasonReadError   = "read error"
	ReasonLineTooLong = "line exceeds buffer limit"
)

// BufferLimit es el tamaño máximo de línea al escanear (4 MB).
const BufferLimit = 4 * 1024 * 1024

// IsRegular indica si la entrada del walk es un archivo regular. Descarta
// directorios, FIFOs, sockets y dispositivos, que colgarían o fallarían al
// abrirse. Pensada para usarse directo con el os.DirEntry de filepath.WalkDir.
func IsRegular(d os.DirEntry) bool {
	return d.Type().IsRegular()
}

// FailsGate indica si un motivo de omisión debe romper --fail. Solo lo hacen
// los motivos que ocultan contenido no analizado: error de lectura y línea
// que excede el buffer.
func FailsGate(reason string) bool {
	return reason == ReasonReadError || reason == ReasonLineTooLong
}

// AnyFailsGate indica si alguno de los skips rompe el gate.
func AnyFailsGate(skips []Skip) bool {
	for _, s := range skips {
		if FailsGate(s.Reason) {
			return true
		}
	}
	return false
}
