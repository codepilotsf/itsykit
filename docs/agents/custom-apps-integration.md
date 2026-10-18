# Custom apps integration handoff

Reconciliation of the pre-existing custom managed-app and app-owned-session work
onto current main, together with the already-released literal-brace fix.

## Provenance

- Archive (all original source bytes preserved): commit `719c504` on branch
  `archive/custom-apps-before-reconcile-20261018`, parent `f9db8ad` (old 0.10.0
  base). Full private checkout backup and source manifest:
  `/Users/sean/flea-workspace-backups/2026-10-18-before-reconcile`.
- Integration base: current `origin/main` `b32b74e` (newer Blog, Theme removal,
  LAN hosts/ports, modular `.agents` skills).
- Integration branch: `integrate/custom-apps-20261018`.
- Literal-brace fix: source `57d8378` on `release/0.16.1`, receipt commits
  `f438622` and `f54a7ff`. Its compiler and regression coverage are identical to
  the archive copy of `internal/site/template_lex.go` and
  `internal/site/template_lex_test.go`; the archive `template.go` differs only in
  that the released tree also carries the Blog body token, which the integration
  keeps from current main.
- Historical design note: the archive `docs/agents/custom-apps-handoff.md` is not
  carried forward as active guidance. This file replaces it.

## Version and protocol boundaries

- `internal/buildinfo/version.go` is `0.17.0-dev`. Feature changes stay under
  `## Unreleased`. The public and installed version remains 0.16.1 and the
  published 0.16.1 server remains `FLEA_DEPLOY=25`.
- Development deployment protocol is `FLEA_DEPLOY=26`
  (`internal/release/protocol.go`). Protocol 26 adds the custom-app package
  contract (confined `flea/apps/<name>/` manifests and per-platform native
  executables extracted 0555). A protocol-25 server is refused before upload and
  cannot extract a package; there is no fallback.
- App wire protocol is `2` (`internal/appapi`). v1 services are rejected.
- No deployment, tag, publish or installed-binary replacement is part of this
  integration.

## What was integrated

Custom managed apps: strict shared manifest/package policy in
`internal/sitefiles` (`apppackage.go`, `executable.go`, `hardlink_unix.go`),
load-time confinement and identity recording (`internal/site/app_package.go`),
frozen-executable launch and private storage (`internal/appmanager`), release
packaging/extraction and preflight (`internal/release/archive.go`,
`internal/setup`), stable per-site runtime root (`ManagedRuntimeRoot`), and the
standalone `samples/session-app` harness.

App-owned sessions: `internal/site/session.go` (per-site/per-app cookie, raw
Cookie-header duplicate handling, `X-Flea-Session` bridge) and the
`internal/appapi` envelope `session` validation (`token`/`max_age`, clear,
omitted). Contact keeps fixed read-only routes and the mount-only POST; custom
apps may read on GET.

Blog and Blog contract are preserved unchanged. Theme support is not
reintroduced. Modular skills and the `.agents` workflow are unchanged.

The full disposition inventory (integrated / superseded by current code /
archived historical guidance, per original change) lives outside the repository
at `/Users/sean/flea-workspace-backups/2026-10-18-before-reconcile/integration-disposition.md`.

## Verification completed

The final tree was checked before handoff. Commands and evidence logs live under
`/tmp/deepseek-task.CwLX3m/`:

- `go vet ./...` exit 0 (`logs/vet-2.txt`; the exit is recorded in `worker.log`).
- `go test ./...` exit 0 (`logs/test-3.txt`), after the stable-version fixture
  adaptation to `internal/buildinfo/staging_test.go` needed because the recorded
  version is the `0.17.0-dev` prerelease. `logs/test-2.txt` is the earlier run
  that exposed and led to that fix.
- `go test -race ./...` exit 0 (`logs/race-1.txt`).
- `scripts/build.sh` exit 0 with the real Contact bundles, including its
  `FLEA_REQUIRE_APP_BUNDLE=1` `TestManaged*` run (`logs/build-1.txt`); the built
  CLI reports `0.17.0-dev`.
- Real sample-app lifecycle `TestCustomAppHostLifecycleAcrossReleases` (Pack to
  Extract to Load to launch, a session across a release change, alias isolation
  and cross-site rejection) plus the sample build and vet passed.
- Fresh temporary-site smoke: `flea new`, canonical starter parity and
  `flea status` exit 0 (`smoke/`).

The independent read-only review of the integration diff
(`/tmp/deepseek-task.kaG0zA/report.md`) found no material code, test, security or
provenance defects. Its remaining items were low/documentation only; this handoff
correction resolves the published-version and disposition-provenance notes, and
the accepted `internal/buildinfo/staging_test.go` fixture adaptation is recorded
above.

The final test/documentation revision added the real
`TestBlogSaveKeepsCustomAppSessionAndManagedRuntime` regression: Blog and the
sample session app coexist, the literal JSON attribute decodes correctly, and a
Blog save plus host reload preserve both session authentication and the stable
managed runtime directory. It and the A-to-B lifecycle passed under race;
`FLEA_REQUIRE_APP_BUNDLE=1 go test -race ./internal/host/ -count=1` and the
real-bundle `TestManaged*` race checks also passed, with no skips.
Evidence: `/tmp/deepseek-task.CH2sdh/logs/`. Lead review of this narrow delta
confirmed no runtime source changes after the independent review. Permanent
reports and archived verification evidence are kept with the private checkout
backup named above.

## Remaining release gates

These are genuine future release and deployment gates, not local test work:

- A matching server release at deployment protocol 26. The published 0.16.1
  server is protocol 25 and refuses a custom-app package.
- Native Linux/systemd acceptance, clean-Mac/AMD64 acceptance and real email
  acceptance. None are established by local tests.

## Canonical checkout

The reviewed integration is commit `0c850e7` and has been fast-forwarded into
`main` at `/Users/sean/Projects/flea`. Use this canonical checkout for further
framework development. Its `bin/flea` and generated embedded bundles match the
verified development build (`0.17.0-dev`). Installed/public runtimes remain
`0.16.1`; this source consolidation is not a production deployment.

The exact original snapshot remains on the pushed branch
`archive/custom-apps-before-reconcile-20261018` (`719c504`) and in the private
full-checkout archive. The integration branch also remains available for audit.
Private implementation/review/revision reports and verification evidence are
retained under the backup directory above. No custom-app source was discarded.
