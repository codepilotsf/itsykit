# FleaCMS 0.20.0 release verification

Published and deployed October 30, 2026. The public channel, shared server and
normal Kirin/Mac CLIs report 0.20.0. Deployment protocol is 28; app protocol is 2.

## Source and native checks

- Immutable `v0.20.0` selects `f5f47d1251c25da2271b8dc3bf353313ed122968`,
  merged as `14dde06` through framework MR !22.
- The release reconciles migration redirects with the SEO discovery feature
  merged through MR !21 while builds were running. The earlier redirect-only
  candidate was retained privately and never tagged or published.
- Final source archive SHA-256 on both build machines:
  `f3d2fbf51c8ad6a22b4788b1d104e9ca9e529c6ff8acc2e06acacb49c9d6135e`.
- Native Linux AMD64, Go 1.26.9: canonical bundled build, then the server and
  client release scripts with real Contact/server bundles required. Full tests,
  full race checks and vet passed. The ARM64 Linux server was cross-compiled.
- Native Mac ARM64, Go 1.27.1: canonical bundled build, then native client
  release checks with the same final Linux server artifacts. Full tests, full
  race checks, vet and final ad-hoc signature verification passed.
- Exact final CLIs passed current-directory creation, initial Git commit,
  minimal/starter parity, nonempty-directory refusal and status checks. Real
  GET/HEAD checks covered 301 destinations, query retention/replacement,
  200/404/405 behavior, sitemap, robots.txt, canonicals and noindex exclusions.
- The exact Linux server passed shared-root host isolation, including staging
  noindex on generated discovery documents. Sources and snapshots stayed private.
- Tag pipeline 658 was pending native runners at publication. The release used
  the manual native checks above; no remote CI success is claimed.

## Public artifacts and updates

Anonymous HTTPS downloads match the final signed-byte manifest:

```text
f9820e1ca70976f2489b4149839173052e1cd160cf0c074a83f8a98229d05826  flea-darwin-arm64
2a9411fd9d243c58be201dcd10d048a8724fd855bb1ce9550f5139e3f31bf1f9  flea-linux-amd64
```

The version directory was staged on disk, checked for ownership, modes and
checksums, moved without overwrite under deploy.lock, and validated before the
supported executable upgrade restarted the service and loaded the catalogue.
The 0.19.0 manifest is byte-identical to the retained baseline. No published
versions, site releases, content or media were pruned.

On both supported native platforms, public fresh install, reinstall,
0.19.0-to-0.20.0 update and already-current checks passed in isolated directories,
including paths with spaces. Installed hashes and exact CLI smoke checks passed.
The public installer is byte-identical to canonical source. Normal Kirin/Mac
CLIs were then updated through `flea update`, with prior executables retained.
Kirin's `.local/bin/flea` symlink still resolves to the project `bin/flea`.
Existing user dev processes were not restarted.

## Shared server and preservation

Supported `flea update-server` completed against root@fleacms.com with its normal
interactive confirmation. The installed server SHA-256 is
`138829102b581dc9e70e05338f1a6fa54effabad5401e007c0401eb73b6482f8`.
No recovery was needed; the service is active with no pending recovery journal.
The existing 320 MiB Go memory target and active 1 GiB swap remain unchanged.

Across the executable upgrade, all 17 site roots, source-file hashes, live content
hashes and captured host-policy fields matched exactly. All 23 HTTPS hostnames
retained their expected 200/308 status, redirects and indexing headers.
No existing live site had a `seo` setting; the upgrade did not opt sites in.

After website deployment, every live content hash and all 16 unaffected site
roots/source hashes still match. Host indexing, TLS, WWW redirects, download
mappings and site/branch ownership are unchanged. The Flea website changed only
`site/docs.html`, `site/_docs-toc.html`, `site/index.html` and
`site/releases/latest.txt` in its deployed source.

## Website and documentation

Website MR !9 merged as `9190acb`, deployed as
`20261030-004255-1885fadf131e62d999b94551`. The channel selects 0.20.0; homepage,
download links, README and public docs cover migration redirects, URL-inventory
skills and opt-in SEO discovery. Historical release notes remain intact.
The live browser verified both new sections, parsed configuration examples,
contents links and protocol text at desktop and mobile widths without horizontal
overflow. No layout, stylesheet, installer or live editable content was changed.

Framework README, SEO guide and dated changelog describe the released behavior.
The website's RELEASES.md and handoff.md record the publication and activation.

## Backups and limits

Private artifacts and evidence are retained under
`/home/sean/fleacms-releases/0.20.0-candidate2` and the corresponding Mac path
`/Users/sean/fleacms-releases/0.20.0-candidate2`. The first candidate and website
backup remain under the adjacent `0.20.0` archive. Previous remote config and
executable are retained in `/var/backups/flea-0200.XcBnqhIR`; supported upgrade
recovery also retains its previous executable.

Native ARM64 Linux server execution, clean-Mac acceptance, Apple notarization,
real email delivery and hosted crawler/Search Console acceptance were not tested.
Source browser cases requiring FLEA_TEST_BROWSER were not run. HTTP/browser
checks do not establish search indexing or rankings. Task-owned dev servers and
temporary network transport were stopped after verification.
