package main

import (
	"fmt"
	"os"
	"strings"
)

// runTutorial imprime la guía estática de configuración y devuelve 0. Parsea
// los args de forma simple: sólo reconoce --no-color para suprimir el ANSI.
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

// printTutorial vuelca la guía a stdout. Los marcadores « » delimitan los
// títulos en negrita; con noColor se eliminan sin dejar secuencias ANSI.
func printTutorial(noColor bool) {
	text := tutorialText
	if noColor {
		text = strings.NewReplacer("«", "", "»", "").Replace(text)
	} else {
		text = strings.NewReplacer("«", colorBold, "»", colorReset).Replace(text)
	}
	fmt.Fprint(os.Stdout, text)
}

const tutorialText = `«scopelens — guía de configuración»

scopelens mide el diff acumulado (rama vs base) y lo compara contra un
presupuesto de archivos. Config en scopelens.json (o pyproject/package.json/
composer.json vía cadena). Genera un ejemplo con: scopelens init

«CLAVES DE CONFIG (scopelens.json)»

  «maxFiles» (int, default 15)
    Presupuesto de archivos contables del diff. 0 o ausente usa el default;
    negativo es error (exit 2).  Ej.:  "maxFiles": 20

  «base» (string, default "")
    Rama base por defecto para el diff. Vacío autodetecta origin/HEAD > main
    > master. --base la sobreescribe.  Ej.:  "base": "develop"

  «excludeTests» (bool, default false)
    Descuenta del presupuesto los archivos de test. Equivale al flag
    --exclude-tests (se combinan por OR).  Ej.:  "excludeTests": true

  «exclude» ([]string, default: lista integrada)
    Globs excluidos del conteo (lockfiles, generados, node_modules, etc.).
    OJO: el array es ATÓMICO — si lo defines, REEMPLAZA por completo la lista
    integrada (no se suma).  Ej.:  "exclude": ["dist/**", "*.gen.go"]

«FLAGS DEL TOOL»

  scopelens check [opciones]
    --max-files <int>   presupuesto; sobreescribe maxFiles de la config
    --base <ref>        rama base; sobreescribe base de la config
    --dir <path>        directorio del repo (default: .)
    --staged-only       cuenta sólo el índice (staged)
    --exclude-tests     descuenta los tests del presupuesto
    --fail              exit 1 si excede el presupuesto (para git hooks)
    --no-color          salida sin secuencias ANSI
  scopelens init [--output <archivo>]   crea un scopelens.json con defaults
  scopelens --tutorial [--no-color]     esta guía
  scopelens version | help

«EXIT CODES»
  0 dentro del presupuesto (o fuera, sin --fail)
  1 excede el presupuesto (con --fail)
  2 no se pudo medir: git ausente, no es repo, shallow, base no resoluble o
    config inválida (maxFiles negativo, JSON malformado)

«COMPORTAMIENTO 0.3.0»
  La config se carga en modo estricto: una clave desconocida emite un warning
  a stderr (no falla). En la cadena pyproject/package.json/composer.json cada
  campo lo aporta el primer archivo que lo defina; el array exclude es atómico.
`
