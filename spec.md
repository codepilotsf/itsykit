# Nanoweb

## Product and build specification

The developer CLI also provides `nanoweb admin [email]`: prompt when omitted,
create or append to admin.emails in the existing JSON config, skip duplicates,
validate the 20-address limit, and preserve unrelated settings. It does not
send email or deploy. Restart dev or deploy to apply changed access settings.

**Status:** first implementation specification  
**Audience:** developer building Nanoweb in Go  
**Core promise:** a developer can create, run, and deploy a small editable website with one CLI, one server executable, plain HTML, JSON-backed content, and automatic HTTPS. No Node, Ruby, PHP, database server, Docker, or separately configured web server is required.

**Current implementation boundary:** The implementation provides local `new`, `dev`, and `status`, routing, escaped templates, sanitized rich snips with dev extraction and preserved deployment content, assets, automatic reloads, SSH `setup` and `deploy`, and opt-in managed HTTPS through `https` or `deploy -https`. Deployments validate immutable releases, atomically replace host config, restart briefly, verify readiness, and restore the prior config on failure. Email-code login and basic owner edits at /admin are implemented. A rollback CLI, public forms, schemas, per-edit revisions, managed snip image publishing, and unrestricted raw HTML remain deferred. The sections below also describe future product direction, not just shipped capabilities. See README.md for the implemented contract and exact commands. Configuration is JSON only. Build requires Go 1.26 or newer for the pinned ACME library.

Nanoweb is for brochure-style sites: small businesses, restaurants, consultants, event services, portfolios, and similar sites where a conventional application framework or WordPress installation is disproportionate to the job.

It is not intended to become a general application platform, a no-code site builder, or a WordPress replacement with plugins.

## Product principles

1. **The site remains an understandable folder.** A developer should be able to open it in a text editor or file browser and understand its structure immediately.
2. **HTML is primary.** Pages are ordinary HTML files. Nanoweb adds a small templating layer but does not require a component framework or a build toolchain.
3. **Public routes come from files.** No route file is needed for normal pages. URLs are clean and extensionless.
4. **The owner edits structured content, not templates.** Developers own templates and partials; owners manage snips and their images through a simple editor.
5. **One install, one operational unit.** The production host is a Nanoweb Go executable running as a system service. It owns HTTP(S), site selection, TLS certificates, static serving, the editor, deployments, and limited server-side actions.
6. **No separate Caddy requirement.** Nanoweb must not ask users to install, configure, or operate Caddy. It may use mature Go libraries for ACME/TLS internally; it must not implement TLS certificate management from scratch.
7. **Safe by default.** Releases are atomic and reversible; content writes are validated and atomic; private files are never served; arbitrary code execution is not a feature.

## User experience

The intended local workflow:

```sh
brew install nanoweb
nanoweb new my-site
cd my-site
nanoweb dev
```

The intended initial production workflow:

```sh
nanoweb setup
nanoweb deploy
```

`nanoweb setup` runs from the Mac, prompts for an SSH target, verifies the connection, and extracts the matching bundled Linux server for the detected architecture and verifies its bundled checksum. It uses native SSH keys/agents or interactive password authentication, uploads the executable, installs its systemd service, and checks readiness. The current installation implementation supports Ubuntu LTS 22.04/24.04/26.04 on amd64/arm64 and listens publicly on HTTP port 80, with an empty JSON host configuration. It does not open firewall ports. `nanoweb deploy -target user@host -domain example.com` publishes a validated site. `nanoweb https` enables certificates for deployed domains using an explicit contact email and terms acceptance. DNS and firewalls are configured by the administrator.

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

Every site follows this shape. the project root is the public document root. `_/` is Nanoweb private data and configuration.

```text
my-site/
  _/
    nanoweb.config
    snips.json
    data/
      team.json

  _layout.html
  _404.html
  _500.html
  _header.html
  _footer.html
  _nav.html
  index.html
  about-us/
    index.html
    history.html
  css/
    theme.css
    global.css
    typography.css
  images/
    logo.svg
```

### Ownership and visibility

| Location            | Owner                                   | Served publicly?               | Purpose                                           |
| ------------------- | --------------------------------------- | ------------------------------ | ------------------------------------------------- |
| the project root             | Developer                               | Yes, except reserved `_` names | Pages and public assets                           |
| `_layout.html` | Developer                               | No                             | Default page wrapper                              |
| `**/_name.html` | Developer                              | No                             | Reusable partial templates, except reserved names |
| `_404.html`    | Developer                               | Only as the 404 response       | Not-found document                                |
| `_500.html`    | Developer                               | Only as the 500 response       | Server-error document                             |
| `_/nanoweb.config`  | Developer                               | No                             | Site configuration                                |
| `_/snips/`          | Site owner through editor, or developer | No direct file serving         | Editable structured content and associated images |

