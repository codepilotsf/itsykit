# Installation and updates

Version 0.13.0 is the current test release for Apple Silicon macOS and AMD64/ARM64
Linux. Published October 12, 2026; the public installer selects it from
`https://fleacms.com/releases/latest.txt`. The shared server remains 0.9.0-rc.1 /
protocol 23; this release needs no remote server upgrade.

## 0.14.0 server and local CLI rollout

0.14.0 removes browser Theme editing and CSS synchronization while retaining
`site/css/theme.css`, Prewind and source styling. Normal deployment publishes
committed CSS. Snip editing and synchronization remain supported. Protocol 24
requires reviewed removal of obsolete host theme mappings before upgrading an
existing protocol-23 server; preserve any remote customizations in source first.
The public update channel remains 0.13.0 until a separately verified publication.
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
