package configload

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Strict deserializa data en v y, además, detecta la primera clave desconocida
// del JSON (una que no corresponde a ningún campo de v). Puebla v con los
// valores conocidos de todos modos, de manera que el llamador pueda avisar de
// la clave desconocida sin descartar la configuración (retrocompatibilidad:
// las versiones previas ignoraban las claves desconocidas en silencio).
//
// Devuelve el nombre de la clave desconocida en unknown ("" si no hay). Un
// error de JSON malformado o de tipo sí se devuelve como err.
func Strict(data []byte, v any) (unknown string, err error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if e := dec.Decode(v); e != nil {
		const marker = "unknown field "
		msg := e.Error()
		if i := strings.Index(msg, marker); i >= 0 {
			unknown = strings.Trim(msg[i+len(marker):], `"`)
			if e2 := json.Unmarshal(data, v); e2 != nil {
				return "", e2
			}
			return unknown, nil
		}
		return "", e
	}
	return "", nil
}
