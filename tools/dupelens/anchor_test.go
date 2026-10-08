package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// blockX genera una función sin repetición interna (ni literal ni
// estructural) y bastante larga para superar minTokens/minLines por defecto.
func blockX(tag string) string {
	return fmt.Sprintf(`func process_%s(items []string, limit int) (map[string]int, error) {
	counts := make(map[string]int, len(items))
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	for idx, item := range items {
		if idx >= limit {
			break
		}
		trimmed := strings.TrimSpace(item)
		counts[trimmed]++
	}
	total := 0
	for key, value := range counts {
		if value > 1 && len(key) > 3 {
			total += value
		}
	}
	switch {
	case total > 100:
		log.Printf("large total", total)
	case total == 0:
		return counts, io.EOF
	}
	return counts, nil
}
`, tag)
}

// blockY tiene una estructura distinta de blockX, para que su forma
// normalizada no colisione con la de X.
func blockY() string {
	var b strings.Builder
	for i := 0; i < 4; i++ {
		fmt.Fprintf(&b, "type Holder%d struct { items []string; size int }\n", i)
		fmt.Fprintf(&b, "func (h *Holder%d) Push(s string) { h.items = append(h.items, s); h.size++ }\n", i)
		fmt.Fprintf(&b, "func (h *Holder%d) Len() int { return len(h.items) + h.size*0 }\n", i)
	}
	return b.String()
}

func pairsOf(ms []Match) map[string]int {
	out := map[string]int{}
	for _, m := range ms {
		out[m.FileA+"|"+m.FileB]++
	}
	return out
}

func TestScan_threeIdenticalCopies_anchoredToFirst(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"a.go", "b.go", "c.go"} {
		writeTempFile(t, dir, n, blockX("x"))
	}
	matches, _, _, err := scan(dir, defaultConfig, 0, memGuard{})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	got := pairsOf(matches)
	if len(matches) != 2 || got["a.go|b.go"] != 1 || got["a.go|c.go"] != 1 {
		t.Fatalf("se esperaban exactamente a-b y a-c; got %v", got)
	}
	if exact, _ := countByKind(matches); exact != 2 {
		t.Errorf("exactCount = %d; want 2", exact)
	}
}

func TestScan_blockSharedOnlyByNonCanonicalCopies_isReported(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "a.go", blockX("x"))
	writeTempFile(t, dir, "b.go", blockX("x")+blockY())
	writeTempFile(t, dir, "c.go", blockX("x")+blockY())
	matches, _, _, err := scan(dir, defaultConfig, 0, memGuard{})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if pairsOf(matches)["b.go|c.go"] != 1 {
		t.Fatalf("el bloque Y (solo en b y c) debe reportarse como b-c; got %v", pairsOf(matches))
	}
}

func TestRun_threeExactCopies_failStillExits1(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"a.go", "b.go", "c.go"} {
		writeTempFile(t, dir, n, blockX("x"))
	}
	if code := run([]string{"check", "--dir", dir, "--fail", "--no-color", "--config", writeCfg(t, dir)}); code != 1 {
		t.Errorf("tres copias exact con --fail: exit = %d; want 1", code)
	}
}

// writeCfg deja una config explícita con los defaults para no depender del cwd.
func writeCfg(t *testing.T, dir string) string {
	t.Helper()
	p := filepath.Join(dir, "dupelens.json")
	if err := os.WriteFile(p, []byte(defaultConfigJSON()), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// allocFor mide los bytes asignados por scan sobre k copias de un paquete.
func allocFor(t *testing.T, k int) uint64 {
	t.Helper()
	dir := t.TempDir()
	pkg := blockX("p") + blockY()
	for i := 0; i < 40; i++ {
		pkg += blockX(fmt.Sprintf("q%d", i))
	}
	for c := 0; c < k; c++ {
		writeTempFile(t, dir, fmt.Sprintf("copy_%02d.go", c), pkg)
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	if _, _, _, err := scan(dir, defaultConfig, 0, memGuard{}); err != nil {
		t.Fatalf("scan: %v", err)
	}
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// Duplicar las copias de un mismo paquete debe duplicar el costo, no
// cuadruplicarlo: es el eje que hacía caer a proyectos chicos con código
// generado o vendorizado.
func TestMemory_copiesScaleLinearly(t *testing.T) {
	if testing.Short() {
		t.Skip("omitido en -short: escanea 48 copias de un paquete")
	}
	a16, a32 := allocFor(t, 16), allocFor(t, 32)
	if float64(a32) >= 2.5*float64(a16) {
		t.Errorf("32 copias asignaron %d KB, 16 copias %d KB: crece más rápido que lineal",
			a32>>10, a16>>10)
	}
	t.Logf("16 copias: %d KB, 32 copias: %d KB", a16>>10, a32>>10)
}

// Si un archivo anterior en el orden canónico comparte solo una parte del
// bloque, el ancla cambia a mitad de camino. El par posterior igual debe
// reportarse completo: la semilla se extiende sobre su diagonal.
func TestFindDuplicates_anchorFlipMidBlockDoesNotFragment(t *testing.T) {
	block := []string{"t01", "t02", "t03", "t04", "t05", "t06", "t07", "t08", "t09", "t10"}
	got := find(t, 3, 1, 10,
		tf{"a.go", block[:5]},
		tf{"m.go", block},
		tf{"n.go", block})
	if pairsOf(got)["m.go|n.go"] != 1 {
		t.Fatalf("m-n comparten 10 tokens y deben reportarse completos; got %+v", got)
	}
}

// El conteo de tokens de un hallazgo es el tramo real que abarca: 0.4.x sumaba
// un token por cada par de ventanas fusionado, aunque vinieran de alineaciones
// distintas, y en código periódico declaraba más tokens que los del bloque.
func TestFindDuplicates_tokensNeverExceedSpan(t *testing.T) {
	seq := strings.Fields("p q r s p q r s p q r s p q r s")
	for _, m := range find(t, 3, 1, 0, tf{"a.go", seq}) {
		if span := m.EndLineA - m.StartLineA + 1; m.Tokens > span {
			t.Errorf("match %d-%d declara %d tokens en un tramo de %d", m.StartLineA, m.EndLineA, m.Tokens, span)
		}
	}
}
