# FleaCMS 0.18.0 release verification

Published October 21, 2026. The immutable downloads and public update channel
are live. Framework and website changes are merged; the website is deployed.

## Source and platform scope

- Release source: `7ab8f6ea022f276600bc373d7353559c8552bc11`, immutable tag
  `v0.18.0`; merged into framework main as `f02af1d` through MR !16.
- Feature base: `7e137e445a0b9f144559bac0aff7fd83b151ed2c`.
- Adds `flea new .` and new/existing empty destination directory paths.
  Nonempty directories, including hidden files and `.git`, are refused without
  overwriting. Existing naming, first-run prompts and Git defaults are preserved.
- Linux ARM64 CLI downloads are discontinued by owner decision. The supported
  CLI targets are Apple Silicon macOS and x86-64 Linux. Installer, updater, native
  release CI, build guards and release staging enforce that matrix.
- Both CLIs retain x86-64 and cross-compiled ARM64 Linux servers. Native ARM64
  server acceptance is a separate check and was not run for this release;
  Asahi was offline. No ARM64 Linux machine was used to build the server.
- Deployment protocol remains `FLEA_DEPLOY=26`; app wire protocol remains 2.
- Carries the global-only Impeccable and starter agent-guidance changes on main.
  Unrelated local Kirin skill-install commits and source checkouts were preserved.

## Native build and checks

Private archive: `/home/sean/fleacms-releases/0.18.0`; Mac counterpart:
`/Users/sean/fleacms-releases/0.18.0`. Evidence is private and outside the public
website. Exact source archive SHA-256 on both machines:
`82aed00f6c35705416b0d4b3374923dc368b6989a637dc292be10c01e7c821e6`.

- Native Linux AMD64, Go 1.26.9: `scripts/build.sh`, then
  `FLEA_REQUIRE_BUNDLE=1 CI_COMMIT_TAG=v0.18.0 scripts/release.sh server amd64`.
  Full tests, full race tests and vet passed with real Contact/server bundles.
  This is a native AMD64 build/full-suite result, closing that gap for 0.18.0.
- Native macOS ARM64, Go 1.27.1: canonical `scripts/build.sh`, full tests, full
  race tests and vet passed with both bundle-required flags. Source came from
  the exact release archive, without modifying the normal Mac checkout.
- Final CLIs embed the same Linux server binaries produced on AMD64: native
  x86-64 and cross-compiled ARM64. Final bundle extraction checks passed. Final
  parent builds use `CGO_ENABLED=0`, `-trimpath` and the recorded version.
- Mac final ad-hoc codesign and strict signature verification passed.
- Exact final binaries passed current-directory creation with an initial Git
  commit, default/starter file parity, status and repeat/hidden-file/symlink
  refusals on both native platforms.
- Installer/updater regressions verify that unsupported platforms preserve the
  existing CLI; staging tests require exactly the two supported artifacts and
  omit even a leftover Linux ARM64 CLI. Shell syntax and diff checks passed.
- Linux commands used `umask 022`; source browser test cases requiring an
  explicitly configured browser were not run. Collaborative browser checks
  verified local and live installation text, quickstart and download links.
- Tag pipeline 481 was pending native runners at publication. This release used
  the manual native checks above; no remote CI success is claimed.

## Published artifacts

The public manifest at `https://fleacms.com/releases/0.18.0/SHA256SUMS` contains
exactly these two binaries, and anonymous HTTPS downloads match both hashes:

```text
e6b9a3dde2a7e832501d7df2bfe1acb6c9c7a2bf1789e88de14205631fc81a5f  flea-darwin-arm64
eb0c6659cd032c72a38fc1989252217d5629cdec66e10ee35189e3b7cbb5d6ae  flea-linux-amd64
```

The canonical installer and public website installer are byte-identical. Native
public fresh install, reinstall, update from 0.17.0 and already-current checks
passed on both supported platforms as the ordinary user; installed hashes match
the manifest. The freshly installed CLIs also passed the artifact smoke checks.
Normal user CLI installations were not replaced by these isolated checks.

## Publication and website deployment

- Downloads were uploaded to private disk-backed staging, checked for exact
  filenames, checksums, ownership and modes, then moved without overwrite into
  `/var/lib/flea/downloads/fleacms/0.18.0` under the deployment lock.
- Host validation passed with `GOMEMLIMIT=320MiB`; the existing 0.17.0 server was
  restarted to load its new download catalogue. The first immediate loopback
  probe raced startup and failed to connect; subsequent readiness and public
  probes passed. No rollback or recovery was needed and no journal remains.
- No executable server upgrade was needed for this CLI feature. Production
  remains 0.17.0 / protocol 26, with its existing memory target and swap intact.
- Website MR !5 merged to source `1c786f2daad2c5489207eac840ed954654536655`,
  deployed as `20261021-174934-4dbfcc67281f82d2e68b36a4`.
- `latest.txt` selects 0.18.0. Public homepage/docs, the two download links and
  the new empty-directory instructions were verified. Final website status
  reports matching deployed code and local content from that deployed commit.
- All 19 configured HTTPS hostnames return 200. Host policy and non-website
  release roots are unchanged, and all editable-content hashes are preserved.
- The previous 0.17.0 manifest and Linux ARM64 binary remain available; the
  latter still matches `8c0ac734abf2dcbb2e9aac6e808a74c83040c21d067c2a3c1906c7fe0b96b09d`.
  No published version, content, media or retained release was pruned.
- Task-owned dev preview was stopped. Production and both original source
  checkouts were kept separate from release build and installer fixtures.

## Remaining limits

Native ARM64 Linux server execution, clean-Mac acceptance, Apple notarization
and real email delivery were not tested for this release. Historical verification
records retain their original platform matrix and evidence.
