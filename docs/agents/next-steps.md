# FleaCMS V1 final sprint

Agreed September 28, 2026. V1 is ready for other developers to download and use
without Sean's help. Release when the checks pass, with no fixed deadline.
This is the working plan, not a claim that planned commands or releases exist.
Use [checklist.md](checklist.md) for manual results, [CHANGELOG.md](../../CHANGELOG.md)
for completed changes, and [README.md](../../README.md) for current usage.

## Theme removal follow-up

Before updating older site sources, remove their obsolete Theme toolbar button
while preserving custom layouts. Existing hosted toolbar partials have not been
rewritten by the shared executable rollout. Preview noindex and Hub branch
workflow verification are separate planned tasks. Theme removal is included in
public 0.15.1. Native AMD64 build/full-suite, clean-Mac and notarization
acceptance remain open.

## Current release status (October 18, 2026)

### 0.17.0 release candidate (local preparation)

The local 0.17.0 release candidate is prepared on `main` at `f09d877`. It records
version `0.17.0`, deployment protocol 26 and app wire protocol 2, and moves the
Unreleased changelog entries into the dated 0.17.0 section. It is **not built,
tagged, published, installed or deployed**. Local full source checks are recorded
in [0.17.0 verification](release-0170-verification.md).

The next phase is the lead's candidate commit, then a canonical bundled build
from that exact commit so VCS metadata is correct, followed by native Mac/Linux
checks, immutable download staging and the protocol 25 to 26 server rollout. Do
not claim public or native acceptance before those are recorded.

The integrated custom apps are described in [custom apps
integration](custom-apps-integration.md); preserved pre-reconcile source remains
on `archive/custom-apps-before-reconcile-20261018` and in the private backup.

### Public 0.16.1

Public 0.16.1, the shared server and installed Mac/Linux CLIs are verified.
The literal-brace fix is on `release/0.16.1` / immutable `v0.16.1` (`57d8378`),
based on published 0.16.0. All 18 configured HTTPS hosts are healthy; content
hashes are preserved and recovery is clear. Native Mac/ARM64 Linux full tests,
vet, race and bundled builds passed, plus native artifact/install/update smoke
on all advertised platforms. Native AMD64 full-suite/build, clean-Mac and
notarization remain open. See [release evidence](release-0161-verification.md).
The older dirty local framework checkout and unrelated main skill work are
preserved. No source merge into main is implied by the release branch/tag.

## Historical release status (October 13, 2026)

- Public 0.16.0 is published from `1c85aae` / `v0.16.0`; website `a49e002`,
  activation `20261013-221944-d1f6169314346e36fd5ab5c7`. Linux ARM64 and M5 Mac
  normal CLIs are installed from the verified public bytes. Includes detected
  hostnames/Tailscale peers in LAN mode, automatic default dev-port selection
  and server-update progress. Host/port, peer, Origin, CSRF and session checks
  remain intact. Native Mac/ARM64 Linux full tests, vet, race and bundled builds,
  relevant named-host/preview browser race checks and public install/update
  checks passed. Actual M5-to-asahi access and wrong Host/port rejection passed.
  All 12 HTTPS domains are healthy, content/media/CSS and the latest unrelated
  deployments are preserved. Server stays 0.15.1 / protocol 25; recovery is clear.
  Native AMD64 build/full-suite, clean-Mac/notarization and a pre-existing
  navigation browser timing check under race remain open. See
  [release evidence](release-016-verification.md). Draft MR 8 records integration;
  no source branches were merged into main.

- Public 0.15.1 is published from source/tag `d5958ac` / `v0.15.1`; website
  `d6cd9c9`, activation `20261013-213030-9a92d56ccb61b7b62b20a800`.
  The installer/update channel and download links agree. Native Mac/ARM64 Linux
  full tests, race, vet and bundled builds passed. Public install/reinstall,
  update from 0.13.0, already-current and fresh-site/dev checks passed on all
  three platforms. AMD64 was cross-compiled and smoke-tested unprivileged;
  native AMD64 build/full-suite, clean-Mac and notarization remain open.
  Server remains 0.15.1 / protocol 25; all eleven HTTPS domains are healthy,
  content/media/CSS and unrelated deployments are preserved, recovery is clear.
  No source merges occurred. See [release evidence](release-0151-verification.md).

