# Working on this ItsyWeb site

This is `website/` in the combined ItsyWeb CP Git repository. The CLI/server
source and generated starter live in sibling `../system/`. Paths in this
guide are relative to `website/`; run `itsy` site commands here. Commit changes
explicitly before deployment because the site is inside a larger repository.
Preserve both saved branch deployment mappings in `itsy.config.json`; `main`
targets fleacms.com and `t3/reframe-flea-html-focus` targets itsyweb.dev. This
repository move does not authorize deploying or changing those mappings.

Read README.md before changing templates. Preserve existing custom layouts,
content and assets. There is no frontend build step or Node requirement.
`itsy new` is minimal; `--starter`/`-s` adds example pages. The shared navigation
partial drives desktop links and the native mobile popover. `itsy reset` replaces
all local files/settings except Git history after confirmation and keeps a private
sibling backup. Never run it as an incidental cleanup step or deploy automatically.

Starting dev again replaces this site's existing dev server;
pull/reset remain exclusive. Stop previews you start before handing control back
unless the user asks to leave them running.
New sites enable dev.email_preview in itsy.config.json. Local sign-in/contact
mail opens in the browser without Resend credentials. Preview mode and sender
details require a dev restart after changes; production ignores the toggle.
Do not write preview messages or sign-in codes into site files.

Public pages, partials, layouts and assets belong in site/. Use main.layout.html
as the default wrapper and *.layout.html for custom layouts. Root itsy.config.json
holds configuration and deploy.branches; never put secrets there. Blog posts and
cover-photo records live in itsy/snips.json, cover files in itsy/media/, and
ignored runtime in itsy/runtime/. Ordinary page copy lives in HTML; title and
description belong in one leading settings comment with optional layout.
Mail templates live in root mailers/, named *.email.html and *.email.txt with
main.layout.html. Keep public URLs site-relative; never expose private directories.

Blog keeps `/admin` sign-in and `/admin/blog` management, with its own admin
layout and CSRF controls. Preserve `{admin.csrf}`, form actions/names, exact-Origin
checks, sessions and stale-version checks. Public layouts have no page editor
mount or admin CSRF meta. Never configure `app.current_admin` in JSON. Crossfades
remain ordinary css/navigation.css imported by global.css.

Use `{#unless dotted.path}...{/unless}` for negative conditions. Like if, it
supports `{:else}` and `{:else if dotted.path}`; conditions are dotted paths only.

Write page copy as ordinary HTML and use itsy/data/ for structured lists. Use
ordinary image/video tags and source-controlled assets in site/. Developer CSS
and theme tokens remain in site/css/. Blog retains its post editor and cover
uploads; page editing and browser theme controls are removed.

This website requires the removal/metadata candidate described in README.md.
The public 0.21.0 installer is insufficient. Use the rebuilt candidate for checks;
protocol 30 deployment and a matching release are prerequisites for publication.
Do not deploy or upgrade a server as part of local maintenance.

At task start, read [.agents/skills/core/SKILL.md](.agents/skills/core/SKILL.md) once.
For site authoring, read [itsyweb-site](.agents/skills/itsyweb-site/SKILL.md).
Read [Prewind](.agents/skills/prewind/SKILL.md) for styling,
[design](.agents/skills/design/SKILL.md) for visual/UX work,
[copywriting](.agents/skills/copywriting/SKILL.md) for copy and
[task-delivery](.agents/skills/task-delivery/SKILL.md) for checks and Git delivery.
Use only runtime-discovered global Impeccable; keep its installation outside this site.

## Available skills

For requests like "add a blog" or "let's create a blog", read
`itsy/skills/add-blog/SKILL.md`. For "add a contact form" or an email-enabled
contact page, read `itsy/skills/add-contact-form/SKILL.md`. These focused additions
use the existing site's design and do not require a full-site consultation.

- **import-brochure**: import an existing small brochure website into this site,
  preserving its content and navigation with basic Prewind styling. Read
  `itsy/skills/import-brochure/SKILL.md` when doing that work. This is a portable
  instruction file, not an automatically installed agent plugin.

Agent guidance stays in Git, but is excluded from public serving and deployment.
Do not put credentials in guidance or templates. Ask before adding dependencies,
backward compatibility or changing deployment/access settings.
