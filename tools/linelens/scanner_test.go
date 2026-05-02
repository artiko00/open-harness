package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func lines(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("line\n")
	}
	return b.String()
}

func TestScan_DetectsViolations(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "short.go", lines(50))
	writeFile(t, dir, "long.go", lines(150))
	writeFile(t, dir, "node_modules/pkg/index.js", lines(200))

	cfg := Config{
		Default: DefaultConfig{MaxLines: 100},
		Exclude: []string{"node_modules"},
	}

	results, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatal(err)
	}

	var violations []FileResult
	for _, r := range results {
		if r.IsViolation() {
			violations = append(violations, r)
		}
	}

	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(violations))
	}
	if len(violations) > 0 && violations[0].RelPath != "long.go" {
		t.Errorf("expected long.go violation, got %s", violations[0].RelPath)
	}
}

func TestScan_RuleOverride(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "auth.service.ts", lines(150))
	writeFile(t, dir, "auth.spec.ts", lines(250))

	cfg := Config{
		Default: DefaultConfig{MaxLines: 100},
		Rules: []Rule{
			{Pattern: "**/*.spec.*", MaxLines: 300},
		},
		Exclude: []string{},
	}

	results, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range results {
		if strings.HasSuffix(r.RelPath, ".spec.ts") && r.IsViolation() {
			t.Errorf("spec file should not be a violation (250 < 300): %+v", r)
		}
		if strings.HasSuffix(r.RelPath, ".service.ts") && !r.IsViolation() {
			t.Errorf("service file should be a violation (150 > 100): %+v", r)
		}
	}
}

func TestScan_SkipRule(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "db/migrations/001_init.sql", lines(500))

	cfg := Config{
		Default: DefaultConfig{MaxLines: 100},
		Rules: []Rule{
			{Pattern: "**/migrations/**", Skip: true},
		},
		Exclude: []string{},
	}

	results, err := scan(dir, cfg, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 0 {
		t.Errorf("migration files should be skipped, got %d results", len(results))
	}
}

func TestScan_MaxOverride(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "big.py", lines(180))

	cfg := Config{
		Default: DefaultConfig{MaxLines: 100},
		Exclude: []string{},
	}

	results, err := scan(dir, cfg, 200)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range results {
		if r.IsViolation() {
			t.Errorf("180 lines should pass with --max 200 override")
		}
		if r.MaxLines != 200 {
			t.Errorf("override not applied, got maxLines=%d", r.MaxLines)
		}
	}
}
