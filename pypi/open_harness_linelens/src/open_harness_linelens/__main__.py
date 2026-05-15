"""Entry-point that delegates argv to the bundled native binary."""
import os
import sys
from pathlib import Path

_TOOL = "linelens"


def _binary_path() -> Path:
    """Locate the native binary shipped inside the wheel."""
    name = f"{_TOOL}.exe" if sys.platform == "win32" else _TOOL
    return Path(__file__).resolve().parent / "bin" / name


def main() -> int:
    binary = _binary_path()
    if not binary.exists():
        sys.stderr.write(
            f"{_TOOL}: native binary not found at {binary}. "
            f"The installed wheel is incomplete or built for a different "
            f"platform. Reinstall: pip install --force-reinstall "
            f"open-harness-{_TOOL}\n"
        )
        return 1
    os.execv(str(binary), [str(binary), *sys.argv[1:]])


if __name__ == "__main__":
    raise SystemExit(main())
