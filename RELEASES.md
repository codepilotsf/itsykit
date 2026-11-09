# Publishing FleaCMS downloads

## 0.20.5 published November 9, 2026

Adopt the approved navy-on-light flea treatment throughout the primary logos,
icons, website header/footer/admin wrappers, sharing artwork, starter and
framework diagnostic branding. Preserve the hero layout, mascot geometry,
Nunito lettering and customer branding. Sticky-header checks include 320px
navigation spacing and background blur while scrolling.

Framework MR !36 merged as `5084a26`; immutable `v0.20.5` selects tested source
`00b2e36`. Source archive SHA-256 on Linux/Mac is
`55909f2680b71aba8485609e7ce3ad9c0bdec077fc68eb4358b574a98917fbb2`.
Website MR !22 merged as `53eb18e` and activated
`20261109-165104-562591a96cda87412e103bef`, including the update channel.
All live logo/icon/social/CSS bytes match reviewed source. Desktop/mobile/
320px homepage, docs and admin branding checks passed.

Native Linux/Mac full tests, race tests, vet, canonical builds, exact-artifact/
starter checks, signature checks and public install/reinstall/update checks passed.
Normal Kirin/Mac CLIs match the public manifest:

```text
c3870e6d1605da19674fd8ea84a5d795fa9b5703b8a9118f57e31b59f3a7ef34  flea-darwin-arm64
6be769f644468f56af9189432facb75b3fb896de463dbf73827357b808f26dbb  flea-linux-amd64
```

The compatible shared executable stays 0.20.3. Live content, host policy,
other 19 sites and all 29 HTTPS responses/indexing are preserved. Old downloads
are unchanged. Private evidence is in `fleacms-releases/0.20.5` on Kirin and Mac.
Tag pipeline 799 was pending runners; completed native checks supplied evidence.
Existing ARM64 server, clean-Mac/notarization, packaging and email limits remain.
Task development preview stopped; no unrelated source work was overwritten.

## 0.20.4 published November 9, 2026

Browser and touch icons now stay white `#f5f9ff` on header navy `#142b43` in both
browser themes. SVG, PNG, ICO and touch files match the reviewed exports and use
refreshed URLs. Logo artwork and page design are unchanged. Framework MR !34
merged as `f6f5430`; immutable `v0.20.4` selects tested source `f0269b2`.
Website icon MR !19 merged as `59f7ced`, activation
`20261109-012644-c58945602a2360256d46434d`; channel MR !20 merged as `6694e69`,
activation `20261109-013053-45656596029f6bb3f0a161bd`.

Native Linux/Mac full tests, race tests, vet, canonical builds, exact/starter
checks and public install/reinstall/update/already-current checks passed. Normal
CLIs match the public manifest. The compatible shared executable remains 0.20.3;
all 29 HTTPS hosts, live content, host policies and other 19 sites are preserved.

```text
38bec9013fa3c224648b2c84dc63896b5982042efe6d43eecf1420c3dbea8760  flea-darwin-arm64
05698c04913aca2839ab657a6153ac24dc8410ee41f2bab88786647eea2d256c  flea-linux-amd64
```

Tag pipeline 798 was pending runners; completed native checks supplied evidence.
Existing ARM64 server, clean-Mac/notarization, packaging and email limits remain.
Logs, previous CLIs and source provenance are in `fleacms-releases/0.20.4` on
Kirin and Mac. Older downloads and customer branding are unchanged.

## 0.20.3 published November 9, 2026

The approved filled Little Skip concept D replaces the outline throughout the
website, starter, framework website and development diagnostic. Nunito lettering,
lockup dimensions and existing palettes are preserved. SVG, PNG, ICO, touch icons
and the social image derive from the same artwork; all eight public assets match
reviewed source. Customer branding is preserved.

Framework MR !32 merged as `1071144`; immutable `v0.20.3` selects tested source
`26bc0e8`. Website branding MR !16 merged as `6bce343`, activation
`20261109-004926-199b26ab9e0c6aff256e0ac1`. Channel MR !17 merged as `577a722`,
activation `20261109-005527-66744fb1ea339e9e65840e66`.

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) canonical builds, full
tests, race tests, vet, exact-artifact/starter checks and public fresh install,
reinstall, 0.20.2 update and already-current checks passed. Mac ad-hoc signature
verification passed. Source archive SHA-256 on both machines is
`dc3c85513b677c8a056bad20daa5e133774fbfe2be13102de2a5cd8745d0bbee`.

