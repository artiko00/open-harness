package main

// dropCoveredByExact descarta los matches renamed que ya cubre un match exact
// del mismo par de archivos (todo clon exact es también estructuralmente igual;
// reportarlo dos veces sería ruido). Se conservan los renamed genuinos.
func dropCoveredByExact(renamed, exact []Match) []Match {
	var out []Match
	for _, r := range renamed {
		if !coveredByExact(r, exact) {
			out = append(out, r)
		}
	}
	return out
}

// coveredByExact indica si el match r se solapa con algún match exact en ambos
// archivos del par.
func coveredByExact(r Match, exact []Match) bool {
	for _, e := range exact {
		if e.FileA == r.FileA && e.FileB == r.FileB &&
			overlaps(e.StartLineA, e.EndLineA, r.StartLineA, r.EndLineA) &&
			overlaps(e.StartLineB, e.EndLineB, r.StartLineB, r.EndLineB) {
			return true
		}
	}
	return false
}

// overlaps indica si dos rangos [s1,e1] y [s2,e2] se intersecan.
func overlaps(s1, e1, s2, e2 int) bool {
	return s1 <= e2 && s2 <= e1
}

// gateCount cuenta los matches que rompen --fail según --fail-on:
// "exact" (default) solo cuenta exact, "renamed" solo renamed, "all" todos.
func gateCount(matches []Match, failOn string) int {
	n := 0
	for _, m := range matches {
		if failOn == "all" || m.Kind == failOn {
			n++
		}
	}
	return n
}
