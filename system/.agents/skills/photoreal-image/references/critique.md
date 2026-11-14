# Critique – scoring drafts and making the one fix

Inspect the generated image itself, including the intended crop. Apply only checks
relevant to the brief; a deliberately deep-focus product image need not imitate a
shallow-focus portrait. Scoring the prompt is not evidence about the output.

---

## How to run the pass

1. Inspect the candidate. When alternatives were requested, compare them against
   the same brief and select the strongest composition and moment.
2. Check applicable truth failures first, then whether the image serves its job.
3. If it meets the brief, stop. Do not invent a flaw to justify another render.
4. Otherwise identify its worst material failure and make a targeted correction.
   Preserve successful details; reconsider the whole prompt only if the concept
   or physical setup is the underlying failure.

When repeated attempts fail on the same axis, revise that physical decision rather
than repeating the same prompt.

---

## The rubric

### Truth – does it read as a photograph?

| # | Check | Failing looks like |
|---|---|---|
| 1 | **One light source story.** Do all shadows agree on direction, hardness, and color? | Shadows pointing two ways; a soft shadow and a hard shadow from the same light; a lit face with no corresponding highlight anywhere else |
| 2 | **Catchlights match the modifier.** Window = rectangle, softbox = round, practical = point, overcast = broad and dim. | Two identical circular catchlights in a room with one window; catchlights with no source in the scene |
| 3 | **Selective sharpness.** Is there a focal plane, or is everything equally crisp? | Uniform micro-detail front to back – the single most reliable render tell |
| 4 | **Skin.** Pores, asymmetry, subsurface scatter at ears/nostrils/fingers near light, sebum on the T-zone. | Wax, airbrush, uniform tone, poreless cheeks, symmetrical face |
| 5 | **The four failure zones** – hands, text, reflections, eyes. | Extra or fused fingers; garbled lettering; a reflection that doesn't match what's in front of it; mismatched pupils or dead gaze |
| 6 | **Materials behave.** Metal shows anisotropic streaks, fabric shows weave and drape under gravity, glass refracts and has thickness, contact shadows exist where objects touch. | Objects floating a millimetre above surfaces; plastic-looking metal; fabric with no weight |
| 7 | **Scale and perspective hold.** Do object sizes and the horizon agree with the stated lens and camera height? | Wide-angle framing with telephoto compression; furniture the wrong size; two vanishing points fighting |

### Merit – is it worth looking at?

| # | Check | Failing looks like |
|---|---|---|
| 8 | **Figure-to-ground separation.** Does the subject stand clear by tone, color, or focus? | Subject's dark jacket merging into a dark background; the frame reads as texture, not as a subject |
| 9 | **Edge discipline.** Nothing kissing the frame edge; no background line growing out of a head; no limb cropped at a joint. | Unresolved tangents – the most common amateur tell, and models produce them constantly |
| 10 | **A moment is visible.** Can you say what just happened? | A person posed mid-nothing; an activity averaged into no instant |
| 11 | **Depth in layers.** Is there something in the foreground, midground, and background? | A subject pasted onto a backdrop; two planes where there should be three |
| 12 | **Palette hierarchy.** Is there a dominant, a support, and an accent – or six colors of equal weight? | Mud. Every hue at the same saturation |
| 13 | **Distinctive detail.** When the brief needs a specific detail, did it survive? | A requested distinguishing feature is absent or changed |
| 14 | **Fitness for use.** Does the result meet the brief at its intended crop? | A concrete composition, content or quality gap that affects its job |

---

## Failure → fix

Change one thing. The right-hand column is the edit, not a rewrite.

| Symptom | Edit |
|---|---|
| Everything equally sharp | Block 4: drop the aperture (f/1.4–f/2.8) and name the focal plane *and* what falls off – "sharp on the near eye, the far shoulder already soft" |
| Waxy or airbrushed skin | Name plausible skin texture and light response for this subject; preserve reference identity |
| Flat, no drama | Block 3: state a ratio (4:1 or 8:1) and kill the fill – "no fill from camera right; that side falls to near-black" |
| Muddy color | Block 10: state the relationship before the tones. "Near-monochrome cool grey with a single sodium-orange break" |
| Subject lost in background | Block 5: name the separation mechanism – "the pale wall behind him is two stops brighter, silhouetting the shoulder line" |
| Feels staged | Reconsider action, vantage and the specific moment; add incidental objects only if they belong |
| Feels generic | Block 9: the punctum was dropped or too vague. Make it a physical object with an implied history |
| Tangents at the edges | Block 5: "the doorframe clears his head by a hand's width; no line touches the frame edge" |
| Symmetrical AI face | Block 7: "asymmetrical expression, one brow slightly higher, weight on the left foot, mid-blink" |
| Composition is boring | Block 1: change the vantage, not the words – through a doorway, over a shoulder, from below the work surface, reflected in glass |
| Text is garbled | Use the selected tool’s text-capable edit/generation mode when available; shorten and quote the exact requested text |
| Hands wrong | Block 7: give the hands a specific task and state the grip – "left hand braced flat on the panel, right hand closed around the wrench handle" |
| Too clean / too pretty | `references/looks.md` §4: add the relevant clichés to the block 10 negative list |
| Right idea, wrong intensity | Do not touch the blocks. Generate another candidate only within the authorized count and budget – some of this is sampling luck |

---

## When to stop

Stop as soon as the image meets the brief. After two unsuccessful targeted
correction rounds, reconsider the concept or disclose the unresolved limit.
