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
  --fail        exit 1 if duplicates      (use in git hooks)
  --no-color    disable colored output

EXAMPLES:
  dupelens check
  dupelens check --fail
  dupelens check --min-tokens 30 --dir ./src

LEFTHOOK INTEGRATION (lefthook.yml):
  pre-commit:
    commands:
      dupelens:
        run: ./tools/dupelens/dupelens check --fail
`)
}
