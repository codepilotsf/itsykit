# Publishing ItsyWeb

Run framework commands in this guide from `system/` in the combined
`itsyweb` repository. Build outputs are `system/bin/` and `system/dist/`
from the repository root; the public site is the sibling `website/`. The root
GitLab CI configuration changes into `system/` and collects its artifacts.
CP Git hosts source and has no automatic release pipeline; native release checks
and explicit publication remain required. Historical release tags retain their
original layouts and must not be moved or recreated.

## ItsyWeb 0.22.0 rollout

Version 0.22.0 removes the page CMS and adds comment metadata, requiring deployment
protocol 30. Existing page-CMS sites are incompatible: back up and explicitly
convert their saved content into ordinary HTML/JSON and comment metadata before
upgrading. Preserve Blog posts/media, runtime state and host policies. The normal
upgrader rejects incompatible sites; it does not migrate them. Shared hosts need
a coordinated, validated cutover with paired executable/configuration rollback.
Both native CLIs and the shared server are published/upgraded. All 22 sites were
explicitly converted with verified backups and rendered-content comparisons.
See [the 0.22.0 receipt](release-0220-verification.md) for actual release evidence.

## ItsyWeb 0.21.0 rollout

Public 0.21.0, both native CLIs and the shared server are updated. All native
release and public installation checks passed; see [the rollout receipt](release-0210-verification.md).

Development builds produce `bin/itsy` and a retired `bin/flea` link. Canonical
public artifacts are `itsy-darwin-arm64` and `itsy-linux-amd64`, hosted at
`https://itsyweb.dev/releases/`. Stage byte-identical `flea-*` aliases alongside
them so released Flea updaters can still upgrade from fleacms.com. The host
catalogue accepts both names; retained versions and bytes stay immutable.

Invoke downloaded artifacts through an `itsy` name for CLI smoke checks. The
private old `.flea-update-* version` probe remains solely for released updaters.
From a pre-rename CLI, the first `flea update` installs the transition binary;
a second `flea update` installs `itsy`. Verify both steps, installer aliases,
and the strong notice on both native supported platforms before publication.

The 0.21.0 rollout introduced `itsy.config.json` and `itsy/` with explicit legacy
aliases and protocol 29. Its historical validation covered old site files, sessions,
custom-app storage and the then-supported toolbar initializers.
Publish verified artifacts before advancing either website's latest.txt or
installer. Native Mac checks remain required even when the Linux server and
local site conversions are ready. Existing release receipts below retain
historical names; record actual rollout evidence separately.

For the manual public release workflow, SemVer rules, low-memory host safeguards
and recovery steps, follow the [deployment runbook](deployment-runbook.md).
Since 0.7.0, versioned binaries live in the host's separate
[disk-backed download store](download-hosting.md), not the website Git tree.
Use the store publication procedure before advancing website latest.txt. Older
sections describing website version directories are historical staging guidance.
0.18.0 adds creation in empty destination directories and removes the Linux ARM64
CLI from the supported release matrix. Apple Silicon macOS and x86-64 Linux
remain supported CLI platforms. Both CLIs retain x86-64 and cross-compiled ARM64
Linux servers; native ARM64 server acceptance is recorded separately and does
not block CLI publication. That release uses deployment protocol 26 and app protocol 2.
See [0.18.0 verification](release-0180-verification.md) for actual publication,
artifact and acceptance status. Historical published downloads remain immutable.

