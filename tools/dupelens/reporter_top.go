package main

import "sort"

// FileMatchCount es par usado en summary "top duplicated files".
type FileMatchCount struct {
	File  string `json:"file"`
	Count int    `json:"count"`
}

// topDuplicatedFiles cuenta apariciones de cada archivo en el set de matches
// (sumando FileA y FileB) y retorna los N con más matches, ordenado descendente.
func topDuplicatedFiles(matches []Match, n int) []FileMatchCount {
	counts := make(map[string]int)
	for _, m := range matches {
		counts[m.FileA]++
		if m.FileB != m.FileA {
			counts[m.FileB]++
		}
	}
	out := make([]FileMatchCount, 0, len(counts))
	for f, c := range counts {
		out = append(out, FileMatchCount{File: f, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].File < out[j].File
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}
