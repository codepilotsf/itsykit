# Installation and updates

## ItsyWeb rename (unreleased)

Flea is now ItsyWeb, and the command is `itsy`. In the renamed CLI,
`flea update` is the only working old command: it displays a prominent notice,
verifies the update and installs `itsy` beside the old command. It also installs
`itsy` when the version is already current. All other `flea` commands print
“Flea is now ItsyWeb. From now on, please use itsy.” and exit with status 2.
Use `itsy update` and `itsy update-server` for subsequent updates.

The installer places `itsy` in `~/.local/bin` and retains `flea` as a relative
symlink to it for the notice and update exception. The updater will not replace
an unrelated existing `itsy` during conversion. Config filenames, content paths,
server services, protocol identifiers, release URLs and downloadable artifact
names retain their existing spelling. Local updates do not modify servers.

Already released Flea binaries cannot display new notices themselves. Their first
`flea update` installs the renamed binary under the old name, using the preserved
private version probe. Run `flea update` once more to see the notice and install
`itsy`, or use the new installer. Publishing the renamed CLI and installer is a
separate release step; the public channel has not changed in this task.

Earlier release notes below retain their historical command names.

## 0.20.5 navy-on-light branding

Primary logos and icons use the approved navy foreground on light surfaces.
New sites and development diagnostics receive the matching brand treatment.
Native Linux/Mac and public install/update checks passed. The compatible shared
server remains 0.20.3; protocols stay 28 and 2. See
[verification](agents/release-0205-verification.md).

## 0.20.4 fixed icon palette

Browser and touch icons stay white on navy, matching the public header in both
browser themes. Native Linux/Mac release and public install/update checks passed.
The compatible shared server remains 0.20.3; protocols remain 28 and 2. See
[verification](agents/release-0204-verification.md).

## 0.20.3 filled branding

The filled Little Skip mascot approved on November 9, 2026 replaces the outline
artwork in new sites and development diagnostics. The Nunito wordmark and existing
site palettes are preserved. CLI/server updates retain custom site branding.
The public channel, shared server and normal Kirin/Mac CLIs are 0.20.3. Native
Linux/Mac release and public install/update checks passed; protocols remain 28
and 2. See [verification](agents/release-0203-verification.md) for evidence and limits.

## 0.20.2 approved branding

Published November 2, 2026 from immutable `v0.20.2` (`279ec50`). New sites use
the approved Little Skip mascot, outlined Nunito wordmark and updated browser
and touch icons. CLI/server updates preserve existing custom site branding.
The public website, docs, shared server and normal Kirin/Mac CLIs are updated.
Native release, public installer and artifact checks passed on both supported
platforms. Protocols remain 28 and 2. See
[release verification](agents/release-0202-verification.md) for hashes and limits.

## 0.20.1 temporary storage

Published November 2, 2026 from immutable `v0.20.1` (`c14efd7`). Deployment,
setup and Mac installer packaging respect `TMPDIR`; private invocation directories
and cleanup are preserved. Space/quota errors identify the temporary location and
explain how to choose another one. See [development storage](development-storage.md)
for private disk-backed per-job placement. Deployment protocol stays 28 and app
protocol stays 2. CLI and server updates remain separate commands.

Native Linux AMD64 and Mac ARM64 full tests, race, vet, bundled builds, artifact
smoke and public install/reinstall/update checks passed. The public channel,
shared server and normal Kirin/Mac CLIs are 0.20.1. All 24 hosted HTTPS names are
healthy with live content preserved. Asahi is decommissioned and no longer part
of the installation inventory. See
[release verification](agents/release-0201-verification.md) for provenance and limits.

## 0.19.0 hostname indexing

Published October 24, 2026 from `da7d0c7`, immutable `v0.19.0`. Native Linux AMD64
and Mac ARM64 full tests, race, vet, bundled builds, public install/reinstall and
0.18.0 update checks passed. The public channel, shared server and normal
Kirin/Mac CLIs are 0.19.0. Native staging enforcement is verified after removing
the consumer meta fallback. See the release receipt below for remaining limits.

