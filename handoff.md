# fleacms.com website handoff

## Blog documentation deployment — October 13, 2026

Website source `54815a7` on `docs/blog-setup`, based on current main `7e55082`.
[Draft MR 2](https://gitlab.codepilot.com/sean-schertell/flea-website/-/merge_requests/2)
remains open. Deployed an isolated main checkout of the exact committed snapshot
to the existing fleacms.com destination; original working checkout/branch was
preserved and no source merge was performed.

Activation: `20261013-194213-97642e56ad639a7acdb3ac8a`. Blog setup is public at
https://fleacms.com/docs#blog and linked from the homepage and docs contents.
The guide covers config/template adoption, editing, agent JSON, tokens and sync.
It labels the 0.15.0 / protocol 25 requirement and the public 0.13.0 channel
accurately. No release artifacts or channel files changed.

Validation: `flea status`, rendered anchors/literal tokens, highlighted JSON and
HTML examples, desktop/mobile browser checks and live HTTPS verification passed.
At 375px there is no horizontal page overflow. Live JSON examples parse as shown.
After deploy, `flea status` reports code and content matching the committed source.
All eleven HTTPS homepages are healthy, recovery is clear, and website saved
content/media/CSS plus all unrelated-site fingerprints are preserved. Runtime
remains 0.15.0 / protocol 25. No real email was sent.

Private backup: `/tmp/flea-website-blog-rollout/website-before.flea-backup`.
Before/after fingerprints are in that private directory. The deployment used the
existing 320 MiB admin-process memory target; no service-policy change occurred.
No preview process remains. These are website documentation checks, not new
native binary-release acceptance results.


## Current release: 0.13.0 published October 12, 2026

MR !4 merged; framework source/tag `bd668a1` / `v0.13.0`, Go 1.27.1. Status now
prints local information first, then automatically compares the current branch's
configured deployment with terminal progress. Code, content and theme are distinct;
change direction needs a scoped pull baseline. No remote flag or status mutation.

Website `1f8d4bf`, activation `20261012-182752-488ee26bcaf1fb17ddf715ec`, publishes
homepage/docs/latest.txt 0.13.0 after all three immutable HTTPS downloads were
verified. Native Mac/ARM64 Linux full checks and all three platforms' public
install/reinstall/update, initial-commit, status and dev checks passed. The final
ARM64 CLI's animated remote status check passed. No test dev processes remain.

Shared server remains **0.9.0-rc.1 / protocol 23**. All hosted domains are healthy,
content/themes/routing hashes are preserved, recovery clear and memory/swap policy
unchanged. Existing user CLIs and published old artifacts are unchanged. Native
AMD64 full-suite/build, native CI runner setup (pipeline 358 pending), clean-Mac,
notarization and live non-root password/sudo status verification remain open.

Private records: framework worktree `/home/sean/Projects/flea-release-013`, Mac
build `/Users/sean/flea-release-013.w6iFK0`, native AMD64 isolated checks
`/var/tmp/flea-013-native.soNcFH1s`, artifact/download archive `/tmp/flea-013-public`.
Website backup `/tmp/flea-013-website-before.flea-backup`; before/after state hashes
`/tmp/flea-013-before-state.json` and `/tmp/flea-013-after-state.json`. RELEASES.md
records final hashes and provenance. Keep private source/backups outside site/.

## Previous release: 0.12.0 published October 12, 2026

Source/tag `a227ff4` / `v0.12.0`, Go 1.27.1. First public Linux ARM64 CLI,
alongside Mac ARM64 and Linux AMD64. Website `6ebdb00`, activation
`20261012-174635-ec70336ae4e81090c4bc0c21`; homepage/docs/latest.txt show 0.12.0.
Installer errors now identify missing versions/platforms; the same curl command
works on aarch64. Docs highlighting is preserved.

Native Mac/ARM64 Linux full tests, race checks, vet and builds passed. All CLIs
bundle identical verified Linux server executables. Public installer/reinstaller,
checksum/version, updater and already-current checks passed on all three native
platforms, including default/starter initial commits, status and dev routes.
ARM64 updated a local 0.11.0 fixture; Mac/AMD64 updated public 0.11.0. Native
AMD64 build/full-suite, native CI runner setup (pipeline 357 pending), clean-Mac
acceptance and notarization remain open. Test dev processes are stopped.

Shared server remains **0.9.0-rc.1 / protocol 23**. Six sites are healthy, live
content/themes are preserved, recovery clear and memory/swap policy unchanged.
Normal user CLIs were not replaced. Old version URLs/hashes are unchanged.
RELEASES.md records hashes and provenance; backups/build copies remain private
outside website Git.

## Previous release: 0.11.0 published October 12, 2026

Framework tag `v0.11.0` (`9759a51`). Website `26ed807`, activation
`20261012-164409-6fcf3b6569f135a51848d882`; homepage/docs/latest.txt show 0.11.0.
`flea new` creates a clean `Initial site` commit on main and prints remote
connection/push commands. Downloads remain Mac ARM64 and Linux AMD64 only.

Native Mac full tests, vet, race checks, bundled build and ad-hoc signature checks
passed. Both Linux server bundles were rebuilt. Linux AMD64 was cross-compiled
off production and smoke-tested there unprivileged. Public download hashes,
fresh install, reinstall, update from 0.10.0 and already-current checks passed
on both platforms, including default/starter initial commits, status and dev
routes. Test dev processes are stopped. The tag pipeline is pending for absent
native runners; native AMD64 build/full-suite, clean-Mac acceptance and
notarization remain open.

The shared server remains **0.9.0-rc.1 / protocol 23**. All six hosted sites are
healthy, content/themes are preserved, recovery is clear and memory/swap policy
is unchanged. No normal user CLI installation was replaced. Build copies are
isolated from the unfinished framework and website edits. See RELEASES.md for
hashes and full evidence.

## Previous release: 0.10.0 published October 5, 2026

Framework tag `v0.10.0` (`d6a7f7c`). Website `b95eabb` is deployed as
`20261005-203837-d9261af00e49e2d96f2fd52f`; the homepage, docs and `latest.txt`
show 0.10.0 and the installer is unchanged. The 0.10.0 directory was staged
beside the download store, verified and validated under the deployment lock
before the shared service restarted to load it; no journals remain. Anonymous
HTTPS downloads matched both published hashes and an old-release manifest check
passed. Public fresh install, reinstall, 0.9.0-to-0.10.0 update and
already-current checks passed on Mac ARM64 and native unprivileged Linux AMD64
with matching hashes. `flea update` replaced Sean's installed CLI, which now
reports 0.10.0. Renames `flea server sites` to `flea server list` and prompts
for the server login when the target is omitted.

The shared server executable remains **0.9.0-rc.1 / protocol 23**; it was not
upgraded, only restarted to load the new catalogue. All four active sites are
healthy, live content and themes were preserved, recovery is clear and
memory/swap safeguards are unchanged. Test dev processes are stopped. Native
Linux full-suite, clean-machine Mac and notarization remain open. See RELEASES.md.

## Previous release: 0.9.0 published October 5, 2026

Framework tag `v0.9.0` (`e599a5e`). Website `5ac7101` is deployed as
`20261005-172223-0d03e392e05da991928a068b`; the homepage, docs and
`latest.txt` show 0.9.0 and the installer is unchanged. The 0.9.0 directory was
staged beside the download store and validated under the deployment lock before
the shared service restarted to load it; no journals remain. Anonymous HTTPS
downloads matched both published hashes and an old-release manifest check passed.
Public fresh install, reinstall, 0.8.1-to-0.9.0 update and already-current checks
passed on Mac ARM64 and native unprivileged Linux AMD64 with matching hashes.
`flea update` replaced Sean's installed CLI, which now reports 0.9.0.

The shared server executable remains **0.9.0-rc.1 / protocol 23**; it was not
upgraded, only restarted to load the new catalogue. All four active sites are
healthy, live content and themes were preserved, recovery is clear and
memory/swap safeguards are unchanged. Test dev processes are stopped. Native
Linux full-suite, clean-machine Mac and notarization remain open. See RELEASES.md.

## Previous private rollout: 0.9.0-rc.1, October 5, 2026 (not a public release)

Shared server is **0.9.0-rc.1 / protocol 23**, framework `3447cb7`. Website
`7ef4e6e` was deployed as `20261005-163126-76b53507e24be701230d2e89` with Contact.
Atlas `b2f768a` was deployed; Bookworm is unregistered with all files retained.
Kent and Naves were not redeployed. All four sites are healthy and recovery is
clear. Live content and themes were preserved. Both Contact forms passed CSRF,
validation and no-email success checks; real delivery was not tested this time.
Public downloads, latest.txt and Sean's installed CLI remained **0.8.1** at that
time. No RC tag or public artifacts were published. See RELEASES.md and the
framework deployment runbook for details. The entries below describe earlier
releases.

Latest release 0.8.1: framework `209b5d3`, tag `v0.8.1`; website `9fd1397`,
activation `20261003-230705-de4bfab5fbec8834e722a460`. Fixes duplicate `.gitignore`
rules and preserves existing bytes. Public Mac/Linux installs, reinstalls and
updates from 0.8.0 passed; Sean's installed CLI and the public channel are 0.8.1.
The shared server stays at 0.8.0; no executable upgrade was needed for this CLI fix.
Recovery is clear, memory/swap safeguards are unchanged, and test dev processes
are stopped. See RELEASES.md for hashes and remaining distribution limits.
Framework APPS.md and the previous installed binary backup remain untouched.

Previous release 0.8.0: framework `c8a1d96`, tag `v0.8.0`; website `3188000`,
activation `20261003-171440-ca76afb853e32d0d0309c6ea`. Server, public channel
and Sean's installed CLI agree. Public downloads, fresh installs, updates from
0.7.1 and already-current checks passed on Mac and unprivileged Linux.
The website, Naves and Kent received the required new toolbar wrapper. Bookworm
was excluded at Sean's request: its public page works, but the old editor toolbar
is incompatible with the shared 0.8.0 runtime. No OTP mailer work was performed.
No recovery journals remain; memory/swap policy is unchanged. See RELEASES.md.
Framework APPS.md and Kent's unrelated skill edit remain untouched. Naves has
no Git remote, so its toolbar commit is local only. Earlier entries are history.

Latest patch 0.7.1: framework `f4a72ee`, tag `v0.7.1`; website `f75d63f`,
activation `20261003-151507-4d9b7bc05ab6b470b4ae679b`. Server/channel/local CLI
agree. Routine deployment's shorter HTTPS message was observed live. Public
fresh installs and updates from 0.7.0 passed on Mac/Linux. See RELEASES.md.

## Current update: October 3, 2026

0.7.0 is published from framework `1199747`, tag `v0.7.0`.
Website `a41a758`, activation `20261003-141847-a863aaed91fd5f4c8fa761b8`.
Public fresh installs/updates passed on Mac ARM64 and unprivileged Linux AMD64.
Downloads now live in `/var/lib/flea/downloads/fleacms` on the existing host,
outside this repository and website snapshots. Every retained version and 0.7.0
passed public manifest/hash checks. Local copies are at
`/Users/sean/fleacms-releases/`; prior Git history is unchanged. Keep only
latest.txt under site/releases. See RELEASES.md for final deployment evidence and
the framework's docs/download-hosting.md for future publishing and backup steps.
Do not follow older notes below that say to commit versioned binaries in this site.

Updated October 1, 2026.

Version 0.6.1 is published from framework commit `89e3472` and tag `v0.6.1`.
Current website copy and download links target 0.6.1. It adds automatic graceful
dev replacement, including LAN/port changes, without interrupting pull/reset.
Both release binaries passed replacement/crash checks on their native platforms.
Sean approved `GOMEMLIMIT=320MiB` in the service's `60-memory.conf` override;
also apply that target to separate deployment and upgrade processes. The approved
1 GB /swapfile is active and persisted in /etc/fstab. Public fresh installs and
updates from 0.6.0 passed on Mac ARM64 and unprivileged Linux AMD64. See RELEASES.md for
the active release and final publication evidence.

Social preview update: commit `200edf7`, deployed as
`20261002-005223-ea8f62a137bd69e5998efc78`. The layout has page-specific canonical,
Open Graph and large-image card metadata; `site/images/social-preview.png` is
the shared card, with editable HTML artwork under `design/`.
This activation also required `GOMEMLIMIT=320MiB` on the separate remote
`flea-server -deploy` process. The service override alone does not constrain that
process. The one-time environment setting preserved ordinary SSH checks and
made no additional persistent host changes. Consider this for future deployments
while the site serves several retained binary releases from memory.

## Historical state after the September 29 website deployment

Version 0.5.0 is published. Framework commit `0994c1f` and tag
`v0.5.0` are pushed. Both native downloads matched their public HTTPS checksums
before advancing the update pointer. Website docs now cover `--lan`/`-l`,
localhost links, device sign-in and in-place saves. See `RELEASES.md` for the
final deployment and installation checks. Active website release:
`20260930-010405-485949623ee87304b3304819` (commit `f2d3c60`). Public fresh installs
and updates from 0.4.1 passed on Mac ARM64 and unprivileged Linux AMD64, including
local and LAN dev startup. Temporary Linux build files were removed. The server
executable and Bookworm data were not changed. The entries below are historical.

Installation update: Sean approved the temporary private-source transfer and
public tester installation on September 29. Version 0.3.0 native Mac ARM64 and
Linux AMD64 binaries are now published under `/releases/0.3.0/`, with checksums,
`/releases/latest.txt` and `/install.sh`. Both actual public installer/update
flows passed in isolated directories; Linux ran unprivileged. Homepage and docs
now include the one-liner and label this an early test release. Clean-machine
Mac testing remains pending; the Mac binary is ad-hoc signed, not notarized.
See `RELEASES.md` for provenance and checks. This supersedes the older statements
below that downloads were not yet published.

Installation-page deployment: commit `2205161`, release
`20260929-202718-8ebaae1f4db520dfa3f53919`. Live homepage/docs command, installer
bytes, release pointer and links verified. Temporary Linux source/toolchain/test
directories were removed; the source snapshot remains privately under ignored
`release-work/0.3.0/`. Bookworm remains unchanged.

The website now lives in `/Users/sean/fleacms-website`, the separate repository
Sean supplied for this task. The earlier proposed directory and preview-only
instructions below describe the previous plan. Sean subsequently requested
building and deploying the site and confirmed `fleacms.com` as server and domain.

- Built `/` and `/docs` with bright white surfaces, green accents, the preserved
  Skip logo, and an actual local-editor screenshot at `site/images/editor-demo.png`.
- Site commit: `1ca1012` (initial website implementation and deployment settings).
- Live: `https://fleacms.com`, with HTTPS and a WWW-to-apex redirect.
- SSH: `root@fleacms.com`; remote site name: `fleacms-website-main`.
- Active release: `20260929-195606-b19e2095c1c32bb2802f4125`.
- Verified both public pages over HTTPS, the image checksum, HTTP/WWW redirects,
  local template validation, links and anchors. Bookworm's existing release,
  content and theme paths remain unchanged and its homepage returns HTTP 200.
- Desktop rendering was inspected. Browser automation failed during the later
  checks, so final desktop spacing and mobile rendering still need visual review.
- Public installer downloads are not published by this work. Docs say so.
- The production owner allowlist remains empty. No sign-in delivery or access
  settings were added. The screenshot used an isolated local demo with preview
  mail and a fictional address; no mail was sent.

Continue in this checkout. Website-only changes do not require a framework build.
Normal `flea deploy` preserves live snips and theme; pull before any explicit sync.

## Original design brief

## Start here

Build the public FleaCMS website as a separate Flea site and Git repository at
`/Users/sean/Projects/fleacms.com`. Sean agreed to separating it from the framework.
That directory does not exist yet. No files have been moved, no new repository
has been created, and the public website has not been deployed by this work.

The immediate task is a working **homepage preview for Sean's review**, at desktop
and mobile widths. After he accepts the design, carry it into the Docs page.
Do not publish or configure the apex domain yet.

## Agreed scope: two pages

- `/`: homepage with a direct introduction, a real editor demonstration, and
  short explanations paired with code examples.
- `/docs`: one long-form documentation page with anchored sections. Use a sticky
  contents list on desktop and a compact native disclosure on mobile.

Installation, quickstart, templates, owner editing, deployment, email, recovery,
updates, limitations, release notes and support all belong within `/docs`.
No separate Features, FAQ, Contact or Getting started pages. Authentication and
error templates are infrastructure, not extra public content pages.

## Design direction

Sean's reference is [zjax.dev](https://www.zjax.dev/) and its
[single-page docs](https://www.zjax.dev/docs). Both were inspected in Chrome.
Borrow the direct, example-led structure and compact navigation, not its dark
theme. Sean specifically asked for a **fresh, bright style**.

Use bright white surfaces, crisp charcoal type, vivid green accents, generous
whitespace and strong typography. Inter is a reasonable starting point; reserve
monospace for code. Preserve the selected Skip flea logo. Give the marketing site
its own composition, not a lightly modified starter site.

Sean rejected the first boilerplate polish because it felt unchanged. Do not
repeat a spacing-only pass or assume that maintaining the existing palette is a
virtue here. Make the visual direction tangible in the browser before asking him
to approve it. Keep functionality and page count simple.

Inspected Mobbin references:

- [Clerk](https://mobbin.com/screens/b57c3420-a006-4fe1-b903-4ca0ca658a5f): clear
  navigation, strong headline hierarchy and one prioritized action.
- [Greptile](https://mobbin.com/screens/ddd061c9-6056-4cd2-a671-4f433c84f2de):
  confident left-aligned typography and room for a distinctive product visual.

Do not copy their branding, illustrations or social proof. Mobbin MCP was
available and used. Sean requested a design skill earlier, but no matching skill
was available in this session. Check the new session's available skills; do not
claim to use a skill that is absent. Prewind and inspected references remain the
fallback.

Proposed copy, not yet approved:

> Write the HTML. Let them handle the edits.
>
> A small CMS for developers who build with plain files, and people who just
> want to update their website.

Header: logo/home, Docs, Get started (`/docs#installation`). No hamburger needed
for two destinations. Hero actions can be Get started and See how editing works.
Follow with an actual editor screenshot or recording, then a short `snip` example
and the new/dev/deploy workflow. Avoid feature-card filler, invented testimonials,
usage counts, unsupported claims and decorative mock interfaces presented as real.

## Sources and setup

Framework: `/Users/sean/Projects/flea`.
Existing unpublished scaffold: `/Users/sean/Projects/flea/website`.
Design notes: `website/design-direction.md`.
Roadmap: `next-steps.md`, section 5. Manual release evidence: `checklist.md`.

Read framework `AGENTS.md`, relevant `docs/agents/` guides, and
`/Users/sean/Projects/prewind/README.md` before implementation. Follow generated
site guidance in the new project too. The framework worktree contains extensive
uncommitted work. Do not reset, bulk-stage, commit or clean it.

Recommended setup:

1. Verify the destination is still absent and the installed CLI path/version.
   `bin/flea` in the framework was rebuilt successfully at version 0.2.0.
2. Generate a clean site beside the framework with `flea new fleacms.com`.
   This initializes Git without a commit. Do not use `--starter`: the marketing
   site does not need Team/Contact/demo pages. Do not create a remote or push.
3. Selectively port useful product copy and the preserved logo from the scaffold.
   Keep the old scaffold intact until the new project is verified; do not blindly
   copy its entire tree. It contains old sibling directories and stale guidance.
4. Build `/` and the `/docs` shell using Flea and Prewind. Keep the framework's
   canonical starter separate from marketing-specific design work.
5. Run local validation and inspect actual desktop/mobile rendering. Present the
   homepage preview for review before writing out the full docs design.

If permissions prevent creating the sibling project, request scoped access rather
than silently building the finished website inside the framework again.

## Flea conventions and current behavior

- Plain HTML/CSS/JSON, no separate frontend framework, npm pipeline or database.
- Public templates/assets in `site/`, private emails in `mailers/`, content/media
  in `flea/`, config in `.flea.config.json` with public globals under `app`.
- Shared wrapper is `site/main.layout.html`; use `{slot}`. Preserve Skip at
  `site/images/mark.svg`. Keep header/footer inline unless there is a real reason
  to introduce a partial.
- Prewind 1.3.1, global-first styling, layers `reset, styles, prewind`. Theme
  tokens for colors/fonts/spacing; supported utilities for one-off adjustments.
  No invented Tailwind utilities. Respect reduced motion and keyboard focus.
- `flea new foobar` is clean by default. `--starter` / `-s` adds example pages.
  `flea reset` confirms before resetting local files, preserving Git and a private
  sibling recovery copy. It never resets a live site.
- Local email preview is enabled for new sites. Dev prompts for site name and
  optional owner email, not Resend. Production delivery requires configuration.
- Owners sign in by email code, then edit on-page. Edit defaults on and persists
  per tab. Page Save/Cancel are in the formatting bar; Theme has its own Save/X.
- Text, photos, supported MP4 video, collection editing/reordering and theme
  controls exist. Do not repeat scaffold claims that managed media is pending or
  collections use a side panel with duplicate buttons.
- `flea update` updates the CLI; `flea update-server` updates the server.
  Rollback, backup/restore and pull are implemented. Verify exact syntax using
  current CLI help and framework docs before writing instructions.
- Normal deploy preserves live snips and theme. Explicit `--with-theme` or
  `--with-snips` sync requires the relevant pull baseline for existing sites.
  Never discard owner content to make screenshots or deploys easier.

The scaffold's README and pages are not authoritative. They contain stale
Homebrew installation, Resend onboarding, collection UI and media/recovery claims.
Use current framework README, implementation and tests to reconcile copy.

## Distribution and hosting boundaries

Source remains private. Binaries are intended to be free without an account or
license key for now. Do not describe Flea as open source or add a public source link.
The planned installation is one copy-paste shell command at
`https://fleacms.com/install.sh`, with release metadata under `/releases/`.
Public downloads and installation remain unverified. Do not show the installer
as working, advertise untested platforms, or tell visitors to execute it yet.
Primary target platforms are Apple Silicon macOS and x86-64 Linux; additional
architectures require release verification before public support claims.

Final apex hosting destination/DNS and support route still need confirmation.
Existing test infrastructure is not blanket permission to replace fleacms.com.

## Existing test site: reference only

- Live: `https://bookworm.fleacms.com`.
- Local source: `/tmp/flea-live-qa.YOhVfCy1/v1-acceptance`.
- Saved SSH destination: `root@fleacms.com`, branch `main`.
- Latest test-site commit: `4715e8f`, swapping homepage buttons to put Meet the
  team before Get in touch. Last release:
  `20260929-185956-b63c97e8e70f6549ae70d3dd`.
- It has the recent Inter/forest-green theme, white panels and warm backgrounds.
  Its live owner content must be preserved. Do not turn it into the marketing site.

These are last-known values, not a substitute for checking live state. No local
preview server was intentionally left running by the completed design work.

## Verification and communication

Use `flea status` and `flea dev` in the new site; inspect real browser output at
desktop and mobile widths. Check routes, anchors, code wrapping, contrast,
keyboard navigation and absence of horizontal overflow. Do not submit live email
forms or request real sign-in codes just to obtain screenshots.

If framework/starter code changes are needed, work in the framework and follow
its required tests, vet and bundled build workflow:

```sh
GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go test ./...
GOCACHE=/tmp/flea-go-cache /usr/local/go/bin/go vet ./...
GOCACHE=/tmp/flea-go-cache GO_BINARY=/usr/local/go/bin/go bash scripts/build.sh
```

Plain `go build` can embed stale Linux servers. Website-only template changes do
not require rebuilding the framework. Do not deploy, publish releases, push or
create a public repository until Sean asks.

Keep responses succinct and natural. No emojis, em dashes or AI marketing cliches.
Ask one question at a time when a decision is genuinely needed. The next step is
to build the homepage preview, not to ask Sean to approve the two-page scope again.
