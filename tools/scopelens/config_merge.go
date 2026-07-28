package main

// mergeConfig completa los campos aún sin definir de dst con los de src. Para
// cada campo gana el primer archivo de la cadena que lo aporte (no-cero /
// no-vacío); el array Exclude es atómico.
func mergeConfig(dst *Config, src Config) {
	if dst.MaxFiles == 0 {
		dst.MaxFiles = src.MaxFiles
	}
	if dst.Base == "" {
		dst.Base = src.Base
	}
	if !dst.ExcludeTests {
		dst.ExcludeTests = src.ExcludeTests
	}
	if len(dst.Exclude) == 0 {
		dst.Exclude = src.Exclude
	}
}

func defaultConfigJSON() string {
	return `{
  "maxFiles": 15,
  "base": "",
  "excludeTests": false,
  "exclude": [
    ".git/**",
    "node_modules/**",
    "vendor/**",
    "dist/**",
    "build/**",
    "coverage/**",
    "package-lock.json",
    "pnpm-lock.yaml",
    "yarn.lock",
    ".next/**",
    ".nuxt/**",
    "out/**",
    "**/__snapshots__/**",
    "poetry.lock",
    "Pipfile.lock",
    "uv.lock",
    "**/__pycache__/**",
    "*.egg-info/**",
    ".venv/**",
    "go.sum",
    "**/*.pb.go",
    "**/zz_generated*.go"
  ]
}
`
}
