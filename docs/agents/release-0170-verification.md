# FleaCMS 0.17.0 release verification

Published October 18, 2026. Public release: built, tagged, published, installed
and deployed. This record separates the earlier local candidate checks from the
canonical build, native acceptance, public installer matrix and live rollout.

## Scope

0.17.0 adds custom managed apps, app-owned sessions and the deployment protocol
26 package contract. It also carries the modular agent skills already on main.
The app wire protocol is 2. Deployment protocol moves from 25 to 26: a
protocol-25 server refuses a package-bearing release, and a protocol-26 CLI
refuses a protocol-25 server before upload. There is no fallback or migration.

The initial preparation task was local and documentation-only on top of the
already reviewed integration: it changed the recorded version and release
documentation, and added no runtime code, Impeccable code, design, dependency or
site layout change. Later phases added the canonical build, native acceptance,
publication and rollout recorded below.

## Provenance

- Base: `main` `f09d877` (clean working tree at task start).
- Published source: `main` `fe15b8934165a90390c8bd4ff88cc5e9c9e0d753` (clean
  tree), immutable tag `v0.17.0`.
- Version: `internal/buildinfo/version.go` = `0.17.0`.
- Deployment protocol: `internal/release/protocol.go` = `FLEA_DEPLOY=26`.
- App protocol: `2`.
- Preserved pre-reconcile source: commit `719c504` on branch
  `archive/custom-apps-before-reconcile-20261018`, plus the private checkout and
  manifests under `/Users/sean/flea-workspace-backups/2026-10-18-before-reconcile`.
- Integrated custom apps were independently reviewed before this task; see
  [custom apps integration](custom-apps-integration.md).

## Local source checks (candidate)

Run from `/Users/sean/Projects/flea` with Go 1.27.1 and real development bundles
present (`FLEA_REQUIRE_APP_BUNDLE=1`, `FLEA_REQUIRE_BUNDLE=1`):

| Check | Command | Result |
| --- | --- | --- |
| Full tests | `GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go test ./...` | Pass, exit 0 |
| Vet | `GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go vet ./...` | Pass, exit 0 |
| Race | `GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go test -race ./...` | Pass, exit 0 |

Go 1.27.1 on darwin/arm64. The full-suite commands exited 0, and the
bundle-required Contact/managed-app checks ran against the real development
bundles. The default source runs did not set `FLEA_TEST_BROWSER`, so the Chrome
tests below were skipped in those runs and are recorded separately. Private
logs: `/tmp/deepseek-task.GFVMwd/logs/` (`test.txt`, `vet.txt`, `race.txt` and
their `.exit` files).

### Optional browser checks

Run once against the current source (Go 1.27.1, darwin/arm64) with the
documented Chrome workflow and private headless profiles only:

```sh
FLEA_TEST_BROWSER='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' \
  GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go test \
  ./internal/editor ./internal/site -run Browser -count=1 -v
```

Exit 0. `internal/editor`: `TestExternalEditorBrowser`,
`TestCodeInputBrowser`, `TestPhotosBrowser`, `TestMultisnipsBrowser`,
`TestOwnerNavigationBrowser`, `TestPublicPreviewBrowser`,
`TestOwnerReloadBrowser`, `TestVideosBrowser`. `internal/site`:
`TestBrowserCannotInjectAppMetadata`, `TestNavigationStylesBrowser`,
`TestDevErrorBrowser`. All pass; no runtime code was changed. Evidence:
`evidence/browser-checks/` (`browser-test.log`, `browser-test.exit`).

## Canonical build and native checks

Local archive `/Users/sean/fleacms-releases/0.17.0`; private evidence under
`evidence/`. Recorded against the `fe15b89` source:

- macOS ARM64 canonical `scripts/build.sh` (Go 1.27.1, `CGO_ENABLED=0`), the two
  bundle-required tests, Mac CLI VCS metadata and ad-hoc codesign verification:
  pass.
