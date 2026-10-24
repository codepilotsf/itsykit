# FleaCMS 0.19.0 release verification

Published and rolled out October 24, 2026. The public channel, shared server and
normal Kirin/Mac CLIs report 0.19.0. Deployment protocol is 27; app protocol is 2.

## Source and native checks

- Reviewed release source: `da7d0c70775dea352e00cfc6cf2c86e31f9cea16`, immutable
  tag `v0.19.0`, merged as `a6bab4e` through framework MR !18.
- Reconciles indexing implementation `3b7d21d` with published 0.18.0/current main
  `0002a84`. Existing-empty-directory creation and the two supported CLI targets
  are preserved; unrelated local Kirin skill-install changes are excluded.
- Exact source archive SHA-256 on both build machines:
  `1e7d0926d1c06edd80592eecf22503321fb67fac7c48a19bdca32a261f05a906`.
- Native Linux AMD64, Go 1.26.9: canonical bundled build, then
  `FLEA_REQUIRE_BUNDLE=1 CI_COMMIT_TAG=v0.19.0 scripts/release.sh server amd64`.
  Full tests, full race checks and vet passed with real Contact/server bundles.
  ARM64 server was cross-compiled on this machine.
- Native macOS ARM64, Go 1.27.1: canonical bundled build, full tests, full race
  checks and vet passed with both bundle-required flags. Final Mac CLI embeds
  the same two Linux server artifacts built on AMD64. Ad-hoc codesign and strict
  signature verification passed after the final parent build.
- Exact final CLIs passed current-directory creation, initial Git commit,
  minimal/starter file parity, status and nonempty-directory refusal checks.
  The exact Linux server artifact also passed a local two-host/shared-root test
  for GET/HEAD, assets, 404s, health and different indexing policies.
- Tag pipeline 610 was pending native runners at publication. The release used
  the manual native checks above; no remote CI success is claimed.

## Public artifacts and updates

Anonymous HTTPS downloads match the immutable public manifest exactly:

```text
e68848a053a8cec0d31313554a1dcd12570ef1fff9f6f5bdb4a2b04982d842ec  flea-darwin-arm64
004c36487647c1967faa7ecf1afbd977aab0f39e9af703636d3d6e25e40c46bd  flea-linux-amd64
```

The version directory was staged on disk, checked for ownership/modes and
checksums, moved without overwrite under deploy.lock, and validated before the
supported executable upgrade restarted the service and loaded the catalogue.
The previous 0.18.0 manifest is byte-identical to its retained release record.
No old downloads, releases, content or media were pruned.

On both supported native platforms, public fresh install, reinstall, update
from 0.18.0 and already-current checks passed in separate directories, including
paths with spaces, with matching installed hashes and exact CLI smoke checks.
The public installer is byte-identical to canonical source. Normal user CLIs
were then updated through `flea update`, with their prior bytes retained:
Kirin resolves `/home/sean/.local/bin/flea` to the preserved project `bin/flea`
symlink target; Mac uses `/Users/sean/.local/bin/flea`. Both final hashes match
above. Existing dev processes were not restarted.

## Server and preservation

The supported `flea update-server` completed against root@fleacms.com. An initial
nonterminal confirmation was refused before executable mutation; the ordinary
interactive confirmation then completed through a PTY. No recovery was needed.
Installed Linux server SHA-256:
`b1b70b0012682158c6c97ee94ae68099072793ea9e5e18d389b806c623d53661`.

Before and after the executable upgrade, all 15 release roots, 1,339 source-file
hashes, every live content hash and all host-policy fields matched. Only
`staging.codepilot.com` was subsequently set to `noindex` through the supported
SSH command. The existing 320 MiB memory target and active swap remain intact;
no pending deployment or executable recovery journal remains.

All 21 configured HTTPS hostnames have valid TLS and their expected 200/308
status. After both website deployments, all live content hashes and all 13
unaffected release roots/source hashes still match the baseline. The two
intentional redeployments changed exactly these source files:

- Flea website: docs.html, _docs-toc.html, index.html and releases/latest.txt.
- CodePilot staging: .flea.config.json and the public/admin/500 layout files,
  removing only the temporary indexing flag/meta blocks.

Routing, redirects, TLS, download mappings and site/branch ownership are unchanged.

## Website and consumer rollout

Public website MR !7 merged to `e8ee4aa`, deployed as
`20261024-204456-c663ba719e4e604fe65116f6`. The channel selects 0.19.0; README,
installation text, release notes and download links are updated. The live browser
confirmed the Search indexing section/contents link, command examples, download
URLs and no horizontal overflow at 1280px. Historical release notes are retained.

Native staging headers were verified while its meta fallback was still present.
Only afterward, consumer commit `efd26efa187605701554a0189e17bd276efeb017` was
published on staging and deployed as
`20261024-204619-be34ce7d4a930362a4fa8147`. Final live GET/HEAD checks prove native
noindex/nofollow on the homepage, asset, admin, missing page and robots.txt, plus
the HTTP redirect. Health retains its ordinary contract. No robots meta fallback
remains. Homepage content other than that removed meta tag, all four signup URLs
and robots.txt bytes are unchanged. A spoofed forwarding header cannot add the
staging policy to fleacms.com. The policy survived both ordinary deployments.

The consumer production guard now requires an explicit HTTPS origin and checks
local signup/crawl configuration plus live headers/meta/robots.txt. Regression
fixtures cover production success, native staging rejection, repeated headers,
meta/crawl blocks, redirects, unavailable hosts, invalid URLs and bounded reads.
Its live staging run intentionally fails on native response headers.

No codepilot.com registration, DNS changes, Hub code or billing changes occurred.
Signup remains https://hub.codepilot.com/get-started. Future rollback must retain
staging protection as described in [host indexing](../host-indexing.md).

## Backups and remaining limits

Private evidence/artifacts are under
`/home/sean/fleacms-releases/host-indexing-rollout` and the corresponding
`/Users/sean/fleacms-releases/host-indexing-rollout` on Mac. Site backups for both
deployed sites are retained there. Prior remote config/executable are retained
in `/var/backups/flea-host-indexing.D6ChkUu8Kj` (private, root-owned); supported
upgrade recovery also retains its previous binary. These are private operator
records, not public website assets or portable site backups.

Native ARM64 Linux server execution, clean-Mac acceptance, Apple notarization and
real email delivery were not tested for this release. Source browser cases that
require FLEA_TEST_BROWSER were not run. Live browser checks were limited to the
public documentation. No task-owned dev server remains running.
