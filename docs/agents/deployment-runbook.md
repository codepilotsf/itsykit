# FleaCMS release and deployment runbook

Updated October 3, 2026 for disk-backed download hosting. This is an operator guide,
not authorization to publish or change a server. Recheck paths, branch mappings,
host settings and live versions before each release. Read the repository's
AGENTS.md, [release guide](releasing.md), [deployment requirements](deployment.md)
and [recovery guide](../recovery.md) first.

## Repositories and destinations

- Framework: `/Users/sean/Projects/flea`, branch `main`.
- Public website: `/Users/sean/fleacms-website`, branch `main`. This is separate
  from the framework's `website/` source directory. Preserve its custom layout.
- Production SSH: `root@fleacms.com`. The shared `flea.service` also serves
  `bookworm.fleacms.com`; an executable upgrade restarts all hosted sites.
- Website deployment uses the current branch's saved `.flea.config.json` mapping.
  Confirm it targets the intended host/domain before running anything.
- Public binaries: `https://fleacms.com/releases/<version>/`.
- Stable channel: `https://fleacms.com/releases/latest.txt`.
- Installer: `https://fleacms.com/install.sh`.

Never copy framework source, SSH keys, private runtime state or backups into the
public website. Do not change Bookworm content as part of a marketing-site release.

## 1. Scope, version and source

Inspect `git status --short` in both repositories. Preserve unrelated work,
including pending planning changes. Obtain permission for commits, pushes,
tags, publication and any new persistent host changes. Stage explicit files,
not the entire worktree. Deployment requires a clean, committed website snapshot.

The canonical version is `internal/buildinfo/version.go`:

- During 0.x, use a patch for fixes (0.6.0 to 0.6.1), and a minor for features or
  breaking changes (0.6.x to 0.7.0). Do not infer that a release should become 1.0.
- After 1.0, SemVer uses patch for compatible fixes, minor for compatible features,
  and major for incompatible changes. Agree the release number with the user.
- `VERSION` overrides and binary output omit `v`; Git tags include it, such as
  `v0.6.1`. Release scripts require `CI_COMMIT_TAG` to match the recorded version.
- Keep changes under Unreleased until cutting a dated CHANGELOG.md section.
- Update current-version docs, website copy, download links and handoff records.
  Preserve historical changelog versions, old checksums, old download directories
  and protocol numbers. A global version-string replacement is not safe.
- Published tags and version directories are immutable. If published bytes need
  fixing, use a new version rather than replacing the old artifacts.

Review and commit the intended source. Record the exact commit, toolchain and
build commands. Tag that reviewed source, not a later documentation-only commit.
Push the branch and tag only when authorized. A tag push or local build is not
a public deployment: CI's publish stage only prepares artifacts.

## 2. Test and build off the production host

Use Go 1.26 or newer. On Sean's Mac, from the framework repository:

```sh
GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go test ./...
GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go vet ./...
GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go test -race ./...
GOCACHE=/tmp/flea-go-cache GO_BINARY=/usr/local/go/bin/go bash scripts/build.sh
./bin/flea version
```

Race checks are required for concurrency, reload and shared-state changes.
Run applicable browser/manual checks too. Do not overlap the bundled build with
tests compiling bundle assets. Socket tests may need sandbox bind permission.
`scripts/build.sh` rebuilds both embedded Linux servers before the local CLI;
plain `go build` alone can embed stale servers.

Use a fresh temporary site for `new` and `status`, compare affected starter files,
and exercise changed behavior using the exact final artifacts. Stop test servers.
Confirm the installed CLI path/version separately from `bin/flea`.

The preferred release matrix uses native runners through `scripts/release.sh`:
Linux AMD64/ARM64 server builds, then Mac ARM64 and Linux AMD64 CLI builds embedding
both servers. See releasing.md for prerequisites. Run
`CI_COMMIT_TAG=v<version> bash scripts/stage-release.sh` only with verified
`dist/flea-darwin-arm64` and `dist/flea-linux-amd64` artifacts. It stages files;
it does not upload them. Use a fresh staging directory rather than overwriting one.

For 0.6.1, Linux AMD64 was cross-compiled on the Mac after refreshing bundles:

```sh
GOCACHE=/tmp/flea-go-cache CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  /usr/local/go/bin/go build -trimpath \
  -ldflags '-s -w -X flea/internal/buildinfo.Version=0.6.1' \
  -o /tmp/flea-061-linux-amd64 ./cmd/flea
```

