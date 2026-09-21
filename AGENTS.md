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

Currently implemented: `nanoweb new <site-name>`, `nanoweb status`,
`nanoweb dev [-port 3000]`, and `nanoweb setup [-source /path/to/nanoweb]`.
Run status/dev from a generated site. Setup runs from the Mac over native SSH.
The dev server listens on loopback; the persistent host defaults to public HTTP
port 80. TLS remains deferred.

README.md describes the current implementation. spec.md also contains future
plans: do not assume its editor, deployment, HTTPS, forms, schemas, revisions,
or managed images are implemented. Raw HTML rendering is not implemented.
Verify current code and tests when documentation disagrees.

## Repository map

- `cmd/nanoweb/main.go`: executable entry point.
- `cmd/nanoweb-server/main.go`: standalone persistent server entry point.
- `internal/host/`: strict server JSON config, hostname routing, readiness, and shutdown.
- `internal/setup/`: interactive SSH bootstrap, cross-build, installer, and systemd unit.
- `internal/cli/cli.go`: command handling and site creation.
- `templates/`: canonical starter files, mirroring the generated site structure.
- `templates/embed.go`: embeds HTML, CSS, SVG, JSON, and README files in the binary.
- `internal/cli/cli_test.go`: CLI tests and executable/server smoke test.
- `internal/site/project.go`: loading, validation, routes, snapshots, and rendering.
- `internal/site/template.go`: template parser, partial resolution, and data lookup.
- `internal/site/icons.go`: on-demand pinned Lucide downloads, private cache,
  geometry-only SVG validation, and icon directive compilation.
