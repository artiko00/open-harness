package main

import (
	"encoding/binary"
	"unicode/utf16"
)

// decodeContent devuelve el contenido en UTF-8. Si detecta un BOM UTF-16
// (FF FE little-endian o FE FF big-endian) lo decodifica; en otro caso
// devuelve los bytes tal cual (UTF-8 o texto de un byte).
func decodeContent(data []byte) []byte {
	if len(data) >= 2 {
		if data[0] == 0xFF && data[1] == 0xFE {
			return utf16ToUTF8(data[2:], binary.LittleEndian)
		}
		if data[0] == 0xFE && data[1] == 0xFF {
			return utf16ToUTF8(data[2:], binary.BigEndian)
		}
	}
	return data
}

func utf16ToUTF8(b []byte, order binary.ByteOrder) []byte {
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u = append(u, order.Uint16(b[i:i+2]))
	}
	return []byte(string(utf16.Decode(u)))
}
