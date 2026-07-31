package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_ignoreImportsDefaultsToTrue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	writeTempFile(t, dir, "cfg.json", `{"default":{"minTokens":10}}`)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("config válida no debe fallar: %v", err)
	}
	if !cfg.ignoreImports() {
		t.Error("ignoreImports ausente debe resolver a true")
	}
}

func TestLoadConfig_ignoreImportsExplicitFalseIsRespected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	writeTempFile(t, dir, "cfg.json", `{"default":{"ignoreImports":false}}`)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("config válida no debe fallar: %v", err)
	}
	if cfg.ignoreImports() {
		t.Error(`"ignoreImports": false debe respetarse`)
	}
}

func TestLoadConfig_ignoreImportsExplicitTrueIsRespected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	writeTempFile(t, dir, "cfg.json", `{"default":{"ignoreImports":true}}`)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("config válida no debe fallar: %v", err)
	}
	if !cfg.ignoreImports() {
		t.Error(`"ignoreImports": true debe respetarse`)
	}
}

func TestTokenize_dropsImportDeclarations(t *testing.T) {
	src := "import { UserService } from './user.service';\nfunction run() { return compute(); }"
	tokens := tokenize(src, ".ts", true)
	for _, tok := range tokens {
		if tok.Value == "import" || tok.Value == "from" {
			t.Fatalf("los imports debían descartarse, got %v", tokens)
		}
	}
	if !containsToken(tokens, "compute") {
		t.Fatalf("el cuerpo debía conservarse, got %v", tokens)
	}
}

func TestTokenize_keepsImportsWhenDisabled(t *testing.T) {
	src := "import { UserService } from './user.service';\nfunction run() {}"
	if !containsToken(tokenize(src, ".ts", false), "import") {
		t.Error("con el descarte apagado los tokens de import deben conservarse")
	}
}

func TestTokenize_importStripPreservesLineNumbers(t *testing.T) {
	src := "import a from 'a';\nimport b from 'b';\nfunction later_marker() {}"
	tokens := tokenize(src, ".ts", true)
	for _, tok := range tokens {
		if tok.Value == "later_marker" && tok.Line != 3 {
			t.Fatalf("la línea del token = %d; want 3", tok.Line)
		}
	}
}

func containsToken(tokens []Token, v string) bool {
	for _, tok := range tokens {
		if tok.Value == v {
			return true
		}
	}
	return false
}

// nestHeader genera una cabecera de imports al estilo NestJS: idéntica en
// estructura entre archivos, con módulos distintos. Alterna las tres formas de
// import de JS para que el primer token varíe por línea: así el bloque no es de
// baja entropía y solo StripImports puede neutralizarlo.
func nestHeader(prefix string) string {
	forms := []string{
		"import { %sSym%s } from './%s/mod%s';\n",
		"export { %sSym%s } from './%s/mod%s';\n",
		"const %sSym%s = require('./%s/mod%s');\n",
	}
	var b strings.Builder
	for i := 0; i < 12; i++ {
		up, low := string(rune('A'+i)), string(rune('a'+i))
		b.WriteString(fmt.Sprintf(forms[i%len(forms)], prefix, up, prefix, low))
	}
	return b.String()
}

func TestScan_importHeadersAloneAreNotDuplicates(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "a.ts", nestHeader("alpha")+
		"\nexport function alphaTotal(rows) {\n  let acc = 0;\n  for (const r of rows) acc += r.price;\n  return acc;\n}\n")
	writeTempFile(t, dir, "b.ts", nestHeader("beta")+
		"\nexport class BetaGate {\n  check(user) {\n    if (!user.active) throw new Error();\n    return user.id;\n  }\n}\n")

	cfg := defaultConfig
	cfg.Default.MinTokens = 20
	matches, _, _, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("las cabeceras de imports no son duplicación, got %d matches: %+v", len(matches), matches)
	}
}

func TestScan_realDuplicateStillDetectedWithImportsStripped(t *testing.T) {
	body := "\nexport function totalize(rows, rate) {\n" +
		"  let acc = 0;\n  for (const row of rows) {\n    acc += row.price * rate;\n" +
		"    if (row.discount) acc -= row.discount;\n  }\n" +
		"  const tax = acc * 0.21;\n  const shipping = acc > 100 ? 0 : 12;\n" +
		"  return { acc, tax, shipping };\n}\n"
	dir := t.TempDir()
	writeTempFile(t, dir, "a.ts", nestHeader("alpha")+body)
	writeTempFile(t, dir, "b.ts", nestHeader("beta")+body)

	cfg := defaultConfig
	cfg.Default.MinTokens = 20
	cfg.Default.MinLines = 3
	matches, _, _, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("el bloque de lógica duplicado debía reportarse")
	}
	if matches[0].StartLineA <= 12 {
		t.Errorf("el match debe ubicarse en el cuerpo, no en la cabecera; startLineA=%d", matches[0].StartLineA)
	}
}

func TestScan_ignoreImportsFalseRestoresHeaderNoise(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "a.ts", nestHeader("alpha")+"\nfunction alphaOnly() { return 1; }\n")
	writeTempFile(t, dir, "b.ts", nestHeader("beta")+"\nclass BetaOnly { run() { return 2; } }\n")

	cfg := defaultConfig
	cfg.Default.MinTokens = 20
	off := false
	cfg.Default.IgnoreImports = &off
	matches, _, _, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("con ignoreImports=false la cabecera vuelve a colisionar (comportamiento previo)")
	}
}
