# Setup, deployment and recovery

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

## Recovery

`flea update` changes only the local CLI and bundled servers using bounded HTTPS
downloads, checksum/platform/version checks and same-directory replacement.
It shares `.flea-install.lock` with the installer. Preserve the old executable on
failed validation. `flea update-server` explicitly updates the remote executable;
never add an `upgrade` alias or trigger remote updates from local updates.

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
from the uploaded config. Runtime app/admin/contact settings are retained.
Deploy packages committed site/ pages/assets/partials, .flea.config.json,
mailers/, flea/snips.json, flea/media/, flea/data/, legacy flea/snips/, and flea/icons/. Validate
the actual packaged snapshot locally and the extracted release remotely.
Reject symlinks, hard links, special files, duplicate/traversing archive paths,
and releases exceeding 256 MiB or 10,000 entries.

Deployment administration runs through SSH/sudo via the installed server binary,
never HTTP. The browser editor has content-only permissions, not deployment powers.
Protocol 22 requires persistent-session runtime directories, the contact
runtime and unquoted snip syntax.
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
