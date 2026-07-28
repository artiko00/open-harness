package main

import (
	"errors"
	"strings"
	"testing"
)

// fullRepo devuelve un fakeGit con HEAD y base main resolubles.
func fullRepo(staged, committed string) *fakeGit {
	return &fakeGit{
		refsOK:    map[string]bool{"HEAD": true, "main": true},
		staged:    &resp{out: staged},
		committed: &resp{out: committed},
	}
}

func TestMeasureUnionDedup(t *testing.T) {
	f := fullRepo("src/a.ts\n", "src/a.ts\nsrc/b.ts\n")
	res, err := measure(f.runner(), "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 2 {
		t.Fatalf("union con duplicado debe dar 2: %v", res.Files)
	}
	if res.Files[0] != "src/a.ts" || res.Files[1] != "src/b.ts" {
		t.Fatalf("orden lexicográfico esperado: %v", res.Files)
	}
	if res.Base != "main" || res.MergeBase != "abcdef1" {
		t.Fatalf("encabezado mal: %+v", res)
	}
}

func TestMeasureCincoCommits(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 20; i++ {
		b.WriteString("file")
		b.WriteByte(byte('a' + i))
		b.WriteString(".go\n")
	}
	f := fullRepo("", b.String())
	res, err := measure(f.runner(), "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 20 {
		t.Fatalf("esperaba 20, got %d", len(res.Files))
	}
}

func TestMeasureStagedOnly(t *testing.T) {
	f := fullRepo("s1.ts\ns2.ts\ns3.ts\ns4.ts\n", "c1.ts\nc2.ts\n")
	res, err := measure(f.runner(), "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 4 {
		t.Fatalf("staged-only debe ignorar commits: %v", res.Files)
	}
	if res.Base != "" || res.MergeBase != "" {
		t.Fatalf("staged-only no debe resolver base: %+v", res)
	}
}

func TestMeasureSinCommits(t *testing.T) {
	f := &fakeGit{
		refsOK: map[string]bool{}, // HEAD no existe
		staged: &resp{out: "a.txt\nb.txt\n"},
	}
	res, err := measure(f.runner(), "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 2 || res.MergeBase != "" {
		t.Fatalf("sin commits cuenta sólo staged: %+v", res)
	}
}

func TestMeasureEnsureError(t *testing.T) {
	f := &fakeGit{shallow: &resp{out: "true\n"}}
	if _, err := measure(f.runner(), "", "", false); err != errShallow {
		t.Fatalf("shallow → %v", err)
	}
}

func TestMeasureStagedError(t *testing.T) {
	f := fullRepo("", "")
	f.staged = &resp{err: &gitCmdError{stderr: "boom"}}
	if _, err := measure(f.runner(), "", "", false); err == nil {
		t.Fatal("esperaba error de staged")
	}
}

func TestMeasureBaseUnresolved(t *testing.T) {
	f := &fakeGit{refsOK: map[string]bool{"HEAD": true}, staged: &resp{out: ""}}
	if _, err := measure(f.runner(), "", "", false); err != errBaseUnresolved {
		t.Fatalf("base → %v", err)
	}
}

func TestMeasureMergeBaseError(t *testing.T) {
	f := fullRepo("", "")
	f.mergeBase = &resp{err: errors.New("boom")}
	if _, err := measure(f.runner(), "", "", false); err == nil {
		t.Fatal("esperaba error de merge-base")
	}
}

func TestMeasureShortError(t *testing.T) {
	f := fullRepo("", "")
	f.short = &resp{err: errors.New("boom")}
	if _, err := measure(f.runner(), "", "", false); err == nil {
		t.Fatal("esperaba error de rev-parse --short")
	}
}

func TestMeasureCommittedError(t *testing.T) {
	f := fullRepo("", "")
	f.committed = &resp{err: &gitCmdError{stderr: "boom"}}
	if _, err := measure(f.runner(), "", "", false); err == nil {
		t.Fatal("esperaba error de diff committed")
	}
}
