package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	errGitMissing     = errors.New("git no se encontró en PATH; instalá git para poder medir el alcance del cambio")
	errNotRepo        = errors.New("el directorio no es un repositorio git; ejecutá scopelens dentro de un repo")
	errShallow        = errors.New("el repositorio es un clon shallow; usá fetch-depth: 0 para obtener el historial completo")
	errBaseUnresolved = errors.New("no se pudo resolver la rama base (probé origin/HEAD, main, master); pasá --base <ref> explícitamente")
)

// operationalErr traduce el error de una invocación a git en un error de
// medición (exit 2), distinguiendo timeout de ausencia de git.
func operationalErr(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("git excedió el timeout de %s; el repositorio puede estar bloqueado", gitTimeout)
	}
	if errors.Is(err, exec.ErrNotFound) {
		return errGitMissing
	}
	return fmt.Errorf("git falló: %w", err)
}

// ensureMeasurable verifica que se pueda medir: git presente, cwd es repo y no
// es un clon shallow. Cada fallo devuelve un error accionable (exit 2).
func ensureMeasurable(ctx context.Context, run gitRunner) error {
	if _, err := run(ctx, "rev-parse", "--is-inside-work-tree"); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, exec.ErrNotFound) {
			return operationalErr(err)
		}
		return errNotRepo
	}
	out, err := run(ctx, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return operationalErr(err)
	}
	if strings.TrimSpace(string(out)) == "true" {
		return errShallow
	}
	return nil
}

// resolveBase elige la rama base. Si el usuario fijó una base explícita (--base
// o config), esa base DEBE resolver: si no existe es un error de medición
// (exit 2), no un fallback silencioso a otra base — un typo en --base no puede
// pasar un presupuesto contra la rama equivocada. Sin base explícita, prueba
// origin/HEAD > main > master.
func resolveBase(ctx context.Context, run gitRunner, flagBase, cfgBase string) (string, error) {
	explicit := flagBase
	if explicit == "" {
		explicit = cfgBase
	}
	if explicit != "" {
		if _, err := run(ctx, "rev-parse", "--verify", "--quiet", explicit); err == nil {
			return explicit, nil
		}
		return "", fmt.Errorf("la base %q no existe; verificá el nombre de la ref pasada en --base o config", explicit)
	}
	for _, ref := range []string{"origin/HEAD", "main", "master"} {
		if _, err := run(ctx, "rev-parse", "--verify", "--quiet", ref); err == nil {
			return ref, nil
		}
	}
	return "", errBaseUnresolved
}

// branchName devuelve el nombre de la rama de HEAD para el encabezado, o "HEAD"
// si está detached o el repo no tiene commits.
func branchName(ctx context.Context, run gitRunner) string {
	out, err := run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "HEAD"
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return "HEAD"
	}
	return name
}
