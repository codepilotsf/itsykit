# Setup, deployment and recovery

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

For the recorded public-release procedure and constrained production host settings,
read the [deployment runbook](deployment-runbook.md). Recheck live state and
obtain authorization before applying host changes or publishing.

## Recovery

`flea update` changes only the local CLI and bundled servers using bounded HTTPS
downloads, checksum/platform/version checks and same-directory replacement.
It shares `.flea-install.lock` with the installer. Preserve the old executable on
failed validation. `flea update-server` explicitly updates the remote executable;
never add an `upgrade` alias or trigger remote updates from local updates.
After confirmation, update/recovery shows a simple terminal spinner through
upload and execution; redirected output uses plain progress text. Clear animation
before diagnostics, interactive sudo prompts and final output, including failures
and cancellation. Preserve confirmation, native SSH and upgrade safeguards.

Read docs/recovery.md for commands, data boundaries and limits.
Rollback selects reachable per-site receipts, never directory ordering or the
whole-server previous config. It republishes code while retaining live editable
state and current host policy. Backup captures live overlays under deploy.lock;
restore checks an expected state digest and requires explicit CLI confirmation.
Keep credentials, certificates, sessions and deployment mappings out of backups.
Upgrade validates pinned executable bytes, journals an executable-only swap and
restores on failed restart/probes. Pending upgrades block deployments and saves.
Do not add format or service-unit migrations under the guise of upgrading.

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
Build CLIs on Apple Silicon macOS or x86-64 Linux; Linux ARM64 CLI builds are
no longer supported. ARM64 Linux servers remain supported through cross-compilation,
with native server acceptance recorded separately.
Build with bash scripts/build.sh: refresh both CGO-disabled Linux variants before
embedding into the native CLI. Plain go build may embed stale assets. CI still
builds/tests on native Linux/Mac runners; see docs/agents/releasing.md.
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

### Host indexing

Read [hostname indexing](../host-indexing.md). Optional server `indexing` maps
exact registered hosts to `noindex` or `allow`; omission adds no restriction.
Use `flea server indexing [ssh-target] <hostname> <noindex|allow>` for explicit,
reviewed changes. It shares the deployment lock, candidate validation, journal,
restart and recovery transaction with removal. A fresh revision prevents stale
readiness. Site/removal review digests include every owned indexing entry.
Policies persist on ordinary deploy, sync, rollback and site restore; backups
exclude them. Removal clears only owned aliases, and redeployment after removal
does not recover prior policy. Full journal recovery restores prior host config.
Protocol 27 requires matching clients/servers; older typed writers could discard
this optional map. No migration or site flag is added. Keep release/server update
and CodePilot consumer fallback removal separately authorized and ordered.

### Deployment and HTTPS

Local deployment and setup use private, unique scratch directories selected by
Go (`TMPDIR` on Linux/macOS), with deferred removal on completion and handled
cancellation. Preserve independent timeout contexts for remote/SSH cleanup.
Storage errors retain their cause and identify the local temporary location.
Do not fall back after an explicit temporary-directory setting fails or clean
other invocations' leftovers. Remote staging remains separate. See
[development storage](../development-storage.md) for disk-backed job placement.

Optional host `downloads` maps configured hostnames to absolute administrator-owned
release stores outside site snapshots. Aliases must match; normal deployment keeps
this host policy. See [download hosting](download-hosting.md) for publication,
verification and rollback boundaries. Never delete retained downloads to shrink a
site archive, or include the store in site backups. The 256 MiB archive limit remains.

Deploy/https select only the current Git branch's saved destination from
.flea.config.json's deploy.branches map. Use the actual branch name without a production
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
The server chooses activation progress from its installed TLS domain list: show
certificate setup guidance only when enabling a new domain, otherwise a concise
site-health/HTTPS check. Keep verification and renewal behavior unchanged.
Ask for the SSH username (default root) only when the target does not include
user@host; save the explicit username. Existing bare targets ask only for the
missing username. Do not ask about NAT/proxy IP lists during normal setup; retain
expected_ips as an advanced JSON/flag setting. Saving occurs before the clean check; the user must
commit the config or explicitly ignore it. The wizard can run before the initial
commit so settings and site files can be committed together. Upload still requires
a commit and clean status. Non-Git and detached HEAD cannot configure deployment.
Use -configure to confirm and save changed settings; CLI flags cannot silently
override an existing mapping. Do not accept flat legacy deployment settings.
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
The deploy section contains settings only, never credentials, and is stripped
from the uploaded config. Runtime app/admin/apps settings are retained.
Top-level site `redirects` is runtime config too: preserve it in the committed
snapshot/archive and validate its destination routes on both sides. Protocol 28
requires matching clients/servers because older servers silently ignore these
rules. Keep server updates separate from site deployment.
Deploy packages committed site/ pages/assets/partials, .flea.config.json,
mailers/, flea/snips.json, flea/media/, flea/data/, legacy flea/snips/, flea/icons/,
and each custom app's `flea/apps/<name>/app.json` with exactly the executables it
declares. Validate the actual packaged snapshot locally and the extracted release
remotely. Reject symlinks, hard links, special files, duplicate/traversing
archive paths, and releases exceeding 256 MiB or 10,000 entries. A malformed app
manifest or an archive containing an unlisted package file fails extraction.
App executables are extracted mode 0555 so the `flea` service user can read and
execute them without owning the root-owned release tree; package directories stay
administrator-owned. Custom-app storage is stable per-site runtime state outside
the release, so it is never packaged and survives a release change; it is not
backed up automatically.

