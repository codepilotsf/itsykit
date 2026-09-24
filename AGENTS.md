# Working on FleaCMS

## Keep this guide current

Update this file as part of any change that affects FleaCMS's architecture,
template language, CLI, starter conventions, supported features, or development
workflow. Update the relevant sections in the same task, remove stale guidance,
and keep this a practical guide rather than a running changelog. No separate
reminder from the user is needed. Keep README.md and applicable sections of
spec.md consistent with behavior changes too.

Keep `next-steps.md` organized as the working roadmap: priorities, open decisions,
verification still needed, and deferred work. Update it as tasks are completed
or priorities change. Remove stale notes; use CHANGELOG.md for completed changes
and README.md for usage rather than accumulating implementation history there.

## Product and scope

### Backward compatibility policy

FleaCMS is early in development. Do not add backward compatibility, legacy
aliases, format fallbacks, migration layers, or support for superseded behavior
by default. Prefer the current design, even when it introduces a breaking change.
Ask the user before adding any backward compatibility layer. This applies to
commands, configuration, templates, stored data, deployment, and server protocols.
Existing legacy support is not a precedent for adding more.

FleaCMS is a small website server built with Go 1.26 or newer, the standard
library, and pinned golang.org/x/crypto/acme/autocert for managed TLS. Sites use plain HTML, local assets, and JSON content. Keep the design
simple and avoid adding dependencies or a frontend build system without a
concrete need.

Currently implemented: `flea new <site-name>`, `flea status`,
`flea dev [-port 3000]`, `flea setup`, `flea version`,
`flea deploy`, `flea admin [email]`, and `flea https`.
`flea email` configures local Resend delivery interactively.
Run status/dev from a generated site. Setup runs from the Mac over native SSH.
Bare `flea`, `flea help`, `flea -h`, and `flea --help` show the grouped command
overview on stdout and exit successfully. Each command supports `--help` and
`-h` without requiring a site or starting work. Keep the overview spacious;
put flags and deployment requirements in command-specific help.
Dev uses pinned github.com/fsnotify/fsnotify directory notifications with a
100 ms debounce, not periodic file scans. Register included subdirectories,
reconcile new/renamed/deleted directories, and keep atomic saves working.
Fingerprint only at startup or after events to suppress self-generated writes.
On overflow rescan; on watcher failure stop dev with an error. No polling fallback.
Browser reload polling is separate and unchanged. Preserve invalid-edit recovery.
The dev server listens on loopback; the persistent host defaults to public HTTP
port 80, with opt-in automatic HTTPS on port 443 per domain.

README.md describes the current implementation. spec.md also contains future
plans: do not assume its rollback CLI, general-purpose forms, schemas, content revisions,
or managed images are implemented. Rich fleas use sanitized HTML; unrestricted
raw HTML rendering is not implemented.
Verify current code and tests when documentation disagrees.

## Repository map

Read `ARCHITECTURE.md` for dependency direction, trust boundaries, transaction
ordering, and review follow-ups. Keep it current when those responsibilities change.

- `cmd/flea/main.go`: executable entry point.
- `cmd/flea-server/main.go`: standalone persistent server entry point.
- `internal/host/`: strict server JSON config, hostname routing, readiness, and shutdown.
- `internal/setup/`: interactive SSH bootstrap, installer, and systemd unit.
- `internal/serverbundle/`: compressed embedded Linux servers and verified extraction.
- `internal/setup/deploy.go`: native SSH/SCP deployment and local deployment settings.
- `internal/release/`: bounded, confined runtime-only archive packing and extraction.
- `internal/admin/`: privileged release activation, lock, recovery journal, and probes.
- `internal/editor/`: owner dashboard, in-memory OTPs, persistent rolling sessions,
  CSRF, and the hash-pinned activity script. Page markup lives in templates/admin/.
