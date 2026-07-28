package pathmatch

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

var binaryExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".svg": true, ".ico": true, ".bmp": true, ".tiff": true,
	".mp3": true, ".mp4": true, ".wav": true, ".ogg": true, ".webm": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
	".exe": true, ".bin": true, ".dll": true, ".so": true, ".dylib": true,
	".wasm": true, ".class": true, ".pyc": true,
	".ttf": true, ".woff": true, ".woff2": true, ".eot": true,
	".db": true, ".sqlite": true,
}

// IsBinaryPath decide por extensión si el path es un binario conocido.
func IsBinaryPath(path string) bool {
	return binaryExtensions[strings.ToLower(filepath.Ext(path))]
}

// IsBinaryContent inspecciona los primeros bytes del archivo. Un BOM UTF-8 o
// UTF-16 marca el contenido como texto (un .env en UTF-16 no es binario);
// sin BOM, la presencia de un byte NUL indica binario.
func IsBinaryContent(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	b := buf[:n]
	if tieneBOMTexto(b) {
		return false
	}
	return bytes.ContainsRune(b, 0)
}

// tieneBOMTexto detecta los BOM de UTF-8 y UTF-16 (LE/BE).
func tieneBOMTexto(b []byte) bool {
	switch {
	case len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF:
		return true
	case len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE:
		return true
	case len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF:
		return true
	}
	return false
}
