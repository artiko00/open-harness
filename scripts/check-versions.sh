#!/usr/bin/env bash
#
# check-versions.sh — verifica que la version de cada tool este sincronizada
# en TODAS sus fuentes, incluidos los manifiestos de distribucion npm y PyPI.
#
# Fuente de verdad: el `const version` de cada tools/<tool>/main.go.
# Se compara contra:
#   - el bloque del tool en open-harness.json (y la version del manifiesto),
#   - cada mencion de version (vX.Y.Z) en README.md y AGENTS.md,
#   - npm: package.json del tool, de sus 4 paquetes de plataforma, y los pins
#     (optionalDependencies del tool -> plataformas; dependencies del meta -> tools),
#   - PyPI: pyproject.toml del tool, del meta, sus pins, y __version__ del meta.
#
# Sale con 1 si alguna diverge, 0 si todas coinciden. Un solo comando verificable
# antes de publicar.

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT}" || exit 2

TOOLS=(linelens dupelens secretlens testlens)
PLATFORMS=(linux-x64 darwin-arm64 darwin-x64 win32-x64)
MANIFEST="open-harness.json"
DOCS=(README.md AGENTS.md)
NPM_DIR="npm/@open_harness"
META_NPM="${NPM_DIR}/open-harness/package.json"
META_PY="pypi/open_harness/pyproject.toml"
META_PY_INIT="pypi/open_harness/src/open_harness/__init__.py"

ok=1
pass() { printf '  OK    %s\n' "$1"; }
fail() { printf '  DIFF  %s\n' "$1"; ok=0; }

# expect <label> <actual> <expected>: compara y reporta.
expect() {
  if [[ "$2" == "$3" ]]; then
    pass "$1 = $2"
  else
    fail "$1 = ${2:-<vacio>} (esperado $3)"
  fi
}

# --- lectores de version por tipo de archivo (grep/sed, sin dependencias) ---

const_version() { grep -E '^const version = "' "$1" | sed -E 's/.*"([^"]+)".*/\1/' | head -n1; }

# .version de un package.json (primera aparicion del campo).
pkg_version() { grep -m1 -E '"version"[[:space:]]*:' "$1" 2>/dev/null | sed -E 's/.*"version"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/'; }

# version de un pin exacto "<name>": "X" en un package.json (evita colisiones
# entre @open_harness/linelens y @open_harness/linelens-linux-x64).
pkg_dep_version() { grep -E "\"$2\"[[:space:]]*:" "$1" 2>/dev/null | sed -E 's/.*:[[:space:]]*"([0-9][^"]*)".*/\1/' | head -n1; }

# version = "X" de un pyproject.toml.
pyproj_version() { grep -m1 -E '^version[[:space:]]*=' "$1" 2>/dev/null | sed -E 's/.*"([^"]+)".*/\1/'; }

# pin open-harness-<tool>==X de un pyproject.toml.
pyproj_pin() { grep -E "open-harness-$2==" "$1" 2>/dev/null | sed -E 's/.*==([0-9][0-9.]*).*/\1/' | head -n1; }

# __version__ = "X" de un __init__.py.
py_dunder_version() { grep -E '__version__[[:space:]]*=' "$1" 2>/dev/null | sed -E 's/.*"([^"]+)".*/\1/' | head -n1; }

manifest_tool_version() {
  awk -v tool="$1" '
    /"name":[[:space:]]*"/ { cur=$0; gsub(/^.*"name":[[:space:]]*"/,"",cur); gsub(/".*$/,"",cur); in_tool=(cur==tool) }
    in_tool && /"version":[[:space:]]*"/ { v=$0; gsub(/^.*"version":[[:space:]]*"/,"",v); gsub(/".*$/,"",v); print v; exit }
  ' "${MANIFEST}"
}
manifest_top_version() {
  awk '/"version":[[:space:]]*"/ { v=$0; gsub(/^.*"version":[[:space:]]*"/,"",v); gsub(/".*$/,"",v); print v; exit }' "${MANIFEST}"
}
doc_tool_versions() {
  grep -E "\b$2\b" "$1" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sed -E 's/^v//' | sort -u
}

# check_tool_dist <tool> <expected>: verifica npm (tool + plataformas + pins) y PyPI.
check_tool_dist() {
  local tool="$1" expected="$2"
  local f="${NPM_DIR}/${tool}/package.json"
  if [[ -f "$f" ]]; then
    expect "npm ${tool}/package.json" "$(pkg_version "$f")" "$expected"
  else
    fail "npm ${tool}/package.json: no existe"
  fi
  local plat pf
  for plat in "${PLATFORMS[@]}"; do
    pf="${NPM_DIR}/${tool}-${plat}/package.json"
    if [[ -f "$pf" ]]; then
      expect "npm ${tool}-${plat}" "$(pkg_version "$pf")" "$expected"
    else
      fail "npm ${tool}-${plat}/package.json: no existe"
    fi
    if [[ -f "$f" ]]; then
      expect "npm ${tool} pin ${plat}" "$(pkg_dep_version "$f" "@open_harness/${tool}-${plat}")" "$expected"
    fi
  done
  local py="pypi/open_harness_${tool}/pyproject.toml"
  if [[ -f "$py" ]]; then
    expect "pypi open_harness_${tool}" "$(pyproj_version "$py")" "$expected"
  else
    fail "pypi open_harness_${tool}/pyproject.toml: no existe"
  fi
}