Deployment administration runs through SSH/sudo via the installed server binary,
never HTTP. The browser editor has content-only permissions, not deployment powers.
Protocol 26 requires persistent-session runtime directories, the bundled Contact
child/runtime, custom-app packages extracted at service-readable mode and current
site/config syntax. Old top-level contact settings are rejected, not migrated.
Before upload the CLI probes the server architecture and rejects a package that
does not declare the server's `<goos>-<goarch>` executable, or whose declared
target executable is not a native binary for that platform. The check runs
against the immutable staged snapshot, not the mutable working tree. Build each
matching Contact executable before its parent. Verify child launch/cleanup under
the real systemd unit separately; local tests do not establish native Linux or
email acceptance.
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
Retain releases and server.previous.json; rollback uses per-site history receipts.
Automatic pruning is deferred.
On normal deployment, rich snips are committed seeds only. Preserve remote content by ID and remote
page metadata, seed new IDs, and restore reintroduced IDs from archived content.
Store content outside releases in sites/<name>/content/<release-id>.json;
host.Config.Content maps hostnames to absolute snapshots. Activation journals
switch both paths; failed activation leaves old content intact. After successful
verification, prune unreferenced active IDs into archive, retaining prior snapshots.
Scan lexical references in all templates, including unused partials and false/empty
branches; never prune based on rendered output. No automatic archive pruning.
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

### Server-wide site administration

`flea server list [ssh-target]` lists installed sites without a local project:
each site's name and branch, then HTTPS domains, then HTTP-only domains, then
redirects, in deterministic name order, without the release or repeated plain
domain lines. It is read-only and may inspect a server with a pending recovery
journal. `flea server remove [ssh-target] [site-name]` unregisters one site after
typing the exact name; when the site name is omitted, it lists the managed sites
and asks for one by its exact name.

Removal is unregister/offline only: it deletes no files. Site files remain under
`/var/lib/flea/sites/<name>` and cached certificates under `/var/lib/flea/certs`,
but retained content is not restored automatically.
Re-deploying the same stable name starts from the current project; because
content lookup follows active routing, it cannot recover content left behind.
Back up or pull the site before removal if a later deployment must keep its
current content. Removal restarts the shared service, so other sites are briefly
unavailable. No restore feature is added.

`flea server` connects only over native SSH. Give the target explicitly, or omit
it to be prompted for the server login; an SSH alias or bare host is kept as-is.
It never needs a local config, Git repository or site. The target must run the current
deployment protocol; an older or incomplete server is reported and no remote
configuration is rewritten by hand. Non-root targets use the existing sudo
policy. Validate the target and site name, and single-quote every remote
argument.

The server computes an expected-state digest over exactly the routing and
ancillary entries removal would change. Review prints it; removal recomputes it
under the existing deployment/content lock and refuses if the site changed since
review. The operation blocks on a pending deployment or executable recovery and
removes only the target's `sites`, `content`, `downloads`, `indexing`, owned
redirects, TLS allowlist entries and site/branch/commit bookkeeping. The digest
covers branch and commit, so a commit-only change since review also fails closed.
Other sites' routing and ancillary entries are preserved. Removing the site that
owns the last HTTPS hostname clears the TLS allowlist and disables automatic
HTTPS, while cached certificates stay on disk; removal that leaves other HTTPS
hostnames preserves those TLS settings. A redirect that touches the target at
only one end, or a site whose root is outside the managed data directory, fails
closed.

Removal reuses the existing atomic config publication, recovery journal,
restart/probe and failure rollback paths. The prior config is journaled to
`deploy.pending.json` before replacement; a failure restores it. Because config
activation restarts the shared service, other sites are briefly unavailable. An
interrupted removal is recovered with `sudo flea-server -recover`, exactly like
an interrupted deployment. Removing the last site is supported and probes generic
server readiness. Removal never prunes retained content or media.

## Source CSS contract

Protocol 24 publishes committed `site/css/theme.css` with the code release on
normal deploy. No live theme overlay or CSS sync plan exists. Rollback selects
the target release's CSS while retaining current snips; backup/restore includes
CSS as source. Servers with a removed `themes` mapping must be reviewed before
upgrade: preserve remote customizations explicitly in Git using the old runtime
first. No migration, compatibility mapping or automatic theme deletion is added.

## Blog

Read [Blog](../blog.md) for the optional built-in blog, JSON and template contract.
Posts share the snip/media snapshot, atomic save and preservation/synchronization
paths. Drafts remain outside public routes and template data. The blog admin uses
existing login, exact Origin, CSRF, upload bounds and content-version checks.
