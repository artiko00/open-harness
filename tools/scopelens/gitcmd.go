package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// gitTimeout acota cada invocación a git; superarlo es error operativo (exit 2).
const gitTimeout = 15 * time.Second

// gitRunner ejecuta git con los argumentos dados y devuelve su stdout. Aísla el
// adaptador de disco: los tests inyectan fakes con salidas fijas.
type gitRunner func(ctx context.Context, args ...string) ([]byte, error)

// gitCmdError transporta el stderr de git para que el mensaje de exit 2 nombre
// la causa. Unwrap expone el error subyacente (p. ej. exec.ErrNotFound).
type gitCmdError struct {
	stderr string
	err    error
}

func (e *gitCmdError) Error() string {
	if e.stderr != "" {
		return e.stderr
	}
	return e.err.Error()
}

func (e *gitCmdError) Unwrap() error { return e.err }

// newGitRunner devuelve un gitRunner real que invoca git en dir con
// --no-optional-locks para no interferir con otros procesos git en curso.
func newGitRunner(dir string) gitRunner {
	return func(ctx context.Context, args ...string) ([]byte, error) {
		full := append([]string{"-C", dir, "--no-optional-locks"}, args...)
		cmd := exec.CommandContext(ctx, "git", full...)
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			if ctx.Err() != nil {
				return out.Bytes(), ctx.Err()
			}
			return out.Bytes(), &gitCmdError{stderr: strings.TrimSpace(stderr.String()), err: err}
		}
		return out.Bytes(), nil
	}
}
