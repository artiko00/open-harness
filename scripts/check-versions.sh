#!/usr/bin/env bash
#
# check-versions.sh — verifica que la version de cada tool este sincronizada
# en todas sus fuentes.
#
# Fuente de verdad: el `const version` de cada tools/<tool>/main.go.
# Se compara contra:
#   - el bloque del tool en open-harness.json (y la version del manifiesto),
#   - cada mencion de version (vX.Y.Z) en README.md y AGENTS.md que aparezca
#     en una linea que nombre al tool.
#
# Imprime cada fuente y si coincide. Sale con 1 si alguna diverge, 0 si todas
# coinciden.

set -u

# Ubicarse en la raiz del repo (el script vive en scripts/).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT}" || exit 2

TOOLS=(linelens dupelens secretlens testlens)
MANIFEST="open-harness.json"
DOCS=(README.md AGENTS.md)

ok=1

pass() { printf '  OK    %s\n' "$1"; }
fail() { printf '  DIFF  %s\n' "$1"; ok=0; }

# Lee el `const version = "X.Y.Z"` de un main.go.
const_version() {
  local file="$1"
  grep -E '^const version = "' "${file}" \
    | sed -E 's/.*"([^"]+)".*/\1/' \
    | head -n1
}

# Lee la version del bloque de un tool en open-harness.json.
# Busca el objeto cuyo "name" coincide y devuelve su "version".
manifest_tool_version() {
  local name="$1"
  awk -v tool="${name}" '
    /"name":[[:space:]]*"/ {
      cur = $0
      gsub(/^.*"name":[[:space:]]*"/, "", cur)
      gsub(/".*$/, "", cur)
      in_tool = (cur == tool)
    }
    in_tool && /"version":[[:space:]]*"/ {
      v = $0
      gsub(/^.*"version":[[:space:]]*"/, "", v)
      gsub(/".*$/, "", v)
      print v
      exit
    }
  ' "${MANIFEST}"
}

# Version de nivel superior del manifiesto (primera clave "version").
manifest_top_version() {
  awk '
    /"version":[[:space:]]*"/ {
      v = $0
      gsub(/^.*"version":[[:space:]]*"/, "", v)
      gsub(/".*$/, "", v)
      print v
      exit
    }
  ' "${MANIFEST}"
}

# Extrae todas las menciones vX.Y.Z de las lineas de un doc que nombren al tool.
# Imprime una version por linea (sin el prefijo "v").
doc_tool_versions() {
  local doc="$1" name="$2"
  grep -E "\b${name}\b" "${doc}" \
    | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' \
    | sed -E 's/^v//' \
    | sort -u
}

echo "Verificando sincronizacion de versiones (fuente: const version de main.go)"
echo

REFERENCE=""

for tool in "${TOOLS[@]}"; do
  main="tools/${tool}/main.go"
  if [[ ! -f "${main}" ]]; then
    fail "${tool}: no existe ${main}"
    continue
  fi

  expected="$(const_version "${main}")"
  if [[ -z "${expected}" ]]; then
    fail "${tool}: no se pudo leer const version en ${main}"
    continue
  fi

  echo "[${tool}] const version (fuente de verdad) = ${expected}"

  # Todas las const deben coincidir entre si (version unificada del repo).
  if [[ -z "${REFERENCE}" ]]; then
    REFERENCE="${expected}"
  elif [[ "${expected}" != "${REFERENCE}" ]]; then
    fail "${tool}: const version ${expected} difiere de la referencia ${REFERENCE}"
  fi

  # open-harness.json (bloque del tool).
  mtv="$(manifest_tool_version "${tool}")"
  if [[ "${mtv}" == "${expected}" ]]; then
    pass "${MANIFEST} (tool ${tool}) = ${mtv}"
  else
    fail "${MANIFEST} (tool ${tool}) = ${mtv:-<vacio>} (esperado ${expected})"
  fi

  # README.md y AGENTS.md.
  for doc in "${DOCS[@]}"; do
    if [[ ! -f "${doc}" ]]; then
      fail "${doc}: no existe"
      continue
    fi
    versions="$(doc_tool_versions "${doc}" "${tool}")"
    if [[ -z "${versions}" ]]; then
      fail "${doc} (${tool}): sin mencion de version vX.Y.Z"
      continue
    fi
    while IFS= read -r v; do
      [[ -z "${v}" ]] && continue
      if [[ "${v}" == "${expected}" ]]; then
        pass "${doc} (${tool}) = ${v}"
      else
        fail "${doc} (${tool}) = ${v} (esperado ${expected})"
      fi
    done <<< "${versions}"
  done

  echo
done

# Version del manifiesto a nivel raiz debe coincidir con la referencia.
mtop="$(manifest_top_version)"
echo "[manifiesto] version raiz = ${mtop} (referencia ${REFERENCE})"
if [[ "${mtop}" == "${REFERENCE}" ]]; then
  pass "${MANIFEST} (version raiz) = ${mtop}"
else
  fail "${MANIFEST} (version raiz) = ${mtop:-<vacio>} (esperado ${REFERENCE})"
fi

echo
if [[ "${ok}" -eq 1 ]]; then
  echo "RESULTADO: todas las versiones coinciden."
  exit 0
else
  echo "RESULTADO: hay versiones divergentes."
  exit 1
fi
