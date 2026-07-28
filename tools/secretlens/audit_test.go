package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Los secretos del fixture se ENSAMBLAN por partes en tiempo de ejecución para
// que ningún literal completo con formato de credencial real quede en el código
// fuente (los escáneres de secretos —incluido secretlens— los marcarían). El
// valor ensamblado sí matchea los patrones; las partes sueltas no.
func pre(prefix, body string) string { return prefix + body }

// auditSecret es un secreto del fixture con una subcadena única para verificarlo.
type auditSecret struct {
	line   string // línea completa tal como aparece en el archivo escaneado
	needle string // subcadena única esperada en el hallazgo
}

func auditSecrets() []auditSecret {
	stripe := pre("sk_", "live_51H8xR2eZvKYlo2CqDm4nFpQrStUvWxYz012")
	slack := pre("xox", "b-2401234567-2401234567890-AbCdEfGhIjKlMnOpQrStUvWx")
	google := pre("AIza", "SyC3aBcDeFgHiJkLmNoPqRsTuVwXyZ012345")
	openai := pre("sk-", "proj-aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789")
	gitlab := pre("glpat-", "aBcDeFgHiJkLmNoPqRs1")
	npmtok := pre("npm_", "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ab")
	sendgrid := pre("SG.", "aBcDeFgHiJkLmNoPqRsTuV.wXyZ0123456789aBcDeFgHiJkLmNoPqRsTuVwXyZ012")
	ghp := pre("ghp_", "aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789ab")
	aws := pre("AKIA", "QNB7Q7AIIVSMBGPF")
	awsSecret := "kLnQ8vTpR3mWxY7zA1bC4dE6fG9hJ2kMnP5qRs8t"
	jwt := pre("eyJ", "hbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0In0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV")
	slackHook := pre("hooks.slack.com/services/", "T00000000/B11111111/aBcDeFgHiJkLmNoPqRsTuVwX")
	ghoTok := pre("gho_", "16C7e42F292c6912E7710c838347Ae178B4a")
	pkBody := "mK9pL2xQ7wZ4tR8nB3vC6dF1gH5j"
	return []auditSecret{
		{"STRIPE_KEY=" + stripe, stripe},
		{"SLACK_BOT=" + slack, slack},
		{"GOOGLE_API=" + google, google},
		{"OPENAI_KEY=" + openai, openai},
		{"GITLAB_PAT=" + gitlab, gitlab},
		{"NPM_TOKEN=" + npmtok, npmtok},
		{"SENDGRID_KEY=" + sendgrid, sendgrid},
		{"GITHUB_TOKEN=" + ghp, ghp},
		{"AWS_ACCESS_KEY_ID=" + aws, aws},
		{"aws_secret_access_key = \"" + awsSecret + "\"", awsSecret},
		{"JWT=" + jwt, jwt},
		{"SLACK_HOOK=https://" + slackHook, "T00000000"},
		{"DB_URL=postgres://admin:SuperSecretPass99@db.internal:5432/app", "SuperSecretPass99"},
		{"CACHE_URL=redis://default:R3disStrongPwd42@cache.internal:6379/0", "R3disStrongPwd42"},
		{"MONGO_URL=mongodb+srv://svc:M0ngoStrongPwd77@cluster0.mongodb.net/prod", "M0ngoStrongPwd77"},
		{"API_KEY=aB3xK9mQ2pL7wZ4tR8n", "aB3xK9mQ2pL7wZ4tR8n"},
		{"password: \"P@ssw0rdVeryS3cret!\"", "P@ssw0rdVeryS3cret!"},
		{"auth_token = \"" + ghoTok + "\"", ghoTok},
		{"-----" + pre("BEGIN ", "RSA PRIVATE KEY") + "-----", "PRIVATE KEY"},
		{"private_key: \"" + pkBody + "\"", pkBody},
	}
}

var auditDecoys = []string{
	"DEBUG=true", "your_key_here", "changeme", "placeholder",
	"xxxxxxxxxxxxxxxx", "LOG_LEVEL=debug", "aaaaaaaa",
}

// writeAuditFixture materializa el fixture en un dir temporal a partir de los
// secretos ensamblados + los señuelos, y devuelve el dir.
func writeAuditFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	var b strings.Builder
	b.WriteString("# Fixture de auditoria FASE 5\n\n")
	for _, s := range auditSecrets() {
		b.WriteString(s.line + "\n")
	}
	for _, d := range auditDecoys {
		b.WriteString(d + "\n")
	}
	// Asignación partida en líneas consecutivas (JSON pretty) para ejercitar el
	// lookahead de una línea (tarea 5.16).
	b.WriteString("{\n  \"private_key\":\n    \"" + "aX9bYc2dZ4eW6fV8gT1hR3jQ5kP7mN0" + "\"\n}\n")
	if err := os.WriteFile(filepath.Join(dir, "leak-config.txt"), []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func auditConfig() Config {
	var cfg Config
	applyConfigDefaults(&cfg)
	return cfg
}

func TestAudit_RecallYPrecision(t *testing.T) {
	dir := writeAuditFixture(t)
	findings, _, _, err := scan(dir, auditConfig())
	if err != nil {
		t.Fatal(err)
	}

	secrets := auditSecrets()
	detected := 0
	for _, sec := range secrets {
		if algunHallazgoContiene(findings, sec.needle) {
			detected++
		} else {
			t.Logf("secreto NO detectado: %s", sec.needle)
		}
	}
	if detected < 18 {
		t.Errorf("recall = %d/20, want >= 18/20", detected)
	}

	needles := make([]string, len(secrets))
	for i, s := range secrets {
		needles[i] = s.needle
	}
	truePos := 0
	for _, f := range findings {
		if contieneAlguno(f.Content, needles) {
			truePos++
		}
	}
	precision := float64(truePos) / float64(len(findings))
	if precision < 0.8 {
		t.Errorf("precision = %.2f (%d/%d), want >= 0.80", precision, truePos, len(findings))
	}

	for _, d := range auditDecoys {
		if algunHallazgoContiene(findings, d) {
			t.Errorf("señuelo detectado como secreto: %s", d)
		}
	}
}

func algunHallazgoContiene(findings []Finding, sub string) bool {
	for _, f := range findings {
		if strings.Contains(f.Content, sub) {
			return true
		}
	}
	return false
}

func contieneAlguno(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
