---
name: itsykit-site
description: Build or maintain ItsyKit websites with the itsy CLI, plain HTML/JSON, Contact, optional Blog posts and cover photos, managed apps, migration redirects, SEO and content synchronization. Includes existing Flea sites; excludes CLI/server framework development.
---

# Build an ItsyKit website

Read the site's `AGENTS.md`, relevant README sections and installed version
(`itsy version`). Inspect affected config, layouts, content and assets. Preserve
custom branding, ordinary HTML/JSON and the no-build setup. Project instructions
and explicit user choices take precedence; this skill grants no action permissions.
Resolve conflicting historical docs against the installed implementation. Roadmaps
and `docs/agents/spec.md` include future plans, not just supported behavior.

Use `itsy`, not retired `flea` commands. New sites
use `itsy.config.json`, `itsy/` and `<!-- itsy` page settings. Existing
`.flea.config.json`, `flea/` and `<!-- flea` settings remain supported. Prefer the
Itsy config and directory independently when present; otherwise use the legacy
ones. All reads and writes stay in the selected files. Never create parallel
content trees, silently rename files or migrate sites. Examples below use the
new names; adapt paths to the selected existing layout.

For installation, transition, deployment, pull or recovery, read
[compatibility and operations](references/compatibility-operations.md).
The page-CMS removal contract below requires ItsyKit 0.22.0 or newer and
server protocol 30. Earlier page-CMS sites need an explicitly approved conversion
with verified backups before an upgrade. See the operations reference for
compatibility and upgrade limits.

## Load only relevant guidance

- Drafting or editing copy: read [copywriting](../copywriting/SKILL.md).
- Styling, layout or markup: read [prewind](../prewind/SKILL.md).
- Visual direction, UX, redesign or polish: also read [design](../design/SKILL.md).
  Use only runtime-discovered global Impeccable; never install, vendor, copy or
  link it into a site or starter. Keep design context and artifacts in the site.
- Template syntax, layouts, page metadata or static media: read
  [templates and content](references/templates-content.md).
- Contact, Blog or custom managed apps: read [apps](references/apps.md) and the
  relevant installed guide in `itsy/skills/` (or the selected `flea/skills/`).
- Existing-site migration, URL inventory or 301s: read
  [migration redirects](references/migration-redirects.md).
- Sitemap, robots, canonicals or search exclusions: read [SEO](references/seo.md).

Read installed files explicitly; links do not load dependencies. Report missing
skills and continue permitted work from available guidance. Do not fetch sibling
copies or load Rails/Preflight application conventions for an ItsyKit site.

## Content and direction

For a new site or substantial redesign, establish the primary visitor action
from the brief: directions/call, bookings, sales or inquiries. Ask one focused
question only if consequential direction remains unclear. Do not invent a booking
or commerce workflow. For an existing-site redesign, follow the design skill's
early choice between exact current copy and rethinking the messaging. A content
import or migration alone does not authorize rewriting it.

For a real business, inspect first-party materials and supplied assets before
asking for facts. Distinguish owner-confirmed facts, corroborated facts,
unverified claims and lore. Resolve conflicting hours, menus and prices with the
owner. Do not invent history, awards or capacity. Quote reviews verbatim with
attribution; label lore rather than stating it as fact.

Use real business photos with suitable permission. Generated images cannot prove
its premises, food or people. Preserve supplied logos; derive a logo from signage
as vector SVG rather than embedding a bitmap. Use the business and audience to
inform design. Follow the authored design skill's task-sized workflow, research
and effort defaults, concrete direction review and sample-page review for
substantial work. Honor settled choices and delegated decisions. Routine
maintenance and focused app additions do not restart a full-site consultation.

## Site authoring defaults

Write page copy directly in HTML. Use `itsy/data/*.json`, `data.*` and `#each`
for structured lists. Images and videos are ordinary static assets under `site/`.
Keep developer CSS, Prewind and `site/css/theme.css` tokens with code. Ordinary
page title/description and optional layout belong in one leading settings comment;
see [templates and content](references/templates-content.md) for parsing and fallbacks.

The page CMS is removed: do not add text/media snips, named/structured snips,
multisnips, extraction, page editing/formatting, a page toolbar or browser theme
controls. Blog retains its rich admin editor, drafts, publishing and cover uploads.
The selected `itsy/snips.json` (or legacy `flea/snips.json`) stores only optional
`blog` and `media`; `{}` is valid. Its name does not imply page-snips support.

Old page-CMS fields, named-content directories and pending extraction journals
fail closed without rewriting. Preserve a backup and require an explicitly scoped
source conversion before using this runtime on an old site. Never discard content
or media to pass validation. This skill authorizes no automatic conversion, bulk
migration, compatibility layer, pruning or live-site cleanup.

- Public pages/assets belong in `site/`; mail templates in private `mailers/`.
  Config globals use `app.*`, not `site.*`. Keep secrets out of config and Git.
- Preserve custom layouts, partials, navigation and native mobile behavior.
  `main.layout.html` uses `{slot}`; custom layouts use leading page settings.
- Public layouts may link to `/admin`; they need no admin CSRF meta or toolbar
  mount and have no `app.current_admin` rendering state. Preserve private admin
  CSRF/login/Blog controls and Contact's separate `{csrf}` input.
- Keep theme decisions in `site/css/theme.css`, shared rules in
  `site/css/global.css` under `@layer styles`, and vendored assets unchanged.
  Use installed Prewind utilities and tokens; do not invent Tailwind classes,
  a parallel component system or a build step. Custom CSS follows site policy.
- Use `{icon name}` and the selected `itsy/icons/` cache. Dev resolves missing
  referenced icons; status is read-only. Keep caches committed.
- Agent guidance stays committed but private and excluded from deployment.
  Keep runtime state in the selected private `itsy/runtime/`, never public assets.

## Edit and verify

Dev validates and reloads page source without extraction or content rewrites.
Coordinate with an existing process owner before replacing their dev server;
restart when required by config changes. Dev can resolve referenced icons.
`itsy status` validates without extracting or downloading icons. It also contacts
the current branch's saved remote automatically, if configured; it is not a
local-only or offline check in that case.

Inspect rendered HTTP responses, title/copy and unresolved directives. Check
changed desktop/mobile layouts, navigation, forms and interactions. Exercise
Blog Save/Cancel when affected, using disposable fixtures or explicit authority
for publication. Review changed CSS for duplication or unintended overrides.
Use local email preview for sign-in/Contact checks when enabled; keep messages,
codes and capability links out of site files.

Finish with `itsy status` and review templates with changed config, metadata and
content. Verify ordinary pages use HTML/JSON and comment metadata, with no page-CMS
attributes, bindings, collections, metadata overrides or public admin state.
Stop task-owned servers unless asked to keep them. Report actual checks and
unverified boundaries. Do not run reset, pull, restore, publish posts, deploy
or upgrade a server as incidental cleanup.
