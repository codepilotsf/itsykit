# ItsyWeb 0.21.0 rollout verification

Updated November 14, 2026. The public test release and all 22 hosted-site updates are complete.
The owner authorized the rename, Kirin/server deployments and public rollout,
then chose to preserve existing customer designs and update only Atlas’s preview.

## Source and Linux checks

Framework executable source: `04669d71d988b0425571834426f7a678723bb5ff` on `main`, pushed to the
private framework repository. Immutable tag `v0.21.0` selects that tested source. Website source `6affc545`
activates the approved naming/domain cutover, installer and 0.21.0 channel on both
`main` and `t3/reframe-flea-html-focus`. Historical version directories are unchanged.

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

Native Mac ARM64 used Go 1.27.1 on Sean’s M5 Mac. The same committed source
archive and verified Linux server pair were copied into a fresh private build
folder. The native client release script passed full tests, full race tests,
vet, real Contact/bundle checks, Chrome browser checks and fresh-site smoke.
The final Mac executable passed ad-hoc signature verification. Its public hash is:

```text
1dcde316b29933b7e88882eef9dbee0a56cba3f5dc15cba367831bd6b1ee5e2a  itsy-darwin-arm64
```

Canonical and retired `flea-*` artifact aliases are byte-identical and have
separate regular files/checksum entries. All four public binaries matched the
manifest over HTTPS before channel activation. The 0.20.5 manifest remains
available. Both native platforms then passed actual public fresh install,
reinstall, `new`/`status`, retired-command stop, update from the released 0.20.5
binary, second-update conversion/prominent notice and already-current checks.
The normal Mac and Kirin CLIs match these public hashes. The Mac’s original CLI
is retained privately and its interactive shell resolves the new `itsy` and
retired `flea` entries in `~/.local/bin`.

The ARM64 Linux server is cross-compiled, without native ARM64 acceptance.
No Linux ARM64 CLI was built or published. Tag pipeline 843 is pending runners;
completed native manual runs supply the release evidence.

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
Customer live content is byte-identical to the baseline. Only the approved
official-site branding/copy changed. Recorded runtime-file metadata and session
hashes are unchanged. All 34 HTTPS root
responses retain their expected status/routing/indexing. Site 4's varying HTML
is its ordinary per-request visitor CSRF token.

All 22 sites now have active immutable releases using `itsy.config.json` and
`itsy/`. The following table records the final observed sources; an independent
CubiTalk publication was preserved:

| Hosted site | Active source |
| --- | --- |
| `atlas-main` | `880b1661` |
| `battertonbrokerage-main` | `64f1b566` |
| `battertonbrokerage-preview` | `522eba3e` |
| `codepilot-site-1-main` | `b67c4019` |
| `codepilot-site-2-main` | `27fd3c89` |
| `codepilot-site-3` | `ff5668cc` |
| `codepilot-site-4-main` | `45331e65` |
| `codepilot-site-5` | `931ad79d` |
| `codepilot-website-main` | `6585e3ad` |
| `codepilot-website-staging` | `69619b9d` |
| `cubitalk-website` | `d879987a` |
| `dropr-preview` | `493d727c` |
| `fleacms-website-main` | `6affc545` |
| `gestalt-haus-main` | `8035f173` |
| `gestalt-haus-preview` | `247e4d51` |
| `itsy-web` | `6affc545` |
| `kent-matheson-main` | `bd56e8fe` |
| `naves-main` | `3cfb03c0` |
| `oninopizza-main` | `adc9722d` |
| `prewind-website` | `eb1bb09a` |
| `sunlux-main` | `0b8f31a6` |
| `t3docs-main` | `670df0b6` |

Local project checkouts, the actual active monorepo website sources and the
required main/preview branches were converted with config/data/media retention
checks. Read-only status validated the converted sites. Existing unrelated files
in the website root checkout were preserved. Dormant historical task worktrees
were not rewritten. T3 project titles now say ItsyWeb and ItsyWeb Website.

## Completed final rollout

After Tailscale on the Mac was reconnected, SSH/native builds succeeded. Naves
and the current Kent source were cloned from `/Users/sean/Projects/` into
Kirin’s `Projects/naves` and `Projects/kent-matheson`. Their source matched the
published runtime and Kent’s existing admin access was retained. The older,
task-created Kent GitLab copy is retained privately, outside active Projects.

For CodePilot main, Site 4, Gestalt main and Onino, the owner chose to preserve
published designs. Isolated deployment checkouts used their exact published
sources; Onino came from its fresh verified snapshot where the source commit
was unavailable locally. Config, media and content were retained. Public source
files differed only in naming identifiers, and the four rendered homepages
matched the prior published HTML after normalizing visitor CSRF tokens. The
unpublished designs in the original checkouts were preserved.

Atlas now deploys through its existing `atlas.fleacms.com` alias, as explicitly
selected by the owner. Its other host aliases and all public DNS were preserved.
No DNS records were changed.

The immutable download directory was uploaded and verified before channel
activation. Under the deployment lock, a backed-up, validated host config added
`itsyweb.dev` and `itsy.fleacms.com` to the existing download store. The initial
readiness probe ran before catalogue startup finished; the prior configuration
was restored. A retry waited for the exact new revision, then passed the native
server readiness check. The server is healthy and no pending journal remains.

Both website branches then activated 0.21.0. Fleacms.com’s approved branding/copy
was published explicitly with `--with-snips` after pulling and reviewing its
baseline; customer sites used ordinary deployment. Customer content hashes,
runtime/session checks, TLS names, branch ownership, redirects and indexing
match the pre-rollout baseline. Only the approved official-site copy changed.
All 34 HTTPS root responses retain their expected routing/status/indexing. All
four website aliases serve latest.txt = 0.21.0 and the exact canonical installer.
Public homepage rendering and public docs version/protocol/download links were
verified. Temporary browser tabs were closed.

Private backups, source archive, hashes, statuses, inventory and deployment logs
are in `/home/sean/itsyweb-releases/0.21.0/`, outside every served/deployed tree.
Existing clean-Mac/notarization, native installer packaging, real-email and native
ARM64 server acceptance limits remain unverified.
