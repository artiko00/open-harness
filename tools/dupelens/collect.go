package main

import (
	"path/filepath"
	"strings"
)

// addFile tokeniza el contenido, calcula fingerprints crudos y normalizados y
// registra el archivo solo si produjo fingerprints.
func addFile(files *[]fileData, rawFps, normFps *[]Fingerprint, relPath, content string, windowSize int) {
	ext := strings.ToLower(filepath.Ext(relPath))
	raw := tokenize(content, ext)
	norm := normalizeTokens(raw)
	fileID := len(*files)
	rfp := fingerprintCode(raw, fileID, windowSize)
	nfp := fingerprintNormalized(norm, fileID, windowSize)
	if len(rfp) == 0 && len(nfp) == 0 {
		return
	}
	*files = append(*files, fileData{name: relPath, raw: raw, norm: norm})
	*rawFps = append(*rawFps, rfp...)
	*normFps = append(*normFps, nfp...)
}
