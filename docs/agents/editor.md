# Owner editing, authentication and email setup

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

Successful login returns to `/`. Public HTML pages receive owner-only snip marker
comments and the owner bar only after the editor authenticates the session.
`site.Server.ServeOwner` is a trusted in-process rendering entry point, never a URL
or query switch. Do not mutate published templates or content to add editor state.
`internal/editor/inline/` embeds pinned Squire 2.4.9 and DOMPurify 3.4.16 plus the
editing controller; no npm or frontend build step. Keep vendor files unmodified.
The bar uses a shadow root to isolate controls from arbitrary site CSS and stays
in document flow. Existing snip wrappers remain intact. Only an entire marked
wrapper is editable, never navigation, buttons, forms or mixed template content.
Detect wrappers from server-injected start/end comments around a rich directive;
the source snip attribute is consumed during extraction and is not required at runtime.
Prose containers get full Squire formatting; heading/paragraph leaf wrappers get
inline-only formatting through Squire in a temporary offscreen prose buffer.
Never put Squire's block wrappers inside the actual heading or paragraph.
Restore selection by text offsets and sanitize paste to inline content;
allow inline marks/links, never nested block markup. Disable heading/list/alignment
controls for constrained wrappers and validate the same restriction server-side.
The toolbar supports H1–H4, basic marks, current-link editing/removal, bullet and
numbered lists, left/center/right alignment, and a Page details modal. No undo/redo
buttons. Formatting is disabled without an active rich region. Editing is a native
switch labelled Edit, defaults on and persists in sessionStorage per owner/origin/tab. Clear
its preference on Sign out. Switching off with dirty content requires confirmation;
Restore the preference with the canonical hash-allowed inline initializer before
paint, using the escaped owner email on the mount. Do not wait for editor context
or permit arbitrary inline scripts; the full editor remains external and deferred.
Initialize Page Details from that same preference before paint. The toolbar stays
inert until hydration even when Page Details looks enabled.
navigation retains native beforeunload protection. Cancel restores the page and
keeps editing on; Save applies the canonical server receipt in place. Cancel/Save sit at the
right of the second row, disabled until changes exist. Theme stays in the first
row, suspends content editing and restores the preferred mode on close.
Keep the first row compact: no product wordmark or account separator, dark-off/
green-on switch, and 30px pill-shaped Theme/Sign out buttons. Keep the account
name close to Sign out. Omit idle instructional hints; retain unsaved/error
status and the shared-content warning.
Editable overlays use 2px dashed green (#16a34a) boundaries with 6px corners,
including collection cards and Add tiles. Keep these outside saved content.
Keep opaque white collection controls stacked above editable field outlines.
Fade overlays in/out over 160ms; reduced motion is immediate. Exit outlines are
inert, border-only clones with no controls, cleared before re-entering editing.
Dialogs retain the prototype's compact titles, left text Cancel, right primary action,
fade/slide motion and reduced-motion behavior. Reuse canonical ui.css switch styling.
POST /admin/page saves one bounded PageEdit under the existing content/deployment
lock. Require session, exact Origin and CSRF. Bind the version to the actual rendered
content/templates, not a later Backend.Read. Reject stale versions, duplicate IDs,
hidden/cross-page IDs and block markup in leaf wrappers. Shared IDs update all uses;
repeated on-page instances stay in sync. Validate the whole candidate, persist once,
then publish once. Failed requests preserve the draft and a visible error message.
Return the committed version and canonical changed values from inside that same
save lock, never from a later Read. Rebase text, collection membership/order,
photos and media baselines after success; subsequent Cancel restores the last save.
Retain existing wrappers and editor instances. Dev revision changes matching the
saved content version are acknowledged through authenticated editor context,
without a reload; differing source versions retain the usual reload protection.
Authenticated responses are private/no-store with the exact external editor script URL
and exact-origin activity/logout. Templates explicitly include admin/_toolbar.html
inside {#if app.current_admin}; the head explicitly includes its csrf-token meta.
Reserve app.current_admin for request identity, reject it in config, and clone App
per render. Never mutate shared Project data with session fields. The toolbar
mount carries rendered content version and dev revision and a declarative open
shadow-root shell from site/admin/_toolbar.html. Render both rows and escaped admin
identity with the page; keep the shell inert until behavior is attached. Never
clear/recreate this shell during bootstrap. Dialogs and external CSS remain cached
assets. An authenticated
CSRF-header-checked GET /admin/editor-context supplies metadata, rejecting unknown
routes. Bootstrap refuses a context version different from the rendered mount.
Static /admin/assets files contain no session data and use revalidating ETags.
Do not inject the toolbar, move its mount, or inline its libraries/styles into HTML.
Dev owner pages carry the
rendered snapshot revision and poll the existing reload endpoint without cookies.
Reload clean pages automatically and restore edit mode/scroll via short-lived,
tab-local sessionStorage. Never reload dirty drafts, saving/expired sessions,
open dialogs or an open Theme panel. Dirty drafts show a source-change notice;
canceling allows the pending reload. Production has no reload marker or polling.
Other site scripts and inline scripts remain blocked by the owner CSP.
Anonymous pages keep their original scripts and reload behavior. Assets/errors are not
decorated. Do not loosen CSP or add a login bypass to make testing easier.

Admin markup lives in developer-editable admin/index.html, code.html,
setup.html, and error.html, shipped by templates/. Compile these
with the selected admin.layout.html and render escaped admin.* data through
Server.RenderAdmin against the current immutable snapshot. Never compile editor
values as source. Exclude the entire admin directory from public routes/assets;
only editor state selects these templates. Missing files have no embedded fallback.
Without a layout, render the page as written. Preserve developer layout changes.
Use FleaCMS as the product name, without a space. /admin is sign-in only and
redirects authenticated owners to /. The old dashboard template and /admin/save
route are removed. Do not add a fallback. Browse using the site's navigation;
unlinked-page discovery remains deferred. Pages without snips still have Page details.
Sign-in CSP permits HTTPS styles/fonts/images and only the hash-pinned six-box
code-entry script on verification pages, with no frames. Keep the single code
field usable without JavaScript; enhancement supports paste, autofill and keyboard
navigation. Login emails include a prominent HTML code and plain-text fallback.
Never interpolate template or owner values into JavaScript.
Referrer-Policy must be same-origin,
not no-referrer: native browser POSTs otherwise send Origin: null. Keep exact
Origin matching and CSRF validation; never accept null or missing origins.

Interactive `flea dev` offers onboarding only on the first run, before app.site_name
is saved: it asks for the missing name, accepting its folder default on Enter, then
offers an optional admin who can sign in and edit. Ask once; a skipped admin prints
the `flea admin` next step and is not offered again. Never prompt for Resend
credentials or sender details during dev startup. New sites enable
dev.email_preview; the CLI-only preview sender handles both auth and contact
mail without provider credentials. Read docs/email-templates.md for its boundaries.
Confirm the site name with "Config file updated." and an added admin with
"Added <email> to .flea.config.json." Noninteractive stdin never prompts. Only
the explicit email wizard checks Resend. Reload after adding an owner, before
initializing sessions and serving. Keep secrets in ignored .env,
never site config. Malformed settings or wizard failures stop with an error.
Owner onboarding uses a developer-shared /admin link, not invitation emails or
separate account registration. User-facing copy calls these people "admins"
(matching flea admin and admin.emails); code, DOM ids and internal docs keep "owner".
The CLI explains applying access changes, sharing the live link, local mail setup,
and viewing/removing addresses via admin.emails. Local /admin shows a setup page
even with no owners; production remains disabled with an empty allowlist.
Missing-mail pages use the editor layout; locally direct developers to flea email
and returning to /admin, while production keeps an owner-facing message.
Keep request-code responses identical for approved and unapproved addresses.
One Save publishes all changed current-page regions and metadata together.

flea admin [email] appends to admin.emails, prompting when the argument is
omitted. Use the shared AdminAccess validator, skip case-insensitive duplicates,
and retain unrelated JSON fields with RawMessage and confined atomic writes.
Do not send email, commit, or deploy. Check effective local mail settings and
run the email wizard if missing or invalid only when dev.email_preview is off.
Never replace malformed configuration.
Keep output terse: confirmation, /admin, and the live commit/deploy reminder.
Dev uses the current validated owner list and rereads local mail per admin request.
Rebuild the local editor on access, idle-window or mail changes, discarding pending
codes but retaining valid sessions. Prune removed owners and expired sessions.
Production applies configuration on startup/deploy. Contact settings still require restart.
The email wizard asks for an asterisk-masked API key, sender name,
with Backspace/Ctrl-U editing and restored terminal mode after cancellation,
and separate verified-domain from address, then checks Resend without sending.
Inspect only safe provider error codes; never expose response bodies or secrets.
The documented 401 restricted_api_key response indicates a sending-only key;
403 restricted_api_key does not indicate success. Never claim delivery was tested.
Save only after a successful check, preserving unrelated .env settings, using
0600 permissions and a final /.env Git ignore rule. Refuse tracked .env files.
Successful email setup prints only the saved-file confirmation after checking.
Keep permission limitations in documentation, not routine success output; never
claim that the connection check verifies the sender domain or email delivery.
Local /admin must show actionable environment setup instructions when email is
unconfigured; production keeps an owner-facing message without local shell steps.

`admin.emails` in .flea.config.json is a developer-owned list (max 20), empty by
default. /admin supports email OTP; the public-page editor supports title/description
rich content, designated images/videos, collections and supported theme values.
Templates, schemas, archived entries and developer configuration remain private.
Codes use crypto/rand, expire in 10 minutes and allow five attempts. Store hashed
codes in bounded expiring maps. Sessions use admin.session_idle_days (omitted/zero
means 30, otherwise 1–365). Persist only token hashes, email, host, CSRF, last activity
and expiry. Store up to 2048 sessions in a bounded 2 MiB JSON file, 0600, at local
flea/runtime/sessions.json or production sites/<name>/flea/runtime/sessions.json outside releases.
The production runtime directory is flea-owned 0700; site parents/releases remain
root-owned. Use confined reads, reject links and unsafe permissions, and lock
read-modify-write with .sessions-lock. Sync an exclusive private temporary file,
rename, and sync the parent. Never silently reset corrupt/unwritable storage.
Each transaction rereads disk; stale handlers cannot restore revoked sessions.
Keep session locking separate from the content/deployment lock.
Prune against current owners and expiry at startup, local handler rebuild and
session access. Host.Load is read-only, including candidate validation; actual
host startup initializes/prunes session state. Deployment creates runtime directory
permissions without copying sessions. Sessions survive restart and release changes;
logout revokes the current token, owner removal revokes their tokens.
Authenticated page views/forms renew cookie and disk expiry. The activity script
renews on trusted visible input/click/key/scroll events, at most once per minute,
never with an idle timer. Its POST requires the same Origin and CSRF as saves.
Do not require email delivery to remain configured for already authenticated users.
Dev ensures session ignore rules before creating state, refuses Git-tracked files,
and status reports the local path. Runtime inclusion excludes sessions/lock/temp
files from watchers, snapshots and archives; underscore URL privacy still applies.
Require exact Origin and CSRF checks on every POST, host-bound HttpOnly/SameSite
cookies, HTTPS and __Host- cookies in production. Local dev permits loopback HTTP
and explicit `--lan`/`-l` requests authorized by devaccess middleware. Keep the
same checks for authenticated public rendering and admin routes. LAN is HTTP on
a trusted private IPv4 network, never a login bypass. Do not trust forwarded headers. Rate-limit requests by
address, actual client IP and site. Do not reveal allowlist membership or log codes.

Resend uses mailer.Sender, not an SDK-specific dependency in auth code. Delivery is
asynchronous with a timeout, and bounded by the send limits; errors never contain
provider bodies, tokens or recipients. Deployment prompts for admin-enabled or contact-enabled
sites lacking saved mail settings. Stream secrets over SSH without a PTY into a
private staging file; sudo may then consume it with an interactive terminal.
Store /etc/flea/mail.env root-only (0600), loaded by systemd EnvironmentFile.
No secrets in site config, deploy settings or Git. Local dev and deployment mail
setup use localenv for bounded literal .env reads; process environment values win.
Never execute or expand .env values, mutate the global process environment, or
load .env into the production host. FLEA_EMAIL_NAME supplies the sender display
name. Keep Sender and setup Checker interfaces separate from the Resend adapter.
Admin addresses and sign-in mail settings update without a dev restart.

Protocol 22 also requires the service unit and lock/content/runtime permissions, not
just a binary swap. Do not reset a live installation as an upgrade. Setup creates
the lock owned by flea, and activation grants write access only to each content
and private runtime directory. Keep release/config parents root-owned. Browser edits and extraction
share an in-process mutex; remote edits and activation share deploy.lock and refuse
pending recovery journals. Test stale forms, single-use codes, CSRF, isolation,
sanitization, content preservation and concurrent requests. Live email delivery
requires real credentials and remains unverified unless actually exercised.

Tests cover real local HTTP/TLS, a local ACME issuance flow and offline cache
reuse, activation failure/recovery, archive attacks, and simulated SSH/SCP.
These do not establish live Let's Encrypt issuance, long-term renewal, or real
Ubuntu/systemd deployment. Record those as pending until exercised on a target.

## Prototype and shared interaction rules

The isolated inline-editing experiment lives in `design/inline-editor/`.
Run `go run ./design/inline-editor` for its loopback-only preview. It uses pinned
Squire and DOMPurify with simulated identity and tab-only saves, not real site
content. The runtime has authenticated inline editing with atomic content/metadata
saves, H1–H4, marks, lists, links and alignment. See its README for browser checks.
The prototype edits the current page with a right-aligned Edit toggle, not a page
menu. Edit becomes Cancel/Save; Sign out stays visible after a divider.
Page details and contextual formatting
appear in a second bar. Unlinked-page discovery is deferred. Serve the selected
local Lucide SVGs without adding npm or a frontend build step.
Prototype dialogs use compact titles, a bold text-link-style Cancel button on
the left, and the primary action on the right. Cancel discards dialog changes.
Keep the second bar visible whenever signed in, including browsing and other modes,
so leaving Edit does not collapse the toolbar. Disable Page details and formatting
outside Edit; within Edit, formatting requires an active text region. No content
outlines, including on keyboard focus. Modal motion and page-content view transitions
are subtle and respect reduced motion.
Keep dialogs open until their exit animation completes, including Cancel, Escape
and submit. Page crossfades use explicit final-opacity fill and hide the old root
snapshot; rapid navigation finishes/skips the prior transition before starting anew.
Do not add undo/redo toolbar buttons; keyboard history remains available.

A complete six-digit code pasted into the sign-in boxes submits the form using
requestSubmit, preserving native validation and CSRF. Partial paste and manual
typing retain the Sign in button and Enter behavior. Never auto-submit saved values.
The link button opens the current URL for editing when inside a link, with a
separate Remove link action in the dialog. Preserve link text and formatting.
Prototype links default bare domains to HTTPS and preserve explicit schemes and
relative paths. The new-tab option defaults off and stores only _blank with
noopener noreferrer. The production sanitizer permits this exact new-tab policy.
Reusable CSS-only switch visuals belong in templates/site/css/ui.css, targeting native
checkboxes with role="switch"; preserve keyboard, focus and reduced-motion behavior.
