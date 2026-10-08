package main

import (
	"errors"
	"runtime/debug"
	"runtime/metrics"
	"unsafe"
)

// errOverBudget indica que el escaneo no entra en el presupuesto de memoria.
var errOverBudget = errors.New("memory budget exceeded")

// memGuard hace cumplir el presupuesto de memoria en los puntos de control del
// escaneo: después de cada archivo, antes de cada reserva grande y al crecer los
// slices de candidatos, semillas y hallazgos. limit 0 = sin límite.
type memGuard struct{ limit uint64 }

// memUsed devuelve la memoria que el runtime tiene tomada del SO (total menos lo
// ya devuelto): la misma magnitud que acota debug.SetMemoryLimit, y una cota
// superior del heap más las pilas. Variable para inyectarla en tests.
var memUsed = func() uint64 {
	s := []metrics.Sample{{Name: "/memory/classes/total:bytes"}, {Name: "/memory/classes/heap/released:bytes"}}
	metrics.Read(s)
	return s[0].Value.Uint64() - s[1].Value.Uint64()
}

// reserve verifica que n bytes más entran en el presupuesto. Ante un exceso
// aparente fuerza un GC con devolución de memoria al SO y vuelve a medir: solo
// el exceso que sobrevive a eso es real.
func (g memGuard) reserve(n uint64) error {
	if g.limit == 0 || memUsed()+n <= g.limit {
		return nil
	}
	debug.FreeOSMemory()
	if memUsed()+n <= g.limit {
		return nil
	}
	return errOverBudget
}

// check es un punto de control sin reserva adicional.
func (g memGuard) check() error { return g.reserve(0) }

// grow duplica la capacidad de s (mínimo 1024 elementos) verificando antes que
// la nueva reserva entra en el presupuesto, para que el crecimiento de un slice
// grande no se salte los puntos de control. Ante un exceso devuelve s intacto.
func grow[T any](s []T, g memGuard) ([]T, error) {
	var zero T
	n := max(2*cap(s), 1024)
	if err := g.reserve(uint64(n) * uint64(unsafe.Sizeof(zero))); err != nil {
		return s, err
	}
	out := make([]T, len(s), n)
	copy(out, s)
	return out, nil
}