- Three CLIs (Mac ARM64; Linux AMD64 cross-built on the Mac; Linux ARM64 built
  natively from an exact `git archive`) and both bundled Linux servers;
  `stage-release.sh CI_COMMIT_TAG=v0.17.0`, `SHA256SUMS` and archive hashes
  verified.
- Native Linux ARM64 `go test ./...`, `vet` and `-race` exit 0 with real bundles,
  plus a native lifecycle smoke with the managed session app.
- The exact published Linux AMD64 server passed an isolated native systemd
  acceptance (TLS, app and CSRF, login/session reuse and rotation, restart
  persistence, immutability, 403 negatives) and the exact published Linux AMD64
  CLI passed native `new`, default-template parity, `status`, raw JSON attribute,
  development reload and shutdown checks as an unprivileged user. Linux AMD64
  was cross-built on the Mac; there was no native AMD64 build or full suite.
- Independent test input `evidence/session-linux-amd64`
  (`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ./samples/session-app`, sha256
  `e124952f9e48a42410f13c3e83d24cf43bda7dddaf9e157aaf23e40249a06af7`); not a
  published release binary and not part of any CLI.
- The Mac binary is ad-hoc signed and verified on disk; it is not notarized.

## Published artifacts

Final artifacts match the public immutable directory
`https://fleacms.com/releases/0.17.0/` for all three hashes:

```text
5f9cb8d9bda64d6f00f52eb43e1d38b82cb9d28832f30df5e3b4bbc71973dec5  flea-darwin-arm64
919ed61f7ac31f5ba3163a16b2bce05d1fcd7fcbce8c82c2a667c15a1f16cad4  flea-linux-amd64
8c0ac734abf2dcbb2e9aac6e808a74c83040c21d067c2a3c1906c7fe0b96b09d  flea-linux-arm64
```

Anonymous HTTPS downloads pass, the older published `0.16.1` manifest still
resolves, and published bytes were not overwritten.

## Public installer and update matrix

The downloaded public `install.sh` is byte-identical to the canonical
`dist/public/install.sh` and `website/site/install.sh`
(`a15fda6a2d3ebabe776b81f088938c82ff402f51ac9e37c6444eb59ee0dcbeed`).

All checks below passed on macOS ARM64 (local), native Asahi Linux ARM64 and
native Linux AMD64 as `nobody` in a disposable isolated directory with an
explicit absolute `FLEA_INSTALL_DIR`:

- fresh install reported `FleaCMS 0.17.0` and the installed hash matched the
  public manifest;
- reinstall of the same version kept the version and hash;
- the public `0.16.1` CLI was installed in a second isolated directory, matched
  its published checksum, then its own `update` reached 0.17.0 with a matching
  hash;
- a further `update` reported already current and changed nothing;
- the fresh CLI ran `new --skip-git` and `status` once on each native platform;
- macOS ad-hoc `codesign --verify` passed.

Per-platform hashes matched `5f9cb8d9…dec5` (Mac), `8c0ac734…b09d` (Asahi ARM64)
and `919ed61f…cad4` (AMD64). Evidence and redacted logs:
`evidence/public-installers/`.

## Shared server and website rollout

- Supported `flea update-server` completed as 0.17.0 / `FLEA_DEPLOY=26`. The
  production unit is active with protocol 26, no pending recovery journal, and
  four matching bundled Contact children running.
- All 18 configured HTTPS hosts returned 200. Routing, all editable content and
  all deployed source hashes were unchanged across the executable upgrade.
- The existing 320 MiB `GOMEMLIMIT` override and swap are unchanged.
- Website source `115a73c`, activation
  `20261018-195731-803d4f4dc7db80fad2a5fd8a`; `latest.txt` selects 0.17.0.
- Final website check: global revision matched the expected activation ID, all 18
  content SHA values were unchanged even after the website deploy, and non-Flea
  website sources, routing, commits, TLS, downloads, site branches and redirect
  policy were unchanged. Public docs, custom-app availability and every new
  download link were checked in a browser.
- Normal Mac `/Users/sean/.local/bin/flea` and Asahi
  `/home/sean/.local/bin/flea` were updated through the public updater and report
  0.17.0 with matching hashes (`5f9cb8d9…dec5` and `8c0ac734…b09d`). Previous
  binaries were retained privately.