```text
7eb2c7e094ffca80e569261a7657633b81eeda4bf86cc3f51a831dded9bbae33  flea-darwin-arm64
be70a532a2a86e5be2dd6d4c270d3c45a89a4b0eda425a10791126ac5f1b8308  flea-linux-amd64
```

The shared server, public channel and normal Kirin/Mac CLIs report 0.20.3.
Server SHA-256 is
`4ba946401b91ceb723f30b73b8c44803f8e6e306e5b31d60bae1ef1324ef7f75`.
All 29 HTTPS hostnames retain expected status, redirects and indexing headers.
Live content and host policies are preserved; a concurrent Cubitalk deployment
advanced its own code pointer and was left intact. Other site sources match the
baseline. The first website activation hit a busy lock without changing the live
site; retry succeeded after confirming no pending recovery. Old public downloads
are unchanged. No host policy, TLS, access or customer branding was changed.

Tag pipeline 797 was pending runners; completed native manual checks supplied
publication evidence. Native ARM64 server, clean-Mac/notarization, installer
packaging and real email limits remain. Private evidence and previous CLIs are
in `fleacms-releases/0.20.3` on Kirin and Mac; the remote backup is
`/var/backups/flea-0203.tC2FUbNf`. Temporary previews stopped.

## 0.20.2 published November 2, 2026

Framework MR !29 merged as `f00132e`; immutable `v0.20.2` selects tested source
`279ec50`. Website MR !14 merged as `648e922` and activated normally as
`20261102-203411-64cea67e0d016a92327e5c39`. The owner approved the heavier
Little Skip mascot, Nunito 800 lettering and intermediate spacing. The website
header/footer, docs/admin wrappers, social preview and SVG/PNG/ICO/touch favicons
now use the matching outlined artwork. Current release links/channel are 0.20.2.

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) passed full tests, race
tests, vet, canonical bundled builds and release checks. Exact/public clients
passed fresh-site creation, initial commit, starter parity, status and overwrite
refusal. Anonymous fresh install, reinstall, 0.20.1 update and already-current
checks passed on both platforms. Mac ad-hoc signature verification passed.

```text
015c092a3622335f3e78a9cee364d8b9bef2ad907875ec0c7ebc1bb6f87b4bbc  flea-darwin-arm64
6ccfa190ac5db9165db3a9327e5195e3f17727484239dc3bd4c9b02a488b0e7a  flea-linux-amd64
```

Shared server, public channel and normal Kirin/Mac CLIs report 0.20.2. The server
hash is `83af63cae6cc3b5fe6c89f10143f88b255d898b7c8806eb8dcf9804ec986d840`.
All 24 HTTPS names retained expected statuses, redirects and indexing headers.
Live CMS content and all other 17 sites' source hashes are preserved; existing
customer logos and an active Dropr development job were not overwritten.
Kirin's CLI symlink is preserved. Asahi remains decommissioned. No old downloads,
retained content or media were removed. Memory/swap policy, TLS, quotas, mounts
and cleanup configuration are unchanged. Only task-owned scratch is cleaned up.

All eight public logo/favicon/social files match the reviewed source; desktop
and mobile website/docs/admin checks show loaded logos and no overflow.
The source archive SHA-256 matches on both platforms:
`0991e4848a4e5c6ee6f31848e76d0ef530b097e8e470257bd740f2552f698513`.
Private evidence is under `fleacms-releases/0.20.2` in Sean's home on both
machines; the server backup is `/var/backups/flea-0202.BArcPnY2`.

Tag pipeline 709 was pending runners; successful native manual checks supplied
publication evidence. Native ARM64 server execution, clean-Mac acceptance,
notarization, native installer packaging and real email remain unverified.
No Rails, payment or native app distribution work was performed.

## 0.20.1 published November 2, 2026

Immutable framework `v0.20.1` selects `c14efd7`, merged through MR !27 as
`cba671b`. Deployment protocol stays 28; app protocol stays 2. Packaging respects
TMPDIR, storage errors explain recovery, and the release client removes its own
smoke directory. Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) full
tests, full race checks, vet, canonical builds and release scripts passed. Exact
artifacts passed creation/starter/status and host-isolation smoke checks; the Mac
signature passed. Tag pipeline 704 was pending runners, not reported as passed.

