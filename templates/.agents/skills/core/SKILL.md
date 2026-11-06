---
name: core
description: Shared task continuity, decision, evidence and handoff practices. Read when project AGENTS.md requires it; project rules and personal preferences retain their authority.
---

# Core working practices

Read once at task start when directed by the project's AGENTS.md. Load only
skills and references relevant to the current task. Reuse context already read;
revisit changed sections, conflicts or missing evidence rather than whole guides.

## Scope and continuity

Keep the active outcome and unresolved decisions in view. Follow-up corrections
and small edits steer that outcome unless the user cancels or replaces it.
Resume remaining work after an intermediate request; a plan, progress update,
saved draft or passing check alone does not finish the task.

Continue actionable work until the outcome is complete or a concrete decision
or blocker needs input. Do independent work while awaiting an answer. Honor
pauses, cancellations and scope limits; persistence grants no extra permission.

Preserve personal instructions and configuration. Do not rewrite global AGENTS.md,
settings or personal skills as a side effect of project work. Shared skills must
work without a sibling checkout or a particular developer's home directory.
Project requirements, security and permissions remain authoritative under the
runtime's instruction hierarchy.

## Decisions and conversation

Inspect the brief and existing patterns first. Ask one focused question before
an unresolved choice materially changes product experience, creative direction,
architecture, integration flow, recurring cost or data/security ownership.
Present concrete options, their practical tradeoff and a recommendation; wait
before dependent implementation. Silence is not approval.

Decide routine, reversible details within approved direction. Do not reopen
settled choices without material new evidence. When the user delegates a choice,
decide and state the assumption. Git permission does not delegate product direction.

- For sustained interviews, read [efficient-collaboration](../efficient-collaboration/SKILL.md).
  A clear implementation request needs no interview.
- Prefer one clickable multiple-choice question at a time, including Yes/No,
  unless personal preferences or the user say otherwise. Use free text when
  options would distort the decision; ask plainly if structured input is absent.
- Keep updates concise. Explain findings, decisions and what the next step resolves.
- Never use em dashes (U+2014) in authored prose or responses; use ` – ` instead.

## Browser cleanup

Reuse task-owned tabs. Close temporary inspection tabs through the normal tab
lifecycle after checks; stop task-owned recordings when finished. Preserve user
and other-task tabs, requested review previews and boards awaiting an answer.
Opening a preview without showing it still creates a page. Never terminate a
shared browser or close all tabs for routine cleanup. Report a cleanup failure.

## Evidence and delivery

Inspect Git status and preserve unrelated or concurrent work. Load
[task-delivery](../task-delivery/SKILL.md) for implementation, checks and Git
mechanics; permissions come from the project or user. Report only observed checks
and actions, preserve test isolation and complete reviewable work before asking
for any missing authorization.

Use structured input when a required answer is pending and leave the task
awaiting input, not Done. If the runtime cannot expose that state, say so. A
consultation remains open while required decisions or actionable work remain.
At completion, report the result, verification and material limits, then stop.

## Shared-skill updates

In the canonical repository identified by sync/consumers.json, finish deliberate
skills/ changes and checks before consumer updates. Continue already-authorized
sync scope; otherwise offer the local workflow for the configured Flea CMS and
Preflight templates. Editing a shared skill does not authorize consumer changes.
Documentation-only edits, consumer-local changes and global Impeccable do not
trigger this offer. Generated sites/apps need separate enrollment and scope.

When sync is authorized, follow the repository's synchronization guide, use the
selected exact source commit and preserve drift checks. Standing commit permission
can supply a missing source commit; otherwise request only that authorization.
