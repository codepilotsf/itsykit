# Public 0.16.0 release verification

October 13, 2026. User authorized committing hostname work, local Linux/M5
installation, public publication and inclusion of the spinner/port changes.
One agent; unrelated main/Mac changes and the original hostname handoff were preserved.

## Source and scope

Source/tag `1c85aaedbf8699cf9c6504b4bfe14251d62ec4de` / `v0.16.0`.
Integration branch `fix/lan-hostnames-delivery`, worktree
`/home/sean/Projects/flea-lan-hostnames-delivery`, based on published 0.15.1
plus receipts (`caff6bc`). Spinner source `f3f5f6c` and port source `5686dbd`
were committed and integrated as `5f6660c` and `4267f27`. Hostname integration
is `bb7ce4a`; the release cut is `9cd262a`. Final `1c85aae` fixes only the preview
browser-test initialization race. The initial unpublished local tag was moved
before publication; the pushed/public tag is the final source and is immutable.

Original uncommitted hostname patch remains in `/home/sean/Projects/flea-lan-hostnames`
and private `/tmp/flea-lan-hostnames-handoff.patch`. Its unrestricted Host behavior
was narrowed to detected interface addresses and OS hostname/local aliases at the
actual listener port. Explicit LAN mode admits private, loopback and RFC 6598
shared IPv4 peers; forwarded headers remain irrelevant. Exact Origin, CSRF,
host-bound sessions and production HTTPS checks are preserved. Protocol remains 25.

[Draft framework MR 8](https://gitlab.codepilot.com/sean-schertell/flea/-/merge_requests/8)
targets the previous release branch; no main/source merges occurred. Website
[Draft MR 4](https://gitlab.codepilot.com/sean-schertell/flea-website/-/merge_requests/4)
contains the public copy/channel change and receipts. Both are linked to T3.

## Verification

Clean final tagged Go 1.27.1 artifacts (`vcs.modified=false`). Native Linux ARM64
and M5 Mac full `go test ./...`, `go vet ./...`, `go test -race ./...` and standard
`scripts/build.sh` passed with real app/server bundles required. Mac final full
runtime suites used the default browser opt-out. Relevant Chrome LAN editing and
mail-preview tests also passed under race instrumentation explicitly, including
named-host login/edit/save and private preview boundaries.

The first Mac ordinary suite with all browser tests enabled passed. Its race run
found a test-only mutable-handler race in `TestPreviewBrowser`, fixed by initializing
an unstarted test server before serving. It also found `TestOwnerNavigationBrowser`
reporting `Missing page transition name`; repeated race runs reproduce that timing
failure, while repeated ordinary runs pass. No navigation implementation or assertion
was changed. This existing unrelated browser check remains an acceptance limit;
do not describe all browser tests under race as passing.

All final CLIs embed the same Mac-built Linux servers. ARM64 was rebuilt after
selecting those bundles; required bundle/managed-app checks and final artifact
smoke tests passed. AMD64 was cross-compiled on the Mac and tested natively as
nobody on the production host; native AMD64 build/full-suite remains unverified.
Mac ad-hoc signature verification passed; clean-Mac acceptance/notarization remain open.

Default/starter initial commits, status and dev homepage/team/Contact routes passed
for actual artifacts on all platforms. Published fresh-site README/AGENTS and
bundled skills match canonical starter bytes. Port fallback, explicit-port errors,
replacement, cancellation, recovery, spinner/redirected output and diagnostics/sudo
boundaries are covered by regression checks and prior CLI terminal smoke evidence.

Actual M5 DNS resolved `asahi` to Linux's Tailscale IPv4. The initial random test
port was refused by the machine/network; no firewall settings were changed.
Using normal default-port fallback selected 3003 after occupied 3000/3001/3002.
M5-to-`http://asahi:3003` homepage, configured sign-in and reload returned 200;
unrelated Host and wrong port returned 403. Existing dev processes were untouched.
The smoke test used a disposable site with a fictional admin and sent no real mail.

## Publication and local installation

Final anonymous HTTPS and installed hashes:

- darwin-arm64: `c95ba159de8674538ebf03499958332ff66ada9d9cb36b63394af4696e6b3ee3`
- linux-amd64: `edae40e02d8e6174610d46277958313790eee26e786f86b8527b2fd7b0324b2c`
- linux-arm64: `9a88c0e21b0fe890f81c72c8b71d514a1e126f211465d227a1e41e4189a7f53c`

Staged beside the store at `/var/lib/flea/downloads/.flea-016.VWjWihow`.
Exact files, hashes and permissions passed. Installed without overwrite under
`/etc/flea/deploy.lock`, validated with the existing 320 MiB Go target, then
restarted to load the catalogue. All public hashes and the retained 0.15.1
manifest passed before channel advancement. No old artifacts were removed.

The long build-window snapshot comparison detected a separate Gestalt Haus
activation `20261013-220136-62ba8cfc530ca453040636ab`, commit `6601d165`, before
catalogue publication. Host config last changed at 22:01:41 UTC; catalogue service
restart was 22:15:55 UTC. That unrelated new code/content/CSS was preserved,
not rolled back. A fresh pre-channel baseline matched all content/media/CSS,
routing/branch maps and unrelated commits after the website deployment.

Website source `a49e002`, normal deployment activation
`20261013-221944-d1f6169314346e36fd5ab5c7`. Homepage, docs, downloads and latest.txt
agree on 0.16.0. Final `flea status` reports code/content matching source.
All 12 HTTPS domains are healthy; service is active, recovery journals are absent
and the publication-window error journal is empty. Memory/swap policy is unchanged.
Server executable remains 0.15.1 / protocol 25; it was not upgraded.

Public fresh installs, reinstalls, 0.15.1-to-0.16.0 updates, already-current checks
and fresh-site/dev smoke passed on Mac ARM64, Linux ARM64 and unprivileged Linux
AMD64. Normal Linux CLI `/home/sean/.local/bin/flea` and M5 CLI
`/Users/sean/.local/bin/flea` were then updated through the public updater and
both report 0.16.0 with the hashes above. Previous normal CLIs remain privately
at `/tmp/flea-016-release/linux-cli.before` and
`/tmp/flea-016-installed-before/flea` on the Mac. Existing dev servers require a
new invocation to load new Go code; none was restarted as part of installation.

Evidence and private website backup: `/tmp/flea-016-release` and
`/tmp/flea-016-*.log`. Local immutable archive:
`/home/sean/fleacms-releases/0.16.0`. Final Mac build clone:
`/Users/sean/Projects/flea-release-016-final`; Mac public tests:
`/tmp/flea-016-public`; AMD64 test files:
`/var/tmp/flea-016-native.CFWIObxb`. Task-owned dev processes stopped.
No real email, new Ubuntu/systemd setup or new interactive production admin QA
was performed; existing manual checklist answers are preserved. Native CI runner,
AMD64 full-suite/build, clean-Mac/notarization and navigation browser-race acceptance
remain open. This is a public test release.