Final anonymous HTTPS download hashes:

```text
981e4c1d6cd2a6c5c50ee36d731bf2bac773b011d52c81585562c3b0dce06968  flea-darwin-arm64
8c0894c23f2f5298ad01da92f88af4d49b0b837ca021bdc789bba7df480a70b9  flea-linux-amd64
```

The new immutable catalogue was staged on disk, verified and moved without
overwrite under deploy.lock. Prior 0.20.0 hashes are unchanged. Supported
`flea update-server` installed 0.20.1, SHA256
`b0ed25ede09f614eb35f2c167462a08a84090236af065cdd7043d52e847631c4`.
All 18 source trees, live content and byte-identical server config were preserved
across the upgrade. Existing memory/swap policy remains; recovery is clear.

Website MR !12 merged as `619f8a2`, activated normally as
`20261102-185436-78c8e68d3d9d74d6f432dd9a`. Only index.html, docs.html and
releases/latest.txt changed in deployed source. All live content and the other
17 sites' roots/source hashes are unchanged. All 24 HTTPS hosts retained expected
status, redirects and indexing headers before and after publication. Live browser
checks at 1280px/390px verified release text and the TMPDIR example without
horizontal overflow, missing images or observed console errors.

Anonymous public install/reinstall, 0.20.0 update, already-current behavior and
CLI smoke passed on both native platforms. Normal Kirin/Mac CLIs were updated
through `flea update`, preserving prior binaries and Kirin's symlink. Kirin local
site workflows share that CLI; no separate project executable or running Flea
dev process was found. Asahi is decommissioned and excluded from future rollouts;
references in older receipts below are historical, not active deployment targets.

Evidence and artifacts: `fleacms-releases/0.20.1` in Sean's home on Kirin and Mac.
The Linux archive includes a site backup; root-only remote config/executable
backups are in `/var/backups/flea-0201.uVhW4aef`. No retained downloads, site
releases, content or media were pruned. Native ARM64 server execution, clean-Mac,
notarization, native installer packaging and real email delivery remain untested.
No Rails, payments, native distribution, quota, mount or cleanup policy changes.

## 0.20.0 published October 30, 2026

Immutable framework tag `v0.20.0` selects
`f5f47d1251c25da2271b8dc3bf353313ed122968`, merged as `14dde06` through MR !22.
This release combines 301 migration redirects and URL-inventory guidance with
opt-in sitemap/robots discovery, missing canonicals and exact noindex exclusions.
Deployment protocol is 28; app protocol remains 2. Existing sites are not opted
into SEO automatically.

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) full tests, full race,
vet and bundled builds passed on the same source archive. Both CLIs embed the
same final Linux servers. Exact CLI checks covered minimal/starter parity,
initial commit/status and real redirect, query, sitemap, robots, canonical and
exclusion behavior. Final Mac ad-hoc signature verification passed. Tag pipeline
658 was pending runners; manual native checks supplied release evidence.

Anonymous HTTPS downloads and public fresh install, reinstall, 0.19.0 update and
already-current checks passed on both native platforms with these final hashes:

```text
f9820e1ca70976f2489b4149839173052e1cd160cf0c074a83f8a98229d05826  flea-darwin-arm64
2a9411fd9d243c58be201dcd10d048a8724fd855bb1ce9550f5139e3f31bf1f9  flea-linux-amd64
```

The supported server update completed at 0.20.0 / protocol 28, with no recovery.
Website MR !9 merged as `9190acb`, activation
`20261030-004255-1885fadf131e62d999b94551`. Homepage, docs, download links and
latest.txt select 0.20.0. README explains the new settings and migration skill;
public docs include desktop/mobile contents links and configuration examples.
The installer is unchanged. Normal Kirin/Mac CLIs match the public hashes, with
prior executables retained.

All 23 HTTPS hostnames are healthy, every live content hash and all 16 unaffected
site roots/source hashes are preserved, and host indexing/TLS/redirect/download
policies are unchanged. The website changed only docs.html, _docs-toc.html,
index.html and releases/latest.txt in deployed source. Existing memory/swap
safeguards remain intact. Older downloads, including 0.19.0, are unchanged.