- `internal/site/icons_test.go`: offline icon resolution, escaping, and security tests.
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
  slate-800 (#1e293b) circled, slanted lowercase n SVG mark with bold rounded strokes,
  shared by the header and favicon.
- Optional partials are underscore-prefixed HTML files anywhere under `html/`.
  There is no special `_partials/` directory.
- `{render footer}` resolves `_footer.html` beside the calling template.
  `{render /foo/bar/footer}` resolves `html/foo/bar/_footer.html`.
  Paths are unquoted and omit the filename's underscore and `.html` extension.
- Relative paths, including `./` and `../`, resolve from the calling template,
  including inside partials. Paths cannot escape `html/`. Missing and recursive
  includes fail validation, including in unused partials.
- Root `_layout.html`, `_404.html`, and `_500.html` are reserved, not partials.
  `_404.html` uses the global layout; write it as page content for `{slot}`.
  `_500.html` remains standalone so error handling does not depend on the layout.
- Data uses dotted paths such as `{site.name}` and `{snip.home.title}`.
  Conditions use `{#if site.email}...{/if}`; loops use
  `{#each snip.home.items as item}...{/each}`. Includes inherit loop aliases.
- `{snip "notice"}` renders the escaped text field of a named snip. Preserve
  contextual escaping through Go's html/template; never reinterpret JSON as
  template code. Script and style elements do not support interpolation.

## Styling

### Icons

Use `{icon arrow-right}` or `{icon search size="20" class="text-dark"}`.
Names are literal lowercase kebab-case; attributes are quoted HTML-style values.
Dotted-path values inside attribute quotes inherit loop aliases and use Go's
contextual escaping. Defaults are 24px, currentColor stroke, and aria-hidden;
aria-label/aria-labelledby gives the icon role img. Supported attributes and
limitations are listed in README.md. No event handlers, style, href, or raw HTML.

Do not embed the Lucide library or download it wholesale. `LoadDev` fetches only
missing referenced icons from pinned lucide-static 1.47.0 at unpkg before serving,
including on reload. `Load`/status stays read-only and never accesses the network.
Missing icons produce deduplicated `Project.Warnings` and render nothing.
Download/network failures are also warnings; status exits 0 when otherwise valid.
Print warnings at startup/reload, not on every request. Invalid directives and
unsafe or malformed downloaded/cached SVGs remain validation errors.
Cache validated SVG files and the upstream LICENSE under `_/icons/lucide-1.47.0/`;
retain this directory in site version control. Cached icons work offline.
Downloads are bounded, cancellable, and atomically saved via confined handles.
Never fetch icons inside request handlers. Reject symlinks and reconstruct SVG
using the restricted geometry vocabulary before insertion into template source.
No site build command exists yet; a future build should reuse this resolver.
Tests use an injected HTTP transport, not live CDN requests.

### CSS

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

The starter loads Prewind 1.3.1 from unpkg, based on the Preflight template's
global-first CSS conventions. No Node build step or Rails dependency is required.
The main stylesheet is `html/css/global.css`, with layer order `reset, styles,
prewind`. It imports Prewind, theme tokens, typography, native form styles, and
button styles. Put global rules in `@layer styles` so utilities can override them.
The first line of `theme.css` imports Google Fonts: Manrope for headings and
Source Sans 3 for body text, with display=swap and system fallbacks. Customize
colors, fonts, and fluid scales in `theme.css`; site rules belong in `global.css`.
Form styling does not implement submission handling.

Keep the vendored `prewind-1.3.1.css` unchanged; customize the other files.
The theme and general styles are adapted from `../preflight/lib/templates/app/assets/stylesheets/`.
The library matches `../prewind/prewind.css` at version 1.3.1. Preserve the simple
starter with its slate circled lowercase n logo and inline header/footer when making styling changes.

## Routing and isolation

`html/index.html` maps to `/`. `html/about.html` and `html/about/index.html`
both map to `/about`; providing both is a route collision. Explicit `.html`
URLs and trailing slashes other than `/` return 404. Assets retain their paths.

Files and directories beginning with `_` or `.` are private. Preserve URL
validation, confined reads, symlink rejection, and immutable asset snapshots.
Do not replace validated route maps with request-derived filesystem reads.
Keep detailed errors in the terminal and generic/custom error pages in HTTP
responses. Invalid edits produce an error state; fixing them restores serving.
Missing icons are the exception: omit them and warn without making the site invalid.
Dev request logs use `GET /test` for 200 and `404 GET /missing` for non-200
responses, including assets and HEAD. Suppress successful internal reload polls.
Log escaped paths without query strings, using the same concurrency-safe logger
as reload diagnostics. Capture the actual HTTP status, including partial content.

## Framework versus example site

`../bookworm` is a local generated site used to try starter changes. It is not
the framework source. When the user requests a Nanoweb feature or starter
change, implement it here, even if the session begins in Bookworm.

For intentional starter changes, update the ordinary files in `templates/` as well as
the example site in scope. Do not overwrite unrelated site customizations.
Generate a fresh site in a temporary directory and compare its relevant files
with the intended result. Keep local agent instructions out of that comparison
unless they are deliberately part of the generated starter.

## Development checks

### Remote setup

Setup prompts for a hostname/IP (also user@host or an SSH alias), verifies SSH,
Ubuntu LTS 22.04/24.04/26.04 and amd64/arm64, then prints the build message.
Use the native SSH/SCP clients for agents, keys, host verification, and interactive
passwords. Do not store passwords or disable host-key checks. Bare hostnames keep
normal SSH username selection. Non-root installs use sudo with a remote PTY.
The temporary control connection and staged files are cleaned up on completion.

Cross-build cmd/nanoweb-server with GOOS=linux, detected GOARCH, CGO_ENABLED=0.
Source is found through the executable's resolved symlink or the explicit -source
flag. Go is required locally for setup, never on the target. Setup refuses existing
installation paths/accounts and occupied port 80, and does not change SSH,
firewalls, packages, or existing sites. It installs a dedicated non-login service
account, hardened systemd unit, and JSON config. Failure cleanup preserves data
and accounts. This is a first install, not an upgrade or deployment workflow.

The initial persistent host binds :80, uses /etc/nanoweb/server.json,
and has no sites by default. Config maps hostnames to absolute site roots.
It uses site.NewProductionServer: no development reload script/endpoint, watcher,
or icon fetching. Config/site edits require restart. TLS is not implemented yet.
The non-root service receives only CAP_NET_BIND_SERVICE to bind low ports.
Readiness probes wildcard listeners via loopback before setup succeeds; this
does not verify public firewall access. Administrators must allow inbound TCP 80.
Tests mock SSH/SCP, verify both Linux ELF builds, and exercise HTTP locally.
Do not claim real Ubuntu/password/systemd installation was tested without a target.

### Checks

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
