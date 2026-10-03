# FleaCMS agent handoff

## Latest release: 0.8.0, October 3, 2026

Framework `c8a1d96`, immutable tag `v0.8.0`; website `3188000`, activation
`20261003-171440-ca76afb853e32d0d0309c6ea`. Server, public channel and Sean's
installed CLI are 0.8.0; the installed Mac hash matches the public manifest.
Edit row collapses outside Edit; Theme slides from the right and tracks toolbar
height. Existing sites must update the starter toolbar partial with the runtime.
Naves `c59dfff`, activation `20261003-171113-e1b9be7c7e17a28c8aca4ffd`;
Kent `d80eaeb`, activation `20261003-171202-d490ea0173f4c49cbd1be27a`.
Both deployments preserved content/theme. Bookworm was explicitly excluded; its
public page returns 200, but its old editor toolbar is incompatible with 0.8.0.
Do not mistake a public health check for editor acceptance there.

Tests/vet, bundle build, final browser regressions, native Mac signature/new/status
and unprivileged Linux smoke checks passed. Public fresh installs and 0.7.1 updates
passed on both platforms with exact hashes. Linux was cross-compiled; native full
suite/build and clean Mac/notarization remain open. Website RELEASES.md has hashes.
Host memory/swap settings are unchanged; no recovery journals remain.
Private pre-release site backups are /tmp/flea-{naves,kent,website}-before-080.flea-backup.
Kent's unrelated skill edit and framework APPS.md were preserved. Kent deployed
from /tmp/flea-kent-release-080. Naves has no Git remote; its commit is local only.
No OTP mailer changes were made. Earlier release notes below are historical.

## Latest patch: 0.7.1, October 3, 2026

Framework `f4a72ee`, tag `v0.7.1`; website `f75d63f`, activation
`20261003-151507-4d9b7bc05ab6b470b4ae679b`. Server, public channel and local CLI
are 0.7.1. The server now chooses deployment progress from installed TLS domains:
first-time enablement keeps setup guidance; repeat deployments say "Checking site
health and HTTPS...". The latter was observed during this real deployment.
Readiness/certificate verification is unchanged. Full tests/vet, bundled build,
Mac signing/fresh-site and native unprivileged Linux smoke checks passed. Public
fresh installs and 0.7.0 updates passed on both platforms with matching hashes.
Both sites returned 200 and recovery is clear. Distribution acceptance limits
below remain; APPS.md is still unrelated, untracked and untouched.

## Current handoff: October 3, 2026

0.7.0 is published and verified. Framework source `1199747`, immutable tag
`v0.7.0`; website source `a41a758`, activation
`20261003-141847-a863aaed91fd5f4c8fa761b8`. Both repositories were pushed.
Server, public channel and Sean's local CLI report 0.7.0. Tests, vet, race checks,
Mac signature/generated-skill checks and native unprivileged Linux smoke checks
passed. Public fresh installs and updates from 0.6.2 passed on Mac/Linux, including
already-current behavior. Full native Linux suite/build and clean Mac/notarization
remain unverified. The website release record has exact hashes and migration notes.

The bundled skills now include interactive design checkpoints. The review corrected
overbroad watcher/caching and editing claims while preserving the requested process.
`APPS.md` is unrelated untracked planning work, deliberately left untouched.
Older notes below about uncommitted runbook/QR changes are superseded: those were
already included in 0.6.2. QR remains planned, not implemented.

Downloads are now separate from website snapshots. Read
[download hosting](docs/download-hosting.md) and the updated
[deployment runbook](docs/deployment-runbook.md). Production store:
`/var/lib/flea/downloads/fleacms`; local artifact archive:
`/Users/sean/fleacms-releases/`. All seven published versions and hashes were
verified over HTTPS before removing duplicate binaries from the website tree.
Git history and old server releases remain intact. Only latest.txt stays in the
website; do not add versioned binaries back. Public URLs are unchanged.
The small website now has about 0.63 MiB of public assets; deployment uploaded
575 KB. Keep the existing 320 MiB targets and 1 GB swap buffer.

The new host config mapping is explicit administrator policy. Its pre-change
backup is `/etc/flea/server.before-downloads-20261003T141633.json`. Normal site
deployments preserve it. Pre-0.7 executables cannot load the field; see rollback
instructions before attempting an executable downgrade. A migration helper's
config permission mistake was corrected to Flea's standard 0644 mode; final
service, download, range, isolation and public-site checks passed. No recovery
operation or preview server was intentionally left running.

Read current AGENTS.md, inspect both worktrees and live state, and follow Sean's
new request. Do not treat this handoff as standing deployment authorization.

## Historical handoff through 0.6.1

Current handoff recorded October 1, 2026, after release 0.6.1.
Start here. The older website-design brief below is preserved as history, not
an instruction to rebuild, relocate or publish the website again.

