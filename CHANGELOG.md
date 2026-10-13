# Changelog

User-facing changes are recorded here, newest first. Versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).
During 0.x development, interfaces may change without backward compatibility.

## 0.15.1 - 2026-10-13

- Bundle focused `add-blog` and `add-contact-form` agent skills in new sites,
  with request routing, setup steps, template values and content/form conventions.

## 0.15.0 - 2026-10-13

- Add optional Blog with a private admin list/editor, explicit Publish/Unpublish,
  rich text and library/uploaded covers. Public templates receive published
  posts only. Content preservation, explicit synchronization and backups reuse
  the existing content snapshot.
- Deployment protocol 25 adds the blog content and mounted routing contract.

## 0.14.0 - 2026-10-13

- Remove the remote Theme editor, CSS overlays and theme synchronization. Keep
  `site/css/theme.css`, Prewind, theme tokens and custom CSS as developer-managed
  source. Normal deploy publishes committed CSS; rollback restores release CSS
  while keeping live snips, and backup/restore includes source styles.
- Remove `--with-theme`, `--theme-only` and `--snips-only` without aliases. Pull
  retrieves live snips/metadata and referenced media; snip editing and explicit
  `--with-snips` publication retain their validation and conflict checks.
- Bump the version to 0.14.0 and deployment protocol to 24. Removed
  host theme mappings are rejected. Existing remote theme customization needs
  explicit preservation and reviewed maintenance before a live upgrade; no
  migration or automatic file deletion is included.

## 0.13.0 - 2026-10-12

- `flea status` now shows a readable local summary first, then automatically
  checks the current branch's saved deployment with terminal progress. Code,
  editable content and theme are compared separately; scoped pull baselines
  distinguish local, remote and conflicting edits. Status stays read-only.

## 0.12.0 - 2026-10-12

- Add Linux ARM64 CLI release builds and public staging alongside Linux AMD64
  and Apple Silicon macOS. Missing installer downloads now identify the selected
  release and platform while preserving checksum verification.

## 0.11.0 - 2026-10-12

- `flea new` now creates an `Initial site` commit on main and prints commands
  to connect and push to a remote repository. Git identity, hooks and signing
  settings are respected; failed commits preserve the site with retry instructions.
  `--skip-git` still skips Git setup. No remote or hosted project is created.

## 0.10.0 - 2026-10-05

- Rename `flea server sites` to `flea server list` without an alias. `flea server
  list` and `flea server remove` now prompt for the server login (for example
  `root@fleacms.com`) when no target is given, and `flea server remove` without a
  site name lists the installed managed sites and asks for one by exact name
  before the existing exact-name confirmation. Normal listing shows each site's
  name and branch, then HTTPS domains, then HTTP-only domains, then redirects,
  without the release or repeated plain domain lines. Removal review is unchanged.

## 0.9.0 - 2026-10-05

- Add `flea server sites <ssh-target>` and `flea server remove <ssh-target>
  <site-name>` for server-wide site administration independent of a local site.
  Listing is read-only; removal is unregister/offline only and retains every
  release, content file, media file, session and certificate. Removal requires
  typing the exact site name, rejects state that changed since review including a
  commit-only change, blocks on pending recovery and reuses the deployment lock,
  journal, probe and rollback paths. It restarts the shared service, so other
  sites are briefly unavailable, and disables automatic HTTPS when no HTTPS
  hostnames remain while cached certificates stay on disk. Retained content is
  not restored automatically by redeploying the same name.
- Use mount-relative `thank-you` for Contact success and remove the starter's
  duplicate delivery confirmation.
- Add the bundled Contact mini app with explicit `{csrf}`, private managed
  process startup, authenticated JSON forwarding and a single-use email capability.
  Validation preserves safe fields; success redirects with visitor-bound flash.
  App email.from can override the server sender, including in dev previews.