- `internal/editstore/`: conflict-checked content saves and validated snapshot publication.
- `internal/mailer/`: provider-neutral Sender and Resend HTTP adapter; never expose provider errors or secrets.
- `internal/state/`: the server-side file lock shared by browser saves and deployment.
- `internal/cli/cli.go`: command handling and site creation.
- `templates/`: canonical starter files, mirroring the generated site structure.
- `templates/embed.go`: embeds HTML, CSS, SVG, JSON, and README files in the binary.
- `internal/cli/cli_test.go`: CLI tests and executable/server smoke test.
- `internal/site/project.go`: loading, validation, routes, snapshots, and rendering.
- `internal/site/routes.go` and `data.go`: shared route conversion and JSON-directory loading.
- `internal/site/regions.go`: source-preserving region scanner and lexical references.
- `internal/site/extraction.go`: dev-only `LoadEditing` and flea extraction orchestration.
- `internal/site/flea_transaction.go`: content-first extraction journal and replay.
- `internal/site/files.go`: pinned project reads and durable atomic writes, also used by icons.
- `internal/admin/content.go`: remote content merge/staging; activation retains the deployment lock.
- `internal/setup/ssh.go`: shared SSH/SCP connection policy and target formatting.
- `internal/release/protocol.go`: shared CLI/server archive compatibility marker.
- `internal/site/template.go`: template parser, partial resolution, and data lookup.
- `internal/site/icons.go`: on-demand pinned Lucide downloads, private cache,
  geometry-only SVG validation, and icon directive compilation.
- `internal/site/icons_test.go`: offline icon resolution, escaping, and security tests.
- `internal/site/server.go`: HTTP serving, reload watching, and URL validation.
- `internal/site/site_test.go`: rendering, routing, isolation, and reload tests.
- `bin/flea`: local build output; rebuilding is necessary for CLI changes to take effect.

## Sites and templates

