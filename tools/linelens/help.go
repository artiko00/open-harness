package main

import "fmt"

func printUsage() {
	fmt.Print(`linelens — file length linter for any language

USAGE:
  linelens <command> [options]

COMMANDS:
  check     scan files and report violations
  init      create a default linelens.json config
  version   show version

CHECK OPTIONS:
  --config    path to config file    (default: linelens.json)
  --max       override max lines     (default: from config or 100)
  --dir       directory to scan      (default: .)
  --fail      exit 1 if violations   (use in git hooks)
  --no-color  disable colored output

EXAMPLES:
  linelens check
  linelens check --fail
  linelens check --max 200 --dir ./src
  linelens check --config myproject.json --fail

HUSKY INTEGRATION (.husky/pre-commit):
  ./tools/linelens check --fail
`)
}
