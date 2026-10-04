# CLI, runtime and routing

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

Core commands include `flea new <site-name>`, `flea status`,
`flea dev [-port 3000]`, `flea setup`, `flea version`,
`flea deploy`, `flea admin [email]`, and `flea https`.
See README.md for pull, synchronization, recovery and the complete command list.
`flea email` configures local Resend delivery interactively.
Run status/dev from a generated site. Setup runs from the Mac over native SSH.
Bare `flea`, `flea help`, `flea -h`, and `flea --help` show the grouped command
overview on stdout and exit successfully. Each command supports `--help` and
`-h` without requiring a site or starting work. Keep the overview spacious;
put flags and deployment requirements in command-specific help.
Successful `flea new` output ends with `cd <site>` and `flea dev`; do not append
publishing instructions or deployment commit guidance.
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
port. `--lan`/`-l` binds IPv4 interfaces and prints detected RFC1918 addresses.
The outer devaccess middleware validates exact listener Host/port and actual
private/loopback peers, then marks the request context. Forwarded headers never
authorize access. Only marked dev requests may use LAN admin, contact, diagnostics
and private capability previews. Keep exact Origin, CSRF, host-bound sessions and
production HTTPS checks. Warn that LAN HTTP requires a trusted network; restart
after network changes. Fail clearly when no private IPv4 address is available.
The persistent host defaults to public HTTP
port 80, with opt-in automatic HTTPS on port 443 per domain.

## Email templates

mailers is a private, snapshot-loaded directory separate from page/partial parsing
and extraction. HTML/text bodies support escaped app.* and email.* values only;
the HTML layout has one {slot}. Keep recipient, subject, sender and Reply-To outside
templates. Both renderers use a capped writer, not a post-allocation size check.
Site runtime always wires RenderEmail; missing templates fail delivery without a
compatibility fallback. The optional nil renderer in low-level auth/contact tests
is not a site-level fallback. flea email preview emits fake-data HTML without sending.
Contact success_path is developer-configured, local-only, never visitor-selected.

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

Email-only contact forms live in internal/contact. Configure contact.to in site
JSON; POST /_flea/contact accepts only name/email/message/website. Bind handlers
and recipients to the loaded site, never request-selected IDs or addresses.
Aliases share limits; distinct roots must not share a content file, including
hardlinks. Keep exact Origin checks, production HTTPS, bounded bodies, per-site/IP
limits and a shared server contact budget. Visitor email is Reply-To only; send
plain text through mailer.Sender. Do not store submissions or queue work.
Local contact settings require restarting dev. Deployment configures mail for
contact-only sites too. See README for limits and setup. Protocol 22 requires
the contact runtime. Test cross-site recipients and aliases without live email.

`index.html` maps to `/`. `about.html` and `about/index.html`
both map to `/about`; providing both is a route collision. Explicit `.html`
URLs and trailing slashes other than `/` return 404. Assets retain their paths.

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
