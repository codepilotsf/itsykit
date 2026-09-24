# Architecture and review guide

Email-only forms use internal/contact, configured by contact.to. Host routing
binds handlers to fixed recipients and shares them across aliases. Validate Origin,
TLS and bounded fields before synchronous, timed mailer.Send. Visitor input sets
Reply-To and text, never recipient, sender, site or paths. Per-site/IP limits and
a shared contact budget bound work. No submission storage exists. Distinct site
roots cannot share content inodes. These are trusted in-process services, not an
OS sandbox; provider quota remains shared with login mail. Configuration is
captured at startup. See README for the implemented limits.

FleaCMS has two entry points: a developer CLI and a persistent website server.
Both use the same site loader and renderer. There is no application runtime,
database, background job framework, or administration API exposed over HTTP.

Read this alongside README.md for the supported behavior and AGENTS.md for
development rules. spec.md also describes features that are not implemented.

## Where to start

| Responsibility | Code |
| --- | --- |
| CLI dispatch, flags, local server lifecycle, new-site creation | `internal/cli/cli.go` |
| Validated site snapshot and rendering | `internal/site/project.go` |
| Template grammar and alias scope | `internal/site/template.go` |
| Page filename to canonical route | `internal/site/routes.go` |
| Legacy fleas and structured JSON loading | `internal/site/data.go` |
| HTML region scanning and lexical flea references | `internal/site/regions.go` |
| Dev-only content extraction | `internal/site/extraction.go` |
| Content-first write journal and replay | `internal/site/flea_transaction.go` |
| Confined project reads and atomic writes | `internal/site/files.go` |
| Lucide resolution, SVG validation, icon compilation | `internal/site/icons.go` |
| Request handling and reload publication | `internal/site/server.go` |
| Native directory notifications and debounced reloads | `internal/site/watch.go` |
| Flea model, IDs, HTML sanitization, content merge | `internal/fleas/` |
| Shared runtime-file inclusion policy | `internal/sitefiles/` |
| Hostname routing, configuration, listeners, ACME | `internal/host/` |
| Local Git snapshot, prompts, SSH, installation and upload | `internal/setup/` |
| Embedded Linux servers and bounded verified extraction | `internal/serverbundle/` |
| Bounded release archives and compatibility marker | `internal/release/` |
| Privileged activation, readiness, recovery | `internal/admin/deploy.go` |
| Preserved-content staging during activation | `internal/admin/content.go` |
| Owner authentication and server-rendered dashboard | `internal/editor/` |
| Content-save authorization boundaries and publication | `internal/editstore/` |
| Resend delivery behind a small Sender interface | `internal/mailer/` |
| Cross-process content/deployment lock | `internal/state/` |
| Embedded starter source | `templates/` |

Dependencies point toward site/content primitives, not toward command handlers.
`cli` calls `setup` and `site`; `setup` uses `release`, `site`, and shared admin
option validation. `admin` uses `host`, `release`, `site`, and `fleas`. `host`
uses `site` and `fleas`. The site package never imports CLI, setup, or admin.
`fleas` does not know about HTTP, SSH, templates, or deployment orchestration.

Only CLI/setup imports serverbundle; the server never embeds copies of itself.
Build both Linux servers before embedding their compressed bytes and hashes into
the CLI. Setup validates extraction before reset confirmation/upload. There is
no runtime download/compiler fallback. scripts/build.sh refreshes generated assets.

## Loading and serving

`site.Load` validates the runtime tree, reads JSON, scans templates, resolves
routes, compiles partials/pages, snapshots assets, and checks rendering. All
partials are checked, including unused ones. Routes and flea page ownership use
the same filename conversion. Legacy named fleas and `data.*` share a namespace
and one JSON-directory loader; collisions are errors.

A `Project` is immutable after publication. Its exported maps exist for internal
inspection, not concurrent editing. HTTP handlers select from validated maps;
they never turn request paths into filesystem reads. Template output is buffered
before sending headers, so a render failure can still produce a proper 500.

The template compiler translates FleaCMS directives into Go `html/template`.
`{flea ID}` becomes a sanitized rich-content call. The HTML region scanner
recognizes the same unquoted markers for extraction, editor access and deployment
pruning. Quoted markers are rejected. HTML comments and script/style contents
do not contribute references; comments never execute template instructions.
Go retains responsibility for contextual escaping. JSON is passed as values,
never compiled as template source. Rich fleas are the narrow exception to normal
text escaping: the HTML allowlist reconstructs them first, and only the result
becomes `template.HTML`. SVG icons have a separate geometry-only validator.
Keep these two sanitizers separate: their permitted markup and uses differ.

There are three intentionally distinct loading entry points:

- `Load` / `LoadWithFleas`: read-only; no downloads or source changes.
- `LoadDev`: may fetch missing icons; also used on a temporary committed snapshot
  during deployment. It does not extract fleas.
- `LoadEditing`: extracts local editable regions, then calls `LoadDev`. Only the
  local dev command and watcher should use it.

Reload builds a replacement snapshot outside the request lock. A writer mutex
serializes reloads; the shorter state lock publishes the result. Invalid edits
publish an error state, not an apparently current old page. The prior snapshot
remains available for a custom error response. Production has no watcher or
reload endpoint. Host aliases for one release share one serving snapshot and
must select the same content file.

## Filesystem boundaries

`sitefiles.Included` is a runtime inclusion policy, not a public-URL policy.
Configuration and content belong in an archive but remain private over HTTP.
Do not combine this predicate with the server's URL checks.

