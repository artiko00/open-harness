package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/artiko00/open-harness/tools/_shared/configload"
)

// defaultMinEntropy es el umbral de entropía de Shannon (bits/carácter)
// aplicado SOLO a las reglas genéricas KEY=VALUE. Calibrado contra el fixture
// de auditoría (testdata/audit): recall 20/20, precision 100%.
const defaultMinEntropy = 3.0

type Config struct {
	Patterns               []PatternRule `json:"patterns"`
	Allowlist              []string      `json:"allowlist"`
	Exclude                []string      `json:"exclude"`
	MinEntropy             float64       `json:"minEntropy"`
	DisableDefaultPatterns bool          `json:"disableDefaultPatterns"`
}

type PatternRule struct {
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Severity string `json:"severity"`
	// EntropyGate marca las reglas genéricas cuyo valor capturado debe superar
	// MinEntropy para reportarse. Las reglas de prefijo fuerte lo dejan en false.
	EntropyGate bool `json:"entropyGate"`
}

var defaultConfig = Config{
	Patterns:   defaultPatterns(),
	Allowlist:  []string{"example", "placeholder", "your_key_here", "changeme", "xxxx", "****"},
	MinEntropy: defaultMinEntropy,
	Exclude: []string{
		"node_modules", "vendor", ".git", "dist", "build",
		"coverage", "__pycache__", "target", ".next", "out",
		".cache", "*.lock", "go.sum",
	},
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return loadConfigFallbackChain(filepath.Dir(path))
		}
		return defaultConfig, err
	}

	var cfg Config
	unknown, err := configload.Strict(data, &cfg)
	if err != nil {
		return defaultConfig, fmt.Errorf("config %q: %w", path, err)
	}
	if unknown != "" {
		fmt.Fprintf(os.Stderr, "warning: config %q: clave desconocida %q (ignorada)\n", path, unknown)
	}
	applyConfigDefaults(&cfg)
	return cfg, nil
}

// applyConfigDefaults combina defaults. Los patterns son ADITIVOS: se antepone
// defaultPatterns() a los custom, salvo que DisableDefaultPatterns sea true.
func applyConfigDefaults(cfg *Config) {
	if !cfg.DisableDefaultPatterns {
		cfg.Patterns = append(defaultPatterns(), cfg.Patterns...)
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = defaultConfig.Exclude
	}
	if len(cfg.Allowlist) == 0 {
		cfg.Allowlist = defaultConfig.Allowlist
	}
	if cfg.MinEntropy == 0 {
		cfg.MinEntropy = defaultMinEntropy
	}
}

func defaultConfigJSON() string {
	return `{
  "patterns": [],
  "allowlist": [
    "example", "placeholder", "your_key_here", "changeme", "xxxx", "****"
  ],
  "exclude": [
    "node_modules", "vendor", ".git", "dist", "build",
    "coverage", "__pycache__", "target", ".next", "out"
  ]
}
`
}
