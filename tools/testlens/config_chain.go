package main

import "github.com/artiko00/open-harness/tools/_shared/configload"

// loadConfigFallbackChain acumula los archivos presentes en orden de prioridad
// (pyproject.toml, package.json, composer.json) y hace merge por campo: gana el
// primer archivo que defina cada campo. Al agotar la cadena, los defaults
// compilados rellenan lo que quede.
func loadConfigFallbackChain(dir string) (Config, error) {
	loaders := []func(string) (Config, bool, error){
		func(d string) (Config, bool, error) { return configload.Pyproject[Config](d, "tool.testlens") },
		func(d string) (Config, bool, error) { return configload.PackageJSON[Config](d, "testlens") },
		func(d string) (Config, bool, error) { return configload.Composer[Config](d, "testlens") },
	}

	var chain []Config
	for _, load := range loaders {
		cfg, found, err := load(dir)
		if err != nil {
			return defaultConfig, err
		}
		if found {
			chain = append(chain, cfg)
		}
	}

	merged := mergeConfigChain(chain)
	applyConfigDefaults(&merged)
	return merged, nil
}
