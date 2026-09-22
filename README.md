# Nanoweb

Nanoweb is a small website server for people who like working with HTML and CSS.
Write your pages, keep reusable content in JSON, and preview changes as you save.
When you're ready to publish, deploy from your Mac to an Ubuntu server over SSH.
Nanoweb serves the site and handles HTTPS too.

It's built for small sites: a business website, a portfolio, or a few pages for
your next project. Your site stays in ordinary files you can open in any editor
and track in Git. You don't need Node, Docker, a database, or a separate web server.

## What you get

- Plain HTML pages with clean URLs, shared layouts, and reusable partials.
- JSON content, with simple template tags for values, conditions, and lists.
- A local preview server that reloads your browser when you save.
- A responsive starter with themeable CSS, Prewind utilities, and Lucide icons on demand.
- Git-based deployment, with a separate saved destination for each branch.
- An Ubuntu server that can host multiple sites, with automatic HTTPS and WWW redirects.

Nanoweb is still in development. The local workflow is available from source;
release publishing and real-server verification are still pending. There is no
form submission handler yet. The owner dashboard supports email-code login and
basic content editing. See [current limitations](#current-limitations)
for the boundaries of what's implemented.

## Quick start

Working on NanoWeb itself? Start with the [architecture and review guide](ARCHITECTURE.md)
for package responsibilities, content transactions, and remaining review work.

If `nanoweb` is already installed, you're three commands away:

```sh
nanoweb new my-site
cd my-site
nanoweb dev
```

Open http://127.0.0.1:3000, edit `index.html`, and save. Your browser reloads
automatically. Change the welcome text in `_/snips.json`, or give the site
your own colors and fonts in `css/theme.css`. Press Ctrl-C to stop the server.

Need the executable first? With Go 1.26 or newer installed, run these commands
from this repository:

```sh
bash scripts/build.sh
export PATH="$PWD/bin:$PATH"
```

Then follow the three commands above. This adds Nanoweb to your current shell's
PATH. The build script includes both Linux servers, so setup works without
publishing a release or downloading a server. No Homebrew package is
published by this project yet.

## Everyday commands

Run these from your site's folder:

| Command | What it does |
| --- | --- |
| `nanoweb dev` | Preview your site at http://127.0.0.1:3000 with browser reloads. |
| `nanoweb dev -port 8080` | Preview on a different port. |
| `nanoweb status` | Check configuration, routes, templates, and content without starting a server. |
| `nanoweb deploy` | Configure the current branch's destination, then publish its committed files. |
| `nanoweb deploy -configure` | Change that branch's saved deployment settings and deploy. |
| `nanoweb https` | Enable HTTPS for an already deployed site. |
| `nanoweb version` | Show the installed version. |

The preview server is local to your machine. Use `-port 0` to choose an available
port automatically. `status` works offline; missing icons produce warnings while
otherwise valid sites still pass.

`nanoweb new my-site` creates a Git repository on `main`, ready for your first
commit. Use `nanoweb new my-site --skip-git` if the site belongs in an existing
repository. Git isn't required for local previews, but it is required to deploy.

Start with [project structure](#project-structure) and [templates](#templates)
below. When you're ready to publish, see [deploy a site](#deploy-a-site) and
[server setup](#set-up-an-ubuntu-server-from-your-mac).

## Project structure

```text
my-site/
  _/
    nanoweb.config       JSON configuration
    snips.json           Rich content and page metadata
  _layout.html           optional global wrapper
  _404.html              optional not-found page using the layout
  _500.html              optional standalone error document
  index.html             /
  css/global.css         /css/global.css
  images/mark.svg         /images/mark.svg
```

Start with `index.html` for your home page. Add `about.html` and it's available
at `/about`. Put shared page markup in `_layout.html`, styles in `css/`, and
images in `images/`.

Every site needs `_/nanoweb.config`; JSON snips are optional. The starter gives
you a welcome page, a layout with a header and footer, error pages, an SVG logo,
and CSS you can make your own. It loads Google Fonts and Prewind from external
services.

The project root is the document root. Dotfiles/directories, `node_modules`,
`README.md`, `AGENTS.md`, `ARCHITECTURE.md`, `spec.md`, `next-steps.md`, `LICENSE`, `LICENSE.md`,
`go.mod`, and `go.sum` are excluded from serving, reloads, and deployment
(these metadata names are case-insensitive). Keep other development files or
secrets in dot-prefixed directories or private `_/` files. Only configuration,
snips, and icon caches inside `_/` are loaded and deployed.

Existing sites must move the contents of `html/` into the project root, preserving
their relative paths, then remove the empty `html/` directory. Check for name
collisions before moving files. There is no legacy `html/` document-root fallback.
Update the server binary before deploying this new layout; older binaries expect
`html/`. Nanoweb does not move existing projects automatically.

## Templates

Write the shared HTML once in `_layout.html`, then use `{slot}` where each
page's content should appear. A page only needs the markup that goes inside
that slot. If you don't need a shared layout, leave it out and write complete
HTML pages instead.

```html
<!-- _layout.html -->
<!doctype html>
<html lang="en">
  <head><title>{page.title} | {site.name}</title></head>
  <body>
    <header><a href="/">{site.name}</a></header>
    {slot}
  </body>
</html>
```

The not-found page, `_404.html`, uses the layout too, so write it as page content.
The server-error page, `_500.html`, is a standalone HTML document; keep it simple
so it can work even when the rest of the site has a problem. Both are optional.
Nanoweb supplies basic error pages if you leave them out, and keeps detailed
errors in the terminal.

Partials are underscore-prefixed HTML files anywhere under the project root. `{render footer}` includes `_footer.html` beside the calling template. `{render /foo/bar/footer}` includes `foo/bar/_footer.html`. Relative directory paths, including `./` and `../`, resolve from the calling template; nested renders resolve from the partial containing them. Paths are unquoted and omit the filename's underscore and `.html`. Paths escaping the project root are rejected. The root `_layout.html`, `_404.html`, and `_500.html` are reserved, not partials. This replaces the former `{partial "name"}` syntax and `_partials/` convention.

Missing and recursive includes fail validation, even in unused partials. Includes inherit the surrounding loop aliases. `page.path` is the canonical route. `page.title` and `page.description` come from that route's metadata in `_/snips.json`. Without a title, the default is `Home` at `/`, otherwise the final route component with hyphens replaced by spaces and its first letter capitalized. Description defaults to empty. Metadata is contextually escaped.

```html
<h1>{snip.home.title}</h1>
{#if site.email}
  <a href="mailto:{site.email}">Email us</a>
{:else if site.phone}
  <p>{site.phone}</p>
{:else}
  <p>Contact details coming soon.</p>
{/if}
{#each snip.home.services as service}
  <article><h2>{service.name}</h2><p>{service.description}</p></article>
{/each}
```

Values and conditions accept dotted paths only. Missing values render empty. Missing/null values, false, zero, empty strings, empty arrays, and empty objects are false in conditions. Empty or missing/null collections render no loop body; other non-array collections are errors. Nested loops and alias shadowing work; the outer alias is restored afterward. `site`, `page`, `data`, and legacy `snip` are reserved root names and remain accessible inside loops. Aliases start with a letter and contain letters, digits, or underscores.

`{snip a7Km2pQ9xB}` renders sanitized rich HTML from `_/snips.json`, only in HTML element content. Ordinary dotted-path values remain contextually escaped, including attributes and URLs. Legacy `{snip notice}` still reads the escaped `text` field from `_/snips/notice.json`; legacy structured snips remain supported during migration. JSON content is never parsed as template syntax. There is no unrestricted raw-HTML directive.

There are no expressions, comparisons, functions, filters, array indexes, or arbitrary code execution. Braces outside script/style elements are template directives; use `&#123;` and `&#125;` for literal text braces. Script/style bodies and their opening tags are left untouched by the Nanoweb parser, so they do not support interpolation. The internal delimiter `[[NW` is reserved and rejected in template source. Developers are trusted to author HTML, JavaScript, CSS, and public assets; snip JSON is treated as data.

Loops can include an empty-list fallback:

```html
{#each snip.team.members as member}
  <p>{member.name}</p>
{:else}
  <p>No members yet.</p>
{/each}
```

The fallback also runs for missing/null arrays. It has no current item; any
outer alias with the same name is restored. `{:else if dotted.path}` is only
valid in conditions. A final `{:else}` cannot be followed by another branch.

## JSON configuration and content

### Editable snip regions

While `nanoweb dev` is running, add an editable region to a template:

```html
<h1 snip="title">Welcome to <em>NanoWeb</em></h1>
```

On save, dev extracts the inner HTML and rewrites only that region:

```html
<h1 snip="title">{snip a7Km2pQ9xB}</h1>
```

IDs are cryptographically random, case-sensitive, 10-character alphanumeric
strings, checked against active and archived local IDs. IDs never contain braces,
quotes or whitespace. Use `{snip ID}` without quotes. Quoted markers are rejected;
HTML comments do not render snips. Labels are descriptions,
not identities. Reusing an ID intentionally shares content. IDs referenced from
multiple routes, layouts, or partials are grouped under `shared`.

Private `_/snips.json` stores rich content and page metadata:

```json
{
  "pages": {
    "/": {
      "title": "Home",
      "description": "Welcome to our site.",
      "snips": {
        "a7Km2pQ9xB": {
          "label": "title",
          "content": "Welcome to <em>NanoWeb</em>"
        }
      }
    }
  },
  "shared": {"snips": {}}
}
```

Use a `div` or `section` region for multiple paragraphs. Formatting includes
paragraphs, emphasis, lists, headings h2–h6, blockquotes, code, and safe links.
Scripts, styles, event handlers, images, forms, embedded content, SVG/MathML,
and arbitrary attributes are removed. Allowed link schemes are HTTP, HTTPS,
mailto, tel, and relative URLs; protocol-relative links are removed. Rich snips
cannot appear in attributes, script/style interpolation, textarea, title, or
select contexts. Ordinary `{data.name.field}` values remain escaped text.

New regions must have explicit balanced closing tags, cannot nest, and cannot
contain NanoWeb directives. Use HTML entities for literal braces during initial
extraction. Existing JSON content can contain braces: it is never template code.
Snip files are limited to 5 MiB and each rich content value to 1 MiB.

Extraction happens only in dev, including startup. Status is read-only. Deploy
rejects unextracted regions and asks you to run dev and commit the resulting
files. Every relevant save rescans all templates, extracts new regions, updates
labels and route ownership, and moves unreferenced IDs to `archive`. Undoing a
removal restores archived content by ID. Existing referenced content and page
metadata are preserved. Dev prints `Snips updated in _/snips.json.` when it writes
changes. Missing IDs and malformed regions are reported in the terminal.

Local writes are coordinated in memory, with no lock file. Run only one dev
process per project. Stop older dev processes before updating; the new version
removes an old empty `_/.snips.lock` file when it next scans the project.
Server-side deployment/content locks are unchanged.
A `_/.snips-pending.json` recovery record saves the content and
before/after templates before any replacement. Dev resumes interrupted writes;
if files were independently edited, it stops rather than overwriting them.
Reconcile those files with the recovery record and restart dev. Do not delete
the recovery record without preserving its content.

Developer-managed structured JSON belongs in `_/data/*.json`, available as
`{data.team.members}` and in `{#each data.team.members as member}` loops. It is
deployed with code, not treated as owner-editable rich content. Legacy
`_/snips/*.json` remains supported; new sites no longer generate that directory.

### Site settings

`_/nanoweb.config` must contain JSON with a nonempty `site.name` string:

```json
{
  "site": {
    "name": "Nanoweb",
    "email": "hello@example.com"
  }
}
```

Other fields within `site` are available to templates. `admin.emails` controls owner
access as described below. Other unknown top-level settings are ignored; redirects
and deploy settings have no effect here. JSON comments and trailing commas are errors.
Rich snips live in `_/snips.json`; structured data lives in `_/data/*.json`.
Legacy named snippets in `_/snips/*.json` remain supported but are not browser-editable.

## Owner dashboard

From your site's folder, run `nanoweb admin` and enter an email, or provide it directly:

```sh
nanoweb admin owner@example.com
```

This creates or appends to `admin.emails` in `_/nanoweb.config`, preserves other
settings, and skips addresses already present (case-insensitive). It accepts up
to 20 admins and does not send email, commit, or deploy. Restart dev to apply
local access changes; commit and deploy to update the live site. Email delivery
still requires the Resend settings below. You can also edit the JSON directly:

```json
{
  "site": {"name": "My site"},
  "admin": {"emails": ["owner@example.com", "editor@example.com"]}
}
```

The starter has an empty list, so admin access is disabled until you configure it.
There is no registration, password, or user-management screen. All listed addresses
have the same editing access. Up to 20 unique addresses are supported; matching is
case-insensitive. Only the developer can change the list.

Visit `https://your-domain/admin`. Enter an approved email, then its six-digit code.
Codes expire after 10 minutes, allow five attempts, and work once. Sessions last up
to eight hours and end on logout or server restart/deployment. Cookies are host-only,
HttpOnly, Secure in production, and SameSite=Strict. POST requests require the same
Origin and a CSRF token. Login requests have per-address, client-IP and per-site
limits (3, 10 and 40 per 15 minutes). Unlisted addresses receive the same generic
response but no email. Forwarded IP/TLS headers are not trusted.

The dashboard lists public pages, titles, descriptions, and shared/page snips.
Edit content as plain text or simple HTML in a textarea. This first editor is not
a visual rich-text toolbar. The existing HTML allowlist removes unsupported markup
on save. Page details remain escaped. Saving validates the site, atomically writes
the managed content file, and updates the live serving snapshot without restart.
Stale forms are rejected instead of overwriting newer edits. Archived snips,
templates, access settings, deployment settings, and structured JSON are not editable.

### Email credentials

Resend is the only implemented provider. The code uses its
[send-email API](https://resend.com/docs/api-reference/emails/send-email) through a
small provider-neutral `mailer.Sender` interface, without adding a provider SDK.
Use a sending API key and a sender address from a verified domain in your Resend account.

On the first deployment of an admin-enabled site, NanoWeb checks the server's mail
settings. If missing, it prompts for a hidden API key and sender email. Subsequent
deployments reuse them. Settings are shared across sites on that server and stored
in `/etc/nanoweb/mail.env` (root-owned, mode 0600), outside site archives and Git.
systemd reads that file and passes the values to the unprivileged service. The key
travels over SSH stdin to a private temporary file, never as a command argument.
The temporary file is removed on normal completion; an interrupted connection may
leave a mode-0600 file in its mode-0700 `/tmp/nanoweb-mail.*` directory.

For local development, export `RESEND_API_KEY` and `NANOWEB_EMAIL_FROM` before
running `nanoweb dev`, then open `http://127.0.0.1:3000/admin`. The same email-code
flow applies; there is no authentication bypass and codes are never printed.
For a private local `.env`, use these names and load it yourself with `set -a;
source .env; set +a` in your shell. NanoWeb does not interpret dotenv files locally.
The starter ignores `.env` files. Never place keys in `_/nanoweb.config` or commit them.
Restart dev after changing approved emails or environment variables.

To rotate server credentials, edit `/etc/nanoweb/mail.env` as root, keeping the
following two quoted assignments and mode 0600, then restart `nanoweb.service`:

```dotenv
RESEND_API_KEY="re_your_key_here"
NANOWEB_EMAIL_FROM="login@your-verified-domain.com"
```

Production admin access requires HTTPS. An admin-enabled HTTP-only deployment is
refused. Protocol 7 requires both the updated server binary and systemd unit, a
service-owned `/etc/nanoweb/deploy.lock`, and writable managed content directories.
Fresh setup installs these. Existing servers need an administrator-reviewed upgrade;
do not use setup's destructive reset to upgrade a server with sites. Existing
content directories must be owned by `nanoweb:nanoweb` (0750); release directories
and host config remain root-owned. Redeployment prepares the new content directory.
Back up your server before changing installation permissions.

## Styling with Prewind

Start with `css/theme.css` to change colors, fonts, and spacing. Put your site's
shared styles in `css/global.css`, and use Prewind utilities for adjustments
in individual pages. There is no CSS build step.

The starter loads Prewind 1.3.1 from unpkg and uses a global-first stylesheet
structure adapted from Preflight.
The first line of `theme.css` imports Manrope for headings and Source Sans 3
for body text from Google Fonts, with `display=swap` and system fallbacks.
The entry point `css/global.css` declares
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
embedded from `templates/css/` into the executable when built.

The full starter lives in `templates/`, mirroring a generated site's structure:
edit `templates/*.html` for markup, `templates/css/` for styles,
`templates/images/` for assets, and `templates/_/` for JSON defaults.
`templates/README.md` is the generated site's readme. Rebuild Nanoweb after
editing these files. They are embedded in the executable, so `nanoweb new`
works from any directory without access to this repository. The embedding
includes underscore-prefixed layouts, error pages, and private configuration.

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

Before serving, `nanoweb dev` downloads the icons you've used from the pinned
`lucide-static@1.47.0` package
on unpkg. It stores them with the upstream license under
`_/icons/lucide-1.47.0/`. New references are resolved during template reloads.
No browser-side icon library, JavaScript, or icon network requests are required:
the server renders cached geometry as inline SVG. Keep the cache and LICENSE
in version control for reproducible, offline use. Unused cached files are retained.
Deploy uses the same resolver before packaging; no standalone site build command exists.

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

## Routing

`index.html` maps to its directory: `/` at the root or `/about` for `about/index.html`. `about.html` also maps to `/about`; having both is a validation error. Page/asset route collisions are also errors. Assets keep their exact relative paths. There are no directory listings or implicit asset indexes.

URLs are case sensitive. Trailing slashes other than `/` return 404, without redirects. URLs containing `.html` (case insensitive), dot segments, doubled slashes, backslashes, NUL, or residual percent signs return 404. Standard percent encoding is decoded once by Go's HTTP parser and then validated; encoded private names are still private. Filenames that cannot produce an accepted URL fail validation.

Files and directories beginning with `_` or `.` are private, including configuration and snips. `/admin` is reserved and cannot be a page or asset. `/_nanoweb/reload` is a development-only version endpoint, not a filesystem route. Symlinks and nonregular files in the included site tree are rejected. Reads use confined directory handles with file identity checks. Requests serve only preloaded route maps and immutable asset bytes, never request-derived filesystem paths.

## Reload behavior

Dev logs requests to the terminal without timestamps. Successful requests use
`GET /test`; non-200 responses include the status, such as
`404 GET /no-such-page` or `500 GET /test`. Assets and HEAD requests are logged
too. Successful internal reload polls are omitted. Logs include the escaped
route path, not query strings, so query tokens are not printed and encoded
control characters cannot create extra log lines.

The dev server hashes files under the project root and `_/` every 500 ms, including configuration, assets, partials, and snips. Changes trigger full validation and a new in-memory snapshot. Served HTML includes a small script that checks the version every 700 ms and refreshes the browser after changes. No WebSocket, bundler, or browser extension is required. With JavaScript disabled, refresh manually. Full reloads do not preserve browser form state.

Invalid edits are logged and cause public responses to return a generic/custom 500 using the last valid error template. Fixing the source publishes a working snapshot and the browser reloads again. Requests already using an older snapshot can finish safely. Assets are held in memory, so disk edits cannot change a response halfway through serving it. A multi-file edit is not a transaction; save related changes together and allow validation to recover. Polling and full snapshots are intended for small local sites, not large media libraries.

## Deploy a site

You'll need Git, SSH access to a supported Ubuntu server, and a domain pointing
at that server (or an IP address for an HTTP-only site). For HTTPS, allow inbound
TCP ports 80 and 443. Use a complete release or build with `bash scripts/build.sh`
to include the bundled servers needed by setup.

From your site's folder, on the branch you want to publish, run:

```sh
nanoweb deploy
```

On first use for each branch, Nanoweb asks for the target server, SSH username
(default `root`), domain/IP,
WWW-to-primary redirect, and HTTPS (default yes). Deployments use the actual
branch name, with no separate environment label or question. The internal site
name is generated automatically from the project folder and branch. A new
`staging` branch suggests `staging.example.com` using the saved main mapping,
but requires approval. Staging defaults to no WWW alias; main defaults
to redirecting WWW. An IP destination uses HTTP without a WWW alias.
HTTPS uses one combined question: `Enable HTTPS and accept Let's Encrypt terms?`
(default yes), with the terms link displayed first. It then asks for a contact
email. The saved choice is reused on future deploys.
Supplying `user@host` skips the username question. Bare targets saved by older
versions ask for the username once and save it; commit that config change before
retrying. Deployment never silently chooses your Mac username.

After connecting, deploy checks whether Nanoweb is installed. If it is missing,
it offers to run setup on that same server, then continues deploying. The existing
SSH connection is reused, including the same verified bundled-server extraction.
Failed SSH authentication does not trigger setup,
and an older or incomplete installation requires an update rather than reinstall.

Selections are saved in `_/deploy.json` before attempting deployment:

```json
{
  "branches": {
    "main": {
      "target": "ubuntu@1.2.3.4",
      "domain": "example.com",
      "name": "my-site-main",
      "www_redirect": true,
      "https": true,
      "email": "you@example.com",
      "agree_tos": true
    },
    "staging": {
      "target": "ubuntu@1.2.3.4",
      "domain": "staging.example.com",
      "name": "my-site-staging",
      "www_redirect": false,
      "https": true,
      "email": "you@example.com",
      "agree_tos": true
    }
  }
}
```

Nanoweb offers to commit pending changes, including the new configuration, then
continues deploying. It lists every changed file, asks for confirmation and a
commit message (`init` for the first commit; otherwise required with no default), and uses normal Git
hooks and signing settings. Declining or a failed commit stops deployment.
All listed repository changes are staged, not just public site files. Sites
inside a larger repository require a manual commit to avoid staging unrelated
work. Changes made later during deployment still stop it. Future deploys
use only the current branch's saved destination and print the branch/commit.
An unmapped branch prompts instead of falling back to main. Detached HEAD,
repositories without a commit, and non-Git folders cannot deploy. The wizard can
save settings before the first commit, so you can commit the site and deployment
settings together. Upload still requires that commit and a clean status. Settings are
never uploaded and contain no credentials. You can explicitly ignore the file
if you prefer local-only settings. Otherwise commit and share the branch map
across the branches that should know about it.

Use `nanoweb deploy -configure` to change a saved mapping. Flags such as
`-target`, `-domain`, `-name`, `-https`, `-www-redirect`, `-email`, `-agree-tos`,
and `-expected-ips` seed the questions on first use or with `-configure`.
They cannot silently override an existing branch destination. Flat settings from
the previous version are offered for confirmation and migrated to a branch map.
Different branches cannot share a hostname or the same remote site name on a
target. The server also records and enforces site branch ownership.

Git must be clean, including staged, unstaged, and untracked files, before SSH.
The repository is checked again before upload and activation. The snapshot is
read from the selected commit's Git objects, never working files, so ignored
assets and uncommitted edits cannot slip into a release. Missing icons are
resolved into the temporary committed snapshot without changing the checkout.
Sites in a repository subdirectory are supported; cleanliness covers the whole
repository. Git and a valid committed site are required on the deploying Mac.

Before uploading, the target server resolves every configured hostname (including
WWW when enabled). Lookup failures, no A/AAAA records, or any address that does
not match the target fail deployment. Both IPv4 and IPv6 answers must match.
By default the expected addresses come from server interfaces. For NAT or a
proxy/CDN, set `expected_ips` to the explicitly approved public addresses in
the config or with `-configure -expected-ips`. This advanced setting is not asked
about in the normal wizard. IP destinations skip DNS lookup.
This checks the server's resolver; it does not prove public DNS propagation or
that firewall ports are open. DNS names must point at the server for ACME.

Deploy resolves referenced icons in a temporary snapshot, validates it, and
uploads it over native SSH/SCP. Only root-level site files, `_/nanoweb.config`,
`_/snips.json`, `_/data/`, legacy `_/snips/`, and
`_/icons/` are included. Repository files and other private workspace files are
excluded. Archives are limited to 100 MiB uncompressed and 10,000 entries;
symlinks, hard links, special files, duplicates, and traversal paths are rejected.

On the server, each release gets a new directory:

```text
/var/lib/nanoweb/sites/my-site/releases/<timestamp-random-id>/
```

The installed server validates the extracted site, atomically replaces its JSON
host configuration, restarts the service, and verifies the active revision and
hostname. All existing aliases for the same site move together; unrelated sites
remain configured. A hostname already assigned to another site is refused.
Omitting a previous alias does not delete it. Changing aliases/removing sites is
currently a manual host configuration task.

There is a brief interruption while the single service restarts. If activation
or certificate verification fails, Nanoweb restores the previous configuration
and restarts it. Concurrent operations are locked. Prior releases are retained,
and the previous successful configuration is saved as
`/etc/nanoweb/server.previous.json`. Failed staged releases are retained too;
automatic pruning and a user-facing rollback command remain future work.

An interrupted SSH session can leave an unfinished transaction. Inspect the
service and `/etc/nanoweb/deploy.pending.json` on the server. If the journal
exists, restore the previous configuration with:

```sh
sudo /usr/local/bin/nanoweb-server -recover
```

Further deployments refuse to proceed until recovery succeeds. Do not delete
the recovery journal to bypass that check.

Rich snip deployment preserves remote content by ID and existing page metadata.
Local content seeds only missing IDs. Code and structured data are still replaced.
Content lives outside code releases at
`/var/lib/nanoweb/sites/<site>/content/<release-id>.json`; the server config's
`content` map selects the active snapshot for each hostname. Aliases pointing to
the same release must select the same content snapshot; conflicting paths are
rejected. Valid aliases share one loaded site snapshot. The existing lock
and recovery journal switch code and content together. Failed activation leaves
the previous content snapshot untouched.

After successful activation, unreferenced IDs are removed from active page/shared
snip maps and retained in the content snapshot's `archive` map. References are
scanned from template source, including false branches, empty loops, and unused
partials. Reintroducing an archived ID restores its last remote content, not a
stale local seed. Archives and old snapshots are not automatically deleted.
Page metadata is retained even when a page disappears. There is no rollback command
yet. Browser saves share the deployment lock and update the live snapshot. Manual
content edits require a service restart and must not run concurrently with deployment.

## Enable HTTPS

Deployment settings are saved without a final save-confirmation prompt. Deployment
shows the branch and domain with a Ctrl-C cancellation notice, and prints the
site URL after successful activation. A `nanoweb rollback` command is not yet
available.

After deploying a domain, run:

```sh
nanoweb https
```

The first-deploy wizard enables HTTPS by default, so you normally won't need a
separate step. To change an existing HTTP-only branch mapping and deploy it, run:

```sh
nanoweb deploy -configure -https
```

Confirm the settings and use the offered commit step, or commit them yourself
before retrying. `-configure` continues into deployment after saving settings.

`-agree-tos` records acceptance of the [Let's Encrypt subscriber terms](https://letsencrypt.org/repository/).
The domain must resolve publicly to the server, inbound TCP 80 and 443 must be
open, and the server needs outbound HTTPS to Let's Encrypt. HTTP challenges use
port 80; TLS-ALPN challenges use 443. No DNS-provider credentials are needed.
IP certificates and wildcard certificates are not supported. Contact email and
consent are required explicitly. They are saved within the selected branch's
settings. WWW redirect adds both hostnames to the certificate allowlist, and
redirects to the primary domain over HTTP or HTTPS while preserving path/query.

Nanoweb uses pinned `golang.org/x/crypto/acme/autocert` for certificate issuance,
caching, and automatic renewal. The daemon handles both ports without a proxy.
Only explicitly enabled site domains can obtain or use certificates. TLS uses
version 1.2 or newer and checks that the request Host matches its TLS hostname.
HTTP requests for enabled domains redirect to HTTPS with status 308, preserving
the path and query. Other sites stay on HTTP. Challenges and the generic health
endpoint remain available on HTTP. HSTS is not enabled in this first version.

Certificates and ACME account keys live under `/var/lib/nanoweb/certs` with
restrictive permissions. Preserve this directory across server updates and back
it up securely. A restart reuses cached certificates. The service account can
write certificates but cannot modify deployed releases or host configuration.

Activation waits up to two minutes for readiness and certificate verification,
using a locally dialed HTTPS connection with normal certificate/hostname checks.
A failure restores the previous configuration. Diagnose DNS, firewall, CAA, and
ACME errors with `sudo journalctl -u nanoweb -n 100 --no-pager` before retrying;
repeated issuance attempts can hit CA rate limits. Previously enabled HTTPS is
controlled by the saved branch preference on later deploys. Explicitly changing
that preference to false and redeploying disables TLS for that site's aliases;
cached certificates are retained. Other sites' HTTPS settings are preserved.

The resulting host config includes a `tls` object alongside `listen`, `sites`,
and the generated `revision`:

```json
{
  "listen": ":443",
  "email": "you@example.com",
  "agree_tos": true,
  "cache_dir": "/var/lib/nanoweb/certs",
  "domains": ["example.com"]
}
```

This is the value of `tls`, not a replacement for the whole server config. The
outer HTTP `listen` stays `:80`. Test public access from your Mac:

```sh
curl -I http://example.com/
curl -I https://example.com/
```

Tests simulate SSH/SCP failures, exercise real local HTTP/TLS listeners and
deployment recovery, and issue a certificate through a local ACME test authority.
They also verify persistent certificate reuse without the CA. No real Ubuntu
server or live Let's Encrypt account was provisioned during implementation.

## Set up an Ubuntu server from your Mac

From any directory, run `nanoweb setup`:

```sh
nanoweb setup
```

Enter your server's hostname, IP, or SSH config alias. Standalone setup always
connects as root, even if you supply another username. Custom ports and keys
belong in `~/.ssh/config`.
Nanoweb uses the installed `ssh` and `scp`: keys, agents, passphrases, normal
host-key checks, and interactive password authentication are handled by SSH.
If you use a password, password login must be enabled on the server. Passwords are never stored by
Nanoweb, and host-key verification is not disabled. Non-root installation uses
`sudo` with a remote terminal for its password prompt. A temporary SSH control
connection reuses authentication and is closed when setup finishes.

Targets: Ubuntu LTS 22.04, 24.04, or 26.04 with systemd, on amd64 or arm64.
The target needs OpenSSH/SFTP and standard Ubuntu tools including `ss` (iproute2),
plus sudo for non-root deployment targets. No compiler is needed on either machine.
Setup checks SSH and the remote OS/architecture, extracts the matching compressed
Linux server bundled inside the CLI, verifies its bundled SHA-256 and static ELF
architecture, then uploads it. Extraction is limited to 100 MiB. No server download,
compiler or registry is needed during setup. Verification happens before upload
or reset confirmation. There is no download or compiler fallback.

`nanoweb version` prints the CLI version; `nanoweb-server -version` prints the
server version. `bash scripts/build.sh` produces a complete `dev` build with both
servers from the same source. Plain `go build` does not refresh embedded servers
and is not the supported installation build. See [RELEASING.md](RELEASING.md).

Setup installs `/usr/local/bin/nanoweb-server`, `/etc/nanoweb/server.json`, and
`/etc/systemd/system/nanoweb.service`. It creates a dedicated non-login `nanoweb`
account and `/var/lib/nanoweb`. The service starts at boot and restarts on failure.
Setup verifies systemd activity and HTTP readiness before reporting success.
Before changing the server, the installer checks TCP and UDP ports 80 and 443
on IPv4 and IPv6. It refuses occupied ports unless every listener is verified as
the existing Nanoweb systemd main process during a confirmed reinstall. Unknown
owners or failed port checks also stop installation; other services are never stopped.

Setup connects as root; enter just the hostname or IP. When the server binary
and configuration already exist, it warns that all sites will go offline and
requires typing `DESTROY` to reset the installation. The replacement binary is
extracted and verified before this confirmation; extraction failures leave the
existing installation unchanged. Old files are moved into a
root-only `/var/backups/nanoweb.*` directory, including certificates and releases.

Setup and deploy automatically trust and save previously unseen SSH host keys.
Changed or revoked keys stop the connection with a friendly explanation.
Verify a changed fingerprint with your server provider before updating your
known_hosts entry; Nanoweb never removes saved keys automatically.
Restoration is manual; this is a reset, not an upgrade. Unexpected partial
installations and symlinked installation paths are refused. It does not modify firewalls,
SSH settings, packages, or existing sites. If installation fails, its newly
installed service/binary/config files are removed; any created account and data
directory remain for inspection. After a connection loss, inspect remote state
before retrying, since interrupted installs may leave partial state.

The host listens publicly on HTTP port 80 and has no sites configured:

```json
{
  "listen": ":80",
  "sites": {}
}
```

The `sites` object maps hostnames to absolute site directories, for example
`"example.com": "/var/lib/nanoweb/sites/example"`. The service account must be
able to read them. Sites/configuration load at startup; restart after changes.
The host serves validated snapshots without browser reload scripts, file watching,
or icon downloads. Resolve icons locally before copying a site. Unknown hosts
return 404, including requests to the server IP unless that IP is a configured
site key. `/_nanoweb/health` publicly reports only the service name and readiness.
The empty listen address binds all available interfaces; explicit IPv4/IPv6
addresses and numeric ports 1-65535 are also accepted. The service remains
non-root, with only `CAP_NET_BIND_SERVICE` granted for binding low ports.
Local `nanoweb dev` remains loopback-only.

Allow inbound TCP 80 in your server and cloud firewalls. Setup does not change
either firewall, and its local readiness check does not prove public reachability.
From your Mac, test public access (replace the example IP):

```sh
curl http://1.2.3.4/_nanoweb/health
```

On the server:

```sh
sudo systemctl status nanoweb --no-pager
sudo journalctl -u nanoweb -f
sudo systemctl restart nanoweb
```

Existing installations are not upgraded by rerunning setup. Deployment requires
the current server executable (deployment protocol 7, unquoted snip syntax).
For an older installation,
update the executable and systemd unit deliberately, preserve server.json and all
sites, and create `/var/lib/nanoweb/certs` owned by `nanoweb:nanoweb` with mode
0700. The data directory itself should be root-owned, mode 0755. The current
unit permits writes only to the certificate directory. Reload systemd and restart
after replacing the unit. An automated server upgrade command is not included.

## Local installation details

For a persistent local installation, from the repository:

```sh
mkdir -p "$HOME/.local/bin"
bash scripts/build.sh
cp bin/nanoweb "$HOME/.local/bin/nanoweb"
export PATH="$HOME/.local/bin:$PATH"
```

Add that PATH setting to your shell profile to keep it available in new terminals.

For framework development, keep your global symlink to `bin/nanoweb` and replace
your old build alias in `~/.zshrc` with:

```sh
alias nanoweb-build='bash /Users/sean/Projects/nanoweb/scripts/build.sh'
```

Run `source ~/.zshrc`, then `nanoweb-build` from any directory. The script finds
Go on PATH or at `/usr/local/go/bin/go`; `GO_BINARY` overrides that location.
Contributors need Go, Bash, gzip and shasum; installed users do not need them.

`new` accepts a simple name containing letters, digits, underscores, and hyphens, starting with a letter or digit. It refuses existing destinations, including empty directories and symlinks. An interrupted creation may leave a partial new directory for inspection. `status` validates configuration, routes, templates, and current content without listening on a port. Exit codes are 0 for success, 1 for project or runtime failures, and 2 for invalid CLI usage.

`nanoweb new my-site` also initializes Git on `main`, without staging or committing
anything. The starter includes a `.gitignore` for macOS metadata, temporary editor
files, and local environment secrets. To use a parent repository instead, run
`nanoweb new my-site --skip-git` (the flag also works before the name).
If Git is not installed, the site is still created with instructions for setting
up Git before deployment.

## Checks

From the repository root:

```sh
gofmt -w cmd internal
go test ./...
go test -race ./...
go vet ./...
bash scripts/build.sh
go test ./internal/cli -run TestExecutableWorkflow -count=1 -v
```

The executable smoke test builds the binary, creates a fresh starter, runs status, starts dev on an available loopback port, checks pages/assets/private URLs, edits snips/config/assets/templates, checks invalid-edit errors and recovery, and sends an interrupt to verify clean shutdown. Tests need permission to bind a local TCP listener.

## Current limitations

Editable regions, short IDs, sanitized HTML, page metadata, content-preserving
deployment, and archived orphan pruning are implemented. Deployment protocol 7
requires an updated server binary; older servers are refused before upload.
Local extraction preserves unreferenced content until deployment. There is no
Markdown renderer, visual rich-text toolbar, schema migration engine, or automatic
remote-to-local content synchronization yet. Keep backups before manual edits.

This release provides local development, SSH setup/deployment, automatic HTTPS,
email-code authentication and basic owner content editing. A rollback command,
server upgrades, release pruning, remote logs/status commands, public forms,
snip schemas, per-edit revision history, and managed snip image publishing remain
deferred. Deployment restarts the service briefly and preserves remote rich snips.
Put developer-managed images under `images/`; the editor cannot upload images.
Unrestricted raw HTML is unsupported. Real Ubuntu installation, live Resend
delivery and live CA verification remain pending; automated tests use fake email
delivery and local HTTP/TLS. This is not a production security certification.