echo "Verificando sincronizacion de versiones (fuente: const version de main.go)"
echo

REFERENCE=""

for tool in "${TOOLS[@]}"; do
  main="tools/${tool}/main.go"
  [[ -f "${main}" ]] || { fail "${tool}: no existe ${main}"; continue; }
  expected="$(const_version "${main}")"
  [[ -n "${expected}" ]] || { fail "${tool}: no se pudo leer const version en ${main}"; continue; }

  echo "[${tool}] const version (fuente de verdad) = ${expected}"
  if [[ -z "${REFERENCE}" ]]; then REFERENCE="${expected}"
  elif [[ "${expected}" != "${REFERENCE}" ]]; then fail "${tool}: const version ${expected} difiere de la referencia ${REFERENCE}"; fi

  expect "${MANIFEST} (tool ${tool})" "$(manifest_tool_version "${tool}")" "${expected}"
  for doc in "${DOCS[@]}"; do
    [[ -f "${doc}" ]] || { fail "${doc}: no existe"; continue; }
    versions="$(doc_tool_versions "${doc}" "${tool}")"
    [[ -n "${versions}" ]] || { fail "${doc} (${tool}): sin mencion de version vX.Y.Z"; continue; }
    while IFS= read -r v; do [[ -z "${v}" ]] || expect "${doc} (${tool})" "${v}" "${expected}"; done <<< "${versions}"
  done
  check_tool_dist "${tool}" "${expected}"
  echo
done

# scopelens: ciclo de versiones propio (0.1.0), no comparte REFERENCE.
SCOPELENS_VER=""
sl_main="tools/scopelens/main.go"
if [[ -f "${sl_main}" ]]; then
  SCOPELENS_VER="$(const_version "${sl_main}")"
  if [[ -z "${SCOPELENS_VER}" ]]; then
    fail "scopelens: no se pudo leer const version en ${sl_main}"
  else
    echo "[scopelens] const version (fuente de verdad) = ${SCOPELENS_VER}"
    expect "${MANIFEST} (tool scopelens)" "$(manifest_tool_version scopelens)" "${SCOPELENS_VER}"
    for doc in "${DOCS[@]}"; do
      [[ -f "${doc}" ]] || { fail "${doc}: no existe"; continue; }
      versions="$(doc_tool_versions "${doc}" scopelens)"
      [[ -z "${versions}" ]] && continue
      while IFS= read -r v; do [[ -z "${v}" ]] || expect "${doc} (scopelens)" "${v}" "${SCOPELENS_VER}"; done <<< "${versions}"
    done
    check_tool_dist scopelens "${SCOPELENS_VER}"
    echo
  fi
fi

# --- Meta-paquete (comparte REFERENCE con los 4 nucleos) ---
echo "[meta] open-harness (referencia ${REFERENCE})"
expect "${MANIFEST} (version raiz)" "$(manifest_top_version)" "${REFERENCE}"

# npm meta: version + dependencies -> cada tool a su version.
if [[ -f "${META_NPM}" ]]; then
  expect "npm meta open-harness" "$(pkg_version "${META_NPM}")" "${REFERENCE}"
  for tool in "${TOOLS[@]}"; do
    expect "npm meta dep ${tool}" "$(pkg_dep_version "${META_NPM}" "@open_harness/${tool}")" "${REFERENCE}"
  done
  if [[ -n "${SCOPELENS_VER}" ]]; then
    expect "npm meta dep scopelens" "$(pkg_dep_version "${META_NPM}" "@open_harness/scopelens")" "${SCOPELENS_VER}"
  fi
else
  fail "npm meta: no existe ${META_NPM}"
fi

# pypi meta: version + __version__ + pins -> cada tool a su version.
if [[ -f "${META_PY}" ]]; then
  expect "pypi meta open_harness" "$(pyproj_version "${META_PY}")" "${REFERENCE}"
  for tool in "${TOOLS[@]}"; do
    expect "pypi meta pin ${tool}" "$(pyproj_pin "${META_PY}" "${tool}")" "${REFERENCE}"
  done
  if [[ -n "${SCOPELENS_VER}" ]]; then
    # scopelens puede o no estar en los pins del meta pypi (packaging opcional).
    sl_pin="$(pyproj_pin "${META_PY}" scopelens)"
    [[ -n "${sl_pin}" ]] && expect "pypi meta pin scopelens" "${sl_pin}" "${SCOPELENS_VER}"
  fi
else
  fail "pypi meta: no existe ${META_PY}"
fi
if [[ -f "${META_PY_INIT}" ]]; then
  expect "pypi meta __version__" "$(py_dunder_version "${META_PY_INIT}")" "${REFERENCE}"
else
  fail "pypi meta: no existe ${META_PY_INIT}"
fi

echo
if [[ "${ok}" -eq 1 ]]; then
  echo "RESULTADO: todas las versiones coinciden (main.go, docs, npm y PyPI)."
  exit 0
else
  echo "RESULTADO: hay versiones divergentes."
  exit 1
fi
