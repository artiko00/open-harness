package main

import (
	"context"
	"strconv"
	"strings"
)

// lineStat es el churn de un archivo: líneas agregadas y borradas.
type lineStat struct{ Added, Deleted int }

// parseNumstat convierte la salida de `git diff --numstat` en churn por archivo.
// Cada línea es "added\tdeleted\tpath"; los binarios muestran "-" (cuentan 0) y
// los renames (con -M) traen el path en formato "{old => new}" o "old => new".
func parseNumstat(out []byte) map[string]lineStat {
	m := map[string]lineStat{}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		path := numstatPath(parts[2])
		st := m[path]
		st.Added += atoiOrZero(parts[0])
		st.Deleted += atoiOrZero(parts[1])
		m[path] = st
	}
	return m
}

// atoiOrZero devuelve el entero, o 0 si el campo es "-" (binario) o inválido.
func atoiOrZero(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// numstatPath extrae la ruta destino de un campo de numstat, resolviendo el
// quoting octal y el formato de rename (se queda con el "new").
func numstatPath(field string) string {
	if len(field) >= 2 && field[0] == '"' && field[len(field)-1] == '"' {
		field = unquotePath(field[1 : len(field)-1])
	}
	if i := strings.Index(field, "{"); i >= 0 {
		if j := strings.Index(field[i:], "}"); j >= 0 {
			j += i
			inner := field[i+1 : j]
			if k := strings.Index(inner, " => "); k >= 0 {
				field = field[:i] + inner[k+4:] + field[j+1:]
			}
		}
	} else if k := strings.Index(field, " => "); k >= 0 {
		field = field[k+4:]
	}
	return field
}

// accumulateChurn corre `git diff --numstat -M <spec>` y suma el churn de cada
// archivo en el mapa dado (un archivo tocado en varios specs acumula).
func accumulateChurn(ctx context.Context, run gitRunner, spec string, churn map[string]lineStat) error {
	out, err := run(ctx, "diff", "--numstat", "-M", spec)
	if err != nil {
		return operationalErr(err)
	}
	for p, st := range parseNumstat(out) {
		cur := churn[p]
		cur.Added += st.Added
		cur.Deleted += st.Deleted
		churn[p] = cur
	}
	return nil
}
