# FleaCMS

## Product and build specification

Deployment progress uses installed server TLS state: first-time HTTPS enablement
shows setup guidance, routine deployments show a concise health/HTTPS check.
This changes messaging only, not readiness checks or certificate renewal.

Implemented in 0.7.0: optional administrator-owned `downloads` host mappings serve
checksum-validated versioned Flea CLI downloads from pinned disk handles, outside
website snapshots and RAM asset maps. This is not a general filesystem mount.
The website retains latest.txt; the 256 MiB site archive limit is unchanged.
See [download hosting](download-hosting.md) for boundaries and publication.

The agreed public V1 scope and release gates are in [next-steps.md](next-steps.md).
V1 keeps source private and distributes free binaries without accounts or license
keys. The one-line installer, local `flea update`, and remote
`flea update-server` are implemented; public hosting is still pending. Homebrew
is deferred. The roadmap governs V1 scope where older future proposals below differ.

Typed image/video snips, multisnip photo fields, branch-scoped rollback, site backup/restore
and executable-only server upgrades are implemented. See docs/photos.md and
docs/recovery.md for exact bounds. Live systemd/SSH failure testing and signing/
publishing remain outstanding. Local optimized MP4 video uploads, posters, native
playback and deployment/backup preservation are implemented; see docs/videos.md.
No automatic conversion, WebM fallback or hosted-provider integration for v1.

Page-owned multisnips are implemented with direct ordered arrays, named
`snip.field` template values, extraction and conflict-checked editing. See
docs/multisnips.md for adjacent identical-markup seed extraction, photo fields
and limits. Developer seeds never replace an existing owner array. Editing
defaults on and persists per tab across navigation. Dashed overlays identify
editable content. Collections use inset handles, remove icons and per-group Add tiles, with no separate toggle, sidebar or
group dropdown. Pointer/keyboard reordering updates the same ordered draft;
shared page Save/Cancel applies all changes together. Photo fields offer existing
library images in a thumbnail picker and JPEG/PNG uploads, with a selected-photo
preview and contextual alt text. Private previews publish only with page Save.
Editable `mailers/` templates and local
email preview are implemented; see docs/email-templates.md. Protocol 24 requires
these runtime features. Adaptive streaming and automatic retention remain planned.

Dev reloads authenticated public pages when safe, restoring edit mode and scroll.
Unsaved drafts, open dialogs, saves block automatic reload.
Production owner pages do not poll for development reloads.
Page Save applies server-confirmed content in place without a full reload and
updates the version and Cancel baseline for subsequent edits.

Interactive dev startup asks for a missing site name and optional admin email,
never Resend setup. Noninteractive startup never prompts. New sites enable
dev.email_preview: actual local login/contact mail opens in an email-style browser
window without sending. Previews are bounded, expire in 10 minutes and live only
in memory. Production ignores the toggle. Preview-mode changes require restart;
real local delivery uses the explicit email wizard. Secrets stay in ignored .env,
while approved owners remain developer-owned configuration.

`flea dev --lan`/`-l` explicitly enables full development over trusted-network
HTTP, using detected private IPv4 URLs. Loopback remains the default. Exact
Host/port and actual-peer checks gate LAN requests; sessions, Origin, CSRF and
production HTTPS requirements remain intact. Previews open on the dev computer
and existing private capability URLs also work on approved network addresses.

Editable state: `flea pull` retrieves live snips, metadata and referenced media
into local files. Plain deployment preserves live content; `--with-snips`
publishes intentional local changes using scoped pull baselines and explicit
conflict resolution. No config/session/secret or CSS sync. Protocol 24 checks
reviewed state under the activation lock.

Styles are developer-managed in `site/css/theme.css`, using Prewind, theme tokens
and ordinary CSS. No remote Theme editor, CSS overlay or theme synchronization.
Normal deploy publishes committed CSS, backups include source styles, and rollback
restores the selected release's CSS while keeping current snip content.
Page Cancel/Save remain on the second row and keep Editing on. The switch label
is Edit, with dark-off/green-on styling. Editable overlays use 2px dashed green
boundaries with 6px corners; Sign out stays a compact pill. Keep visible focus,
source-change protection and reduced-motion behavior.

`flea email` configures Resend with an asterisk-masked API key, separate sender name and
verified-domain address, a non-sending connection check, and a private .env save.
Dev reads local .env mail settings automatically; shell variables take precedence.
Provider delivery and setup checks use separate interfaces for future adapters.

Versions follow SemVer, currently 0.2.0, with release notes in CHANGELOG.md.
Owner onboarding uses a developer-shared /admin link and email codes, without
registration or invitations. Local /admin provides first-owner and email setup
guidance; production with no approved owners remains disabled.
Admin pages use the site's layout and styles, with scripts disabled. Browser
form submissions retain their origin through the same-origin referrer policy.
The canonical version is internal/buildinfo/version.go; builds use the same
version for the CLI and bundled servers. The 0.x interface remains unstable.

CLI help: bare `flea` shows the grouped command overview and exits successfully.
Use `flea <command> --help` for command details without starting the command.

Development policy: prefer the current design over backward compatibility at
this early stage. Ask the user before adding compatibility or migration layers.

