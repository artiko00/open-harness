package main

import "fmt"

func printUsage() {
	fmt.Print(`secretlens — secret and credential detector for any codebase

USAGE:
  secretlens <command> [options]

COMMANDS:
  check     scan files and report potential secrets
  init      create a default secretlens.json config
  version   show version

CHECK OPTIONS:
  --config    path to config file    (default: secretlens.json)
  --dir       directory to scan      (default: .)
  --format    output format          (console | json, default: console)
  --fail      exit 1 if secrets found (use in git hooks)
  --no-color  disable colored output

EXAMPLES:
  secretlens check
  secretlens check --fail
  secretlens check --dir ./src --no-color
  secretlens check --config myproject.json --fail

HUSKY / LEFTHOOK INTEGRATION:
  secretlens check --fail

SEVERITY LEVELS:
  critical  known secret formats (AWS keys, GitHub tokens, PEM keys)
  high      generic secret assignments with long values
  medium    token assignments, bearer headers
`)
}