A generated site has the project root for templates/assets and `_/` for configuration
and content. `_/flea.config` is JSON and requires a nonempty `site.name`.
Owner-editable fleas live in _/fleas.json: pages map canonical routes to title,
description, and fleas; shared.fleas holds reusable regions; archive holds removed
IDs. Entries have content and an optional label. Bare flea attributes need no
name or value. IDs are 10 case-sensitive alphanumeric chars,
crypto/rand with rejection sampling and active/archive collision checks.
Generate {flea ID} without quotes. Quoted markers are rejected; HTML comments
are ordinary comments and never render content or count as references, except
the leading flea page-settings comment described below.
Reject old snip attributes, directives, content paths and recovery journals;
never silently ignore or migrate them. Protocol 10 enforces the renamed format.
Validate rich references identically in the parser and region scanner;
IDs cannot contain braces, whitespace or punctuation.
Developer-managed structured objects live in _/data/*.json and use data.*.
Legacy _/fleas/*.json and escaped named directives remain supported.

LoadEditing (dev startup/watch only) extracts inner HTML of bare flea regions,
preserving surrounding source. Nested regions, malformed nesting, void/raw-text
regions, and directives in new content are errors. Existing IDs retain content;
labels/routes are reindexed on dev saves, cross-route/layout/partial IDs are shared.
Use the in-process LockContent mutex through local writes and publication; acquire
it before Server.reloadMu. No local content filesystem lock; multiple dev processes on
one project are unsupported. No legacy lock cleanup or automatic migration.
On every relevant save, archive unreferenced IDs and restore reintroduced IDs,
preserving content and metadata. Log successful flea updates. Persist
_/.fleas-pending.json with before/after
templates and content before saving content, then templates. Replay on restart;
stop if independently edited files conflict. Status and deploy never extract.
Deploy refuses unextracted regions. Do not ignore recovery records.

Reconstruct rich HTML with an allowlist using x/net/html. Never compile JSON into
template source. Only sanitized values become template.HTML through richFlea at
execution; require element-text contexts, reject attributes/raw-text contexts.
Script/style source remains untouched. Block scripts, styles, SVG/MathML, forms,
embeds, images, arbitrary attributes, and unsafe links. Page metadata and data.*
remain escaped. Flea JSON is limited to 5 MiB and each content value to 1 MiB.

Use internal/sitefiles as the shared exclusion policy for loading, reloads,
Git snapshots, and archives. Exclude dotpaths, node_modules, and project metadata
(README.md, AGENTS.md, ARCHITECTURE.md, spec.md, next-steps.md, LICENSE, LICENSE.md, go.mod, go.sum).
Only config, fleas, structured data, and icons inside _/ enter release snapshots;
private session runtime files do not. No legacy html/
document-root fallback exists. Existing sites require an explicit, collision-checked
move of html/ contents to the root. Deployment protocol 10 prevents old servers
from receiving incompatible root-level archives.

`flea new` initializes Git on main without staging files or creating a commit.
Support --skip-git before or after the site name for existing repositories.
If Git is unavailable, still create the site and explain how to initialize Git
before deployment. Embed the starter .gitignore explicitly; retain deployment
settings and icon caches in version control by default.

- `_layout.html` is the optional global wrapper; `{slot}` inserts the page.
- A leading `<!-- flea` comment with `layout: /_layout_admin.html` on its own
  line and a closing `-->` selects a custom layout. Use one block before markup,
  root-relative private .html paths, and only supported key:value settings.
  Reject unknown/duplicate keys, missing/unsafe layouts, nested layouts and settings
  in partials. Layouts support {slot} and relative partials, cannot be included as
  partials, and stay private. _500.html remains standalone. No YAML dependency.
- The starter's header and footer live directly in `_layout.html`. Do not
  generate separate `_header.html` or `_footer.html` files by default.
- The starter is a minimal FleaCMS welcome page. Its logo is the selected
  option C (Skip): a happy side-profile flea with a folded jumping leg, in
  slate-800 rounded line art. templates/images/mark.svg is the shipped logo,
  shared by the header and favicon. Keep the selected design unchanged.
- Optional partials are underscore-prefixed HTML files anywhere under the project root.
  There is no special `_partials/` directory.
- `{render footer}` resolves `_footer.html` beside the calling template.
  `{render /foo/bar/footer}` resolves `foo/bar/_footer.html`.
  Paths are unquoted and omit the filename's underscore and `.html` extension.
- Relative paths, including `./` and `../`, resolve from the calling template,
  including inside partials. Paths cannot escape the project root. Missing and recursive
  includes fail validation, including in unused partials.
- Root `_layout.html`, `_404.html`, and `_500.html` are reserved, not partials.
  `_404.html` uses the global layout; write it as page content for `{slot}`.
  `_500.html` remains standalone so error handling does not depend on the layout.
- Data uses dotted paths such as `{site.name}` and `{data.home.title}`.
  Conditions use `{#if site.email}...{/if}`; loops use
  `{#each data.home.items as item}...{/each}`. Includes inherit loop aliases.
  Support `{:else}` in if/each and `{:else if dotted.path}` in if only.
  Each fallbacks run for empty/missing arrays and restore any outer shadowed
  alias; no current item exists there. Reject duplicate else or branches after
  else, and retain dotted-path-only conditions and contextual escaping.
- `{flea a7Km2pQ9xB}` renders sanitized owner HTML. Legacy `{flea notice}`
  renders the escaped text field of a named flea. Preserve
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
Deploy also uses LoadDev to resolve icons before packaging. No standalone site
build command exists yet; a future build should reuse this resolver.
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
put `container` on an inner wrapper. Do not add `max-w-full` alongside container;
container already handles its width. Avoid wrappers duplicating its purpose.
Use global textarea `field-sizing: content`, a one-field minimum and 400px maximum
height with overflow scrolling. Do not add fixed rows or autosizing JavaScript.

The starter loads Prewind 1.3.1 from unpkg, based on the Preflight template's
global-first CSS conventions. No Node build step or Rails dependency is required.
The main stylesheet is `css/global.css`, with layer order `reset, styles,
prewind`. It imports Prewind, theme tokens, typography, native form styles, and
button styles. Put global rules in `@layer styles` so utilities can override them.
The first line of `theme.css` imports Google Fonts: Manrope for headings and
Source Sans 3 for body text, with display=swap and system fallbacks. Customize
colors, fonts, and fluid scales in `theme.css`; site rules belong in `global.css`.
Form styling does not implement submission handling.

Keep the vendored `prewind-1.3.1.css` unchanged; customize the other files.
The theme and general styles are adapted from `../preflight/lib/templates/app/assets/stylesheets/`.
The library matches `../prewind/prewind.css` at version 1.3.1. Preserve the simple
starter with its inline header/footer when making styling changes.

## Routing and isolation

Email-only contact forms live in internal/contact. Configure contact.to in site
JSON; POST /_flea/contact accepts only name/email/message/website. Bind handlers
and recipients to the loaded site, never request-selected IDs or addresses.
Aliases share limits; distinct roots must not share a content file, including
hardlinks. Keep exact Origin checks, production HTTPS, bounded bodies, per-site/IP
limits and a shared server contact budget. Visitor email is Reply-To only; send
plain text through mailer.Sender. Do not store submissions or queue work.
Local contact settings require restarting dev. Deployment configures mail for
contact-only sites too. See README for limits and setup. Protocol 10 requires
the contact runtime. Test cross-site recipients and aliases without live email.

`index.html` maps to `/`. `about.html` and `about/index.html`
both map to `/about`; providing both is a route collision. Explicit `.html`
URLs and trailing slashes other than `/` return 404. Assets retain their paths.

Files and directories beginning with `_` or `.` are private. Preserve URL
validation, confined reads, symlink rejection, and immutable asset snapshots.
Treat all Project maps and Content as immutable after publication. Build a new
Project for reloads. Keep project-file confinement separate from administrator-owned
absolute-path writes; their trust assumptions differ.
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
the framework source. When the user requests a FleaCMS feature or starter
change, implement it here, even if the session begins in Bookworm.

For intentional starter changes, update the ordinary files in `templates/` as well as
the example site in scope. Do not overwrite unrelated site customizations.
Generate a fresh site in a temporary directory and compare its relevant files
with the intended result. Keep local agent instructions out of that comparison
unless they are deliberately part of the generated starter.

## Development checks

The isolated inline-editing experiment lives in `design/inline-editor/`.
Run `go run ./design/inline-editor` for its loopback-only preview. It uses pinned
Squire and DOMPurify with simulated identity and tab-only saves, not real site
content. Keep production editing unchanged while evaluating it. See its README
for browser checks and integration gaps, especially H1/alignment sanitization.
The prototype edits the current page with a right-aligned Edit toggle, not a page
menu. Edit becomes Cancel/Save; Sign out stays visible after a divider.
Page details and contextual formatting
appear in a second bar. Unlinked-page discovery is deferred. Serve the selected
local Lucide SVGs without adding npm or a frontend build step.
Prototype dialogs use compact titles, a bold text-link-style Cancel button on
the left, and the primary action on the right. Cancel discards dialog changes.
Keep the second bar visible throughout edit mode, hidden while browsing. Formatting
is disabled without an active text region; Page details stays enabled. No content
outlines except keyboard focus. Modal motion and page-content view transitions
are subtle and respect reduced motion.
Keep dialogs open until their exit animation completes, including Cancel, Escape
and submit. Page crossfades use explicit final-opacity fill and hide the old root
snapshot; rapid navigation finishes/skips the prior transition before starting anew.
Undo/redo toolbar buttons are deferred; keyboard history remains available.
The link button opens the current URL for editing when inside a link, with a
separate Remove link action in the dialog. Preserve link text and formatting.
Prototype links default bare domains to HTTPS and preserve explicit schemes and
relative paths. The new-tab option defaults off and stores only _blank with
noopener noreferrer. Production sanitizer support remains an integration task.
Reusable CSS-only switch visuals belong in templates/css/ui.css, targeting native
checkboxes with role="switch"; preserve keyboard, focus and reduced-motion behavior.

### Versioning and changelog

Use Semantic Versioning. The current source release is 0.2.0. The canonical version is
`internal/buildinfo/version.go`; local builds and both server bundles use it.
Local VERSION overrides must be valid SemVer without a leading v. Release tags
must be v plus the recorded version; release and publish scripts validate this.
Keep user-facing changes in CHANGELOG.md under Unreleased until preparing the
next version, then move them into a dated version section. During 0.x, use patch
versions for fixes and minor versions for features or breaking changes. This
does not change the policy requiring approval before adding compatibility layers.
Do not create or push release tags or publish artifacts unless requested.

### Remote setup

Setup prompts for a hostname/IP (also user@host or an SSH alias), verifies SSH,
Ubuntu LTS 22.04/24.04/26.04 and amd64/arm64, then extracts a bundled server.
Use the native SSH/SCP clients for agents, keys, host verification, and interactive
passwords. Do not store passwords or disable host-key checks. Standalone setup
always connects as root, even if a username is supplied. Deployment-triggered
setup retains its saved target. Non-root installs use sudo with a remote PTY.
SSH/SCP use StrictHostKeyChecking=accept-new: trust new host identities without
a prompt, but reject changed/revoked keys. Preserve SSH diagnostics and add a
friendly explanation. Never remove or replace known_hosts entries automatically.
The temporary control connection and staged files are cleaned up on completion.

No Go or source checkout is needed on either machine. Extract the matching gzip
server from the CLI, verify its bundled SHA-256 and static Linux ELF architecture,
then upload. Limit extraction to 100 MiB, honor cancellation and remove failed
outputs. No network/compiler fallback. Never import serverbundle from the server.
Build with bash scripts/build.sh: refresh both CGO-disabled Linux variants before
embedding into the native CLI. Plain go build may embed stale assets. CI still
builds/tests on native Linux/Mac runners; see RELEASING.md.
Setup refuses unexpected existing
installation paths/accounts and occupied port 80, and does not change SSH,
firewalls, packages, or existing sites. It installs a dedicated non-login service
account, hardened systemd unit, and JSON config. Failure cleanup preserves data
and accounts. A recognized binary/config installation offers a destructive reset
requiring typed DESTROY confirmation, only after extracting and verifying the
replacement binary locally. Move canonical paths to a root-only backup
under /var/backups before recreating an empty installation; refuse symlinked
paths and retain the service account. This is not an in-place upgrade.
Before any reset/install changes, check TCP/UDP 80 and 443 on IPv4/IPv6.
Only confirmed reinstall may exempt the systemd MainPID whose /proc executable
is /usr/local/bin/flea-server. Reject unknown owners, other PIDs, and failed
checks. Never stop another service. Recheck after stopping FleaCMS.

The initial persistent host binds :80, uses /etc/flea/server.json,
and has no sites by default. Config maps hostnames to absolute site roots.
It uses site.NewProductionServer: no development reload script/endpoint, watcher,
or icon fetching. Config/site edits require restart. TLS is opt-in per domain.
Aliases for the same release must select the same content snapshot and share one
loaded handler. Reject conflicting alias content paths instead of choosing arbitrarily.
The non-root service receives only CAP_NET_BIND_SERVICE to bind low ports.
Readiness probes wildcard listeners via loopback before setup succeeds; this
does not verify public firewall access. Administrators must allow inbound TCP 80.
Tests mock SSH/SCP, verify both Linux ELF builds, and exercise HTTP locally.
Do not claim real Ubuntu/password/systemd installation was tested without a target.

### Deployment and HTTPS

Deploy/https select only the current Git branch's saved destination from
_/deploy.json's branches map. Use the actual branch name without a production
label or environment question. Generate the internal remote site name from the
project folder and branch; do not ask the user to name it. First use prompts,
suggests staging.<main domain> for staging, and saves server, domain, stable
branch-specific site name, WWW redirect, HTTPS (default true), and contact/terms.
Combine HTTPS enablement and terms acceptance in one question, showing the
Let's Encrypt terms link first. Save both flags from that answer; do not ask a
second terms question. Saved mappings skip these questions on future deploys.
Default the certificate email to Git user.email only when no saved, inherited,
or explicitly supplied email exists. Keep the prompt editable; do not change Git config.
Save validated settings without a final save-confirmation question. Deployment
shows the branch/domain and Ctrl-C notice, then a success URL after activation.
Do not advertise a rollback command until it is implemented and tested.
Ask for the SSH username (default root) only when the target does not include
user@host; save the explicit username. Existing bare targets ask only for the
missing username. Do not ask about NAT/proxy IP lists during normal setup; retain
expected_ips as an advanced JSON/flag setting. Saving occurs before the clean check; the user must
commit the config or explicitly ignore it. The wizard can run before the initial
commit so settings and site files can be committed together. Upload still requires
a commit and clean status. Non-Git and detached HEAD cannot configure deployment.
Use -configure to confirm and save changed settings; CLI flags cannot silently
override an existing mapping. Flat legacy settings require branch confirmation.
Reject hostname or same-server site-name collisions across branches. The server
records site_branches and commits and enforces ownership too.

Require clean Git status (staged, unstaged, untracked, submodules) before SSH,
but offer to commit pending changes first: display status, obtain confirmation,
ask for a commit message, then stage and commit. Default to init only without a
HEAD commit; otherwise require a nonempty message without a default. Never bypass hooks/signing or
commit during later deployment checks. Subfolder sites require manual commits.
The commit helper runs before the deployment workflow; it never pushes.
Require clean Git status again before
upload, and activation; fail if the branch/commit changes during deploy. Build
the snapshot from immutable Git objects, including when the site is a subfolder.
Resolve missing icons in the temporary snapshot, never the working directory.
This JSON file contains settings only, never credentials, and is not uploaded.
Deploy packages committed root-level pages/assets/partials, _/flea.config,
_/fleas.json, _/data/, legacy _/fleas/, and _/icons/ only. Validate
the actual packaged snapshot locally and the extracted release remotely.
Reject symlinks, hard links, special files, duplicate/traversing archive paths,
and releases exceeding 100 MiB or 10,000 entries.

Deployment administration runs through SSH/sudo via the installed server binary,
never HTTP. The browser editor has content-only permissions, not deployment powers.
Protocol version 10 requires persistent-session runtime directories, the contact
runtime and unquoted flea syntax.
Setup does not upgrade existing installations.
If an authenticated server probe confirms that FleaCMS is absent, offer setup
and continue deployment after a successful install and fresh server check. Reuse
the deployment SSH connection and target. Do not offer installation after failed
SSH authentication or when an old/incomplete server executable already exists.
Keep SSH errors friendly and separate connection failure from missing software.
Before upload invoke the server's unprivileged -check-dns. Resolve every domain
and WWW alias using the server resolver and fail on lookup errors, no answers,
or any mismatched A/AAAA address. Expected IPs default to non-loopback server
interfaces; explicitly saved expected_ips supports NAT/proxies. IP destinations
skip DNS. No public reachability guarantee is implied by this resolver check.
Use fresh releases under /var/lib/flea/sites/<name>/releases/<id> and absolute
config paths, without current symlinks. Move all aliases of a site together and
refuse hostnames owned by another site. Preserve unrelated sites and old releases.
Activation locks operations, journals the prior config, atomically replaces
server.json, restarts, and probes the active revision/hostname. Failure restores
the prior config; interrupted operations require flea-server -recover.
Retain releases and server.previous.json; rollback CLI and pruning are deferred.
Rich fleas are committed seeds only. Preserve remote content by ID and remote
page metadata, seed new IDs, and restore reintroduced IDs from archived content.
Store content outside releases in sites/<name>/content/<release-id>.json;
host.Config.Content maps hostnames to absolute snapshots. Activation journals
switch both paths; failed activation leaves old content intact. After successful
verification, prune unreferenced active IDs into archive, retaining prior snapshots.
Scan lexical references in all templates, including unused partials and false/empty
branches; never prune based on rendered output. No archive pruning or rollback CLI.
Code/structured-data deploys still replace snapshots. Browser saves share the
deployment lock, check the active release and content version, and publish a new
validated Project. Manual remote content edits need restart and must not overlap deployment.
WWW redirects use 308 with preserved path/query and certificates for both names.
Saved HTTPS=false removes this site's TLS allowlist entries on redeploy but
retains the certificate cache; other sites remain unchanged.

Use autocert for Let's Encrypt issuance and renewal, with exact configured
domain allowlists including cached certificate checks. Require a contact email
and explicit terms acceptance. No IP/wildcard certificates or custom production
ACME endpoints yet. Keep HTTP :80 and HTTPS :443 for ACME, redirect enabled
domains with 308, preserve path/query, and keep generic readiness on HTTP.
Reject TLS Host/SNI mismatches. Tests use a local CA without weakening production
verification. Keep certificates in /var/lib/flea/certs (flea-owned 0700);
the data parent and releases are root-owned. Only certificates, managed content
and private runtime directories and the precreated deployment lock are writable
by the service.
Never include keys in archives or logs.

### Owner login and editing

Admin markup lives in developer-editable admin/index.html, code.html,
dashboard.html, setup.html, and error.html, shipped by templates/. Compile these
with the selected _layout_admin.html and render escaped admin.* data through
Server.RenderAdmin against the current immutable snapshot. Never compile editor
values as source. Exclude the entire admin directory from public routes/assets;
only editor state selects these templates. Missing files have no embedded fallback.
Without a layout, render the page as written. Preserve developer layout changes.
Use FleaCMS as the product name, without a space. Dashboard navigation lists all
public routes, including pages without fleas, with / labeled Homepage. Sidebar
links use /admin?page=<canonical route>, showing only that page's metadata and
fleas. Shared content uses /admin?shared=1. Unknown routes return 404. Saves keep
the selected page. The full-width, footer-free admin shell pins the sidebar at
the top left below its header; menu and editor scroll independently. On narrow
screens navigation sits above content. Dashboard CSS lives in css/global.css.
Admin CSP permits HTTPS styles/fonts/images and only the SHA-256-pinned built-in
activity script, plus same-origin fetches. Other scripts, frames, and cross-origin
forms remain blocked. Inject the exact script bytes on authenticated dashboard
pages only; never interpolate template or owner values into JavaScript.
Referrer-Policy must be same-origin,
not no-referrer: native browser POSTs otherwise send Origin: null. Keep exact
Origin matching and CSRF validation; never accept null or missing origins.

Owner onboarding uses a developer-shared /admin link, not invitation emails or
separate account registration. Use "site owner" in developer and owner guidance.
The CLI explains applying access changes, sharing the live link, local mail setup,
and viewing/removing addresses via admin.emails. Local /admin shows a setup page
even with no owners; production remains disabled with an empty allowlist.
Missing-mail pages use the editor layout; locally direct developers to flea email
and returning to /admin, while production keeps an owner-facing message.
Keep request-code responses identical for approved and unapproved addresses.
Explain that each Save publishes one form and other unsaved fields are discarded.

flea admin [email] appends to admin.emails, prompting when the argument is
omitted. Use the shared AdminAccess validator, skip case-insensitive duplicates,
and retain unrelated JSON fields with RawMessage and confined atomic writes.
Do not send email, commit, or deploy. Check effective local mail settings and
run the email wizard if missing or invalid. Never replace malformed configuration.
Keep output terse: confirmation, /admin, and the live commit/deploy reminder.
Dev uses the current validated owner list and rereads local mail per admin request.
Rebuild the local editor on access, idle-window or mail changes, discarding pending
codes but retaining valid sessions. Prune removed owners and expired sessions.
Production applies configuration on startup/deploy. Contact settings still require restart.
The email wizard asks for a hidden API key, sender name,
and separate verified-domain from address, then checks Resend without sending.
Inspect only safe provider error codes; never expose response bodies or secrets.
The documented 401 restricted_api_key response indicates a sending-only key;
403 restricted_api_key does not indicate success. Never claim delivery was tested.
Save only after a successful check, preserving unrelated .env settings, using
0600 permissions and a final /.env Git ignore rule. Refuse tracked .env files.
Successful email setup prints only the saved-file confirmation after checking.
Keep permission limitations in documentation, not routine success output; never
claim that the connection check verifies the sender domain or email delivery.
Local /admin must show actionable environment setup instructions when email is
unconfigured; production keeps an owner-facing message without local shell steps.

`admin.emails` in _/flea.config is a developer-owned list (max 20), empty by
default. The editor at /admin supports email OTP, page title/description and rich
flea HTML-textarea edits, not templates, schemas, images, archived entries or settings.
Codes use crypto/rand, expire in 10 minutes and allow five attempts. Store hashed
codes in bounded expiring maps. Sessions use admin.session_idle_days (omitted/zero
means 30, otherwise 1–365). Persist only token hashes, email, host, CSRF, last activity
and expiry. Store up to 2048 sessions in a bounded 2 MiB JSON file, 0600, at local
_/sessions.json or production sites/<name>/_/sessions.json outside releases.
The production runtime directory is flea-owned 0700; site parents/releases remain
root-owned. Use confined reads, reject links and unsafe permissions, and lock
read-modify-write with .sessions-lock. Sync an exclusive private temporary file,
rename, and sync the parent. Never silently reset corrupt/unwritable storage.
Each transaction rereads disk; stale handlers cannot restore revoked sessions.
Keep session locking separate from the content/deployment lock.
Prune against current owners and expiry at startup, local handler rebuild and
session access. Host.Load is read-only, including candidate validation; actual
host startup initializes/prunes session state. Deployment creates runtime directory
permissions without copying sessions. Sessions survive restart and release changes;
logout revokes the current token, owner removal revokes their tokens.
Authenticated page views/forms renew cookie and disk expiry. The activity script
renews on trusted visible input/click/key/scroll events, at most once per minute,
never with an idle timer. Its POST requires the same Origin and CSRF as saves.
Do not require email delivery to remain configured for already authenticated users.
Dev ensures session ignore rules before creating state, refuses Git-tracked files,
and status reports the local path. Runtime inclusion excludes sessions/lock/temp
files from watchers, snapshots and archives; underscore URL privacy still applies.
Require exact Origin and CSRF checks on every POST, host-bound HttpOnly/SameSite
cookies, HTTPS and __Host- cookies in production. Local dev permits loopback HTTP
only, with no login bypass. Do not trust forwarded headers. Rate-limit requests by
address, actual client IP and site. Do not reveal allowlist membership or log codes.

Resend uses mailer.Sender, not an SDK-specific dependency in auth code. Delivery is
asynchronous with a timeout, and bounded by the send limits; errors never contain
provider bodies, tokens or recipients. Deployment prompts for admin-enabled or contact-enabled
sites lacking saved mail settings. Stream secrets over SSH without a PTY into a
private staging file; sudo may then consume it with an interactive terminal.
Store /etc/flea/mail.env root-only (0600), loaded by systemd EnvironmentFile.
No secrets in site config, deploy settings or Git. Local dev and deployment mail
setup use localenv for bounded literal .env reads; process environment values win.
Never execute or expand .env values, mutate the global process environment, or
load .env into the production host. FLEA_EMAIL_NAME supplies the sender display
name. Keep Sender and setup Checker interfaces separate from the Resend adapter.
Admin addresses and sign-in mail settings update without a dev restart.

Protocol 10 also requires the service unit and lock/content/runtime permissions, not
just a binary swap. Do not reset a live installation as an upgrade. Setup creates
the lock owned by flea, and activation grants write access only to each content
and private runtime directory. Keep release/config parents root-owned. Browser edits and extraction
share an in-process mutex; remote edits and activation share deploy.lock and refuse
pending recovery journals. Test stale forms, single-use codes, CSRF, isolation,
sanitization, content preservation and concurrent requests. Live email delivery
requires real credentials and remains unverified unless actually exercised.

Tests cover real local HTTP/TLS, a local ACME issuance flow and offline cache
reuse, activation failure/recovery, archive attacks, and simulated SSH/SCP.
These do not establish live Let's Encrypt issuance, long-term renewal, or real
Ubuntu/systemd deployment. Record those as pending until exercised on a target.

### Checks

Inspect the current worktree before editing and preserve unrelated changes.
Format changed Go files with gofmt. For implementation changes, run:

```sh
go test ./...
go vet ./...
bash scripts/build.sh
```

Use `go test -race ./...` for concurrency, reload, or shared-state changes.
The regular suite includes `TestExecutableWorkflow`, which creates a fresh
site, starts a local server, checks edits/recovery, and shuts it down. It needs
loopback binding permission; a sandbox bind failure is not a product failure.
Add meaningful regression coverage for changed behavior. Documentation-only
changes do not need the runtime suite.

On the current Mac, Go is available at `/usr/local/go/bin/go`. If the default
build cache is restricted, use `GOCACHE=/tmp/flea-go-cache`. This checkout is now
at `/Users/sean/Projects/flea`.
`/usr/local/bin/flea` points to `/Users/sean/Projects/flea/bin/flea` and was
verified with `flea version`. See README for the installation command.
Rebuild after implementation or starter changes,
then run `flea status` in the example/generated site. Restart an existing
dev process when the executable changes; site-file reloads do not reload Go code.

## Communication

Always write developer-facing CLI output in simple, friendly English. Use line
breaks and blank lines for readability, especially between an explanation and
the next command. State what happened and what to do next. Avoid raw subprocess
commands, exit codes, or Git jargon in routine errors. Keep messages brief without
removing useful context needed to fix an actual failure.

Use a natural voice. Avoid AI cliches, emojis, and em dashes. Question assumptions
when warranted. State which repository changed, what was checked, and any
unfinished work. Do not claim a starter change is complete based only on edits
to Bookworm.
