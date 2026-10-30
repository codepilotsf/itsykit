# CLI, runtime and routing

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

Core commands include `flea new <directory>`, `flea status`,
`flea dev [-port 3000]`, `flea setup`, `flea version`,
`flea deploy`, `flea admin [email]`, and `flea https`.
`flea server list [ssh-target]` and `flea server remove [ssh-target] [site-name]`
administer sites on a server directly, without a local site, Git repository or
saved deployment settings. An omitted target is prompted for (the server login,
for example root@fleacms.com); an omitted site name lists the managed sites and
asks for one by its exact name.
See README.md for pull, synchronization, recovery and the complete command list.
`flea server indexing [ssh-target] <hostname> <noindex|allow>` sets one exact
registered host policy through SSH with hostname confirmation and shared-service
restart. List/status show effective policies and static robots.txt warnings.
Policy enforcement belongs outside cached site handlers, at writes and flushes,
including downloads/errors and host redirects. Never buffer response bodies or
trust forwarding headers for selection. Keep readiness and TLS rejection unchanged.
See [host indexing](../host-indexing.md); this requires protocol 27.
`flea email` configures local Resend delivery interactively.
Run status/dev from a generated site. Setup uses native SSH.
`flea status` prints local validation, counts and Git state before automatically
checking the current named branch's saved deployment. Do not add a remote flag,
prompt for deployment setup, or select another branch's destination. Without a
mapping it finishes locally. Terminal output animates Checking remote; redirected
output stays plain. Stop animation before diagnostics or sudo prompts and always
join its goroutine on completion/cancellation. Remote work has a 60-second deadline.
Read-only status must not create a pull-baseline directory or refresh the Git index.
Use the existing SSH/protocol and bounded snapshot/history/inventory readers;
compare exact active release/branch/domain identity before reporting code state.
Website HEAD probes validate normal TLS and do not follow redirects. Unavailable
remote checks exit 1 while preserving the local summary; differences exit 0.
Bare `flea`, `flea help`, `flea -h`, and `flea --help` show the grouped command
overview on stdout and exit successfully. Each command supports `--help` and
`-h` without requiring a site or starting work. Keep the overview spacious;
put flags and deployment requirements in command-specific help.
`flea new <directory>` accepts relative or absolute paths, including `.` for the
current directory. Create a missing final directory; its parent must exist.
Existing targets must be real, completely empty directories, including hidden
entries and `.git`. Reject file targets and directory symlinks before writing.
Do not prompt to merge or overwrite files. Confine scaffold writes with os.Root
and retain exclusive file creation. Use the actual destination folder for naming
defaults. For `.` end with `flea dev`; otherwise show a shell-quoted `cd` when needed.
`flea new` initializes Git on main and creates an `Initial site` commit of the
generated files, respecting Git identity, hooks and signing. Keep inherited Git
location overrides out of every Git subprocess so another repository is untouched.
Commit failure preserves the site and prints diagnostics plus manual retry commands;
missing Git also preserves site creation. `--skip-git` skips all Git operations.
Do not configure a remote or create a hosted project automatically. After a successful
commit, show `git remote add origin <repository-url>` and `git push -u origin main`.
Output ends with local development instructions; do not add deployment instructions.
Dev uses pinned github.com/fsnotify/fsnotify directory notifications with a
100 ms debounce, not periodic file scans. Register included subdirectories,
reconcile new/renamed/deleted directories, and keep atomic saves working.
Fingerprint only at startup or after events to suppress self-generated writes.
On overflow rescan; on watcher failure stop dev with an error. No polling fallback.
Browser reload polling is separate and unchanged. Preserve invalid-edit recovery.
Dev admin pages also receive the revision reload script; production admin pages never do.
Starting dev again gracefully replaces the same user's dev server for the same
canonical site root, then applies the new flags. `internal/cli/dev_control.go`
uses an owner-checked mode-0700 directory under /tmp and a per-site Unix socket;
there is no HTTP shutdown endpoint or PID-based killing. Hold the existing
dev/pull/reset lock through shutdown, watcher completion and socket cleanup.
Never interrupt pull/reset, delete their locks, or proceed if shutdown times out.
Stale sockets may be removed only after acquiring the site lock; reject other
file types. Older binaries without this endpoint must be stopped manually once.
The dev server defaults to loopback and prints localhost URLs with the actual bound
port. Without an explicit `--port`, bind 3000 directly and retry consecutive
ports only for address-in-use errors, explaining each occupied port and the
selected replacement. Explicit ports are fixed; `--port 0` remains OS-selected.
Never probe then bind, retry permission/other failures, or bypass same-site
replacement and the dev/pull/reset lock. Derive URLs, email previews and LAN
Host/Origin checks from the actual listener port.
`--lan`/`-l` binds IPv4 interfaces and prints detected private/Tailscale addresses
and OS hostname/local aliases. The outer devaccess middleware validates these
exact Hosts at the listener port and actual private, RFC 6598 shared-address-space
(100.64.0.0/10) or loopback IPv4 peers, then marks the request context. Forwarded headers never
authorize access. Only marked dev requests may use LAN admin, contact, diagnostics
and private capability previews. Keep exact Origin, CSRF, host-bound sessions and
production HTTPS checks. Warn that LAN HTTP requires a trusted network; restart
after network changes. Fail clearly when no private or Tailscale IPv4 address is available.
The persistent host defaults to public HTTP
port 80, with opt-in automatic HTTPS on port 443 per domain.

