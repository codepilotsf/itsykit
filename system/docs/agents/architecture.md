# Architecture and review guide

The parent repository has `system/` (this Go module) and independent `website/`.
Only `templates/` is embedded for generated sites; never overwrite the public site
with the starter. Run framework commands here. Preserve historical releases/tags.

## Loading and serving

`site.Project` is immutable after publication. The loader reads selected config,
private mail templates, Blog storage and developer JSON, resolves validated routes
and layouts, compiles developer templates, and snapshots assets. Reload builds a
complete new Project. Failed loads close unpublished handles. Requests select only
validated maps and pinned file handles; never read request-derived filesystem paths.
Large static files and managed cover photos stream from validated handles with
size/mtime/link checks; atomic replacement preserves old readers. Small files and
source theme CSS retain byte snapshots. Protect descriptor lifetimes and budgets.

Ordinary page metadata belongs in one leading HTML settings comment, before markup:

```html
<!-- itsy
title: Visit the bookshop
description: Find our opening hours and plan a visit.
layout: /custom.layout.html
-->
<h1>Visit the bookshop</h1>
```

Each value is single-line plain text, without surrounding quotation marks. Split
at the first colon; later colons and braces stay literal. `layout` is optional and
must name an existing site-relative private `*.layout.html` file. The existing
`<!-- flea` spelling is also supported. Unknown and duplicate settings fail
validation, including duplicate empty values. This is not YAML.

`{page.title}` and `{page.description}` render with contextual escaping, for
example `<title>{page.title}</title>` and
`<meta name="description" content="{page.description}">`. Missing or empty title
uses `Home` at `/`, otherwise the final URL component with hyphens replaced by
spaces and its first letter capitalized. Missing or empty description is empty.
Comment values are never compiled as template source or copied into JSON.
Blog posts use their saved title and excerpt instead of template comment metadata.
There are no saved ordinary-page metadata overrides or page metadata editor.

Developer JSON is data, never executable template source. The shared brace lexer
preserves literal groups and source offsets. Only Blog body HTML is sanitized and
trusted in ordinary element content; attribute/raw-text/foreign contexts fail.
Removed page attributes/directives fail validation, with no extraction writes.

## Blog and authentication

`internal/snips` retains its historical internal name and `itsy/snips.json` path,
but its Store has only Blog records and the photo library. No Page/Entry/archive/
collection/video catalogue model remains. `site/blog.go` constructs published
views/routes; drafts never enter public template data. `editor/blog.go` owns
private forms, rich editing assets and uploads; `editstore/blog.go` owns mutation.
Public pages are identical for signed-in visitors, with no authenticated rendering
path, toolbar, wrappers or public admin context. `/admin` selects the Blog entry.
See [editor](editor.md), [Blog](../blog.md) and [photos](../photos.md).

Admin retains bounded single-use OTPs, exact Origin/CSRF/Host/HTTPS, rate limits,
host-bound cookies and durable mode-0600 hashed sessions. Runtime and content locks
remain separate. Renew authenticated navigation/activity, never idle timers. Logout
and access removal revoke sessions; validation is read-only. Host startup prunes.
Local previews are private bounded capabilities; production never enables them.
Resend Sender and non-sending Checker are separate; secrets stay out of logs/Git.

## Transactions and preservation

Local Blog saves and reload use the in-process content mutex before reloadMu.
Production saves and activation share deploy.lock. Refuse pending recovery, changed
release/content pairs and stale form versions. Write immutable processed cover bytes,
validate the entire candidate, durably replace content once, then publish. A receipt
is captured under that lock. Failures preserve drafts and may retain orphan photos.
There is no page extraction or replay transaction; old pending journals fail closed.

Normal deployment stages new immutable code and a separate content snapshot,
preserving remote posts/photos and adding local-only IDs. Check an explicit sync
plan's expected remote state under lock before writing. Candidate host validation,
readiness, recovery journal, atomic config publication and failed-probe rollback
retain their ordering. Never mutate active code or prune retained content/media.
Rollback selects code/CSS from per-site receipts and retains current Blog data.
Backup/restore exclude secrets/sessions and validate bounds, identity and digest.
Explicit executable upgrades preserve verified replacement and recovery safeguards.

### Blog content synchronization

Commit `itsy/snips.json` when seeding Blog posts or cover metadata. Normal
`itsy deploy` preserves live posts and photos by stable ID and adds source-only
IDs. Omitting an ID never deletes a live post or photo. Publish code/CSS normally.

```sh
itsy pull
# Review and commit local Blog content and photos.
itsy deploy --with-content
```

`pull` retrieves Blog posts and cover photos from this branch's saved destination,
keeps local-only IDs and never pulls CSS, templates, configuration or sessions.
Stop dev first; selected local files must be committed. Explicit publication
requires a destination-scoped pull baseline for existing sites. Posts and photos
compare as whole records; conflicting edits require an explicit choice and stop
noninteractively. The expected remote digest is checked under the deployment lock.
`itsy status` reports content differences read-only. `--with-snips` is removed
without an alias. This contract requires deployment protocol 30.

Backups include Blog content and library photos. Rollback preserves current
content while selecting earlier code/CSS. Restore requires an expected snapshot
and explicit confirmation. Retained files are never pruned automatically.

## Other trust boundaries

`sitefiles` shares runtime inclusion and confined package rules across loading,
watching, Git snapshots and archives. Dotpaths, agent guidance and private runtime
are excluded. `site/`, `mailers/`, selected config, developer data, Blog content,
media and icon caches are included. Legacy filenames are selected once per operation;
never rename/migrate automatically. No directory traversal or symlink fallback.

Contact and custom apps retain `appapi`, `appruntime`, `appmanager` and private
capability delivery. Never forward browser cookies/Authorization as service credentials.
Anonymous visitor CSRF and app-owned sessions remain separate from admin sessions.
Child executables use validated packages, target hashes, inherited credential pipes,
liveness and bounded restart/response behavior. Loading/status never launch them.
Managed runtime is stable outside immutable releases. Apps do not authorize admins.

SEO/redirects and exact host indexing remain independent policies; never derive
production origins from requests, overwrite custom discovery/canonicals, discard
host policy on deploy/recovery, or weaken Host/TLS readiness. Native fsnotify watches
included directories with debounce and overflow recovery, never idle polling.

Build with scripts/build.sh, serially refreshing Contact and both embedded Linux
servers before the CLI. Self-update verifies bounded platform/checksum downloads
under the installer lock. Release/production actions require separate authority.

## Verification

Run go test ./..., go test -race ./..., go vet ./... and bash scripts/build.sh.
Use isolated HTTP fixtures and fresh minimal/starter sites. Real Ubuntu/systemd,
public TLS issuance/renewal and real email need separately authorized environments;
local checks do not establish those acceptances. Never weaken permission checks.