That exact binary passed native unprivileged Linux smoke checks. This was not a
native Linux build or full Linux suite. Record such deviations and remaining
acceptance work explicitly; do not silently treat cross-compilation as native QA.
Mac ad-hoc signing is not notarization or clean-machine acceptance.

Do not compile on the 1 GB production host as the normal workflow. It needs only
the CLI's bundled server, not Go or source. If a separately authorized native build
must run on a constrained host, `GOMAXPROCS=1 GOFLAGS=-p=1` reduces build concurrency;
put temporary build/cache files on verified disk-backed storage and monitor RAM.

## 3. Preflight the low-memory host

Run read-only checks before any operation:

```sh
ssh root@fleacms.com 'systemctl is-active flea; /usr/local/bin/flea-server -version; free -m; swapon --show; df -h / /tmp /var/tmp; systemctl show flea -p Environment -p MemoryHigh -p MemoryMax'
ssh root@fleacms.com 'ls -l /etc/flea/*pending* 2>/dev/null; ps -eo pid,rss,args | grep -E "[f]lea-server|[f]lea-upgrade"'
```

An absent journal glob may make the second command return nonzero. Inspect the
actual output. Never print server configuration, credentials or session files.
Take an authorized `flea backup` to a new private path outside the website before
risky changes. It is a site backup, not a whole-host backup.

### Safeguards approved for this host

The 0.6.1 host had about 956 MiB usable RAM. Loading retained release downloads
and running an admin process alongside the service exhausted it. The service
alone being healthy did not establish enough headroom for an upgrade.

The existing `/etc/systemd/system/flea.service.d/60-memory.conf` contains:

```ini
[Service]
Environment=GOMEMLIMIT=320MiB
```

This is a soft Go runtime memory target, not an RSS ceiling. It does not eliminate
the memory needed for live data. A hard cap can kill the service; do not introduce
one as a substitute. Preserve existing systemd hardening and unrelated overrides.

Sean also approved a 1 GB `/swapfile`, root-owned, mode 0600. It is enabled with
`swapon` and persisted through this entry in `/etc/fstab`:

```text
/swapfile none swap sw 0 0
```

The pre-change fstab was retained at `/etc/fstab.before-flea-swap-20261002`.
Check that swap is active, not merely present on disk. Do not recreate, truncate
or run `mkswap` on an existing swap file. Repairing or resizing swap, or changing
service policy, needs explicit approval and filesystem checks. Do not duplicate
fstab entries. The existing buffer survived the 0.6.1 upgrade and deployment;
it does not guarantee enough capacity for indefinitely growing release assets.

`/tmp` is RAM-backed on this host. Old build caches and uploaded archives consume
the same scarce memory. Inspect exact paths and open-file users before any cleanup.
During earlier releases, completed, unopened upload archives were moved intact
to `/var/tmp` after extraction. That was an exceptional recovery measure, not a
routine deploy step. Never move an active upload or prune retained releases,
content or media. Prefer disk-backed staging in a future implementation.

### Apply the target to admin processes too

The service override does not affect the separate SSH `-deploy`, `-upgrade` or
`-recover-upgrade` processes. Setting GOMEMLIMIT only on the local CLI does not
forward it to the remote command.

For 0.6.1, a temporary local SSH wrapper supplied the remote environment while
preserving all original SSH arguments. To reproduce it, create a new private
directory with `mktemp -d /tmp/flea-memory-ssh.XXXXXX`, save the following as its
`ssh` file, and make it executable. Do not assume an old /tmp wrapper still exists.
Review against current CLI command construction if that code has changed.

```bash
#!/bin/bash
args=("$@")
last=$((${#args[@]} - 1))
if [[ $last -ge 0 ]] && {
  [[ "${args[$last]}" == *flea-server* && "${args[$last]}" == *" -deploy "* ]] ||
  [[ "${args[$last]}" == *flea-upgrade.* && (
    "${args[$last]}" == *" -upgrade "* ||
    "${args[$last]}" == *" -recover-upgrade"*
  ) ]]
}; then
  args[$last]="env GOMEMLIMIT=320MiB ${args[$last]}"
fi
exec /usr/bin/ssh "${args[@]}"
```

Use the chosen directory as `task_ssh_dir`, scoped to each invocation:

