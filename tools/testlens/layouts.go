package main

import (
	"path/filepath"
	"strings"
)

// testSearchDirs returns the list of directories where the test for
// sourcePath might live, according to the language's layout policy.
// Order matters: cheaper / more specific lookups first.
func testSearchDirs(sourcePath string, lang languageMapping) []string {
	sourceDir := filepath.Dir(sourcePath)
	dirs := []string{sourceDir}

	for _, td := range lang.testDirs {
		dirs = append(dirs, filepath.Join(sourceDir, td))
	}

	for _, m := range lang.mirrors {
		if mirrored, ok := applyMirror(sourceDir, m[0], m[1]); ok {
			dirs = append(dirs, mirrored)
		}
	}

	return dirs
}

// applyMirror swaps the first occurrence of fromPrefix with toPrefix in
// the given path's segments. Returns (newPath, true) when the prefix
// matched a contiguous run of segments, (path, false) otherwise.
//
// Example: applyMirror("src/main/java/com/foo", "src/main/java", "src/test/java")
//          → "src/test/java/com/foo", true
func applyMirror(path, fromPrefix, toPrefix string) (string, bool) {
	fromSegs := splitSegments(fromPrefix)
	toSegs := splitSegments(toPrefix)
	pathSegs := splitSegments(path)

	idx := indexOfSubslice(pathSegs, fromSegs)
	if idx < 0 {
		return path, false
	}

	out := make([]string, 0, len(pathSegs)-len(fromSegs)+len(toSegs))
	out = append(out, pathSegs[:idx]...)
	out = append(out, toSegs...)
	out = append(out, pathSegs[idx+len(fromSegs):]...)

	joined := filepath.Join(out...)
	if filepath.IsAbs(path) {
		joined = string(filepath.Separator) + joined
	}
	return joined, true
}

func splitSegments(p string) []string {
	if p == "" {
		return nil
	}
	parts := strings.Split(filepath.ToSlash(p), "/")
	out := parts[:0]
	for _, s := range parts {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func indexOfSubslice(haystack, needle []string) int {
	if len(needle) == 0 || len(haystack) < len(needle) {
		return -1
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if equalSegments(haystack[i:i+len(needle)], needle) {
			return i
		}
	}
	return -1
}

func equalSegments(a, b []string) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