Any file or directory whose basename begins with `_` is private/reserved and must never be reachable through a public URL. `_/` is additionally reserved as Nanoweb's private workspace.

## File-based routing

Nanoweb maps page files in the project root to extensionless routes.

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
3. URLs containing `.html` are not alternate routes and return 404. Nanoweb does not redirect `/about.html` to `/about`.
4. An asset retains its filename and extension.
5. Reserved/private underscore paths are not routes.
6. A duplicate route, such as both `about.html` and `about/index.html`, is a build error. `nanoweb dev` reports it and `nanoweb deploy` refuses it.
7. An unmatched route returns `_404.html` with HTTP status 404, wrapped in `_layout.html` when present. The 404 template contains page content for `{slot}`, not a separate HTML document.
8. An unexpected request-time/internal rendering failure returns `_500.html` with HTTP status 500. This fallback must remain deliberately simple.

The first version has no `routes.json`. Add redirects only when there is a real need, through `_/nanoweb.config`.

## Templates

Nanoweb borrows the structural readability of Svelte templates but does not execute JavaScript. It uses single curly braces.

### Layouts and slots

`_layout.html` is the default layout for normal pages. A page's rendered body replaces `{slot}`.

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>{page.title} | {site.name}</title>
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

Do not add page-specific or nested layouts in the first implementation. A global layout plus partials is enough, keeps behavior obvious, and leaves room to add an explicit layout directive later if experience justifies it.

### Partials

Partials are developer-owned, underscore-prefixed HTML templates anywhere under the project root. They can be reused on multiple pages and can render data, other partials, and snips.