Private evidence is retained under `fleacms-releases/0.20.0-candidate2` on Kirin
and Mac; the superseded, unpublished redirect-only candidate and site backup are
retained in adjacent `0.20.0`. Native ARM64 server execution, clean-Mac acceptance,
notarization, real email and hosted crawler/Search Console acceptance remain
unverified. Task-owned dev servers and temporary transport were stopped.

## 0.19.0 published October 24, 2026

Framework source `da7d0c70775dea352e00cfc6cf2c86e31f9cea16`, immutable `v0.19.0`,
merged as `a6bab4e` through MR !18. Adds native exact-host indexing policy,
protocol 27 and matching client/server enforcement; app protocol remains 2.
Both supported CLIs retain the same AMD64/cross-compiled ARM64 Linux servers.

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) full tests, full race,
vet and bundled builds pass. Exact CLIs passed new/current-directory, initial
commit, starter parity, status and refusal checks; the exact Linux server also
passed shared-root hostname/header isolation checks. Mac final ad-hoc signature
verification passes. Tag pipeline 610 was pending runners at publication; the
release used these manual native checks and does not claim remote CI success.

Anonymous HTTPS downloads match the final signed-byte manifest:

```text
e68848a053a8cec0d31313554a1dcd12570ef1fff9f6f5bdb4a2b04982d842ec  flea-darwin-arm64
004c36487647c1967faa7ecf1afbd977aab0f39e9af703636d3d6e25e40c46bd  flea-linux-amd64
```

The public installer matches canonical source. Public fresh install, reinstall,
0.18.0-to-0.19.0 update and already-current checks pass on both supported native
platforms, including install paths with spaces and matching final hashes. Normal
Kirin/Mac CLIs were updated through `flea update`, with prior binaries retained.
The old 0.18.0 manifest is unchanged; all older downloads remain available.

Supported `flea update-server` completed. The shared server is 0.19.0 / protocol
27. All 15 roots, 1,339 source hashes and every live content hash were unchanged
across upgrade and policy activation. Only staging.codepilot.com was set to
noindex. Native headers were verified before separately removing the consumer
flag/meta fallback. They remain on live GET/HEAD page, asset, admin, 404 and
robots.txt responses and the HTTP redirect, with health unchanged. The consumer
homepage content, signup URLs and robots.txt bytes are preserved. No production
hostname registration, DNS, Hub code or billing changes occurred.

