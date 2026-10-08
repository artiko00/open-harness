package main

// vocab interna los tokens de una ejecución: cada string distinto recibe un id
// estable junto con su hash de token y el id de su forma normalizada. Crece con
// el vocabulario del proyecto —sublineal en el tamaño del fuente— y evita
// retener cada token como string, crudo y normalizado, por posición.
type vocab struct {
	ids  map[string]uint32
	hash []uint64 // hash de token, por id
	norm []uint32 // id de la forma normalizada, por id
}

func newVocab() *vocab { return &vocab{ids: map[string]uint32{}} }

// id devuelve el id de s y, la primera vez, registra su hash y su forma
// normalizada, que también se interna. La recursión es corta: la forma
// normalizada de "NUM" es "ID", y la de "ID" es ella misma.
func (v *vocab) id(s string) uint32 {
	if id, ok := v.ids[s]; ok {
		return id
	}
	id := uint32(len(v.hash))
	v.ids[s] = id
	v.hash = append(v.hash, hashToken(s))
	v.norm = append(v.norm, id)
	if n := normalizeValue(s); n != s {
		nid := v.id(n)
		v.norm[id] = nid
	}
	return id
}

// at devuelve el id del token i de f en la vista pedida: crudo o normalizado.
func (v *vocab) at(f fileData, i int, norm bool) uint32 {
	if norm {
		return v.norm[f.ids[i]]
	}
	return f.ids[i]
}
