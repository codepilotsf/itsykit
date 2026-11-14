---
name: photoreal-image
description: Generate or refine believable photographic images with concrete camera, light, material and moment choices. Use for editorial, lifestyle and product imagery; broader concept selection belongs to art-direct.
---

# Photoreal Image

Use the runtime's built-in image-generation tool and its installed imagegen
skill when available. This skill supplies photographic direction and critique;
it does not install tools, configure credentials, choose a paid external provider
or change project permissions. Preserve the selected subject, brand and scope.

## Brief and prompt

Read the brief, intended placement and existing image. Inspect local references
with the image viewer and label their role: identity, geometry, composition or
style. Resolve consequential open direction through the project's consultation
rules; use [art-direct](../art-direct/SKILL.md) when concept selection is needed.
Honor delegated choices without starting another interview.

Write a focused natural-language prompt with the physical facts that matter:
subject and moment, composition/crop, motivated light, camera viewpoint and focus,
and believable materials. Commit to one coherent idea. State reference elements
that must survive, exact requested text and output placement. Avoid conflicting
style adjectives and unsupported claims about model controls or dimensions.

Use scene-appropriate texture and a distinctive detail when helpful. Clean
product references need no invented wear, clutter or narrative props. Preserve
identity and geometry rather than filling a fixed number of prompt blocks.
Save the exact prompt in the project's private working artifacts.

Read [craft and examples](references/craft.md) for a difficult prompt detail;
read [looks](references/looks.md) when lighting, camera or color choices need
examples. Neither is a mandatory read for a brief that already settles the look.
Describe visual attributes in prompts rather than relying on artist names.

## Generate and inspect

1. Generate one candidate on the intended finishing model. Alternatives are for
   requested exploration or a concrete failure that warrants another attempt.
   Respect the user's count/budget and the tool's supported controls. An empty
   CLI key probe does not establish that a native generation tool is absent.
2. Inspect the actual pixels and intended crop using
   [the critique rubric](references/critique.md). Check light, focal plane,
   materials, perspective, anatomy, text and reference fidelity where applicable.
   Judge whether the image communicates the brief, not whether the prompt says so.
3. Deliver an acceptable candidate without a ceremonial correction round.
   Otherwise fix its worst material failure and preserve what works. After at
   most two targeted correction rounds, reconsider the concept or explain the
   limitation rather than generating indefinitely.

If the built-in tool is unavailable or fails, report the limit and retain the
prompt. Use an external API/CLI only when the user explicitly chooses it, following
its current supported workflow. Never silently switch model families or install
dependencies to complete an image request.

## Delivery

Copy the selected asset to the requested project location with its exact prompt
sidecar. Keep existing assets unless replacement is requested. Report the saved
path, actual dimensions and generation route. Distinguish generated imagery from
photographs documenting real people, products or events; generation is not proof
of a real business or customer result.

Adapted from Oleg Spiridonov's skill; see the pinned
[source and adaptation record](references/sources.md) and [license](LICENSE).