```sh
PATH="$task_ssh_dir:$PATH" /Users/sean/Projects/flea/bin/flea update-server
PATH="$task_ssh_dir:$PATH" /Users/sean/Projects/flea/bin/flea deploy
```

Run these from the website repository, sequentially, not concurrently. This
wrapper is an operator workaround for the confirmed root destination, not a
general sudo/SSH configuration feature. Do not disable host-key checks, change
authentication, or globally replace SSH. Do not rely on it for unmatched commands.

## 4. Publish downloads, then the website

From 0.7.0 onward, follow [download hosting](download-hosting.md): stage artifacts
outside the website, upload and validate a new immutable version directory in
`/var/lib/flea/downloads/fleacms`, restart to load the catalogue, and verify public
hashes before deploying the small website/channel update. Existing versions stay
in that store and out of site snapshots. Do not copy dist/public/releases version
directories into website Git. The store is backed up separately from site backups.
Upgrade the server before enabling the host field on the first migration; keep
the old channel until all checks pass. Preserve approved memory and swap settings.

### Historical procedure through 0.6.2 (not the current workflow)

1. Put only the new immutable version directory in the website:
   `site/releases/<version>/flea-darwin-arm64`, `flea-linux-amd64`, `SHA256SUMS`.
   Hash the exact final bytes after any signing. Preserve all old directories.
2. Keep `latest.txt` and current public install links on the previous release.
   Commit the artifact-only change with permission, then deploy using the wrapper.
   If future website edits already exist, use a clean, reviewed staging checkout
   with the correct branch mapping. Never reset the user's work to stage a release.
3. Download both binaries and SHA256SUMS anonymously over HTTPS into a fresh
   temporary directory. Verify hashes (`shasum -a 256 -c SHA256SUMS` on Mac or
   `sha256sum -c SHA256SUMS` on Linux), version and native execution. Do not just
   compare local staging files. Validate the Mac signature separately.
4. Upgrade the shared server using the new CLI's `flea update-server`, with the
   wrapper and confirmation. It validates and journals the executable swap.
   Verify the remote version, service health and every affected site.
5. Update `site/releases/latest.txt`, homepage/docs download links and release
   notes to the new version. Check installer references and social metadata/assets.
   Commit the website changes, then run the final normal `flea deploy` with the
   wrapper. Normal deployment preserves live snips and theme. Do not use explicit
   sync merely to publish binaries or docs.
6. Record both activation IDs and source commits. Push authorized source and
   website commits. A deploy does not push Git, and a Git push does not deploy.

Monitor `free -m`, `swapon --show`, process RSS and `journalctl -u flea` while
activating. Keep the user informed. Never overlap deploy, upgrade or recovery.

## 5. If activation or upgrade fails

Stop new release attempts. Check service state, exact running processes, journal
presence and kernel/service logs for OOM. A timeout is not proof of an OOM; obtain
evidence. Do not keep retrying while the same memory pressure remains.

- Allow automatic rollback to finish. It may need separate verification time.
- For a pending executable upgrade, after the previous operation exits, run
  `flea update-server --recover` from the website using the wrapper.
- For interrupted site activation, use the documented server-side recovery,
  after confirming no other admin operation is running:
  `ssh root@fleacms.com 'env GOMEMLIMIT=320MiB /usr/local/bin/flea-server -recover'`.
  Do not confuse site activation recovery with executable upgrade recovery.
- Never delete recovery journals or locks to unblock deployment, manually swap
  executables, reset the host, or overwrite live configuration/content.
- If a recovery verifier itself starves the service, investigate its exact PID
  and phase. In the 0.6.1 incident, the old executable was already restored;
  terminating that specific verifier released RAM and restored availability but
  left recovery pending. This is not a routine shortcut. Preserve the journal,
  resolve capacity and rerun supported recovery before another deployment.
- Confirm HTTPS availability and cleared recovery state before claiming success.
  If new host changes or more capacity are required, explain the blocker and ask.

## 6. Verify the public result and hand off

Check the public homepage, docs, `latest.txt`, both download URLs, checksum manifest,
installer and social-preview image. Check Bookworm after shared-service restarts.
Verify both public HTTPS status and relevant body contents, not status alone.

On Mac ARM64 and native Linux AMD64:

1. Download the public installer and install into a fresh absolute directory using
   `FLEA_INSTALL_DIR`. Keep this separate from the user's normal installation.
