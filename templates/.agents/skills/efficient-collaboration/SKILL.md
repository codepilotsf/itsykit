---
name: efficient-collaboration
description: Run focused planning interviews and multi-step development collaboration with prepared question rounds, one clickable question at a time, batched documentation, and a clear finish line. Use when gathering requirements or coordinating sustained work; do not add an interview to a straightforward task.
---

# Efficient collaboration

Reach a useful, reviewable outcome quickly. Reduce avoidable questions and tool
round trips without dropping required checks or crossing authorization boundaries.
Explicit user preferences and project requirements take precedence over these defaults.

## Establish the finish line

- Identify the requested deliverable, decisions already settled, acceptance criteria
  and what can be refined later. Inspect existing evidence before asking the user
  to repeat information. Read relevant documents once, then revisit only changed
  sections, conflicts or missing evidence.
- For a workflow, draft its main stages and branches early. Use a simple flowchart
  when helpful. Settle the basic path before reminders, recovery and edge cases.
- Ask only questions whose answers change the next meaningful decision. There is
  no question quota. A clear task needs no interview; a consequential unresolved
  product, creative or integration direction needs a choice before implementation.
  Follow the decision rule in [core](../core/SKILL.md#ask-about-consequential-direction).

## Prepare rounds, deliver one question at a time

- Prepare the next round internally: prioritize choices affecting scope, cost,
  customer experience, architecture or authorization. Include a recommendation
  and identify work independent of the answer.
- Present one question at a time using the runtime's clickable multiple-choice
  input, including for Yes/No choices. Explain unfamiliar terms or meaningful
  tradeoffs briefly before asking. Use free text when choices would distort the
  answer, or a plain question if no structured tool is available.
- After an answer, capture the decision and promptly ask the next relevant
  question. Do not interleave routine document edits, repeated reads, lengthy
  recaps or another broad audit after every answer.
- Adapt the round to corrections and new requirements. Replace obsolete or
  lower-priority questions instead of extending the interview automatically.
  Reopen a settled decision only when new evidence materially changes it.
- Decide routine, reversible details yourself. Do not invent questions to keep
  the consultation active. Separate an actual blocker from a recommendation or
  detail that can be refined during implementation.

## Persist lightly, document in batches

- Keep a lightweight decision log with confirmed choices, recommendations,
  unresolved blockers and the next question. Use existing task state or a small
  durable note when needed for recovery; do not require a new document per answer.
- Reconcile the main document at round boundaries, at a significant scope change,
  before handoff or context loss, or when the user requests a save. Update sooner
  when correctness or authorization requires it. Distinguish user decisions from
  agent recommendations and unresolved implementation details.
- Summarize only the changes and remaining consequential decisions. Avoid
  regenerating the full document or visual for a minor answer. Check affected
  sections for contradictions when batching the edits.

## Keep development moving

- Work in small, coherent deliverables with observable acceptance criteria.
  Inspect the affected surfaces, implement the change, run the relevant checks
  and deliver the reviewable result under the project's worktree/review policy.
- Batch independent reads and searches. Reuse current evidence and passing
  checks; repeat them when changes or unresolved concerns invalidate that evidence.
  Do not trade away required security, review or deployment checks for speed.
- Interrupt only for a genuine missing decision or permission. Continue useful
  independent work while an answer is pending; never treat silence as approval.
  Do not create busywork, expand scope or start external actions to appear active.

## Close deliberately

- At each round boundary, reassess whether enough is known to deliver. Stop
  optional questions once direction is clear. If a substantial new requirement
  needs another round, explain its impact; honor the user's requested stopping point.
- If the user says to wrap up, stop optional questions. Deliver the baseline,
  preserve genuine blockers and move technical refinements into implementation
  notes. Do not silently classify an unanswered decision as user-approved.
- During active work, either continue an actionable step or present the specific
  question/blocker needed to proceed. Do not end merely after saving one answer.
- When the requested outcome is complete, state that clearly, give the artifact
  and next milestone, and stop. Do not keep interviewing or imply work continues
  after the turn. Planning completion does not itself authorize implementation,
  commits, synchronization, deployment or unrelated work.
