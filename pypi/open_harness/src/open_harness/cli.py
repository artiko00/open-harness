"""`open-harness init` — delega en el init de cada tool.

Por cada tool crea su archivo de config (<tool>.json) en el directorio actual.
No sobrescribe: si el archivo ya existe lo reporta y sigue. El init de cada
tool sigue siendo el dueño de generar su propio JSON; aquí solo lo orquestamos
invocando el comando instalado por cada wrapper (linelens, dupelens, ...).
"""
import os
import shutil
import subprocess
import sys

TOOLS = [
    ("linelens", "linelens.json"),
    ("dupelens", "dupelens.json"),
    ("secretlens", "secretlens.json"),
    ("testlens", "testlens.json"),
    ("scopelens", "scopelens.json"),
]


def _run_init() -> int:
    cwd = os.getcwd()
    created, existed, failed = [], [], []

    for name, config in TOOLS:
        target = os.path.join(cwd, config)
        if os.path.exists(target):
            existed.append(config)
            continue

        exe = shutil.which(name)
        if exe is None:
            failed.append(f"{config} (comando {name} no encontrado)")
            continue

        res = subprocess.run([exe, "init"], stdout=subprocess.DEVNULL)
        if res.returncode == 0 and os.path.exists(target):
            created.append(config)
        else:
            failed.append(config)

    print("open-harness init:")
    print(f"  creados:    {', '.join(created) if created else '(ninguno)'}")
    print(f"  existentes: {', '.join(existed) if existed else '(ninguno)'}")
    if failed:
        print(f"  fallidos:   {', '.join(failed)}")

    return 1 if failed else 0


def _usage() -> None:
    print("uso: open-harness <comando>")
    print("")
    print("comandos:")
    print("  init    crea los archivos de config de cada tool en el directorio actual")
    print("")
    print("tools: linelens, dupelens, secretlens, testlens, scopelens")


def main() -> int:
    sub = sys.argv[1] if len(sys.argv) > 1 else None
    if sub == "init":
        return _run_init()
    _usage()
    return 0 if sub in (None, "-h", "--help") else 2


if __name__ == "__main__":
    raise SystemExit(main())
