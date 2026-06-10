---
description: Bring .ai/knowledge notes + index back in sync with the code. drift = report only; refresh = rewrite stale notes. Optionally scope to modules.
argument-hint: "[drift|refresh] [module ...]"
---

Run the workflow defined at `.ai/skills/update-knowledge/SKILL.md` using the engine at `.ai/skills/update-knowledge/workflow.js`.

Parse `$ARGUMENTS`:

* First word `drift` — drift mode. Read-only: report stale notes and dead citations. Write nothing under `.ai/knowledge/`.
* First word `refresh`, or no mode word — refresh mode. A cheap per-module drift gate runs first; only stale or new modules are rewritten. The index is updated. All changes stay in the working tree. No commits.
* First word `pr` — PR mode, for autonomous runs. Aborts on a dirty tree; pulls the default branch; runs refresh; commits verified changes to `<git user>/docs/update-knowledge-<date>`; pushes; opens a pull request. Requires push rights and an authenticated `gh`.
* Remaining words are module names that scope the run. No module words — whole repo.

Invoke with the Workflow tool:

* drift, all modules: `Workflow({ scriptPath: ".ai/skills/update-knowledge/workflow.js", args: { mode: "drift" } })`
* refresh, all modules: `Workflow({ scriptPath: ".ai/skills/update-knowledge/workflow.js" })`
* pr, all modules: `Workflow({ scriptPath: ".ai/skills/update-knowledge/workflow.js", args: { mode: "pr" } })`
* one module: `args: "nodes"`
* several modules: `args: ["nodes", "objects"]`
* scoped drift or pr: `args: { mode: "drift", modules: ["nodes", "objects"] }`

The no-args refresh is self-gating and idempotent. A run where nothing drifted costs only the gate and writes nothing. The form `/loop 144h /update-knowledge` is safe.

Scope guards: never write to `.ai/system/`, never git-commit, update the index at `.ai/knowledge/README.md` only (never the recipe READMEs under `modules/` or `concepts/`).

On completion, summarize: in drift mode, the stale notes and their findings; in refresh mode, the notes and index rows changed, the modules that failed verification, and the concept gaps. Leave all changes for the user to review.
