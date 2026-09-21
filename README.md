# Nanoweb

A small local website server: plain HTML, JSON content, clean URLs, and automatic browser reloads. Phase one has no dependencies outside the Go standard library. Once built, the executable needs no Go installation, Node, Docker, database, or external web server.

## Build and run

Requires Go 1.25 or newer. From this repository on the supplied Mac:

```sh
export PATH="/usr/local/go/bin:$PATH"
export GOCACHE="/tmp/nanoweb-go-cache"
go build -o bin/nanoweb ./cmd/nanoweb
export PATH="$PWD/bin:$PATH"
nanoweb new my-site
cd my-site
nanoweb status
nanoweb dev
```

Open http://127.0.0.1:3000. `nanoweb dev -port 8080` selects a different port; `-port 0` picks an available one and prints the address. The listener is IPv4 loopback only. Ctrl-C or SIGTERM shuts it down. Run `dev` and `status` from the site root.

For a persistent local installation, from the repository:

```sh
mkdir -p "$HOME/.local/bin"
/usr/local/go/bin/go build -o "$HOME/.local/bin/nanoweb" ./cmd/nanoweb
export PATH="$HOME/.local/bin:$PATH"
```

Add that PATH setting to your shell profile if desired. No Homebrew package is published by this project.

`new` accepts a simple name containing letters, digits, underscores, and hyphens, starting with a letter or digit. It refuses existing destinations, including empty directories and symlinks. An interrupted creation may leave a partial new directory for inspection. `status` validates configuration, routes, templates, and current content without listening on a port. Exit codes are 0 for success, 1 for project or runtime failures, and 2 for invalid CLI usage.

## Project structure

```text
my-site/
  _/
    nanoweb.config       JSON configuration
    snips/home.json      JSON content objects
  html/
    _layout.html         optional global wrapper
    _404.html            optional not-found page using the layout
    _500.html            optional standalone error document
    index.html           /
    css/style.css        /css/style.css
    images/mark.svg       /images/mark.svg
```

`html/` and `_/` are required. `_/snips/` is optional. The starter includes a minimal, responsive Nanoweb welcome page, an SVG logo, local theme CSS, a layout containing the header and footer directly, error documents, and a JSON snip for the welcome text. It loads Google Fonts and Prewind from external services.

## Styling with Prewind

The starter loads Prewind 1.3.1 from unpkg and adapts Preflight's global-first
stylesheet structure. There is no CSS build step or Rails dependency.
The first line of `theme.css` imports Manrope for headings and Source Sans 3
for body text from Google Fonts, with `display=swap` and system fallbacks.
The entry point `html/css/global.css` declares
`@layer reset, styles, prewind` and imports:

- `theme.css`: colors, fonts, fluid spacing and text scales.
- `https://unpkg.com/prewindcss@1.3.1`: the utility library, including its reset.
- `typography.css`: headings, prose, links, lists, code, and tables.
- `forms.css`: native controls and optional `.form-field` / `.form-row` wrappers.
- `ui.css`: buttons, including `.ui-button.ui-brand-1`.
- `global.css`: imports and site-specific styles.

Keep shared rules in `@layer styles` and use utilities such as `flex gap-sm`,
`mt-lg`, or `text-brand-1` for one-off adjustments. Utilities override shared
styles through the layer order. Customize the theme rather than editing the
vendored library. Form styling does not add submission processing. CSS is
embedded from `templates/html/css/` into the executable when built.

The full starter lives in `templates/`, mirroring a generated site's structure:
edit `templates/html/*.html` for markup, `templates/html/css/` for styles,
`templates/html/images/` for assets, and `templates/_/` for JSON defaults.
`templates/README.md` is the generated site's readme. Rebuild Nanoweb after
editing these files. They are embedded in the executable, so `nanoweb new`
works from any directory without access to this repository. The embedding
includes underscore-prefixed layouts, error pages, and private configuration.

## JSON configuration and content

`_/nanoweb.config` must contain JSON with a nonempty `site.name` string:

```json
{
  "site": {
    "name": "Nanoweb",
    "email": "hello@example.com"
  }
}
```

Other fields within `site` are available to templates. Unknown top-level settings are ignored; redirects and deploy settings have no effect. JSON comments and trailing commas are errors. Snips are top-level `*.json` files in `_/snips/`, each containing a JSON object. Names contain letters, digits, underscores, and hyphens, and start with a letter or digit. Nested snip directories are not loaded or published. Edit content with a text editor; saving by atomic rename avoids partially written JSON being observed.

## Routing

