---
name: core
description: Shared task continuity, consultation, evidence and handoff practices for CodePilot projects. Read at task start when required by project AGENTS.md; complements personal preferences without granting project or external-action permissions.
---

# Core working practices

Read this file when the project's `AGENTS.md` directs it, even if automatic skill
selection did not choose it. Then load only task-relevant skills and guides.

## Compose with personal and project instructions

Keep each developer's personal instructions intact. Do not install or rewrite
their global `AGENTS.md`, tool configuration or personal skills as a side effect
of working in a project. Project-local skill copies must work without a sibling
checkout or a particular developer's home directory.

The conversation preferences below are shared defaults. Explicit user requests
and available personal conversation preferences override these defaults. That
exception does not waive project security, product requirements or authorization
boundaries. Follow the runtime's instruction hierarchy; report a consequential
conflict rather than pretending this skill changes that hierarchy.

## Carry the task through

Keep the user's active goal and unresolved decisions in view. Treat follow-up
corrections, preferences and small edits as steering unless the user explicitly
cancels or replaces the task. After an intermediate request, resume the remaining
work in the same turn. A plan, progress report or completed subtask is not the
overall outcome.

Until the task is complete, continue actionable work or ask the concrete question
needed to proceed. Do not invent questions when the next step is clear. Continue
independent work while awaiting an answer without assuming a dependent decision.
For an external blocker, state what prevents progress and what specific input or
action would resume it.

During an ongoing consultation, draft files and passing checks do not by
themselves conclude the discussion. Present the concrete result and resolve
outstanding decisions. Honor cancellation, an explicit pause and scope limits;
persistence does not authorize additional work or external actions.

## Ask about consequential direction

Before implementing an unresolved choice that materially changes user experience,
creative direction, architecture, integration flow, recurring cost or data/security
ownership, inspect the brief and existing patterns, then ask one focused question.
Present a few concrete options, explain the user-visible tradeoff and recommend one.
For example, a request to connect payments does not decide whether customers use
hosted pages, provider-rendered embedded forms or an application-owned flow backed
by APIs. Confirm that experience before selecting the implementation.

Do not infer a product preference from the easiest SDK or a vendor default. Wait
for the answer before dependent implementation; silence is not approval. Continue
independent work where useful. A small comparison or reversible exploration may
inform the choice, but must not silently commit the product to an unchosen direction.

Proceed without questions for established patterns, routine internals and small
refinements within an approved direction. Do not reopen settled choices without
new evidence. Permission to commit and merge does not delegate product or creative
direction. If the user explicitly delegates a choice, decide and state the assumption.

## Shared-skill updates

When working in the canonical shared-skills repository identified by its
`sync/consumers.json`, if the task deliberately changes an authored skill under
`skills/`, finish the change and its checks. If the user already requested sync,
continue that scope without asking again. Otherwise offer the local sync workflow
for the configured Flea CMS and Preflight template repositories. This is opt-in;
a shared-skill edit alone does not authorize consumer modifications. When authorized,
follow that repository's local skill synchronization instructions, use the
user-selected pinned source revision and preserve drift checks. If the change
is not yet present in a commit, use standing commit permission when available;
otherwise explain why the pinned workflow needs a source commit and request only
that missing permission. Never treat missing standing permission as already granted.

This offer applies to authored shared skills, not documentation-only changes,
consumer-local edits or Impeccable. The configured consumers are the Flea and
Preflight templates; generated sites and applications are not enrolled unless
they are separately added to the consumer configuration.

## Conversation defaults

- For sustained requirements interviews or multi-step collaboration, read
  [efficient-collaboration](../efficient-collaboration/SKILL.md). Prepare small
  question rounds internally, present one question at a time, and reconcile the
  main document at round boundaries. Inspect evidence first, recommend sensible
  defaults and decide routine reversible details yourself.
- During interviews, move directly from an answer to the next prepared question.
  Skip repeated confirmations and duplicate question text; put brief context in
  the clickable question and batch routine tool work and document updates.
- Explain unfamiliar concepts and the practical consequences of a choice before
  asking the user to choose.
- Prefer clickable structured multiple-choice questions for clear alternatives,
  including Yes/No questions. Allow custom answers when supported. Use free text
  when options would misleadingly constrain the answer. If the runtime lacks a
  structured input tool, ask plainly and state a material capability limitation.
- Keep progress updates concise and explain findings, decisions and what the
  next step resolves. Do not end a turn merely to announce intended work.
- Never use em dashes (U+2014) in authored prose or responses; use an en dash
  surrounded by spaces instead: ` – `.

## Evidence and handoff

Inspect Git status and preserve unrelated/concurrent changes. Extend existing
patterns when appropriate. Load [task-delivery](../task-delivery/SKILL.md) for
worktree ownership, verification and delivery mechanics when applicable; the
project supplies permissions, not this skill.

Report only checks and actions actually observed. Preserve security boundaries
and test isolation; do not bypass runtime components to make checks pass.
Complete the authorized, reviewable work before seeking a missing permission.

When a user answer is pending, use the runtime's structured input mechanism and
keep the thread awaiting input rather than Done/settled. Ending a turn is not
task completion. If the runtime cannot expose the Input state, state the limit
instead of claiming the GUI changed. Do not settle or mark work complete while
required decisions or actionable work remain. At actual completion, report the
result, verification and remaining limits.
