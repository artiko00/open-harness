package main

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestOperationalErr(t *testing.T) {
	if got := operationalErr(context.DeadlineExceeded); !strings.Contains(got.Error(), "timeout") {
		t.Errorf("deadline → %v", got)
	}
	if got := operationalErr(exec.ErrNotFound); got != errGitMissing {
		t.Errorf("notfound → %v", got)
	}
	generic := &gitCmdError{stderr: "fatal: bad object"}
	if got := operationalErr(generic); !strings.Contains(got.Error(), "bad object") {
		t.Errorf("genérico debe incluir stderr, got %v", got)
	}
}

func TestEnsureMeasurableGitAusente(t *testing.T) {
	f := &fakeGit{inside: &resp{err: &gitCmdError{err: exec.ErrNotFound}}}
	if err := ensureMeasurable(context.Background(), f.runner()); err != errGitMissing {
		t.Fatalf("git ausente → %v", err)
	}
}

func TestEnsureMeasurableNoRepo(t *testing.T) {
	f := &fakeGit{inside: &resp{err: &gitCmdError{stderr: "fatal: not a git repository"}}}
	if err := ensureMeasurable(context.Background(), f.runner()); err != errNotRepo {
		t.Fatalf("no repo → %v", err)
	}
}

func TestEnsureMeasurableTimeoutInside(t *testing.T) {
	f := &fakeGit{inside: &resp{err: context.DeadlineExceeded}}
	err := ensureMeasurable(context.Background(), f.runner())
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("timeout → %v", err)
	}
}

func TestEnsureMeasurableShallow(t *testing.T) {
	f := &fakeGit{shallow: &resp{out: "true\n"}}
	if err := ensureMeasurable(context.Background(), f.runner()); err != errShallow {
		t.Fatalf("shallow → %v", err)
	}
}

func TestEnsureMeasurableShallowError(t *testing.T) {
	f := &fakeGit{shallow: &resp{err: &gitCmdError{stderr: "boom"}}}
	if err := ensureMeasurable(context.Background(), f.runner()); err == nil {
		t.Fatal("esperaba error de shallow")
	}
}

func TestEnsureMeasurableOK(t *testing.T) {
	f := &fakeGit{}
	if err := ensureMeasurable(context.Background(), f.runner()); err != nil {
		t.Fatalf("ok → %v", err)
	}
}

func TestResolveBasePrecedencia(t *testing.T) {
	f := &fakeGit{refsOK: map[string]bool{"develop": true, "main": true}}
	ref, err := resolveBase(context.Background(), f.runner(), "develop", "main")
	if err != nil || ref != "develop" {
		t.Fatalf("--base debe ganar: %q %v", ref, err)
	}
}

func TestResolveBaseConfigYFallback(t *testing.T) {
	f := &fakeGit{refsOK: map[string]bool{"master": true}}
	ref, err := resolveBase(context.Background(), f.runner(), "", "")
	if err != nil || ref != "master" {
		t.Fatalf("fallback a master: %q %v", ref, err)
	}
}

func TestResolveBaseSinBase(t *testing.T) {
	f := &fakeGit{refsOK: map[string]bool{}}
	if _, err := resolveBase(context.Background(), f.runner(), "", ""); err != errBaseUnresolved {
		t.Fatalf("sin base → %v", err)
	}
}

// Una base EXPLÍCITA (--base o config) que no resuelve NO debe caer al fallback:
// es un error de medición (exit 2), no un pase silencioso contra la base equivocada.
func TestResolveBaseFlagInexistenteNoCaeAlFallback(t *testing.T) {
	f := &fakeGit{refsOK: map[string]bool{"main": true, "master": true}}
	_, err := resolveBase(context.Background(), f.runner(), "typo/ref", "")
	if err == nil {
		t.Fatal("--base inexistente debe fallar, no caer a main/master")
	}
	if !strings.Contains(err.Error(), "typo/ref") {
		t.Errorf("el error debe nombrar la ref, got %v", err)
	}
}

func TestResolveBaseConfigInexistenteNoCaeAlFallback(t *testing.T) {
	f := &fakeGit{refsOK: map[string]bool{"main": true}}
	if _, err := resolveBase(context.Background(), f.runner(), "", "typo/ref"); err == nil {
		t.Fatal("config base inexistente debe fallar, no caer al fallback")
	}
}

func TestBranchName(t *testing.T) {
	if got := branchName(context.Background(), (&fakeGit{}).runner()); got != "feature" {
		t.Errorf("nombre = %q", got)
	}
	errF := &fakeGit{abbrev: &resp{err: errors.New("x")}}
	if got := branchName(context.Background(), errF.runner()); got != "HEAD" {
		t.Errorf("error → %q", got)
	}
	emptyF := &fakeGit{abbrev: &resp{out: "\n"}}
	if got := branchName(context.Background(), emptyF.runner()); got != "HEAD" {
		t.Errorf("vacío → %q", got)
	}
}
