# Nanoweb next steps

## Backlog: independent site runtime versions

Deferred by agreement: keep the current single shared NanoWeb server for now.
Do not introduce proxy/runtime separation as part of current work.

Future option: a lightweight host handles HTTP/HTTPS, certificates and domain
routing, while each deployed site carries its matching NanoWeb runtime for
templates, snips and /admin. Separate systemd-managed processes could communicate
through private Unix sockets, allowing sites to use different runtime versions.
Deployments could health-check a replacement before switching traffic and retain
the previous runtime alongside its release for future rollback.

Revisit only when independent versions justify the added process supervision,
memory, permissions, stable host/runtime contract and content-migration handling.
This is a backlog proposal, not implemented behavior.

## Owner dashboard

Implemented /admin email-code login and basic page/snippet editing. Configure
admin.emails in _/nanoweb.config. Resend credentials are server-only in
/etc/nanoweb/mail.env, prompted on deployment when missing. Local dev uses
RESEND_API_KEY and NANOWEB_EMAIL_FROM environment variables. No real email was
sent during implementation; live Resend and Ubuntu/systemd checks are still needed.
Protocol 7 requires the new binary, service unit, service-owned deploy.lock and
managed content-directory permissions. Do not reset an existing server as an
upgrade. See README's Owner dashboard section. Visual rich-text controls, image
uploads, revision history, rollback, and a nondestructive upgrade workflow remain
next steps. Tests cover auth, CSRF, expiry/replay, content conflicts and locking.

## Earlier implementation notes

