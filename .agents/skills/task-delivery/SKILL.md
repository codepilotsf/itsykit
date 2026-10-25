---
name: task-delivery
description: Finish a repository task as a scoped, verified, reviewable change. Use for implementation handoff, committing completed work, and opening, updating or merging a PR or merge request. Uses owner-authorized routine Git delivery without repeated approval; preserves checks, project constraints and unrelated work.
---

# Task delivery

Read the project's root `AGENTS.md` and applicable guides. Explicit user instructions
take precedence. Determine whether the project authorizes local commits, branch pushes,
PR/MR creation and merging. Existing standing permission covers those steps without
another approval question. The owner-approved CodePilot default is autonomous routine
Git delivery after verification; a consuming project records its authorization in root
instructions. This skill alone grants no permission. Honor explicit stop-at-review requests.

## Start and scope

- Inspect status, the current branch, remotes and `git worktree list`. Identify pre-existing work and the
  task's files. Preserve other people's changes, staged content and untracked files.
- Default each new independent feature or implementation task to its own worktree and
  descriptive task branch. Reuse its existing worktree on continuation. Keep the main
  checkout available to the user; never switch its branch to start another agent's task.
  Create worktrees through the host's supported workspace tools or Git, without launching
  a new top-level conversation merely to obtain an isolated delegated workspace.
- Record the task ID, owner, absolute worktree path, branch, base commit, scope and runtime
  resources in the host's task metadata or a local coordination record outside the repo.
  Report the path so the user can open the correct editor and preview. Give every delegated
  brief that exact path; require the worker to verify cwd and branch before editing.
- Start from an explicit agreed base commit, normally the target branch. A new worktree
  does not include another checkout's uncommitted or ignored files. Do not copy its whole
  tree, secrets or runtime state; identify required task-owned prerequisites and obtain
  authorization for any necessary transfer. Do not bypass branch-in-use guards.
- Existing tasks with uncommitted work may finish in their current checkout; do not move
  them silently to apply this default. An unborn repository needs an authorized initial
  commit before normal branch-based worktrees. Tiny read-only or prose tasks may use an
  existing checkout when no concurrent writer owns it; state the exception.
- Keep the task's acceptance criteria and required checks explicit. Ask only for an
  unresolved product, architecture, security or authorization decision.

## Concurrency and integration

- One editing agent owns each task worktree at a time. Agents collaborating on that task
  hand off sequentially. Review a settled diff with the writer paused; a reviewer sharing
  the worktree remains read-only. Different authorized feature tasks may run concurrently
  in different worktrees. Worktree isolation does not itself authorize delegation.
- Plan overlap before delegating: identify shared APIs, schemas, dependencies and files.
  Assign a coordinator and integration order when tasks depend on or overlap each other.
  Keep dependent work sequential or based on an explicitly recorded prerequisite branch.
  Separate worktrees prevent overwriting working files; they do not prevent merge conflicts
  or incompatible behavior. Surface newly discovered overlap promptly.
- Give each task its own dev-server port, writable test data/database and build output
  paths when applicable. Do not share writable runtime state or restart another task's
  server. If a tool cannot isolate a resource, serialize that operation through the
  coordinator. Worktrees share Git objects, refs and usually configuration; coordinate
  repository-wide Git changes, and never modify another task's branch or worktree.
- Deliver a scoped change per independent task, using a PR/MR when that is the project's
  integration workflow. Before merging, inspect target changes since the recorded base.
  Reconcile overlap in the task worktree and rerun checks invalidated by the combination.
  Integrate overlapping tasks sequentially. With standing merge permission, continue
  without asking again; earlier checks on an outdated diff do not verify the combination.
- Keep worktrees and branches through review. Remove them only after merge or explicit
  abandonment, with the owner's approval and confirmation that no needed dirty, untracked,
  ignored or unpushed work remains and task-owned servers have stopped. Never force-remove
  worktrees, automatically prune other agents' work, or delete another task's branch.

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

## Complete and review

Finish implementation and affected documentation. Run the project's checks appropriate
to behavior and risk, inspect the final diff, and resolve findings before delivery.
For delegated work, follow the project's orchestration skill and authorization.
Reviewers are advisory; their verdict does not authorize Git or external mutations.

Preserve check evidence and state unverified limits. Failed required checks mean the
task is not ready. Report a blocker or, when authorized and useful, open an explicitly
incomplete draft for review; never claim it is completed or ready to merge.

## Commit, review and integrate

If standing project policy already authorizes these steps, perform them without another
permission question. Otherwise finish the concrete diff and checks first, then request
only the missing authorization with the policy's location and reason.

1. Inspect status and the index again. Stage only task-owned paths or hunks; never use
   bulk staging to sweep in unrelated files. Inspect the entire staged diff before a
   scoped conventional commit. If unrelated work is staged, resolve ownership rather
   than committing or unstaging it silently.
2. Push only the task branch to the intended remote. Inspect applicable CI/deployment
   configuration before pushing or merging. If it triggers deployment or another action
   outside standing permission, obtain that authorization first. Do not force-push,
   rewrite another person's history, push release tags or publish releases without authorization.
3. Open or update the task's PR/MR when the project uses that workflow; avoid duplicates.
   Describe the problem, resulting behavior, checks and limits. Use draft status while
   incomplete or when requested. Keep secrets and private state out.
4. Register the full review URL with the app's PR-linking tool when available. Before
   finishing, verify that every review request worked on is linked. Report a linking
   failure accurately. Use app-native watching only if the user asks for monitoring.

5. With standing merge permission, finish required reviews and CI and merge the verified
   task through the project's supported workflow. Do not bypass branch protections,
   unresolved objections or failing checks. For local integration, use a clean target
   checkout and prefer fast-forward integration when possible. Preserve unrelated edits;
   do not reset, stash or switch another task's checkout to force integration.

Stop at review when explicitly requested or when a real blocker prevents integration.
Deployment, production changes, releases and destructive Git operations still require
separate authorization. Routine merge permission does not authorize their CI side effects.

For a requested undo, inspect subsequent changes and prefer a scoped revert commit.
Handle merge reverts with the correct mainline; do not reset or force-push shared history.

## Handoff

Report what changed, checks actually run, branch/commit and review URL when created,
and remaining limits. If pushing or creating a review fails, retain the local commit
and branch, state the failure and next action, and avoid repeated blind retries.
