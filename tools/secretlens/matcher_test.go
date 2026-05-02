package main

import "testing"

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"**/*.go", "src/main.go", true},
		{"**/*.go", "main.go", true},
		{"**/*.go", "src/main.ts", false},
		{"*.go", "main.go", true},
		{"*.go", "src/main.go", true},
		{"src/*.go", "src/main.go", true},
		{"src/*.go", "other/main.go", false},
		{"**/migrations/**", "db/migrations/001.sql", true},
		{"**/migrations/**", "src/auth.go", false},
		{"node_modules", "node_modules", true},
		{"src/**", "src/main.go", true},
		{"src/**", "other/main.go", false},
		{"src/**/config.go", "src/deep/config.go", true},
		{"src/**/config.go", "other/deep/config.go", false},
		{"**", "foo.go", true},
	}

	for _, c := range cases {
		got := matchGlob(c.pattern, c.path)
		if got != c.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", c.pattern, c.path, got, c.want)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	excludes := []string{"node_modules", ".git", "dist", "**/*.min.js", "*.lock"}

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
		{"go.sum", false},
		{"yarn.lock", true},
	}

	for _, c := range cases {
		got := isExcluded(c.path, excludes)
		if got != c.want {
			t.Errorf("isExcluded(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
