# FleaCMS 0.17.0 release verification

Candidate preparation: October 18, 2026. Public release candidate; **not built,
tagged, published, installed or deployed**. This record separates the local
candidate checks from the later canonical build and native acceptance work.

## Scope

0.17.0 adds custom managed apps, app-owned sessions and the deployment protocol
26 package contract. It also carries the modular agent skills already on main.
The app wire protocol is 2. Deployment protocol moves from 25 to 26: a
protocol-25 server refuses a package-bearing release, and a protocol-26 CLI
refuses a protocol-25 server before upload. There is no fallback or migration.

This preparation task is local and documentation-only on top of the already
reviewed integration: it changes the recorded version and release documentation,
and adds no runtime code, Impeccable code, design, dependency or site layout
change.

## Provenance

- Base: `main` `f09d877` (clean working tree at task start).
- Candidate source commit/tag: **TBD by the lead**; the lead commits the candidate
  before the canonical build so binaries carry correct VCS metadata.
- Version: `internal/buildinfo/version.go` = `0.17.0`.
- Deployment protocol: `internal/release/protocol.go` = `FLEA_DEPLOY=26`.
- App protocol: `2`.
- Preserved pre-reconcile source: commit `719c504` on branch
  `archive/custom-apps-before-reconcile-20261018`, plus the private checkout and
  manifests under `/Users/sean/flea-workspace-backups/2026-10-18-before-reconcile`.
- Integrated custom apps were independently reviewed before this task; see
  [custom apps integration](custom-apps-integration.md).

## Local source checks (this candidate)

Run from `/Users/sean/Projects/flea` with Go 1.27.1 and real development bundles
present (`FLEA_REQUIRE_APP_BUNDLE=1`, `FLEA_REQUIRE_BUNDLE=1`):

| Check | Command | Result |
| --- | --- | --- |
| Full tests | `GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go test ./...` | Pass, exit 0 |
| Vet | `GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go vet ./...` | Pass, exit 0 |
| Race | `GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go test -race ./...` | Pass, exit 0 |

Go 1.27.1 on darwin/arm64. No test was skipped and the bundle-required
Contact/managed-app checks ran against the real development bundles. Private
logs: `/tmp/deepseek-task.GFVMwd/logs/` (`test.txt`, `vet.txt`, `race.txt` and
their `.exit` files).

No build is recorded here. `scripts/build.sh`, release artifacts, hashes,
signatures and native checks belong to the next phase.

## Not yet established

Do not claim any of the following until the corresponding phase is recorded:

- Canonical bundled build (`scripts/build.sh` / `scripts/release.sh`) on the
  candidate commit.
- macOS ARM64, Linux AMD64 and Linux ARM64 artifacts, SHA256SUMS, installer and
  `latest.txt` staging.
- Public immutable download upload and anonymous HTTPS verification.
- Native platform install, reinstall, already-current and update checks.
- Linux/server native checks and the live protocol 25 to 26 server upgrade.
- Live custom app acceptance under the real systemd unit.
- Real email, clean-Mac and notarization acceptance.

## Rollout plan (later phase, requires separate authorization)

1. Commit the candidate, then build canonically from that exact commit so VCS
   metadata is correct. Do not reuse the development `bin/flea` for release.
2. Run the native CLI checks for macOS ARM64, Linux AMD64 and Linux ARM64, then
   stage the immutable `releases/0.17.0/` directory and `latest.txt`.
3. Upload the version directory first and verify anonymous HTTPS downloads and
   hashes. Advance `latest.txt` and the website only after that check.
4. Upgrade the shared server to 0.17.0 / protocol 26 with `flea update-server`
   before deploying a package-bearing site. Preserve the existing 320 MiB
   `GOMEMLIMIT` service override and the 1 GB swap file; apply the memory target
   to separate SSH deploy/upgrade processes too. Never compile on the production
   host and never overlap deploy, upgrade or recovery.
5. Run the custom app lifecycle (launch, handshake, session across a release
   change, graceful shutdown) under the real systemd unit on the target host.
6. Verify every hosted site after the restart: HTTPS, routing, retained content,
   media, CSS and recovery journals. Existing sites keep protocol-25 releases
   until redeployed.
7. Update the website, `RELEASES.md` and framework status docs with actual
   activation IDs and remaining limits.

## Recovery plan

- Failed executable upgrade: let rollback finish, then run
  `flea update-server --recover`; never delete journals or swap executables by
  hand.
- Interrupted site activation: run
  `ssh root@fleacms.com 'env GOMEMLIMIT=320MiB /usr/local/bin/flea-server -recover'`
  after confirming no other admin operation is running.
- Custom app failure: one app failing should affect only its own routes; other
  pages stay up. Check the app's stable runtime storage before restarting.
- Keep immutable downloads: publish a new version instead of replacing 0.17.0
  bytes. Old versions and hashes remain available.
- Preserve existing content, media, CSS, certificates, session state and custom
  app storage. Do not prune retained releases or media automatically.

## Blockers and open gates

- The lead must commit the candidate before the canonical build.
- No release binaries, hashes or tags exist yet.
- Protocol 26 requires a matching server; 0.16.1 / protocol 25 cannot deploy a
  custom app package.
- Live systemd, native Linux AMD64 full-suite, clean-Mac, notarization and real
  email acceptance remain unverified.
- The 1 GB host still depends on the 320 MiB memory target and swap under
  deployment peaks.

## Evidence

Detailed private logs live under the runner task directory
(`/tmp/deepseek-task.GFVMwd/logs/`). They are not part of the release artifact.
