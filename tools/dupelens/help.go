package main

import "fmt"

func printUsage() {
	fmt.Print(`dupelens — code duplication detector for any language

USAGE:
  dupelens <command> [options]

COMMANDS:
  check     scan files and report duplicate code blocks
  init      create a default dupelens.json config
  version   show version

CHECK OPTIONS:
  --config      path to config file       (default: dupelens.json)
  --min-tokens  override token threshold  (default: from config or 50)
  --dir         directory to scan         (default: .)
  --format      output format             (default: console; "json" available)
  --fail        exit 1 if duplicates      (use in git hooks)
  --fail-on     kinds that break --fail   (default: exact; "renamed"|"all")
  --no-color    disable colored output
  --verbose     print scan timings to stderr

EXAMPLES:
  dupelens check
  dupelens check --fail
  dupelens check --format=json > report.json
  dupelens check --min-tokens 30 --dir ./src --verbose

LEFTHOOK INTEGRATION (lefthook.yml):
  pre-commit:
    commands:
      dupelens:
        run: ./tools/dupelens/dupelens check --fail
`)
}
