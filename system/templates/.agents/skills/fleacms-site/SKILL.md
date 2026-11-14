---
name: fleacms-site
description: Build or maintain Flea CMS websites, including templates, optional editable snips, Contact forms and migration redirects. Excludes development of the Flea CMS framework itself.
---

# Build a Flea CMS website

Use the site's already-read `AGENTS.md`; consult relevant README sections and inspect
the installed Flea version and affected layouts, content and assets. Local instructions and explicit user choices
take precedence. Keep ordinary HTML/JSON and the existing no-build setup.

## Load the relevant shared guidance

- Drafting or editing website, marketing or product copy: read
  [copywriting](../copywriting/SKILL.md).
- Styling, layout, theme or markup: read [prewind](../prewind/SKILL.md).
- Visual direction, UX, redesign or polish: also read
  [design](../design/SKILL.md), which routes to Impeccable.
- Read local task-specific guides, such as `flea/skills/import-brochure/SKILL.md`,
  when the task requires their import workflow.

Read installed files explicitly; this is not an automatic dependency loader. If
a shared skill is missing, report it and continue from available project guidance.
Do not load Rails or Preflight conventions for a Flea site.

## Content and direction

For a new site or substantial redesign, establish the primary visitor action
from the brief: directions/call, bookings, sales or inquiries. Ask one question
only if that job is unclear. Do not invent a booking or commerce workflow.
For an existing-site rebuild or redesign, also follow the design skill's early
choice between exact current copy and rethinking the messaging. Settle that
scope before visual directions; a content import or migration alone does not
authorize rewriting the message.

For a real business, research available first-party materials and supplied assets
before asking for facts. Distinguish owner-confirmed facts, corroborated facts,
unverified claims and lore. Resolve conflicting hours, menus and prices with the
owner; do not invent history, awards or capacity. Quote reviews only verbatim
with attribution. Label lore rather than presenting it as established fact.

Use real photos of the business with suitable permission. Do not use generated
images as evidence of its premises, food or people. Preserve supplied logos; when
deriving one from signage, keep the deliverable vector SVG rather than embedding
a bitmap. Derive visual direction from the business and its audience.

For substantial design work, show concrete direction options and one sample page
before expanding the site. Follow the design skill for visible comps. Ask one
question at a time by default, respecting personal conversation preferences and
an explicit instruction to proceed without pauses.
Routine maintenance does not require restarting this process.

## Templates and content

Build agent-managed Flea websites without snips by default. Write copy directly
in HTML templates, keep media in `site/`, and use `flea/data/` for structured lists
when useful. Do not add text, image or video snips, `snip` attributes, generated
snip bindings or multisnips unless the user explicitly requests browser-editable
content. A request to build, import or redesign a site is not such a request.
Preserve existing snips unless their conversion is requested; do not silently
remove a client's editing workflow. Route titles and descriptions may still live
in `flea/snips.json` without any editable snip entries.

- Public pages and assets live in `site/`. `site/index.html` maps to `/` and
  `site/about.html` to `/about`; use site-relative public URLs. Keep private
  config, mailers, runtime and agent guidance out of public assets.
- `site/main.layout.html` wraps pages with `{slot}`. Preserve existing custom
  layouts, partials and shared navigation, including native mobile behavior.
- Keep the conditional admin CSRF meta in `<head>` and conditional
  `{render /admin/toolbar}` directly inside `<body>`. Preserve the toolbar's
  version attributes, declarative shadow-root shell and external script.
  `app.current_admin` is request-supplied, not a JSON setting.
- Public settings belong under `app` in `.flea.config.json`; secrets do not.
  Developer-managed JSON lives in `flea/data/` and mail templates in `mailers/`.
- Only when browser editing is explicitly requested, use bare `snip` attributes
  for the requested regions and let dev extract IDs.
  Existing values and route metadata live in `flea/snips.json`; edit that source
  rather than replacing a generated reference. Do not invent IDs or nest snips.
- When requested, media snips extract image src/alt or video src/poster bindings.
  Keep classes, dimensions, links and playback behavior in templates. Do not combine
  a media snip with competing `picture`, `source` or `srcset` bindings.
- Use multisnips only when requested; follow local examples. Preserve page-level
  groups with a single root and named fields; do not nest groups or place them in
  layouts/partials.
- Use `{icon name}` and the `flea/icons/` cache. Do not introduce another icon or
  component system to follow a design example.

## Migrations and permanent redirects

For an existing-site migration, URL inventory or redirect request, read
[migration redirects](references/migration-redirects.md). This workflow can run
on its own without a redesign consultation: discover public URLs from the current
website, map them to the new site's pages, and update the top-level `redirects`
object in `.flea.config.json`. Preserve existing rules and verify actual 301
responses locally. Check installed runtime support before relying on the config;
older Flea versions silently ignore it. Do not deploy or upgrade as an incidental
part of an inventory.

## Styling and Contact

Keep theme decisions in `site/css/theme.css`; approved shared rules belong in
`site/css/global.css` under `@layer styles`. Preserve vendored assets. Do not
introduce BEM or page-specific styling systems. Before adding custom CSS that
Prewind cannot express, explain the exception and obtain the approval required
by the site's styling policy. Preserve useful sticky headers, blur, navigation
and transitions unless their behavior is part of the requested change.

Contact uses ordinary HTML controls and an explicit `{csrf}` directive. Inspect
the site's Contact documentation and example before changing it: enable
`apps.contact` with its mount and `email.to`, post to the mount, and provide its
`index.html` and `thank-you.html`. Use `contact.data.*`, `contact.invalid.*`,
`contact.error` and `contact.flash.*`. Do not replace admin CSRF contracts or
assume arbitrary app installation support. Restart dev after app config changes.

For an installed optional Blog, read the site's `flea/skills/add-blog/SKILL.md`
and current Blog guide before editing post records, templates or publishing.
Preserve stable IDs, draft visibility and explicit Publish/Unpublish. Other app
support depends on the site's installed version; do not infer it from this skill.

## Edit and verify safely

Dev rewrites templates when snip extraction is needed. Stop a task-owned dev watcher
before bulk rewrites, then restart for rendering and any required icon downloads
or explicitly requested snip extraction. Coordinate with an existing process owner
before replacing or stopping their server.
`flea status` validates but does not perform extraction or download icons.

Use local email preview for sign-in/contact checks when enabled. Keep preview
messages and codes out of site files. Do not run reset, pull, restore or deployment
as incidental cleanup. Follow the site's authorization and process-lifecycle rules.

Check rendered HTTP responses, title/copy and unresolved template directives.
Inspect desktop/mobile rendering and exercise changed interactions, navigation,
forms and edit Save/Cancel where affected. Review changed CSS for duplication and
unintended overrides; verify behavior beyond screenshots. Finish with `flea status`
and a review of templates together with any changed metadata or content. For new
sites without requested browser editing, verify there are no snip attributes,
snip bindings, multisnips or editable entries in `flea/snips.json`. Report unrun
checks and remaining limits; this skill grants no commit or deployment permission.
