# FleaCMS 0.20.2 release verification

Published November 2, 2026 after the owner approved the Nunito logo with adjusted
spacing and requested project-wide deployment and a release. Public downloads,
website/docs, shared server and normal Kirin/Mac CLIs are 0.20.2. Deployment
protocol remains 28; app protocol remains 2.

## Source and artwork

- Framework MR !29 merged as `f00132e`. Immutable `v0.20.2` selects tested
  source `279ec50`; source archive SHA-256 on Linux and Mac is
  `0991e4848a4e5c6ee6f31848e76d0ef530b097e8e470257bd740f2552f698513`.
- Website MR !14 merged as `648e922`. Normal deployment activated
  `20261102-203411-64cea67e0d016a92327e5c39` from that committed snapshot.
- The owner-approved `f5b19f8` proof supplies the heavier Paired Little Skip and
  Nunito 800 lockup. Canonical masters, original prompt/source, font license and
  export provenance are in [branding](../branding/README.md).
- Logos are SVG paths with outlined lettering and no external image/font requests.
  The standalone mascot, adaptive SVG favicon, 16/32px ICO, PNG and Apple touch
  icon derive from the same approved artwork. Website assets match the masters.
- Starter and framework website layouts, diagnostic page, READMEs, public header,
  footer, docs/admin wrappers and 1200x630 social preview use the new identity.
  Body typography and existing custom site branding are preserved. Diagnostic
  branding works without site assets and retains its original restrictive CSP.

## Verification

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) passed the canonical
bundled build and native release client checks: full tests, full race tests, vet,
real Contact child and server bundle checks. Linux also passed the native server
release checks and cross-built the ARM64 server. Both final clients embed the
same Linux server artifacts. Mac ad-hoc signature verification passed.

Exact local and publicly downloaded clients passed current-directory creation,
initial Git commit, starter asset parity, status and overwrite refusal. The exact
native Linux server passed GET/HEAD, host/indexing isolation, asset, 404 and health
checks. A deliberate scratch-site template error verified the branded diagnostic
response and unchanged CSP. Live website/docs/admin and starter branding were
checked at desktop and mobile widths, including loaded logos and no overflow.
All eight public logo/favicon/social assets match the reviewed website files.

Both supported native platforms passed anonymous public fresh installation,
reinstallation, 0.20.1-to-0.20.2 update and already-current behavior, including
paths containing spaces. Installed normal CLI hashes match the public manifest.
Kirin's symlink still resolves to its existing project `bin/flea`. A pre-existing
Dropr preview was left running without interruption. Asahi is decommissioned and
was not contacted.

Tag pipeline 709 was pending native runner availability. Publication used the
completed native checks above; no remote CI success is claimed. An initial Linux
build hit the existing `/tmp` quota; the release job used private disk-backed
scratch thereafter. Native tests used a private job umask so the shell's group
write default did not invalidate strict download-store permission fixtures.
No product security checks or global environment/permissions were weakened.

## Artifacts and preservation

Anonymous HTTPS downloads match these final signed-byte checksums:

```text
015c092a3622335f3e78a9cee364d8b9bef2ad907875ec0c7ebc1bb6f87b4bbc  flea-darwin-arm64
6ccfa190ac5db9165db3a9327e5195e3f17727484239dc3bd4c9b02a488b0e7a  flea-linux-amd64
```

The shared server hash is
`83af63cae6cc3b5fe6c89f10143f88b255d898b7c8806eb8dcf9804ec986d840`.
The new immutable download directory was checked and installed under deploy.lock;
no old published bytes were replaced. Retained 0.20.1 public downloads still
verify. Supported `flea update-server` completed without recovery or a remaining
journal. All 18 source trees, live content and host policies matched across the
executable upgrade. After website deployment, all live CMS content and the other
17 sites' source trees matched the baseline. All 24 HTTPS hostnames retained
expected statuses, redirects and indexing headers. The update channel is 0.20.2.

Private evidence and previous CLIs are under `/home/sean/fleacms-releases/0.20.2`
and `/Users/sean/fleacms-releases/0.20.2`. The website backup is in the Linux
archive; the remote config/executable backup is `/var/backups/flea-0202.BArcPnY2`.
Only task-owned scratch/previews are cleaned up. Existing 320 MiB Go memory target,
1 GiB swap, quotas, mounts, cleanup policies, TLS and host ownership are unchanged.

Native ARM64 Linux server execution, clean-Mac acceptance, notarization, native
installer packaging and real email delivery remain unverified. Source browser
tests requiring FLEA_TEST_BROWSER were not run. No Rails, payments or native app
distribution changed.