- Replace top-level contact settings and /_flea/contact with apps.contact and a
  configured mount. Starter pages now use /contact and /contact/thank-you.
  No aliases or automatic migration; deployment protocol 23 requires the new
  server. Admin/editor remain core. Custom apps, remote registration and storage
  are still planned. Live mail, Linux/systemd and capacity acceptance remain open.

- Stream large static assets and managed photos from validated file handles
  instead of keeping their full contents on the Go heap. Keep small assets and
  theme CSS in memory. Stream fingerprinting and photo validation as well.
- Reject hard-linked streamed assets. Detected in-place changes fail closed;
  atomically replaced files retain their original bytes in existing snapshots.

## 0.8.1 - 2026-10-03

- Append only missing runtime and media rules to `.gitignore`, regardless of
  where existing rules appear. Preserve existing content and trailing blank lines.

## 0.8.0 - 2026-10-03

- Collapse the second toolbar row outside Edit with a slide transition. Theme
  slides from the right and uses the space below the changing toolbar height.
  Reduced-motion preferences keep both changes immediate.
- Existing sites must update `site/admin/_toolbar.html` from the starter alongside
  the executable. The new edit-row wrapper is required; custom layouts stay unchanged.

## 0.7.1 - 2026-10-03

- Show certificate setup guidance only when enabling HTTPS for a new domain.
  Routine deployments say "Checking site health and HTTPS..." while retaining
  all readiness, certificate and rollback checks. HTTP-only sites use a health check message.

## 0.7.0 - 2026-10-03

- Expand the bundled import skill to brochure, portfolio and artist-site redesigns,
  with two design directions and an approved sample page before the full build.
- Add matching review checkpoints to the site-building skill and update generated
  README/agent guidance. Explicitly delegated design decisions remain supported.
- Clarify large-gallery asset handling, media constraints, safe editing, dev reloads
  and visual verification in both skills.
- Serve administrator-managed, checksum-verified release downloads from disk,
  outside website snapshots. Existing public URLs remain unchanged; site archive
  limits and the deployment protocol are unchanged.

## 0.6.2 - 2026-10-01

- New sites bundle a second agent skill, `fleacms-site`: build or redesign a
  site for one real local business. It covers deciding the site's job,
  research-first vibe discovery, real photos and SVG logos taken from the
  actual signage, content truth rules, genre-driven design, and a safe
  editing plus dev/verify workflow.
- List both bundled skills (`import-brochure` and `fleacms-site`) in the
  generated AGENTS.md and README.

## 0.6.1 - 2026-10-01

- Starting `flea dev` again gracefully replaces the same site's existing dev
  server, including LAN/port changes. Pull/reset remain protected, and crashed
  dev instances do not leave a blocking control socket.
- Preserve the underlying error when an OS lock fails for a reason other than
  contention, rather than incorrectly reporting another running command.

## 0.6.0 - 2026-10-01

- Clarify the message shown when `flea dev`, `pull` or `reset` finds the site
  already in use, without repeating the generic lock text.

- On the first `flea dev`, explain that the optional email adds an admin who can
  sign in and edit, point to `flea admin` when skipped, and ask only once. Space
  out the prompts and startup links for readability.

- Use "admin" for the people who sign in throughout user-facing copy, help and
  documentation; internal code and DOM identifiers keep "owner".

## 0.5.0 - 2026-09-29

- Add `flea dev --lan`/`-l` with detected private IPv4 URLs and full authenticated
  development, forms, reloads, diagnostics and private email-preview links on a
  trusted network. Keep ordinary dev loopback-only and production HTTPS-only.

- Display dev website and sign-in links using localhost, retaining loopback-only access.

- Save page edits in place without reloading fonts or the page. Keep editing on
  and make Cancel restore the latest successful save, including collections/media.

## 0.4.1 - 2026-09-29

- Keep `flea new` completion output focused on starting local development.

## 0.4.0 - 2026-09-29

- Record the current framework development snapshot as 0.4.0. Public downloads
  remain on the separately published 0.3.0 test release.

## 0.3.0 - 2026-09-29

- Prepare the first public binary distribution for Apple Silicon macOS and
  x86-64 Linux, with checksum-verified installation and local CLI updates.
