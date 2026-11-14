# FleaCMS 0.20.4 verification

Published November 9, 2026 to keep browser and touch icons white on a fixed navy
background, matching the public header in both browser themes. The approved
mascot, wordmarks, page design and customer branding are unchanged.

Tested source `f0269b2`, immutable `v0.20.4`, merged through framework MR !34
as `f6f5430`. Source archive SHA-256 on the release machines is
`453823d7f253e8d891aeb2a2ca5f91085c3935821ce47ffd4ed23c5fd8f7d9d6`.
Website icon MR !19 merged as `59f7ced`, activation
`20261109-012644-c58945602a2360256d46434d`. Channel MR !20 merged as `6694e69`,
activation `20261109-013053-45656596029f6bb3f0a161bd`.

SVG icons have no theme override. PNGs have opaque backgrounds matching header
navy `#142b43` and foreground `#f5f9ff`. All framework/starter/website copies
match, and ICO entries contain the exact validated 16px and 32px exports.
All four public icon files match reviewed source; asset URLs refresh browser caches.

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) canonical builds and
release scripts passed full tests, race tests and vet. Both clients contain the
same canonical Linux server bundles. Mac ad-hoc signature verification passed.
Exact and public clients passed fresh site creation, starter asset parity, status,
overwrite refusal, public install/reinstall, update from 0.20.3 and already-current
checks. Normal Kirin/Mac CLIs now match the public 0.20.4 checksums:

```text
38bec9013fa3c224648b2c84dc63896b5982042efe6d43eecf1420c3dbea8760  flea-darwin-arm64
05698c04913aca2839ab657a6153ac24dc8410ee41f2bab88786647eea2d256c  flea-linux-amd64
```

The validated immutable download catalogue was made live with a service restart.
The compatible shared executable stays at 0.20.3; this asset-only correction
requires no executable upgrade. Protocols remain 28 and 2. All 29 HTTPS hosts
retain expected statuses, redirects and indexing; live content, host policies and
the other 19 sites' source trees match the baseline. Old 0.20.3 downloads remain
unchanged. Private release logs and previous CLIs are in `fleacms-releases/0.20.4`
on Kirin and Mac. No previews were started or customer files replaced.

Tag pipeline 798 was pending native runners; completed native checks supplied
publication evidence. Existing native ARM64 server, clean-Mac/notarization,
installer packaging and real-email verification limits remain.