`index.html` maps to its directory: `/` at the root or `/about` for `about/index.html`. `about.html` also maps to `/about`; having both is a validation error. Page/asset route collisions are also errors. Assets keep their exact relative paths. There are no directory listings or implicit asset indexes.

URLs are case sensitive. Trailing slashes other than `/` return 404, without redirects. URLs containing `.html` (case insensitive), dot segments, doubled slashes, backslashes, NUL, or residual percent signs return 404. Standard percent encoding is decoded once by Go's HTTP parser and then validated; encoded private names are still private. Filenames that cannot produce an accepted URL fail validation.

Files and directories beginning with `_` or `.` are private, including configuration and snips. `/admin` is reserved and cannot be a page or asset. `/_nanoweb/reload` is a development-only version endpoint, not a filesystem route. Symlinks and nonregular files anywhere under `html/` or `_/` are rejected. Reads use confined directory handles with file identity checks. Requests serve only preloaded route maps and immutable asset bytes, never request-derived filesystem paths.

## Templates

Pages and `_404.html` use `_layout.html` if present. Put `{slot}` in the layout to insert the page. Write `_404.html` as page content without its own HTML document wrapper. `_500.html` stays standalone; keep it simple and independent of assets. Without custom error templates, plain built-in documents are used. HTTP statuses remain 404 or 500. Errors are detailed in the terminal, not in responses. Existing sites with a full-document `_404.html` should remove its outer HTML/head/body wrapper when upgrading.

```html
<!-- html/_layout.html -->
<!doctype html>
<html lang="en">
  <head><title>{page.title} | {site.name}</title></head>
  <body>{render header}{slot}</body>
</html>
```

Partials are underscore-prefixed HTML files anywhere under `html/`. `{render footer}` includes `_footer.html` beside the calling template. `{render /foo/bar/footer}` includes `html/foo/bar/_footer.html`. Relative directory paths, including `./` and `../`, resolve from the calling template; nested renders resolve from the partial containing them. Paths are unquoted and omit the filename's underscore and `.html`. Paths escaping `html/` are rejected. The root `_layout.html`, `_404.html`, and `_500.html` are reserved, not partials. This replaces the former `{partial "name"}` syntax and `_partials/` convention.

Missing and recursive includes fail validation, even in unused partials. Includes inherit the surrounding loop aliases. `page.path` is the canonical route; `page.title` is `Home` at `/`, otherwise the final route component with hyphens replaced by spaces and its first letter capitalized. There is no page metadata/front matter yet.

```html
<h1>{snip.home.title}</h1>
{#if site.email}<a href="mailto:{site.email}">Email us</a>{/if}
{#each snip.home.services as service}
  <article><h2>{service.name}</h2><p>{service.description}</p></article>
{/each}
```

Values and conditions accept dotted paths only. Missing values render empty. Missing/null values, false, zero, empty strings, empty arrays, and empty objects are false in conditions. Empty or missing/null collections render no loop body; other non-array collections are errors. Nested loops and alias shadowing work; the outer alias is restored afterward. `site`, `page`, and `snip` are reserved root names and remain accessible inside loops. Aliases start with a letter and contain letters, digits, or underscores.

`{snip "notice"}` renders the escaped `text` string from `_/snips/notice.json`. It requires the snip and string field to exist. Structured content uses dotted paths. Values are escaped using Go's `html/template`, including attribute and URL contexts; unsafe URL schemes become `#ZgotmplZ`. Content is never parsed again as template syntax. No raw HTML directive or trusted-HTML conversion is provided.

There are no expressions, comparisons, functions, filters, array indexes, `else`, or arbitrary code execution. Braces outside script/style elements are template directives; use `&#123;` and `&#125;` for literal text braces. Script/style bodies and their opening tags are left untouched by the Nanoweb parser, so they do not support interpolation. The internal delimiter `[[NW` is reserved and rejected in template source. Developers are trusted to author HTML, JavaScript, CSS, and public assets; snip JSON is treated as data.

## Lucide icons

Use literal lowercase Lucide names directly in templates, layouts, and partials:

```html
<a href="/">{icon arrow-left size="18"} Home</a>
{icon search class="text-dark" size="24" aria-label="Search"}
{icon check width="20" height="20" stroke-width="3"}
{#each snip.home.items as item}
  {icon info aria-label="{item.title}"}
{/each}
```

Nanoweb does **not** bundle the Lucide library. Before serving, `nanoweb dev`
downloads only referenced icons from the pinned `lucide-static@1.47.0` package
on unpkg. It stores them with the upstream license under
`_/icons/lucide-1.47.0/`. New references are resolved during template reloads.
No browser-side icon library, JavaScript, or icon network requests are required:
the server renders cached geometry as inline SVG. Keep the cache and LICENSE
in version control for reproducible, offline use. Unused cached files are retained.
There is no site build/deploy command yet; resolution currently happens in dev.