Project files are untrusted. `site/files.go` pins directory handles, rejects
symlinks/nonregular files, and compares opened file identities. Writes use an
exclusive temporary file in the pinned parent, sync, rename, then sync the
directory. The icon cache uses these same write primitives.

Administrator-owned configuration has a different boundary. `admin.atomicWrite`
works with trusted absolute paths; `realDirs` rejects linked managed directories.
`fleas.ReadFile` pins an administrator-selected content directory and checks its
file. These helpers are deliberately not forced into one general filesystem
abstraction. Changing their trust assumptions requires tests, not just deduplication.

## Two separate transactions

### Local extraction

The HTML tokenizer records byte ranges, so extraction preserves source formatting
outside the region. The extractor reserves IDs against active and archived content,
checks all regions, and builds edits before writing anything.

While holding the in-process content mutex, it saves `_/.fleas-pending.json`, writes content,
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
   active fleas, saves the previous config, and removes the journal. Cleanup
   failure keeps the journal and reports that the release is already active.

Flea references are lexical across all templates, not collected from one render.
Otherwise empty loops or conditional sections could incorrectly lose content.
Browser saves share the deployment lock, reject pending recovery journals, verify
the active release/content pair, and compare a content digest from the submitted
form. They sanitize and validate a replacement snapshot before atomically saving
and publishing it. Local saves use the extraction lock instead. HTTP requests never
choose a filesystem destination. Sessions persist independently of releases.

Authentication/session persistence belongs to `editor`, content persistence to `editstore`, and delivery to
`mailer`. The only email-provider dependency is the Sender interface; only Resend
is implemented. Credentials remain in the root-only systemd EnvironmentFile, not
site snapshots. The service can write content/private runtime directories and the shared lock,
but not code releases or host configuration. This is a change from the previous
read-only service model and requires the updated installation permissions.

Dev always mounts the editor so a fresh site's /admin can explain owner setup.
The editor enforces loopback host checks before showing local instructions.
The dev admin wrapper selects the current validated access list and reads local
mail settings per request. A serialized handler replacement retains valid sessions
and discards pending codes when settings change, with a matching edit-store allowlist. Invalid snapshots
or unreadable mail settings fail closed. Production configuration remains fixed
until restart/deploy; local contact mail remains startup-bound.
Production mounts it only for sites with approved owners. Setup and missing-mail
pages grant no editing access; the normal email-code flow is still required.

The editor passes a page state and plain data to a renderer supplied by the
site's Server. Navigation derives from the validated public route list, not the
flea store: a page needs no fleas to edit its metadata. Query parameters select
one validated route at a time, never a filesystem path. Shared editing is separate.
The renderer selects a precompiled developer-owned admin/*.html
template and its selected layout from the current immutable snapshot. Values are
contextually escaped, never compiled as source. The admin directory is excluded
from public routes and assets; authentication remains in the editor handler.
Missing templates fail without an embedded fallback. This shares site chrome and
reflects template reloads without an editor dependency on site.
Admin styling permits HTTPS CSS/fonts/images. Only the built-in activity script
is permitted by a hash in CSP; other scripts remain blocked. Same-origin referrer policy preserves browser form origins;
origin and CSRF validation remain exact.

### Session runtime state

Both environments use a private `_/sessions.json`: below the project locally,
below `/var/lib/flea/sites/<name>/` in production. Production's `_/` is service-owned
0700, beside content/releases, not inside either. Code and parent directories remain
administrator-owned. Deployment protocol 10 requires these runtime permissions.
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
activity POST, not an idle heartbeat. The script is injected with exact bytes and
hash-pinned in CSP; no dynamic JavaScript interpolation. OTPs/rate limits remain
bounded in-memory state. Browser/device tokens are independent and host-bound.

The loader resolves leading flea settings comments before compiling partials.
Selected layouts are private HTML files removed from the partial namespace;
they may render partials but cannot select another layout. Only the layout
compiler accepts slots. Unknown/duplicate settings and invalid targets fail load.
The admin starter selects _layout_admin.html; public _layout.html is independent.

`flea email` uses the setup-only mailer.Checker interface for a non-sending
connection check. Resend implements both Checker and Sender; auth/contact code
depends only on Sender. Future provider adapters should use these boundaries.
The localenv package handles bounded, literal .env reads and owner-only writes,
rejects links, and preserves unrelated settings. Setup refuses Git-tracked .env
files and ensures an ignore rule before saving secrets. Dev and deployment mail
setup merge only mail settings with process-environment precedence, without
changing process environment. Production still reads systemd's mail.env.

## Review findings and decisions

This refactor addressed mixed responsibilities in flea handling, icon-specific
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
- Provide a nondestructive server upgrade path. Setup currently offers a reset,
  not an upgrade. Publishing release artifacts also remains operational work.
- Decide and implement deployment-config consolidation separately. Settings still
  live in `_/deploy.json`; this refactor does not migrate user files.
- Local dev uses pinned fsnotify directory notifications, registering included
  subdirectories and reconciling new/moved/deleted directories after a 100 ms
  debounce. Content fingerprints run only at startup or after events, suppressing
  duplicate reloads from extraction's own writes. Overflow triggers a rescan;
  watcher failures stop dev visibly, without a polling fallback. Reload still
  validates a complete snapshot. Browser reload notification remains a separate
  700 ms HTTP poll; the admin does not auto-refresh unsaved forms.
- Keep the published `Project` immutable. Consider hiding exported maps if more
  in-process consumers appear; current callers do not need a broad accessor API.
- Releases, content snapshots, and archived fleas are retained. A retention policy
  and rollback CLI need explicit product behavior and separate tests.
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
