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

## Conversation defaults

- Ask one focused question at a time, then wait for the answer before the next.
  Keep the total reasonable. Inspect available evidence, recommend sensible
  defaults and decide routine reversible details yourself.
- Explain unfamiliar concepts and the practical consequences of a choice before
  asking the user to choose.
- Prefer clickable structured multiple-choice questions for clear alternatives,
  including Yes/No questions. Allow custom answers when supported. Use free text
  when options would misleadingly constrain the answer. If the runtime lacks a
  structured input tool, ask plainly and state a material capability limitation.
- Keep progress updates concise and explain findings, decisions and what the
  next step resolves. Do not end a turn merely to announce intended work.

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