## Email templates

mailers is a private, snapshot-loaded directory separate from page/partial parsing
and extraction. HTML/text bodies support escaped app.* and email.* values only;
the HTML layout has one {slot}. Keep recipient, subject, sender and Reply-To outside
templates. Both renderers use a capped writer, not a post-allocation size check.
Site runtime always wires RenderEmail; missing templates fail delivery without a
compatibility fallback. The optional nil renderer in low-level auth tests
is not a site-level fallback. flea email preview emits fake-data HTML without sending.
Contact's success destination is fixed as mount-relative `thank-you`, never visitor-selected.

New sites enable dev.email_preview. Only the dev CLI wires internal/mailpreview
as a Sender and mounts its random-token loopback route. Keep at most 20 messages
in memory, with 10-minute expiry, no-store responses, exact host/loopback checks
and sandboxed HTML. Explicit LAN mode also admits middleware-validated requests
to existing unguessable preview URLs, with exact request Origin checks. Previews
still open on the developer's computer; no inbox or login bypass is exposed.
Do not write previews to disk or let production choose this
sender. Browser launch failure prints a local capability link, never message data.
Preview mode/sender changes require restart; real local delivery remains explicit.

## Routing and isolation

Optional `seo.url` enables snapshot-generated `/sitemap.xml`, a default
`/robots.txt` with sitemap discovery, and missing public-page canonicals. Exact
`seo.exclude` paths stay crawlable with noindex. Preserve custom endpoint files,
canonical markup, immutable snapshots, pinned asset streaming and outer host
indexing enforcement. Read [SEO](../seo.md) for route eligibility, limits and the
separate production-origin and host-policy boundaries.

Mini apps use internal/appapi and internal/formguard. The included Contact app
is publicly mounted by apps.contact and launched only during actual dev/host
startup, never Project loading or status. The developer-facing contract is
docs/apps.md; see apps/contact/README.md for the included app.
internal/appmanager supervises children; internal/appruntime implements the
child transport; internal/appbundle packages the target-specific Contact
executable. Custom apps are site-local managed executables confined to
flea/apps/<name>/. Keep credentials in inherited pipes and an already-bound
listener handoff. Never pass provider environment to children. Preserve
parent-liveness cleanup, bounded restarts and per-site process isolation; this is
not an OS sandbox.
Contact is bundled; Optional Blog uses the built-in content snapshot/editor.
Custom managed packages are configurable; remote or registry-based registration
remains unsupported. App-owned sessions ride a separate per-site/per-app cookie
and the X-Flea-Session header; validate them strictly and never let an app session
grant admin access.
Never forward browser cookies or Authorization as service credentials. Keep
upstream redirects and POST retries disabled, response/deadline/concurrency
bounds enforced, and app values request-local. Anonymous form cookies are signed,
site/origin-bound and independent of admin sessions; reject duplicate/missing/
expired tokens before any app call. Preserve keys across Project reloads, never
inside Projects, logs or archives. Full restart invalidates forms. Only an
actually evaluated `{csrf}` may issue a cookie after a successful render.

Contact validation lives in apps/contact. internal/contact provides its narrow
site-bound email capability, with a separate bearer token and an exact single-use
reservation for the admitted submission. Recipients and optional email.from come
from apps.contact.email, never request data. Visitor email is Reply-To only.
Aliases share limits; distinct roots cannot share content files, including hardlinks.
Keep exact Origin/CSRF, production HTTPS, bounded bodies, site/IP limits and the
shared server budget. Private HTML/text mailers use mailer.Sender, including dev
preview. Do not store submissions or queue work. No old route/config aliases.

