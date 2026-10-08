package main

import "sort"

// Match representa un par de bloques duplicados detectados. Kind es "exact"
// (tokens idénticos) o "renamed" (misma estructura, identificadores distintos).
// startIdxA/endIdxA son internos: el rango de tokens [start, end) del bloque en
// el archivo A, para contar los tokens al fusionar corridas.
type Match struct {
	FileA      string
	StartLineA int
	EndLineA   int
	FileB      string
	StartLineB int
	EndLineB   int
	Tokens     int
	Kind       string

	startIdxA int
	endIdxA   int
}

// fileData guarda, por archivo, el id internado y la línea de cada token. Los
// tokens normalizados no se copian: se derivan del id vía vocab.norm.
type fileData struct {
	name  string
	ids   []uint32
	lines []uint32
}

// findDuplicates corre dos pasadas: exact (vista cruda) y renamed (vista
// normalizada). Los renamed que ya cubre un exact se descartan para no
// reportarlos dos veces. Ordena los archivos por ruta para que el anclaje a la
// primera ocurrencia sea canónico, y devuelve el resultado ordenado.
func findDuplicates(files []fileData, v *vocab, windowSize, minLines, minTokens int, g memGuard) ([]Match, error) {
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	exact, err := detect(files, v, false, windowSize, minLines, minTokens, g)
	if err != nil {
		return nil, err
	}
	renamed, err := detect(files, v, true, windowSize, minLines, minTokens, g)
	if err != nil {
		return nil, err
	}
	out := append(exact, dropCoveredByExact(renamed, exact)...)
	sortMatches(out)
	return out, nil
}

// detect encadena prefiltro, semillas ancladas y fusión de corridas sobre una
// vista, filtra por líneas y tokens mínimos y etiqueta el tipo de hallazgo.
func detect(files []fileData, v *vocab, norm bool, w, minLines, minTokens int, g memGuard) ([]Match, error) {
	recs, err := candidates(files, v, w, norm, g)
	if err != nil {
		return nil, err
	}
	seeds, err := anchorSeeds(files, v, recs, w, norm, g)
	if err != nil {
		return nil, err
	}
	out, err := mergeRuns(files, v, seeds, w, norm, g)
	if err != nil {
		return nil, err
	}
	out = filterByMinTokens(filterByMinLines(out, minLines), minTokens)
	kind := "exact"
	if norm {
		kind = "renamed"
	}
	for i := range out {
		out[i].Kind = kind
	}
	return out, nil
}
