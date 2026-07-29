package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// resp es una respuesta canned de git: salida y/o error.
type resp struct {
	out string
	err error
}

func (r *resp) resolve(def string) ([]byte, error) {
	if r == nil {
		return []byte(def), nil
	}
	if r.err != nil {
		return nil, r.err
	}
	return []byte(r.out), nil
}

// fakeGit despacha por los args exactos que emite measure. Los campos nil caen
// a un default exitoso, de modo que cada test sólo declara lo que le importa.
type fakeGit struct {
	inside    *resp
	shallow   *resp
	abbrev    *resp
	refsOK    map[string]bool // refs que verifican (incluye "HEAD")
	mergeBase *resp
	short     *resp
	staged    *resp
	committed *resp
	// Numstat opcional: nil devuelve churn vacío (los tests que no miden líneas
	// no necesitan declararlo).
	stagedNum    *resp
	committedNum *resp
}

var errRef = errors.New("ref inexistente")

func (f *fakeGit) run(_ context.Context, args ...string) ([]byte, error) {
	j := strings.Join(args, " ")
	switch {
	case j == "rev-parse --is-inside-work-tree":
		return f.inside.resolve("true")
	case j == "rev-parse --is-shallow-repository":
		return f.shallow.resolve("false")
	case j == "rev-parse --abbrev-ref HEAD":
		return f.abbrev.resolve("feature")
	case strings.HasPrefix(j, "rev-parse --verify --quiet "):
		if f.refsOK[args[3]] {
			return []byte(args[3] + "\n"), nil
		}
		return nil, errRef
	case args[0] == "merge-base":
		return f.mergeBase.resolve("abcdef1234567890")
	case j == "rev-parse --short abcdef1234567890" || (len(args) == 3 && args[1] == "--short"):
		return f.short.resolve("abcdef1")
	case args[0] == "diff" && contains(args, "--numstat") && contains(args, "--cached"):
		return f.stagedNum.resolve("")
	case args[0] == "diff" && contains(args, "--numstat"):
		return f.committedNum.resolve("")
	case args[0] == "diff" && contains(args, "--cached"):
		return f.staged.resolve("")
	case args[0] == "diff":
		return f.committed.resolve("")
	}
	return nil, fmt.Errorf("args de git inesperados: %v", args)
}

func (f *fakeGit) runner() gitRunner { return f.run }

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
