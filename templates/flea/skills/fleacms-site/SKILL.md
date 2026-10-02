---
name: fleacms-site
description: Build or redesign a website in a FleaCMS project, especially for one real local business (bar, restaurant, shop, venue). Use when asked to research a business and build or redo its site, and for any substantial FleaCMS authoring task. Covers deciding the site's job, research-first vibe discovery, real assets (including SVG logos taken from the actual signage), content truth rules, genre-driven design, and the safe editing + dev/verify workflow for this repo.
---

# Build a site in FleaCMS

## Why this exists

Two ways an agent ruins a site, and one way it wrecks the repo:

- **Inventing content** — a generic template, fabricated hours/menu/history, and AI
  photos of some other building. For a characterful place, *the character is the
  product*; a slick site for a gritty dive bar is a failed site.
- **Over-interrogating the owner** — running a long interview for facts it could have
  researched itself, or building before understanding the place.
- **Racing file writes** — parallel edits and a running dev watcher corrupt templates.

This skill is the middle path plus the guardrails.

## 0. Decide the job before anything

Name the single primary action the site must drive: **walk-ins** (directions + call),
**bookings**, **sales**, or **inquiries**. Match the CTA to what the business can
actually deliver — never invent e-commerce, reservations, or a contact workflow it does
not run. A cash-only, walk-in bar's honest job is a *brochure that gets the right people
through the door*: correct hours/address/phone and a truthful picture of the place.
Confirm the job with the owner in **one** question.

## 1. Research first — ask only what you can't find

Do the homework the internet allows *before* asking anything: Google Business Profile,
Yelp (photos especially), the business's own Facebook/Instagram, review text, local
press. Build a specific, evidence-backed picture — and pull the **real photos and the
real logo/sign**.