```html
<!-- _header.html -->
<header>
  <a href="/">{site.name}</a>
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

`{render header}` resolves to `_header.html` beside the calling template. `{render /foo/bar/footer}` resolves to `foo/bar/_footer.html`. Paths are unquoted and omit the filename's leading underscore and `.html`. Relative directories, including `./` and `../`, resolve from the calling template, including inside nested partials. Paths escaping the project root are invalid. The root `_layout.html`, `_404.html`, and `_500.html` are reserved and cannot be rendered as partials. A missing partial or recursive partial inclusion is a dev/build error, even in unused partials.

### Icons

`{icon arrow-right}` renders an inline Lucide SVG. Optional attributes use
quoted HTML syntax, for example `{icon search size="20" aria-label="Search"}`.
Names are literal lowercase kebab-case. Attribute values support escaped
dotted-path interpolation. Icons inherit currentColor and are decorative by
default; labeled icons use role img.

Do not bundle the library. Phase-one dev resolves missing referenced icons
from pinned lucide-static 1.47.0 into `_/icons/lucide-1.47.0/` before serving,
including during reloads. Cache the upstream license too. Status is read-only
and reports missing icons as warnings without a failing exit code. Missing icons
or download failures render nothing and do not cause a 500 response. Malformed
directives and unsafe SVGs remain errors. Downloads are bounded, cancellable, and atomic;
SVGs are reconstructed from a strict geometry-only vocabulary. No runtime
browser library or request-time downloads. Deploy resolves icons before publishing.
See README.md for supported attributes.

### Snips

Snips are owner-editable sanitized HTML stored in private `_/snips.json`.
They are distinct from partials and structured developer data:

- A partial is a reusable developer template.
- A snip is owner content identified by a stable 10-character alphanumeric ID.
- Structured objects for loops live in `_/data/*.json`, exposed as `data.*`.

Create an editable region while dev runs:

```html
<h1 snip="title">Welcome to NanoWeb</h1>
```

Dev extracts inner HTML and inserts a stable reference:

```html
<h1 snip="title">{snip a7Km2pQ9xB}</h1>
```

Storage groups entries by route, with shared regions for layouts/partials:

```json
{"pages":{"/":{"title":"Home","description":"Welcome","snips":{"a7Km2pQ9xB":{"label":"title","content":"Welcome to NanoWeb"}}}},"shared":{"snips":{}}}
```

Extraction is dev-only, journaled and serialized. Content is persisted before
rewriting templates; interrupted work resumes or stops on conflicting edits.
Local serialization is in-process only, with no lock file; run one dev process
per project. Every relevant save extracts additions, reindexes labels/routes,
archives removals and restores reintroduced IDs. Existing content and metadata
are retained. Markers use {snip ID} without quotes; quoted forms are rejected.
HTML comments do not render snips. IDs contain only ASCII letters and digits.
Nested regions and template directives in extracted content are errors. Rendering
reconstructs a formatting/link allowlist and never interprets JSON as template
syntax. Rich output is only allowed in element content, not attributes or raw-text
elements. Page metadata and structured values remain escaped. Legacy named snips
remain supported for migration. Basic browser editing is implemented; managed images are deferred.

Deploy preserves remote rich content/metadata and seeds missing IDs. Content
snapshots live outside code releases and are switched by the activation journal.
Aliases pointing to the same release must select the same content snapshot;
conflicting paths are rejected and valid aliases share a loaded serving snapshot.
After successful activation, unreferenced entries move to an archive map, retaining
last remote content. References are lexical, including false/empty branches and
unused partials. Failed activation leaves previous content untouched. Protocol 7
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
{site.name}
{page.title}
{snip.about-us.title}

{#if site.phone}
  <a href="tel:{site.phone}">{site.phone}</a>
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

- Value lookup: dotted paths only, for example `{site.name}` and `{member.photo}`.
- Condition: a dotted-path truthiness check only, for example `{#if site.phone}`.
- Conditions support `{:else if dotted.path}` and a final `{:else}`. Loops support
  `{:else}` for empty or missing arrays, with outer aliases restored in the fallback.
- Loop: a collection dotted path and one alias only, for example `{#each team.members as member}`.
- An empty loop renders nothing.
- No function calls, arithmetic, comparisons, `.length`, filters, inline JavaScript, or arbitrary expressions in v1.
- Nanoweb does not parse template directives inside `<script>` or `<style>` elements.

Values interpolated by `{path}` are contextually escaped. Short-ID snips render
sanitized HTML by default; unrestricted raw HTML is not supported.

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

Nanoweb validates a snip before writing it. It writes content atomically: create a temporary file in the same directory, fsync as appropriate, then rename it over the existing JSON file. It retains prior revisions/snapshots for rollback.

## Configuration

`_/nanoweb.config` contains JSON only. Comments and trailing commas are not supported. The site loader reads the `site` object. Local `_/deploy.json` stores a `branches` map of confirmed deployment settings (target, domain, name, www_redirect, https, email, agree_tos, expected_ips). First deploy prompts and saves; later deploys select only the current branch. Changed settings require `-configure` confirmation. It contains no credentials and is excluded from uploads. Use branch names directly without environment labels or questions. Generate the internal site name automatically; staging suggests a subdomain. General page redirect settings remain deferred; WWW-to-primary redirects are supported.

Example:

```json
{
  "site": {
    "name": "Acme Catering",
    "email": "hello@example.com"
  }
}
```

Configuration should not contain private server credentials. The CLI stores deploy target credentials/settings locally outside version control, or uses standard SSH configuration and keys. The deploy target can be recorded in a `.gitignore`d local file if the developer wants convenience.

## Editor

The implemented editor is served at `/admin`, separate from public template routes.
Approved addresses are configured by the developer in `admin.emails` in JSON site
config. Login uses an emailed, six-digit, single-use code with expiry and attempt
limits, not passwords. Production requires HTTPS. Resend is the first provider,
behind a provider-neutral Sender interface; credentials live in a root-only server
EnvironmentFile and are prompted for during deployment when missing.

Implemented editor capabilities:

- Email-code authentication and secure host-bound session cookies
- List available snips
- Edit rich-snippet HTML in textareas and page title/description
- Publish validated content without restarting the service
- Validate before saving
- Serialize saves with deployment and reject stale forms

Visual rich-text controls, schema-defined fields, images and per-edit revision
history remain future work. The editor cannot change the email allowlist.

Editor exclusions for v1:

- Editing HTML, layouts, or partials
- Arbitrary filesystem browsing
- Arbitrary HTML/script injection
- Multiple-editor collaboration
- A visual drag-and-drop page builder

## Server and TLS

The installed Nanoweb server is a single Go executable and system service. A separate Caddy process is not part of the required architecture.

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
activation; interrupted transactions can be restored with `nanoweb-server -recover`.
Old releases are retained, with no pruning or rollback CLI yet. Committed rich
snips seed missing IDs; existing remote content/metadata wins. Content snapshots
live separately under content/<release-id>.json, selected by the host config.
Successful activation archives/prunes unreferenced IDs; failed activation restores
the prior code and content paths. The rollback workflow below is future work.

Each deployment creates an immutable release. The host switches to it atomically.

```text
/var/lib/nanoweb/sites/acmecatering.com/
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
- On deploy, Nanoweb preserves existing owner-edited snips and uploads unless the developer intentionally runs an explicit content migration/reset command.

This avoids accidentally erasing an owner’s edits every time a developer redeploys HTML.

Required deployment properties:

- Validate routes, templates, snips, and configuration before activating a release.
- Upload a staged release before changing `current`.
- Switch releases atomically.
- Preserve the prior active release.
- Provide `nanoweb rollback`.
- Report the deployed domain, release identifier, and any certificate state clearly.

## CLI surface

Keep the initial command surface small:

```text
nanoweb new <site-name>
nanoweb dev
nanoweb setup
nanoweb deploy
nanoweb status
nanoweb logs
nanoweb rollback
```

### Command behavior

`nanoweb setup`

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
  only the verified NanoWeb service during reinstall. Preserve unrelated services.
- Initial host listens publicly on HTTP port 80. Deploy and HTTPS are subsequent
  explicit commands; no firewall changes are made.

`nanoweb new <site-name>`

- Creates the standard folder structure and a polished minimal starter site.
- Initializes Git on main without making a commit. --skip-git skips initialization;
  missing Git does not prevent site creation. Includes a starter .gitignore.
- Creates default `_/nanoweb.config`, a default snip, a layout with inline header/footer, error pages, and a minimal Nanoweb welcome page with its SVG mark.
- Bundles local Prewind 1.3.1 and Preflight-inspired theme, typography, native form, button, and application styles. CSS layers are ordered reset, styles, prewind. No CSS build step or CDN is required; form styling does not implement submission handling.

`nanoweb dev`

- Must run without Docker or a database.
- Watches source files and snip content.
- Logs requests as `GET /test` for 200 or `404 GET /missing` for non-200.
  Include assets and HEAD, but omit successful internal reload polls and query strings.
- Runs a local server with the same route and rendering behavior as production.
- Reports template errors, invalid route collisions, missing partials, schema errors, and unsafe/unserved paths clearly.
- Provides local editor access.

`nanoweb deploy`

- Before deployment, offer to commit pending changes with a file preview,
  confirmation, and commit message. Respect Git hooks and signing. Declining or
  failure stops deployment. Sites within larger repositories commit manually.

- Requires deployment protocol 7 before upload, offering first-time setup when absent. Validates the
  packaged site, uploads over SSH/SCP, stages, and atomically switches host config.
- Restarts and verifies revision/hostname; restores the prior config on failure.
- Uses native SSH and the saved current-branch mapping in _/deploy.json. The first
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

`nanoweb https`

- Enables Let's Encrypt certificates for existing domains. Uses autocert for
  issuance, cache and renewal, with exact host allowlists and TLS Host/SNI checks.
- Requires DNS pointing at the server, inbound TCP 80/443, contact email and terms
  acceptance. Redirects HTTP to HTTPS for enabled domains; health stays on HTTP.
- Tests use a local ACME authority; real Ubuntu/live CA validation is pending.

`nanoweb status`

- Shows local project validity or remote deployment/certificate/release status, depending on flags/context.

`nanoweb logs`

- Streams or retrieves host logs without requiring the developer to know system-service commands.

`nanoweb rollback`

- Lists/selects or defaults to the immediately previous release, then atomically reactivates it.

## Docker

Docker is not required for local development or the first production path. The core product must work on a blank Ubuntu LTS host using the Nanoweb executable and OS service management.

Later, Docker can be an optional deployment target for users who already operate containers. It must not dictate the local project format or force every customer to understand containers.

## Server-side features

The first server-side feature should be a constrained contact-form action, not arbitrary application code.

A form can declare a named Nanoweb endpoint and validation configuration. Nanoweb handles rate limiting, spam defenses, payload validation, delivery/storage, and logging. The implementation may later offer email or webhook destinations, but those credentials live in host configuration, never in public site files.

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

### Milestone 1: local static Nanoweb

- CLI skeleton and `nanoweb new`
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
- Host-level users/sessions/site registry in SQLite

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
2. **Public URL prefix for snip-managed images:** use a stable virtual prefix and do not expose `_/` directly.
3. **Rich text model:** sanitized HTML versus a constrained Markdown-like format.
4. **First form destination:** email, host inbox, webhook, or a combination.
5. **Editor user model:** one site-owner account initially versus host-managed multiple editor accounts.
6. **Default-content deploy policy:** preserve live editor changes by default, with explicit content migration/reset commands.

## Acceptance criteria for the first usable release

Nanoweb is meaningfully usable when a developer can:

1. Install the CLI through Homebrew.
2. Run `nanoweb new my-site` and receive a working starter site.
3. Run `nanoweb dev` with no language runtime, Node modules, database, Docker, or external web-server setup.
4. Add `about-us/history.html` and see it at `/about-us/history`.
5. Confirm `/about-us/history.html` returns 404.
6. Reuse a developer partial in several pages.
7. Wrap pages with `_layout.html` using `{slot}`.
8. Render values, simple conditions, and loops from a JSON snip.
9. Edit a schema-defined snip and associated image from a protected editor without editing source HTML.
10. Deploy to a blank supported Ubuntu server with one command after entering the initial domain/server details.
11. Receive a valid automatically managed HTTPS certificate.
12. Deploy a second release and roll back safely.
13. Host more than one domain/site from the same Nanoweb server.