Adds server-owned exact-host `noindex`/`allow` policy with
`flea server indexing root@server staging.example.com noindex`. Policies survive
deployment, content synchronization, code rollback and site restore without a
website app flag or different templates. Keep robots.txt crawlable. Inspect the
effective policy with `flea server list` or `flea status`.

Deployment protocol moves to 27 with no fallback. Update the CLI with `flea update`
and explicitly update the shared server with `flea update-server` before using
protocol-27 deployment or host administration. Local updates do not change the
remote executable, site releases or running dev servers. Existing configs without
an indexing map keep their behavior. See [host indexing](host-indexing.md) and
[0.19.0 verification](agents/release-0190-verification.md) for rollout status.
Supported CLIs remain Apple Silicon macOS and x86-64 Linux; ARM64 Linux servers
are cross-compiled and native server acceptance is recorded separately.

## 0.18.0

Published October 21, 2026 from `7ab8f6e`, immutable tag `v0.18.0`. Both supported
native CLIs passed full tests, race, vet, bundled builds and public install,
reinstall, update from 0.17.0 and already-current checks with matching hashes.
The live website and public update channel select 0.18.0. Production remains
0.17.0 / protocol 26; this feature required no executable server upgrade.

0.18.0 adds `flea new .` and creation in existing empty directories. Hidden
files and `.git` count as content; nonempty directories are refused without
an overwrite prompt. Folder-derived naming and Git setup are preserved.

Supported CLI downloads are Apple Silicon macOS and x86-64 Linux. Linux ARM64
CLI downloads are discontinued; an existing ARM64 CLI cannot update to 0.18.0,
and the installer rejects ARM64 Linux before changing an installation. Older
published artifacts remain available at their immutable versioned URLs.
Both supported CLIs still deploy to x86-64 or ARM64 Linux servers. ARM64 servers
are cross-compiled; native server acceptance is recorded separately. Deployment
protocol stays 26 and app wire protocol 2. No server upgrade is required for
this CLI creation feature.

See [0.18.0 verification](agents/release-0180-verification.md) for publication,
artifact hashes and the actual platform checks. Historical records below describe
their original platform matrix, not current download support.

## 0.17.0

Published October 18, 2026. Source `fe15b8934165a90390c8bd4ff88cc5e9c9e0d753`
on `main`, immutable tag `v0.17.0`. It adds custom managed apps and app-owned
sessions, with deployment protocol 26 and app wire protocol 2.

A 0.17.0 server or CLI refuses a protocol mismatch in both directions; there is
no fallback or migration. Deploying a site that carries a `flea/apps/<name>/`
package requires the matching protocol-26 server. See [Flea apps](apps.md) and
[release verification](agents/release-0170-verification.md).

Artifacts use Go 1.27.1 with `CGO_ENABLED=0`. Mac ARM64 and Linux ARM64 passed
full tests, vet, race checks and canonical bundled builds; Linux AMD64 was
cross-built on the Mac, then its native CLI and server/systemd paths were tested.
There was no native AMD64 build or full suite. The Mac binary is ad-hoc signed,
not notarized.

Published HTTPS binaries and manifest:

```text
5f9cb8d9bda64d6f00f52eb43e1d38b82cb9d28832f30df5e3b4bbc71973dec5  flea-darwin-arm64
919ed61f7ac31f5ba3163a16b2bce05d1fcd7fcbce8c82c2a667c15a1f16cad4  flea-linux-amd64
8c0ac734abf2dcbb2e9aac6e808a74c83040c21d067c2a3c1906c7fe0b96b09d  flea-linux-arm64
```

The public installer is byte-identical to the canonical
`dist/public/install.sh` and `website/site/install.sh`
(`a15fda6a2d3ebabe776b81f088938c82ff402f51ac9e37c6444eb59ee0dcbeed`). Public
installer checks passed on Mac ARM64, native Asahi ARM64 and native AMD64 as
`nobody`: fresh install, reinstall, update from the public 0.16.1 CLI and an
already-current update. Every installed and updated hash matched the published
manifest, and the old 0.16.1 manifest still resolves. Normal Mac and Asahi CLIs
were updated through `flea update` to 0.17.0 with matching hashes; previous
binaries were retained privately.

