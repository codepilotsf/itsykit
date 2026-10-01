# Working on this FleaCMS site

Read README.md before changing templates. Preserve existing custom layouts,
content and assets. There is no frontend build step or Node requirement.
`flea new` is minimal; `--starter`/`-s` adds example pages. The shared navigation
partial drives desktop links and the native mobile popover. `flea reset` replaces
all local files/settings except Git history after confirmation and keeps a private
sibling backup. Never run it as an incidental cleanup step or deploy automatically.

New sites enable dev.email_preview in .flea.config.json. Local sign-in/contact
mail opens in the browser without Resend credentials. Preview mode and sender
details require a dev restart after changes; production ignores the toggle.
Do not write preview messages or sign-in codes into site files.
`flea dev --lan`/`-l` enables full HTTP development on a trusted private IPv4
network. Default dev remains loopback-only. Email previews open on the host
computer; enter the code on the other device. Never forward the development port.

Public pages, partials, layouts and assets belong in site/. Use main.layout.html
as the default wrapper and *.layout.html for custom layouts. Root .flea.config.json
holds configuration and deploy.branches; never put secrets there. Editable content
lives in flea/snips.json, uploads in flea/media/, and ignored runtime in flea/runtime/.
Mail templates live in root mailers/, named *.email.html and *.email.txt with
main.layout.html. Keep public URLs site-relative; never expose private directories.

Keep the explicit `{#if app.current_admin}` blocks in the layout: a csrf-token
meta in head and `{render /admin/toolbar}` directly inside body. The reserved
current_admin object is supplied per request, never configured in JSON. Keep the
toolbar mount's version/revision attributes, declarative shadow-root shell and
external script. The visible shell renders before editor initialization; do not
replace it with an empty mount. Crossfades are
ordinary css/navigation.css imported by global.css, not runtime-injected styles.

Use `{#unless dotted.path}...{/unless}` for negative conditions. Like if, it
supports `{:else}` and `{:else if dotted.path}`; conditions are dotted paths only.

Use bare snip attributes to designate editable text, images or videos. Text
extracts to {snip ID}; an img extracts to src="{snip ID.src}" and
alt="{snip ID.alt}", a video to src and poster bindings. The generated JSON
entry is the only source of editable values. Keep classes, dimensions, links
and playback attributes in the template. Do not nest snips or use competing
srcset/picture/source elements for a media snip. Local JPEG/PNG and optimized
H.264 MP4 seeds are imported by dev. Do not invent IDs. Admins click media in
Edit mode and use the same page Save/Cancel. No Photos/Videos library buttons.

## Available skill

- **import-brochure**: import an existing small brochure website into this site,
  preserving its content and navigation with basic Prewind styling. Read
  `flea/skills/import-brochure/SKILL.md` when doing that work. This is a portable
  instruction file, not an automatically installed agent plugin.

Agent guidance stays in Git, but is excluded from public serving and deployment.
Do not put credentials in guidance or templates. Ask before adding dependencies,
backward compatibility or changing deployment/access settings.
