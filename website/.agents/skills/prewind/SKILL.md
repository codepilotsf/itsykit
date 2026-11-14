---
name: prewind
description: Style Prewind sites and applications using their installed tokens, utilities and responsive conventions. Framework-specific forms and file locations belong to their authoring skills.
---

# Prewind styling

Read the project's instructions and inspect its loaded Prewind stylesheet, version,
theme and nearby markup. Use that project's Prewind documentation as the class
reference. Do not assume Tailwind classes, arbitrary-value syntax, responsive
prefixes, grid utilities or animation utilities exist in the installed version.

- Start with Prewind's shared element defaults and theme tokens. Use documented
  utilities for layout, spacing, sizing, alignment and responsive adjustments.
- Reuse the theme's fluid type and spacing scale. Change shared tokens in the
  project's theme file rather than scattering competing values through templates.
- Use `.container` for content that follows the site's shared measure. An approved
  full-bleed image, graphic background, oversized display treatment or asymmetric
  composition may sit outside it with deliberate responsive bounds. Do not use
  `max-w-full` to defeat the container; express that composition explicitly.
- Before adding custom CSS, check existing defaults and utilities. Do not recreate
  a supported utility with a custom selector or add a parallel utility system.
- Use custom CSS for behavior or visual detail the installed system cannot
  reasonably express. Follow the application's stylesheet boundaries, cascade
  layers and approval rules; a reusable component policy is application-specific.
- Preserve the established header, main, footer and responsive navigation unless
  the task changes them. Keep keyboard focus and reduced-motion behavior usable.

Prewind supplies styling primitives, not an art direction. For new sites and
approved redesigns, do not treat starter colors, font choices, card examples or
default scale values as approved brand constraints. Implement the selected visual
system through theme tokens and supported utilities, using permitted custom CSS
for authored graphics and motion that utilities cannot reasonably express. Keep
the stack and project approval rules; report a concrete policy conflict instead
of quietly reducing the design to the starter's appearance.

Inspect changed markup and selectors for unsupported classes, utility duplication
and unintended overrides. Verify responsive rendering and interactions in the
browser; source inspection alone does not establish visual correctness.

This skill does not choose Simple Form, Zjax, a frontend build step or stylesheet
paths. Load the installed ItsyWeb or Preflight authoring skill when applicable;
use the design skill when visual direction or UX decisions are part of the task.
