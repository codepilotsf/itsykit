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
    _404.html            optional standalone error document
    _500.html            optional standalone error document
    index.html           /
    css/style.css        /css/style.css
    images/mark.svg       /images/mark.svg
```

`html/` and `_/` are required. `_/snips/` is optional. The starter includes a minimal, responsive Nanoweb welcome page, an SVG logo, local CSS, a layout containing the header and footer directly, error documents, and a JSON snip for the welcome text. It loads no external fonts or services.

## Styling with Prewind

The starter bundles Prewind 1.3.1 locally and adapts Preflight's global-first
stylesheet structure. There is no CSS build step, CDN, external font request,
or Rails dependency. The entry point `html/css/style.css` declares
`@layer reset, styles, prewind` and imports:

- `theme.css`: colors, fonts, fluid spacing and text scales.
- `prewind-1.3.1.css`: the unmodified library, including its reset.
- `typography.css`: headings, prose, links, lists, code, and tables.
- `forms.css`: native controls and optional `.form-field` / `.form-row` wrappers.
- `ui.css`: buttons, including `.ui-button.ui-brand-1`.
- `application.css`: the Nanoweb starter's layout and site-specific styles.

Keep shared rules in `@layer styles` and use utilities such as `flex gap-sm`,
`mt-lg`, or `text-brand-1` for one-off adjustments. Utilities override shared
styles through the layer order. Customize the theme rather than editing the
vendored library. Form styling does not add submission processing. CSS is
embedded from `internal/cli/starter/css/` into the executable when built.

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

Pages use `_layout.html` if present. Put `{slot}` in the layout to insert the page. Error templates are standalone and do not use the layout; keep `_500.html` simple and independent of assets. Without custom error templates, plain built-in documents are used. HTTP statuses remain 404 or 500. Errors are detailed in the terminal, not in responses.

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

## Reload behavior

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