The supported shared-server upgrade completed as 0.17.0 / protocol 26. All 18
configured HTTPS hosts return 200, and routing, all editable content and all
deployed source hashes were unchanged across the executable upgrade. The
existing 320 MiB memory target and swap are unchanged, and no recovery journal
remains. Website source `115a73c`, activation
`20261018-195731-803d4f4dc7db80fad2a5fd8a`; `latest.txt` now selects 0.17.0.

### Custom apps in 0.17.0

Custom managed apps, app-owned sessions and the `FLEA_DEPLOY=26` package contract
ship in 0.17.0. The sample app ran under a private transient TLS namespace with
production hardening using the exact release artifact. That is a controlled
acceptance run, not a claim that a customer custom app is deployed in production.

Clean-Mac acceptance, notarization and real email delivery remain unverified.
See [next steps](agents/next-steps.md) for any remaining gates.

## 0.16.1 literal-brace fix

Published October 18, 2026. The public update channel, shared server and normal
Mac/Linux CLI installations were verified at 0.16.1. Public installer/reinstaller,
update and checksum checks passed for all three advertised platforms. Protocol
remains 25. Existing dev processes need a new `flea dev` invocation to load the
updated executable.

The published 0.16.1 release preserves literal brace groups and raw JSON in ordinary
HTML text and attributes, including unchanged Impeccable Live parameters.
Recognizable malformed commands remain errors, and ordinary interpolation,
contextual escaping and CSRF checks remain active. Encode exact Flea syntax
with HTML entities when displaying it literally. See README.md for lexical rules.

Install the published release with `flea update`. Server updates are separate:
run `flea update-server` from a site with the intended saved SSH destination.
The parser fix requires updating both the CLI used for dev and the production
server. Deployment protocol remains 25; no content migration is introduced.
See [0.16.1 verification](agents/release-0161-verification.md) for actual rollout
status and remaining acceptance limits.

## Historical 0.16.0 release

Version 0.16.0 was a public test release for Apple Silicon macOS and
AMD64/ARM64 Linux, published October 13, 2026, and remains the base for 0.16.1.
The installer and `flea update` now select 0.16.1 from
`https://fleacms.com/releases/latest.txt`.

## Updating to 0.16.0

Run `flea update` to update the local CLI and its bundled servers. This includes
optional Blog and the Blog/Contact setup skills. LAN dev now accepts detected
computer names/local aliases and Tailscale IPv4 peers while retaining exact
Host/port, Origin, CSRF and peer checks. Default dev explains occupied ports and
tries consecutive ports from 3000; explicit ports stay fixed. Server updates and
recovery show a terminal spinner after confirmation, or plain redirected output.
Existing project files remain unchanged; adopt Blog or the new skills explicitly as described below.

CSS is developer-managed from 0.14.0 onward. Browser Theme editing, CSS overlays,
`--with-theme`, `--theme-only` and `--snips-only` are removed. Normal deployment
publishes committed CSS and preserves live editable content. Pull retrieves live
content and referenced media; `--with-snips` remains the explicit content sync.

Remote server updates are separate. The 0.16.0 CLI requires server protocol 25. Before
upgrading a protocol-23 server, preserve saved theme customizations in source and
arrange reviewed removal of obsolete host theme mappings; no automatic migration
is provided. Remove the obsolete Theme control from existing toolbar templates,
preserving custom layouts. The shared fleacms.com server remains compatible at 0.15.1 / protocol 25.

Native Mac ARM64 and Linux ARM64 full tests, vet, race and bundled builds passed.
Public install/reinstall, 0.15.1-to-0.16.0 update, already-current and fresh-site
checks passed on all three advertised platforms. Linux AMD64 was cross-compiled
and checked natively as an unprivileged user; its native build/full suite remains
open, as do clean-Mac acceptance and notarization. See
[public release verification](agents/release-016-verification.md).

## 0.15.1 agent skills rollout

