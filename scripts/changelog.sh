#!/usr/bin/env bash
#
# changelog.sh — genera entradas de changelog desde los commits convencionales,
# con git-cliff, por tool y para la suite.
#
# Preserva los CHANGELOG.md curados a mano: por defecto imprime un BORRADOR de la
# version sin publicar (para revisar y curar antes de commitear). Con `--write
# <tag>` prepende la seccion nueva al changelog existente sin tocar lo anterior.
#
#   scripts/changelog.sh                 # borrador a stdout, por tool + suite
#   scripts/changelog.sh --write v0.3.2  # prepende la seccion v0.3.2 a cada CHANGELOG.md
#
# Requiere git-cliff en PATH:  cargo install git-cliff  (o brew/apt/npm).

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

if ! command -v git-cliff >/dev/null 2>&1; then
  echo "git-cliff no esta en PATH. Instalalo con: cargo install git-cliff" >&2
  exit 2
fi

TOOLS=(linelens dupelens secretlens testlens scopelens)
CFG="cliff.toml"

MODE="draft"
TAG=""
if [[ "${1:-}" == "--write" ]]; then
  MODE="write"
  TAG="${2:?uso: changelog.sh --write <tag>   (ej. v0.3.2)}"
fi

gen_tool() {
  local t="$1" path="tools/$1/**" file="tools/$1/CHANGELOG.md"
  if [[ "${MODE}" == "write" ]]; then
    git-cliff --config "${CFG}" --include-path "${path}" --unreleased --tag "${TAG}" --prepend "${file}"
    echo "  prepend ${file} (${TAG})"
  else
    echo "=== ${t} ==="
    git-cliff --config "${CFG}" --include-path "${path}" --unreleased 2>/dev/null || echo "(sin cambios convencionales)"
    echo
  fi
}

for t in "${TOOLS[@]}"; do
  gen_tool "${t}"
done

# Suite / meta (todos los commits).
if [[ "${MODE}" == "write" ]]; then
  git-cliff --config "${CFG}" --unreleased --tag "${TAG}" --prepend CHANGELOG.md
  echo "  prepend CHANGELOG.md (${TAG})"
else
  echo "=== suite (root) ==="
  git-cliff --config "${CFG}" --unreleased 2>/dev/null || echo "(sin cambios convencionales)"
fi
