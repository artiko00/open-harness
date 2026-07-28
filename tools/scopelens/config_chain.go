package main

import "github.com/artiko00/open-harness/tools/_shared/configload"

// loadConfigChain acumula los archivos de la cadena en orden de prioridad
// (pyproject.toml, package.json, composer.json) y hace merge por campo: para
// cada campo gana el primero de la cadena que lo defina. Al agotar la cadena,
// validateAndDefault rellena lo que quede con los defaults.
func loadConfigChain(dir string) (Config, error) {
	loaders := []func(string) (Config, bool, error){
		func(d string) (Config, bool, error) { return configload.Pyproject[Config](d, "tool.scopelens") },
		func(d string) (Config, bool, error) { return configload.PackageJSON[Config](d, "scopelens") },
		func(d string) (Config, bool, error) { return configload.Composer[Config](d, "scopelens") },
	}
	var merged Config
	for _, load := range loaders {
		cfg, found, err := load(dir)
		if err != nil {
			return Config{}, err
		}
		if found {
			mergeConfig(&merged, cfg)
		}
	}
	return validateAndDefault(merged, "defaults")
}