Snip redesign agreed: extract inner HTML from snip="label" regions during dev
into private _/snips.json, keyed by route with shared regions and page metadata.
Use sanitized HTML by default and 10-character crypto-random alphanumeric IDs.
Extraction, short IDs, sanitized rendering, page metadata, external deployment
content snapshots, preservation, archived orphan pruning, and recovery tests are
implemented. New starters use _/snips.json; legacy named JSON stays supported.
Structured data now uses _/data/*.json with data.*. Deployment protocol 7 is
required; update the server binary before deploying rich snips. No live server
or example site was modified. A rollback CLI, remote sync,
schema migrations, and consolidation of deploy settings into nanoweb.config
are still pending.

Latest change: setup extracts its matching compressed Linux binary from the CLI,
checks the bundled SHA-256 and static ELF architecture, then uploads over SSH.
Build with bash scripts/build.sh to refresh both Linux variants and the native CLI.
Development setup no longer requires publishing or a release download service.
This supersedes older runtime cross-build and download notes below. See RELEASING.md.
Next: test real Ubuntu installation on a disposable server. Release runner setup
and publication can happen separately when distribution is needed.
No release or runner configuration has been published by this local work.

Updated after adding branch-specific deployment, committed snapshots, DNS preflight,
and saved WWW/HTTPS preferences.

## Current state

Nanoweb has a runnable local site workflow:

- `nanoweb new <site-name>` creates the embedded starter site and initializes Git
  on main without committing. --skip-git supports existing parent repositories.
  A starter .gitignore is included; missing Git does not block site creation.
- `nanoweb status` validates a local project.
- `nanoweb dev` serves a local site with automatic reloads.
- File-based routes, static assets, layouts and `{slot}`, partials, escaped JSON
  snips, dotted paths, conditions, loops, custom error pages, and on-demand
  pinned Lucide icons are implemented.
- The starter templates are ordinary files under the repository's `templates/`
  directory. They use the slate circled lowercase `n` logo and the current
  Prewind/Google-font theme.
- `/usr/local/bin/nanoweb` is a symlink to this checkout's `bin/nanoweb`, so
  rebuilding `bin/nanoweb` updates the global command without reinstalling it.

The persistent server, deployment, and HTTPS are implemented. They have not yet
been exercised against a real remote Ubuntu machine or live Let's Encrypt:

- `nanoweb setup` prompts for an SSH target, verifies supported Ubuntu/systemd
  and architecture, extracts the bundled `nanoweb-server`, uploads it, and installs a
  hardened `nanoweb` systemd service.
- New installations default to public HTTP on port 80 (`"listen": ":80"`).
  The service is non-root and receives only `CAP_NET_BIND_SERVICE` to bind the
  low port.
- `nanoweb deploy` validates/packages the site, uploads over SSH/SCP, and invokes
  the installed server to activate a release. `_/deploy.json` saves approved
  destinations per Git branch. The server validates again, atomically switches configuration,
  restarts, and checks the revision/hostname. Failure restores the prior config.
- `nanoweb https` enables automatic Let's Encrypt certificates for deployed
  domains. `nanoweb deploy -https` combines the two. Contact email and explicit
  acceptance of the CA terms are required.
- Host configuration lives at `/etc/nanoweb/server.json`. Its `sites` map
  connects a request hostname to an absolute site directory. It starts empty,
  so only `/_nanoweb/health` works immediately after setup; other requests
  return 404 until a site is deployed.
- Setup does not change cloud or host firewalls. TCP 80 and 443 must be allowed
  by the administrator. DNS records must point at the server. Setup remains a
  first-install operation; there is no automatic server binary upgrade command.

## Commands to use next

With a current server installed, run from a committed site on a named Git branch:

```sh
nanoweb deploy
```

First use asks for the server/domain, uses the branch name directly, and saves
WWW redirect and HTTPS preferences (HTTPS defaults true). Commit the saved
settings before retrying. Subsequent deploys use the current branch's mapping.
The remote site name is generated automatically; there is no environment or
remote-name question. To change the mapping:

```sh
nanoweb deploy -configure
```

Deploy asks for an SSH username (default root) if the target has no user@ prefix.
Existing bare targets ask only for that missing username and save it. Commit the
config change before retrying. There is no NAT/proxy IP question in the normal
wizard; expected_ips remains an advanced config/flag option.

Once SSH succeeds, a missing server installation offers setup and then resumes
deployment using the same target and SSH connection. Go and the source checkout
are required locally for that installation. Failed SSH authentication does not
trigger setup, and older installations are not overwritten.

Plain HTTP can use an IPv4 address; HTTPS requires DNS names. Example saved
`_/deploy.json` (see README for a full main/staging example):

```json
{
  "branches": {
    "main": {
      "target": "ubuntu@1.2.3.4",
      "domain": "example.com",
      "name": "my-site-main",
      "www_redirect": true,
      "https": true,
      "email": "you@example.com",
      "agree_tos": true
    }
  }
}
```

This file is excluded from upload and contains no credentials. Flags seed the
questions only on first use or with -configure; they cannot silently override
a saved mapping. Legacy flat config requires confirmation before migration.

Git must be clean before SSH/upload/activation, including staged, unstaged, and
untracked files. The wizard can save settings before the first commit, allowing
site and settings to be committed together. Upload still requires a commit.
Detached HEAD and non-Git folders cannot configure deployment.
Snapshots come from commit objects; missing icons resolve only in the temporary
snapshot. Ignored working files cannot be uploaded. Subfolder sites are supported.

New staging mappings suggest staging.<main domain> but require approval. Both
local config and server metadata prevent cross-branch destination collisions.
The server checks A/AAAA records, including WWW, before upload. Every answer must
match its interface addresses or explicitly saved expected_ips for NAT/proxies.
IP literals skip DNS. This is not a public firewall/reachability test.

## Verification already run

The current implementation passed:

```sh
GOCACHE=/tmp/nanoweb-go-cache /usr/local/go/bin/go test ./...
GOCACHE=/tmp/nanoweb-go-cache /usr/local/go/bin/go test -race ./...
GOCACHE=/tmp/nanoweb-go-cache /usr/local/go/bin/go vet ./...
GOCACHE=/tmp/nanoweb-go-cache bash scripts/build.sh
GOCACHE=/tmp/nanoweb-go-cache /usr/local/go/bin/go build -o bin/nanoweb-server ./cmd/nanoweb-server
```

The full suite includes the executable starter/dev/reload smoke test, Linux
amd64/arm64 cross-builds, simulated SSH/SCP failures, archive security tests,
real HTTP deployment/redeployment and recovery tests, and TLS hostname isolation.
Branch tests cover saved first-use answers, future reuse, staging suggestions,
detached HEAD, destination collisions, dirty staged/unstaged/untracked changes,
commit-only snapshots, subfolder sites, ignored files, and committed symlinks.
DNS tests cover mismatched A/AAAA answers, timeouts, no records, WWW, IP bypass,
and stopping before upload on failure. Redirect and saved HTTPS preference tests
verify that other sites are preserved. Full tests, race tests, vet and builds
were rerun successfully after these additions.
A local ACME authority tests actual HTTP-01 challenge handling, CSR issuance,
trusted HTTPS, certificate cache permissions, and cache reuse while the CA is
offline. Race tests also passed. These tests do not prove a real Ubuntu install,
SSH password flow, systemd permissions, public reachability, live CA issuance,
or renewal over a real certificate lifetime.

## Decisions made while you were away

- HTTPS lives inside Nanoweb using pinned `golang.org/x/crypto/acme/autocert`
  v0.57.0, as required by spec.md. No reverse proxy or certificate shell scripts.
  This dependency raises the build requirement to Go 1.26; the installed Mac Go
  is 1.27.1. The dependency versions are recorded in go.mod/go.sum.
- HTTPS is enabled explicitly per domain. HTTP-only sites can coexist on the
  same host. HTTPS-enabled domains redirect with 308; unknown TLS hostnames and
  Host/SNI mismatches are rejected. No HSTS, wildcard, or IP certificates yet.
- Certificates/account keys persist under `/var/lib/nanoweb/certs`, owned by
  nanoweb with mode 0700. The daemon can write only that directory; the data
  parent, releases, and host configuration are administrator-owned.
- Deploy uses separate release directories and an atomic host config switch,
  followed by a brief service restart. It does not promise zero downtime.

```text
/var/lib/nanoweb/sites/<site-name>/releases/<timestamp-random-id>/
```

- All aliases belonging to the same site move together. Deploy cannot steal a
  hostname from another site. Removing aliases is a manual config change.
- Only runtime files are uploaded: , _/nanoweb.config, _/snips/, _/icons/.
  Archives are bounded at 100 MiB and 10,000 entries, with no links or traversal.
- Committed snips ship as a complete snapshot. Remote content edits are not merged;
  per-edit content history and visual rich-text controls remain deferred.
- A lock serializes activation. A durable recovery journal preserves the old
  config through interruption. If `/etc/nanoweb/deploy.pending.json` remains,
  further deploys stop until `sudo /usr/local/bin/nanoweb-server -recover` succeeds.
- Prior releases and server.previous.json remain available. There is no release
  pruning, rollback CLI, or automated server upgrade in this change.

## Next work: real-server verification

The remaining operational check needs an actual target and domain:

1. Run `nanoweb setup` against it.
2. Allow inbound TCP 80/443 in the provider and host firewall, and point the
   domain's A record at the server. Any AAAA record must also work.
3. Deploy a test site with the commands above, then enable HTTPS.
4. Confirm `http://SERVER_IP/_nanoweb/health`, HTTP-to-HTTPS redirect, valid HTTPS,
   pages/assets, and private-path rejection from another machine.
5. Check `sudo systemctl status nanoweb --no-pager` and service logs.
6. Redeploy changed content, verify restart/cache reuse, and check recovery.
7. Monitor real certificate renewal. Avoid repeatedly requesting certificates
   while debugging DNS/firewalls because CA issuance limits apply.

Older installed server binaries lack deployment protocol 7. They must be updated
deliberately before deploy works; rerunning setup refuses an existing installation.
README.md documents the executable/unit/certificate-directory requirements.

After the live verification, likely next improvements are a server upgrade
command, explicit rollback, release retention/pruning, and remote status/logs.

## Working-tree note

Developer-facing output must always use simple, friendly English and readable
line breaks. This standing guideline is recorded in AGENTS.md.

The deploy/HTTPS implementation, dependency files, tests, documentation, and this
handoff are uncommitted. Earlier setup/HTTP work was already clean in git when
this implementation began. README.md, AGENTS.md, spec.md, and the generated-site
README were updated. Both binaries were rebuilt, including the globally linked
CLI. No commits, pushes, real server changes, or live CA registrations were made.
# Root-level site layout

Pages, layouts, partials, CSS, and images now live directly in the site root.
The starter mirrors this layout. Private JSON stays in _/. Shared exclusions
keep Git metadata and project docs out of serving, reloads, and deployment.
Deployment protocol is now 3. Existing sites need html/ contents moved to the
root after checking for collisions; existing servers need the updated binary.
No existing example site or remote server was automatically migrated.
