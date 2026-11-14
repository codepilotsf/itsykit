# ItsyWeb 0.21.0 rollout verification

Updated November 14, 2026. This is a partial rollout receipt, not a completed
public release. The owner authorized the rename, Kirin updates and all existing
hosted-site deployments. Public publication remains gated on native Mac checks.

## Source and Linux checks

Framework executable source: `04669d71d988b0425571834426f7a678723bb5ff` on `main`, pushed to the
private framework repository. Website `main` is `b251dfa`, including the approved
ItsyWeb naming/domain cutover and the compatible installer. No 0.21.0 release tag,
public version directory or channel advance has been published.

Native Linux AMD64 used `go version go1.26.9 linux/amd64`. The final candidate passed:

- `CI_COMMIT_TAG=v0.21.0 bash scripts/release.sh server amd64` and the matching
  `client amd64` script: full tests, full race tests, vet and native builds.
- Both runs enabled the documented `FLEA_TEST_BROWSER` workflow with Kirin's
  installed Chromium headless shell. Editor, media, collection, navigation,
  development and diagnostic browser checks passed on the final candidate.
- `bash scripts/build.sh`, real Contact bundle checks and verified extraction of
  both embedded Linux servers. The client script generated a fresh temporary
  site and passed `itsy status`; generated paths use `itsy.config.json`/`itsy/`.
  A separate full starter matched all nine checked canonical config/content,
  toolbar, CSS and branding files byte-for-byte and passed status.
- Canonical/legacy config, layout, archive, media and session regression tests;
  legacy toolbar CSP/initialization and canonical/legacy artifact catalogue tests.
- The pinned shared-skill SHA-256 manifest remains valid. Canonical shared skills
  were preserved, with project-owned ItsyWeb adapters for current paths.

Browser verification found and fixed missing legacy/new root resolution in the
photo and collection scripts. The navigation test now checks the computed page
name after the native transition finishes while still checking its actual named
animation during the transition. This avoids Chromium's transient computed name.

Verified Linux artifact hashes:

```text
e9f1f33e96f10a37cdc69b1a2ff0564ce4dc190101edf962bc896b27431d5823  itsy-server-linux-amd64
330395647a365674b4ec35797ec35f3ecc0a2fd28aa394fc273e92f1dd380f54  itsy-server-linux-arm64
2ff6e2d0d7155d2f69e3c4e508f37285e59fa60c2c245e3002dba064129b936f  itsy-linux-amd64
```

The ARM64 Linux server is cross-compiled, without native ARM64 acceptance.
No Linux ARM64 CLI was built or published. Native Mac ARM64 checks, signing,
public install/reinstall and both updater transition steps remain pending.

## Kirin and shared server

Kirin's normal `itsy` resolves to `/home/sean/Projects/flea/bin/itsy`. Its hash
matches the Linux artifact above. The prior CLI is retained privately. Existing
`flea` remains available for the notice and exact `flea update` exception.
Actual `flea status` prints the notice and stops with status 2. Actual
`flea update` prints the strong rename notice and handles the already-current
0.21.0 CLI successfully.

The supported `itsy update-server` flow upgraded `root@fleacms.com` from
0.20.3/protocol 28 to 0.21.0/protocol 29. The installed server hash matches the
native AMD64 artifact above. It validated all existing sites before the swap,
retained the prior executable, restarted successfully and left no pending journal.
The existing service/account/paths remain under their Flea names for compatibility.

The 320 MiB service memory target and active 1 GiB swap were preserved. Admin
operations used the same memory target and task-private disk-backed temporary
storage through a reviewed, task-only SSH wrapper. No Go build ran on production.

All 22 managed sites were backed up to private verified archives before the
upgrade. Post-upgrade and post-deployment comparisons preserve the 34 configured
hostnames, branch ownership, TLS names, redirects, downloads and indexing rules.
Live content is byte-identical to the baseline across every site. Recorded
runtime-file metadata and session hashes are unchanged. All 34 HTTPS root
responses retain their expected status/routing/indexing. Site 4's varying HTML
is its ordinary per-request visitor CSRF token.

The following 13 sites have new immutable releases using `itsy.config.json`
and `itsy/`, with live edits and custom branding preserved:

| Hosted site | Deployed source |
| --- | --- |
| `battertonbrokerage-main` | `64f1b566` |
| `battertonbrokerage-preview` | `522eba3e` |
| `codepilot-site-1-main` | `b67c4019` |
| `codepilot-site-2-main` | `27fd3c89` |
| `codepilot-site-3` | `ff5668cc` |
| `codepilot-site-5` | `931ad79d` |
| `codepilot-website-staging` | `69619b9d` |
| `cubitalk-website` | `84c32456` |
| `dropr-preview` | `493d727c` |
| `gestalt-haus-preview` | `247e4d51` |
| `prewind-website` | `eb1bb09a` |
| `sunlux-main` | `0b8f31a6` |
| `t3docs-main` | `670df0b6` |

Local project checkouts, the actual active monorepo website sources and the
required main/preview branches were converted with config/data/media retention
checks. Read-only status validated the converted sites. Existing unrelated files
in the website root checkout were preserved. Dormant historical task worktrees
were not rewritten. T3 project titles now say ItsyWeb and ItsyWeb Website.

## Remaining rollout blockers

- Native Mac release access: SSH to `sean@100.79.67.33` on the online Mac times
  out. Enable Remote Login or provide a reachable native Mac build host. The
  exact committed source archive and verified Linux server artifacts are ready
  privately on Kirin. Do not substitute cross-compilation for native acceptance.
- Public distribution: the stable channel still selects 0.20.5. The currently
  served itsyweb.dev installer requests canonical `itsy-*` files, which are not
  present in that historical release (HTTP 404). The old fleacms.com installer
  and retained downloads remain available. Complete the native gate, publish the
  four 0.21.0 files (canonical plus byte-identical legacy aliases), enable the new
  hosts' validated download mapping, verify HTTPS hashes, then activate both
  website branches' matching installer/latest.txt and native update checks.
- Atlas: saved `atlastitlesurveys.com` resolves to 147.79.120.158 rather than this
  server's expected addresses. Deployment stopped before upload/activation. The
  existing Atlas release stays live; DNS/target selection needs resolution.
- CodePilot main, Site 4, Gestalt main and Onino: local committed presentation
  changes differ from the published sites. The owner's pending choice determines
  whether to publish those changes or prepare naming-only updates from the
  existing published source. Their compatible current releases remain intact.
- Kent: the cloned GitLab source's public pages match the verified live backup,
  but its admin email list is empty while the live site has an admin. Local paths
  are converted; deployment is held to preserve access until current source/config
  is identified or the intended alignment is confirmed.
- Naves: no local/GitLab source was found; the usual CP Git URL denied access.
  Its source path/URL is requested. The verified backup and compatible live
  installation remain preserved.

Private backups, source archive, hashes, statuses, inventory and deployment logs
are in `/home/sean/itsyweb-releases/0.21.0/`, outside every served/deployed tree.
Existing clean-Mac/notarization, native installer packaging, real-email and native
ARM64 server acceptance limits remain unverified.
