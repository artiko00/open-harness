package main

import (
	"github.com/artiko00/open-harness/tools/_shared/configload"
)

// loadConfigFallbackChain acumula los archivos presentes en orden de prioridad
// (pyproject.toml, package.json, composer.json) y hace merge por campo: para
// cada campo gana el primero de la cadena que lo defina. Al agotar la cadena,
// applyConfigDefaults rellena lo que quede con los defaults compilados.
func loadConfigFallbackChain(dir string) (Config, error) {
	loaders := []func(string) (Config, bool, error){
		func(d string) (Config, bool, error) { return configload.Pyproject[Config](d, "tool.linelens") },
		func(d string) (Config, bool, error) { return configload.PackageJSON[Config](d, "linelens") },
		func(d string) (Config, bool, error) { return configload.Composer[Config](d, "linelens") },
	}
	var merged Config
	for _, load := range loaders {
		cfg, found, err := load(dir)
		if err != nil {
			return defaultConfig, err
		}
		if found {
			mergeConfig(&merged, cfg)
		}
	}
	applyConfigDefaults(&merged)
	return merged, nil
}