The included Contact mini app is implemented locally. apps.contact specifies
mount and email.to, with optional email.from. POST at the mount validates
name/email/message through an authenticated JSON child service; Flea owns
private email templates and delivery. Success redirects to mount/thank-you;
validation returns 422 and operational failures 503, preserving safe fields.
No submissions are stored. Old contact config/routes are not supported.
See apps/contact/README.md for the complete contract and acceptance limits.

The developer CLI also provides `flea admin [email]`: prompt when omitted,
create or append to admin.emails in the existing JSON config, skip duplicates,
validate the 20-address limit, and preserve unrelated settings. It does not
send email or deploy. It prompts for email settings when needed. Dev reloads
owner access and sign-in mail without restart; live changes require deployment.

Admin sessions are persistent JSON runtime state, not a database: local
flea/runtime/sessions.json and production sites/<name>/flea/runtime/sessions.json outside releases.
admin.session_idle_days defaults to 30 (omitted/zero), accepts 1–365, and renews
on authenticated activity. Active editing renews at most once a minute; idle tabs
do not. Sessions survive browser closure, restart and deploy. Logout and owner
removal revoke sessions. Store hashed tokens and host/CSRF/expiry metadata only,
with private permissions, bounded reads, locking and atomic saves. Never ship
session state in Git or release archives. Deployment protocol 24 requires the
private writable runtime directory; OTPs remain memory-only.

**Status:** first implementation specification  

Owners edit the current public page through a persistent top bar. Edit opens a
second formatting bar with Page details, available even without snips. One Save
publishes changed regions and metadata together. /admin is sign-in only.
**Audience:** developer building FleaCMS in Go
**Core promise:** a developer can create, run, and deploy a small editable website with one CLI, one server executable, plain HTML, JSON-backed content, and automatic HTTPS. No Node, Ruby, PHP, database server, Docker, or separately configured web server is required.

**Current implementation boundary:** The implementation provides local `new`, `dev`, and `status`, routing, escaped templates, sanitized rich snips with dev extraction and preserved deployment content, assets, automatic reloads, SSH `setup` and `deploy`, and opt-in managed HTTPS through `https` or `deploy -https`. Deployments validate immutable releases, atomically replace host config, restart briefly, verify readiness, and restore the prior config on failure. Email-code login, inline owner editing, multisnips, photo uploads, rollback, backups and executable upgrades are implemented. General-purpose forms, general field schemas, per-edit revisions and unrestricted raw HTML remain deferred. The sections below also describe future product direction, not just shipped capabilities. See README.md for the implemented contract and exact commands. Configuration is JSON only. Build requires Go 1.26 or newer for the pinned ACME library.

FleaCMS is for brochure-style sites: small businesses, restaurants, consultants, event services, portfolios, and similar sites where a conventional application framework or WordPress installation is disproportionate to the job.

It is not intended to become a general application platform, a no-code site builder, or a WordPress replacement with plugins.

## Product principles

1. **The site remains an understandable folder.** A developer should be able to open it in a text editor or file browser and understand its structure immediately.
2. **HTML is primary.** Pages are ordinary HTML files. FleaCMS adds a small templating layer but does not require a component framework or a build toolchain.
3. **Public routes come from files.** No route file is needed for normal pages. URLs are clean and extensionless.
4. **The owner edits structured content, not templates.** Developers own templates and partials; owners manage snips and their images through a simple editor.
5. **One install, one operational unit.** The production host is a FleaCMS Go executable running as a system service. It owns HTTP(S), site selection, TLS certificates, static serving, the editor, deployments, and limited server-side actions.
6. **No separate Caddy requirement.** FleaCMS must not ask users to install, configure, or operate Caddy. It may use mature Go libraries for ACME/TLS internally; it must not implement TLS certificate management from scratch.
7. **Safe by default.** Releases are atomic and reversible; content writes are validated and atomic; private files are never served; arbitrary code execution is not a feature.

## User experience

The intended local workflow:

```sh
brew install flea
flea new my-site
cd my-site
flea dev
```

The intended initial production workflow:

```sh
flea setup
flea deploy
```

`flea setup` runs from the Mac, prompts for an SSH target, verifies the connection, and extracts the matching bundled Linux server for the detected architecture and verifies its bundled checksum. It uses native SSH keys/agents or interactive password authentication, uploads the executable, installs its systemd service, and checks readiness. The current installation implementation supports Ubuntu LTS 22.04/24.04/26.04 on amd64/arm64 and listens publicly on HTTP port 80, with an empty JSON host configuration. It does not open firewall ports. `flea deploy -target user@host -domain example.com` publishes a validated site. `flea https` enables certificates for deployed domains using an explicit contact email and terms acceptance. DNS and firewalls are configured by the administrator.

The initial server target is **Ubuntu LTS only**. Do not support Arch or multiple distributions until the Ubuntu flow is stable and tested.

## Non-goals for v1

