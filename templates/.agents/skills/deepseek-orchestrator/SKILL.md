---
name: deepseek-orchestrator
description: Choose direct Sol implementation or bounded, flat DeepSeek delegation in T3 Code. Use for deciding whether to delegate coding work, coordinating up to three independent developer/QA workstreams, reviewing worker results, or routing child tasks to the configured provider and model. Covers live model discovery, reasoning levels, worktree ownership, risk-based review and authorized delivery.
---

# DeepSeek orchestrator

T3 Code owns orchestration. Prefer its injected tools and their live descriptions over any
schema copied here; those tool descriptions have priority when they differ.

## Roles and models

- **Lead / final scrutiny:** GPT-6.1 Sol. Implements small tasks directly, owns the decision,
  reviews and presents the settled diff, and opens a linked PR/MR only when authorized. The lead
  never edits the same task worktree concurrently with an active worker.
- **Implementation, research, debugging worker:** DeepSeek V4.1 Flash through OpenCode. It
  handles routine fixes and test diagnosis.
- **Independent QA when warranted:** a fresh DeepSeek V4.1 Flash Go worker, read-only and
  sequential within its workstream. Sol retains final scrutiny; a DeepSeek review never
  replaces the lead's.

Discover the actual IDs at runtime with `orchestrator_capabilities` and use the returned
provider instance and model. Never assume every environment matches the defaults below.

- Lead default currently: `codex` / `gpt-6.1-sol`.
- Worker default currently: provider instance `opencode`, model
  `opencode-go/deepseek-v4.1-flash`.

Use the Go subscription route **only**. Never substitute `opencode/deepseek-v4.1-flash` or
another pay-as-you-go route for the worker. If the configured model is unavailable, report
it and stop; do not silently switch models, providers, or routes.

## Choose the smallest useful team

Optimize time to verified completion while avoiding unnecessary token usage and handoffs.
Within the project's delegation authorization:

- Handle small, clear edits and straightforward fixes directly with Sol when briefing and
  reviewing a worker would cost more than doing the work. Required checks and reviews still apply.
- Prefer one DeepSeek developer for substantial, cohesive implementation. Sol retains
  planning, consequential decisions, integration and final scrutiny.
- Actively consider parallel work for larger tasks. Expand to two or at most three
  developer/QA workstreams only when each can make independent progress with a clear scope,
  acceptance criteria and agreed shared interfaces. Keep dependent steps sequential.
- Three workstreams is a ceiling, not a target. Allow at most three active developers and
  three active QA reviewers across the request; other research/debugging workers consume
  developer capacity. Respect any lower runtime limit and queue work rather than bypass it.
- Keep delegation flat: only the main Sol lead creates workers. Do not create Sol subleads
  or let workers delegate further. Include this restriction in every delegated brief.
- A pair describes a developer followed by independent QA when the review tier requires it,
  not two agents launched together. Start QA on a settled candidate with its writer paused;
  other independent workstreams may continue. Do not launch idle reviewers or require QA
  solely because developer capacity is available.

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
  the `task-delivery` skill for scoped staging, verification and integration. Standing
  merge permission lets the lead merge verified work without asking again, subject to
  required reviews, CI and branch protections. Deployment, releases, production mutations
  and destructive Git operations still require separate authorization, including automatic
  CI side effects. When permission is missing, return the reviewable result and name it.
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
  editors. If findings remain unresolved after two fix/review rounds, the lead takes
  ownership of diagnosis and resolution instead of continuing worker handoffs. Pause the
  editor before Sol takes over; surface any decision or blocker that needs user input.
- Use a distinct `clientRequestId` per round, stable across retries of that round, and
  retain each returned `taskId`.

## Review tiers

- **Low risk:** Sol implements and inspects the diff directly, or reviews the delegated
  implementer's diff. No separate QA agent is required by this skill.
- **High risk / explicitly requested independent review:** for changes affecting
  security boundaries, money movement, data integrity, concurrency or irreversible
  operations, run a fresh read-only reviewer after implementation, then lead scrutiny.
  Size, multiple files or several implementation steps alone do not require another
  reviewer. Existing project-required reviews still apply. Reuse check evidence; the
  reviewer reruns checks only for a concrete gap or concern. Never overlap the reviewer
  with an editing worker in the same checkout.

A fresh review round is a **new** `delegate_task`. Include the original brief, the baseline,
prior findings, fixes, and any unresolved objections. Do not continue a review round by
messaging the child thread; track it as its own task.

## Async behavior

- Let async completion notifications wake you; do not poll.
- Call `task_status` only when you need a result mid-turn, and `task_cancel` to stop work
  safely when it is no longer wanted.
- Link every pull request or merge request you create or work on with the T3 PR-linking
  tools. Watch a PR only when explicitly asked to.
