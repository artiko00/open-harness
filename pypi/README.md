# pypi/

PyPI distribution layer for open-harness. 5 Python packages publish the same native binaries that the npm wrappers under `npm/@open_harness/` use:

| Package | PyPI name | Version |
|---|---|---|
| `open_harness_linelens/`   | `open-harness-linelens`   | 0.1.3 |
| `open_harness_dupelens/`   | `open-harness-dupelens`   | 0.1.3 |
| `open_harness_secretlens/` | `open-harness-secretlens` | 0.1.2 |
| `open_harness_testlens/`   | `open-harness-testlens`   | 0.1.2 |
| `open_harness/`            | `open-harness` (meta)     | 0.1.0 |

## Distribution model

Same model as `ruff`, `uv`, `esbuild` on PyPI: **native wheels per platform** with the Go binary embedded.

For each tool we publish 4 platform-tagged wheels:

```
open_harness_linelens-0.1.3-py3-none-manylinux2014_x86_64.whl
open_harness_linelens-0.1.3-py3-none-macosx_11_0_arm64.whl
open_harness_linelens-0.1.3-py3-none-macosx_10_9_x86_64.whl
open_harness_linelens-0.1.3-py3-none-win_amd64.whl
```

pip selects the matching wheel automatically based on `sys.platform` and `platform.machine()`. The `__main__.py` entry-point resolves to the bundled binary inside the wheel and `os.execv`'s into it — no Python startup cost per invocation past the entry-point.

The meta-package `open-harness` is **pure Python** (no binary). It only declares dependencies on the 4 per-tool packages, so `pip install open-harness` installs the suite in one shot.

## Building

Requires `go`, `python3`, `python -m build` (`pip install build`).

```bash
bash scripts/build-pypi.sh linelens     # builds 4 wheels into pypi/open_harness_linelens/dist/
bash scripts/build-pypi.sh dupelens
bash scripts/build-pypi.sh secretlens
bash scripts/build-pypi.sh testlens
bash scripts/build-pypi.sh open-harness  # builds the pure-Python meta wheel
```

## Publishing

```bash
pip install twine
twine upload pypi/open_harness_linelens/dist/*       # repeat per tool
twine upload pypi/open_harness/dist/*                # the meta last
```

Order matters: publish the 4 per-tool packages **before** the meta. The meta has hard `==` version pins so it will refuse to resolve if any per-tool package is missing on the index.

## Manual / first time setup

1. Create an account on https://pypi.org and enable 2FA.
2. Create an API token scoped to the 5 package names (or to your whole account).
3. Save the token in `~/.pypirc`:
   ```ini
   [pypi]
     username = __token__
     password = pypi-<your-token>
   ```
4. Reserve the 5 package names by uploading at least one version (the names cannot be reserved without an upload).

## Roadmap (F-012 in `.agent/feature-list.json`)

- [ ] `[tool.<name>]` table in `pyproject.toml` as an alternate config source (mirrors the `package.json` fallback on npm). Requires extending each tool's Go `config_pkg.go`.
- [ ] Smoke test in a fresh venv: `pip install open-harness && linelens version && …`.
- [ ] Add a `Usage in Python projects` section to the root README.
- [ ] ADR-016 documenting the wheel-per-platform decision and why we don't ship a universal `py3-none-any` wheel.
