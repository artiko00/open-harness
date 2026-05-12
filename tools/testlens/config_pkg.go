package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const packageJSONKey = "testlens"

func loadConfigFromPackageJSON(dir string) (Config, bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, false, nil
		}
		return defaultConfig, false, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return defaultConfig, false, err
	}

	section, ok := raw[packageJSONKey]
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