- Correct the starter contact-page test to match its six editable regions.

- Give the starter an Inter/forest-green theme with warm off-white pages, white
  access panels and team cards, softer controls and a narrower content measure.

- Refine boilerplate spacing, heading proportions, team cards and form hierarchy.
  Sign-in and code pages share a narrow layout with clearer primary actions;
  the standalone server-error page stays styled without loading site assets.

- Enlarge the borderless mobile hamburger; animate the menu and backdrop below
  the measured owner toolbar, respecting reduced motion.

- Make `flea new` minimal by default; `--starter` / `-s` adds example pages and
  six fictional team members. Share navigation with a CSS-styled native mobile popover.
- Add confirmed local `flea reset`, preserving Git history and a private sibling
  recovery copy while replacing site files, uploads, credentials and configuration.

- Improve the starter with Team/Contact navigation, a responsive contact page and
  more editable copy. Put installation and `flea new foobar` first in the README.

- Keep Page Details visually enabled across navigation when Edit is on, while
  retaining the inert toolbar until the editor is ready.

- Restore the tab's Edit preference before rendering the toolbar, preventing an
  on/off flash during page navigation while editor context is loading.

- Shorten page actions to Save. Theme uses its own fixed-footer Save and header X,
  with Escape, discard confirmation and reduced-motion-aware closing animation.

- Keep white collection drag/remove controls above editable outlines so dashed
  borders cannot cross their faces.

- Simplify the owner bar: green-on Edit switch, compact Theme/Sign out pills,
  no wordmark/account divider or idle hints. Load native navigation transitions
  directly from layout heads instead of through the general CSS imports.
- Default to persistent per-tab Editing, with a first-row switch and Theme button,
  second-row Save/Cancel, and 2px green dashed editable boundaries with rounded
  corners and a subtle reduced-motion-aware fade. Cancel keeps editing on.
- Remove the Collections toggle. Use inset grip/remove controls, a floating drag
  preview, visible landing space and reduced-motion-aware reorder animations.

- Replace the collection-photo dropdown with a thumbnail picker, clear selection,
  Upload photo and a larger preview. Show alt text alongside the selected photo;
  keep new uploads private until page Save changes.

- Add a user-directory shell installer and `flea update` with verified downloads
  and failed-update preservation. Public release hosting is still pending.
- Rename remote updates to `flea update-server`, including `--recover`, with no
  old-command alias. Local updates never update remote hosts.
- Simplify dev diagnostics, use normal 404 pages, locate unclosed conditionals
  at their opening directive, and start Theme accordion sections closed.

- Preview actual development emails in the browser, with sender, recipient,
  subject, timestamp and HTML/plain-text bodies. New sites enable
  `dev.email_preview`; first-run dev no longer asks for Resend credentials.
  Production delivery is unchanged; previews stay temporary and in memory.

- Store icons directly in `flea/icons/`, with `.keep` and license notices inside
  each SVG. Render formatting icons with the toolbar instead of inserting them
  after navigation.

- Add `{#unless dotted.path}` negative conditionals, with else and else-if branches.
- Render the admin toolbar shell with the page so it remains visible while
  editor data loads during navigation. Existing sites need the updated toolbar partial.

- Collection photo fields accept new JPEG/PNG uploads as well as library choices.
  Preview privately, then publish the photo and item together with Save changes.

- Replace the collection sidebar with on-page Collections controls: drag handles,
  remove icons, Add tiles and keyboard reordering. Keep page Save/Cancel and direct
  photo selection. Show neutral portraits for empty photos, share select styling,
  and compile collection Lucide icons through the existing icon directive.

- Add typed image/video snips from bare snip attributes, with source/alt/poster
  values stored once in JSON. Click designated media in Edit mode and save or
  cancel it with the page. Remove Photos/Videos topbar library modes. Reject
  nested media/text snips and competing sources; preserve developer attributes.
  Sync compares complete typed values. Protocol 22 requires the typed contract.

