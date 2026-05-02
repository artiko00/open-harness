package main

import "fmt"

const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

// report imprime los matches encontrados al stdout y retorna la cuenta total.
// Si noColor=true, descarta los códigos ANSI para output plano.
func report(matches []Match, noColor bool) int {
	if len(matches) == 0 {
		fmt.Printf("%s%sOK%s: no duplicate code blocks found\n",
			boldOpt(noColor), greenOpt(noColor), resetOpt(noColor))
		return 0
	}

	fmt.Printf("%s%sDUPLICATES%s (%d match(es) found):\n\n",
		boldOpt(noColor), redOpt(noColor), resetOpt(noColor), len(matches))

	for _, m := range matches {
		fmt.Printf("  %s%s:%d-%d%s  ↔  %s%s:%d-%d%s  %s(%d tokens)%s\n",
			boldOpt(noColor), m.FileA, m.StartLineA, m.EndLineA, resetOpt(noColor),
			boldOpt(noColor), m.FileB, m.StartLineB, m.EndLineB, resetOpt(noColor),
			grayOpt(noColor), m.Tokens, resetOpt(noColor))
	}
	fmt.Printf("\nSUMMARY: %d duplicate match(es)\n", len(matches))
	return len(matches)
}

func boldOpt(no bool) string  { if no { return "" }; return colorBold }
func redOpt(no bool) string   { if no { return "" }; return colorRed }
func greenOpt(no bool) string { if no { return "" }; return colorGreen }
func grayOpt(no bool) string  { if no { return "" }; return colorGray }
func resetOpt(no bool) string { if no { return "" }; return colorReset }
