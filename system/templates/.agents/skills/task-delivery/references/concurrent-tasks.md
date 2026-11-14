# Concurrent task coordination

Read when authorized tasks overlap or share runtime resources. For a single
task, the ownership rules in task-delivery are sufficient.

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