Website MR !7 merged as `e8ee4aa`, deployed as
`20261024-204456-c663ba719e4e604fe65116f6`. Channel, homepage, installation/docs,
release notes and links now show 0.19.0. The live browser confirmed the new
[Search indexing section](https://fleacms.com/docs#indexing), contents link,
commands/download URLs and no horizontal overflow at 1280px. The consumer
`efd26ef` deployed as `20261024-204619-be34ce7d4a930362a4fa8147`. After both normal
deploys, every live content hash and all 13 unrelated roots/source hashes still
match. All 21 HTTPS hosts have valid TLS and expected 200/308 responses. Memory
and swap safeguards are unchanged; no recovery journal remains.

Private evidence/artifacts and site backups are under
`/home/sean/fleacms-releases/host-indexing-rollout`, with the Mac counterpart under
`/Users/sean/fleacms-releases/host-indexing-rollout`. Remote pre-upgrade config and
binary are retained privately at `/var/backups/flea-host-indexing.D6ChkUu8Kj`.
These receipt-only follow-ups do not change the deployed website snapshot.
Native ARM64 server execution, clean-Mac acceptance, notarization and real email
remain unverified. Source browser tests requiring FLEA_TEST_BROWSER were not run;
the live browser check covered public documentation. No task-owned dev server
remains running.

## 0.18.0 published October 21, 2026

Framework source `7ab8f6ea022f276600bc373d7353559c8552bc11`, immutable tag
`v0.18.0`, merged as `f02af1d` (MR !16). Adds `flea new .` and relative/absolute
paths into new or existing completely empty project directories. Hidden files
and `.git` count as content; nonempty destinations are refused without overwriting.
Existing folder-derived naming and initial Git commit behavior are preserved.

Linux ARM64 CLI downloads are discontinued. Supported CLIs are Apple Silicon
macOS and x86-64 Linux; both retain the cross-compiled ARM64 Linux server.
Native ARM64 server acceptance is recorded separately and was not run because
Asahi was offline. Published older binaries and hashes remain immutable.

Native Mac ARM64 (Go 1.27.1) and Linux AMD64 (Go 1.26.9) full tests, race, vet and
bundled builds pass. Both final CLIs embed the same native AMD64/cross-compiled
ARM64 Linux server binaries. Exact final binaries and public installed binaries
pass current-directory creation, initial commit, starter parity, status and
nonempty/hidden-file/symlink refusal checks. Mac ad-hoc signature verification
passes; clean-Mac acceptance and notarization remain unverified.

```text
e6b9a3dde2a7e832501d7df2bfe1acb6c9c7a2bf1789e88de14205631fc81a5f  flea-darwin-arm64
eb0c6659cd032c72a38fc1989252217d5629cdec66e10ee35189e3b7cbb5d6ae  flea-linux-amd64
```

The two-file public manifest and anonymous HTTPS downloads match those hashes.
The new installer is byte-identical to canonical. Fresh install, reinstall,
update from public 0.17.0 and already-current checks passed as the ordinary user
on both supported platforms, with matching installed hashes. Checks used
isolated installation directories; normal CLI installations were not replaced.

The version directory was validated and installed without overwrite under the
deployment lock. The existing shared server restarted to load its catalogue;
an immediate loopback probe raced startup, while subsequent readiness and all
public checks passed. Production remains **0.17.0 / protocol 26**; no executable
server upgrade is needed for this CLI feature. App wire protocol stays 2.
Memory/swap safeguards are intact and no recovery journal remains.

Website MR !5 merged as `1c786f2daad2c5489207eac840ed954654536655`, deployed as
`20261021-174934-4dbfcc67281f82d2e68b36a4`. `latest.txt` selects 0.18.0 and live
homepage/docs show the current platform matrix and empty-directory instructions.
Website status reports matching deployed code/content from that commit. All 19
configured HTTPS hostnames return 200, host policy and non-website release roots
are unchanged, and every editable-content hash is preserved.

The previous 0.17.0 manifest and ARM64 CLI still resolve; the old ARM64 binary
still matches `8c0ac734abf2dcbb2e9aac6e808a74c83040c21d067c2a3c1906c7fe0b96b09d`.
Task-owned preview was stopped. Native ARM64 server execution, clean-Mac,
notarization and real email remain unverified. Tag CI pipeline 481 was pending
runners at publication; this release used the recorded manual native checks.
Private evidence is under `/home/sean/fleacms-releases/0.18.0` and
`/Users/sean/fleacms-releases/0.18.0`. These receipt-only follow-up changes do not
alter the deployed website snapshot.

## 0.17.0 release October 18, 2026

Framework source `fe15b8934165a90390c8bd4ff88cc5e9c9e0d753`, immutable tag
`v0.17.0`, adds managed custom apps, app-owned sessions and modular starter
skills. Deployment protocol is 26; app wire protocol is 2. The literal-brace fix
from 0.16.1 is included without changing Impeccable.

Final artifacts are archived privately in `/Users/sean/fleacms-releases/0.17.0`.
Mac ARM64 and Linux ARM64 full tests, vet, race checks and canonical bundled
builds passed. Linux AMD64 was cross-built on Mac and passed an unprivileged
native CLI smoke check. The exact server and sample app passed TLS, sessions,
restart persistence, CSRF/Origin rejection and shutdown under a transient
systemd unit with production hardening and a separate network namespace. The
transient run used the exact artifact; it is not a claim that a customer custom
app is deployed in production.

The version directory was published without overwriting older versions. All
three anonymous HTTPS downloads match SHA256SUMS and the older 0.16.1 manifest
still resolves. The supported shared-server upgrade completed as 0.17.0 /
protocol 26; all 18 HTTPS hostnames returned 200. Routing, editable-content
hashes and deployed source hashes were unchanged. The existing memory target,
swap and service policy remain in place, and no recovery journal remains.

Website source `115a73c`, activation
`20261018-195731-803d4f4dc7db80fad2a5fd8a`. This website commit advances the
channel and current documentation; `latest.txt` selects 0.17.0 and public docs,
custom-app availability and all download links were checked in a browser. Final
website verification confirmed the expected global revision, unchanged content
hashes after deploy, and unchanged non-Flea sources, routing, commits, TLS,
downloads, site branches and redirect policy. A full post-deploy check returned
200 for all 18 HTTPS hosts again, and a missing-CSRF production Contact POST
returned 403 with no email sent. The installed CLI `flea status` passed all 8
known Mac site folders, including remote comparisons.

The public installer is byte-identical to canonical and passed fresh install,
reinstall, update from public 0.16.1 and already-current checks on Mac ARM64,
native Asahi ARM64 and native AMD64 as `nobody`; every installed hash matched the
published manifest. Normal Mac `/Users/sean/.local/bin/flea` and Asahi
`/home/sean/.local/bin/flea` were updated through the public updater and report
0.17.0 with matching hashes; previous binaries were retained privately. Two Asahi
previews were gracefully replaced preserving cwd and flags, and their
`/proc/<pid>/exe` hashes match the release.

```text
5f9cb8d9bda64d6f00f52eb43e1d38b82cb9d28832f30df5e3b4bbc71973dec5  flea-darwin-arm64
919ed61f7ac31f5ba3163a16b2bce05d1fcd7fcbce8c82c2a667c15a1f16cad4  flea-linux-amd64
8c0ac734abf2dcbb2e9aac6e808a74c83040c21d067c2a3c1906c7fe0b96b09d  flea-linux-arm64
```

Native AMD64 full-suite/build, clean-Mac acceptance, notarization and real email
remain unverified. The Mac binary is ad-hoc signed.

## 0.16.1 release October 18, 2026

Framework source `57d8378df75c9981bb5fe1af173dbad81ac8afb4`, immutable tag
`v0.16.1`, based on published 0.16.0. This patch fixes literal brace groups and
raw JSON in template text/attributes, with matching rich/media/multisnip scanning.
Impeccable code, metadata and workflow are unchanged. Unreleased modular skills
on framework main and the original older dirty checkout are excluded/preserved.

Mac ARM64 and Linux ARM64 full tests, vet, full race checks and bundled builds
passed with real server/Contact bundles. Native new/status/raw-JSON/HTTP/dev-reload
smoke passed on both and on Linux AMD64 as nobody. AMD64 was cross-compiled on
Mac; its native build/full suite remains open. Mac signature validation passed;
clean-machine Mac/notarization remain open. All artifacts use Go 1.27.1 and the
same Mac-built embedded Linux servers. The ARM64 CLI was built natively from a
private archive of the exact tagged source; Mac/AMD64 VCS metadata is clean.

The unchanged Impeccable helper passed initial raw JSON parsing, development
reload, real Tune slider changes (0.5 to 0.8, padding 10px to 16px), and discard
cleanup in a disposable site using the exact Mac release binary. Temporary dev
and helper processes were stopped.

Published HTTPS binaries and manifest match the local immutable archive:

```text
8191404446e6025408b0e23f7bf84b740bc2d475a22164418b015635deee2390  flea-darwin-arm64
d3162dab92e5e3376cb93aaa4f0aad0b8fc2d36ddb320d45424df02d80fe8199  flea-linux-amd64
37443e9b9c5e24f58fece7d2185d2d38bccad6e1e62eb190d121863f53de4bc5  flea-linux-arm64
```

The shared server is verified at 0.16.1 / protocol 25 after supported update-server.
Service is active, recovery is clear, all 18 configured HTTPS hosts return 200,
and routing/content hashes match the pre-upgrade baseline. Existing memory/swap
policy is unchanged. Website backup is private at
`/Users/sean/fleacms-backups/fleacms-before-0.16.1-20261018.tar.gz`.
Website source `d59d74e`, activation
`20261018-174259-d8975e4a784b2f5837ed6ef4`. Public homepage, docs, download links
and latest.txt agree on 0.16.1. Public fresh installs, reinstalls, updates from
0.16.0 and already-current checks passed on Mac ARM64, Linux ARM64 and Linux
AMD64 as nobody; installed hashes match the published manifest. Normal Mac
`/Users/sean/.local/bin/flea` and Linux `/home/sean/.local/bin/flea` were updated
through the public updater and report 0.16.1 with matching hashes. The shared
server also reports 0.16.1, service active and recovery clear.

Only fleacms.com and www.fleacms.com moved release roots; every hosted content
hash, TLS policy and download mapping matches the baseline. No real email,
new Ubuntu/systemd installation, clean-machine Mac or notarization acceptance
was performed. Native AMD64 build/full-suite remains open. Original older
framework worktree and unrelated website design work remain untouched.


## 0.16.0 published and installed October 13, 2026

Framework/tag `1c85aae` / `v0.16.0`, clean Go 1.27.1 artifacts. Includes detected
hostnames/local aliases and Tailscale IPv4 LAN peers with exact Host/port checks,
automatic default dev-port selection, and server-update progress. Website source
`a49e002`, activation `20261013-221944-d1f6169314346e36fd5ab5c7`. Public homepage,
docs, download links and latest.txt now agree on 0.16.0.

Native Mac and Linux ARM64 full runtime tests, vet, race and bundled builds
passed. Relevant named-host editing and mail-preview Chrome checks passed under
race instrumentation. A test-only preview handler race was fixed. The existing
owner-navigation browser check reports `Missing page transition name` in repeated
race runs; repeated ordinary runs pass. No navigation implementation/assertion
was changed. Do not claim all browser tests under race passed. Native AMD64
build/full-suite, clean-Mac/notarization and native CI acceptance remain open.

AMD64 was cross-compiled on the Mac and smoke-tested natively as nobody. All
CLIs embed identical Linux servers. Public hashes, fresh installs, reinstalls,
0.15.1-to-0.16.0 updates, already-current checks and fresh-site/dev smoke passed
on all three advertised platforms. Native Mac signature checks passed. Normal
Linux `/home/sean/.local/bin/flea` and M5 `/Users/sean/.local/bin/flea` were updated
through the public updater and both report 0.16.0 with matching public hashes.

The immutable directory was verified and installed under the deployment lock,
validated with the existing 320 MiB admin target, then the service restarted to
load the catalogue. Public bytes and the old 0.15.1 manifest passed before channel
advancement. During the longer build window, Gestalt Haus was separately deployed
as `6601d165`, activation `20261013-220136-62ba8cfc530ca453040636ab`; this newer
code/content/CSS was preserved. Config last changed at 22:01:41 UTC, before
catalogue restart at 22:15:55 UTC. A refreshed pre-channel snapshot matched
content/media/CSS and unrelated deployments after the website update.

All 12 HTTPS domains are healthy. Final website status reports code/content
matching source. Recovery is clear and the publication-window error journal is
empty; memory/swap policy and retained downloads are unchanged. Server executable
remains 0.15.1 / protocol 25. No source branches were merged or real email sent.
Actual M5-to-`asahi:3003` dev homepage/sign-in/reload returned 200; unrelated Host
and wrong port returned 403. No firewall settings or existing dev servers changed.
Restart site dev with the installed CLI to load the new code.

- Mac ARM64: `c95ba159de8674538ebf03499958332ff66ada9d9cb36b63394af4696e6b3ee3`
- Linux AMD64: `edae40e02d8e6174610d46277958313790eee26e786f86b8527b2fd7b0324b2c`
- Linux ARM64: `9a88c0e21b0fe890f81c72c8b71d514a1e126f211465d227a1e41e4189a7f53c`

Framework [draft MR 8](https://gitlab.codepilot.com/sean-schertell/flea/-/merge_requests/8)
and website [draft MR 4](https://gitlab.codepilot.com/sean-schertell/flea-website/-/merge_requests/4)
are linked to T3. Full provenance/limits are in framework
`docs/agents/release-016-verification.md`. Private evidence/website backup:
`/tmp/flea-016-release`. Retained artifacts:
`/home/sean/fleacms-releases/0.16.0`; remote staging:
`/var/lib/flea/downloads/.flea-016.VWjWihow`. Prior installed binaries remain at
`/tmp/flea-016-release/linux-cli.before` and `/tmp/flea-016-installed-before/flea`
on the Mac. Task-owned previews stopped; existing manual QA answers are preserved.

## 0.15.1 published October 13, 2026

Framework source/tag `d5958ac` / `v0.15.1`; all three clean tagged artifacts use
Go 1.27.1 (`vcs.modified=false`). Includes optional Blog, Blog/Contact setup
skills and the removal of browser Theme editing/CSS synchronization. The public
installer, updater, homepage/docs and download links now select 0.15.1.
Website source `d6cd9c9`, activation
`20261013-213030-9a92d56ccb61b7b62b20a800`, deployed through an isolated clean
main clone. This task merged no source branches.

Native Mac ARM64 and Linux ARM64 full tests, required real-bundle race checks,
vet and bundled builds passed. Final CLIs embed identical Linux server bundles.
Linux AMD64 was cross-compiled on the Mac and tested natively as nobody; its
native build/full-suite remains unverified. Clean/starter initial commits,
status, dev routes and new skills passed. Generated Blog templates/skills matched
canonical starter bytes. Mac ad-hoc signature checks passed; clean-Mac acceptance
and notarization remain open.

The immutable download directory was uploaded beside the store, hash/permission
checked, moved without overwrite under the deployment lock, validated, then the
service restarted to load its catalogue. Anonymous HTTPS hashes on all three
platforms and the retained 0.13.0 manifest passed before advancing latest.txt.
Public fresh installs, reinstalls, 0.13.0-to-0.15.1 updates and already-current
checks passed on all three native platforms, with installed hashes matching.
Downloaded binaries passed fresh-site/dev checks; test dev servers stopped.
Ordinary installed CLIs were left unchanged so owners can run `flea update`.

Server executable remains the already-installed 0.15.1 / protocol 25. All eleven
HTTPS domains return 200; content/media/CSS and unrelated site roots/commits are
preserved. Host config fingerprints matched across catalogue publication; only
the website's normal deployment changed its routing/commit receipt afterward.
Service is active, recovery is clear and the deployment-window error journal is
empty. Memory/swap policy is unchanged and no old downloads were removed.
Final website `flea status` reports deployed code/content matching source.
No real email or new interactive production editing was tested.

- Mac ARM64: `784ca438a304faf344f1a2a6bbb4e2fd5685ced77411b5fc5c313a581b94c9f4`
- Linux AMD64: `6c31975387b3712967066cd78533ef0aa6bdd92f27fba53860bb7a6ba36451f3`
- Linux ARM64: `ee9d4ede230d14f0d3db74983148c58d526d3a576021a6ac0b37625d9b4e66a4`

Evidence/private website backup: `/tmp/flea-0151-release`. Local retained artifacts:
`/home/sean/fleacms-releases/0.15.1`. Remote retained staging:
`/var/lib/flea/downloads/.flea-0151.FVF6MDwf`; Mac public test files:
`/tmp/flea-0151-public`; native AMD64 test files:
`/var/tmp/flea-0151-native.T0IR9yVs`. Full provenance is recorded in framework
`docs/agents/release-0151-verification.md`. Native AMD64 full-suite/build,
clean-Mac/notarization and native CI runner acceptance remain open.

## Blog and Contact skills documentation deployment - October 13, 2026

Website source `5380c92` on `docs/app-skills`, stacked on `docs/blog-setup`.
[Draft MR 3](https://gitlab.codepilot.com/sean-schertell/flea-website/-/merge_requests/3)
remains open. The exact snapshot was normally deployed through a clean isolated
main checkout to the existing fleacms.com mapping; no source merge occurred.
Activation: `20261013-201235-eb349a3b66e33aaf2fadb520`.

Live https://fleacms.com/docs#agents documents `add-blog`, `add-contact-form`,
request phrases, setup conventions, template values and existing-site adoption.
The website's own agent guidance includes both skills. Source-built framework
`1feae24`, draft MR 7, supplies Flea 0.15.1 / protocol 25. The shared server and
local Linux ARM64 CLI were updated through the supported update-server workflow
and a verified atomic local installation. Public downloads remain 0.13.0;
no artifacts, tags or channel changes were published.

Verification: full native ARM64 framework tests with real bundles, vet and
standard bundled build passed; skill validation and clean/starter packaging
checks passed. Website `flea status` validates and reports live code/content
matching source. Browser checks at 1280px and 375px confirm literal tokens,
valid Blog/Contact links and no page/section overflow. All eleven HTTPS domains
return 200. Skill URLs return 404. Content/media/CSS fingerprints and site/domain
counts were preserved. Service is active, recovery is clear and the deployment
window has no service error entries. No real mail or new editor QA was performed;
native Mac/AMD64 release acceptance and public publication remain open.

All seven live-site backups and old server executable remain under private host
`/var/backups/flea-app-skills-0151-20261013`; website backup, previous local CLI and
before/after evidence are in `/tmp/flea-app-skills-rollout`. Existing memory/swap
policy was preserved; the established 320 MiB target applied to admin commands.
No task-owned preview remains. Receipt-only commits do not change the site snapshot.

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
(`linux-amd64`). Linux ARM64 CLI downloads are discontinued from 0.18.0;
both supported CLIs retain ARM64 Linux server deployment. Do not list additional
CLI platforms without native checks.
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
