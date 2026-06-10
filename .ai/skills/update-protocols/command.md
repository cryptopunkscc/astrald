---
description: Generate protocol doc packages from mod/ source into .ai/artifacts/protocols/, in the exact structure and style of .ai/system/protocols. drift = report only. Optionally scope to modules.
argument-hint: "[drift|refresh] [module ...]"
---

Run the workflow defined at `.ai/skills/update-protocols/SKILL.md` using the engine at `.ai/skills/update-protocols/workflow.js`.

Parse `$ARGUMENTS`:

* First word `drift` — drift mode. Read-only: report stale and missing protocol docs per module. Write nothing.
* First word `refresh`, or no mode word — refresh mode. A cheap per-module drift gate runs first; only stale or undocumented packages are authored. Complete drop-in packages are staged at `.ai/artifacts/protocols/<protocol>/`. No commits.
* Remaining words are module names that scope the run. No module words — every module with ops.

Invoke with the Workflow tool:

* drift, all modules: `Workflow({ scriptPath: ".ai/skills/update-protocols/workflow.js", args: { mode: "drift" } })`
* refresh, all modules: `Workflow({ scriptPath: ".ai/skills/update-protocols/workflow.js" })`
* one module: `args: "nodes"`
* several modules: `args: ["nodes", "objects"]`
* scoped drift: `args: { mode: "drift", modules: ["nodes", "objects"] }`

The no-args refresh is self-gating and idempotent. A fingerprint pass against `.ai/artifacts/protocols-state.json` skips modules unchanged since they last gated clean or shipped a package; a run where nothing changed costs only the discovery and fingerprint passes.

Scope guards: never write to `.ai/system/` (the astral-docs submodule — staging lives at `.ai/artifacts/protocols/`), never git-commit, never run astrald, astral-query, or any op — all content is synthesized from reading source.

On completion, summarize: in drift mode, the stale packages and their findings; in refresh mode, the packages staged (ops authored, files copied verbatim, orphaned docs excluded), the packages dropped after failed verification, the modules skipped as unchanged, and the findings left for a human. State the move procedure: replace `.ai/system/protocols/<protocol>/` wholesale with the staged directory, review the submodule diff, commit in astral-docs.
