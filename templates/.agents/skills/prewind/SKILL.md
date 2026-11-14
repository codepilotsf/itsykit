---
name: prewind
description: Style sites and applications that use Prewind. Use for theme tokens, utility classes, fluid typography and spacing, layout, and responsive styling. Applies across ItsyWeb and Rails; framework-specific forms, components and file locations belong to their authoring skills.
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
- Put content inside `.container`; a full-width band may wrap it. Do not add
  `max-w-full` as a workaround that defeats the intended content width.
- Before adding custom CSS, check existing defaults and utilities. Do not recreate
  a supported utility with a custom selector or add a parallel utility system.
- Use custom CSS for behavior or visual detail the installed system cannot
  reasonably express. Follow the application's stylesheet boundaries, cascade
  layers and approval rules; a reusable component policy is application-specific.
- Preserve the established header, main, footer and responsive navigation unless
  the task changes them. Keep keyboard focus and reduced-motion behavior usable.

Inspect changed markup and selectors for unsupported classes, utility duplication
and unintended overrides. Verify responsive rendering and interactions in the
browser; source inspection alone does not establish visual correctness.

This skill does not choose Simple Form, Zjax, a frontend build step or stylesheet
paths. Load the installed ItsyWeb or Preflight authoring skill when applicable;
use the design skill when visual direction or UX decisions are part of the task.
