package pathmatch

var codeExtensions = map[string]bool{
	".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
	".mjs": true, ".cjs": true, ".py": true, ".rb": true, ".rs": true,
	".java": true, ".php": true, ".c": true, ".cc": true, ".cpp": true,
	".h": true, ".hpp": true, ".cs": true, ".dart": true, ".kt": true,
	".kts": true, ".swift": true, ".scala": true, ".m": true, ".mm": true,
}

// CodeExtensions devuelve el conjunto compartido de extensiones de código
// fuente (con el punto inicial, en minúsculas). El mapa se comparte; los
// consumidores no deben mutarlo.
func CodeExtensions() map[string]bool {
	return codeExtensions
}
