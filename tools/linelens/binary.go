package main

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

func isBinaryPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return binaryExtensions[ext]
}

func isBinaryContent(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	return bytes.ContainsRune(buf[:n], 0)
}
