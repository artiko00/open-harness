# openspec/

Spec-Driven Development (SDD) entry point for `open-harness`. AI agents working in this repo MUST read `config.yaml` before drafting proposals, specs, designs, or task lists. Humans can browse `changes/` to see the current SDD pipeline state.

## Layout

```
openspec/
├── config.yaml     ← project context + rules per artifact (proposal/specs/design/tasks)
├── README.md       ← this file
└── changes/        ← one directory per active change, kebab-case name (e.g. add-bigo-tool)
    └── <change>/
        ├── proposal.md
        ├── specs/
        ├── design.md
        ├── tasks.md
        └── state.yaml
```

## Workflow

1. **Propose**: create `changes/<verb>-<scope>/proposal.md` answering "what problem, what scope, why now".
2. **Spec**: write `specs/` with RFC 2119 requirements and Given/When/Then scenarios (4-hashtag headings).
3. **Design**: write `design.md` covering architecture, testing strategy, ADR impact.
4. **Tasks**: break down into `tasks.md` ordered red → green → refactor → docs.
5. **Apply**: implement tasks, mark each `[done]` only when tests pass and coverage holds at 100%.
6. **Verify**: check completeness, correctness, coherence before archiving.
7. **Archive**: move folder out of `changes/` once merged.

## Tool reference

If the `@open-spec/cli` is installed, `openspec update` re-syncs AI instructions from the latest `config.yaml`. The CLI is Node-based and is optional for this Go repo — `config.yaml` is the source of truth either way.

## Linked feature list

The `.agent/feature-list.json` is the authoritative backlog. Each `changes/<name>/` SHOULD link to a feature ID (e.g. F-011) in its `proposal.md`. The reverse is also true: features that move into active work get a `changes/` folder created.

See [AGENTS.md](../AGENTS.md) section 7 for the end-to-end workflow including SDD + TDD + quality gates.