`nanoweb status` remains read-only and offline. Missing icons are warnings,
not validation failures: status exits successfully if the rest of the project
is valid. Missing icons and failed downloads render nothing in their place;
the rest of the page serves normally, with no browser error message or 500.
Dev prints one warning per unresolved icon at startup or reload, including the
template location and download failure when available. Warnings are not repeated
for every request. Malformed directives and unsafe/invalid SVGs remain errors.
A first use needs internet access; cached icons work offline. Failed downloads
do not install incomplete SVGs. Fix the name or connection and save the
template again (or restart dev) to retry. More than 256 new icons in one load
requires restarting dev to resolve the next batch. Only literal names are
supported; dynamic icon names, custom sources, and automatic version upgrades
are not supported. Icons referenced in unused partials or conditional branches
are resolved too. Script/style contents and snip strings are never scanned as icons.

Defaults: 24 by 24, `viewBox="0 0 24 24"`, no fill, `currentColor` stroke,
stroke width 2, rounded joins and caps, and `focusable="false"`.
Supported quoted attributes are `class`, `id`, `size`, `width`, `height`,
`fill`, `stroke`, `stroke-width`, `stroke-linecap`, `stroke-linejoin`, `color`,
`role`, `tabindex`, and `aria-*` / `data-*`. `size` sets both dimensions;
explicit width/height take precedence. Both single and double quotes work.
Attributes can contain escaped dotted-path values, including loop aliases.
Event handlers, inline `style`, links, and structural SVG attributes are rejected.
Use CSS classes for styling. Each icon has `lucide` and `lucide-<name>` classes.

Icons are decorative (`aria-hidden="true"`) unless given `aria-label` or
`aria-labelledby`, which defaults their role to `img`. Label an icon-only button
on the button itself. Unknown icon names are omitted with warnings; invalid name
syntax, duplicate/invalid attributes, and malformed SVGs are validation errors.
Downloaded and cached SVGs are reconstructed using
only basic geometry; scripts, external references, and arbitrary HTML cannot be
introduced through this feature. Browse names at [Lucide](https://lucide.dev/icons/).

## Reload behavior

Dev logs requests to the terminal without timestamps. Successful requests use
`GET /test`; non-200 responses include the status, such as
`404 GET /no-such-page` or `500 GET /test`. Assets and HEAD requests are logged
too. Successful internal reload polls are omitted. Logs include the escaped
route path, not query strings, so query tokens are not printed and encoded
control characters cannot create extra log lines.

The dev server hashes files under `html/` and `_/` every 500 ms, including configuration, assets, partials, and snips. Changes trigger full validation and a new in-memory snapshot. Served HTML includes a small script that checks the version every 700 ms and refreshes the browser after changes. No WebSocket, bundler, or browser extension is required. With JavaScript disabled, refresh manually. Full reloads do not preserve browser form state.

Invalid edits are logged and cause public responses to return a generic/custom 500 using the last valid error template. Fixing the source publishes a working snapshot and the browser reloads again. Requests already using an older snapshot can finish safely. Assets are held in memory, so disk edits cannot change a response halfway through serving it. A multi-file edit is not a transaction; save related changes together and allow validation to recover. Polling and full snapshots are intended for small local sites, not large media libraries.

## Checks

From the repository root:

```sh
export PATH="/usr/local/go/bin:$PATH"
export GOCACHE="/tmp/nanoweb-go-cache"
gofmt -w cmd internal
go test ./...
go test -race ./...
go vet ./...
go build -o bin/nanoweb ./cmd/nanoweb
go test ./internal/cli -run TestExecutableWorkflow -count=1 -v
```

The executable smoke test builds the binary, creates a fresh starter, runs status, starts dev on an available loopback port, checks pages/assets/private URLs, edits snips/config/assets/templates, checks invalid-edit errors and recovery, and sends an interrupt to verify clean shutdown. Tests need permission to bind a local TCP listener.

## Current limitations

This release is local development only. Browser editing and authentication, production hosting/HTTPS, deployment/rollback, remote logs/status, forms, snip schemas, managed content writes and revisions, and managed snip image publishing are deferred. Put developer-managed images under `html/images/` and reference their public URLs in JSON. `_/snips/` never becomes a public image directory. Raw HTML remains deferred until there is a schema-approved, sanitized design. JSON parsing validates syntax and object shape, not content schemas. No production release or security review is implied by the broader roadmap in spec.md.
