---
name: design
description: Apply CodePilot design preferences when designing, redesigning, critiquing or visually refining a website or application. Coordinates visual research and visible comps with the installed Impeccable toolkit and the project's design system.
---

# CodePilot design guidance

Use the globally installed Impeccable skill discovered by the agent runtime
(currently `~/.agents/skills/impeccable/SKILL.md`) for the relevant design workflow,
alongside these preferences. Never install, vendor, copy or link Impeccable into
a project or starter template. Run its global launcher with the project as cwd.
Preserve the consuming project's stack, design system, product context, pinned
brand assets and security rules. Do not add another framework or build step to
follow a toolkit example. These instructions grant no Git, deployment, delegation
or external-action permission.

If Impeccable is missing or its launcher cannot run, state the limitation and
continue permitted work from existing project context and these preferences.
Do not claim toolkit execution or create a project-local fallback. Report a missing
global installation so it can be resolved at user scope.

## Model selection

Run Impeccable work with GPT-6.1 Sol (`gpt-6.1-sol`) at high reasoning. This includes
design planning, UI implementation and delegated Impeccable text/code helpers, such as
the asset producer, manual edit applier, finish reviewer and documenter. Image-generation
tools use their own image models.

In T3 Code, discover supported models and reasoning options with
`orchestrator_capabilities`. Select Sol with `reasoningEffort: high` through
`t3_thread_configure` before starting a design turn. Pass the provider, model and
reasoning option explicitly to authorized child tasks. Verify the effective settings
before proceeding. If the runtime cannot select or verify them, report the limit and
leave Impeccable work pending. Never rely on upstream helper metadata or a lower
inherited default.

## Visual research

Use Mobbin when an authorized connection or browser session is available. Check
available access before choosing a visual direction. Make at most two focused
searches for the relevant surface and user goal; inspect useful examples and
show actual screen links with the transferable principles. Never invent
references, claim unseen screens were inspected, or copy distinctive branding.

If Mobbin is unavailable, say so briefly and continue with existing project
patterns and available references. Its absence must not block presenting comps.

## Consultation and comps

Use the project's core conversation defaults, including personal overrides.
Inspect existing context and honor direction the user has already supplied;
do not restart an interview for a routine refinement.

During a design interview, present visible mockups/comps to establish direction
before implementation. Use a rendered HTML prototype or image mockup; prose,
wireframe descriptions and reference links alone do not count. Include useful
desktop/mobile views and alternatives when directions meaningfully differ.

Present comps for feedback and agree on direction before implementing it. Revise
the same comps when feedback arrives. Honor explicit instructions to delegate
direction or proceed without pauses. Outside a design interview, use comps when
they resolve a meaningful design decision; routine UI refinements do not require
an extra approval checkpoint. Visual approval does not authorize deployment.

## Implementation handoff

Load [Prewind](../prewind/SKILL.md) only for a project using it, and the relevant
application authoring skill for forms, components, paths and runtime behavior.
Design preferences do not replace those implementation contracts. Resume the
active implementation task after consultation unless a needed answer is pending.
