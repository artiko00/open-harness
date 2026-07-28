package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0"

var osExit = os.Exit

func main() {
	osExit(run(os.Args[1:]))
}

// run despacha subcomandos. Devuelve 0 (ok), 1 (excede presupuesto con --fail)
// o 2 (no se pudo medir, error de uso, git o config). Nunca entra en pánico.
func run(args []string) int {
	if len(args) < 1 {
		printUsage()
		return 2
	}

	switch args[0] {
	case "check":
		return runCheck(args[1:])
	case "init":
		return runInit(args[1:])
	case "--tutorial", "tutorial":
		return runTutorial(args[1:])
	case "version", "--version", "-v":
		fmt.Printf("scopelens %s\n", version)
		return 0
	case "help", "--help", "-h":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n\n", args[0])
		printUsage()
		return 2
	}
}

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	output := fs.String("output", "scopelens.json", "nombre del archivo de salida")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if _, err := os.Stat(*output); err == nil {
		fmt.Fprintf(os.Stderr, "%s ya existe, no se sobrescribe\n", *output)
		return 2
	}

	if err := os.WriteFile(*output, []byte(defaultConfigJSON()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error escribiendo config: %v\n", err)
		return 2
	}

	fmt.Printf("creado %s\n", *output)
	return 0
}
