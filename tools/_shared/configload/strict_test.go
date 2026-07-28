package configload

import "testing"

type strictCfg struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func TestStrict_sinClaveDesconocida(t *testing.T) {
	var c strictCfg
	unknown, err := Strict([]byte(`{"name":"a","enabled":true}`), &c)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if unknown != "" {
		t.Fatalf("no debia haber clave desconocida, got %q", unknown)
	}
	if c.Name != "a" || !c.Enabled {
		t.Fatalf("valores mal poblados: %+v", c)
	}
}

func TestStrict_claveDesconocidaSePoblaIgual(t *testing.T) {
	var c strictCfg
	unknown, err := Strict([]byte(`{"name":"a","extra":1,"enabled":true}`), &c)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if unknown != "extra" {
		t.Fatalf("clave desconocida esperada 'extra', got %q", unknown)
	}
	// Los valores conocidos se poblaron pese a la clave desconocida.
	if c.Name != "a" || !c.Enabled {
		t.Fatalf("valores conocidos deben poblarse igual: %+v", c)
	}
}

func TestStrict_jsonMalformadoEsError(t *testing.T) {
	var c strictCfg
	if _, err := Strict([]byte(`{"name":`), &c); err == nil {
		t.Fatal("JSON malformado debe devolver error")
	}
}

func TestStrict_tipoInvalidoEsError(t *testing.T) {
	var c strictCfg
	if _, err := Strict([]byte(`{"name":123}`), &c); err == nil {
		t.Fatal("tipo invalido debe devolver error")
	}
}

func TestStrict_reDecodeMalformadoTrasClaveDesconocida(t *testing.T) {
	// La clave desconocida aparece antes de un error de tipo: el re-decode
	// lenient falla y se propaga el error.
	var c strictCfg
	if _, err := Strict([]byte(`{"extra":1,"name":123}`), &c); err == nil {
		t.Fatal("error de tipo tras clave desconocida debe propagarse")
	}
}