0.15.1 bundles `add-blog` and `add-contact-form` guidance in new sites. Generated
`AGENTS.md` routes requests like "add a blog" and "add a contact form" to setup,
template-value and content/form instructions. Blog itself requires 0.15.0 or
newer with protocol 25; this guidance update leaves the protocol unchanged.
Updating a CLI does not modify existing projects: copy the two skill directories
from a fresh matching-version site into `flea/skills/` and add their routing to
that project's `AGENTS.md`, preserving its custom instructions. Skills are
committed agent guidance and never public runtime assets.

The shared server and local Linux ARM64 CLI were updated to 0.15.1 on October
13, 2026. Full tests, vet, bundled build and fresh-site skill checks passed.
All eleven HTTPS domains and preserved content/media/styles were verified.
See [rollout evidence](agents/site-app-skills-verification.md). During that private
rollout, public downloads remained 0.13.0. The subsequent public 0.15.1 release
is recorded in [its verification](agents/release-0151-verification.md).

## 0.14.0 server and local CLI rollout

0.14.0 removes browser Theme editing and CSS synchronization while retaining
`site/css/theme.css`, Prewind and source styling. Normal deployment publishes
committed CSS. Snip editing and synchronization remain supported. Protocol 24
requires reviewed removal of obsolete host theme mappings before upgrading an
existing protocol-23 server; preserve any remote customizations in source first.
The public update channel stayed on 0.13.0 during that private rollout; the
subsequent public 0.15.1 release includes this change.
The shared server and local Linux ARM64 CLI were updated to 0.14.0 on October
13, 2026. All seven sites' saved CSS matched source before the approved one-time
host config cleanup; content/styles and all eleven public HTTPS hostnames passed
post-upgrade checks. Existing sites' toolbar partials still need their Theme
control removed in source. This source-built rollout did not publish downloads
or tags.

## Upgrading to 0.13.0

Run `flea update` for the improved `flea status`. Local validation, page/asset/
editable-area counts and Git state appear first. The command automatically checks
this branch's saved deployment over SSH, showing terminal progress and separate
website, code, content and theme results. No remote flag is needed. With no saved
destination it finishes locally. Status remains read-only; a scoped pull baseline
is needed to identify which side changed. Remote failures exit 1; differences exit 0.

Published October 12, 2026 from clean tag `v0.13.0` (`bd668a1`), Go 1.27.1.
Native Linux ARM64 and Mac full tests, real-bundle race checks, vet and bundled
builds passed. All three CLIs embed identical verified Linux server payloads.
Linux AMD64 was cross-compiled off production and smoke-tested natively as an
unprivileged user. Native AMD64 full-suite/build, clean-Mac acceptance and
notarization remain open; tag pipeline 358 is pending without native runners.

Public install/reinstall, checksums/version, 0.12.0-to-0.13.0 updates and
already-current checks passed on all three platforms. Published CLIs passed
initial commits, default/starter status and dev routes. The final ARM64 CLI also
passed local-first output, animated progress and remote comparison against the
configured website. Older published versions and hashes remain unchanged.

## Upgrading to 0.12.0

This is the first public Linux ARM64 (`aarch64`) CLI release. The same curl
installer now selects its verified download. Each platform has a separate binary,
so ARM64 support does not enlarge the Mac or Linux AMD64 downloads. Existing users
run `flea update`. Missing-download errors identify the selected release/platform.
Published 0.11.0 and older artifacts and hashes remain unchanged.

Public install/reinstall, checksum/version and already-current checks passed on
all three native platforms, including default/starter initial commits, status and
dev routes. Mac/Linux AMD64 updates from public 0.11.0 passed; ARM64 updated from
a local 0.11.0 fixture because no public ARM64 0.11.0 binary exists.

Clean tag `v0.12.0` (`a227ff4`), Go 1.27.1. Native Linux ARM64 and Mac full tests,
race checks with real bundles, vet and bundled builds passed. All CLIs contain the
same verified Linux server binaries. Linux AMD64 was cross-compiled and checked
natively without a production compiler; native AMD64 build/full-suite acceptance
remains open. Mac ad-hoc signatures passed; clean-Mac/notarization remain open.

## Upgrading to 0.11.0

