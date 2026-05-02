package main

import (
	"path/filepath"
	"strings"
)

func isExcluded(relPath string, excludes []string) bool {
	relPath = filepath.ToSlash(relPath)
	parts := strings.Split(relPath, "/")

	for _, excl := range excludes {
		excl = filepath.ToSlash(excl)

		for _, part := range parts {
			if matched, _ := filepath.Match(excl, part); matched {
				return true
			}
		}

		if matchGlob(excl, relPath) {
			return true
		}
	}
	return false
}

func matchGlob(pattern, relPath string) bool {
	relPath = filepath.ToSlash(relPath)
	pattern = filepath.ToSlash(pattern)

	if !strings.Contains(pattern, "**") {
		matched, _ := filepath.Match(pattern, relPath)
		if matched {
			return true
		}
		matched, _ = filepath.Match(pattern, filepath.Base(relPath))
		return matched
	}

	parts := strings.SplitN(pattern, "**/", 2)
	if len(parts) != 2 {
		matched, _ := filepath.Match(pattern, relPath)
		return matched
	}

	prefix, suffix := parts[0], parts[1]
	if prefix != "" && !strings.HasPrefix(relPath, prefix) {
		return false
	}

	base := filepath.Base(relPath)
	if matched, _ := filepath.Match(suffix, base); matched {
		return true
	}

	pathParts := strings.Split(relPath, "/")
	for i := range pathParts {
		sub := strings.Join(pathParts[i:], "/")
		if matched, _ := filepath.Match(suffix, sub); matched {
			return true
		}
	}
	return false
}
