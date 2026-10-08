package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// fakeMemUsed reemplaza la medición: devuelve los valores en orden y, agotados,
// repite el último.
func fakeMemUsed(t *testing.T, vals ...uint64) {
	t.Helper()
	orig, calls := memUsed, 0
	memUsed = func() uint64 {
		v := vals[min(calls, len(vals)-1)]
		calls++
		return v
	}
	t.Cleanup(func() { memUsed = orig })
}

func TestMemUsed_reportsRuntimeMemory(t *testing.T) {
	if memUsed() == 0 {
		t.Error("el runtime siempre tiene memoria tomada del SO")
	}
}

func TestReserve(t *testing.T) {
	if (memGuard{}).reserve(1<<62) != nil {
		t.Error("limit 0 = sin límite")
	}
	fakeMemUsed(t, 100)
	if (memGuard{limit: 200}).reserve(50) != nil {
		t.Error("dentro del presupuesto no debe fallar")
	}
	fakeMemUsed(t, 300, 100)
	if (memGuard{limit: 200}).check() != nil {
		t.Error("un exceso que desaparece tras liberar memoria no es real")
	}
	fakeMemUsed(t, 300)
	if !errors.Is((memGuard{limit: 200}).check(), errOverBudget) {
		t.Error("un exceso que sobrevive a la liberación debe devolver errOverBudget")
	}
}

func TestGrow_overBudgetLeavesSliceIntact(t *testing.T) {
	fakeMemUsed(t, 1<<40)
	s := []int{1, 2}
	got, err := grow(s, memGuard{limit: 1 << 30})
	if !errors.Is(err, errOverBudget) || len(got) != 2 || cap(got) != cap(s) {
		t.Errorf("grow con exceso = len %d cap %d, %v; want slice intacto y errOverBudget", len(got), cap(got), err)
	}
}

// Recorre cada punto de control del escaneo haciéndolo fallar: todos deben
// propagar errOverBudget sin entrar en pánico ni devolver hallazgos parciales.
func TestScan_everyCheckpointPropagatesOverBudget(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "a.go", blockX("x")+blockY())
	writeTempFile(t, dir, "b.go", blockX("x")+blockY())
	writeTempFile(t, dir, "c.go", strings.ReplaceAll(blockX("x"), "counts", "tally"))
	for n := 0; ; n++ {
		vals := make([]uint64, n, n+1)
		vals = append(vals, 1<<40)
		fakeMemUsed(t, vals...)
		matches, _, _, err := scan(dir, defaultConfig, 0, memGuard{limit: 1 << 30})
		if err == nil {
			if n == 0 || len(matches) == 0 {
				t.Fatalf("n=%d: el escaneo debía cortar antes o encontrar hallazgos", n)
			}
			return
		}
		if !errors.Is(err, errOverBudget) || matches != nil {
			t.Fatalf("n=%d: err = %v, matches = %d; want errOverBudget sin hallazgos", n, err, len(matches))
		}
	}
}

func TestScan_withinBudgetSameResultAsUnlimited(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "a.go", blockX("x"))
	writeTempFile(t, dir, "b.go", blockX("x"))
	free, _, _, _ := scan(dir, defaultConfig, 0, memGuard{})
	capped, _, _, _ := scan(dir, defaultConfig, 0, memGuard{limit: 1 << 40})
	if len(free) == 0 || !reflect.DeepEqual(free, capped) {
		t.Errorf("dentro del presupuesto el resultado debe ser idéntico:\n%+v\n%+v", free, capped)
	}
}
