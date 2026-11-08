---
name: design
description: Apply CodePilot design defaults to new designs and existing UI refinements. Routes relevant work to global Impeccable while preserving the project's visual system and implementation contracts.
---

# CodePilot design guidance

Use the runtime-discovered global Impeccable skill and launcher, with the project
as cwd. Never install, copy, vendor or link Impeccable into a project. Keep project
context and artifacts there. Preserve the stack, product truth, approved brand,
design system and security rules. Do not add a framework or build step to follow
toolkit examples. If the toolkit is missing or cannot run,
report the limit and continue permitted work from existing context.

## Match the workflow to the request

- **Routine refinement:** inspect the affected UI and nearby patterns, preserve
  its identity and behavior, make the requested change, then verify its rendering
  and changed interactions. Use Impeccable's incumbent-refinement guidance. Do not
  start new-work, an interview, external research, a concept tournament or a
  separate comp approval just because a UI file changes. Run additional audit or
  polish workflows only when requested or needed for a concrete unresolved issue.
- **New surface or substantial redesign:** use the relevant Impeccable workflow.
  Resolve consequential open direction with visible options and honor choices
  already supplied; a fully specified brief needs no confirmation interview.
  Default to building the working UI in code rather than
  reconstructing a generated screenshot.
- **Explicit comp-led work:** when the user requests generated design exploration,
  measured comp fidelity, or the project already pins that path, follow the full
  upstream comp, asset, comparison and review gates. Keep an approved comp's
  authority; never silently downgrade it to bypass a failed gate.

For new work without an established build-path preference, record `buildPath` as
`code-led` in the project's .impeccable/config.json before starting its direction
round. Preserve other settings. Explicit user choices and existing project
preferences win; do not overwrite them or ask an extra execution-mode question.
Use the toolkit's supported board/toggle workflow for direction decisions.

## Present reviews in the owner's browser

Default to the owner's normal external browser for direction boards, concept
comps and final review artifacts. Honor an explicit browser/transport preference;
keep personal machine details in supported owner/environment configuration, not
shared skills or invented Impeccable config keys. T3 preview is for agent
inspection and does not establish access from the owner's device.

Before presenting any review, read [browser review](references/browser-review.md).
Track internal preview and owner review URLs separately. Remote loopback is never
an owner link without an established forwarding route. Reuse an existing reachable
route, provide a portable local-file ZIP for standalone concepts, and keep the
task-owned review available while an answer is pending. Apply the same destination
to inline board buttons, expanded images, mobile views and downloads.

## Model and effort

Use GPT-6.1 Sol for substantial Impeccable work and its authorized text/code
helpers, at high reasoning or higher. Routine refinements use task-appropriate
effort. Image-generation tools use their own models.

Verify effective settings from runtime metadata when available. Higher effort
satisfies the minimum; unavailable metadata is a disclosed limit, not an automatic
stop. Do not claim settings were verified when they were not. Respect an explicit
user-selected model and report any concrete capability gap.

Never reconfigure the thread executing the current turn, even for a same-model
or reasoning-only change; it can interrupt the run. Do not require a new turn
merely to normalize suitable settings. Discover available IDs/options only when
needed for authorized child work and pass those choices explicitly. This skill
grants no delegation, Git or deployment permission.

## Research when direction needs evidence

Use supplied references and existing project evidence first. Seek external
examples when they resolve an actual open composition, UX or quality decision,
not as a prerequisite to every refinement. Reuse current inspected references.

For meaningful external research, inspect rendered pages or screens at the first
viewport, representative lower content and relevant interaction. Search snippets
alone are not visual evidence. When Mobbin fits and authorized access is available,
make at most two focused searches, inspect useful examples and return actual
screen links with the transferable principles. If access is absent, continue
with project patterns and available references; it must not block the work.

Translate evidence into a short account of what the design must achieve. Borrow
principles rather than distinctive branding. Report inspection limits honestly.

## Craft and consultation

Make the product, its mechanism or the customer's result carry marketing work.
Use convincing product detail, useful demonstrations or appropriate real imagery;
a larger slogan and a generic feature grid do not resolve a generic composition.
Choose imagery, framing, type and contrast together, with a clear focal moment
and quieter supporting content. Use generated raster assets when they materially
improve the result; preserve identity, factual claims and asset provenance.

Give each section a purpose and composition appropriate to its content. Preserve
native conventions on task-focused screens. Do not turn one site's palette,
visual device or successful layout into a universal house style, or invent
capabilities, testimonials and commercial claims for visual richness.

During a design interview, present rendered HTML or image comps with realistic
content, desktop/mobile views where useful and alternatives when directions
meaningfully differ. Agree on consequential direction before implementation;
revise the same artifacts after feedback. Honor delegated direction and explicit
instructions to proceed without pauses. Routine refinements need no extra approval.
External presentation preserves the consultation, question key and re-roll/update
flow. A standalone export asks for feedback in the conversation and records the
actual answer; opening or downloading never counts as approval. Unanswered required
choices remain pending, including when a toolkit timeout suggests proceeding.

## Verify and hand off

Compare the rendered build with approved references and any chosen comp at
comparable sizes. Check the opening, representative lower content, mobile layout
and the changed main action. Batch fixes and follow the upstream bounded review
workflow for the selected path; do not add an open-ended polishing loop.

User evidence of a material mismatch outranks an earlier automated pass. Preserve
what they liked and fix the underlying composition or product presentation.

Load [Prewind](../prewind/SKILL.md) only when the project uses it, and the relevant
authoring skill for forms, components, paths and runtime behavior. Resume the
active task after consultation unless a required answer is pending. Visual
approval does not authorize deployment. Deliver review links and offline artifacts
through the same [browser-review workflow](references/browser-review.md), with
client reachability limits and task-owned server lifecycle recorded.
