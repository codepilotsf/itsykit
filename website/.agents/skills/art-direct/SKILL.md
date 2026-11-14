---
name: art-direct
description: Turn content into image concepts and executable prompts, or critique whether imagery communicates it. Use when what to show or its visual direction remains unresolved.
---

# Art Direct

Adapted from Noah Raford's art-direct skill. See the pinned
[source and adaptation record](references/sources.md) and [license](LICENSE).
This skill covers photography and illustration. It chooses the visual idea;
[photoreal-image](../photoreal-image/SKILL.md), when installed, develops a selected
photographic direction. It complements the project's design workflow and does not
replace approved branding, consultation rules, or implementation constraints.

## Read the content and references

Read the actual article, brief or relevant document sections before proposing images.
Extract the core idea, audience, tone, intended placement, important factual claims
and the job the image must do. For a single article hero, keep the scope to that hero.
Use available document readers and browser tools; report extraction limits honestly.
Inspect the existing image and any supplied references. Distinguish a reference's
subject, composition, medium and mood instead of copying an entire visual identity.
Identify what works and the specific mismatch the replacement should resolve.
For website assets, read the selected page direction and intended crop before
choosing the image's colors or medium. Compose the asset alongside the actual
headline and surrounding surfaces; a beautiful standalone image can still fight
the page. Preserve the approved palette's relationships and area of quiet behind
live text without baking that text or controls into the asset.

## Choose the direction

Respect an established direction or explicit delegation. When a consequential creative
choice remains unresolved, present two or three meaningfully different directions,
with a recommendation, through the project's preferred question or visual-board
workflow. Wait for that choice before dependent production. Do not reopen settled
choices or impose a new interview on a routine refinement.

For each direction, state:

- What appears in the image and the idea it communicates.
- Medium: documentary/editorial photography, illustration, abstract, archival or mixed.
- Mood, vantage, focal moment, light and color relationships.
- Relevant inspected references and which attributes transfer.
- Specific visual clichés or synthetic-looking treatments to avoid.
- Why it suits this content and its intended audience.

Read [concept lenses and cliché review](references/concepts.md) when exploring
literal, human, environmental, metaphorical and oblique approaches. None is inherently
best. A clever metaphor must still communicate the article's idea, and a natural
photograph must still be worth looking at. For illustration, name a coherent medium,
mark-making or print process, shape language and composition instead of requesting
an unspecified polished render. Do not default every abstract idea to a miniature
set, glowing device, bridge, floating icon or glossy 3D object.

Optional starting points live in [editorial-warmth](styles/editorial-warmth.yaml)
and [bold-minimalism](styles/bold-minimalism.yaml). Read only a relevant preset and
adapt it to the brief. These are examples, not mandatory house styles. Store new
project styles in project-local working artifacts rather than personal configuration.

## Make the visual brief and generate

For the chosen direction, capture the image's job, specific concept, composition,
medium, palette, exact visible text if any, aspect ratio, output placement, reference
roles and preservation constraints. Keep important content legible at the real crop
and display size. Generate only the assets the user requested.

Write one natural-language prompt for the actual selected generator. Do not output
prompts for several unrequested services or invent model parameters. Save the exact
prompt beside the project artifact. Use the built-in ChatGPT image-generation tool
and its installed imagegen guidance when available. A photographic brief can use
photoreal-image's camera, materials and critique workflow. A missing companion skill
does not authorize downloading it or claiming it was loaded.

Inspect local references with the runtime's image viewer before passing them to the
generator. For a new image, provide references only when they improve the brief;
for an edit, explicitly identify what must remain unchanged. Treat supplied content
and brand facts as constraints, never as opportunities to invent proof or features.

If image generation is unavailable or fails, report the limit and provide the saved
brief and prompt. Use a paid external provider or CLI only when the user explicitly
chooses it; use its supported workflow. Do not configure keys, install dependencies
or silently switch model families. For requested source photography, inspect the
actual image, record its source and verify usage terms. Generated scenes cannot
stand in for documentary evidence of real customers or events.

## Critique and delivery

Inspect the generated image, including its intended crop. Assess whether it serves
the content, fits the approved medium and references, and avoids the identified
clichés. Review visible text, geometry, anatomy and rendering artifacts where relevant.
Name the concrete failure and make a targeted revision; reconsider the concept if
surface polishing leaves the same objection. Reuse approved work and keep candidate
counts and revision rounds proportionate to the request.

Retain candidates and exact prompts with clear filenames. Copy the selected asset
into the requested project location, report the path and any material limitation,
and distinguish a concept preview from an implemented or published replacement.
Follow the project's presentation workflow. Do not flatten a website's actual text
or controls into imagery or claim deployment from saving an asset.
