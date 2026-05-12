package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Default DefaultConfig `json:"default"`
	Rules   []Rule        `json:"rules"`
	Exclude []string      `json:"exclude"`
}

type DefaultConfig struct {
	MaxLines int `json:"maxLines"`
}

type Rule struct {
	Pattern  string `json:"pattern"`
	MaxLines int    `json:"maxLines"`
	Skip     bool   `json:"skip"`
}

var defaultConfig = Config{
	Default: DefaultConfig{MaxLines: 100},
	Rules:   []Rule{},
	Exclude: []string{
		"node_modules", "vendor", ".git", "dist", "build",
		"coverage", "__pycache__", "target", ".next", ".nuxt",
		"out", ".cache",
	},
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg, found, perr := loadConfigFromPackageJSON(filepath.Dir(path))
			if perr != nil {
				return defaultConfig, perr
			}
			if found {
				return cfg, nil
			}
			return defaultConfig, nil
		}
		return defaultConfig, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig, err
	}
	applyConfigDefaults(&cfg)
	return cfg, nil
}

func applyConfigDefaults(cfg *Config) {
	if cfg.Default.MaxLines == 0 {
		cfg.Default.MaxLines = defaultConfig.Default.MaxLines
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = defaultConfig.Exclude
	}
}

func defaultConfigJSON() string {
	return `{
  "default": {
    "maxLines": 100
  },
  "rules": [
    { "pattern": "**/*.spec.*",  "maxLines": 300 },
    { "pattern": "**/*.test.*",  "maxLines": 300 },
    { "pattern": "**/*_test.go", "maxLines": 300 },
    { "pattern": "**/migrations/**", "skip": true }
  ],
  "exclude": [
    "node_modules",
    "vendor",
    ".git",
    "dist",
    "build",
    "coverage",
    "__pycache__",
    "target",
    ".next",
    "out"
  ]
}
`
}
