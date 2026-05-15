package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const composerToolKey = "secretlens"

// composerExtra mirrors the "extra" → "open-harness" → "<tool>" nesting
// inside composer.json, the same shape phpstan and friends use.
type composerExtra struct {
	Extra struct {
		OpenHarness map[string]json.RawMessage `json:"open-harness"`
	} `json:"extra"`
}

func loadConfigFromComposerJSON(dir string) (Config, bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, "composer.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, false, nil
		}
		return defaultConfig, false, err
	}

	var doc composerExtra
	if err := json.Unmarshal(data, &doc); err != nil {
		return defaultConfig, false, err
	}
	section, ok := doc.Extra.OpenHarness[composerToolKey]
	if !ok {
		return defaultConfig, false, nil
	}

	var cfg Config
	if err := json.Unmarshal(section, &cfg); err != nil {
		return defaultConfig, false, err
	}
	applyConfigDefaults(&cfg)
	return cfg, true, nil
}
