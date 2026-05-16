package main

import (
	"path/filepath"
	"strings"
)

// testSearchDirs returns the list of directories where the test for
// sourcePath might live, according to the language's layout policy.
// Order matters: cheaper / more specific lookups first.
//
// When root is non-empty, ancestor walk-up is enabled: for each testDir,
// every ancestor of the source between sourceDir and root is also probed
// with the relative path appended, so that centralised __tests__/ layouts
// that mirror the source tree (F-016) match.
func testSearchDirs(sourcePath, root string, lang languageMapping) []string {
	sourceDir := filepath.Dir(sourcePath)
	dirs := []string{sourceDir}

	for _, td := range lang.testDirs {
		dirs = append(dirs, filepath.Join(sourceDir, td))
		dirs = append(dirs, ancestorTestDirs(sourceDir, root, td)...)
	}

	for _, m := range lang.mirrors {
		if mirrored, ok := applyMirror(sourceDir, m[0], m[1]); ok {
			dirs = append(dirs, mirrored)
		}
	}

	return dirs
}

// ancestorTestDirs returns dirs like <ancestor>/<testDir>/<rel> for each
// ancestor of sourceDir (excluding sourceDir itself) up to (and including) root.
func ancestorTestDirs(sourceDir, root, testDir string) []string {
	if root == "" {
		return nil
	}
	absSource, err1 := filepath.Abs(sourceDir)
	absRoot, err2 := filepath.Abs(root)
	if err1 != nil || err2 != nil || !strings.HasPrefix(absSource, absRoot) {
		return nil
	}

	var dirs []string
	rel := ""
	ancestor := sourceDir
	for ancestor != root && filepath.Dir(ancestor) != ancestor {
		rel = filepath.Join(filepath.Base(ancestor), rel)
		parent := filepath.Dir(ancestor)
		dirs = append(dirs, filepath.Join(parent, testDir, rel))
		if absParent, _ := filepath.Abs(parent); absParent == absRoot {
			break
		}
		ancestor = parent
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

