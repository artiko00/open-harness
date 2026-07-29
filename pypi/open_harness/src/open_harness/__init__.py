"""open-harness — meta-package that pulls the linters into your project.

Installing this package brings in:
    - open-harness-linelens
    - open-harness-dupelens
    - open-harness-secretlens
    - open-harness-testlens
    - open-harness-scopelens

After `pip install open-harness`, those CLI commands are available in your
virtualenv, plus the `open-harness` command whose `init` subcommand delegates
to each tool's own init to scaffold the config files (including scopelens.json
when the scopelens command is available).
"""
__version__ = "0.3.2"