2. Verify installed version and hash against the public manifest.
3. Put the previous public CLI in a second isolated directory and run its
   `flea update`. Verify the resulting version/hash, plus already-current behavior.
4. Run site creation/status and relevant feature smoke tests. Run Linux checks
   unprivileged, not as root; stop any test dev servers afterward.
5. Update the user's normal CLI only when authorized, then verify the resolved
   binary path and `flea version`.

Update the website's RELEASES.md and handoff.md with provenance, activation IDs,
exact checks, recovery events and remaining limits. Update current status in the
framework's docs/installation.md and docs/agents/next-steps.md. Do not rewrite historical
release records. Documentation-only follow-up commits excluded from the deployed
snapshot do not require another site deployment.

Run `git diff --check`, review and commit only authorized files, push both repos,
and confirm their intended branch matches the remote. Preserve unrelated changes.
Finish with website `flea status`, live server version/health, and any remaining
acceptance limits. Do not turn a successful smoke check into a claim of full QA.

## Private server rollout: 0.9.0-rc.1, October 5, 2026

Framework `3447cb7`; standard bundled build, Mac CLI SHA-256
`505642a47e0c558b4bfb3a4737bb4a408894768d2c836391784f53397e437fe6`.
Shared server is 0.9.0-rc.1 / protocol 23. Public downloads and the installed
CLI remain 0.8.1; no RC tag or public artifacts were published.

Atlas interim `e9dd946`, activation `20261005-162654-c4165eb23f351ba4f70e2c8c`,
temporarily disabled Contact while retaining its routing and live content.
Bookworm (`v1-acceptance-main`) was then unregistered through explicitly approved
one-time protocol-22 maintenance, with all 20 releases and other files retained.
Config backup: `/var/backups/flea/server.json.before-bookworm-removal-20261005-162928`.
Five verified site backups remain locally in `/Users/sean/fleacms-backups`.

Supported `update-server` succeeded, followed by Atlas `b2f768a`, activation
`20261005-163105-49d3e6e8f1cdb06b0ee73e86`, and website `7ef4e6e`, activation
`20261005-163126-76b53507e24be701230d2e89`. Kent and Naves were not redeployed.
All four sites are healthy. Themes and content match backups except Atlas's
intentional thank-you route move, whose snips were preserved. Both Contact forms
passed CSRF rejection, validation feedback and honeypot success redirect checks;
no real email was sent. Two Contact children run under the existing service policy.
No rollback was required and no recovery journals remain. Memory/swap safeguards
are unchanged. Live inventory passed; the new removal command still needs live QA.

The one-time script is not a supported reusable tool: its review mode used BSD
`mktemp -t` syntax and needed a transient shell adapter on Linux. Do not reuse it
on protocol 23. Use the new site commands for future approved administration.
Native Linux full-suite and clean-Mac acceptance remain open.

## Historical completed release: 0.8.0

Framework `c8a1d96`, tag `v0.8.0`; website `3188000`, activation
`20261003-171440-ca76afb853e32d0d0309c6ea`. Public Mac/Linux hashes, fresh
installs and updates from 0.7.1 passed; server, channel and local CLI are 0.8.0.
Naves and Kent received toolbar-only deployments. Existing sites require the
new starter `site/admin/_toolbar.html` edit-row wrapper with this runtime;
update those partials before upgrading the shared server. Preserve custom layouts.
Bookworm was explicitly excluded by Sean: public pages remain healthy, but its
old toolbar is incompatible with the 0.8.0 editor. No compatibility fallback was
added. Native Linux full-suite/build and clean Mac/notarization remain open.
Memory/swap settings are unchanged, and recovery is clear. See website RELEASES.md.

## Historical completed release: 0.7.0

Framework `1199747`, tag `v0.7.0`; website `a41a758`, activation
`20261003-141847-a863aaed91fd5f4c8fa761b8`. All seven version directories were
copied to the separate store and every public manifest/binary hash verified.
The website then removed its duplicate version directories and deployed latest.txt
and docs as 0.7.0. Local binaries are preserved in /Users/sean/fleacms-releases.
Upload size was 575 KB. Tests, race, vet and Mac/Linux public installer/update
checks passed. Native Linux full build/suite and clean Mac/notarization remain open.
See the website RELEASES.md and docs/agents/download-hosting.md for evidence and procedure.

## Historical completed release: 0.6.1

