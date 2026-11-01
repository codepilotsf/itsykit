---
name: photoreal-image
description: Direct and generate images that look like believable photographs, using concrete camera, lighting, material and moment choices followed by visual critique. Use for editorial, documentary, lifestyle and product photography or requests to make an image look less synthetic. Illustration and broader concept selection belong to art-direct.
---

# Photoreal Image

Adapted from Oleg Spiridonov's photoreal-image skill. Read
[the source record](references/sources.md) for the pinned revision and adaptations.

Use the runtime's built-in image-generation tool, following its installed imagegen
skill when available. This skill supplies photographic direction and critique.
It does not install tools, select an external paid provider, or change project
permissions. Preserve the user's selected subject, model, brand and output scope.
For unresolved consequential creative choices, follow the project's consultation
rules; when the user delegates direction, choose and state the assumption.

## Overview

A short request ("AC repair scene, symmetrical, from inside the unit") produces a generic image. The same request expanded into the physical facts a camera would have recorded produces a photograph. Expanded into the *decisions a photographer would have made* – it produces a good one.

**Two axes, not one.**

- **Truth** – would a camera have recorded this? Served by blocks 3, 4, 7, 8.
- **Merit** – is this frame worth looking at? Served by blocks 2, 5, 9, 10.

A prompt that only serves truth produces a convincing stock photo: real, and boring. Realism comes from constraint, not adjectives – "photorealistic, 8k, ultra detailed" is noise, while a named lens, a named source, a named flaw and a named material are signal. But merit comes from **commitment**: one moment, one light idea, one detail that could not be anywhere else.

**Never ship a prompt without blocks 2 (Moment), 8 (Imperfections) and 9 (Punctum).** Imperfection is what stops it reading as CGI. Moment and punctum are what stop it reading as stock.

**Commit to one idea.** Dramatic *and* soft, wide *and* intimate, moody *and* bright – the model averages conflicting instructions into mush. Pick the idea, then make every block serve it.

## Workflow

1. Read the brief, intended placement and any existing image. Inspect local references
   with the available image viewer and label each reference's role. When a concept
   is unresolved, use [art-direct](../art-direct/SKILL.md) if installed; otherwise
   resolve it from the brief without inventing an installed dependency.
2. Choose a coherent photographic look. Read [looks](references/looks.md) for lighting,
   camera and color relationships; use named artists as research context and describe
   the actual visual attributes in the generation prompt.
3. Write a focused prompt, usually 250 to 400 words, using the contract below.
   Save its exact wording in the project's private working artifacts.
   Use only imperfections and incidental details that plausibly belong in this scene.
   Preserve factual content, identity and product geometry in reference-based edits.
4. Generate two draft candidates on the intended finishing model when the user's
   scope and budget allow. Use one built-in image call per candidate when required
   by that tool. Respect any requested limit and use available controls only.
   Do not infer tool availability from an API-key probe or promise unsupported sizes.
5. Open and inspect every candidate with [the critique rubric](references/critique.md).
   Judge the actual pixels: light consistency, focal plane, anatomy, text, materials,
   reflections, perspective and the frame's specific moment. A prompt's claims are
   not evidence that the output achieved them.
6. Select the strongest candidate and fix its single worst failure. Preserve what
   works. After at most two targeted correction rounds, reconsider the concept or
   explain the remaining limitation instead of generating indefinitely.
7. Copy the selected image into the requested project location with its exact prompt
   sidecar. Keep existing assets unless replacement is requested. Report the saved
   path, actual dimensions and generation route; distinguish generated illustration
   from photographs documenting real people, products or events.

If the built-in tool is unavailable or fails, report the limitation and retain the
prompt. Use an external API or CLI only when the user explicitly chooses that route;
follow its installed workflow and verify current controls and costs then. Do not
silently switch model families, configure credentials or install dependencies.

## The Prompt Contract

One paragraph per block, in this order. All 10 are REQUIRED.

