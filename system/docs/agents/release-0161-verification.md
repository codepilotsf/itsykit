# FleaCMS 0.16.1 release verification

October 18, 2026. Sean authorized committing and releasing the literal-brace fix,
updating the Mac installation and the shared server. The original checkout at
`/Users/sean/Projects/flea` is based on 0.10.0 with unrelated uncommitted work;
it is preserved. The release worktree is `/Users/sean/Projects/flea-release-0.16.1`,
branch `release/0.16.1`, based on immutable 0.16.0 tag `1c85aae`.
Unreleased modular skill changes on main are excluded from this patch release.

## Scope

A shared lexical scanner distinguishes exact Flea commands and paths from
literal brace groups in compiler and rich/media/multisnip detection. Original
source is rendered with Go contextual escaping. Sanitized content stays data;
no Impeccable code, parameter format or workflow changes are included.

## Verification and rollout

Source/tag `57d8378df75c9981bb5fe1af173dbad81ac8afb4` / `v0.16.1` are pushed.
The tag is immutable. Mac ARM64 and Linux ARM64 full tests, vet, full race suites
and standard bundled builds passed with real Contact/server bundles required.
Go 1.27.1 was used on both. AMD64 was cross-compiled on the Mac and checked
natively as nobody; native AMD64 build/full-suite remains open. The ARM64 CLI
was built natively from a private archive of exact tagged source; all CLIs embed
the same Mac-built Linux servers. Mac and AMD64 artifacts have clean tagged VCS
metadata. Mac ad-hoc signature verification passed; clean-Mac/notarization remain open.

Public hashes:

```text
8191404446e6025408b0e23f7bf84b740bc2d475a22164418b015635deee2390  flea-darwin-arm64
d3162dab92e5e3376cb93aaa4f0aad0b8fc2d36ddb320d45424df02d80fe8199  flea-linux-amd64
37443e9b9c5e24f58fece7d2185d2d38bccad6e1e62eb190d121863f53de4bc5  flea-linux-arm64
```

Actual artifact smoke passed new/status, canonical minimal starter comparison,
raw JSON, HTTP 200, source reload and graceful shutdown on Mac/ARM64 Linux and
AMD64 as nobody. The exact Mac release binary and unchanged Impeccable helper
passed browser JSON.parse, reload/session restoration, real Tune slider changes
from 0.5 to 0.8 (padding 10px to 16px), and durable discard cleanup. Temporary
helper/dev processes stopped. Impeccable and site design remain unchanged.

Public HTTPS downloads and manifest match the immutable local archive at
`/Users/sean/fleacms-releases/0.16.1`. Fresh installation, reinstallation,
0.16.0-to-0.16.1 updates and already-current checks passed on all three platforms.
Normal Mac `/Users/sean/.local/bin/flea` and Linux `/home/sean/.local/bin/flea`
were updated using the public updater, with version and hash verified.

Supported `flea update-server` updated root@fleacms.com's shared server from
0.15.1 to 0.16.1 / protocol 25. Service active, no recovery journal. The existing
320 MiB admin/service memory target and swap policy were preserved. All 18
configured HTTPS hosts return 200. Content hashes, TLS and downloads mappings
match baseline; only Flea marketing site's release roots changed on deployment.
Private pre-update website backup:
`/Users/sean/fleacms-backups/fleacms-before-0.16.1-20261018.tar.gz`.

Website source `d59d74e`, activation
`20261018-174259-d8975e4a784b2f5837ed6ef4`; homepage/docs/latest.txt/download
links show 0.16.1. Website receipt commit `4b28bb9` follows the deployed source.
Public installer, local Mac/Linux CLIs and shared server agree on 0.16.1.
No real email, new server setup or clean-machine acceptance was performed.
The original 0.10.0-based dirty framework worktree remains preserved; use the
release worktree/branch for this release rather than rebuilding that older tree.