Then ask the owner only what research cannot settle, **one question at a time**
(typically 2–4 total): a read-back to confirm/correct ("here's what I found — right?"),
the facts that must be exact (hours, what is/isn't served, name, phone), asset
permission, and anything never to say. A question good research could answer is wasted
on the owner. Never dump a list of questions at once.

## 2. Source discipline

Tag every fact: **owner-confirmed** / **verified** (multiple independent sources) /
**unverified** / **lore**. Publish only the first two; label lore as lore. Directory
aggregators and "best of" listicles are frequently AI-generated — treat their prose as
unusable even when it sounds specific.

## 3. Vibe, then genre-driven design

Learn what the place *is* from real material before designing. Then pick the aesthetic
from the **category and the real interior**, not from a pretty template. A dive bar is
not a café.

| Type | Design language |
| --- | --- |
| Dive bar | Dark/high-contrast, neon & letterboard, cheap-beer Americana, gritty real photos, big hours + phone |
| Craft cocktail / wine | Moody, dim, refined serif, restrained |
| Diner / burger joint | Bright, retro, red/cream, big food photos |
| Coffee shop | Warm minimal, editorial, airy |
| Taqueria | Bold, colorful, loud, menu-first |
| Music venue | Poster art, listings/calendar-first, high energy |

Test: swap the photos and copy out — would the layout fit *any* place? If yes, it's too generic.

## 4. Real assets — and no bitmaps

- **Real photos of *this* business** first (owner's own; else Yelp/Google/Facebook/press),
  downloaded into `site/images/`. Factual alt text. Raise the rights point out loud.
- **Never** use AI to depict the business's premises, food, drinks, or people, and never
  present an AI image as the place.
- **Logos:** derive from the *actual signage*. Build them as **SVG** (paths/type), keep
  them pure vector (no `<image>`, no base64), and **do not rasterize**. Do not introduce
  bitmap tooling (ImageMagick, etc.); if you must preview for yourself, keep it in a
  scratch directory and never ship it.
- Derive palette and type from the real paint, signage, and photos.

## 5. Content truth rules

- **Quotes: verbatim only**, attributed. Never paraphrase a review into a "quote".
- **Hours / menu / prices: owner only.** If listings disagree, don't average them.
- No invented history, awards, founding years, capacity, or claims.
- Mood-words ("quaint", "charming", "hidden gem") are a smell. Use the owner's and the
  regulars' own language.

## 6. Working rhythm

Small loops: **research → a one-paragraph vibe read → one question → a tiny draft →
adjust.** Show work early; do not build everything before the direction is confirmed.

## 7. FleaCMS project mechanics

- Pages live in `site/`: `index.html` = `/`, `about.html` = `/about`. Public images in
  `site/images/`. Public values in `.flea.config.json` under `app.*`; secrets never go
  there (use the ignored `.env`).
- `site/main.layout.html` wraps pages via `{slot}`. Keep the `{#if app.current_admin}`
  CSRF meta in `<head>` and `{render /admin/toolbar}` right inside `<body>`.
- Editable text/images use a bare `snip`; on save, dev rewrites it to `{snip ID}` and
  stores the value in `flea/snips.json` (**the source of truth**). Don't invent IDs.
  Multisnips (`multisnip="name"`) for lists.
- Page `title`/`description` live per-route in `flea/snips.json`.
- Theme tokens live in `site/css/theme.css`; site-wide rules go in
  `site/css/global.css` under `@layer styles`. Icons via `{icon name}`; dev downloads
  them into `flea/icons/`.
- No npm or build step. **No deploy or commit unless asked.**

## 8. Safe editing in this environment (critical)

- **One writer per file.** Never issue parallel edit/write calls to the same file — they
  race and interleave, corrupting it. Batch across *different* files only, or write
  sequentially.
- **Do not run `flea dev` while hand-editing templates**: its watcher rewrites files as
  it extracts snips and will fight your writes. Stop dev, edit, then start it again.
- Editors with **format-on-save** (e.g. Zed) reflow HTML/CSS buffers; close those files
  or expect reformatting.
- **Write atomically** (a single write, or one shell heredoc) and **verify after every
  write**: line count, a unique marker (e.g. count of `<!doctype html>`), and no
  duplicated fragments.

## 9. Dev watcher, caching, and the extract/verify loop

- Run `flea dev` **once** to extract new `snip` regions and download referenced icons,
  then stop it. `flea status` validates but does not extract or download.
- Dev **caches** page metadata and snip content: after editing `flea/snips.json` or page
  metadata directly, **restart dev** to see the change.
- To change a snip's stored text, edit `flea/snips.json`; the template only holds the
  `{snip ID}`. Do not invent IDs.

## 10. Verify without bitmaps

Verify by **fetching the rendered HTML** and checking: HTTP 200, `<title>`, resolved
`{snip}`/`{app}`/`{icon}` (zero unresolved tokens), and the specific copy you changed.
Avoid rasterizing just to check. If a screenshot is genuinely needed, keep it in a
scratch directory and never ship it.

## 11. CSS and header gotchas

- Watch for **leftover selectors overriding new ones** (e.g. a `.brand img { width:42px;
  height:42px }` rule squashing a wide logo). Remove or rescope stale rules and verify
  the computed size and the **aspect ratio**.
- Give the **header** room — adequate `padding-block` and gaps. It is the most-seen
  element; check it first.
- Use the shipped theme tokens and Prewind utilities, and confirm a utility exists before
  using it.

## 12. Definition of done

- Every fact is owner-confirmed or verified; every image is of the actual business.
- The owner/requester agrees it **feels like the place**.
- `flea status` passes; new snips extracted; icons cached.
- Nothing deployed or committed without an explicit ask.

## Anti-patterns — do not

- Build before deciding the job or learning the vibe.
- Invent menu/hours/history; paraphrase reviews into quotes.
- Use AI photos of a real business; use bitmaps for logos.
- Copy a template aesthetic unrelated to the category.
- Ask the owner things you could research, or dump many questions at once.
- Edit the same file in parallel, or while a dev watcher runs.
- Rasterize to "check" and leave artifacts behind.