| # | Block | Must contain |
|---|-------|--------------|
| 1 | **Frame** | Medium + subject + action + where the camera is standing. Then one clause on what the eye does: where it enters the frame and what it lands on. |
| 2 | **Moment** | The *instant*, not the activity. Name what just happened and what is about to. Prefer the beat before or after peak action – that is where photographs live. "He has just stopped turning the wrench and tilted his head toward the compressor; his hand is still on the tool." |
| 3 | **Light** | Hour, season and weather as one physical fact. Then every source and its quality (key, fill, practicals, bounce), the **key-to-fill ratio in stops**, and the *light idea* in one clause – what light is doing in this frame. Mixed color temperatures read as real. |
| 4 | **Camera** | Height, distance, angle. Body class + focal length + aperture. **Rendering character** – diffusion, flare, bokeh shape, grain, vignetting. Exactly which plane is sharp and how fast it falls off. Shutter if anything moves. |
| 5 | **Composition** | Layering across foreground / midground / background. How the subject separates from the ground – by tone, by color, or by focus; name which. Subject scale ("the figure fills about a third of the frame height"), negative space, lookroom. **Edge discipline: nothing tangent to the frame edge or to a background line.** |
| 6 | **Environment** | Specific objects that belong in that real location, distributed across the three planes. Clutter and utility, never staging. Include at least one thing that does not serve the picture – real photographs contain irrelevant objects. |
| 7 | **Subject truth** | Exact garments, fit, condition, drape. Then the human physics: an asymmetrical micro-expression, where the gaze is aimed *and at what*, what the hands are doing, where the weight sits. If no person is in frame, replace with **Materials & surfaces** – substrate, finish, age, and how light behaves on each. |
| 8 | **Imperfections** | **5+ concrete flaws**, drawn from at least three of: skin, fabric, air, surfaces, optics, environment. Never skip this block. |
| 9 | **Punctum** | Exactly **one** small, specific, faintly irrational detail that makes this frame this frame and no other – a child's drawing taped inside the equipment door, one mismatched sock, a wedding ring over a work glove. Not a flaw; a fact with a story behind it. |
| 10 | **Palette, grade & reference** | A color *relationship* – complementary, analogous, near-monochrome with one saturated break, or split-tone – expressed in 4–6 real tones. A named film stock or grade. What tradition the frame belongs to. Then the negative list, last: no CGI look, no plastic or waxy skin, no beauty retouching, no over-sharpening, no HDR halos, no artificial symmetry beyond the intended composition, no malformed hands, no watermark or text unless requested. |

Vocabulary banks and a full worked example: `references/craft.md`.
Presets, photographer/DP references, film stocks, clichés to avoid: `references/looks.md`.
The scoring rubric for step 5: `references/critique.md`.

## Quick Reference

**Lens by intent**

| Shot | Lens | Look |
|---|---|---|
| Interior/POV, environment around subject | 24mm | Wide, immersive, mild edge stretch |
| Documentary / reportage | 35mm | Close to human vision |
| Standard portrait, natural | 50mm | No distortion |
| Flattering portrait, compressed | 85mm f/1.4 | Creamy falloff, isolated subject |
| Product / detail | 100mm macro | Flat field, texture detail |
| Distant subject, compressed layers | 200mm | Stacked planes, thin focus |

**Key-to-fill ratio** – the fastest way to set drama. `1:1` flat, overcast, honest · `2:1` gentle modeling, editorial · `4:1` dramatic, shadow side still reads · `8:1+` low-key, shadows fall to black.

**Depth of field:** `f/1.4` face isolated, background dissolved · `f/2.8` subject sharp, environment legible · `f/5.6` group in focus · `f/11` near-to-far sharp · macro `f/8` with a 2cm plane.

**Rendering character** – what makes digital look photographed: Black Pro-Mist ¼ (halation, lifted blacks) · anamorphic (oval bokeh, horizontal blue flare) · vintage uncoated glass (veiling flare, low contrast) · 7-blade aperture (heptagonal specular highlights, not perfect circles) · cat's-eye bokeh at the edges · 1/48s at 24fps for correct motion blur.

## Common corrections

- Replace quality adjectives with concrete light, focus, materials and action.
- Describe a particular instant rather than an averaged activity or posed smile.
- If a scene feels staged, revise the action, vantage or context before adding grain.
- Texture cannot rescue an implausible concept. A miniature metaphor still reads
  as a miniature when photographed convincingly.
- Correct the actual failure instead of assuming another provider is better.
- Use the resolution appropriate to placement and the tool's documented limits;
  inspect delivered dimensions instead of assuming the requested size was honored.