- A JavaScript runtime or arbitrary JavaScript expressions in templates
- Arbitrary server-side Go, Ruby, Python, shell, or user-provided code
- A plugin marketplace or plugin API
- A database requirement for site content
- Multi-user collaborative editing or merge/conflict resolution
- Docker as a required local or production dependency
- Kubernetes, PaaS, or container orchestration deployment targets
- Dynamic route parameters such as `/posts/:slug`
- WordPress-style page builders, themes, blocks, or WYSIWYG HTML authoring

## Site structure

Only `site/` is the public document root. Project configuration, mailers,
content and runtime state have separate locations.

```text
my-site/
  .flea.config.json
  flea/
    snips.json
    media/
    runtime/
      sessions.json
    skills/
      import-brochure/SKILL.md
      fleacms-site/SKILL.md
    data/
      team.json
    icons/
  mailers/
    main.layout.html
    contact.email.html
    contact.email.txt
    sign-in.email.html
    sign-in.email.txt
  site/
    main.layout.html
    admin.layout.html
    _404.html
    _500.html
    index.html
    about-us.html
    admin/
    css/
      theme.css
      global.css
    images/
      logo.svg
  .gitignore
  AGENTS.md
  README.md
```

### Ownership and visibility

| Location            | Owner                                   | Served publicly?               | Purpose                                           |
| ------------------- | --------------------------------------- | ------------------------------ | ------------------------------------------------- |
| `site/`             | Developer                               | Yes, except private templates | Pages and public assets                           |
| `main.layout.html` | Developer                               | No                             | Default page wrapper                              |
| `**/_name.html` | Developer                              | No                             | Reusable partial templates, except reserved names |
| `_404.html`    | Developer                               | Only as the 404 response       | Not-found document                                |
| `_500.html`    | Developer                               | Only as the 500 response       | Server-error document                             |
| `.flea.config.json`  | Developer                               | No                             | Site configuration                                |
| `flea/snips.json` | Site admin or developer | No | Editable content and page metadata |
| `flea/media/` | Site admin | Catalogue URLs under `/media/` only | Managed photos and optimized videos |
| `flea/runtime/` | Server | No | Ignored sessions and recovery journals |
| `flea/skills/` | Developer | No | Committed agent guidance, excluded from deployment |
| `mailers/` | Developer | No | Private email bodies and layout |

Any file or directory whose basename begins with `_` is private/reserved and must never be reachable through a public URL. `flea/` is additionally reserved as FleaCMS's private workspace.

## File-based routing

FleaCMS maps page files in `site/` to extensionless routes. Neither `site/` nor
`flea/` becomes a URL prefix. All `*.layout.html` templates remain private.
The root config includes branch-specific `deploy.branches` settings, stripped
from runtime archives and backups; credentials remain in ignored `.env`.

| File                         | Route               |
| ---------------------------- | ------------------- |
| `index.html`            | `/`                 |
| `about.html`            | `/about`            |
| `about-us/index.html`   | `/about-us`         |
| `about-us/history.html` | `/about-us/history` |
| `css/global.css`        | `/css/global.css`   |
| `images/logo.svg`       | `/images/logo.svg`  |

Rules:

1. `index.html` represents its containing directory.
2. Any other `.html` page represents the same URL without `.html`.
3. URLs containing `.html` are not alternate routes and return 404. FleaCMS does not redirect `/about.html` to `/about`.
4. An asset retains its filename and extension.
5. Reserved/private underscore paths are not routes.
6. A duplicate route, such as both `about.html` and `about/index.html`, is a build error. `flea dev` reports it and `flea deploy` refuses it.
7. An unmatched route returns `_404.html` with HTTP status 404, wrapped in `main.layout.html` when present. The 404 template contains page content for `{slot}`, not a separate HTML document.
8. An unexpected request-time/internal rendering failure returns `_500.html` with HTTP status 500. This fallback must remain deliberately simple.

Implemented asset storage: ordinary files above 64 KiB and managed photos stream
from pinned, validated regular-file handles. Smaller assets and theme CSS remain
byte snapshots. No request opens a filesystem path. Streamed files reject links;
size/mtime changes fail closed (503 before headers, interrupted transfer afterward).
Published files must be replaced atomically, not changed in place. Managed videos
also stream. OS caching is separate from retained Go heap.

In dev, 404 responses use the site's ordinary error template; 405/500 responses
show self-contained diagnostic pages to loopback peers using a localhost/loopback
Host, or explicitly authorized LAN requests. Include escaped error text,
up to 11 template lines with original line numbers and the error highlighted
(five lines on either side) when available, safe JSON locations,
route guidance, and automatic recovery after save. Never expose source details
in production. Dev errors go to the terminal; production errors go to server logs,
including custom-error-template failures. Initial validation failures precede
server startup and remain terminal-only.

