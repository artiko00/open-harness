package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

type config struct {
	language string
	root     string
	fail     bool
	noColor  bool
	exclude  []string
	notest   []string
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "testlens.json", "path to config file")
	language := fs.String("lang", "auto", "language: go, typescript, python, ruby, rust, java, kotlin, csharp, auto")
	root := fs.String("dir", ".", "directory to scan")
	fail := fs.Bool("fail", false, "exit 1 if untested files found")
	noColor := fs.Bool("no-color", false, "disable colored output")
	format := fs.String("format", "console", "output format: console | json")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing flags: %v\n", err)
		return 1
	}

	if *format != "console" && *format != "json" {
		fmt.Fprintf(os.Stderr, "invalid format %q (valid: console, json)\n", *format)
		return 1
	}

	if *language != "auto" && !esLenguajeSoportado(*language) {
		fmt.Fprintf(os.Stderr, "unsupported language %q (supported: %s)\n",
			*language, strings.Join(supportedLanguages(), ", "))
		return 1
	}

	if _, err := os.Stat(*root); err != nil {
		fmt.Fprintf(os.Stderr, "directory %q not accessible: %v\n", *root, err)
		return 1
	}

	explicitConfig := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			explicitConfig = true
		}
	})
	if explicitConfig {
		if _, err := os.Stat(*configPath); err != nil {
			fmt.Fprintf(os.Stderr, "config file %q not found\n", *configPath)
			return 1
		}
	}

	fileCfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	cfg := mergeConfig(fileCfg, fs, *language, *root, *fail, *noColor)
	cov, err := checkCoverage(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if *format == "json" {
		reportJSON(cov, os.Stdout)
	} else {
		reportConsole(cov, cfg.noColor)
	}
	if cfg.fail && (len(cov.untested) > 0 || pathmatch.AnyFailsGate(cov.skips)) {
		return 1
	}
	return 0
}

func mergeConfig(file Config, fs *flag.FlagSet, language, root string, fail, noColor bool) config {
	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { explicit[f.Name] = true })

	final := config{language: language, root: root, fail: fail, noColor: noColor,
		exclude: file.Exclude, notest: file.NoTest}
	if !explicit["lang"] && file.Language != "" {
		final.language = file.Language
	}
	return final
}
