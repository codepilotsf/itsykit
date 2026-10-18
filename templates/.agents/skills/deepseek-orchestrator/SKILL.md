---
name: deepseek-orchestrator
description: T3-owned multi-agent orchestration for this project. Use when a task should be split between a lead reviewer model and a DeepSeek implementation worker, when delegating or reviewing subagent work, or when routing a child task to a specific provider and model. Covers discovering live provider/model IDs with orchestrator_capabilities, delegate_task briefs and handoffs, risk-tiered review rounds, clientRequestId and taskId tracking, async completion versus task_status and cancellation, git and external-mutation restrictions, and pull request or merge request linking.
---

# DeepSeek orchestrator

T3 Code owns orchestration. Prefer its injected tools and their live descriptions over any
schema copied here; those tool descriptions have priority when they differ.

## Roles and models

- **Lead / final scrutiny:** GPT-6.1 Sol. Owns the decision, reviews the settled diff, and
  presents the reviewable diff and opens a linked PR/MR only when authorized. The lead
  never edits the same task worktree concurrently with an active worker.
- **Implementation, research, debugging worker:** DeepSeek V4.1 Flash through OpenCode. It
  handles routine fixes and test diagnosis.
- **Substantial-tier reviewer:** a fresh DeepSeek V4.1 Flash Go worker, read-only and
  sequential. Sol retains final scrutiny; a DeepSeek review never replaces the lead's.

Discover the actual IDs at runtime with `orchestrator_capabilities` and use the returned
provider instance and model. Never assume every environment matches the defaults below.

- Lead default currently: `codex` / `gpt-6.1-sol`.
- Worker default currently: provider instance `opencode`, model
  `opencode-go/deepseek-v4.1-flash`.

Use the Go subscription route **only**. Never substitute `opencode/deepseek-v4.1-flash` or
another pay-as-you-go route for the worker. If the configured model is unavailable, report
it and stop; do not silently switch models, providers, or routes.

## Reasoning levels

Choose effort from the task, not from a low runtime default. Sol uses at least
`medium` for coding, investigation, planning and final diff review, and `high` for
substantial, ambiguous, cross-cutting, architectural or high-risk work. Reserve
`low` for narrow, deterministic routine tasks with clear acceptance criteria and no
meaningful design or debugging decisions. Raise effort when scope or uncertainty grows.

The DeepSeek Go Flash route offers `low`, `high` and `max`, with no `medium`.
Use `low` only for routine implementation tasks; use `high` whenever medium or higher
would be justified on Sol. Use `high` for independent QA and substantial review.
Reserve `max` for a high-effort pass that still cannot resolve material uncertainty.
Never run independent QA at low effort.

Discover the available options with `orchestrator_capabilities` and pass the chosen
option explicitly to every `delegate_task`. Set and verify the lead's effective option
in the runtime too. If the required level cannot be selected or verified, disclose that
limit instead of implying it was used. Impeccable always follows the separate Sol 6.1 /
high requirement in the design skill.

Use `delegate_task` for T3-owned child work and cross-provider work. Do not launch top-level
threads (`t3_thread_launch`, `create_threads`) merely to run a subagent. Do not use a macOS
path or a legacy shell runner to drive delegation.

## Ownership and limits

- One editing worker per task worktree at a time; the lead does not edit that worktree
  concurrently with its worker. Reviewers are read-only and never overlap an editor in
  the same worktree. Independently authorized feature tasks may run concurrently in
  separate worktrees with coordinated shared resources and integration order.
- Follow `task-delivery` for worktree ownership. The lead prepares or identifies the
  task worktree and supplies its absolute path, branch and base commit in every brief.
  Workers verify cwd and branch before editing; if the runtime cannot target that
  worktree, report the limitation rather than editing another task's checkout.
- The worker may run read-only Git inspection (`status`, `diff`, `log`, `show`) but never
  stages, commits, or otherwise mutates Git or external/production systems. It returns an
  uncommitted candidate in its assigned task worktree.
- Follow the consuming project's root `AGENTS.md` for delegation authorization.
  Loading this skill or classifying work as substantial does not grant that permission.
- The lead may commit, push a scoped task branch and open a linked pull request or GitLab
  merge request when the project's standing policy or the user authorizes it. Follow
  the `task-delivery` skill for scoped staging and review handoff. Never merge, deploy, publish artifacts,
  or mutate production without explicit user approval. Otherwise return an uncommitted
  reviewable diff.
- Keep ownership, scope, and external-mutation restrictions explicit in every brief.

## Briefs and handoffs

- Write a compact, scoped task brief: goal, assigned worktree/branch/base, exact
  files/surfaces, acceptance criteria, boundaries, runtime resources, dependencies
  and required checks.
- Keep handoffs concise (about 300 words or fewer): what changed, exact paths, checks run
  and their results, unresolved risks, and anything the lead must decide.
- Write logs and evidence outside the repository (for example under `/tmp`) and return
  their paths.
- Build the canonical required-check set once, when the final diff is settled; during
  iteration run targeted checks only and repeat the full set only when a change invalidates
  it.
- Batch revision fixes into rounds and incorporate updates safely without overlapping
  editors. If a review round still has unresolved findings after two rounds, escalate to the
  lead instead of looping.
- Use a distinct `clientRequestId` per round, stable across retries of that round, and
  retain each returned `taskId`.

## Review tiers

- **Low risk:** implementer, then direct lead review of the diff.
- **Substantial / security / data / architecture / production behavior / concurrency /
  irreversible operations — or an explicitly requested independent review:** after
  implementation, run a sequential fresh read-only reviewer and a bounded revision round,
  then lead scrutiny. Never run the reviewer concurrently with the editing worker in the
  same checkout.

A fresh review round is a **new** `delegate_task`. Include the original brief, the baseline,
prior findings, fixes, and any unresolved objections. Do not continue a review round by
messaging the child thread; track it as its own task.

## Async behavior

- Let async completion notifications wake you; do not poll.
- Call `task_status` only when you need a result mid-turn, and `task_cancel` to stop work
  safely when it is no longer wanted.
- Link every pull request or merge request you create or work on with the T3 PR-linking
  tools. Watch a PR only when explicitly asked to.
