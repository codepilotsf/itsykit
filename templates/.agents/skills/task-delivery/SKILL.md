---
name: task-delivery
description: Finish a repository task as a scoped, verified, reviewable change. Use for implementation handoff, committing completed work, and opening or updating a draft PR or merge request. Follows the project's standing Git permissions; never grants permission to merge, deploy, publish or include unrelated work.
---

# Task delivery

Read the project's root `AGENTS.md` and applicable guides. Explicit user instructions
take precedence. Determine whether the project authorizes local commits, branch pushes
and draft PR/MR creation. A recommendation in this skill is not authorization.

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
- Deliver one scoped draft PR/MR per independent task. Before an approved merge, inspect
  changes to the target branch since the recorded base. Reconcile overlap in the task's
  worktree and rerun checks invalidated by the combined changes. The coordinator integrates
  overlapping tasks sequentially, with explicit merge approval. Earlier review or passing
  checks alone do not establish that the combined result works.
- Keep worktrees and branches through review. Remove them only after merge or explicit
  abandonment, with the owner's approval and confirmation that no needed dirty, untracked,
  ignored or unpushed work remains and task-owned servers have stopped. Never force-remove
  worktrees, automatically prune other agents' work, or delete another task's branch.

## Complete and review

Finish implementation and affected documentation. Run the project's checks appropriate
to behavior and risk, inspect the final diff, and resolve findings before delivery.
For delegated work, follow the project's orchestration skill and authorization.
Reviewers are advisory; their verdict does not authorize Git or external mutations.

Preserve check evidence and state unverified limits. Failed required checks mean the
task is not ready. Report a blocker or, when authorized and useful, open an explicitly
incomplete draft for review; never claim it is completed or ready to merge.

## Commit and draft review

If standing project policy already authorizes these steps, perform them without another
permission question. Otherwise finish the concrete diff and checks first, then request
only the missing authorization with the policy's location and reason.

1. Inspect status and the index again. Stage only task-owned paths or hunks; never use
   bulk staging to sweep in unrelated files. Inspect the entire staged diff before a
   scoped conventional commit. If unrelated work is staged, resolve ownership rather
   than committing or unstaging it silently.
2. Push only the task branch to the intended remote. Do not force-push, merge, rebase
   another person's history, push tags or publish releases without authorization.
3. Open a draft PR/MR against the intended base branch, or update the existing one for
   this task rather than creating duplicates. Describe the concrete problem, resulting
   behavior, checks and remaining limits. Draft status and rejection leave the target
   branch unchanged, but a push may trigger CI; keep secrets and private state out.
4. Register the full review URL with the app's PR-linking tool when available. Before
   finishing, verify that every review request worked on is linked. Report a linking
   failure accurately. Use app-native watching only if the user asks for monitoring.

Stop at review. Merge, deployment, production changes and release publication require
their own explicit approval. Approval to open a review is not approval for those steps.

## Handoff

Report what changed, checks actually run, branch/commit and review URL when created,
and remaining limits. If pushing or creating a review fails, retain the local commit
and branch, state the failure and next action, and avoid repeated blind retries.
