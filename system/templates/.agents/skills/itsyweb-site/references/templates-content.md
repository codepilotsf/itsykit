# Templates and optional editable content

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
layout: /custom.layout.html
-->
<h1>Page content</h1>
```

Legacy `<!-- flea` blocks still work. Paths are site-relative `*.layout.html`
files. Only `layout` is supported; duplicate/unknown settings, nested layouts,
settings in partials and unsafe/missing layout paths fail validation.

Partials are underscore-prefixed HTML files anywhere in `site/`.
`{render footer}` includes adjacent `_footer.html`;
`{render /shared/footer}` includes `site/shared/_footer.html`. Paths are unquoted,
omit the underscore/extension, and resolve relative to the calling template,
including inside partials. Never escape `site/`. Includes inherit loop aliases;
missing/recursive includes fail even in unused partials. Preserve existing shared
markup; the starter keeps header/footer directly in its main layout.

Public config globals use `{app.site_name}`. Developer-managed
`itsy/data/home.json` uses `{data.home.title}`. Route metadata in
`itsy/snips.json` supplies `{page.title}`, `{page.description}` and `{page.path}`.
This metadata works without editable snips. Values are contextually escaped.

```html
{#if app.email}<a href="mailto:{app.email}">Email</a>
{:else if app.phone}<p>{app.phone}</p>{:else}<p>Contact us soon.</p>{/if}
{#unless app.current_admin}<a href="/admin">Admin</a>{/unless}
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
Legacy escaped named/structured snips under `itsy/snips/*.json` remain supported;
do not rewrite them simply to modernize syntax.

## Editing, media and collections

Only for requested browser editing, mark text with a bare `snip` attribute and
let `itsy dev` extract it into `itsy/snips.json`, leaving `{snip ID}`. Never
invent IDs, quote generated IDs, nest regions or put directives inside a new
region. Preserve existing IDs, stored content and labels. Removed IDs are archived;
reintroducing them restores stored content. Status/deploy never extract; commit
rewritten templates and seeds together. Rich HTML is sanitized, not unrestricted.

For requested media editing, `<img snip src="/images/person.jpg" alt="Person">`
extracts typed src/alt bindings;
`<video snip src="/videos/intro.mp4" controls></video>`
extracts src/poster bindings. Keep classes, dimensions, links, loading and playback
behavior developer-managed. No competing `picture`, `source` or `srcset`, and no
media nested in text snips. Ordinary media tags remain noneditable.

JPEG/PNG imports (up to 8 MiB/20 megapixels) are processed into immutable files
under `itsy/media/`, served at `/media/`. Photos are resized to at most 2048px
and metadata stripped; the library allows 100 records and 64 MiB of files.
Commit processed photo seeds and JSON. MP4 videos
must already be fast-start H.264 with optional AAC-LC, at most 25 MiB/file and
64 MiB retained total; no transcoding. Video bytes are Git-ignored and transferred
separately during deploy/pull/backup. Fresh clones need referenced videos restored
before validation, using an authorized pull when appropriate. No automatic pruning.

For requested repeated editing, use a page-local group with one root and named
fields, such as `<article multisnip="team"><h2 snip="name">Name</h2></article>`.
Dev extracts `{#multisnip team}` with named `{snip.name}` fields. Named image
fields work; video fields do not. No groups in layouts, partials, loops or other
groups. Preserve stable item IDs and the existing field structure.

Admins edit on-page, with explicit Save changes/Cancel for text, metadata, media
and collection order. Media-dialog Done previews a draft; page Save publishes it.
Cancel discards staged uploads. Retain dashed shadow-root overlays, keyboard focus
and reduced motion without changing content layout. Do not add separate library
buttons or a browser Theme editor. General field schemas, item/edit revision
history and adaptive video streaming remain deferred.
