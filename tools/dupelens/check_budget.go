package main

import (
	"flag"
	"fmt"
	"io"
	"runtime/debug"
)

// applyBudget fija el límite blando del GC al 90 % del presupuesto, para que el
// runtime recolecte antes de llegar al tope que hacen cumplir los puntos de
// control. Devuelve la función que restaura el límite anterior.
func applyBudget(b budget) (restore func()) {
	prev := debug.SetMemoryLimit(int64(b.bytes / 10 * 9))
	return func() { debug.SetMemoryLimit(prev) }
}

// printOverBudget explica que la medición no se completó, con el presupuesto
// aplicado, su origen y cómo subirlo o reducir el alcance del escaneo.
func printOverBudget(w io.Writer, b budget) {
	fmt.Fprintf(w, "dupelens: memory budget exceeded (%s, from %s): the scan could not be completed.\n",
		formatBytes(b.bytes), b.source)
	fmt.Fprintln(w, "Raise it with --max-memory, DUPELENS_MAX_MEMORY or the maxMemory config key"+
		" (e.g. 2GiB or 50%), or narrow the scan with exclude.")
}

// flagWasSet indica si el flag name se pasó explícitamente.
func flagWasSet(fs *flag.FlagSet, name string) bool {
	set := false
	fs.Visit(func(f *flag.Flag) { set = set || f.Name == name })
	return set
}
