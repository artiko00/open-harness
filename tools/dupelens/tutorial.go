package main

import (
	"fmt"
	"os"
)

// runTutorial imprime la guía de configuración estática y devuelve 0. Acepta
// --no-color entre los args para omitir códigos ANSI.
func runTutorial(args []string) int {
	noColor := false
	for _, a := range args {
		if a == "--no-color" {
			noColor = true
		}
	}
	printTutorial(noColor)
	return 0
}

// printTutorial escribe la guía a stdout. El único color es el título (bold),
// reutilizando el esquema del reporter; con noColor no emite ANSI.
func printTutorial(noColor bool) {
	fmt.Fprintf(os.Stdout, "%sdupelens — guía de configuración%s\n%s",
		boldOpt(noColor), resetOpt(noColor), tutorialBody)
}

const tutorialBody = `
Detecta código duplicado en cualquier lenguaje. Config: dupelens.json (o
scopelens.json). Ejecuta 'dupelens init' para generar una base.

CLAVES DE CONFIG (struct Config)
  default            objeto con los umbrales globales de reporte.
  default.minTokens  tokens mínimos para reportar un bloque   (default: 50).
  default.minLines   líneas mínimas del bloque duplicado      (default: 5).
  default.windowSize ventana de detección, independiente de minTokens; 0 usa
                     el default interno (25).                 (default: 0).
  rules              array de overrides por patrón glob.
  rules[].pattern    glob del archivo, ej. "**/*_test.go".
  rules[].minTokens  umbral propio de esa regla (opcional).
  rules[].skip       si es true, omite los archivos que casan.
  exclude            directorios excluidos del escaneo
                     (default: node_modules, vendor, .git, dist, build,
                     coverage, __pycache__, target, .next, out).

EJEMPLO dupelens.json
  {
    "default": { "minTokens": 50, "minLines": 5, "windowSize": 25 },
    "rules": [ { "pattern": "**/*_test.go", "skip": true } ],
    "exclude": [ "node_modules", "vendor", "dist" ]
  }

FLAGS (dupelens check)
  --config      ruta al archivo de config      (default: dupelens.json).
  --min-tokens  overridea el umbral de tokens.
  --dir         directorio a escanear           (default: .).
  --format      salida: console | json          (default: console).
  --fail        exit 1 si hay duplicados (hooks).
  --fail-on     qué rompe --fail: exact|renamed|all (default: exact).
  --no-color    desactiva el color (también en --tutorial).
  --verbose     imprime tiempos a stderr.

CAMBIOS 0.3.0
  La config se carga en modo estricto: una clave desconocida se avisa por
  stderr y se ignora (no aborta). Los archivos omitidos por reglas skip se
  reportan al final y pueden romper --fail; ya no se "falla en verde".
`