`flea new` now creates an `Initial site` commit on main and prints commands to
connect and push to a remote repository. It respects Git identity, hooks and
signing settings. A failed commit preserves the site and prints retry instructions;
`--skip-git` skips Git setup. Existing sites and remote repositories are unchanged.
Run `flea update` for the CLI. The bundled Contact app and protocol 23 are
unchanged, so this release needs no remote server update.

Public fresh installation, reinstall, 0.10.0-to-0.11.0 update and already-current
checks passed on Mac ARM64 and unprivileged native Linux AMD64. Installed hashes
matched the public manifest. Default/starter site creation, the initial commit,
status and dev routes passed on both platforms. The release came from clean commit
`9759a51`, built with Go 1.27.1 on Mac ARM64; the Linux AMD64 CLI was
cross-compiled and verified natively, not built or fully tested on a native AMD64
runner. Native Mac tests, race checks, vet and signature verification passed;
clean-machine Mac acceptance and notarization remain open.

## Changes introduced in 0.10.0

0.10.0 renames `flea server sites` to `flea server list` with no alias. Both
`flea server list` and `flea server remove` now prompt for the server login when
the SSH target is omitted, and `flea server remove` without a site name lists the
installed managed sites and asks for one by its exact name before the existing
exact-name confirmation. The listing keeps the site name and branch, groups
domains by HTTPS/HTTP, and puts redirects last; release details and duplicate
domain lines are omitted. Removal safeguards are unchanged. Run `flea update`
to install the CLI.

The bundled Contact app and deployment protocol 23 are unchanged from 0.9.0;
this CLI update does not require upgrading a protocol-23 server. When upgrading
from a pre-0.9.0 server, update the shared server with
`flea update-server` before deploying sites that use the new runtime, and update
existing sites from the current starter where their templates changed. The
top-level `contact` settings and the `/_flea/contact` route are replaced by
`apps.contact` with a configured `mount`; there is no alias or automatic
migration, and old settings are rejected rather than converted. Starter Contact
pages use the mount (for example `/contact`) with a mount-relative `thank-you`
target, and their form includes an explicit CSRF control. Custom layouts and
content are otherwise preserved. Isolated Contact delivery has passed; broader
production mail acceptance, the native Linux full suite and clean-Mac acceptance
remain open.

Install it with:

```sh
curl -fsSL https://fleacms.com/install.sh | bash
```

The installer detects macOS/Linux and processor, downloads over HTTPS, checks
SHA-256 and executable version, and installs without sudo to `~/.local/bin/flea`.
If needed it prints a PATH export to add to your shell profile; it never edits
the profile. `FLEA_INSTALL_DIR` may specify another absolute user-owned directory.
Required V1 platforms are Apple Silicon macOS and x86-64 Linux. Other platforms
must pass native verification before being included in published metadata.

The 0.8.0 public installer passed fresh installation and updates from 0.7.1 on
an existing Mac and Ubuntu host, with installed checksums matching the published
artifacts. Linux checks ran unprivileged. Site creation and validation also passed.
These are not
clean-OS tests. The Mac executable is ad-hoc signed, not notarized; independent
clean-machine testing remains pending. The website presents this as an early
test release so invited testers can provide feedback.

For 0.8.0, update existing sites' `site/admin/_toolbar.html` from the current
starter alongside the executable. Its edit-row wrapper is required for the new
sliding toolbar. Preserve each site's custom layouts and content.

`flea update` updates only the local CLI and embedded Linux servers. Downloads
are bounded and checked for checksum, platform and version before replacement.
Failed validation preserves the installed CLI; newer local versions are not
downgraded. Remote servers change only when you explicitly run this from a site:

```sh
flea update-server
```

This uses the branch's SSH destination and confirms before restarting all sites
on that server. Interrupted updates use `flea update-server --recover`.
There is no `flea upgrade` alias. See [Recovery](recovery.md) for boundaries.

Both local installation paths use `.flea-install.lock` in the destination.
After an interruption, confirm no installer/updater is running before removing
that empty directory. Neither path edits shell profiles or remote servers.

HTTPS delivery from fleacms.com is the release trust boundary. Checksums detect
corruption, not a compromised publisher; they are not independent signatures.
Protect publishing credentials and immutable versioned artifacts. Clean-machine
Mac trust/signing checks and native Linux verification remain release gates.
