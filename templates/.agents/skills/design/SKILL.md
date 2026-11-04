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

In T3 Code, verify the current turn's effective model and reasoning from explicit
runtime metadata. When they already match, continue without changing the thread.
`orchestrator_capabilities` discovers supported provider/model IDs and reasoning
options for authorized child tasks; catalog defaults are not evidence of the
current turn's effective settings. Pass the provider, model and reasoning option
explicitly to those child tasks.

Never call `t3_thread_configure` on the thread executing the current turn, including
same-model or reasoning-only updates. Reconfiguring an active thread can interrupt
the calling run and appear as "Run interrupted by user." If the required settings
do not match or cannot be verified, report the limit and leave Impeccable work
pending until the user selects the required settings in T3 and starts a new turn.
Continue independent authorized work where useful. Never rely on upstream helper
metadata or a lower inherited default.

## Visual research

Establish the quality bar from real products before choosing a composition.
Prioritize references the user supplies, then inspect a few strong examples for
the same audience and task. Open the rendered pages or screens: inspect the first
viewport, representative lower content and useful interaction detail. Search
snippets and a list of URLs do not establish visual evidence. If direct inspection
is unavailable, state that limit rather than claiming reference parity.

Use Mobbin when an authorized connection or browser session is available. Check
available access before choosing a visual direction. Make at most two focused
searches for the relevant surface and user goal; inspect useful examples and
show actual screen links with the transferable principles. Never invent
references, claim unseen screens were inspected, or copy distinctive branding.

If Mobbin is unavailable, say so briefly and continue with existing project
patterns and available references. Its absence must not block presenting comps.

Translate references into a short, concrete account of what the incumbent lacks
and what the new design must achieve. Look at product visibility, composition,
type hierarchy, imagery, depth, density, scroll pacing and the primary action.
Borrow the useful design principle, not another product's branding or unsupported
features. Mobbin patterns supplement this judgment; choosing a familiar pattern
does not establish a distinctive or commercially convincing design.

## Product-specific craft

For marketing work, aim for the finish of a strong commercial product in its
category. Make the actual product, its mechanism or the customer's result carry
the design. A credible app demonstration, carefully directed product photograph
or useful interactive example can supply both personality and proof. Rearranging
generic imagery, enlarging a slogan or recoloring the same template is not enough
to resolve a complaint that the design feels generic.

Treat art direction as part of the work. Choose imagery, framing, lighting,
materials, typography and contrast together, with one clear focal moment and
quieter supporting content. Use image generation when authored raster artwork
would improve the result; supply inspected brand and quality references and say
what each contributes. Preserve approved identity and native conventions. Do not
generalize one successful site's palette, device mockup or visual treatment into
a required style for unrelated products or task-focused application screens.

Give each section a job and a suitable composition. Avoid letting habitual
feature-card grids, interchangeable stock scenes or a repeated FAQ layout decide
the page. Demonstrate important behavior and put useful answers near the decision
they support. Keep claims, sample states and availability honest; visual richness
does not justify invented capabilities, testimonials or release details.

## Consultation and comps

Use the project's core conversation defaults, including personal overrides.
Inspect existing context and honor direction the user has already supplied;
do not restart an interview for a routine refinement.

During a design interview, present visible mockups/comps to establish direction
before implementation. Use a rendered HTML prototype or image mockup; prose,
wireframe descriptions and reference links alone do not count. Include useful
desktop/mobile views and alternatives when directions meaningfully differ.

Make comps finished enough to test the proposed quality, with realistic content,
convincing product detail and a clear next action. Carry the intended imagery and
composition through implementation: produce appropriate assets, retain their
provenance and keep core page text and controls semantic and usable. Do not turn
an approved rich comp into a flattened screenshot or a weaker approximation.

Present comps for feedback and agree on direction before implementing it. Revise
the same comps when feedback arrives. Honor explicit instructions to delegate
direction or proceed without pauses. Outside a design interview, use comps when
they resolve a meaningful design decision; routine UI refinements do not require
an extra approval checkpoint. Visual approval does not authorize deployment.

## Judge the result against the bar

Compare the rendered build with the chosen comp and inspected references at
comparable viewport sizes. Check the opening, a representative scroll sequence,
mobile behavior and the main action. Ask whether product clarity, character and
execution reach the intended standard, not only whether spacing and accessibility
checks pass. Use Impeccable's bounded inspection and finish-review workflow rather
than starting another open-ended polishing loop.

When the user rejects a result as generic or materially below their references,
carry that feedback and the reference evidence into the revised direction and
fresh review. An earlier automated pass or reviewer verdict does not settle the
user's objection. Preserve what they explicitly liked and fix the underlying
composition or product presentation before returning to decorative refinements.

## Implementation handoff

Load [Prewind](../prewind/SKILL.md) only for a project using it, and the relevant
application authoring skill for forms, components, paths and runtime behavior.
Design preferences do not replace those implementation contracts. Resume the
active implementation task after consultation unless a needed answer is pending.
