# Templates, page metadata and static assets

Use the selected config and content directory described in [the skill](../SKILL.md).
Read the installed README and feature guide for version-specific syntax.

## Pages, layouts and data

`site/index.html` maps to `/`. Both `site/about.html` and
`site/about/index.html` map to `/about`; providing both is a collision. Assets
retain their public paths. `.html` URLs and non-root trailing slashes return 404
unless explicitly configured as redirects. Dot/underscore-prefixed files and
layouts are private. `_404.html` uses the main layout; `_500.html` is standalone.

`site/main.layout.html` optionally wraps pages with `{slot}`. A custom layout is
selected by one leading settings block before markup:

```html
<!-- itsy
title: Visit the bookshop
description: Opening hours and directions.
layout: /custom.layout.html
-->
<h1>Page content</h1>
```

Legacy `<!-- flea` blocks still work. Paths are site-relative `*.layout.html`
files. Supported keys are `layout`, `title` and `description`. Duplicate or
unknown keys, nested layouts, layout selection in partials and unsafe/missing
layout paths fail validation.

Partials are underscore-prefixed HTML files anywhere in `site/`.
`{render footer}` includes adjacent `_footer.html`;
`{render /shared/footer}` includes `site/shared/_footer.html`. Paths are unquoted,
omit the underscore/extension, and resolve relative to the calling template,
including inside partials. Never escape `site/`. Includes inherit loop aliases;
missing/recursive includes fail even in unused partials. Preserve existing shared
markup; the starter keeps header/footer directly in its main layout.

Public config globals use `{app.site_name}`. Developer-managed
`itsy/data/home.json` uses `{data.home.title}`. Leading page settings supply
`{page.title}` and `{page.description}`; `{page.path}` is the route. Values are
single-line plain text and contextually escaped, never compiled as templates.
Missing/empty titles use Home or the final route segment; descriptions stay empty.
Blog posts use their title/excerpt for metadata.

```html
{#if app.email}<a href="mailto:{app.email}">Email</a>
{:else if app.phone}<p>{app.phone}</p>{:else}<p>Contact us soon.</p>{/if}
<a href="/admin">Admin</a>
{#each data.home.services as service}
  <h2>{service.name}</h2>
{:else}<p>No services listed.</p>{/each}
```

Conditions are paths, not comparisons or expressions. Empty/missing collections
use the each fallback; other non-array values fail. Do not assume filters,
functions, array indexes, raw HTML or Markdown rendering. Script/style elements
and HTML comments stay inert except leading page settings. To display an exact
command literally, encode braces: `&#123;app.site_name&#125;`. Ordinary balanced
literal braces/JSON stay literal; content JSON is never compiled as templates.

## Static assets and Blog uploads

Keep ordinary images and videos in `site/images/` or `site/videos/`, using native
HTML tags, alt text and dimensions. Commit assets with their pages. Static
assets do not use Blog content synchronization or automatic extraction.

Blog's admin editor retains cover-photo uploads and its photo library. Blog
records and photo metadata use `itsy/snips.json` (only `blog` and `media`), with
photo files in `itsy/media/`. See the Blog guide for the retained schema,
limits, authentication and content sync. Preserve those records and files.
