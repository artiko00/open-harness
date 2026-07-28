package main

import (
	"fmt"
	"os"
)

// tutorialGuide es la guía estática de configuración. Los verbos %[1]s/%[2]s se
// rellenan con los códigos de negrita/reset (vacíos cuando se pide sin color).
const tutorialGuide = `%[1]sGUÍA DE CONFIGURACIÓN DE testlens (v0.3.0)%[2]s

testlens busca archivos de código fuente que no tienen un test asociado.
Se configura con un archivo JSON (testlens.json o scopelens.json).

%[1]sCLAVES DE CONFIG%[2]s

  %[1]slanguage%[2]s  Lenguaje a analizar o "auto" para detección automática.
            Por defecto: "auto".
            Soportados: go, typescript, javascript, python, ruby, rust,
            java, kotlin, csharp.

  %[1]sexclude%[2]s   Directorios/patrones a ignorar durante el escaneo.
            Por defecto: node_modules, .git, vendor, dist, build, coverage,
            __pycache__, target, .next, .nuxt, out, .cache, testdata.

  %[1]snotest%[2]s    Patrones (glob sobre el nombre de archivo) de fuentes que NO
            requieren test propio; evita falsos positivos.
            Por defecto: __init__.py, conftest.py, settings.py, *_pb2.py,
            *.pb.go, *_gen.go, *.g.dart, main.go, doc.go.

Ejemplo de testlens.json (o scopelens.json):

  {
    "language": "auto",
    "exclude": ["node_modules", ".git", "dist", "testdata"],
    "notest": ["main.go", "*_gen.go"]
  }

%[1]sFLAGS%[2]s

  --config <ruta>   Ruta al archivo de config (default: testlens.json).
  --lang <lenguaje> Fuerza el lenguaje; anula "language" del config.
  --dir <ruta>      Directorio a escanear (default: .).
  --fail            Devuelve exit 1 si hay archivos sin test (para CI).
  --no-color        Desactiva los colores ANSI en la salida.
  --format <fmt>    Formato de salida: console | json (default: console).

%[1]sCAMBIOS 0.3.0%[2]s

  - testlens verifica el CONTENIDO de los tests: un archivo de test vacío o
    sin funciones de test reales ya no cuenta como cobertura.
  - Se reportan los archivos omitidos (binarios, etc.) en el resumen y en el
    JSON bajo la clave "skipped".
  - Carga de config estricta: una clave desconocida emite un aviso por stderr.
`

// printTutorial imprime la guía estática a stdout. Sin ANSI si noColor.
func printTutorial(noColor bool) {
	bold, reset := colorBold, colorReset
	if noColor {
		bold, reset = "", ""
	}
	fmt.Fprintf(os.Stdout, tutorialGuide, bold, reset)
}

// tieneFlag indica si name aparece entre los args.
func tieneFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}
