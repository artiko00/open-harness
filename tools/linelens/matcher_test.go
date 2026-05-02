package main

import "testing"

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"**/*.spec.ts", "src/components/foo.spec.ts", true},
		{"**/*.spec.ts", "foo.spec.ts", true},
		{"**/*.spec.ts", "src/foo.ts", false},
		{"**/*.test.*", "tests/auth.test.py", true},
		{"**/*.test.*", "tests/auth_test.go", false},
		{"**/*_test.go", "pkg/util/math_test.go", true},
		{"**/*_test.go", "pkg/util/math.go", false},
		{"*.go", "main.go", true},
		{"*.go", "src/main.go", true},   // sin "/" en patrón → aplica al filename (gitignore-style)
		{"src/*.go", "src/main.go", true},
		{"src/*.go", "other/main.go", false}, // con "/" → path exacto
		{"**/migrations/**", "db/migrations/001_init.sql", true},
		{"**/migrations/**", "src/auth.go", false},
		{"node_modules", "node_modules", true},
		{"src/**", "src/main.go", true},
		{"src/**", "other/main.go", false},
		{"src/**/config.go", "src/deep/config.go", true},
		{"src/**/config.go", "other/deep/config.go", false},
	}

	for _, c := range cases {
		got := matchGlob(c.pattern, c.path)
		if got != c.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", c.pattern, c.path, got, c.want)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	excludes := []string{"node_modules", ".git", "dist", "**/*.min.js"}

	cases := []struct {
		path string
		want bool
	}{
		{"node_modules/lodash/index.js", true},
		{"src/node_modules/foo.ts", true},
		{".git/config", true},
		{"dist/bundle.js", true},
		{"src/app.js", false},
		{"src/vendor/foo.min.js", true},
		{"src/app.ts", false},
	}

	for _, c := range cases {
		got := isExcluded(c.path, excludes)
		if got != c.want {
			t.Errorf("isExcluded(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestRuleForFile(t *testing.T) {
	rules := []Rule{
		{Pattern: "**/*.spec.*", MaxLines: 300},
		{Pattern: "**/migrations/**", Skip: true},
	}

	r := ruleForFile("src/auth.spec.ts", rules)
	if r == nil || r.MaxLines != 300 {
		t.Errorf("expected spec rule, got %v", r)
	}

	r = ruleForFile("db/migrations/001.sql", rules)
	if r == nil || !r.Skip {
		t.Errorf("expected skip rule, got %v", r)
	}

	r = ruleForFile("src/auth.ts", rules)
	if r != nil {
		t.Errorf("expected nil rule, got %v", r)
	}
}