- Private Blog rollout 0.15.0 / protocol 25 completed from `e690ab7`; the matching
  local Linux ARM64 CLI is installed. Seven backups and the old executable remain
  at `/var/backups/flea-blog-015-20261013`. All eleven HTTPS homepages are healthy,
  configuration/content/media/CSS fingerprints match, and recovery is clear.
  [MR 6](https://gitlab.codepilot.com/sean-schertell/flea/-/merge_requests/6) is
  draft/open; that private rollout did not merge, tag or publish public artifacts.
  Its public channel was 0.13.0; the subsequent 0.15.1 publication is recorded above. See [Blog verification](blog-verification.md). Existing sites need
  explicit Blog template/config adoption; custom layouts were not changed.


- Private shared-server rollout 0.14.0 / protocol 24 completed October 13 from
  framework `beb87ab`; the local Linux ARM64 CLI is also 0.14.0. The source branch
  is integrated into main through MR 5. During that private rollout, public downloads
  stayed on 0.13.0 and no release tag or artifacts were published. The owner approved one-time removal
  of 11 obsolete theme mappings across seven sites after verifying every saved
  theme matched source CSS. Root-only backups retain the original config,
  executable and seven site backups in
  `/var/backups/flea-theme-removal-014-20261013`. The standard `flea update-server`
  path completed, active binary hash matches the verified bundle, all eleven
  HTTPS hostnames respond successfully, content/CSS hashes and routing/TLS/release/
  download/commit mappings are preserved, and recovery journals are clear.
  Full tests/vet/race, bundled build, selected Chrome editor/media/reload/LAN
  checks and fresh-site checks passed on native Linux ARM64. Staging tests used
  a task-local SHA-256 adapter for absent shasum. Live email and interactive
  production admin editing were not exercised in this rollout.

- Version 0.13.0 is published: framework `bd668a1`, tag `v0.13.0`; website
  `1f8d4bf`, activation `20261012-182752-488ee26bcaf1fb17ddf715ec`. Status now
  prints local information first and automatically compares the current branch's
  saved deployment with terminal progress. Native ARM64 Linux/Mac full checks
  and public install/reinstall/update, initial-commit and dev smoke checks on all
  three platforms passed. Content/themes and routing are preserved, all hosted
  domains healthy and recovery clear. Shared server stays `0.9.0-rc.1` / protocol
  23. Native AMD64 full-suite/build, clean-Mac/notarization and native CI setup
  remain open; tag pipeline 358 is pending. Release verification was manual.

- Version 0.12.0 is published: framework `a227ff4`, tag `v0.12.0`; website
  `6ebdb00`, activation `20261012-174635-ec70336ae4e81090c4bc0c21`. Linux ARM64
  now has a public CLI. Native ARM64 Linux/Mac full checks and public installation,
  reinstall, update, initial-commit and dev smoke checks on all three platforms
  passed. All six sites are healthy; content/themes preserved and recovery clear.
  Shared server stays `0.9.0-rc.1` / protocol 23. Native AMD64 build/full-suite,
  clean-Mac/notarization and native CI runner setup remain open. Tag pipeline
  357 is pending; this release used manual verification.

- Version 0.11.0 is published from framework `9759a51`, tag `v0.11.0`; website
  `26ed807`, activation `20261012-164409-6fcf3b6569f135a51848d882`. Public Mac ARM64
  and unprivileged Linux AMD64 install/reinstall/update, initial-commit and dev
  smoke checks passed. All six sites are healthy, live content/themes preserved;
  shared server remains `0.9.0-rc.1` / protocol 23. Mac full checks passed; AMD64
  was cross-compiled and smoke-tested natively. Native AMD64 full-suite/build,
  clean-Mac and notarization remain open. Linux ARM64 CLI support is excluded.
  Tag pipeline 356 remains pending because native runner tags are not configured;
  this release used manual verification.

- Version 0.10.0 is published. Framework `d6a7f7c`, tag `v0.10.0`; website
  `b95eabb`, activation `20261005-203837-d9261af00e49e2d96f2fd52f`. The CLI
  renames `flea server sites` to `flea server list` and prompts for the server
  login when the SSH target is omitted. Public Mac/Linux hashes, fresh install,
  reinstall, 0.9.0-to-0.10.0 update and already-current checks passed on Mac ARM64
  and native unprivileged Linux AMD64. All four sites are healthy, content and
  themes preserved and recovery clear; the shared server remains `0.9.0-rc.1` /
  protocol 23. Native Linux full-suite, clean-Mac and notarization remain open.

- Private server rollout `0.9.0-rc.1` completed October 5, 2026 from framework
  `3447cb7`, using the standard bundled build. The shared server runs protocol
  23; Atlas `b2f768a` and website `7ef4e6e` are deployed with Contact. Bookworm
  is unregistered with files retained. All remaining sites are healthy, content
  and themes preserved, and recovery clear. Contact CSRF/validation/no-email
  success checks passed; real delivery was not tested in this rollout. Public
  downloads and the installed CLI remain 0.8.1; no RC tag or artifacts published.
  See the deployment runbook for activation IDs and recovery evidence.

- Version 0.8.1 is published. Framework `209b5d3`, tag `v0.8.1`; website
  `9fd1397`, activation `20261003-230705-de4bfab5fbec8834e722a460`.
  Runtime/media ignore rules append only when missing, preserving existing bytes.
  Package/full tests, vet, bundled build and repeated dev-start checks passed.
  Public Mac/Linux hashes, fresh installs, reinstalls and updates from 0.8.0 passed;
  Sean's installed CLI is the public 0.8.1 binary. The shared server remains 0.8.0.
  Existing distribution acceptance limits and Bookworm toolbar caveat below remain.

- Version 0.8.0 is published. Framework `c8a1d96`, tag `v0.8.0`; website
  `3188000`, activation `20261003-171440-ca76afb853e32d0d0309c6ea`.
  The sliding edit row and Theme panel passed browser and reduced-motion checks.
  Public Mac/Linux downloads, fresh installs and updates from 0.7.1 passed.
  Naves and Kent received toolbar-only deployments, preserving live content/theme.
  Bookworm was excluded at Sean's request: its public pages work, but the old
  toolbar remains incompatible with the shared 0.8.0 editor. Update its partial
  before using that editor again. Existing distribution acceptance limits remain.

- Version 0.7.1 is published: website, download links, update channel and shared
  server executable agree. Public Mac/Linux checksums, fresh installs and updates
  from 0.7.0 passed, with Linux checks running unprivileged. Local tests, vet,
  bundled build and fresh-site validation passed. Routine HTTPS progress was
  verified during the real deployment; certificate checks remain unchanged.
- Versioned downloads now stream from a separate administrator-owned disk store,
  preserving all old URLs and hashes. Website assets fell from 247.7 MiB to
  0.63 MiB; the final archive upload was 575 KB. Keep the 256 MiB archive limit.
  Follow docs/agents/download-hosting.md; back up the store independently of site backups.
- The 1 GB host needed Sean's approved `GOMEMLIMIT=320MiB` service override to
  complete activation. The 0.6.1 upgrade additionally needed an approved 1 GB
  swap file, persisted in /etc/fstab. Apply the memory target to separate upgrade
  and deployment processes too. Preserve these settings during maintenance. Native
  Linux full-suite/build verification, clean-machine Mac acceptance and
  notarization remain open; this is still a test release.
- The older sprint notes below predate the published website and 0.5.0 downloads.
  Use the separate website's RELEASES.md for recorded publication evidence;
  do not treat the earlier "hosting pending" notes as current release status.

## Agreed scope

Separately approved mini-app work is tracked in [the app plan](planning/apps.md).
The approved managed Contact batch now implements launch, fixed render-only
routes, private email capability, redirects/flash and replacement of the old
contact route/config. See apps/contact/README.md and CHANGELOG.md.
The [isolated Linux Contact check](benchmarks/2026-10-04-contact-apps.md)
passed startup, child replacement, graceful shutdown and abrupt parent-exit tests.
Thirty tiny Contact-enabled sites peaked at 55.43 MiB sampled total PSS; the
test unit peaked at 61.6 MiB. Its 221 OS threads also warrant a task-budget review.
An isolated real Resend send using the app sender override also passed provider
acceptance, redirect and flash checks. Sean confirmed inbox receipt, the app's
sender override and Reply-To.
Next acceptance: HTML/text appearance and default-sender checks, representative mixed workloads,
longer soaks and deployment peaks. No deployment is
authorized by this note. Managed custom app packages and app-owned sessions are
implemented in the local 0.17.0 candidate, not yet published, with a stable
per-site runtime storage root and service-readable package extraction. The
deployment protocol is now
`FLEA_DEPLOY=26` (`internal/release/protocol.go`) and the app wire protocol is
`2`; an older server rejects a package-bearing archive, so shipping requires a
matching server release. Remote services, config-declared read routes,
databases, storage moves and ZJAX remain separate work. No server release or
deployment is authorized by this note.

Capacity follow-up: the [isolated 1/10/30-site VPS probe](benchmarks/2026-10-04-multisite.md)
passed on a 1 GB host with synthetic traffic and overlapping read-only validation.
Still measure representative client workloads, editor/uploads, longer soak times,
and actual deployment peaks. Include descriptor limits, total RSS and OS cache,
not just heap. The first Contact measurements are recorded above; heavier apps
and real client workloads still need their own capacity checks.

- Keep the existing CMS feature set. Focus on defects, usability, documentation,
  installation, updates, website design and release verification.
- Include the agreed local email-preview addition: new sites preview outgoing
  mail in an email-style browser window without first-run Resend prompts.
- Source stays private. Binaries are free to download and use for now, with no
  account or license key. Do not describe FleaCMS as open source.
- Primary installation is one copy-and-paste shell command from fleacms.com,
  ideally without sudo. Homebrew and native installer packages are deferred.
- CLI release targets: Apple Silicon macOS and AMD64/ARM64 Linux, all with
  verified public downloads. Intel Mac remains deferred.
- Support single-server Ubuntu deployments. Multi-developer permissions,
  clustering, high availability and separate per-site runtimes are deferred.
- Add `flea update` for the local CLI and its bundled server executables.
  Rename `flea upgrade` to `flea update-server`, including
  `flea update-server --recover`, with no old-command alias.
  Local updates only remind about server updates; never trigger one automatically.
- Build and run fleacms.com on FleaCMS. Use Mobbin references and the requested
  design skill to develop a professional, intentional visual design.
- Include a developer quickstart and a short illustrated owner guide.
- Test with a newly generated site. Sean has a disposable Ubuntu server and
  will do real-device checks; the agent handles automated and available browser
  checks. Existing checklist answers must remain intact.

Local code now provides `flea update` and `flea update-server`, with no upgrade
alias, plus the installer. Public download hosting is not configured or verified.
The local website scaffold and packaging scripts are not published releases.

## Sprint order and acceptance

Work through these stages in order. Website reference research can start while
release infrastructure is being prepared, but screenshots and final instructions
must match the finished product. Keep one agent unless Sean authorizes delegation.

### 1. Resolve the recorded product defects

- [x] Simplify dev diagnostics: remove the eyebrow, actions and footer below the
  fix hint; use the regular 404 template in dev.
- [x] Report unclosed conditionals accurately with useful source locations.
- [x] Keep Theme brand colors closed by default.
- [x] Retest the corresponding checklist failures without erasing original notes
  (Go regressions and Chrome Theme check; broader manual acceptance remains below).
- [ ] Address any additional blocking defects discovered during the sprint.
  Scope new features separately rather than extending V1 during QA.
- [ ] Verify in-place page Save in Safari and on a real mobile device, then include
  it in the next patch release. Zjax adoption remains a separate decision.
- [ ] Investigate Chrome's native-navigation browser check reporting
  "Missing page transition name". The in-place Save, media, collection and dev
  reload browser checks pass; this separate transition check still needs review.
- [ ] Review the Inter/forest-green boilerplate on real devices. Chrome desktop
  and mobile access-page checks passed; the theme is explicitly synced to Bookworm.
- [ ] Review clean default, `new --starter` / `-s`, fictional team and native
  mobile navigation. Verify reset's warning/recovery copy in a disposable local
  site; never use reset as routine cleanup or assume it clears remote content.
- [ ] Sean: review the updated Edit toolbar, green toggle, thicker green rounded
  dashed boundaries and collection
  dragging on Bookworm, including a real-device check. Automated verification
  and test-site rollout are complete; visual acceptance remains open.
- [ ] Confirm the native navigation fade on Bookworm without browser automation
  attached. The early head opt-in clears the observed opt-in error and the local
  native-transition test passes; automated live checks now encounter viewport
  changes. Keep Safari and real-device acceptance open.

Done when targeted regressions and the required implementation checks pass and
the recorded failures have a verified retest result.

### 2. Deliver installation and updates

- [x] Rename the server command, help, errors, tests, recovery instructions,
  starter guidance and website references together. Preserve its explicit
  confirmation, all-site restart warning, validation and recovery behavior.
- [x] Implement a one-line installer with OS/architecture detection, verified
  HTTPS downloads, checksum checks and a writable user installation directory.
  Decide PATH setup deliberately; do not silently overwrite shell configuration.
- [x] Implement `flea update` using the same release metadata and verification.
  Preserve the installed executable on failed downloads or validation and replace
  it safely only after checks succeed. Report already-current versions clearly.
- [x] Explain that local updates include bundled servers but remote servers are
  changed only by an explicit `flea update-server`.
- [ ] Provide versioned public artifacts and release metadata while keeping the
  source repository private. Select hosting and the exact installer URL.
- [ ] Build and test the macOS ARM64 and Linux AMD64 CLIs, each with required
  Linux server bundles. Assess the incremental effort for the optional platforms.
- [ ] Verify fresh installation, PATH, version, site creation/dev/status,
  reinstall, local update and failed-update preservation on both required platforms.
- [ ] Verify the downloaded Mac artifact on a clean machine, including platform
  trust/signing/notarization requirements for the chosen delivery method.
- [x] Update [RELEASING.md](releasing.md), native Mac/Linux CI matrix and user docs.
  The manual publish-stage job prepares artifacts only; actual hosting is pending.

Done when a developer can install and update from the documented command without
Go, source access, an account or administrator privileges on the supported setup.
Do not claim optional platforms work solely because they cross-compile.

### 3. Verify the real server workflow

- [x] Obtain the disposable target, SSH identity, test domain/DNS and mail setup
  when needed. Keep credentials outside repository documents.
- [ ] Start with a fresh generated site and install on a clean supported Ubuntu host.
- [ ] Verify permissions, DNS, public HTTP/HTTPS, certificate issuance, redirects,
  certificate reuse and private-path isolation.
- [ ] Deliver real login and contact email; check custom email templates and Reply-To.
  Login delivery/sign-in and contact email receipt/rendering confirmed by Sean.
  Custom-template variations and explicit Reply-To inspection remain.
- [ ] Test the default local email-preview onboarding on a fresh site, then turn
  preview mode off to verify real delivery. Previews do not satisfy live-mail QA.
- [ ] Exercise owner editing, media upload/playback and theme changes remotely.
  Mobbin-inspired thumbnail picker deployed to the disposable host. New markup
  verified publicly; uploaded photo bytes remain intact. Sean's UI review pending.
- [ ] Redeploy source and prove live text, metadata, collections, theme and media survive.
  Live homepage text survived normal redeployment and was retrieved by pull on
  September 29. A saved Theme color also survived byte-for-byte and theme-only
  pull retrieved it. Metadata, collections and media still need equivalent checks.
- [ ] Verify pull, fresh-clone media retrieval, explicit synchronization, conflicts
  and missing-file failures.
- [ ] Exercise rollback, backup/restore, server updates and interrupted-operation
  recovery on the disposable host. Verify restoration of a usable prior state.
  Normal backup, marker-based rollback, restore, same-version server update and
  pull passed on September 29. Interrupted/failure and version-change checks remain.
- [ ] Record exact tested versions/platforms and outcomes. Track certificate
  renewal monitoring separately; initial issuance does not prove long-term renewal.
- [ ] Exercise `flea server list` and `flea server remove` on the disposable host:
  multi-site/aliases, exact-name cancellation, stale-review rejection including
  a commit-only change, restart/probe failure rollback and interrupted recovery
  with `flea-server -recover`. Confirm the shared-service restart briefly affects
  other sites, that removing the last HTTPS hostname disables automatic HTTPS
  with cached certificates retained, and that redeploying the same name does not
  silently restore retained content. The shared host was upgraded to protocol 23
  on October 5; live inventory passed. Bookworm removal used a separately
  approved one-time protocol-22 maintenance script, not the new removal command,
  so it does not establish live acceptance of that command. No automatic
  migration or aliasing is added.

Done when the documented developer workflow and recovery paths work on the real
target, with no unexplained data loss, permission failures or manual rescue steps.

### 4. Finish browser, security and owner acceptance

Sean accepted the browser/UI pass for now on September 29 and requested moving
to live workflow verification. Do not block that work on further visual polish.
The detailed unverified device/browser items below remain recorded, not claimed
as individually tested.

- [x] Run automated tests, vet, bundled build, race suite and Chrome editor check
  for this sprint batch. Rerun applicable checks after further implementation.
- [ ] Verify current Safari, Chrome and Firefox: sign-in, selection/paste,
  formatting, dialogs, Page details, Theme and Save/Cancel. Review Theme's fixed-footer
  Save, header X, Escape confirmation and animated closing on a real device.
- [ ] Check slow-connection navigation, initial toolbar/icons, dirty-draft
  protection, narrow screens, custom layouts and repeated/shared regions.
- [ ] Sean: test a real phone, including editing, photo uploads, collection drag/
  reorder, empty collections and video playback. Record device/browser versions.
- [ ] Check keyboard navigation, focus, labels, contrast and reduced motion.
- [ ] Review authentication/session handling, Origin/CSRF/CSP, upload bounds,
  parser/file/archive confinement, site isolation and update artifact trust.
  Add focused regression/fuzz/resource-limit tests where the review finds gaps.
- [ ] Exercise the shipped brochure-import guidance on a representative disposable
  site, including responsive rendering, links, assets and validation.

Done when there are no known release-blocking security, data-integrity or core
workflow defects. Record remaining limitations explicitly rather than marking
untested checks as passed. Sean owns real-device acceptance; the agent records
automated and accessible browser evidence.

### 5. Design and build fleacms.com

Agreed structure: exactly two public content pages, homepage and `/docs`, using
zjax.dev's concise example-led structure with a fresh, bright visual style.
Documentation, owner guidance, limitations, release notes and support are anchored
sections of `/docs`, not separate pages. Authentication/error templates remain.
Build this as a separate Flea site/repository at `/Users/sean/Projects/fleacms.com`,
not inside the framework. It has not been created yet; see `handoff.md` for setup,
references and the homepage-review checkpoint. Preserve the existing scaffold
until useful content has been ported and verified.

- [x] Inspect the existing `website/` scaffold and preserve useful content.
- [ ] Locate/enable and read the requested design skill before design work.
  Mobbin search tools are available in this session; a matching design skill was
  not found in installed skills. This does not block the other sprint stages.
- [x] Use Mobbin to select a small set of relevant developer-product references.
  Inspect actual screens, retain canonical reference links, and explain which
  typography, layout and navigation choices suit Flea. Do not copy brand assets.
  Proposed direction and reference links: `website/design-direction.md`.
- [ ] Present a concrete homepage visual direction for Sean's review before
  extending it across the website. Preserve the selected Skip logo.
- [ ] Establish coherent typography, spacing, color, responsive layouts and
  restrained motion. Use actual Flea screenshots/examples and specific copy;
  avoid generic feature-card filler, invented testimonials or unsupported claims.
- [ ] Build in Flea using the project’s Prewind conventions. No separate frontend
  framework or build system. Review rendered desktop and mobile pages.
- [ ] Cover the product, installation, developer quickstart, owner guide,
  limitations, release notes and a way to report problems.
- [ ] Write the developer path from install through new/dev, Git, setup/deploy,
  HTTPS, mail, content preservation, backup and both update commands.
- [ ] Write a short owner guide with screenshots: sign-in, text/Page details,
  images/video, collections, Theme and Save/Cancel.
- [ ] Show the editor through real screenshots or a recorded walkthrough.
  A publicly writable live demo is not required for V1.
- [ ] Verify copy-paste commands, downloads, links, forms, metadata, accessibility
  and responsive behavior against the release candidate.

Done when Sean accepts the visual design and a new developer can follow the
published instructions without private repository access or undocumented help.

### 6. Release and operate

- [ ] Resolve the remaining release choices below and freeze the support matrix.
- [ ] Prepare the V1 version and dated release notes after acceptance passes.
- [ ] Verify immutable public downloads, checksums, installer and update metadata,
  without exposing private source, tokens or credentials.
- [ ] Deploy fleacms.com on FleaCMS with HTTPS; verify the actual public website
  and downloaded artifacts, not only the local build.
- [ ] Run the public quickstart once more using the downloaded release.
- [ ] Record the support/reporting route, backup procedure and certificate/host
  monitoring responsibilities before announcing availability.

Release when all required checks pass. Publishing, tags and live changes happen
as explicit execution steps with the selected destination, not merely because
they appear in this roadmap.

## Details to resolve when their stage starts

### Planned LAN QR code

- [ ] Print a terminal QR code for the LAN URL so a phone can open the dev site
  without typing its address. The phone must be on the same reachable network.
- Decide whether it appears automatically with `flea dev --lan` / `-l` or requires
  an explicit `--qr` option; handle multiple network addresses sensibly.
- Encode only the site URL, never credentials or a sign-in bypass. Preserve the
  existing LAN Host/peer gate and admin authentication. Keep the plain URL visible.
- This is planned work, not implemented or scheduled for a release.

### Planned Markdown support

- [ ] Support `.md` pages out of the box, using the same layouts, routing,
  Svelte-like template syntax and editable snips as HTML pages.
- Allow natural HTML mixed with Markdown, including HTML elements with snip
  attributes, without requiring special wrappers or an extra build step.
- Define Markdown/template/HTML parsing order, code-fence escaping, `.md`/`.html`
  route collision handling and snip extraction/save behavior before implementation.
  Preserve contextual escaping, rich-content sanitization and source formatting.
- Verify conditions, loops, partials, inline HTML and snips together, including
  dev reloads and repeated owner saves. Choose a parser only after reviewing these
  requirements. This is planned work, not yet scheduled for a release.

### Planned no-mail contact fallback

- Keep local `dev.email_preview` working without Resend. It is not a production fallback.
- Compute provider-neutral contact availability server-side from recipient and
  usable sender configuration; expose only a boolean, never credentials or errors.
- When production mail is unavailable, replace the form with a styled unavailable
  notice. Offer a mailto link only from a separately configured, explicitly public
  contact address. Never expose admin.emails or apps.contact.email.to automatically.
- Give owners/developers setup guidance, not visitors. Missing mail must not bypass
  OTP authentication or silently capture/store submissions.
- If delivery fails after a visitor has typed, preserve their fields and show an
  honest retry message in the site layout, not a success page or bare error.
- Verify no-key, no-recipient, preview, configured delivery and transient failure
  states. The availability notice remains planned; safe 503 field redisplay is now implemented by Contact.

## Remaining release decisions

Ask one question at a time; these need not delay independent work.

- Confirm public release storage and clean test access for both required platforms.
  Implemented default: ~/.local/bin, printed PATH instructions without profile
  edits; installer URL https://fleacms.com/install.sh and /releases/ metadata.
- Disposable server address/SSH user, test domain and mail credentials.
  SSH verified at root@fleacms.com. Authorized reset completed with the previous
  installation retained in a root-only backup; current protocol 22 service is
  healthy. Fresh acceptance site is deployed at bookworm.fleacms.com with verified
  public HTTPS. Approved contact: sean@codepilot.com; saved mail settings restored
  without exposing credentials. Baseline backup passed. See checklist.md for
  exact evidence; real delivery and the remaining recovery tests are still pending.
- Final fleacms.com hosting destination and DNS access. Do not assume it is the
  disposable server.
- Availability/name of the requested design skill and the chosen visual direction.
- Binary distribution terms and third-party notices for a free, private-source
  release; support/contact destination. Free use does not imply an open-source license.

## Deferred beyond V1

- Homebrew tap and native installer packages; their existing scaffolding can remain.
- Intel Mac support.
- Cropping/focal points, responsive derivatives, new image formats, video collection
  fields, transcoding/provider integrations and retention/pruning UI.
- Per-edit history, general field schemas/forms and expanded rich-text features.
- Multi-developer deployment authorization, per-site grants/revocation and audit
  design. Do not change SSH privileges before that design is agreed.
- Components beyond partials, new slot/input syntax and broader template abstractions.
- Clustering, high availability, independent site runtimes and service/schema migrations.
- Deployment-setting consolidation, remote logs and automatic release retention.
- Further legacy-format cleanup under the no-backward-compatibility policy.
- Public writable editing demo and unrelated new feature requests.

Maintain this as the remaining work plan. Remove completed task detail as it moves
to the changelog; retain acceptance evidence in the checklist or linked QA records.

## Blog task

Minimal Blog implementation is on `feat/minimal-blog` in its isolated worktree.
Local implementation and checks are complete; see
[Blog verification](blog-verification.md). The 0.15.0 commit, push and private server rollout are complete; results are
recorded in the verification document. Blog and the subsequent setup skills are
now available in public 0.15.1; see [release evidence](release-0151-verification.md).
Optional Blog source adoption must preserve each site's custom layouts. Native
AMD64 build/full-suite, clean-Mac and notarization acceptance remain open.
