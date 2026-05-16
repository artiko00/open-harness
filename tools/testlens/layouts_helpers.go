package main

import (
	"path/filepath"
	"strings"
)

// splitSegments turns a path string into its non-empty path components.
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

// indexOfSubslice returns the start index of needle inside haystack, or -1.
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
