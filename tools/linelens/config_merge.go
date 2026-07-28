package main

// mergeConfig completa los campos aún sin definir de dst con los de src.
// Para cada campo gana el primer archivo de la cadena que lo defina
// (no-cero/no-vacío); los arrays (Rules, Exclude) son atómicos: se toma entero
// el del primer archivo que los aporte.
func mergeConfig(dst *Config, src Config) {
	if dst.Default.MaxLines == 0 {
		dst.Default.MaxLines = src.Default.MaxLines
	}
	if dst.Default.MaxNesting == 0 {
		dst.Default.MaxNesting = src.Default.MaxNesting
	}
	if len(dst.Rules) == 0 {
		dst.Rules = src.Rules
	}
	if len(dst.Exclude) == 0 {
		dst.Exclude = src.Exclude
	}
}
