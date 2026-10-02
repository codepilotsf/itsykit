# FleaCMS V1 final sprint

Agreed September 28, 2026. V1 is ready for other developers to download and use
without Sean's help. Release when the checks pass, with no fixed deadline.
This is the working plan, not a claim that planned commands or releases exist.
Use [checklist.md](checklist.md) for manual results, [CHANGELOG.md](CHANGELOG.md)
for completed changes, and [README.md](README.md) for current usage.

## Current release blockers (October 1, 2026)

- Sean confirmed 0.6.0 as the next release. The framework HEAD and tag match;
  rebuilt candidates and checksums are staged in the separate website checkout
  at `/Users/sean/fleacms-website/site/releases/0.6.0/`. Its installation links
  and update pointer now target 0.6.0, as requested. Superseded candidates are
  preserved outside the public tree; unrelated website changes remain intact.
- Verify matching release artifacts on the required platforms, publish immutable
  versioned downloads and verify public checksums before publishing the prepared
  update pointer and website copy.
  The local 0.6.0 tests, race suite, vet, Chrome sign-in/editing checks, bundled
  build and clean/starter site status checks passed. These do not establish
  native Linux or live server acceptance for 0.6.0.
- The older sprint notes below predate the published website and 0.5.0 downloads.
  Use the separate website's RELEASES.md for recorded publication evidence;
  do not treat the earlier "hosting pending" notes as current release status.

## Agreed scope

- Keep the existing CMS feature set. Focus on defects, usability, documentation,
  installation, updates, website design and release verification.
- Include the agreed local email-preview addition: new sites preview outgoing
  mail in an email-style browser window without first-run Resend prompts.
- Source stays private. Binaries are free to download and use for now, with no
  account or license key. Do not describe FleaCMS as open source.
- Primary installation is one copy-and-paste shell command from fleacms.com,
  ideally without sudo. Homebrew and native installer packages are deferred.
- Required developer platforms: Apple Silicon macOS and x86-64 Linux.
  Add Intel Mac and ARM64 Linux only if building and verifying them takes little
  extra work; otherwise omit them from advertised V1 support.
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
- [x] Update [RELEASING.md](RELEASING.md), native Mac/Linux CI matrix and user docs.
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
  contact address. Never expose admin.emails or contact.to automatically.
- Give owners/developers setup guidance, not visitors. Missing mail must not bypass
  OTP authentication or silently capture/store submissions.
- If delivery fails after a visitor has typed, preserve their fields and show an
  honest retry message in the site layout, not a success page or bare error.
- Verify no-key, no-recipient, preview, configured delivery and transient failure
  states. This is planned, not implemented; existing delivery behavior is unchanged.

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
- Intel Mac/ARM64 Linux support if verification is more than a small addition.
- Cropping/focal points, responsive derivatives, new image formats, video collection
  fields, transcoding/provider integrations and retention/pruning UI.
- Per-edit history, general field schemas/forms and expanded rich-text features.
- Multi-developer deployment authorization, per-site grants/revocation and audit
  design. Do not change SSH privileges before that design is agreed.
- Components beyond partials, new slot/input syntax and broader template abstractions.
- Clustering, high availability, independent site runtimes and service/schema migrations.
- Deployment-setting consolidation, remote status/logs and automatic release retention.
- Further legacy-format cleanup under the no-backward-compatibility policy.
- Public writable editing demo and unrelated new feature requests.

Maintain this as the remaining work plan. Remove completed task detail as it moves
to the changelog; retain acceptance evidence in the checklist or linked QA records.
