package main

// extensionsForLanguage returns source extensions for a given language
func extensionsForLanguage(lang string) []string {
	if m, ok := mapLanguageExtensions()[lang]; ok {
		return m.extensions
	}
	return nil
}

// supportedLanguages returns list of all supported language keys, en ORDEN FIJO.
// El orden es el criterio de desempate de la detección automática (detect.go).
func supportedLanguages() []string {
	return []string{
		"go", "typescript", "javascript", "python",
		"ruby", "rust", "java", "kotlin", "csharp", "dart",
	}
}

// esLenguajeSoportado indica si lang es una clave conocida (no incluye "auto").
func esLenguajeSoportado(lang string) bool {
	for _, l := range supportedLanguages() {
		if l == lang {
			return true
		}
	}
	return false
}
