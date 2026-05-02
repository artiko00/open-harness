#!/usr/bin/env bash
set -e

echo "==> Bootstrapping open-harness (Go workspace + TDD checks)..."

echo "  → go work sync"
go work sync

failed_tools=()
no_test_tools=()

for tool_dir in tools/*/; do
  tool=$(basename "$tool_dir")
  echo "  [$tool] build..."
  if ! (cd "$tool_dir" && go build .); then
    failed_tools+=("$tool (build)")
    continue
  fi

  test_files=$(find "$tool_dir" -maxdepth 1 -name "*_test.go" 2>/dev/null | wc -l)
  if [ "$test_files" -eq 0 ]; then
    no_test_tools+=("$tool")
    echo "    [$tool] ⚠ sin archivos *_test.go — TDD no respetado"
    continue
  fi

  echo "  [$tool] test..."
  if ! (cd "$tool_dir" && go test ./...); then
    failed_tools+=("$tool (test)")
  fi
done

echo ""
echo "==> Resumen:"
for tool_dir in tools/*/; do
  tool=$(basename "$tool_dir")
  echo "      - $tool"
done

if [ ${#no_test_tools[@]} -ne 0 ]; then
  echo ""
  echo "  ⚠ Tools sin tests (viola TDD — ver AGENTS.md y ADR-011):"
  for t in "${no_test_tools[@]}"; do echo "    - $t"; done
fi

if [ ${#failed_tools[@]} -ne 0 ]; then
  echo ""
  echo "  ❌ Tools con fallas:"
  for t in "${failed_tools[@]}"; do echo "    - $t"; done
  exit 1
fi

echo ""
echo "==> Environment ready. Workflow TDD activo (ver AGENTS.md sección 2)."