0.19.0 adds hostname indexing with protocol 27 and needs matching clients and
servers. See [0.19.0 verification](release-0190-verification.md) for build and
rollout status. Never replace existing published bytes. Follow the separate
[host-indexing rollout](../host-indexing.md#codepilot-rollout-and-rollback), keeping
CodePilot's current staging fallback until live native enforcement is verified.

0.20.2 adopts the approved Little Skip/Nunito branding in the starter, diagnostic
page, website and docs, with outlined SVG logos and browser/touch icons. The
public channel, shared server and normal Kirin/Mac CLIs are updated. See
[0.20.2 verification](release-0202-verification.md). Custom customer branding and
live content are preserved. Protocols remain 28 and 2.

0.20.3 adopts the approved filled Little Skip treatment while preserving the
Nunito wordmark, site palettes and protocols. See [0.20.3 verification](release-0203-verification.md)
for native checks, public artifacts, rollout and remaining limits.

0.20.4 fixes browser/touch icons to white on navy in both browser themes. See
[verification](release-0204-verification.md). The shared executable remains 0.20.3.

0.20.5 uses the approved navy-on-light branding throughout primary logos and
icons. See [verification](release-0205-verification.md) for native/public checks
and the light website rollout. The compatible shared executable stays 0.20.3.

## V1 distribution plan

0.20.0 includes migration redirects and opt-in SEO discovery with deployment
protocol 28. The public channel and matching shared server are deployed; see
[0.20.0 verification](release-0200-verification.md). Other hosts still need an
explicit matching server update before deploying these settings; older servers
silently ignore them. Publication and server rollout remain separate operations.

The agreed plan is in [next-steps.md](next-steps.md). Source remains private;
public binaries will be free to use without an account or license key. The
primary path is the implemented `../website/site/install.sh`, intended for
https://itsyweb.dev/install.sh, followed by `itsy update` for local CLI updates.
Remote updates use `itsy update-server` (including `--recover`), without an alias.
The public website source is the sibling `../website/`. Its
[RELEASES.md](../../../website/RELEASES.md) records build provenance and
actual public installer/update checks. Versioned binaries, SHA256SUMS,
latest.txt and install.sh are live on fleacms.com. Public installation, reinstall,
update and already-current checks pass on both currently supported platforms. The
Mac binary is ad-hoc signed; clean-machine Mac testing and notarization remain
unverified, so do not describe this as a completed V1 release.
See [installation and trust](../installation.md).

CLI release targets are macOS ARM64 and Linux AMD64. Each uses a native runner
for tests, race checks, vet and bundled CLI smoke tests. Linux ARM64 and macOS AMD64
are outside the public release matrix. The pipeline has not been run remotely;
manual verification must record any cross-compilation and remaining native gaps.
Public artifacts are hosted in the separate disk-backed download store without
opening the source repository. The registry instructions below are historical.

`scripts/release.sh client <arch>` now supports native Linux as well as Mac
runners. Each needs both previously built Linux server artifacts in dist/.
After native checks for both supported CLIs, `scripts/stage-release.sh` prepares only
`dist/public/install.sh`, `releases/latest.txt`, and immutable
`releases/<version>/` binaries plus SHA256SUMS. It requires a matching stable tag
and refuses an existing version staging directory. It does not upload anything.
Publish the version directory first, verify anonymous HTTPS downloads, then
advance latest.txt. Never overwrite an already published version. The old GitLab
publish job is not the new public channel and must not be used for V1 as-is.

Homebrew and native packages are deferred for V1. Their existing tooling is
described below for reference, not as a required V1 distribution path. Test the
actual downloaded Mac binary and its platform trust requirements regardless of
packaging. Follow the sprint's release criteria before publication.

## Mac distribution tooling

After `bash scripts/build.sh`, `bash scripts/package-macos.sh` creates a native
installer under dist/ without installing it. Optional APPLICATION_IDENTITY and
INSTALLER_IDENTITY sign the executable and package. Public distribution still
requires notarization and independent installation testing. An unsigned package
is a development artifact, not a public release.

Local packaging verification generated a readable package with the CLI payload,
but pkgbuild emitted permission warnings and included AppleDouble metadata,
including after copying without extended attributes and running outside the sandbox.
Resolve this on the signing runner and verify a clean payload before release;
installation, signing and notarization have not been verified.

After building release binaries, set RELEASE_BASE_URL to the HTTPS release
directory and run `bash scripts/homebrew.sh` to generate dist/itsy.rb with real
checksums. Review and publish it to a Homebrew tap separately. These scripts do
not publish artifacts, create a tap, or modify a user's installation.

Local verification built both Mac architectures, generated the formula using a
non-publishing placeholder URL, and passed Ruby syntax validation. A real tap
installation and downloaded-artifact verification remain pending.

The sibling `../website/` is the published website source. Repository moves do
not redeploy it; preserve its design, deployment mappings and live content.

End users install only ItsyWeb. Setup uses native SSH/SCP and extracts the
matching bundled static Linux server; neither machine needs Go or source code.
Both Linux architectures are gzip-compressed inside each CLI, with SHA-256 hashes.
Each CLI/server also embeds a separate matching-platform Contact child. Build it
first with scripts/bundle-app.sh; scripts/build.sh and scripts/release.sh do this
in dependency order. Release tests require FLEA_REQUIRE_APP_BUNDLE=1, including
real child launch, handshake and shutdown. No Go compiler is needed on the VPS.
A source-only go build can omit or reuse stale bundles and is not a release.
Setup needs SSH access, not a release host or package registry connection.
Framework contributors and CI still need Go 1.26 or newer. Deploy also uses Git,
as required by the existing committed-source workflow.

## GitLab prerequisites

The repository's .gitlab-ci.yml builds tagged releases on native runners:

- linux-amd64 for native server tests/builds and cross-compiling the ARM64 server.
- macos-arm64 and linux-amd64 for native CLI builds and tests.

An ARM64 Linux runner is optional for additional native server acceptance. A
cross-built server is not native-tested; record that gap in each release receipt.

Register trusted shell runners with those exact tags. Each needs Go, Git, Bash,
gzip and shasum; the publishing Linux runner also needs curl. Race tests need a C
compiler on runners only. Go does not run on either end user's machine or VPS.
Jobs verify host architecture and refuse cross-architecture release builds.

The older `scripts/publish.sh` targets a GitLab generic package registry and is
not invoked by the V1 pipeline. Do not enable anonymous source access to make
downloads work. Select public binary storage separately.
For the older registry workflow only, enable anonymous package downloads.
The registry distributes releases but is not used by setup. No user tokens are
embedded. Protect release tags and the runners. Disable duplicate
generic package versions so a published version cannot silently be overwritten.
GitLab package API: https://docs.gitlab.com/user/packages/generic_packages/

## Release procedure

1. Finish and review changes. Update `internal/buildinfo/version.go` and move
   the Unreleased changelog entries into a dated section for that version.
   Commit the changes and run the tests.
2. Create and push a new immutable tag, for example v0.2.0.
3. The native Linux AMD64 job tests/builds its server and cross-compiles the ARM64 server.
4. The native Mac ARM64 and Linux AMD64 CLI jobs embed both Linux artifacts using
   scripts/bundle.sh, test extraction and build CLIs with the same version.
5. Review the artifacts, then run the manual publish-stage job. Despite its
   stage name, it only prepares downloadable `dist/public/` CI artifacts.
   It does not upload to fleacms.com or advance the public update channel.
   Upload to the selected public host only after approval and verification.
6. Test the downloaded Mac CLI on a clean Mac and setup on a disposable Ubuntu
   machine for x86-64 before announcing the release or updating a tap. Record
   ARM64 server native acceptance separately when hardware is available.

Public CLI artifacts are itsy-linux-amd64 and itsy-darwin-arm64, with matching
flea-* aliases for released updaters. Build artifacts also include
itsy-server-linux-amd64 and itsy-server-linux-arm64 for embedding.
Linux ARM64 CLI artifacts are no longer built, staged or published.
A Homebrew formula can install the appropriate Mac binary directly, with its release SHA-256 and
no Go dependency. Creating/publishing that tap remains a separate step.

## Development builds

Run from anywhere on this Mac:

```sh
bash /Users/sean/Projects/itsyweb/system/scripts/build.sh
```

Use the normal Go cache and, when temporary storage is constrained, the private
per-job `TMPDIR` pattern in [development storage](../development-storage.md).

The script locates Go on PATH or /usr/local/go/bin/go (GO_BINARY overrides it).
It cross-compiles both Linux servers with CGO disabled, compresses them with
deterministic gzip headers, stores checksums, tests extraction, and builds the
native CLI at bin/itsy. The default version comes from internal/buildinfo/version.go.
Use VERSION=0.2.0-rc.1 for an explicit local build override. Values must be valid
SemVer without a leading v; prerelease and build metadata suffixes are supported.
Release and publish scripts require CI_COMMIT_TAG to match v plus the recorded
version. CLI and server output omit the tag's v prefix.
Contributors need Go 1.26+, Bash, gzip and shasum. Initial builds may download Go
modules. Installed users do not need these build tools.

Generated assets under internal/serverbundle/assets and bin/ are ignored by Git.
Always use the script after changing source: plain go build may embed stale
servers, or none in a fresh checkout. Source-only tests work without bundles;
FLEA_REQUIRE_BUNDLE=1 go test ./internal/serverbundle requires both real
binaries. The server never imports serverbundle, avoiding recursive embedding.
No publishing is necessary for development setup. Extraction is bounded and
checks the hash and static ELF architecture before reset confirmation or upload.

The pipeline is provided but has not been run on the remote GitLab instance.
Runner registration, public registry access, first publication, and real-machine
installation tests must be completed before claiming release readiness.
