# Architecture and review guide

`internal/selfupdate` updates only the local CLI, separate from SSH server updates.
It trusts bounded HTTPS metadata from fleacms.com, checks checksum/platform/version,
and replaces the executable under the shared installer lock. Checksums are not
independent signatures. Public-host and clean-machine verification remain pending.

Project files have explicit boundaries: `site/` supplies templates and public
assets; `mailers/` supplies private `*.email.html`/`*.email.txt` templates and
`main.layout.html`; `flea/` holds content, media, data, icons and agent skills.
Runtime files are under `flea/runtime/`, outside serving, watching and archives.
Template names and public URLs stay relative to site/, not the project root.
Extraction journals use physical site/ paths. All *.layout.html files are private.
Root `.flea.config.json` is the only admitted dotfile. Deployment destinations
are nested in deploy.branches; archive packing removes deploy while retaining
runtime settings. Configuration reads stay bounded and confined. Protocol 23
requires this structure, with no alternate old-layout loader.

Mail templates are private runtime files in `mailers/`. The shared inclusion
policy retains them for watching, deployment and backups; the email renderer
loads them separately from web pages and snip extraction. Protocol 23 requires
this location, with no old-directory fallback.

FleaCMS has a developer CLI and persistent website server sharing the site loader
and renderer. The bundled Contact mini app is a separate child executable, not
a general application framework. There is no database, background job queue or
public deployment administration API.

`apps.contact` enables a validated mount, fixed recipient and optional sender
override. `apps/contact` owns validation and JSON outcomes; `internal/appruntime`
owns the child transport. `internal/appmanager` supervises it only during actual
dev/host runs. Loading, status, validation and extraction never launch apps.
Aliases share a site runtime. `internal/appbundle` embeds only its target's
executable; build scripts compile the child before the parent. One verified
private temporary executable is shared by all sites in the parent process.
Listener fd 3, credential pipe fd 4 and parent-liveness fd 5 avoid port races,
argv/environment secrets and orphaned children. Five attempts use bounded
1/2/4/8-second backoff. Children receive a minimal environment, not provider
credentials. Child processes are not OS sandboxes.

`internal/contact.Delivery` is a private, site-bound email capability. Its separate
bearer credential is not a general relay: an exact currently admitted submission
can send only once. Recipient, sender, subject and templates are chosen by Flea.
Visitor email is Reply-To only. Per-site/IP limits and a shared contact budget
bound work. No submission storage exists. Provider quota is shared with login
mail. App settings are startup-bound; a changed registration fails closed until
restart. Other site pages remain usable.

