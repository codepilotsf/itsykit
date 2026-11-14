---
name: ItsyWeb
description: Navy and ice, readable HTML, and a complete path from source to HTTPS.
colors:
  navy: "#142b43"
  hero: "#0b2842"
  blue: "#245f91"
  ice: "#a4dbff"
  pale: "#f5f9ff"
  code: "#091b2e"
  pane: "#17374f"
  code-border: "#536e85"
  body-ink: "#304962"
  muted: "#4f657b"
  line: "#d8e3ed"
  white: "#ffffff"
typography:
  display:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "clamp(2.25rem, 4.4vw, 3.75rem)"
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: "-0.035em"
  body:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
  code:
    fontFamily: "ui-monospace, monospace"
rounded:
  default: "0.625rem"
  large: "1rem"
  pane: "12px"
spacing:
  page-gutter: "clamp(20px, 3.5vw, 56px)"
components:
  button-primary:
    backgroundColor: "{colors.ice}"
    textColor: "{colors.navy}"
    rounded: "{rounded.default}"
    height: "48px"
    padding: "13px 22px"
  tab-selected:
    backgroundColor: "{colors.ice}"
    textColor: "{colors.code}"
    rounded: "{rounded.default}"
    height: "48px"
---

# Design system: ItsyWeb

## Overview

**Creative North Star: "One simple project, from source to live"**

The approved design uses the existing vector logo, Inter typography and navy/ice identity. Real HTML examples carry the visual focus. Pale reading surfaces separate explanations from the dark source and hosting demonstrations.

**Key Characteristics:**

- Readable source beside its rendered result.
- Ice actions and selected states on navy.
- Flat surfaces with thin borders.
- A shared visual language for marketing and documentation.

## Colors

Ice identifies the primary action on dark surfaces and the active project view. Navy supplies headings on pale and white grounds. Blue supplies reading links. Dark code uses the code surface, pane headers and code-border tokens; secondary text on navy uses `--hero-muted`. Use the existing CSS semantic tokens for forms and feedback.

## Typography

Inter supplies both headings and reading text. Monospace belongs to source, filenames and terminal commands. The homepage display role above has separate phone rules in `homepage.css`; section headings use fluid sizes and tight line heights. Reading paragraphs use relaxed spacing and the existing text measure. Documentation keeps its established heading and article hierarchy.

## Layout

The shared header uses a 1200px container. The upper homepage and agent section use 1120px containers with fluid page gutters; the hero demonstration caps at 960px. The headline and description have narrower centered measures and generous separation. Source and result share two equal columns; at 520px they stack. The earlier five-step card row is removed: Develop links to installation, while Develop and Publish reveal the commands on demand. The HTML explanation precedes the coding-agent section. Lower sections retain their 1280px container.

The agent demonstration puts the file tree beside the file content; at 520px it becomes horizontal file tabs above the content. At 800px the surrounding heading and explanation stack. Documentation retains its sidebar and reading column, with the native Guides disclosure below 640px. The shared header offers Docs and Get started; the homepage agent section and documentation retain access to the coding-agent guide.

Documentation uses the shared page gutter. Its pale sidebar has the large radius and a thin border; the sticky offset accounts for the admin toolbar. Below 640px the Guides disclosure takes the same pale surface and gently curved panel treatment.

## Elevation & Depth

The homepage uses tonal layering and thin borders rather than added shadows. Navy, the darker code surface, pane headers and the white preview establish hierarchy. Existing framework shadow tokens remain available to other components; this redesign adds none.

## Shapes

Actions and tabs use the default radius. Large demonstrations use the large radius, with gently curved inner panes and workflow steps. Preserve the supplied SVG brand assets and consistent library icons.

## Components

**Actions:** ice on navy; the shared header action reverses to navy on pale. Keep visible focus and a generous target. The homepage hero action has a minimum 50px height.

**Project tabs:** semantic links work as section anchors without JavaScript. Enhancement supplies tab roles, selection, arrow/Home/End keys and focus management. Hero views use an ice underline and text; file tabs keep the pale blue selection. Inactive panels are hidden only after enhancement.

**Code panes:** the hero demonstration uses a single outer frame with adjoining source and result panes, without nested card borders or decorative window dots. Source remains scrollable with a filename or terminal label. Syntax accents are restrained and line numbers are decorative. Copy controls use a 44px target and announce success through the shared copy script.

**Example preview:** native HTML, clearly labeled as an example. Its local buttons change authored example content without navigating the visitor away. The displayed port follows ItsyWeb's actual default.

**File tabs:** selected filenames use pale blue. Desktop arrow navigation is vertical, matching the file tree; phone navigation is horizontal, matching its reflow.

**Guide navigation:** the current guide uses ice and navy, matching the homepage view tabs. The current chapter uses pale blue. Code panes have 12px corners and pane-colored copy controls; next-guide links use a pale, bordered panel with a pale blue hover state.

**Motion:** a single 200ms preview refresh with exponential easing. Tab color changes take 160ms. Both are limited to visitors who have not requested reduced motion.

## Do's and Don'ts

- Do use the shared CSS tokens and existing SVG logo.
- Do keep code readable, keyboard focus visible and demonstrations usable without scripts.
- Do preserve documentation's reading structure and native disclosure.
- Don't turn the homepage's composition into a requirement for every page.
- Don't invent benchmarks, customer claims or a public source-repository link.
- Don't imply that server setup provisions a cloud VM or that a coding agent is embedded in ItsyWeb.