- Keep the signed-in formatting row visible with disabled controls outside Edit,
  avoiding its collapse after saving or cancelling. Submit a complete six-digit
  pasted sign-in code automatically, retaining manual submission for typing.

- Reorganize projects around site/, mailers/, flea/media/, flea/runtime/ and
  flea/skills/. Use *.layout.html and *.email.html/txt templates. Consolidate
  branch destinations in root .flea.config.json under deploy.branches; exclude
  those destinations from release archives and backups. Protocol 21 requires
  the new structure; no compatibility loader or automatic migration.

- Make toolbar integration explicit through app.current_admin, a CSRF meta tag
  and admin/_toolbar.html. Serve editor assets separately; move crossfade styles
  into editable starter CSS. Existing custom layouts require explicit opt-in.

- Add local optimized MP4 uploads, validation, posters, replacement/download,
  disk-backed native playback and video preservation through deploy/pull/backup.
  Video files stay outside Git. No server-side conversion or hosted provider.
  Protocol 20 requires the video runtime; release/backup limits are 256/270 MiB.

- Include portable brochure-import agent guidance in new sites, excluded from
  runtime serving and deployment.

- Add branch-scoped rollback preserving live content, verified site backup/restore,
  and nondestructive executable upgrades with automatic failure recovery.
- Add owner photo uploads/replacement/alternative text, bounded JPEG/PNG processing,
  multisnip photo selection, and photo preservation across deploy/pull/backup.

- Add page-owned multisnips with named fields, source extraction, side-panel
  item management, conflict-checked saves, and content-preserving deployment.
- Add editable `mailers/` templates, local email preview, and starter contact,
  thank-you and team pages. Protocol 20 requires the new runtime.
- Add an unpublished Flea website scaffold and local Homebrew/Mac package tooling.

- Move public site-wide variables to `app.*`, including `app.site_name`.
  First interactive dev asks for the name with a title-cased folder default;
  saved names are reused. Deployment requires the saved name and protocol 20.

- Simplify dev onboarding to optional admin/key prompts and sender details,
  with concise file confirmations and asterisk-masked API key entry.

- Default starter typography to Roboto Mono for headings and body text.
- Group text/background token selectors under Page colors, default background
  to White, and enable inline formatting in heading/paragraph snips.

- Use consistent topbar Cancel/Save actions for both Edit and Theme modes,
  keeping Sign out visible and removing the Theme footer action bar.

- Reload clean signed-in dev pages after source/snips changes and restore edit
  mode, while protecting unsaved drafts, dialogs and Theme previews.

- Make Font Color a neutral/brand theme-token dropdown, defaulting to Black in
  the starter and excluding semantic colors.

- Automatically load Google Fonts during theme preview and save managed imports
  to theme.css, preserving developer imports and restoring fonts on Cancel.

- Simplify sign-in headers and headlines. Refine Theme labels, font-family text
  inputs, help bubbles, color groups, focus rings and exclusive accordions.

- Offer missing owner/email setup before interactive dev startup, with skip
  options and concise website/sign-in links. Automated runs never prompt.

- Theme text and spacing scales now have coordinated endpoint controls, immediate
  private previews, and server-side CSS generation without Node. Correct negative
  intercepts and require confirmation before replacing custom scales.

- Simplify Theme panel labels, inset select chevrons, and animate accordion
  expansion/collapse with reduced-motion support.

- Add explicit snip/theme pull and deployment options, three-way conflict review,
  destination-scoped baselines and activation-time concurrency checks.
- Add owner Theme editing with a right-side panel, private preview and Save/Cancel.
  Preserve live theme CSS across normal deployments. Deployment protocol is 14.

- Use `snip` attributes, `{snip ID}`, and `flea/snips.json` consistently for editable
  content. Reject the old content format without migration or fallback.
- Remove editing outlines from content, preserving toolbar and form focus cues.

