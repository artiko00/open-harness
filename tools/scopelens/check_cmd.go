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
	maxLines := fs.Int("max-lines", 0, "presupuesto de líneas (0 = deshabilitado)")
	mode := fs.String("mode", "", "combinación de presupuestos: or | and")
	lineMetric := fs.String("line-metric", "", "métrica de líneas: changed | added")
	base := fs.String("base", "", "rama base a comparar")
	dir := fs.String("dir", ".", "directorio del repositorio")
	stagedOnly := fs.Bool("staged-only", false, "contar sólo el índice")
	excludeTests := fs.Bool("exclude-tests", false, "descontar tests del presupuesto")
	failFlag := fs.Bool("fail", false, "exit 1 si excede el presupuesto")
	noColor := fs.Bool("no-color", false, "salida sin ANSI")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if code := validateFlags(*maxFiles, *maxLines, *mode, *lineMetric, *dir); code != 0 {
		return code
	}

	cfg, err := loadConfig(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error de config: %v\n", err)
		return 2
	}

	limit, limitLines, combMode, metric := cfg.MaxFiles, cfg.MaxLines, cfg.Mode, cfg.LineMetric
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "max-files":
			limit = *maxFiles
		case "max-lines":
			limitLines = *maxLines
		case "mode":
			combMode = *mode
		case "line-metric":
			metric = *lineMetric
		}
	})
	exTests := cfg.ExcludeTests || *excludeTests

	res, err := measure(newRunner(*dir), *base, cfg.Base, *stagedOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scopelens: %v\n", err)
		return 2
	}

	rep := buildReport(res, cfg.Exclude, limit, limitLines, combMode, metric, exTests, *stagedOnly)
	printReport(rep, *noColor, os.Stdout)
	if *failFlag && rep.exceeded() {
		return 1
	}
	return 0
}

// validateFlags valida los flags de check y el directorio; devuelve 2 si alguno
// es inválido (no se puede medir), 0 si todo ok.
func validateFlags(maxFiles, maxLines int, mode, lineMetric, dir string) int {
	if maxFiles < 0 {
		fmt.Fprintf(os.Stderr, "max-files debe ser >= 0, no %d\n", maxFiles)
		return 2
	}
	if maxLines < 0 {
		fmt.Fprintf(os.Stderr, "max-lines debe ser >= 0, no %d\n", maxLines)
		return 2
	}
	if mode != "" && mode != "or" && mode != "and" {
		fmt.Fprintf(os.Stderr, "mode debe ser \"or\" o \"and\", no %q\n", mode)
		return 2
	}
	if lineMetric != "" && lineMetric != "changed" && lineMetric != "added" {
		fmt.Fprintf(os.Stderr, "line-metric debe ser \"changed\" o \"added\", no %q\n", lineMetric)
		return 2
	}
	if _, err := os.Stat(dir); err != nil {
		fmt.Fprintf(os.Stderr, "directorio %q inaccesible: %v\n", dir, err)
		return 2
	}
	return 0
}
