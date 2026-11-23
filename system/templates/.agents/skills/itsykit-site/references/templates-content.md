# Templates and content

Use the selected config and content directory described in [the skill](../SKILL.md).
Read installed requirements for version-specific syntax. This reference describes
its page-CMS removal contract, not a conversion workflow for older sites.

## Pages, layouts and metadata

`site/index.html` maps to `/`. Both `site/about.html` and
`site/about/index.html` map to `/about`; providing both is a collision. Assets
retain their public paths. `.html` URLs and non-root trailing slashes return 404
unless explicitly configured as redirects. Dot/underscore-prefixed files and
layouts are private. `_404.html` uses the main layout; `_500.html` is standalone.

`site/main.layout.html` optionally wraps pages with `{slot}`. Ordinary page
metadata and an optional custom layout use one leading settings comment:

```html
<!-- itsy
title: Visit the bookshop
description: Find our opening hours and plan a visit.
layout: /custom.layout.html
-->
<h1>Visit us</h1>
```

Only `title`, `description` and `layout` are supported. Values are trimmed,
single-line plain text, split at the first colon. Braces, quotes, HTML characters
and later colons remain data, never template source; do not add YAML quoting.
Unknown, malformed and duplicate settings fail, including duplicate empty values.
The existing `<!-- flea` spelling remains supported; no new aliases are added.
Layout paths must name existing site-root-relative private `*.layout.html` files;
retain confinement and traversal checks. Layouts cannot select another layout;
partials cannot select layouts. `_500.html` remains standalone.

Layouts use `{page.title}`, `{page.description}` and `{page.path}` with contextual
escaping. A missing/empty title uses `Home` at `/`, otherwise the final route
component with hyphens replaced by spaces and its first letter capitalized.
Missing/empty description is empty. Published Blog posts use their saved title
and excerpt, overriding post-template comment metadata. There are no saved
ordinary-page metadata overrides, extraction writes or metadata editor.

## Partials, data and literal syntax

Partials are underscore-prefixed HTML files anywhere in `site/`.
`{render footer}` includes adjacent `_footer.html`;
`{render /shared/footer}` includes `site/shared/_footer.html`. Paths are unquoted,
omit the underscore/extension, and resolve relative to the calling template,
including inside partials. Never escape `site/`. Includes inherit loop aliases;
missing/recursive includes fail even in unused partials. Preserve existing shared
markup; the starter keeps header/footer directly in its main layout.

Public config globals use `{app.site_name}`. Developer-managed
`itsy/data/home.json` uses `{data.home.title}`. Values are contextually escaped.

```html
{#if app.email}<a href="mailto:{app.email}">Email</a>
{:else if app.phone}<p>{app.phone}</p>{:else}<p>Contact us soon.</p>{/if}
<a href="/admin">Admin</a>
{#each data.home.services as service}
  <h2>{service.name}</h2>
{:else}<p>No services listed.</p>{/each}
```

Conditions are dotted paths, not comparisons or expressions; `#unless` supplies
negative conditions. Empty/missing collections use the each fallback; other
non-array values fail. Do not assume filters, functions, array indexes, raw HTML
or Markdown rendering. Only exact `{blog.data.post.body}` renders sanitized rich
HTML in ordinary element content; aliases remain escaped.

Script/style elements and HTML comments stay inert except leading page settings.
To show exact template syntax in public HTML examples, encode literal braces:
`&#123;app.site_name&#125;`. Balanced ordinary literal braces/JSON stay literal;
developer JSON, metadata and Blog bodies are never recursively compiled.

## Static assets and retained Blog storage

Write page copy directly in HTML; structured lists use `itsy/data/*.json` and
`#each`. Images and videos are ordinary static files under `site/`, referenced
by public URLs. Keep classes, dimensions, links and playback behavior in source.
Styles, Prewind and theme tokens remain developer-managed and deploy with code.
Public pages have no authenticated CMS state or global page-media namespace.
Blog covers use `blog.data.post.cover` and the private admin photo list; see
[apps](apps.md) for post/photo records, upload limits and administration.

Keep the selected `itsy/snips.json` or supported legacy `flea/snips.json` for
Blog storage. Its only optional top-level fields are `blog` and `media`; `{}` is
valid empty storage. Do not cosmetically rename this file, `internal/snips` or
synchronization `Snips` wire fields: these retain active Blog uses. Rejection
guards and custom-app namespace restrictions also remain live behavior.

Page text/media snips, named/structured snips, multisnips/collections, extraction,
page editing and browser theme controls are removed. Old page-CMS JSON fields,
retired attributes/directives, named-content directories and pending extraction
journals fail closed without conversion or rewriting. Preserve a backup and
require explicitly scoped source conversion before using the incompatible runtime
on an old site. Never delete content/media or journals merely to pass validation;
no bulk migration, compatibility layer, automatic pruning or live cleanup is
implied. Consult [operations](compatibility-operations.md) for version boundaries.
