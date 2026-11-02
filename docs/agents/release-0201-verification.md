# FleaCMS 0.20.1 release verification

Published November 2, 2026 after the owner requested a public patch release and
rollout to Kirin and the shared fleacms.com server. Public channel, shared server,
Kirin CLI and Mac CLI report 0.20.1. Deployment protocol remains 28; app protocol
remains 2. No existing website source was migrated or restyled.

## Source and checks

- Immutable `v0.20.1` selects `c14efd7`, merged through framework MR !27 as
  `cba671b`. The temporary-storage fix was merged through MR !26 as `7cefff1`.
- Source archive SHA-256 on Linux and Mac:
  `7b70502310f87422a769c5790bbbd7e6cfd7b76380b89b249d1e7336ddb9379c`.
- Native Linux AMD64, Go 1.26.9: canonical bundled build, native server release
  script and native client release script passed full tests, full race tests,
  vet and real Contact/server bundle checks. The ARM64 server was cross-compiled.
- Native Mac ARM64, Go 1.27.1: canonical build and native client release checks
  passed full tests, full race tests, vet, real bundles and ad-hoc signature
  verification. Final clients embed the same Linux server artifacts.
- Exact final CLIs passed current-directory creation, initial Git commit,
  starter-file parity, status and refusal to overwrite a nonempty directory.
  The exact Linux server passed GET/HEAD host isolation and staging indexing
  checks, including assets, missing pages, sitemap, robots and health.
- Regression coverage verifies caller-supplied TMPDIR, archive contents,
  ordinary/cancelled cleanup, invalid/unwritable locations and concurrent private
  directories. The release client's smoke directory now has exit cleanup.
- Tag pipeline 704 was pending native runner availability. Publication used the
  completed native checks above; no remote CI success is claimed.

## Public artifacts and installations

Anonymous HTTPS downloads match final signed-byte checksums:

```text
981e4c1d6cd2a6c5c50ee36d731bf2bac773b011d52c81585562c3b0dce06968  flea-darwin-arm64
8c0894c23f2f5298ad01da92f88af4d49b0b837ca021bdc789bba7df480a70b9  flea-linux-amd64
```

The new immutable directory was staged on disk, checked for exact files,
ownership, modes and checksums, moved without overwrite under deploy.lock and
validated. The old 0.20.0 public manifest is unchanged. No published downloads,
retained site releases, content or media were pruned.

Both supported native platforms passed anonymous public install, reinstall,
0.20.0-to-0.20.1 update, already-current behavior and artifact smoke checks in
isolated directories, including paths with spaces. The installer matches the
canonical source. Normal Kirin and Mac installations were updated through
`flea update`; previous executables are retained privately. Installed hashes
match the public manifest. Kirin's CLI symlink still resolves to its project
`bin/flea`, shared by its local site workflows. No separate project executable or
running Flea dev process was found in the Kirin inventory. Local site sources
were not replaced. The owner confirmed Asahi is decommissioned; it is excluded
from the active inventory and future rollouts.

## Shared server and website preservation

Supported `flea update-server` installed 0.20.1 and retained the prior executable.
The server SHA-256 is
`b0ed25ede09f614eb35f2c167462a08a84090236af065cdd7043d52e847631c4`.
No recovery was needed. Existing GOMEMLIMIT=320MiB and active 1 GiB swap are
unchanged; no pending activation or upgrade journal remains.

Across the executable upgrade, server configuration stayed byte-identical and
all 18 deployed source trees and live content hashes matched. All 24 HTTPS
hostnames retained expected status, redirects and indexing headers. Updating the
shared executable covers every registered site without republishing their code.

Website MR !12 merged as `619f8a2` and deployed normally as
`20261102-185436-78c8e68d3d9d74d6f432dd9a`. Fresh pull made no source changes;
normal deployment preserved live editable content. Only `site/docs.html`,
`site/index.html` and `site/releases/latest.txt` changed in the deployed snapshot.
Every live content hash and the other 17 site roots/source trees are preserved.
Host TLS, redirects, indexing, downloads and branch ownership are unchanged.
All 24 HTTPS hosts passed again after website publication. Browser checks at
1280px and 390px confirm current release text and the TMPDIR example, with no
horizontal overflow, missing images or observed console errors.

## Retained evidence and limits

Private artifacts, source archive, logs, downloaded binaries and comparison
evidence are under `/home/sean/fleacms-releases/0.20.1` and the corresponding Mac
path `/Users/sean/fleacms-releases/0.20.1`. The website backup is in the Linux
archive. Remote config/executable backups are in
`/var/backups/flea-0201.uVhW4aef`; supported upgrade recovery also retains its
previous executable. Scratch used private disk-backed per-job directories.

Native ARM64 Linux server execution, clean-Mac acceptance, Apple notarization,
native Mac installer packaging and real email delivery were not tested. Source
browser cases requiring FLEA_TEST_BROWSER were not run. No Rails, payments,
native application distribution, quotas, mount settings or cleanup policy changed.
