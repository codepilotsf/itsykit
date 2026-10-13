# Publishing FleaCMS downloads

## Blog documentation deployment — October 13, 2026

Website source `54815a7` on `docs/blog-setup`, based on current main `7e55082`.
[Draft MR 2](https://gitlab.codepilot.com/sean-schertell/flea-website/-/merge_requests/2)
remains open. Deployed an isolated main checkout of the exact committed snapshot
to the existing fleacms.com destination; original working checkout/branch was
preserved and no source merge was performed.

Activation: `20261013-194213-97642e56ad639a7acdb3ac8a`. Blog setup is public at
https://fleacms.com/docs#blog and linked from the homepage and docs contents.
The guide covers config/template adoption, editing, agent JSON, tokens and sync.
It labels the 0.15.0 / protocol 25 requirement and the public 0.13.0 channel
accurately. No release artifacts or channel files changed.

Validation: `flea status`, rendered anchors/literal tokens, highlighted JSON and
HTML examples, desktop/mobile browser checks and live HTTPS verification passed.
At 375px there is no horizontal page overflow. Live JSON examples parse as shown.
After deploy, `flea status` reports code and content matching the committed source.
All eleven HTTPS homepages are healthy, recovery is clear, and website saved
content/media/CSS plus all unrelated-site fingerprints are preserved. Runtime
remains 0.15.0 / protocol 25. No real email was sent.

Private backup: `/tmp/flea-website-blog-rollout/website-before.flea-backup`.
Before/after fingerprints are in that private directory. The deployment used the
existing 320 MiB admin-process memory target; no service-policy change occurred.
No preview process remains. These are website documentation checks, not new
native binary-release acceptance results.


Public downloads live outside this website in the server's administrator-owned
`/var/lib/flea/downloads/fleacms/<version>/` store. URLs remain
`https://fleacms.com/releases/<version>/`. Never replace published bytes or hashes.
The website contains only `site/releases/latest.txt`; do not add binaries back.
Local retained artifacts are archived in `/Users/sean/fleacms-releases/`.
Older Git commits contain binaries; do not rewrite that history.

Publish a new version's binaries and checksum manifest first. Fetch them over
HTTPS and verify their checksums before advancing `site/releases/latest.txt`.
`site/install.sh` and `flea update` read that pointer. Test actual downloads,
installation, reinstall and update before advertising the command.

Supported downloads: Apple Silicon macOS (`darwin-arm64`), x86-64 Linux
(`linux-amd64`) and ARM64 Linux (`linux-arm64`). Do not list additional platforms without native checks.
The Mac executable is ad-hoc signed, not notarized. This is an early test release;
clean-machine testing remains pending.

## 0.13.0 published October 12, 2026

MR !4 is merged. Source/tag `bd668a1` / `v0.13.0`; Go 1.27.1, clean tagged
builds (`vcs.modified=false`). `flea status` now shows local validation, site
counts and Git state first, then automatically checks this branch's saved remote
with terminal progress. Website reachability, deployed code, editable content and
theme are separate. Scoped pull baselines establish change direction; status
never pulls or changes deployment settings/baselines.

Native ARM64 Linux and Mac full tests, real-bundle race checks, vet and bundled
builds passed. Linux AMD64 was cross-compiled off production and smoke-tested
natively as nobody, including the exact server's version/protocol. Native AMD64
full-suite/build, clean-Mac acceptance and notarization remain open. Tag pipeline
358 is pending without native runners; verification was manual. The Mac executable
passed ad-hoc signature checks.

Every final CLI was checked to contain both canonical Linux server executables
and their matching checksums, comparing uncompressed bytes because gzip output
varies by platform. Server hashes:
AMD64 `02e55e82ee6f49d9a46e7f7454f4f6880dfc25b9ac253a21cf842323eba00b84`,
ARM64 `15841a2c222d0ff65eb787e7bc028bf5c1938cf8f35a2b42338fe47fb039c84e`.
Protocol 23 and Contact behavior are unchanged.

The complete immutable store directory was checksum-checked under the deployment
lock, validated and loaded by restarting the existing service. All anonymous HTTPS
downloads matched before the public channel advanced; the old 0.12.0 manifest
remained unchanged. Website `1f8d4bf`, activation
`20261012-182752-488ee26bcaf1fb17ddf715ec`, then published latest.txt, the homepage,
current download links and status documentation. No binaries entered website Git.

Actual public install/reinstall, installed hashes/version, 0.12.0-to-0.13.0 update
and already-current checks passed on all three native platforms. Published CLIs
passed default/starter initial commits, status and dev routes. The exact final
ARM64 CLI also passed an animated-terminal status check against the configured
website, including local-first output, remote comparison and spinner cleanup.
Test dev processes were stopped. Native password/sudo prompts on a live non-root
remote target were not exercised.

All hosted domains are healthy; content, theme and routing hashes match before/
after. Recovery is clear and memory/swap policy unchanged. Shared server stays
0.9.0-rc.1 / protocol 23; normal user installations were not changed. Backups,
private build copies and retained artifacts are outside website Git.

- Mac ARM64: `683be74c63f09a308bcfff4be11bb3a82cc5a08527611af7ce43bbd50b538052`
- Linux AMD64: `34da6f5ad433adbe2e4b28ed45bd19fa18e7dd461816bc633ecf65a678cb2076`
- Linux ARM64: `84be037ce925e6ff67640ab9abfcac2c7654097152340631ec396ccbb3b3be07`

## 0.12.0 published October 12, 2026

First public Linux ARM64 CLI. Source/tag `a227ff4` / `v0.12.0`; Go 1.27.1,
clean tagged builds (`vcs.modified=false`). Mac ARM64 and Linux ARM64 full tests,
real-bundle race checks, vet and bundled builds passed natively. Linux AMD64 was
cross-compiled off production and verified there as nobody; native AMD64
build/full-suite acceptance remains open. CI pipeline 357 is pending absent
native runners; final verification was manual. Mac signatures are ad-hoc,
not notarized; clean-machine acceptance remains open.

All three CLIs contain one canonical verified set of newly built Linux server
executables and matching Contact children. The tag was synchronized before the
final builds so Go version metadata agrees across build hosts. Server hashes:
AMD64 `b7447fb1485b6a04d3805eb74402682ae4d44777060f249bd58583d3019c197f`,
ARM64 `40a83e9b69f8106bcb00a1646147157f6d68f42878b515b7ebd310547a7fc59d`.
Protocol 23 remains unchanged.

The complete version directory was checksum-verified and installed under the
deployment lock, validated before restarting the shared service, and verified
through anonymous HTTPS downloads. Only then website `6ebdb00`, activation
`20261012-174635-ec70336ae4e81090c4bc0c21`, advanced latest.txt and added ARM64
download links. It preserves the committed docs syntax highlighting and ships
the clearer missing-platform installer error. Old versions/hashes are immutable.

Public fresh install/reinstall, installed hashes/version and already-current
checks passed on all three native platforms. Mac/AMD64 updated from public
0.11.0; ARM64 updated a local 0.11.0 fixture because no prior public ARM64
binary exists. Published CLIs passed default/starter initial commits, status
and dev homepage/team/contact checks. Test dev processes were stopped.

All six hosted sites are healthy; content/theme hashes match before/after,
recovery is clear and memory/swap policy unchanged. Shared server remains
0.9.0-rc.1 / protocol 23, restarted only for catalogue/site publication. Normal
user installations were not replaced. Private website backup and isolated build
copies/artifact archives are retained outside website Git.

- Mac ARM64: `e7d4858a5118ce141a27fce4a87cdf0418dffc06551977c48acce72be050dc41`
- Linux AMD64: `12f5bc368f9611e070224368ce299c09eff3d8ddd36df1efd03c2fb6baf3fc5a`
- Linux ARM64: `e623545617796816a055e81a43cf7aa401543bb2fc5d963791c4cd8448d4e9f5`

## 0.11.0 published October 12, 2026

Framework `9759a51`, tag `v0.11.0`, built with Go 1.27.1 from a clean snapshot
on Mac ARM64 (`vcs.modified=false`). Adds an `Initial site` commit for `flea new`
and prints remote connection/push commands. Git identity, hooks and signing are
respected; failed commits preserve the site. Linux ARM64 CLI support is excluded.
Both Linux server architectures and their matching Contact children were rebuilt
and bundled. Protocol 23 and the Contact behavior are unchanged.

Native Mac full tests, vet, race checks with real bundles, bundled build and
ad-hoc signature verification passed. Linux AMD64 was cross-compiled off the
production host and verified there as nobody. It was not built or fully tested
on a native AMD64 runner. The tag pipeline remains pending for missing native
runner tags; release verification was performed manually. Clean-machine Mac
acceptance and notarization remain open.

The new immutable store directory was uploaded beside the store, checksum-checked,
installed under the deployment lock and validated before restarting the service.
The immediate readiness check raced startup; a subsequent check and all public
site checks passed without rollback. The public channel stayed at 0.10.0 until
anonymous HTTPS downloads matched both new hashes and the old manifest remained
available. Website `26ed807`, activation
`20261012-164409-6fcf3b6569f135a51848d882`, then advanced homepage/docs/latest.txt.
The installer is unchanged.

Public fresh install, reinstall, 0.10.0-to-0.11.0 update and already-current checks
passed natively on Mac ARM64 and unprivileged Linux AMD64. Installed hashes matched
below. Default/starter creation produced clean initial commits on main; status
and dev homepage/team/contact checks passed. Test dev processes were stopped.
All six active sites are healthy and content/theme hashes matched before/after.
No recovery journals remain; existing memory/swap policy is unchanged. Shared
server stays 0.9.0-rc.1 / protocol 23; it was restarted but not upgraded. Normal
user installations were not changed. Website backup is retained privately outside
this repository. Build/artifact records are retained in isolated release paths.

- Mac ARM64: `b4290260582449a599d429eada9fdbd98780588ee6572cbd6a7a09cc0bf66015`
- Linux AMD64: `e3c2303d9eec074a8029f49b39927776617a9e2141ead43284eed363871ffbbf`

## 0.10.0 published October 5, 2026

Framework tag `v0.10.0` (`d6a7f7c`), built with Go 1.27.1 on Mac ARM64 from the
clean tag (`vcs.modified=false`). Renames `flea server sites` to `flea server
list` with no alias; `list` and `remove` now prompt for the server login when the
SSH target is omitted. The bundled Contact app and deployment protocol 23 are
unchanged from 0.9.0.

The 0.10.0 directory was staged beside the store, its manifest verified, and it
was validated and atomically installed under the deployment lock; the shared
service was restarted only after validation to load the catalogue, and no
recovery journals remained. Anonymous HTTPS downloads matched both published
hashes and an 0.9.0 manifest/binary check still passed. Public fresh install,
reinstall, 0.9.0-to-0.10.0 update and already-current checks passed on Mac ARM64
and native unprivileged Linux AMD64 with installed hashes matching below. The
homepage, docs and `latest.txt` show 0.10.0; the installer is unchanged. Website
source `b95eabb`, activation `20261005-203837-d9261af00e49e2d96f2fd52f`.

The shared server executable remains `0.9.0-rc.1` / protocol 23; it was not
upgraded, only restarted to load the new download catalogue. All four active
sites are healthy, live content and themes were preserved, recovery is clear and
memory/swap policy is unchanged. `flea update` replaced Sean's installed CLI.
Test dev processes were stopped. Full native Linux build/suite, clean-machine Mac
testing and notarization remain unverified.

- Mac ARM64: `9502a45980b7abbcfc015fb5fd3268831ed17b6f6f758fd3dd86829b3eadee74`
- Linux AMD64: `f6c2b6f34df19f9fa702741463835d1ab52869f340649c6e92dfa56ffedd9725`

## 0.9.0 published October 5, 2026

Framework tag `v0.9.0` (`e599a5e`), built with Go 1.27.1 on Mac ARM64 from the
clean tag (`vcs.modified=false`). Adds the bundled Contact app with protocol 23
and server-wide site administration, replacing top-level `contact` with
`apps.contact`. Full native Linux build/suite, clean-machine Mac testing and
notarization remain unverified.

Website source `5ac7101`, activation `20261005-172223-0d03e392e05da991928a068b`.
The 0.9.0 directory was staged beside the store, verified and validated under the
deployment lock, then the shared service was restarted to load it; no recovery
journals remained. Anonymous HTTPS downloads matched both published hashes, and
an 0.8.1 manifest/binary check passed. Public fresh install, reinstall,
0.8.1-to-0.9.0 update and already-current checks passed on Mac ARM64 and native
unprivileged Linux AMD64 with installed hashes matching below. The public
channel, homepage and docs show 0.9.0; the installer is unchanged.

The shared server executable remains `0.9.0-rc.1` / protocol 23. It was not
upgraded, only restarted to load the new download catalogue. All four active
sites are healthy, live content and themes were preserved, recovery is clear and
memory/swap policy is unchanged. `flea update` replaced Sean's installed CLI.
Test dev processes were stopped.

- Mac ARM64: `ab076b61597836348000f7d0d0f956f38b269c43fffb03abdd41711824037950`
- Linux AMD64: `b6ab6639c5b3c9f72b7195c7198f41b67872115bc86182fe40fa5775d54608d2`

## Private server rollout October 5, 2026 (not a public release)

Server 0.9.0-rc.1 / protocol 23 from framework `3447cb7`. The standard bundled
build produced Mac CLI SHA-256
`505642a47e0c558b4bfb3a4737bb4a408894768d2c836391784f53397e437fe6`.
Website `7ef4e6e`, activation `20261005-163126-76b53507e24be701230d2e89`, adds
Contact at `/contact` with `/contact/thank-you`. Atlas was also updated;
Bookworm's routing was removed with its files retained. All remaining sites are
healthy, content/themes preserved and recovery clear. Contact CSRF/422/honeypot
303 checks passed on both sites without sending email. Real delivery, native
Linux full-suite and clean-Mac acceptance remain unverified for this rollout.
Public downloads, installer and channel remain 0.8.1. No RC artifacts or tag
were published. Documentation-only follow-up commits are not deployed snapshots.

## 0.8.1 published October 3, 2026

Framework `209b5d3`, immutable tag `v0.8.1`, built with Go 1.27.1 on Mac ARM64.
Fixes repeated `.gitignore` rule appends and preserves existing trailing blank
lines. Package/full tests, vet, bundled build, Mac signature, fresh-site validation
and three consecutive dev starts preserving `.gitignore` bytes passed.
The cross-compiled Linux AMD64 binary passed native unprivileged version/new/status
checks. Full native Linux build/suite, clean-machine Mac testing and notarization
remain unverified.

Website source `9fd1397`, activation `20261003-230705-de4bfab5fbec8834e722a460`.
Anonymous downloads matched both hashes. Public fresh installs, reinstalls,
updates from 0.8.0 and already-current checks passed on Mac and unprivileged
Linux. The downloaded Linux binary also preserved `.gitignore` byte-for-byte
across three dev starts. Sean's installed CLI is the published 0.8.1 binary;
the previous backup remains. The public channel, homepage and docs show 0.8.1.
The shared server remains 0.8.0: this CLI fix did not need an executable upgrade.
Service readiness, Bookworm and retained 0.8.0 manifest checks passed; recovery
journals are clear and memory/swap policy is unchanged. Test dev processes stopped.

- Mac ARM64: `6d9141a3bc4393c2e3d1a0e2f3fc6de1a207b4a9e3955406d865f46c1b37cf31`
- Linux AMD64: `72ab592758addecde4b91666bec10acc735d949679abc7a2bb66600b0616effc`

## 0.8.0 published October 3, 2026

Framework `c8a1d96`, immutable tag `v0.8.0`, built with Go 1.27.1 on Mac ARM64.
The edit row collapses outside Edit; Theme slides from the right and tracks the
available height. Existing sites require the current starter toolbar partial.
Full local tests/vet, browser regressions, bundled server builds, Mac signature
and fresh-site checks passed. The cross-compiled Linux AMD64 CLI passed native
unprivileged version/new/status checks. Full native Linux build/suite and clean
Mac/notarization remain unverified.

Website source `3188000`, activation `20261003-171440-ca76afb853e32d0d0309c6ea`.
Shared server and public channel report 0.8.0. Anonymous downloads matched the
manifest. Fresh installs, updates from 0.7.1 and already-current checks passed
on Mac ARM64 and unprivileged Linux AMD64, with installed hashes matching below.
All four public sites, docs and social image returned 200. Recovery is clear;
existing memory/swap safeguards were preserved. No mailer changes were made.

- Mac ARM64: `f1792257d2c4c62976a934178ef964795a8f91590b18f5a17d8e6fe2c38edbb7`
- Linux AMD64: `e441b4f5ae3bf4a6c60ede86e9ac10a0d7f4740b72723c726214e46737663596`
- Naves toolbar source `c59dfff`, activation `20261003-171113-e1b9be7c7e17a28c8aca4ffd`.
- Kent toolbar source `d80eaeb`, activation `20261003-171202-d490ea0173f4c49cbd1be27a`.
- Bookworm was deliberately excluded at Sean's request. Its existing toolbar
  remains incompatible with the 0.8.0 editor; public pages are unaffected.

## 0.7.1 published October 3, 2026

Framework `f4a72ee`, immutable tag `v0.7.1`. Server-side progress uses the
installed TLS domain list: first-time enablement shows certificate guidance;
routine deployments show "Checking site health and HTTPS...". Checks and renewal
are unchanged. Regression tests cover repeat deploys, first enablement, aliases,
HTTP-only sites and routing-preserving rollback. Full tests/vet and bundled builds
passed; Mac signature/fresh-site and native unprivileged Linux new/status passed.
Linux was cross-compiled; full native Linux tests and Mac clean-machine/notarization
remain unverified. Artifacts are published through the separate download store.

Website source `f75d63f`, activation `20261003-151507-4d9b7bc05ab6b470b4ae679b`.
The real deployment printed "Checking site health and HTTPS..." with no setup
warning. Public Mac/Linux downloads matched their manifests; fresh installs and
updates from 0.7.0 passed, Linux unprivileged. Server, channel and Sean's local CLI
are 0.7.1; both hosted sites returned 200 and recovery journals are clear.

- Mac ARM64: `ca1215301e7728f04a48bedd576d6ee705f71183a167a7ec8133d5e6aef2d66a`
- Linux AMD64: `5a8b032b43e7b6481975251dd1390ff37feeb25c0c505cac5da3c4dcde33753a`

## 0.7.0 published October 3, 2026

Framework commit `1199747`, immutable tag `v0.7.0`. Includes reviewed interactive
design checkpoints in the bundled skills and confined disk-backed release hosting.
Mac full tests, vet and race checks passed. Both embedded Linux servers were
rebuilt; Mac ARM64 signature and generated-skill comparisons passed. Linux AMD64
was cross-compiled and passed native unprivileged version/new/status checks.
This is not a native Linux build/full-suite run or clean-machine Mac acceptance.

Server upgraded to 0.7.0. All seven retained versions (0.3.0 through 0.7.0, with
the existing version gaps) passed public HTTPS manifest and binary hash checks
from the new store before removing their duplicate files from this website.
No published version was deleted or overwritten. latest.txt remains site-managed.
The host mapping was backed up at
`/etc/flea/server.before-downloads-20261003T141633.json`. Migration initially
installed server.json with private permissions; restoring its standard 0644 mode
resolved startup, without weakening the systemd unit. Recovery/health then passed.
Final website commit `a41a758`, activation
`20261003-141847-a863aaed91fd5f4c8fa761b8`. Public assets are about 0.63 MiB,
down from 247.7 MiB; archive upload was 575 KB. Both sites and social image returned
200; download Range returned 206; unlisted files and wrong-host requests returned
404. Fresh public installs, 0.6.2-to-0.7.0 updates and already-current checks passed
on Mac ARM64 and unprivileged Linux AMD64. Server and Sean's local CLI are 0.7.0.
Installed and public hashes:

- Mac ARM64: `4985074561d9a438a63a99ad60f69bc519527401674e6d8e8c207cce32ecd1da`
- Linux AMD64: `63f9ce804c9a72a0a73e43f097a4ce9755e4bd43af85d7095c2e496ed9f6010c`

## 0.6.1 published October 1, 2026

Framework commit `89e3472`, immutable tag `v0.6.1`. Starting dev again replaces
the same site's server gracefully, including LAN and port changes. Pull/reset
remain exclusive, and crash recovery removes only stale control sockets.

Mac tests, vet, race checks, bundled builds and replacement checks passed. The
cross-compiled Linux AMD64 artifact passed native unprivileged checks for site
creation/status, same-port dev replacement, LAN mode and crash recovery. A full
native Linux build/test run and clean-machine Mac/notarization remain unverified.
Both CLIs embed both rebuilt Linux servers. Artifacts live under
`site/releases/0.6.1/`. Both public downloads match SHA256SUMS. Fresh public
installation and updates from 0.6.0 passed on Mac ARM64 and unprivileged native
Linux AMD64. Sean's installed CLI is also 0.6.1.

The shared server is 0.6.1. Final website release
`20261002-015019-3e6152722de7e84b956d7054`, from website commit `1857a45`,
publishes the 0.6.1 homepage, docs links and latest.txt. Historical release
notes and immutable older downloads retain their original versions.

The 1 GB host still exhausted memory with the soft Go target applied to both
service and upgrade processes. After recovery, Sean approved a 1 GB /swapfile,
mode 0600, enabled and persisted in /etc/fstab. The original fstab is retained at
/etc/fstab.before-flea-swap-20261002. Recovery, server upgrade and website
deployment then succeeded. Keep GOMEMLIMIT=320MiB for the service and separate
admin processes. Reducing retained download assets' in-memory footprint remains
a follow-up; swap provides headroom, not a fix for that allocation pattern.

## 0.6.0 published October 1, 2026

Framework commit `04c9b3f`, tag `v0.6.0`. User-facing copy now says "admin"; the first `flea dev`
explains the optional admin email, points to `flea admin` when skipped, and asks
only once; the site-in-use message for `dev`, `pull` and `reset` is clearer.

The framework passed tests, race checks, vet, Chrome sign-in/editing regressions,
bundled build and fresh clean/starter site validation on Apple Silicon macOS.
The Mac CLI is ad-hoc signed, not notarized. The Linux AMD64 CLI is cross-compiled
with `CGO_ENABLED=0`; the actual public download passed version and site validation
on native x86-64 Linux as UID 65534. Site creation also passed with identical staged
bytes. This was not a native Linux build or a full Linux test-suite run.
Both CLIs embed the rebuilt 0.6.0 Linux servers. Published artifacts live under
`site/releases/0.6.0/`, with SHA256SUMS. Public `latest.txt`, installation links
and website copy all target `0.6.0`.

The shared server was updated from 0.2.0 to 0.6.0 with site data/configuration
preserved. Download-only release `20261002-003522-5f12907cc538fdfc0ef629ed` published
the artifacts before the update channel advanced.
Both public binaries match SHA256SUMS; the downloaded Mac binary runs and passes
ad-hoc signature verification. Final active website release:
`20261002-004416-449f529af3c1cafa5f4d8c7f`, from commit `3ebad34` on main.

The 1 GB host exhausted memory during activation. Moving the completed upload
from RAM-backed `/tmp` to `/var/tmp/flea-deploy.cqbA28hiz3` let the download-only
deployment finish. The final website activation then lost its deployment process
to the OOM killer. Sean then explicitly approved a persistent Go memory target:
`/etc/systemd/system/flea.service.d/60-memory.conf` sets `GOMEMLIMIT=320MiB`.
After daemon reload and restart, website/content/channel activation succeeded.
No security settings, site data or credentials were changed by this override.

Fresh public installs and real 0.5.0-to-0.6.0 updates passed on Apple Silicon Mac
and native Linux AMD64, matching the published hashes. Linux checks ran as UID
65534, including already-current update and site validation. The live homepage
shows the revised copy and 0.6.0; docs link the correct binaries and checksums.
The service is active at 0.6.0; Bookworm returns HTTP 200. Sean's installed Mac
CLI was also updated through the public channel to 0.6.0. Clean-machine Mac
acceptance/notarization and a full native Linux build/test run remain unverified.
Superseded candidates are preserved in the ignored release-work directory, outside
the public tree. Their earlier checks do not verify the rebuilt 0.6.0 artifacts.

## 0.5.0 published September 29, 2026

Private framework commit `0994c1f`, immutable tag `v0.5.0`. Adds full opt-in LAN
development (`--lan`/`-l`), localhost startup links and in-place page saves.
Source archive SHA-256:
`6711e8ffc05500eecb555f4c83f05000a48a7b6dc356e0f5d40352c18c33dff8`.

Mac tests, race tests, vet, bundle extraction, native build, ad-hoc signature
verification and starter smoke checks passed. The tagged source passed a real
Chrome sign-in/edit/save test on a private-network HTTP origin.
Native Linux tests, vet, bundled build and new/status smoke checks passed as
UID 65534. The 1 GB host killed a parallel compiler for memory pressure; a rerun
with `GOMAXPROCS=1 GOFLAGS=-p=1` passed. Linux race tests remain unavailable because
the temporary build host has no C compiler. No system toolchain was installed.

Both anonymous HTTPS downloads matched the staged checksum manifest before
advancing latest.txt. Asset-only release: `20260930-005958-5fbeba55cfd5fbaf51a0efe0`.
The initial activation hit host memory pressure and automatically rolled back.
A stale, inactive 188 MB Go build directory was moved from RAM-backed `/tmp`
to `/var/tmp/flea-go-build-archive-1567161428`, preserving it for recovery; the
retry succeeded. No server executable, access settings or Bookworm data changed.
Website commit `f2d3c60`, active release
`20260930-010405-485949623ee87304b3304819`, publishes the update pointer and docs.
Actual public fresh installs, reinstalls, PATH selection, new/status, local and
LAN dev, already-current updates and 0.4.1-to-0.5.0 updates passed on Apple
Silicon macOS and native x86-64 Linux. Linux checks ran as UID 65534. Both
installations used paths containing spaces and matched the published checksums.
Live homepage, LAN docs, installer bytes and private-source 404 were verified.
FleaCMS and Bookworm returned HTTP 200. The temporary Linux source, toolchain,
cache and test installation were removed; the local private archive remains.
Framework main and the annotated tag both resolve remotely to `0994c1f`.
Mac notarization and independent clean-machine/device testing remain pending.

## 0.4.1 published September 29, 2026

Patch release from private framework commit `10dc360`, immutable tag `v0.4.1`.
Shortens `flea new` output to the local development instructions. The preceding
0.4.0 version was a framework development snapshot, not a public download.
Source archive SHA-256:
`9bd091f3babd20891032d162cb4e56c1d8fd7cef195a924ea05fbec38877a8fa`.
Mac tests, race tests, vet, bundle extraction and new/status smoke checks passed.
Native Linux tests (with DAC bypass capabilities dropped), vet, bundle extraction
and new/status smoke checks passed. Linux race tests were not run: this host has
no C compiler. Mac code signature verification passed; notarization remains pending.
Both anonymous HTTPS downloads matched the staged SHA256SUMS before latest.txt
was advanced. Website commit `146fdd5`, active release
`20260929-223440-da9b0ae1b1a6a608fb7c5375`, publishes the 0.4.1 pointer and notes.
Actual 0.3.0-to-0.4.1 updates and fresh installs passed on Apple Silicon macOS and
unprivileged x86-64 Linux. Both verified the shortened new-command output and
site validation. The existing 0.3.0 assets remain unchanged. Bookworm still responds.

## 0.3.0 build provenance

Built from an isolated snapshot of the current private framework working tree,
including the 0.3.0 version update and corrected contact-page and CLI tests. The framework
worktree was not bulk-staged, committed, or pushed.

Private source archive SHA-256:
`01f86b166483c46ff6effa6a2b6ed64061e57cae437ac8e3f9209c3f3c86bbf4`.

The archive is retained locally under ignored `release-work/0.3.0/`, outside
`site/`. Never move it into a public directory. The temporary Linux build copy
was removed after verification, together with the temporary toolchain and Linux
test installation. The local private source archive remains available.

The Mac release pipeline passed tests, race tests, vet, bundle verification,
site creation and validation. Native Linux tests, vet, bundle verification,
site creation and validation also passed. Linux tests ran with DAC bypass
capabilities dropped so root could not mask permission failures; compilation
then ran with ordinary access to generated bundle files. No Linux race run was
performed because this temporary host does not have a C compiler.

Both CLIs contain both Linux server bundles. Only the CLI executables are
published.

## Public verification, September 29, 2026

- Both anonymously downloaded HTTPS executables match the public SHA256SUMS.
- The real public installer passed fresh-directory installation, reinstall,
  PATH selection, version, new-site creation, status and dev HTTP checks on
  Apple Silicon macOS and native x86-64 Ubuntu Linux.
- The Linux installer checks ran as unprivileged UID 65534, not root.
- Install directories containing spaces work on both platforms.
- Both installed CLIs report already-current correctly. A private
  `0.3.0-rc.1` build on each platform successfully updated through the real public
  channel to the exact SHA-256 of the released 0.3.0 executable.
- Installer/updater automated tests passed checksum mismatch, interrupted
  download, invalid executable/version and preservation-of-existing-file cases.
- These were isolated directories on existing machines, not clean OS installs.
  The public pages explicitly invite testing and disclose the Mac notarization
  and platform limits.

The installation pages were deployed from website commit `2205161`, active
release `20260929-202718-8ebaae1f4db520dfa3f53919`. The homepage and docs show the
command; live install.sh matches the reviewed file; latest.txt is 0.3.0. Public
links and anchors resolve. The private source URL returns 404. Bookworm retains
its original release/content/theme and returns HTTP 200.
