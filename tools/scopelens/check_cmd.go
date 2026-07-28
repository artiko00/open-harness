package main

import (
	"flag"
	"fmt"
	"os"
)

// newRunner es un seam: los tests lo reemplazan por una fábrica de gitRunner
// fake para ejercitar runCheck sin tocar el disco.
var newRunner = newGitRunner

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	maxFiles := fs.Int("max-files", 0, "presupuesto de archivos")
	base := fs.String("base", "", "rama base a comparar")
	dir := fs.String("dir", ".", "directorio del repositorio")
	stagedOnly := fs.Bool("staged-only", false, "contar sólo el índice")
	excludeTests := fs.Bool("exclude-tests", false, "descontar tests del presupuesto")
	failFlag := fs.Bool("fail", false, "exit 1 si excede el presupuesto")
	noColor := fs.Bool("no-color", false, "salida sin ANSI")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *maxFiles < 0 {
		fmt.Fprintf(os.Stderr, "max-files debe ser >= 0, no %d\n", *maxFiles)
		return 2
	}
	if _, err := os.Stat(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "directorio %q inaccesible: %v\n", *dir, err)
		return 2
	}

	cfg, err := loadConfig(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error de config: %v\n", err)
		return 2
	}

	limit := cfg.MaxFiles
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "max-files" {
			limit = *maxFiles
		}
	})
	exTests := cfg.ExcludeTests || *excludeTests

	res, err := measure(newRunner(*dir), *base, cfg.Base, *stagedOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scopelens: %v\n", err)
		return 2
	}

	rep := buildReport(res, cfg.Exclude, limit, exTests, *stagedOnly)
	printReport(rep, *noColor, os.Stdout)
	if *failFlag && rep.exceeded() {
		return 1
	}
	return 0
}
