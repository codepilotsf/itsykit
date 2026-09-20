# Working on Nanoweb

## Keep this guide current

Update this file as part of any change that affects Nanoweb's architecture,
template language, CLI, starter conventions, supported features, or development
workflow. Update the relevant sections in the same task, remove stale guidance,
and keep this a practical guide rather than a running changelog. No separate
reminder from the user is needed. Keep README.md and applicable sections of
spec.md consistent with behavior changes too.

## Product and scope

Nanoweb is a small website server built with Go 1.25 or newer and the standard
library. Sites use plain HTML, local assets, and JSON content. Keep the design
simple and avoid adding dependencies or a frontend build system without a
concrete need.

Currently implemented: `nanoweb new <site-name>`, `nanoweb status`, and
`nanoweb dev [-port 3000]`, with local automatic browser reloads. Run status
and dev from a generated site's root. The server listens on IPv4 loopback.

README.md describes the current implementation. spec.md also contains future
plans: do not assume its editor, deployment, HTTPS, forms, schemas, revisions,
or managed images are implemented. Raw HTML rendering is not implemented.
Verify current code and tests when documentation disagrees.

## Repository map

- `cmd/nanoweb/main.go`: executable entry point.
- `internal/cli/cli.go`: command handling and site creation.
- `internal/cli/starter.go`: canonical generated site files, embedded as Go strings.
- `internal/cli/starter/css/`: starter stylesheets embedded using go:embed.
- `internal/cli/cli_test.go`: CLI tests and executable/server smoke test.
- `internal/site/project.go`: loading, validation, routes, snapshots, and rendering.
- `internal/site/template.go`: template parser, partial resolution, and data lookup.
- `internal/site/server.go`: HTTP serving, reload watching, and URL validation.
- `internal/site/site_test.go`: rendering, routing, isolation, and reload tests.
- `bin/nanoweb`: local build output; rebuilding is necessary for CLI changes to take effect.

## Sites and templates

A generated site has `html/` for templates/assets and `_/` for configuration
and content. `_/nanoweb.config` is JSON and requires a nonempty `site.name`.
Snips are JSON objects in `_/snips/*.json`.

- `html/_layout.html` is the optional global wrapper; `{slot}` inserts the page.
- The starter's header and footer live directly in `_layout.html`. Do not
  generate separate `_header.html` or `_footer.html` files by default.
- The starter is a minimal Nanoweb welcome page, with local CSS and a simple
  folded, two-tone violet SVG mark shared by the header and favicon.
- Optional partials are underscore-prefixed HTML files anywhere under `html/`.
  There is no special `_partials/` directory.
- `{render footer}` resolves `_footer.html` beside the calling template.
  `{render /foo/bar/footer}` resolves `html/foo/bar/_footer.html`.
  Paths are unquoted and omit the filename's underscore and `.html` extension.
- Relative paths, including `./` and `../`, resolve from the calling template,
  including inside partials. Paths cannot escape `html/`. Missing and recursive
  includes fail validation, including in unused partials.
- Root `_layout.html`, `_404.html`, and `_500.html` are reserved, not partials.
  Error documents render independently of the layout.
- Data uses dotted paths such as `{site.name}` and `{snip.home.title}`.
  Conditions use `{#if site.email}...{/if}`; loops use
  `{#each snip.home.items as item}...{/each}`. Includes inherit loop aliases.
- `{snip "notice"}` renders the escaped text field of a named snip. Preserve
  contextual escaping through Go's html/template; never reinterpret JSON as
  template code. Script and style elements do not support interpolation.

## Styling

Default to the Prewind method; read `../prewind/README.md` for its philosophy
and supported classes. Keep HTML readable:

- Global defaults (typography, links, forms) belong in shared CSS.
- Use Prewind classes for one-off layout, spacing, and color adjustments.
  Prefer theme tokens and fluid sizes such as `gap-md` and `p-lg`; do not
  invent custom classes that merely duplicate utilities.
