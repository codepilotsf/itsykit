# Flea CMS site requirements

Read README.md before changing templates. Preserve existing custom layouts,
content and assets. There is no frontend build step or Node requirement.
`flea new` is minimal; `--starter`/`-s` adds example pages. The shared navigation
partial drives desktop links and the native mobile popover. `flea reset` replaces
all local files/settings except Git history after confirmation and keeps a private
sibling backup. Never run it as an incidental cleanup step or deploy automatically.

Starting dev again replaces the same site's existing dev server; pull/reset
remain exclusive. Stop any preview process you start before handing control back
unless the user asks to leave it running.

New sites enable dev.email_preview in .flea.config.json. Local sign-in/contact
mail opens in the browser without Resend credentials. Preview mode and sender
details require a dev restart after changes; production ignores the toggle.
Do not write preview messages or sign-in codes into site files.
`flea dev --lan`/`-l` enables full HTTP development on a trusted private IPv4
network or shared address space (100.64.0.0/10, including Tailscale). Hostnames
must match a detected interface address, computer name or local alias at the
selected port. Default dev remains loopback-only. Email previews open on the host
computer; enter the code on the other device. Never forward the development port.

Public pages, partials, layouts and assets belong in site/. Use main.layout.html
as the default wrapper and *.layout.html for custom layouts. Root .flea.config.json
holds configuration and deploy.branches; never put secrets there. Editable content
lives in flea/snips.json, uploads in flea/media/, and ignored runtime in flea/runtime/.
Mail templates live in root mailers/, named *.email.html and *.email.txt with
main.layout.html. Keep public URLs site-relative; never expose private directories.

Optional top-level `seo.url` is the production HTTPS origin for automatic
sitemaps, robots discovery and missing response canonical tags. Use exact
`seo.exclude` page paths for crawlable noindex exclusions. Do not infer the
production domain from staging or dev, add duplicate canonical markup, or
overwrite custom robots/sitemap files during an import. Custom files require
their own sitemap discovery/exclusions. Keep server host indexing separate;
see README.md for the complete site configuration and supported server requirement.

Keep the explicit `{#if app.current_admin}` blocks in the layout: a csrf-token
meta in head and `{render /admin/toolbar}` directly inside body. The reserved
current_admin object is supplied per request, never configured in JSON. Keep the
toolbar mount's version/revision attributes, declarative shadow-root shell and
external script. The visible shell renders before editor initialization; do not
replace it with an empty mount. Crossfades are
ordinary css/navigation.css imported by global.css, not runtime-injected styles.

Use `{#unless dotted.path}...{/unless}` for negative conditions. Like if, it
supports `{:else}` and `{:else if dotted.path}`; conditions are dotted paths only.

`{csrf}` renders the complete hidden input required inside a Contact form.
Enable apps.contact with mount and email.to in .flea.config.json; optional
email.from overrides the server sender. Post to the mount, such as /contact.
Use contact.data.*, contact.invalid.*, contact.error and contact.flash.*.
Contact needs the mount's index.html and thank-you.html. No old contact config or
/_flea/contact alias exists. Restart dev after app settings change. Do not replace
admin CSRF controls. Optional Blog is also bundled. Custom managed apps are
local executables under `flea/apps/<name>/` with a validated `app.json`, enabled
by `apps.<name>.mount`; they own their sessions and private runtime storage.
Remote services, package registries, arbitrary installs and config-declared read
routes are not supported. Follow the Blog requirements below when it is enabled.

Use bare snip attributes to designate editable text, images or videos. Text
extracts to {snip ID}; an img extracts to src="{snip ID.src}" and
alt="{snip ID.alt}", a video to src and poster bindings. The generated JSON
entry is the only source of editable values. Keep classes, dimensions, links
and playback attributes in the template. Do not nest snips or use competing
srcset/picture/source elements for a media snip. Local JPEG/PNG and optimized
H.264 MP4 seeds are imported by dev. Do not invent IDs. Admins click media in
Edit mode and use the same page Save/Cancel. No Photos/Videos library buttons.

## Styling: Prewind first

Follow Prewind's global-first approach, not a parallel component CSS system.
Use existing shared defaults and theme tokens for typography, colors, links,
buttons and forms. Use supported Prewind utilities for layout, spacing, sizing,
alignment and responsive adjustments. Check the loaded Prewind version and its
documentation before choosing classes; do not assume Tailwind syntax exists.

- Before adding custom CSS, check whether existing defaults or Prewind utilities
  already express the design. Do not duplicate utilities with custom selectors.
- Do not introduce BEM naming (`block__element`, `block--modifier`), page-specific
  styling frameworks or parallel utility systems. Preserve existing custom work;
  this rule does not authorize an unrelated rewrite of existing classes.
- Custom CSS is for genuinely unsupported behavior or visual details. Before
  introducing an exception, explain what Prewind cannot reasonably express and
  obtain the user's approval. Put approved shared rules in the existing CSS
  architecture and layers; do not add an overriding layer to bypass utilities.
- Preserve useful boilerplate behavior, including sticky/fixed headers,
  translucent backgrounds, backdrop blur, mobile navigation and transitions,
  unless the user explicitly requests a change. A screenshot artifact is not
  a reason to change the actual site's behavior.
- Pass these constraints to any authorized implementation worker. Delegation
  does not relax them; the lead must review the resulting markup and CSS.

Before expanding an approved sample into the full site, review its markup and
CSS against these rules. Visual approval alone is not approval of a different
styling architecture. Before calling work complete, inspect every added or
changed CSS selector, remove utility duplication, and report any remaining
custom CSS with its reason and approval. Verify the preserved layout behavior
in the browser, not only in screenshots.

Styles are developer-managed in `site/css/theme.css` using Prewind. Ordinary
`flea deploy` publishes committed CSS. There is no browser Theme editor or CSS
synchronization; `flea pull` retrieves snips and media only.

## Optional Blog

`apps.blog.mount` enables private `_index.html` and `_post.html` templates below
the mount directory. Posts are the `blog` array in `flea/snips.json`, sharing
content/media deployment semantics. Keep stable ten-character alphanumeric IDs.
Use `blog.data.posts` and `blog.data.post` published-only views; only the exact
`{blog.data.post.body}` token renders sanitized HTML, in ordinary element content.
Drafts have no public route. The private `/admin/blog` supports explicit Publish
and Unpublish; published saves apply immediately. Preserve live posts with normal
deploy; pull before `--with-snips` source updates. Omitted IDs never delete posts.
