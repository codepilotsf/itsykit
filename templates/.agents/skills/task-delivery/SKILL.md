---
name: task-delivery
description: Deliver a scoped repository change through worktree setup, proportionate verification and authorized Git review/integration. Use for implementation and delivery, not read-only answers.
---

# Task delivery

Use the project's already-read AGENTS.md and relevant guides. Permissions come
from the owner or project; this skill grants none. Standing permission covers
routine Git actions without repeated approval. Honor explicit stop-at-review.

## Scope and ownership

Inspect status and the affected paths. Preserve pre-existing edits, staged
content and untracked files. Routine prose, guidance and mechanical skill updates
normally use the existing checkout with scoped edits and commits. Unrelated dirty
files do not require another worktree when the task's files and staging are clear.
Do not stage someone else's work or manufacture a branch per copied bundle.

Use a separate task branch/worktree for independent implementation, overlapping
writers, required project isolation or an explicit review-worktree request. Reuse
one supplied by the host; never repeat setup on continuation or switch another
task's checkout. Keep the user's main checkout stable.

Use existing host metadata. Add a coordination record only when concurrency or
runtime resources need one; a routine edit needs no new tracking system. Report
the path when it changes or is needed to open the editor/preview. Do not move
uncommitted work or copy another checkout's secrets/runtime state. An unborn
repository needs an authorized initial commit before normal worktrees.

One editing agent owns a worktree at a time. Reviewers are read-only on a settled
diff; pause its writer before review or handoff. Give authorized workers the exact
path, branch, base, scope and required checks; require them to verify their target.
For overlapping tasks or shared runtime resources, read
[concurrent task coordination](references/concurrent-tasks.md).

## Choose proportionate verification

Define checks by changed behavior and failure impact, not file count or a fixed
ritual. Honor required repository checks and CI; these defaults do not weaken them.

- Prose, skill or configuration guidance with no runtime change: review accuracy,
  links, format and consistency. Do not run unrelated application suites.
- Isolated behavior changes: exercise the changed path and meaningful failure cases;
  use focused existing tests and add regression coverage when it proves useful behavior.
- Authentication, permissions, payments, data integrity, concurrency and shared
  infrastructure: include affected integration boundaries and required broader suites.
  Use browser checks for changed critical user journeys, not every internal edit.

For small projects, default to the happy path and consequential failures at the
observable contract, not a test for every private method. Avoid overlapping coverage
across layers unless each protects a distinct risk. Prefer representative boundary
cases to combinatorial permutations; reserve exhaustive coverage for justified
security, permissions, payments, data-integrity risks or complex invariants. For real
bugs, reproduce the failure before the fix and verify it afterward when feasible;
add regression coverage where it meaningfully protects against recurrence.

New tests, harnesses and mocks should resolve concrete uncertainty that affects an
implementation or acceptance decision. Keep fixtures and mocks proportionate; rethink
the design or test approach when scaffolding dwarfs the change. After repeated cycles
produce no useful new evidence, simplify the approach or escalate the unresolved
decision and expose any blocker instead of expanding the machinery. Verify the actual
user flow where relevant; passing mocks or emulators do not establish required
actual-device acceptance. Honor explicit user deadlines and report remaining gaps or
blocked required checks before the deadline; do not silently waive acceptance gates.

Judge existing suites by useful evidence and cost, not test or assertion counts.
A fast existing broad suite can be a sensible final or release check; that does not
justify creating redundant tests or repeatedly rerunning unchanged broad suites.

During iteration, run focused checks. Run the required broader set on the final
candidate and reuse passing results until later changes, target-branch movement or
unresolved concerns invalidate them. Do not write tests that merely mirror the code,
add runtime suites for formatting-only edits, or repeat another agent's checks without
an evidence gap. Never delete tests, weaken assertions, bypass protected components
or silently reduce mandatory coverage for speed. Report unverified boundaries plainly.

## Conditional preparation

For deployment or host resource pressure, read
[deployment preparation](references/deployment-preparation.md). Skip it for
ordinary implementation. A successful build or configured staging destination
does not authorize a rollout.

## Review and integrate

Finish implementation and affected documentation, run required checks and review
the final diff. Resolve material findings. Reuse observed evidence; advisory
reviewers grant no Git or external-action permission. Failed required checks
block readiness. When useful and authorized, an incomplete draft may expose the
blocker, but must not be called ready or complete.

1. Inspect status and index. Stage only task-owned paths or hunks and review the
   entire staged diff before a scoped conventional commit. Do not commit or
   unstage unrelated work to make delivery convenient.
2. Inspect CI/deployment configuration before push or merge. Standing routine
   Git permission excludes deployment, releases, production mutations and
   destructive history operations, including automatic CI side effects. Obtain
   only a genuinely missing authorization after the diff is reviewable.
3. Push and create/update a PR/MR when requested or required by the project's
   integration workflow. A completed local maintenance task need not wait for
   optional publication or administrative cleanup. Avoid duplicates; describe the problem, resulting behavior, checks
   and limits. Register its full URL with the app's linking tool when available
   and confirm all review requests worked on are linked. Report linking failures.
   Use app-native watching only when monitoring was requested.
4. With standing merge permission, finish required reviews and CI, inspect target
   changes since the base and reconcile overlap. Repeat only invalidated checks.
   Integrate overlapping changes sequentially; respect protections and objections.
   Local integration requires a clean target checkout; prefer fast-forward when
   possible. Never reset, stash or switch another task's checkout to force a merge.

Stop at review when requested or blocked. Keep task worktrees through review;
remove only after merge or explicit abandonment with owner approval, no needed
local/unpushed work and task-owned servers stopped. Never force-remove or prune
another task's work. For undo, prefer a scoped revert preserving later changes;
handle merge reverts with the correct mainline.

## Handoff

Report the result, checks observed, branch/commit, review URL and material limits.
On a failed push or review operation, retain local work, state the failure and
next action, and avoid repeated blind retries.
