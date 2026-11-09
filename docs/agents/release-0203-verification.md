# FleaCMS 0.20.3 release verification

Published November 9, 2026 after the owner approved filled mascot concept D and
requested deployment across the website and FleaCMS, including favicons.
Deployment protocol remains 28; app protocol remains 2.

## Source and artwork

Framework MR !32 merged as `1071144`. Immutable `v0.20.3` selects tested source
`26bc0e8`; source archive SHA-256 is
`dc3c85513b677c8a056bad20daa5e133774fbfe2be13102de2a5cd8745d0bbee`
on Linux and Mac. The exact approved source, prompt and dependency-free SVG
export are retained in [branding](../branding/README.md). Filled shell sections,
face and bent legs follow concept D; Nunito outlines and lockup dimensions remain
unchanged. Existing forest and Navy and Ice palettes are retained.

Canonical masters, starter, framework website, development diagnostic and public
website use the filled mark. SVG, 16/32px ICO, PNG, Apple touch and social assets
are updated. Website MR !16 merged as `6bce343` and activated
`20261109-004926-199b26ab9e0c6aff256e0ac1`. Channel MR !17 merged as `577a722`
and activated `20261109-005527-66744fb1ea339e9e65840e66`.

## Verification and rollout

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) passed canonical bundled
builds and native release scripts, including full tests, race tests and vet.
Both final CLIs embed the same Linux server binaries. The Mac final executable
passed ad-hoc signature verification. Exact clients passed current-directory
creation, initial commit, minimal/starter asset parity, status and overwrite
refusal. The exact native Linux server passed host/indexing isolation, GET/HEAD,
assets, 404 and health checks. A scratch-site template error verified the new
static diagnostic SVG and the unchanged restrictive CSP.

Desktop and phone-width homepage/docs/admin checks passed. All eight public
logo/favicon/social assets match reviewed source. Anonymous public downloads
match final signed-byte hashes. Fresh install, reinstall, 0.20.2 update and
already-current checks passed natively on both platforms. Normal Kirin and Mac
CLIs now report 0.20.3 and match the public manifest. Kirin's existing CLI symlink
is preserved; old executables are retained in the private release archive.

```text
7eb2c7e094ffca80e569261a7657633b81eeda4bf86cc3f51a831dded9bbae33  flea-darwin-arm64
be70a532a2a86e5be2dd6d4c270d3c45a89a4b0eda425a10791126ac5f1b8308  flea-linux-amd64
```

The shared server reports 0.20.3; its SHA-256 is
`4ba946401b91ceb723f30b73b8c44803f8e6e306e5b31d60bae1ef1324ef7f75`.
The immutable download directory was validated before the supported executable
upgrade made it available. Old 0.20.2 downloads remain unchanged. All 29 HTTPS
hostnames retained expected statuses, redirects and indexing headers. Live CMS
content and host policies match the initial snapshot. A concurrent Cubitalk
deployment advanced its own code pointer during the rollout; it was preserved.
Aside from FleaCMS and that independent deployment, site sources are unchanged.
The first website activation encountered a busy lock and made no changes; a
clean retry succeeded after checking for active operations and recovery journals.

Tag pipeline 797 was pending native runners. Completed native manual checks
supplied publication evidence; no remote CI success is claimed. No customer
branding, access settings, systemd/memory/swap policy, TLS or host ownership changed.
No old downloads, retained releases or media were deleted. Task previews stopped.
Private logs, checksums, previous CLIs and the website backup are under
`fleacms-releases/0.20.3` in Sean's home on each machine. The remote executable
and config backup is `/var/backups/flea-0203.tC2FUbNf`.

Native ARM64 Linux server execution, clean-Mac acceptance, notarization,
installer packaging and real email delivery remain unverified.