- When styled markup repeats, consider a partial to reuse the markup and
  its utility classes. Keep the starter header/footer inline as specified above.
- Occasionally use inline styles for genuinely arbitrary one-off values,
  such as `style="width: 37px"`; do not invent Tailwind bracket utilities.
- Write shared CSS for behavior Prewind does not cover, such as complex grids
  and animations. Do not assume Tailwind classes or variants exist in Prewind.

Use Prewind's `container` class for inner page content by default, giving it
consistent centered width and theme-controlled padding. For full-width sections,
put `container` on an inner wrapper. Pair it with `max-w-full` for screens
narrower than 360px; avoid custom width wrappers that duplicate its purpose.

The starter bundles Prewind 1.3.1 locally, based on the Preflight template's
global-first CSS conventions. No Node build step, CDN, external fonts, or Rails
dependencies are required. The main stylesheet is `html/css/style.css`, with
layer order `reset, styles, prewind`. It imports the vendored Prewind reset and
utilities, theme tokens, typography, native form styles, button styles, and
application rules. Put global rules in `@layer styles` so utilities can override
them. Customize colors and fluid scales in `theme.css`; site rules belong in
`application.css`. Form styling does not implement submission handling.

Keep the vendored `prewind-1.3.1.css` unchanged; customize the other files.
The theme and general styles are adapted from `../preflight/lib/templates/app/assets/stylesheets/`.
The library matches `../prewind/prewind.css` at version 1.3.1. Preserve the simple
violet starter and inline header/footer when making styling changes.

## Routing and isolation

`html/index.html` maps to `/`. `html/about.html` and `html/about/index.html`
both map to `/about`; providing both is a route collision. Explicit `.html`
URLs and trailing slashes other than `/` return 404. Assets retain their paths.

Files and directories beginning with `_` or `.` are private. Preserve URL
validation, confined reads, symlink rejection, and immutable asset snapshots.
Do not replace validated route maps with request-derived filesystem reads.
Keep detailed errors in the terminal and generic/custom error pages in HTTP
responses. Invalid edits produce an error state; fixing them restores serving.

## Framework versus example site

`../bookworm` is a local generated site used to try starter changes. It is not
the framework source. When the user requests a Nanoweb feature or starter
change, implement it here, even if the session begins in Bookworm.

For intentional starter changes, update `internal/cli/starter.go` as well as
the example site in scope. Do not overwrite unrelated site customizations.
Generate a fresh site in a temporary directory and compare its relevant files
with the intended result. Keep local agent instructions out of that comparison
unless they are deliberately part of the generated starter.

## Development checks

Inspect the current worktree before editing and preserve unrelated changes.
Format changed Go files with gofmt. For implementation changes, run:

```sh
go test ./...
go vet ./...
go build -o bin/nanoweb ./cmd/nanoweb
```

Use `go test -race ./...` for concurrency, reload, or shared-state changes.
The regular suite includes `TestExecutableWorkflow`, which creates a fresh
site, starts a local server, checks edits/recovery, and shuts it down. It needs
loopback binding permission; a sandbox bind failure is not a product failure.
Add meaningful regression coverage for changed behavior. Documentation-only
changes do not need the runtime suite.

On the current Mac, Go is available at `/usr/local/go/bin/go`. If the default
build cache is restricted, use `GOCACHE=/tmp/nanoweb-go-cache`. The installed
`/usr/local/bin/nanoweb` currently links to this repo's `bin/nanoweb`; verify
the link before relying on it. Rebuild after implementation or starter changes,
then run `nanoweb status` in the example/generated site. Restart an existing
dev process when the executable changes; site-file reloads do not reload Go code.

## Communication

Use a natural voice. Avoid AI cliches, emojis, and em dashes. Question assumptions
when warranted. State which repository changed, what was checked, and any
unfinished work. Do not claim a starter change is complete based only on edits
to Bookworm.
