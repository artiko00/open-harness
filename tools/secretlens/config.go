package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Patterns  []PatternRule `json:"patterns"`
	Allowlist []string      `json:"allowlist"`
	Exclude   []string      `json:"exclude"`
}

type PatternRule struct {
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Severity string `json:"severity"`
}

var defaultConfig = Config{
	Patterns:  defaultPatterns(),
	Allowlist: []string{"example", "placeholder", "your_key_here", "changeme", "xxxx", "****"},
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
			return defaultConfig, nil
		}
		return defaultConfig, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig, err
	}

	if len(cfg.Patterns) == 0 {
		cfg.Patterns = defaultPatterns()
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = defaultConfig.Exclude
	}
	if len(cfg.Allowlist) == 0 {
		cfg.Allowlist = defaultConfig.Allowlist
	}

	return cfg, nil
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