- Tagged framework source: `89e3472`, tag `v0.6.1`.
- Deployed website source: `1857a45`.
- Final website activation: `20261002-015019-3e6152722de7e84b956d7054`.
- Final documentation commits: framework `f33db2e`, website `90ab9ac`.
- Public fresh installs and 0.6.0-to-0.6.1 updates passed on Mac ARM64 and native
  unprivileged Linux AMD64; installed hashes matched the public manifest.
- Server and local CLI reported 0.6.1. Both sites and social image returned 200.
- Full native Linux build/test coverage, clean-machine Mac acceptance and
  notarization remained unverified. This was a test release, not completed V1 QA.

These are historical results. Recheck live state instead of assuming it is unchanged.

## Handoff notes for the next agent

### Start without changing anything

1. Read the current AGENTS.md and inspect both repositories' status, branch,
   latest commits and remotes. The website has its own AGENTS.md. Do not assume
   this document's historical commit is still HEAD.
2. Check `command -v flea`, `flea version`, `bin/flea version`, the recorded source
   version, public latest.txt and remote server version separately. These are
   independently updated copies and may legitimately differ during a release.
3. Check pending journals, active admin processes, service status, RAM and swap
   before initiating SSH operations that load sites. After an interrupted agent
   turn, local tool sessions and remote processes may still be running.
4. Establish the user's current request. A previous release authorization does
   not authorize a new release, host migration, cleanup or unrelated repair.

### Work intentionally left open

- LAN terminal QR codes are a planning item in docs/agents/next-steps.md, not implemented.
  Automatic display with `--lan` versus an explicit `--qr` remains undecided.
- Versioned download asset memory was addressed in 0.7.0 by the separate store.
  For other large site assets, consider disk-backed upload staging. Investigate
  the current loader and immutable snapshot requirements before designing a fix.
  Do not delete old public binaries to make the problem disappear. Test memory
  under simultaneous service startup and admin validation, not just steady state.
- Consider a supported way to configure admin-process memory targets so operators
  no longer need the SSH wrapper. This is a proposal, not an implemented option.
- Finish the native Linux/full-suite and clean-machine Mac/notarization acceptance
  work before describing distribution as fully verified. Consult docs/agents/checklist.md;
  do not replace the user's manual answers with inferred passes.

### Local Linux build-machine investigation

Sean identified `buzz` at `192.168.4.31` as his Linux ARM Mac with 24 GB RAM.
Discovery found SSH there and the saved host key matched. However, `ssh root@buzz`
failed for both Sean and the agent. The default `id_ed25519` and explicit
`id_rsa_2` attempts failed; the agent's SSH agent had no identities loaded.
Architecture, RAM and tools were not verified through login. No source was copied
and no software was installed. Sean chose to leave this aside and finish the
production release. Do not assume Buzz is an available runner or resume SSH
configuration changes without a request. Its IP may change.

### Common mistakes to avoid

- Restart dev after rebuilding Go code. File reload does not reload the binary.
  The 0.6.1 same-site replacement feature cannot make an already-running older
  dev binary understand its control protocol; that old process may need one
  manual stop. Never interrupt pull/reset to bypass the site lock.
- Keep the deployment protocol version separate from SemVer. Do not bump it
  merely to make a version search look consistent.
- Compare the published Mac executable's checksum after signing, not before.
  Verify signatures with `codesign --verify --verbose=2 <downloaded-binary>`;
  a passing ad-hoc signature check does not establish notarization.
- An HTTPS 200 can be an error page or wrong revision. Inspect the version/link
  content and activation receipt as well. Social-card HTTP availability does
  not prove Slack or Messages refreshed an already cached preview.
- Temporary paths in old logs are evidence, not reusable infrastructure. Create
  fresh private test directories and verify ownership before unprivileged tests.
- Never paste full configuration files or verbose secret-bearing environment
  output into release notes. Record versions, hashes, paths and outcomes only.

### What to leave in the next handoff

Record the requested version, source commit/tag, artifact hashes and build
provenance; native checks versus cross-compilation; both website activation IDs;
public channel, server and local CLI versions; host changes and recovery state;
remaining tests; and any running processes or uncommitted work. Explicitly say
what was not deployed or verified. Stop temporary previews unless asked otherwise.

At the time this runbook was written, the runbook and its guide links were not
committed. The separate docs/agents/next-steps.md QR-code planning edit was also uncommitted.
Preserve those changes and ask before committing if no current permission exists.
