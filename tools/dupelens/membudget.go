package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	mib         uint64 = 1 << 20
	gib         uint64 = 1 << 30
	autoCap            = gib // tope absoluto del presupuesto automático
	autoPercent uint64 = 25  // porcentaje de la memoria disponible en auto
	autoSource         = "auto: min(1 GiB, 25% of available memory)"
)

// budget es el presupuesto de memoria resuelto y de dónde salió, para el
// mensaje de exceso y para --verbose.
type budget struct {
	bytes  uint64
	source string
}

// memAvailable devuelve la memoria disponible al arrancar (ver memavail_*.go).
// Variable para inyectarla en tests.
var memAvailable = availableMemory

// parseMemValue interpreta "<n>MiB", "<n>GiB" o "<n>%" con n entero positivo
// (el porcentaje, de 1 a 100). Devuelve bytes absolutos o un porcentaje.
func parseMemValue(s string) (bytes, pct uint64, err error) {
	for _, u := range []struct {
		suffix string
		mult   uint64
	}{{"MiB", mib}, {"GiB", gib}, {"%", 0}} {
		num, ok := strings.CutSuffix(s, u.suffix)
		n, perr := strconv.ParseUint(num, 10, 32)
		if !ok || perr != nil || n == 0 || (u.mult == 0 && n > 100) {
			continue
		}
		if u.mult == 0 {
			return 0, n, nil
		}
		return n * u.mult, 0, nil
	}
	return 0, 0, fmt.Errorf("invalid memory value %q (valid: <n>MiB, <n>GiB or <n>%%, e.g. 512MiB, 2GiB, 40%%)", s)
}

// resolveBudget aplica la precedencia flag > DUPELENS_MAX_MEMORY > clave
// maxMemory > auto. Un valor explícito se aplica tal cual; un porcentaje se
// calcula sobre la memoria disponible y, si es indeterminable, cae a 1 GiB con
// un aviso en warn.
func resolveBudget(flagVal string, flagSet bool, cfgVal string, warn io.Writer) (budget, error) {
	val, src := cfgVal, "maxMemory (config)"
	if env := os.Getenv("DUPELENS_MAX_MEMORY"); flagSet || env != "" {
		val, src = env, "DUPELENS_MAX_MEMORY"
		if flagSet {
			val, src = flagVal, "--max-memory"
		}
	}
	if val == "" && !flagSet {
		return autoBudget(), nil
	}
	bytes, pct, err := parseMemValue(val)
	if err != nil {
		return budget{}, fmt.Errorf("%s: %w", src, err)
	}
	if pct == 0 {
		return budget{bytes, src}, nil
	}
	avail, ok := memAvailable()
	if !ok {
		fmt.Fprintf(warn, "warning: available memory unknown; %s %s falls back to 1 GiB\n", src, val)
		return budget{autoCap, src}, nil
	}
	return budget{avail * pct / 100, src}, nil
}

// autoBudget es el default: el menor entre 1 GiB y el 25 % de la memoria
// disponible, o 1 GiB si esta es indeterminable.
func autoBudget() budget {
	b := budget{autoCap, autoSource}
	if avail, ok := memAvailable(); ok && avail*autoPercent/100 < autoCap {
		b.bytes = avail * autoPercent / 100
	}
	return b
}

// formatBytes muestra n en GiB si es múltiplo exacto, o en MiB.
func formatBytes(n uint64) string {
	if n >= gib && n%gib == 0 {
		return fmt.Sprintf("%d GiB", n/gib)
	}
	return fmt.Sprintf("%d MiB", n/mib)
}
