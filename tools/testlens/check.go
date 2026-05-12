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
	exclude  []string
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "testlens.json", "path to config file")
	language := fs.String("lang", "auto", "language: go, typescript, python, ruby, rust, java, kotlin, csharp, auto")
	root := fs.String("dir", ".", "directory to scan")
	fail := fs.Bool("fail", false, "exit 1 if untested files found")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing flags: %v\n", err)
		return 1
	}

	fileCfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	cfg := mergeConfig(fileCfg, fs, *language, *root, *fail)
	violations, err := checkCoverage(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if violations > 0 {
		fmt.Printf("\n%d file(s) without tests\n", violations)
		if cfg.fail {
			return 1
		}
	} else {
		fmt.Println("All source files have tests")
	}
	return 0
}

func mergeConfig(file Config, fs *flag.FlagSet, language, root string, fail bool) config {
	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { explicit[f.Name] = true })

	final := config{language: language, root: root, fail: fail, exclude: file.Exclude}
	if !explicit["lang"] && file.Language != "" {
		final.language = file.Language
	}
	return final
}