`internal/appapi` owns Flea App Protocol v1 types, mount validation, flat form
decoding and bounded authenticated HTTP calls. It never imports site or editor.
`internal/site/app_bridge.go` renders request-local envelopes, confined relative
redirects and bounded flash. Contact GET/HEAD routes render without service reads;
POST forwards JSON. The flash store has at most 256 entries/site, five-minute
expiry, and visitor/host/app/destination binding. Only successful GET rendering
consumes a claim. No credentials or messages enter logs, site files or archives.
Custom-service registration, dynamic read-route declarations and storage remain
planned; see [Contact's contract](../../apps/contact/README.md).
`internal/formguard` owns anonymous visitor CSRF independently of editor sessions.
Each site Server retains a lazy signing key across snapshot reloads; no Project,
disk file or per-visitor server map holds it. `{csrf}` calls a request-local
renderer only when evaluated. Cookies commit after successful rendering; pages
with tokens are private/no-store. Validation creates no tokens. Restarting the
server invalidates open forms. Contact requires this protection; admin handling is unchanged.

Read this alongside README.md for the supported behavior and AGENTS.md for
development rules. docs/agents/spec.md also describes features that are not implemented.

## Where to start

Local site creation composes the canonical starter with embedded minimal content
overrides by default; `--starter` opts into the examples. `cli/reset.go` validates
a fresh minimal tree before confirmation, takes the dev/pull lock, and moves old
entries (except .git) to a mode-0700 sibling backup before publishing the new tree.
Ordinary failures attempt rollback; interruptions leave the printed backup and
remaining original entries for manual recovery. This is local-only, not server
reset, deployment, or content synchronization. Backups may contain secrets.

| Responsibility | Code |
| --- | --- |
| CLI dispatch, flags, local server lifecycle, new-site creation | `internal/cli/cli.go` |
| Validated site snapshot and rendering | `internal/site/project.go` |
| Template grammar and alias scope | `internal/site/template.go` |
| Page filename to canonical route | `internal/site/routes.go` |
| Legacy snips and structured JSON loading | `internal/site/data.go` |
| HTML region scanning and lexical snip references | `internal/site/regions.go` |
| Dev-only content extraction | `internal/site/extraction.go` |
| Content-first write journal and replay | `internal/site/snip_transaction.go` |
| Confined project reads and atomic writes | `internal/site/files.go` |
| Lucide resolution, SVG validation, icon compilation | `internal/site/icons.go` |
| Request handling and reload publication | `internal/site/server.go` |
| Native directory notifications and debounced reloads | `internal/site/watch.go` |
| Snip model, IDs, HTML sanitization, content merge | `internal/snips/` |
| Shared runtime-file inclusion policy | `internal/sitefiles/` |
| Hostname routing, configuration, listeners, ACME | `internal/host/` |
| Local Git snapshot, prompts, SSH, installation and upload | `internal/setup/` |
| Embedded Linux servers and bounded verified extraction | `internal/serverbundle/` |
| Bounded release archives and compatibility marker | `internal/release/` |
| Privileged activation, readiness, recovery | `internal/admin/deploy.go` |
| Read-only site inventory and unregister | `internal/admin/sites.go` |
| SSH server-wide site listing and removal | `internal/setup/server.go` |
| Preserved-content staging during activation | `internal/admin/content.go` |
| Owner authentication and inline editing | `internal/editor/` |
| Content-save authorization boundaries and publication | `internal/editstore/` |
| Resend delivery behind a small Sender interface | `internal/mailer/` |
| Cross-process content/deployment lock | `internal/state/` |
| Embedded starter source | `templates/` |

Dependencies point toward site/content primitives, not toward command handlers.
`cli` calls `setup` and `site`; `setup` uses `release`, `site`, and shared admin
option validation. `admin` uses `host`, `release`, `site`, and `snips`. `host`
uses `site` and `snips`. The site package never imports CLI, setup, or admin.
`snips` does not know about HTTP, SSH, templates, or deployment orchestration.

Only CLI/setup imports serverbundle; the server never embeds copies of itself.
`cli/dev_control.go` handles dev replacement through a per-user private Unix socket,
separate from HTTP/authentication. A new invocation cancels the old dev context
and waits for its watcher, HTTP shutdown and development-lock release before
loading the project. Pull/reset do not expose this endpoint and remain exclusive.
The lock remains authoritative; stale sockets never authorize concurrent writes.
Build both Linux servers before embedding their compressed bytes and hashes into
the CLI. Setup validates extraction before reset confirmation/upload. There is
no runtime download/compiler fallback. scripts/build.sh refreshes generated assets.

## Loading and serving

Multisnips use page-local arrays in snips.Store and persistent field definitions
derived from their developer-owned item template. Extraction consumes a single
seed wrapper before ordinary snip scanning, with one shared recovery journal.
Rendering validates field placement even for empty arrays. Owner-only prototypes
and boundary comments enable preview additions without client-generated templates.
Browser arrays are checked against visible groups, schema and version before
publication. Preserve remote arrays during normal activation, including intentional
empty lists. Explicit sync conflicts at whole-array granularity. Archived ordinary
snips and retained removed fields are not a general content revision system.
Collections arranges real page items in place. Shadow-root controls track item
rectangles; only an empty temporary layout tile enters each group. Controls and
tile content never enter saved fields. Dragging adds an inert, styled clone outside
the editable regions, hides the original in its landing slot, and removes the
clone on drop/cancel. Dashed region outlines are shadow-root overlays, not saved
attributes. Editing preference is tab-local sessionStorage, separate from content.
Pointer/keyboard movement changes the
ordered draft array and its DOM nodes. Cancel restores original nodes, attributes
and order. Empty groups keep an Add tile, with no group picker. Photo fields open
a shared-styled dialog; empty photos render a fixed neutral portrait. Three pinned
editor Lucide icons use the shared {icon ...} compiler and SVG validator.
Collection photo uploads use the same bounded multipart page transaction as typed
media. New IDs must be unused and referenced by visible photo fields in the draft.
Processed immutable bytes precede atomic library metadata and item publication.
Unused drafts are omitted from the request; Cancel revokes private object URLs.

Site-owned email templates compile into immutable HTML/text renderers at load.
They are not pages/partials and are never read based on a request path. Renderers
receive only public app values and message-specific data and stop at output bounds.
Auth/contact retain ownership of delivery, recipients and headers. Render failures
never leak codes or message data in logs or expose account membership.

`site.Load` validates the runtime tree, reads JSON, scans templates, resolves
routes, compiles partials/pages, snapshots assets, and checks rendering. All
partials are checked, including unused ones. Routes and snip page ownership use
the same filename conversion. Legacy named snips and `data.*` share a namespace
and one JSON-directory loader; collisions are errors.

Asset snapshots are hybrid: ordinary files up to 64 KiB and theme CSS keep byte
snapshots; larger assets and all managed photos retain confined regular-file
handles instead of payload buffers. Requests use independent SectionReaders,
never path opens. Streamed files reject hardlinks as well as symlinks. Size/mtime
checks before headers fail with 503; checks around each read stop an active
transfer if the file changes. This assumes published files are not modified in
place, like videos; it is not protection against a privileged writer restoring
file metadata deliberately. Atomic replacement preserves the old open inode.
Failed loads close their handles; published handles are reclaimed by os.File
cleanup after snapshots and active readers become unreachable. This trades heap
retention for file descriptors, whose limits need capacity testing. No disk cache
copies or new writable site directories are created. Fingerprinting streams
bounded file reads rather than allocating each file in full.

A `Project` is immutable after publication. Public config values come from `app`
(app.site_name for the display name), never admin/mail settings. An omitted name
gets a read-only folder default. Interactive dev saves it with confined,
RawMessage-preserving writes; deployment requires a committed name. Protocol 23
requires the app namespace without a site alias. Its exported maps exist for internal
inspection, not concurrent editing. HTTP handlers select from validated maps;
they never turn request paths into filesystem reads. Template output is buffered
before sending headers, so a render failure can still produce a proper 500.

Native cross-document transition CSS lives in the site's navigation.css, linked
directly by both layout heads before global.css. Each head also declares the
small native opt-in and reduced-motion opt-out before external styles resolve.
Rendering never injects it. No JavaScript router is involved.

Dev 405/500 diagnostics are embedded and independent of site templates/assets;
404 responses use the site's ordinary error template with dev reload support.
They require a loopback peer and Host, or CLI-authorized LAN context, as well as
non-production mode. Compiler
SourceError values hold bounded 11-line source excerpts and original locations
at failure time, including region and page-setting validation failures;
JSON syntax errors expose locations only. No request triggers source-file reads.
Error pages escape details and allow only their hash-pinned reload script. Request
failures and failed custom error rendering log quoted details in both modes;
production HTTP stays generic/custom. Startup validation remains terminal-only.

The template compiler translates FleaCMS directives into Go `html/template`.
`{snip ID}` becomes a sanitized rich-content call. The HTML region scanner
recognizes the same unquoted markers for extraction, editor access and deployment
pruning. Quoted markers are rejected. HTML comments and script/style contents
do not contribute references; comments never execute template instructions.
Go retains responsibility for contextual escaping. JSON is passed as values,
never compiled as template source. Rich snips are the narrow exception to normal
text escaping: the HTML allowlist reconstructs them first, and only the result
becomes `template.HTML`. SVG icons have a separate geometry-only validator.
Keep these two sanitizers separate: their permitted markup and uses differ.

There are three intentionally distinct loading entry points:

- `Load` / `LoadWithSnips`: read-only; no downloads or source changes.
- `LoadDev`: may fetch missing icons; also used on a temporary committed snapshot
  during deployment. It does not extract snips.
- `LoadEditing`: extracts local editable regions, then calls `LoadDev`. Only the
  local dev command and watcher should use it.

Reload builds a replacement snapshot outside the request lock. A writer mutex
serializes reloads; the shorter state lock publishes the result. Invalid edits
publish an error state, not an apparently current old page. The prior snapshot
remains available for a custom error response. Production has no watcher or
reload endpoint. Host aliases for one release share one serving snapshot and
must select the same content file.

## Filesystem boundaries

Host configuration may map a site's hostname to an administrator-owned download
root through `downloads`. Aliases must share that root. `host/downloads.go` loads
a bounded catalogue of stable-version directories and SHA256SUMS, rejects links,
unexpected files and writable entries, hashes binaries with streaming reads, and
pins regular file handles. Requests select exact catalogue URLs and use independent
SectionReaders, never request-derived file opens. Changed size/mtime fails closed.
These files are outside Project snapshots, site archives, backup and rollback.
Versioned bytes are immutable; publishing a new version requires a host restart
to validate and load the new catalogue. The website still owns latest.txt.
No deployment protocol change is needed: this is explicit host policy, preserved
by normal activation, not new site archive syntax. Older executables reject this
configuration field, so remove it via a reviewed config rollback before reverting
to a pre-0.7 executable. Never modify existing download bytes in place.

Standalone images/videos are typed snip entries, with direct Image/Video metadata
and optional direct Poster metadata. Extraction replaces source/alt/poster values
with typed bindings while preserving developer markup. Owner-only media markers
enable click-to-edit; template validation rejects nesting and competing sources.
Multisnip photo fields retain their photo catalogue. ImageAssets/VideoAssets
aggregate all referenced typed active/archive entries and catalogue files for
loading, synchronization and recovery. Three-way sync compares a whole typed entry,
excluding its developer-owned label, not merely the text content field.

Hash-named MP4 files share flea/media/, but never enter AssetData. Loading validates
bounded MP4 structure/codecs/checksum and pins a regular file handle. Requests
select only catalogue entries and use independent SectionReaders for Range/HEAD;
no request-derived path opens and no video playback buffers in the asset map.
Normal uploads write immutable replacements, then validate/save metadata and
publish. Failed writes may leave an unreferenced file, never a premature live link.
Old snapshots retain their handles until garbage collection; do not mutate files
in place. Symlinks/hardlinks are refused. A changed size/mtime stops serving.
Local video bytes are excluded from Git, not backups. Deployment supplements the
committed snapshot with exact hash-verified local bytes from committed metadata.
Pull/backup include catalogue files; activation preserves live metadata and files.
The pull-only LoadSyncSource skips video bytes and is explicitly nonservable,
so fresh clones can obtain ignored files. Pull validates a complete normal snapshot
after transfer and before publishing content or advancing its baseline.
Video storage has an independent 64 MiB retained-file quota. Release/backup bounds
are 256/270 MiB to accommodate both photos and videos alongside site code.

Photo metadata belongs to snips.Store, while hash-named bytes live in local media/
or the production content directory's media/ child. LoadWithMedia overlays only
catalogued, checksum/dimension-checked file handles into the serving asset map.
Photo checks stream encoded bytes without retaining full image buffers. Browser uploads
decode/re-encode bounded JPEG/PNG under the content lock before saving metadata.
An interrupted save can leave an unreferenced immutable photo; old bytes are never
silently removed. Deployment seeds missing files and retains live metadata. Pull
verifies a backup's state digest before installing immutable referenced files.

Per-site release receipts form a chain from the current activation. Rollback
requires a reachable receipt and reviewed active release, then reuses activation
with live state and unchanged host policy. Backup/restore use bounded checksummed
site archives, excluding credentials/runtime state. Restore requires a reviewed
live digest and creates a new release. Upgrades use a separate executable journal;
pending upgrades block content/deploy mutations. The new executable validates all
sites before replacement; failed readiness restores the saved old executable.

`sitefiles.Included` is a runtime inclusion policy, not a public-URL policy.
Configuration and content belong in an archive but remain private over HTTP.
Site configuration uses only `.flea.config.json`; protocol 23 rejects older
server contracts before upload, including the inline-save/formatting contract.
There is no extensionless-name fallback.
Do not combine this predicate with the server's URL checks.

Project files are untrusted. `site/files.go` pins directory handles, rejects
symlinks/nonregular files, and compares opened file identities. Writes use an
exclusive temporary file in the pinned parent, sync, rename, then sync the
directory. The icon cache uses these same write primitives.

Administrator-owned configuration has a different boundary. `admin.atomicWrite`
works with trusted absolute paths; `realDirs` rejects linked managed directories.
`snips.ReadFile` pins an administrator-selected content directory and checks its
file. These helpers are deliberately not forced into one general filesystem
abstraction. Changing their trust assumptions requires tests, not just deduplication.

## Two separate transactions

Explicit synchronization uses `internal/syncstate` for pure three-way comparisons,
`setup/sync.go` for branch-scoped baselines and SSH orchestration, and
`admin/sync.go` for a bounded privileged state export. Export and activation use
the same deployment lock as browser saves. A reviewed plan carries an expected
snapshot digest; activation refuses changes since review. Baselines are local Git
metadata, never site runtime files. Pull locks out dev, validates candidates and
checks local Git state again after network/prompts. It saves its baseline last;
disk errors may leave reported partial local file writes, never a false baseline.

Theme CSS is managed beside each content snapshot, selected by host.Themes.
Aliases must share the same content/theme pair. `LoadWithState` overlays immutable
CSS asset bytes without making code releases writable. Theme saves use the existing
editstore lock/version/publication sequence and source-preserving `internal/theme`
value validation. Code, content and theme switch together through the activation
journal. Browser theme editing never writes host configuration or release files.
Scale edits carry numeric endpoint settings, not client-generated CSS. Go validates
and generates the declarations and settings comments under the same save lock.
Browser calculations provide private previews only; parity tests cover the two
implementations. No Node process or extra configuration file is involved.
Font-family saves additionally regenerate a marked Google Fonts import block.
Names are validated and URL-encoded into a fixed HTTPS endpoint; no arbitrary
URL, server fetch or credentials are accepted. Existing developer imports remain
unchanged. Private preview uses a removable stylesheet link and Cancel restores
the original inline variables.

### Local extraction

The HTML tokenizer records byte ranges, so extraction preserves source formatting
outside the region and its consumed snip authoring attribute. Quote-aware attribute
spans preserve all other tag bytes; generated directives need no persistent attribute.
Existing directives with redundant markers use the same transaction without new IDs.
The extractor reserves IDs against active and archived content,
checks all regions, and builds edits before writing anything.

While holding the in-process content mutex, it saves `flea/runtime/.snips-pending.json`, writes content,
then replaces templates. The journal contains before/after values. Replay accepts
either state, but refuses a conflicting independent edit and retains the journal.
Local extraction and editor saves share that mutex; always acquire it before the
server's reload mutex to avoid a publication deadlock. No local content lock file is
created. Multiple dev processes on one project are unsupported. Every relevant
save reindexes references and archives removed IDs; reintroduced IDs recover
their content. Server deployment still uses its cross-process file lock. This is recoverable
multi-file editing, not an atomic filesystem transaction across all files.

### Remote activation

1. The CLI captures a clean Git commit, validates and packages its runtime files,
   verifies the server protocol and DNS, then uploads over SSH.
2. The server holds its deployment lock, extracts a fresh release and validates it.
   `prepareContent` merges committed seeds with remote owner content by stable ID.
   Both the old release and old content file stay untouched.
3. A candidate host configuration points to the new release/content pair. The
   server validates it, persists the previous configuration in a recovery journal,
   atomically replaces the active config, restarts, and checks readiness.
4. Failure restores the previous configuration. Success archives unreferenced
   active snips, saves the previous config, and removes the journal. Cleanup
   failure keeps the journal and reports that the release is already active.

Snip references are lexical across all templates, not collected from one render.
Otherwise empty loops or conditional sections could incorrectly lose content.
Browser saves share the deployment lock, reject pending recovery journals, verify
the active release/content pair, and compare a content digest from the submitted
form. They sanitize and validate a replacement snapshot before atomically saving
and publishing it. Local saves use the extraction lock instead. HTTP requests never
choose a filesystem destination. Sessions persist independently of releases.

Server-wide site administration is read-only listing plus unregister/offline
removal, invoked over SSH without a local site. It deletes no files and reuses the
same deployment lock, expected-state digest (including branch/commit bookkeeping),
recovery journal, restart/probe and rollback paths as activation. It removes only
the target's routing and target-owned ancillary entries, preserving unrelated
sites' entries; when no HTTPS hostnames remain, it clears the TLS allowlist and
disables automatic HTTPS while cached certificates stay on disk. Config activation
restarts the shared service, briefly affecting other sites. Retained content is
not restored automatically. Ambiguous redirects and unmanaged roots fail closed.

Authentication/session persistence belongs to `editor`, content persistence to `editstore`, and delivery to
`mailer`. The only email-provider dependency is the Sender interface; only Resend
is implemented. Credentials remain in the root-only systemd EnvironmentFile, not
site snapshots. The service can write content/private runtime directories and the shared lock,
but not code releases or host configuration. This is a change from the previous
read-only service model and requires the updated installation permissions.

Dev always mounts the editor so a fresh site's /admin can explain owner setup.
Before serving, interactive CLI startup asks for the site name and optionally
reuses AddAdmin; it never prompts for Resend setup. It reloads after writes so session setup and
the first request see the new owner. Nonterminal startup never reads prompt input.
The editor enforces loopback host checks before showing local instructions,
unless devaccess has authorized an explicit LAN request. `flea dev --lan`/`-l`
binds IPv4 interfaces and snapshots detected private IPv4 addresses at startup.
The outer middleware admits only exact approved Hosts with the listener port and
actual private/loopback peers, then adds an unexported context marker. Editor,
authenticated public rendering, contact, diagnostics and previews share this
decision; no HTTP header can enable it. Production never installs this middleware
and still requires HTTPS. LAN mode changes reachability, not authentication,
session host binding, CSRF, exact-Origin checks or content validation.

The dev CLI alone reads dev.email_preview (enabled in the starter) to select
internal/mailpreview instead of Resend for both login and contact delivery.
It holds at most 20 bounded messages in memory behind random 256-bit URLs with
10-minute expiry, exact listener Host/loopback-peer checks and no-store responses.
Explicit LAN mode allows existing capability URLs on middleware-approved network
addresses too, retaining exact Origin and cross-site checks. Browser launch stays
on the developer's computer; no inbox or automatic disclosure to login requesters.
The browser view escapes headers and sandboxes the HTML body; scripts, forms and
external resource loads are blocked. Browser launch is bounded; failure logs only
the local capability URL. No preview data is written to disk. Configuration and
preview sender changes require dev restart. The production host never wires this
sender or route, even when the committed config enables previews.
The dev admin wrapper selects the current validated access list and reads local
mail settings per request. A serialized handler replacement retains valid sessions
and discards pending codes when settings change, with a matching edit-store allowlist. Invalid snapshots
or unreadable mail settings fail closed. Production configuration remains fixed
until restart/deploy; local contact mail remains startup-bound.
Production mounts it only for sites with approved owners. Setup and missing-mail
pages grant no editing access; the normal email-code flow is still required.

Authenticated public-page editing uses `editor.ServePublic` and an injected
`site.Server.ServeOwner` callback. Authentication remains outside site. A private
request-context value enables rich-snip comment markers during rendering; neither
query parameters nor shared Project state can enable them. Only successful public
page responses expose request-only app.current_admin to explicit layout conditionals.
Config cannot define this reserved field. The template places a csrf-token meta
and admin/_toolbar.html declarative shadow-root shell. The template supplies the
visible rows before asynchronous metadata/dialog loading; initialization reuses
the already-rendered formatting SVGs from {icon ...} tags and
the shadow root and enables its inert controls. Rendering never injects UI or libraries. Static
editor assets have revalidating ETags and no private data. Authenticated metadata
fetches require the meta token and return private/no-store JSON. Bootstrap checks
the content version against the rendered mount before allowing editing.
Visitors, assets and errors retain ordinary rendering. Owner responses are private,
no-store; their CSP permits the exact editor asset URL, not other site or inline scripts.
Dev supplies the rendered revision to the external owner controller, which polls without cookies and reloads only
when drafts, dialogs, saves and Theme previews are not at risk. Edit mode and
scroll survive clean reloads in tab-local storage. Production omits this revision.
The controller identifies complete wrappers by the server's directive-boundary
comments, not source attributes. It changes only those wrappers in memory and restores them
on Cancel. The same rendering snapshot supplies metadata and a version digest of
content, templates, structured data, site settings and release root. A later disk
read cannot attach a newer version to older rendered content.
POST /admin/page applies changed IDs and metadata in one content transaction,
with ordinary session/Origin/CSRF checks. Validate rendered complete-wrapper IDs,
reject duplicates and hidden/cross-page changes, and constrain leaf content to
inline marks/links. Inline formatting uses Squire in a temporary offscreen prose
buffer, then sanitizes back to inline HTML; template wrappers never receive
nested blocks. Native selection is restored by text offsets.
Check shared wrapper constraints across routes. Sanitize rich content, validate
the full candidate, write once, then publish. Stale/failed saves retain browser drafts.
Page saves return a canonical receipt captured under the same lock as publication.
The browser applies it in place and advances its version and Cancel baselines;
no post-save Read may pair a newer version with older displayed content.
Typed media use the same transaction. Multipart saves authenticate and check header
CSRF before parsing or acquiring the upload slot; form CSRF is checked afterward.
Uploads are bounded at 40 MiB total, 25 MiB per video and 8 MiB per image/poster.
Write immutable bytes before candidate validation and metadata publication. Failure
can leave an unreferenced file, never a published broken link. Direct image/video
dialogs stage private object-URL previews; no standalone library mode or upload API.

The login handler passes plain data to a renderer supplied by the site's Server.
Signed-in /admin redirects to /. Public-page metadata is editable even without snips.
The renderer selects a precompiled developer-owned admin/*.html
template and its selected layout from the current immutable snapshot. Values are
contextually escaped, never compiled as source. The admin directory is excluded
from public routes and assets; authentication remains in the editor handler.
Missing templates fail without an embedded fallback. This shares site chrome and
reflects template reloads without an editor dependency on site.
Login styling permits HTTPS CSS/fonts/images. Verification pages permit only the
hash-pinned code-input enhancement; other login pages block scripts. Login mail
has escaped HTML with a prominent code and a plain-text alternative. Inline pages allow
the external editor script. Same-origin referrer policy preserves browser form origins;
origin and CSRF validation remain exact.

### Session runtime state

Both environments use a private `flea/runtime/sessions.json`: below the project locally,
below `/var/lib/flea/sites/<name>/` in production. Production's `flea/` is service-owned
0700, beside content/releases, not inside either. Code and parent directories remain
administrator-owned. Deployment protocol 23 requires these runtime permissions.
Host derives this stable location from the administrator-configured managed content
path; aliases share storage, distinct sites must not. No request chooses a path.
Candidate host validation is read-only; actual startup prunes expired/removed owners.

The editor stores SHA-256 token digests, owner, host, CSRF, last activity and expiry.
It pins the parent, rejects linked/nonprivate files, bounds reads, and takes a
dedicated cross-process .sessions-lock around each read/modify/atomic durable write.
The cache is never authoritative: every transaction reloads, preventing stale
handlers from restoring a logout. Store errors fail closed without replacing
malformed data. Sessions never enter Git, release archives, or reload fingerprints.
The local CLI installs ignore rules and refuses already tracked session state.

The idle window is configurable in days, default 30. Current policy can shorten
existing sessions. Valid admin navigation and CSRF-checked forms renew both disk
expiry and the persistent cookie. Trusted visible user events send a throttled
activity POST, not an idle heartbeat. The controller is an external same-origin
asset; no dynamic JavaScript interpolation. OTPs/rate limits remain
bounded in-memory state. Browser/device tokens are independent and host-bound.

The loader resolves leading flea settings comments before compiling partials.
Selected layouts are private HTML files removed from the partial namespace;
they may render partials but cannot select another layout. Only the layout
compiler accepts slots. Unknown/duplicate settings and invalid targets fail load.
The admin starter selects admin.layout.html; public main.layout.html is independent.

`flea email` uses the setup-only mailer.Checker interface for a non-sending
connection check. Resend implements both Checker and Sender; auth/contact code
depends only on Sender. Future provider adapters should use these boundaries.
The localenv package handles bounded, literal .env reads and owner-only writes,
rejects links, and preserves unrelated settings. Setup refuses Git-tracked .env
files and ensures an ignore rule before saving secrets. Dev and deployment mail
setup merge only mail settings with process-environment precedence, without
changing process environment. Production still reads systemd's mail.env.

## Review findings and decisions

This refactor addressed mixed responsibilities in snip handling, icon-specific
naming of shared file writes, duplicated route/data/SSH rules, repeated template
alias restoration, and obsolete single-directory loops. Deployment content
staging is now separate from the activation transaction. A dead prompt reader was
removed; its cancellation test now exercises the prompt actually used by the CLI.

One correctness issue was also addressed: aliases for the same release could
select different owner-content snapshots. That configuration is now rejected
instead of letting deployment choose content by map iteration order. Valid aliases
reuse a snapshot, avoiding duplicate template compilation and asset storage.

The existing package boundaries do not justify introducing repositories, services,
plugin interfaces, a template AST, or a general-purpose transaction framework.
Small functions and explicit state transitions are sufficient at this scale.

Remaining work and limits, not claims of completed sign-off:

- Exercise real Ubuntu/systemd installation, interrupted activation, and live
  certificate issuance/renewal. Local tests and simulated SSH cannot establish these.
- Exercise executable-only `flea update-server` and its recovery path on a real server.
  Setup reset is not an upgrade. Publishing artifacts remains operational work.
- Local dev uses pinned fsnotify directory notifications, registering included
  subdirectories and reconciling new/moved/deleted directories after a 100 ms
  debounce. Content fingerprints run only at startup or after events, suppressing
  duplicate reloads from extraction's own writes. Overflow triggers a rescan;
  watcher failures stop dev visibly, without a polling fallback. Reload still
  validates a complete snapshot. Browser reload notification remains a separate
  700 ms HTTP poll; the admin does not auto-refresh unsaved forms.
- Keep the published `Project` immutable. Consider hiding exported maps if more
  in-process consumers appear; current callers do not need a broad accessor API.
- Releases, content snapshots, archived snips and old photo files are retained.
  Rollback is implemented; automatic retention/pruning still needs a policy.
- A security review should add fuzzing and resource-budget tests for parsers and
  extraction recovery. Existing attack/regression cases are not exhaustive proof.

## Verification

From this repository, using the installed Go executable:

```sh
export GOCACHE=/tmp/flea-go-cache
/usr/local/go/bin/go test ./...
/usr/local/go/bin/go test -race ./...
/usr/local/go/bin/go vet ./...
bash scripts/build.sh
/usr/local/go/bin/go build -o bin/flea-server ./cmd/flea-server
```

`TestExecutableWorkflow` creates a site, starts dev, checks pages/assets/private
paths, exercises edits/extraction/error recovery, and shuts down. Tests also cover
local HTTP/TLS/ACME, alias content consistency, confined file operations, Git and
SSH workflows, and failed deployment recovery. They need loopback bind permission.