- Publish inline content and page details together with conflict-checked atomic
  saves. Restore headings, marks, links, lists, alignment, and metadata dialogs;
  retain drafts on failed saves. Remove the preview warning and old dashboard.
  No undo/redo toolbar buttons. Support H1, restricted alignment, and safe new-tab
  links in the server sanitizer.

- Rename site configuration to `.flea.config.json`, with no old-name fallback.
  Existing sites must rename the file. Deployment now requires protocol 13.

- Expand dev error excerpts to 11 numbered lines and preserve source locations
  for invalid editable regions and page settings.

- Consume the `snip` authoring attribute during extraction. Extracted HTML needs
  only `{snip ID}`; inline editing recognizes directive boundaries instead.

- Enable native cross-page fades on real sites, including signed-in pages, with
  reduced-motion support and no layout edits.
- Add styled local-dev diagnostics with template context, safe JSON locations,
  route guidance and automatic recovery. Log rendering details in both dev and
  production while keeping production error responses private.

- Add an authenticated inline preview on real pages, preserving existing snip
  wrappers and visitor rendering. Login returns to the homepage. Inline Save
  remains disabled; the existing dashboard still publishes content and metadata.

## 0.2.0 - 2026-09-24

Local development release. No release artifacts have been published.

### Added

- Isolated Squire inline-editing prototype with a two-tier admin bar, page
  metadata, shared content, explicit saving, editable links and new-tab settings.
  This is a demo, not yet the production editor. No npm or frontend build step.
- Reusable CSS-only toggle switches in the starter's `css/ui.css`.

### Changed

- Renamed editable content to snip throughout: bare `snip` HTML attributes,
  `{snip ID}` references, `flea/snips.json`, and `snips` JSON fields. No old syntax
  aliases or automatic migration. Deployment protocol is now 10.

- Admin sessions now survive browser closure, restart and deployment, with a
  configurable 30-day rolling idle window and private per-site JSON storage.
  Active editing renews sessions; idle tabs do not.

- Added leading flea page-settings comments for custom layouts.
- Admin now uses a full-width, footer-free layout with pinned page navigation
  and separate editors per page. Textareas grow with content up to 400px.

- Replaced dev's periodic file scans with native directory notifications and
  debounced reloads, including nested directories, renames, and atomic saves.

- Standardized product branding to FleaCMS throughout the project.
- Added scrollable admin page navigation, a Homepage label, and explicit metadata
  editing for pages without snips, preserving the shared site layout.

- Email setup now ends with a single saved-file confirmation, without routine
  connection and untested-delivery notes.

- Shortened `flea admin` output and start email setup when local settings are
  missing. Dev owner access and sign-in email now update without restarting.

- New sites include editable admin/*.html pages sharing the site's layout,
  header, footer, and styles, while authentication stays server-controlled.
- Fixed browser sign-in forms being rejected because the admin referrer policy
  caused browsers to send a null origin.

- Added `flea email` to check Resend connectivity and save an API key, sender name,
  and from address to a private, Git-ignored `.env` file.
- Dev loads local email settings automatically; deployment uses them when server
  email settings are missing. Setup pages now point to the email wizard.

- Clearer owner setup instructions in `flea admin`, including sharing the login
  link and applying access changes.
- Local `/admin` setup guidance for fresh sites and styled missing-email pages.
- First-time sign-in, code delivery help, and per-form saving guidance for owners.

## 0.1.0 - 2026-09-22

First development version. This entry records the local implementation; it does
not indicate that release artifacts have been published.

### Added

- Plain HTML sites with layouts, partials, JSON data, conditions, loops, and routing.
- Editable rich-text snips with sanitized HTML and content preserved during deployment.
- Local development server with automatic reload and site validation.
- Minimal Prewind starter, Skip flea logo, and on-demand Lucide icons.
- SSH server setup and branch-specific deployment with bundled Linux servers.
- Optional automatic HTTPS and WWW redirects.
- Owner login by emailed code, page metadata editing, and snip editing.
- Email-only contact forms using Resend.
- Grouped CLI help and per-command help.

Live email delivery, remote installation, and public certificate issuance still
need verification on a real server.
