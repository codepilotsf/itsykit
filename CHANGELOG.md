# Changelog

User-facing changes are recorded here, newest first. Versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).
During 0.x development, interfaces may change without backward compatibility.

## Unreleased

## 0.2.0 - 2026-09-24

Local development release. No release artifacts have been published.

### Added

- Isolated Squire inline-editing prototype with a two-tier admin bar, page
  metadata, shared content, explicit saving, editable links and new-tab settings.
  This is a demo, not yet the production editor. No npm or frontend build step.
- Reusable CSS-only toggle switches in the starter's `css/ui.css`.

### Changed

- Renamed editable content to flea throughout: bare `flea` HTML attributes,
  `{flea ID}` references, `_/fleas.json`, and `fleas` JSON fields. No old syntax
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
  editing for pages without fleas, preserving the shared site layout.

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
- Editable rich-text fleas with sanitized HTML and content preserved during deployment.
- Local development server with automatic reload and site validation.
- Minimal Prewind starter, Skip flea logo, and on-demand Lucide icons.
- SSH server setup and branch-specific deployment with bundled Linux servers.
- Optional automatic HTTPS and WWW redirects.
- Owner login by emailed code, page metadata editing, and flea editing.
- Email-only contact forms using Resend.
- Grouped CLI help and per-command help.

Live email delivery, remote installation, and public certificate issuance still
need verification on a real server.