## Instructions for the next agent

1. Read [AGENTS.md](AGENTS.md), the complete
   [deployment runbook](docs/deployment-runbook.md), including its handoff notes,
   and [next-steps.md](next-steps.md). Consult [checklist.md](checklist.md) for
   manual acceptance evidence and preserve Sean's answers.
2. Inspect Git status before changing anything. Preserve modified and untracked
   work. Do not reset, bulk-stage or commit without current permission.
3. Ask Sean what to work on next unless his new request already specifies it.
   Pending ideas below are not authorization to implement them.
4. Do not deploy, publish, push tags or change host settings without current
   approval. The completed release's authorization is not standing permission.
5. Use one agent. Keep updates brief and natural, without emojis or em dashes.

## Where things live

- Framework: `/Users/sean/Projects/flea`, branch `main`.
- Actual public website: `/Users/sean/fleacms-website`, branch `main`. Read its
  AGENTS.md before editing. Its RELEASES.md and handoff.md contain release evidence.
- Production: `root@fleacms.com`, shared service for fleacms.com and
  bookworm.fleacms.com. Preserve both sites' live editable content and theme.
- The framework's `website/` directory is not the deployed website repository.
  The historical proposed `/Users/sean/Projects/fleacms.com` path below is obsolete.

## Last verified state

FleaCMS 0.6.1 was deployed and verified: server executable, public website,
versioned Mac ARM64/Linux AMD64 binaries, stable update channel and Sean's local
CLI agreed. Public fresh installs and updates from 0.6.0 passed on both platforms,
with Linux checks run unprivileged and installed hashes matching the manifest.
Both hosted sites and the social-preview image returned HTTP 200. Recovery was
complete. This records the last checks, not a guarantee of current live state.

- Release source: framework `89e3472`, immutable tag `v0.6.1`.
- Deployed website source: `1857a45`.
- Website activation: `20261002-015019-3e6152722de7e84b956d7054`.
- Pushed release-record commits: framework `f33db2e`, website `90ab9ac`.
- Full native Linux build/test coverage, independent clean-machine Mac acceptance
  and notarization remain open. Do not describe this test release as completed V1 QA.

Version 0.6.1 adds graceful replacement of a same-site dev server, including LAN
and port changes. Pull/reset remain exclusive. An already-running older binary
may need one manual stop; never bypass the lock or kill unrelated work.
The website also has social-preview metadata and a public preview image.

## Low-memory host: read before deploying

The approximately 1 GB production host ran out of memory during upgrade and
activation, not compilation. The approved safeguards are:

- `GOMEMLIMIT=320MiB` in the service's `60-memory.conf` systemd override.
- The same target applied separately to remote deploy/upgrade/recovery processes
  through the scoped SSH wrapper documented in the runbook. Local environment
  settings and the service override alone do not cover those processes.
- A 1 GB `/swapfile`, mode 0600, active and persisted in `/etc/fstab`.

Recheck RAM, swap, pending journals and running admin processes before starting.
Never overlap operations, delete journals to bypass recovery, prune old releases,
or assume a soft memory target is a hard ceiling. `/tmp` is RAM-backed on the host.
Use the runbook's full build, SemVer, staging, recovery and verification procedure.

## Pending work and local changes

At this handoff, the framework has uncommitted documentation changes:

- This HANDOFF.md update.
- New `docs/deployment-runbook.md`, with the full release procedure and additional
  handoff notes, plus links in RELEASING.md and docs/agents/deployment.md.
- A separate next-steps.md planning item for a terminal LAN QR code. No QR feature
  has been implemented; automatic versus opt-in behavior remains undecided.

Preserve all of these. Check actual Git status because this list can become stale.
Documentation updates after the release have not been committed or deployed.

Other potential follow-ups are reducing asset-loading memory, disk-backed upload
staging, replacing the operator SSH wrapper with a supported configuration, and
finishing release acceptance checks. None is an instruction to start that work.

Sean's proposed Linux ARM build machine, `buzz` (last found at 192.168.4.31), is
not an established runner. SSH failed for Sean and the agent. The reported 24 GB
RAM and ARM architecture were not verified through login. No source was copied
or software installed. Sean chose to leave it aside; do not resume SSH repairs
without a request. The runbook records the diagnostic details.

Before handing off again, record what changed, exact checks, unresolved limits,
running processes, Git state and any operations still pending. Do not assume an
interrupted agent turn stopped its local or remote processes.

## Historical website-design brief (superseded where noted above)

The remainder is the September 29 brief, retained for design context. Statements
that the website does not exist or has not been published are no longer current.

### Original brief: fleacms.com website handoff

Updated September 29, 2026.

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
  optional admin email, not Resend. Production delivery requires configuration.
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
