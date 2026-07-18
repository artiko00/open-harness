package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	language string
	root     string
	fail     bool
	noColor  bool
	exclude  []string
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "testlens.json", "path to config file")
	language := fs.String("lang", "auto", "language: go, typescript, python, ruby, rust, java, kotlin, csharp, auto")
	root := fs.String("dir", ".", "directory to scan")
	fail := fs.Bool("fail", false, "exit 1 if untested files found")
	noColor := fs.Bool("no-color", false, "disable colored output")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing flags: %v\n", err)
		return 1
	}

	fileCfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	cfg := mergeConfig(fileCfg, fs, *language, *root, *fail, *noColor)
	violations, err := checkCoverage(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	fmt.Println(summaryLine(violations, cfg.noColor))
	if violations > 0 && cfg.fail {
		return 1
	}
	return 0
}

func mergeConfig(file Config, fs *flag.FlagSet, language, root string, fail, noColor bool) config {
	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { explicit[f.Name] = true })

	final := config{language: language, root: root, fail: fail, noColor: noColor, exclude: file.Exclude}
	if !explicit["lang"] && file.Language != "" {
		final.language = file.Language
	}
	return final
}
