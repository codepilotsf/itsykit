# FleaCMS 0.20.5 verification

Published November 9, 2026 after approval of the navy filled flea on a light
background, including the explicit request to restore the website's light header.
Primary logos, icons, starter/framework branding and development diagnostics use
navy `#142b43`. Icons use fixed light `#f5f9ff` backgrounds in all browser themes.
The website header, footer and admin header use the positive wordmark on light
surfaces; the social image has a matching light brand strip. The hero layout,
mascot geometry, Nunito outlines and customer branding are preserved.

Tested source `00b2e36`, immutable `v0.20.5`, framework MR !36 merge `5084a26`.
Source archive SHA-256 matches on Linux/Mac:
`55909f2680b71aba8485609e7ce3ad9c0bdec077fc68eb4358b574a98917fbb2`.
Website MR !22 merged as `53eb18e` and activated
`20261109-165104-562591a96cda87412e103bef`, including the 0.20.5 update channel.

Desktop, phone and 320px browser checks passed for homepage/docs/admin branding.
The narrow header keeps navigation separated and Get started on one line.
The existing sticky divider/translucency behavior is preserved with background
blur to keep underlying text from interfering. Header/footer colors, logo loads,
CTA contrast, icon opacity/palette, social image and all live asset bytes were
verified. No editable content, authentication or CSRF contracts changed.

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) canonical builds and
release scripts passed full tests, race tests and vet. Both clients contain the
same canonical Linux server bundles. Mac ad-hoc signature verification passed.
Exact/public clients passed current-directory creation, initial commit, starter
parity, status, overwrite refusal, fresh install/reinstall, update from 0.20.4
and already-current checks. Normal Kirin/Mac CLIs match the public manifest:

```text
c3870e6d1605da19674fd8ea84a5d795fa9b5703b8a9118f57e31b59f3a7ef34  flea-darwin-arm64
6be769f644468f56af9189432facb75b3fb896de463dbf73827357b808f26dbb  flea-linux-amd64
```

The immutable catalogue was validated and made live before advancing the channel.
The compatible production executable remains 0.20.3; protocols remain 28 and 2.
Live content, host policies and other 19 site trees match the task baseline.
All 29 HTTPS hosts retain expected responses and indexing. Older downloads and
existing custom branding are preserved. Task-owned development preview stopped.
Private source, logs, previous CLIs and provenance are in `fleacms-releases/0.20.5`
on Kirin and Mac. Tag pipeline 799 was observed pending runners; native manual
checks supply release evidence. Existing native ARM64 server, clean-Mac/notarization,
installer packaging and real-email acceptance limits remain.