Contact GET/HEAD renders the form or thank-you page without an upstream read;
POST at the mount invokes the service. Redirects must remain under the mount.
Flash is bounded (256 entries/site, five-minute expiry), visitor/host/app/route
bound, and consumed only after a successful GET render. HEAD, assets, failed
renders and other pages cannot consume it. Changed app settings require restart;
never silently use an old recipient for a changed registration. Deployment
configures email for Contact-only sites too. Blog itself requires no mail; its
admin login still uses the normal mail and HTTPS requirements. Protocol 25 requires the managed runtime.
Test cross-site recipients, credentials, flashes and aliases without live email.

`index.html` maps to `/`. `about.html` and `about/index.html`
both map to `/about`; providing both is a route collision. Explicit `.html`
URLs and trailing slashes other than `/` return 404 unless explicitly listed in
top-level config `redirects`. Assets retain their paths.

`site/redirects.go` validates an exact old-path to local-destination object after
pages, media, apps and published Blog routes are loaded. Store it in the immutable
Project. GET/HEAD may use this map before ordinary safeURL rejects `.html` or a
trailing slash; this exception must never serve files. Decode config paths once,
reject unsafe/private/internal sources, duplicate decoded keys, route/app conflicts,
external or missing destinations, and any chain/self redirect/loop. Former Blog
slugs may redirect; other app subtrees retain their dispatch. Preserve incoming
queries unless the destination supplies one (including an explicit empty `?`).
Never redirect POST/mutations or serve a stale map after a failed reload. Query-
specific and wildcard matching are unsupported. Protocol 28 prevents deployment
to older servers that would silently ignore the rules. See README.md for syntax.

Files and directories beginning with `_` or `.` are private. Preserve URL
validation, confined reads, symlink rejection, and immutable asset snapshots.
Treat all Project maps and Content as immutable after publication. Build a new
Project for reloads. Keep project-file confinement separate from administrator-owned
absolute-path writes; their trust assumptions differ.
Ordinary assets above 64 KiB and all managed photos stream from load-time pinned
regular-file handles. Smaller assets and theme CSS retain byte snapshots. Reject
linked streamed files; requests only select validated map entries, never open
paths. Independent readers preserve concurrent Range/HEAD behavior. Size/mtime
changes return 503 before headers or interrupt an active transfer. Replace files
atomically, never mutate published streamed files in place. Unpublished failed
loads close handles immediately; published snapshots/readers retain them until
os.File cleanup. Monitor file-descriptor use alongside heap at hosting scale.
Fingerprinting and photo checksum validation stream with bounded buffers.
Do not replace validated route maps with request-derived filesystem reads.
Production keeps detailed errors in server logs and generic/custom HTTP error
pages. Local dev uses the ordinary 404 template and embedded diagnostics for 405/500 when
both Host and client are loopback, or the CLI's LAN middleware authorized the
request. Never expose these through production, forwarded headers or an unapproved
Host. SourceError captures bounded compiler-owned excerpts
at failure time; never read source files based on request paths or error strings.
Include five lines before and after the highlighted source line, with original
line numbers. Preserve source locations through region and page-setting validation.
JSON errors include safe file/line information, not configuration source values.
Escape all diagnostics, omit request queries/cookies/bodies, and quote error log
strings to prevent control-character injection. Keep recovery polling on error
pages and pin its exact script hash in CSP. Startup failures remain terminal-only.
Invalid edits produce an error state; fixing them restores serving.
Keep diagnostics limited to the problem, source excerpt and fix hint: no eyebrow,
action buttons or footer. Unclosed blocks highlight their opening directive and
name the missing closing directive. Missing braces must not consume the next directive.

The starter links css/navigation.css directly in both layout heads for native
cross-document view transitions, independently of the general CSS import chain.
Keep the small native-CSS opt-in/reduced-motion opt-out in both layout heads too,
so inbound transitions do not resolve before external styles establish the opt-in.
Never inject navigation CSS into rendered pages. No JS router
or click interception. A sole direct-child body main gets a 240 ms content fade;
other layouts fade the root. Use explicit opacity endpoints/fill, suppress the
old root for main-only transitions, and disable navigation animation for reduced
motion. Preserve custom layouts; existing sites opt in by linking the stylesheet.
Missing icons are the exception: omit them and warn without making the site invalid.
Dev request logs use `GET /test` for 200 and `404 GET /missing` for non-200
responses, including assets and HEAD. Suppress successful internal reload polls.
Log escaped paths without query strings, using the same concurrency-safe logger
as reload diagnostics. Capture the actual HTTP status, including partial content.

## Blog

Read [Blog](../blog.md) for the optional built-in blog, JSON and template contract.
Posts share the snip/media snapshot, atomic save and preservation/synchronization
paths. Drafts remain outside public routes and template data. The blog admin uses
existing login, exact Origin, CSRF, upload bounds and content-version checks.
