# .claude/

Configuración de Claude Code compartida por el repo.

- `settings.json` — versionado. Permisos para los comandos read-only y de build
  del proyecto, para que no pidan confirmación en cada sesión. Los comandos que
  publican o reescriben historia (`npm publish`, `twine upload`, `git push`,
  `git commit --no-verify`) quedan **fuera a propósito**: se confirman cada vez.
- `settings.local.json` — preferencias personales de cada quien; está en
  `.gitignore` y no se versiona.

El contexto del proyecto para agentes vive en [`../CLAUDE.md`](../CLAUDE.md), que
importa [`../AGENTS.md`](../AGENTS.md) como fuente de verdad del workflow.