Successful public pages use native cross-document CSS transitions for same-origin
navigation. Crossfade a sole direct-child main, otherwise the root, over 240 ms.
Honor reduced motion; unsupported browsers retain normal navigation. The starter
imports editable css/navigation.css; there is no runtime CSS injection or JS router.
The layout explicitly gates admin/_toolbar.html with {#if app.current_admin} and
supplies a csrf-token meta tag. This reserved request-only object is absent for
visitors. External editor assets and a protected metadata endpoint keep source small.

The first version has no `routes.json`. Add redirects only when there is a real need, through `.flea.config.json`.

## Templates

FleaCMS borrows the structural readability of Svelte templates but does not execute JavaScript. It uses single curly braces.

### Layouts and slots

`main.layout.html` is the default layout for normal pages. A page's rendered body replaces `{slot}`.

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>{page.title} | {app.site_name}</title>
    <link rel="stylesheet" href="/css/global.css">
  </head>
  <body>
    {render header}
    {slot}
    {render footer}
  </body>
</html>
```

Pages use the default layout without configuration in v1:

```html
<main>
  <h1>About us</h1>
  {snip about-us}
</main>
```

A leading `<!-- flea` settings comment may select `layout: /admin.layout.html`
before its closing `-->`. Root-relative private HTML layouts replace the default
wrapper for that page and support `{slot}`. Unknown or duplicate settings, missing
layouts and nesting are errors. Only layout selection is currently supported;
this is a small key:value format, not YAML. `_500.html` stays standalone.

### Partials

Partials are developer-owned, underscore-prefixed HTML templates anywhere under site/. They can be reused on multiple pages and can render data, other partials, and snips.

```html
<!-- _header.html -->
<header>
  <a href="/">{app.site_name}</a>
  {render nav}
</header>
```

```html
<!-- _nav.html -->
<nav>
  <a href="/about-us">About us</a>
  <a href="/contact">Contact</a>
</nav>
```

`{render header}` resolves to `_header.html` beside the calling template. `{render /foo/bar/footer}` resolves to `foo/bar/_footer.html`. Paths are unquoted and omit the filename's leading underscore and `.html`. Relative directories, including `./` and `../`, resolve from the calling template, including inside nested partials. Paths escaping site/ are invalid. The root `main.layout.html`, `_404.html`, and `_500.html` are reserved and cannot be rendered as partials. A missing partial or recursive partial inclusion is a dev/build error, even in unused partials.

### Icons

`{icon arrow-right}` renders an inline Lucide SVG. Optional attributes use
quoted HTML syntax, for example `{icon search size="20" aria-label="Search"}`.
Names are literal lowercase kebab-case. Attribute values support escaped
dotted-path interpolation. Icons inherit currentColor and are decorative by
default; labeled icons use role img.

Do not bundle the library. Phase-one dev resolves missing referenced icons
from pinned lucide-static 1.47.0 into `flea/icons/name.svg` before serving,
including during reloads. Include the license in each SVG comment and ship
`flea/icons/.keep`, not a version directory or separate license. Status is read-only
and reports missing icons as warnings without a failing exit code. Missing icons
or download failures render nothing and do not cause a 500 response. Malformed
directives and unsafe SVGs remain errors. Downloads are bounded, cancellable, and atomic;
SVGs are reconstructed from a strict geometry-only vocabulary. No runtime
browser library or request-time downloads. Deploy resolves icons before publishing.
See README.md for supported attributes.

### Snips

Snips are owner-editable sanitized HTML stored in private `flea/snips.json`.
Spans and CSS classes on supported content elements survive rendering and editing.
Scripts, event handlers and arbitrary inline styles remain blocked.
They are distinct from partials and structured developer data:

- A partial is a reusable developer template.
- A snip is owner content identified by a stable 10-character alphanumeric ID.
- Structured objects for loops live in `flea/data/*.json`, exposed as `data.*`.

Create an editable region while dev runs:

```html
<h1 snip>Welcome to FleaCMS</h1>
```

Dev extracts inner HTML and inserts a stable reference:

```html
<h1>{snip a7Km2pQ9xB}</h1>
```

The `snip` attribute is an authoring marker and is removed during extraction.
Other attributes and source formatting are preserved. The directive alone
identifies saved content; owner-only rendered boundaries identify it to the editor.

Storage groups entries by route, with shared regions for layouts/partials:

```json
{"pages":{"/":{"title":"Home","description":"Welcome","snips":{"a7Km2pQ9xB":{"content":"Welcome to FleaCMS"}}}},"shared":{"snips":{}}}
```

Extraction is dev-only, journaled and serialized. Content is persisted before
rewriting templates; interrupted work resumes or stops on conflicting edits.
Local content serialization is in-process. A separate OS lock excludes concurrent
dev, pull and reset commands. A new dev invocation asks only the same user's dev
server for that site to shut down through a private Unix socket, then waits for
the lock before loading or writing. Pull/reset are never interrupted. Every
relevant save extracts additions, reindexes labels/routes,
archives removals and restores reintroduced IDs. Existing content and metadata
are retained. Markers use {snip ID} without quotes; quoted forms are rejected.
HTML comments do not render snips. IDs contain only ASCII letters and digits.
Nested regions and template directives in extracted content are errors. Rendering
reconstructs a formatting/link allowlist and never interprets JSON as template
syntax. Rich output is only allowed in element content, not attributes or raw-text
elements. Page metadata and structured values remain escaped. There is no legacy
named-snip compatibility. Inline editing includes developer-designated image/video
snips: src/alt/poster bind to typed JSON entries, layout stays in HTML, and the
page's Save/Cancel covers text and media together. Media cannot nest in text snips.

Deploy preserves remote rich content/metadata and seeds missing IDs. Content
snapshots live outside code releases and are switched by the activation journal.
Aliases pointing to the same release must select the same content snapshot;
conflicting paths are rejected and valid aliases share a loaded serving snapshot.
After successful activation, unreferenced entries move to an archive map, retaining
last remote content. References are lexical, including false/empty branches and
unused partials. Failed activation leaves previous content untouched. Protocol 24
is required. No automatic archive deletion or rollback command exists yet.

### Values, conditions, and loops

Initial supported syntax:

Owner-editable snip IDs are 10 case-sensitive ASCII alphanumeric
characters, generated using cryptographic randomness with collision checks
against active and archived IDs across all routes and shared regions. They are
content identifiers, not UUIDs, credentials, or access-control tokens. The ID
allocator, route-based snips.json, extraction, sanitization, and deployment
preservation workflow are implemented.

```html
{app.site_name}
{page.title}
{snip.about-us.title}

{#if app.phone}
  <a href="tel:{app.phone}">{app.phone}</a>
{/if}

{#each snip.about-us.team as member}
  <article>
    <img src="{member.photo}" alt="{member.name}">
    <h2>{member.name}</h2>
    <p>{member.role}</p>
  </article>
{/each}
```

The expression model is intentionally constrained:

- Value lookup: dotted paths only, for example `{app.site_name}` and `{member.photo}`.
- Condition: a dotted-path truthiness check only, for example `{#if app.phone}`.
- `{#unless dotted.path}...{/unless}` is an inverse if using the same truthiness.
- If/unless conditions support `{:else if dotted.path}` and a final `{:else}`. Loops support
  `{:else}` for empty or missing arrays, with outer aliases restored in the fallback.
- Loop: a collection dotted path and one alias only, for example `{#each team.members as member}`.
- An empty loop renders nothing.
- No function calls, arithmetic, comparisons, `.length`, filters, inline JavaScript, or arbitrary expressions in v1.
- FleaCMS does not parse template directives inside `<script>` or `<style>` elements.

Values interpolated by `{path}` are contextually escaped. Short-ID snips render
sanitized HTML by default; unrestricted raw HTML is not supported.

Mini apps use `{csrf}` as an explicit hidden-input directive in public HTML
element content, independent of admin sessions. Signed visitor cookies and form
tokens expire after 24 hours or restart; template reloads retain the key.
The bundled Contact app has public mount routing, private managed launch,
email capability, mount-relative redirects and bounded visitor flash. Admin
form fields remain unchanged. General app installation, remote registrations,
dynamic read routes and databases remain planned in docs/agents/planning/apps.md.

## Snip format and schemas

The following schema discussion is future work. The current owner-content
format is the route/shared store above, with no schema files or image publishing.

The exact schema file convention is a small deferred implementation choice, but the data model should be developer-defined and owner-editable.

Example snip data:

```json
{
  "title": "About us",
  "introduction": "We make wood-fired pizza for events.",
  "team": [
    {
      "name": "Sam Lee",
      "role": "Founder",
      "photo": "/_assets/snips/about-us-images/sam-lee.png"
    }
  ]
}
```

Recommended schema goals:

- Text, textarea/rich text, number, boolean, URL, image, select, object, and repeatable-list fields
- Required fields and basic validation
- Labels and help text for editors
- Image constraints such as allowed file type and maximum file size
- No arbitrary form generation from untrusted data

FleaCMS validates a snip before writing it. It writes content atomically: create a temporary file in the same directory, fsync as appropriate, then rename it over the existing JSON file. It retains prior revisions/snapshots for rollback.

## Configuration

`.flea.config.json` contains JSON only. Comments and trailing commas are not supported. The site loader reads the `app` object. Local `.flea.config.json` stores a `branches` map of confirmed deployment settings (target, domain, name, www_redirect, https, email, agree_tos, expected_ips). First deploy prompts and saves; later deploys select only the current branch. Changed settings require `-configure` confirmation. It contains no credentials and is excluded from uploads. Use branch names directly without environment labels or questions. Generate the internal site name automatically; staging suggests a subdomain. General page redirect settings remain deferred; WWW-to-primary redirects are supported.

Example:

```json
{
  "app": {
    "site_name": "Acme Catering",
    "email": "hello@example.com"
  }
}
```

Configuration should not contain private server credentials. The CLI stores deploy target credentials/settings locally outside version control, or uses standard SSH configuration and keys. The deploy target can be recorded in a `.gitignore`d local file if the developer wants convenience.

## Editor

Site-wide template variables live in `app.*`, backed by the config's `app` object.
`app.site_name` is prompted once on interactive dev startup when absent, with a
title-cased folder default. Noninteractive loads do not write defaults. Deployment
requires a committed, nonempty name. Page metadata and structured data retain
`page.*` and `data.*`; secrets and access settings are not public app variables.
The old `site.*` namespace is not supported.

The implemented editor is served at `/admin`, separate from public template routes.
Approved addresses are configured by the developer in `admin.emails` in JSON site
config. Login uses an emailed, six-digit, single-use code with expiry and attempt
limits, not passwords. Production requires HTTPS. Resend is the first provider,
behind a provider-neutral Sender interface; credentials live in a root-only server
EnvironmentFile and are prompted for during deployment when missing.

Implemented editor capabilities:

- Email-code authentication and secure host-bound session cookies
- Edit rendered snips in place and page title/description in a modal
- Publish validated content without restarting the service
- Validate before saving
- Serialize saves with deployment and reject stale forms

Schema-defined fields, images and per-edit revision
history remain future work. The editor cannot change the email allowlist.

Signed-in owners return to the homepage with inline editing of existing wrappers.
Rich formatting works in prose containers; leaf elements allow inline marks and
links, never nested blocks. Visitors receive neither the bar nor region markers.
Owner CSP blocks site scripts; dev reload protects drafts and restores edit mode.
The rendered snapshot supplies the content/template version; atomic saves reject
stale drafts, duplicate IDs and regions not editable on the selected route. Failed
saves preserve drafts. /admin is for login, not a dashboard. The signed-in top bar has
right-aligned Edit and Sign out buttons. Edit starts editing the current page and
replaces Edit with Cancel and Save changes. Sign out remains visible after a
vertical divider. Save is disabled until actual
changes exist. Page details and formatting slide out in a second bar during Edit,
collapsing outside Edit and respecting reduced motion. Formatting requires a focused editable region:
normal text, H1–H4, bold, italic, underline, strikethrough, links,
bullet/numbered lists, and left/center/right alignment. Page details holds title
and description. Save applies the page's changes together; Cancel discards them,
with an unsaved-navigation warning. Shared regions are clearly identified.
Media is separate from rich text; repeatable multi-snips remain future work.
The link control opens the current destination for editing, with Remove link in
the dialog. No undo/redo toolbar buttons; keyboard history remains.
Alignment uses three icon buttons. Unlinked-page discovery is deferred.

Editor exclusions for v1:

- Editing HTML, layouts, or partials
- Arbitrary filesystem browsing
- Arbitrary HTML/script injection
- Multiple-editor collaboration
- A visual drag-and-drop page builder

## Server and TLS

The installed FleaCMS server is a single Go executable and system service. A separate Caddy process is not part of the required architecture.

Responsibilities:

- Listen on HTTP port 80 and HTTPS port 443
- Redirect HTTP to HTTPS except as required for ACME challenges
- Obtain and renew certificates through a mature Go ACME/TLS implementation
- Map request hostnames to deployed sites
- Serve assets and rendered pages
- Serve `/admin` and a restricted set of internal APIs
- Enforce request size limits, upload limits, and secure headers
- Write structured logs
- Support multiple independent sites on one server

Do not hand-roll ACME, TLS certificate renewal, password storage, cookie signing, or multipart upload parsing. These are implementation dependencies, not user-visible operational dependencies.

## Deployments and releases

Current implementation: release directories use a stable site name and random
timestamped IDs. The host config points directly at the selected absolute release
directory, without a `current` symlink. Config switching is atomic but service
restart causes a brief interruption. A lock and durable recovery journal protect
activation; interrupted transactions can be restored with `flea-server -recover`.
Old releases are retained, with a branch-scoped rollback CLI but no pruning. Committed rich
snips seed missing IDs; existing remote content/metadata wins. Content snapshots
live separately under content/<release-id>.json, selected by the host config.
Successful activation archives/prunes unreferenced IDs; failed activation restores
the prior code and content paths. See docs/recovery.md for the current rollback contract.

Current implementation: `flea server list [ssh-target]` lists installed sites
read-only and `flea server remove [ssh-target] [site-name]` unregisters one
site's routing after an exact typed confirmation. An omitted target is prompted
for and an omitted site name lists the managed sites and asks for one by its
exact name. Removal deletes no files:
releases, content, media, sessions, certificates and downloads are retained under
the stable site's data directory, but retained content is not restored
automatically by redeploying the same name. It uses the deployment lock, an
expected-state digest covering routing, ancillary entries and branch/commit
bookkeeping, the existing recovery journal, restart/probe and rollback paths, and
fails closed on ambiguous redirects or unmanaged roots. Restarting the shared
service briefly affects other sites; when no HTTPS hostnames remain afterwards,
automatic HTTPS is disabled while cached certificates stay on disk. No public
HTTP administration API is added.

Each deployment creates an immutable release. The host switches to it atomically.

```text
/var/lib/flea/sites/acmecatering.com/
  releases/
    20260920-153012/
    20260919-102417/
  current -> releases/20260920-153012
  shared/
    content-revisions/
    uploads/
```

The exact treatment of editable snips during a developer deployment needs a defined policy before implementation. Recommended default:

- HTML, partials, layouts, CSS, JavaScript, and developer-provided default snips ship in releases.
- Owner edits and uploaded files are stored in a host-managed mutable content layer with revision history.
- On deploy, FleaCMS preserves existing owner-edited snips and uploads unless the developer intentionally runs an explicit content migration/reset command.

This avoids accidentally erasing an owner’s edits every time a developer redeploys HTML.

Required deployment properties:

- Validate routes, templates, snips, and configuration before activating a release.
- Upload a staged release before changing `current`.
- Switch releases atomically.
- Preserve the prior active release.
- Provide `flea rollback`.
- Report the deployed domain, release identifier, and any certificate state clearly.

## CLI surface

Keep the initial command surface small:

```text
flea new <site-name>
flea dev
flea setup
flea deploy
flea status
flea logs
flea rollback
```

### Command behavior

`flea setup`

- Prompt `Server hostname or IP`; standalone setup connects as root.
- Fail on SSH/OS checks before extracting the matching bundled Linux server.
- Verify its embedded SHA-256 and static Linux architecture; no Go or source
  checkout on either machine. scripts/build.sh bundles both Linux architectures
  into development CLIs; native CI builds and tests release artifacts per architecture.
- Upload through SSH/SCP, install as a dedicated systemd user,
  and verify readiness. Support native key and interactive password authentication.
- Extract and verify the replacement before requesting reset confirmation.
- Recognized existing installations require typed DESTROY confirmation and retain
  a root-only backup. Check TCP/UDP 80/443 before changing anything, exempting
  only the verified FleaCMS service during reinstall. Preserve unrelated services.
- Initial host listens publicly on HTTP port 80. Deploy and HTTPS are subsequent
  explicit commands; no firewall changes are made.

`flea new <site-name>`

- Creates a minimal layout/index site by default. `--starter`/`-s` adds the
  complete example site, including six fictional team members and contact pages.
  Native HTML popover navigation supplies the mobile menu without JavaScript.
  `flea reset` confirms before replacing local files/settings with the minimal
  site, retaining .git and a private sibling backup. It never changes remote sites.
- Initializes Git on main and creates an `Initial site` commit using normal Git
  identity, hooks and signing. `--skip-git` skips all Git operations. Missing Git
  or a failed commit preserves the created site with instructions to finish setup.
  Includes a starter .gitignore. No remote or hosted project is created; successful
  commit output explains how to add origin and push main.
- Creates default `.flea.config.json`, editable starter copy, a layout with inline
  header/footer, error pages, and a welcome page with its SVG mark. Navigation links
  Team and Contact; /test remains an unlinked example. Contact uses an introduction
  beside a three-field form, stacked on mobile. Home, Team, Contact, Thank You and
  footer copy are snips; form controls and navigation remain developer-managed.
- Includes editable `admin/index.html`, `code.html`, `setup.html`, and `error.html`, selecting `admin.layout.html` and escaped `admin.*` data. The admin directory has no public routes or assets; server-side editor state selects templates after access checks. Missing templates have no embedded fallback.
  Access screens share a narrow responsive layout; sign-in and verification retain
  an Inter/forest-green theme with warm neutral pages and white form surfaces, plus
  native autofill and the existing six-box code enhancement. The standalone 500
  template includes its own minimal CSS, independent of public stylesheets.
- Bundles local Prewind 1.3.1 and Preflight-inspired theme, typography, native form, button, and application styles. CSS layers are ordered reset, styles, prewind. No CSS build step or CDN is required; form styling does not implement submission handling.

`flea dev`

- Must run without Docker or a database.
- Watches source files and snip content.
- Logs requests as `GET /test` for 200 or `404 GET /missing` for non-200.
  Include assets and HEAD, but omit successful internal reload polls and query strings.
- Use native directory notifications with a 100 ms debounce, including new nested directories and atomic saves. No periodic file scans or polling fallback. Watch errors stop dev; overflow reconciles watches and content. Public browser version polling remains separate.
- Runs a local server with the same route and rendering behavior as production.
- Reports template errors, invalid route collisions, missing partials, schema errors, and unsafe/unserved paths clearly.
- Provides local editor access.

`flea deploy`

- Before deployment, offer to commit pending changes with a file preview,
  confirmation, and commit message. Respect Git hooks and signing. Declining or
  failure stops deployment. Sites within larger repositories commit manually.

- Requires deployment protocol 24 before upload, offering first-time setup when absent. Validates the
  packaged site, uploads over SSH/SCP, stages, and atomically switches host config.
- Restarts and verifies revision/hostname; restores the prior config on failure.
- Uses native SSH and the saved current-branch mapping in .flea.config.json. The first
  invocation confirms server/domain, WWW redirect, and
  HTTPS, then saves configuration for future deploys.
- Ask for an SSH username, default root, when user@host is not supplied. Save it
  explicitly. Do not ask about advanced NAT/proxy IP settings in the normal flow.
- After successful SSH, offer setup if the server executable is missing, then
  resume deployment. Reuse the target and authenticated connection. Do not treat
  login failures or old installations as permission to install or overwrite.
- Requires clean Git status and packages the exact commit, resolving missing
  icons only inside a temporary snapshot. Rejects detached/unborn/non-Git deploys.
- The configuration wizard may run on a branch before its initial commit; settings
  and site files can then be committed together before any upload.
- Before upload, checks DNS on the server. All A/AAAA records must match expected
  interface or explicitly approved NAT/proxy addresses. IP literals skip DNS.
- Enforces distinct site names and hostnames per branch, with server-side ownership.
- Optional -https, -email, and -agree-tos enable managed certificates too.

`flea https`

- Enables Let's Encrypt certificates for existing domains. Uses autocert for
  issuance, cache and renewal, with exact host allowlists and TLS Host/SNI checks.
- Requires DNS pointing at the server, inbound TCP 80/443, contact email and terms
  acceptance. Redirects HTTP to HTTPS for enabled domains; health stays on HTTP.
- Tests use a local ACME authority; real Ubuntu/live CA validation is pending.

`flea status`

- Prints local validity, site counts and Git state first, then automatically checks the current branch's saved remote over SSH. Separately reports website reachability, deployed commit, content differences; CSS is part of code. Direction requires a scoped pull baseline. No remote flag; no mapping finishes locally.

`flea logs`

- Streams or retrieves host logs without requiring the developer to know system-service commands.

`flea rollback`

- Lists/selects or defaults to the immediately previous release, then atomically reactivates it.

## Docker

Docker is not required for local development or the first production path. The core product must work on a blank Ubuntu LTS host using the FleaCMS executable and OS service management.

Later, Docker can be an optional deployment target for users who already operate containers. It must not dictate the local project format or force every customer to understand containers.

## Server-side features

The first server-side feature should be a constrained contact-form action, not arbitrary application code.

A form can declare a named FleaCMS endpoint and validation configuration. FleaCMS handles rate limiting, spam defenses, payload validation, delivery/storage, and logging. The implementation may later offer email or webhook destinations, but those credentials live in host configuration, never in public site files.

Do not add custom server functions in v1. They turn a small static host into a multi-tenant code-execution and security platform.

## Security requirements

- Private/reserved file paths cannot be fetched over HTTP.
- Content JSON and template data are escaped by default.
- Raw HTML is an explicit, schema-controlled capability.
- Editor passwords use a current secure password hash algorithm.
- Cookies are `HttpOnly`, `Secure`, and have appropriate `SameSite` policy.
- Uploads enforce size and MIME/type validation; stored filenames are sanitized/randomized as appropriate.
- Uploads are served with safe headers and must not be executable.
- The host has per-site separation in paths and authorization checks.
- Form actions include rate limiting and abuse/spam protections.
- Deploy archives are validated to prevent path traversal.
- Logs avoid credentials, session tokens, and sensitive form contents.
- Certificate private keys and host operational state have restrictive filesystem permissions.

## Suggested implementation milestones

### Milestone 1: local static FleaCMS

- CLI skeleton and `flea new`
- Parse project layout and configuration
- File route generation with collision detection
- Local HTTP server
- Static asset serving
- `_404.html` and `_500.html` status handling

### Milestone 2: template renderer

- Root layout and `{slot}`
- Partials
- Value lookup, `#if`, and `#each`
- Clear parse/render error reporting
- Escaping and script/style parser boundaries

### Milestone 3: snips

- JSON snip loading
- Basic schema format and validation
- Associated image directories and safe public asset mapping
- Local editor authentication and edit UI
- Atomic writes and content revisions

### Milestone 4: production host

- Single host daemon serving multiple sites
- Domain-to-site lookup
- ACME TLS integration and renewal
- System service installation on Ubuntu LTS
- Host-level account/site registry design (site sessions already use private JSON)

### Milestone 5: deployment

- SSH bootstrap of a blank Ubuntu LTS host
- Staged immutable releases
- Atomic release activation and rollback
- `status` and `logs`

### Milestone 6: constrained forms

- Form declarations
- Validation, spam controls, rate limits, logging
- One delivery mechanism, chosen deliberately

## Open decisions to resolve before their milestone

These are intentionally not blockers for the initial routing/template prototype:

1. **Snip schema filename and syntax:** colocated `about-us.schema.json`, a central schema folder, or schema inside the snip JSON.
2. **Public URL prefix for flea-managed images:** use a stable virtual prefix and do not expose `flea/` directly.
3. **Rich text model:** sanitized HTML versus a constrained Markdown-like format.
4. **First form destination:** email, host inbox, webhook, or a combination.
5. **Editor user model:** one site-owner account initially versus host-managed multiple editor accounts.
6. **Default-content deploy policy:** preserve live editor changes by default, with explicit content migration/reset commands.

## Acceptance criteria for the first usable release

FleaCMS is meaningfully usable when a developer can:

1. Install the CLI through the one-line fleacms.com installer on a supported platform.
   Release jobs cover Mac ARM64 and Linux AMD64/ARM64. Published 0.11.0 lacks the
   ARM64 Linux CLI; that platform requires a new natively verified release.
2. Run `flea new my-site` and receive a working starter site.
3. Run `flea dev` with no language runtime, Node modules, database, Docker, or external web-server setup.
4. Add `about-us/history.html` and see it at `/about-us/history`.
5. Confirm `/about-us/history.html` returns 404.
6. Reuse a developer partial in several pages.
7. Wrap pages with `main.layout.html` using `{slot}`.
8. Render values, simple conditions, and loops from a JSON snip.
9. Edit a schema-defined snip and associated image from a protected editor without editing source HTML.
10. Deploy to a blank supported Ubuntu server with one command after entering the initial domain/server details.
11. Receive a valid automatically managed HTTPS certificate.
12. Deploy a second release and roll back safely.
13. Host more than one domain/site from the same FleaCMS server.