- Two Asahi previews were gracefully replaced preserving cwd and flags:
  `codepilot-website-ink-html` `dev --lan -port 3207` (PID 3053469) and
  `codepilot-website` `dev --lan --port 3001` (PID 3053476). Their
  `/proc/<pid>/exe` hashes match `8c0ac734…b09d`, the old PIDs are gone, both
  return HTTP 200, and config/index/snips source hashes are unchanged. The Asahi
  dirty framework checkout was left untouched. No Mac dev process was running.

## Custom app acceptance

The managed session sample app was exercised under a private transient TLS
namespace with production hardening using the exact release artifact, including
launch, handshake, session across a restart and graceful shutdown. The full
Go suites separately cover the A-to-B release lifecycle and Blog-save boundary. This is
a controlled acceptance run using the sample app; it is not a claim that a
customer custom app is deployed in production.

## Post-deploy receipts and cleanup

- `flea status` from the installed CLI passed all 8 known Mac site folders,
  including remote comparisons, after retrying earlier concurrent-read lock
  contention.
- Website status source `115a73c` matched code and content before the bookkeeping
  edits.
- A full post-deploy check returned 200 again for all 18 HTTPS hosts, and a
  missing-CSRF production Contact POST returned 403 with no email sent.
- Cleanup receipt: the lead removed only completed own fixture directories
  `/var/lib/flea-accept-0170`, `/var/lib/flea-accept-0170-attempt1`,
  `/var/tmp/flea-0170-native-check` and `/var/tmp/flea-0170-cli-smoke`, plus the
  empty publication staging directory, after verifying no active own units or
  processes and copying redacted evidence. All real runtime, snapshots, download
  versions and the previous server binary were retained. No native fixture work
  is pending. The completed isolated public-installer directories on AMD64
  and Asahi were also removed after their logs were archived.
- Asahi intentionally retained its native build stage, the private previous CLI
  and the live previews. The current CodePilot Mac website4 checkout is clean
  after independent favicon/guidance commits `42ad6d0` by other work; it was not
  edited or deployed here, so only the Asahi preview's tracked files and live
  remote content are claimed unchanged, not the Mac website4 source hash.

## Remaining limits

Do not claim any of the following until separately recorded:

- Native AMD64 full-suite/build (the CLI/server were cross-built and then tested
  natively, but the full suite and a native build were not run on AMD64).
- Clean-Mac installation acceptance and Apple notarization.
- Real email delivery acceptance.

## Rollout record (completed)

1. Committed the candidate and built canonically from that exact commit.
2. Ran native checks, including the isolated custom-app systemd acceptance,
   and staged the immutable `releases/0.17.0/` directory.
3. Published the version directory under the deployment lock and validated it.
4. Upgraded the shared server to 0.17.0 / protocol 26, preserving the 320 MiB
   target and 1 GB swap, then verified anonymous download hashes.
5. Deployed the website to advance `latest.txt` and the current documentation.
6. Passed the public installer/update matrix, updated the installed CLIs and
   gracefully replaced the two existing Asahi previews.
7. Verified hosted-site health, content/source preservation, recovery state and
   optional Chrome checks, then recorded the receipts and remaining limits.

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

- No release blocker remains for 0.17.0.
- Protocol 26 requires a matching server; 0.16.1 / protocol 25 cannot deploy a
  custom app package.
- Native Linux AMD64 full-suite/build, clean-Mac, notarization and real email
  acceptance remain unverified.
- The 1 GB host still depends on the 320 MiB memory target and swap under
  deployment peaks.

## Evidence

Private release evidence lives under `/Users/sean/fleacms-releases/0.17.0/`
(`evidence/mac`, `evidence/native`, `evidence/native-amd64-*`,
`evidence/public-installers` and host snapshots). Source-check logs also remain in
the runner task directory (`/tmp/deepseek-task.GFVMwd/logs/`). Evidence is not
part of the public release artifact.
