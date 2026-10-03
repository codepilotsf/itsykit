# FleaCMS

## Install

The planned public installation command is:

```sh
curl -fsSL https://fleacms.com/install.sh | bash
```

Public downloads are not published yet. Until then, authorized source-checkout
users can build with Go 1.26 or newer:

```sh
bash scripts/build.sh
export PATH="$PWD/bin:$PATH"
```

The release installer will use `~/.local/bin` without sudo and print any needed
PATH instructions. See [installation and updates](docs/installation.md).

## Create your first site

```sh
flea new foobar
cd foobar
flea dev
```

FleaCMS is a small website server for people who like HTML and CSS. Sites stay in
ordinary files, with browser-editable content and Git-based deployment to Ubuntu.
No Node, Docker, database, or separate web server is needed.

By default, you get a clean layout, mobile menu, editable homepage, default
styling, and admin/email support files. No demo pages or contact configuration.

Want examples to work from?

```sh
flea new --starter foobar
# Short form: flea new -s foobar
```

The starter adds Team and Contact navigation, six fictional team members with
replaceable photo placeholders, a contact form, a thank-you page and an unlinked
`/test` example. Form fields and navigation destinations remain developer-managed.
Both modes include narrow, single-purpose sign-in screens and styled error pages.
Their colors and typography follow the site theme; the standalone server-error
page keeps its own minimal styles so it works without site assets.

## Start fresh

Run `flea reset` from a site folder to return to the clean default. It asks before
replacing **all local files and settings except `.git`**, including uploads,
custom pages, `.env`, admin access, theme changes and deployment destinations.
Stop `flea dev` first. Git history is retained, but reset does not create a commit.

Old files are moved to a private `.flea-reset-backup-*` folder beside the site;
the command prints its location. Keep it safe: it may contain credentials and
private content. Nothing is deployed or changed remotely. Linked Git worktrees
and symlinked site roots are refused. If interrupted, stop and recover from the
printed backup before trying again; some original files may still be in the site.

## Develop locally

On first startup in a terminal, Flea asks for the site name and an optional admin
email. Press Enter to skip the admin. New sites enable `dev.email_preview` in
`.flea.config.json`: sign-in and contact emails open in your browser instead of
being sent. No Resend setup is needed. The preview shows the sender, recipient,
subject, date/time and actual message, including a usable sign-in code.
Automated runs never prompt. Use `flea admin` to add an admin later.

Running `flea dev` again gracefully replaces this site's existing dev server,
including when switching to `--lan` / `-l` or changing ports. The new command
waits for the old server to stop before opening the site. Pull and reset still
block dev while they work. Dev servers started with an older binary need to be
stopped in their terminal once before using automatic replacement.

Set `"dev": {"email_preview": false}` and restart dev to test real delivery;
configure it explicitly with `flea email`. Production always uses real delivery
and ignores this toggle. See [email previews](docs/email-templates.md) for details.

Open http://localhost:3000, edit `site/index.html`, and save. Your browser reloads
automatically. Change the welcome text in `flea/snips.json`, or give the site
your own colors and fonts in `site/css/theme.css`. Press Ctrl-C to stop the server.
Admin sign-in pages also refresh when their templates or layout change in dev.
Signed-in pages reload too, including newly extracted snips, and retain edit
mode when there are no unsaved changes. Drafts, open dialogs and Theme previews
pause reloads so they are not lost. Cancel a draft to accept pending source changes.

To test from a phone or another computer on the same trusted network, run
`flea dev --lan` (or `flea dev -l`). Open one of the printed Network URLs on that
device. Sign-in, editing, uploads, Theme, forms, reloads and dev diagnostics work
there too. Each device signs in separately. Email previews still open on the
computer running Flea; enter that code on your other device. To share a private
preview, replace its `127.0.0.1` host with a printed network IP, keeping the port
and random path. Never share preview links with anyone you do not trust.

LAN mode uses unencrypted HTTP and detected private IPv4 addresses. Use it only
on a trusted network, never forward its port, and restart after changing networks.
VPN addresses may also appear. Allow the chosen port through your computer's
firewall if needed; guest Wi-Fi may block device-to-device connections.
Without the flag, dev remains accessible only from this computer.

## Everyday commands

Run these from your site's folder:

| Command | What it does |
| --- | --- |
| `flea dev` | Preview your site at http://localhost:3000 with browser reloads. |
| `flea dev -port 8080` | Preview on a different port. |
| `flea status` | Check configuration, routes, templates, and content without starting a server. |
| `flea reset` | Reset the local site to the clean default after confirmation, keeping a private backup. |
| `flea deploy` | Configure the current branch's destination, then publish its committed files. |
| `flea deploy -configure` | Change that branch's saved deployment settings and deploy. |
| `flea https` | Enable HTTPS for an already deployed site. |
| `flea version` | Show the installed version. |

The preview server is local to your machine. Use `-port 0` to choose an available
port automatically. `status` works offline; missing icons produce warnings while
otherwise valid sites still pass.

`flea new my-site` creates a Git repository on `main`, ready for your first
commit. Use `flea new my-site --skip-git` if the site belongs in an existing
repository. Git isn't required for local previews, but it is required to deploy.

Start with [project structure](#project-structure) and [templates](#templates)
below. When you're ready to publish, see [deploy a site](#deploy-a-site) and
[server setup](#set-up-an-ubuntu-server-from-your-mac).

## Project structure

```text
my-site/
  .flea.config.json       App, admin, contact and deploy.branches settings
  flea/
    snips.json           Rich content and page metadata
    media/               Managed photos and optimized videos
    runtime/             Private sessions and extraction recovery
    skills/              Portable agent instructions
    data/                Optional developer-managed JSON
    icons/               Downloaded Lucide cache
  mailers/
    main.layout.html
    contact.email.html
    contact.email.txt
    sign-in.email.html
    sign-in.email.txt
  site/
    main.layout.html     Optional global wrapper
    admin.layout.html    Sign-in page wrapper
    admin/               Editable sign-in templates and toolbar partial
    _404.html            Optional not-found page using the layout
    _500.html            Optional standalone error document
    index.html           /
    css/global.css       /css/global.css
    images/mark.svg       /images/mark.svg
  .gitignore
  AGENTS.md
  README.md
```

Start with `site/index.html` for your home page. Add `site/about.html` and it's
available at `/about`. Put shared markup in `site/main.layout.html`, styles in
`site/css/`, and developer-supplied images in `site/images/`.

Every site needs `.flea.config.json`; JSON snips are optional. The starter gives
you a welcome page, a layout with a header and footer, error pages, an SVG logo,
and CSS you can make your own. It loads Google Fonts and Prewind from external
services.

Only `site/` is the document root. Dotfiles/directories, `node_modules`,
`README.md`, `AGENTS.md`, `ARCHITECTURE.md`, `spec.md`, `next-steps.md`, `LICENSE`, `LICENSE.md`,
`go.mod`, and `go.sum` are excluded from serving, reloads, and deployment
(these metadata names are case-insensitive). Keep other development files or
secrets in ignored `.env` files. Mailers are loaded privately, never served.
Managed media is served by its catalogue under `/media/`, not `/flea/media/`.
`flea/runtime/` and `flea/skills/` are never served or deployed. Runtime is ignored
by Git; skills remain committed. Deployment settings live under `deploy.branches`
in `.flea.config.json` and are stripped from release archives and backups.

This structure is required; there is no old-directory fallback or automatic
migration. Existing projects need an explicit, collision-checked reorganization.
Keep template-relative paths and public URLs unchanged. Update the server before
deploying: protocol 22 requires the new structure.

## Templates

The signed-in formatting row remains visible outside Edit to keep the toolbar's
height stable. Its controls are disabled until Edit is active; text formatting
also requires an editable region to be focused. Pasting a complete six-digit
sign-in code submits it automatically. Typing and partial pastes still allow Enter
or the Sign in button.

Write the shared HTML once in `main.layout.html`, then use `{slot}` where each
page's content should appear. A page only needs the markup that goes inside
that slot. If you don't need a shared layout, leave it out and write complete
HTML pages instead.

```html
<!-- main.layout.html -->
<!doctype html>
<html lang="en">
  <head><title>{page.title} | {app.site_name}</title></head>
  <body>
    <header><a href="/">{app.site_name}</a></header>
    {slot}
  </body>
</html>
```

To select a custom layout, put this settings comment first in a page:

```html
<!-- flea
layout: /admin.layout.html
-->
<h1>Page content</h1>
```

Paths are relative to `site/` and must name a `*.layout.html` file. All layout
files are private, even when no page selects them.
The layout uses `{slot}` for the page and can include relative partials. Without
the comment, `main.layout.html` remains the default. One block is allowed; `layout`
is currently the only setting. Unknown/duplicate settings, missing layouts,
traversal paths, settings in partials, and nested layouts are validation errors.
Selected layouts cannot also be included as partials. Ordinary comments stay inert.

Textareas use native CSS `field-sizing: content` and grow up to 400px, then scroll.

The not-found page, `_404.html`, uses the layout too, so write it as page content.
The server-error page, `_500.html`, is a standalone HTML document; keep it simple
so it can work even when the rest of the site has a problem. Both are optional.
Production uses these custom pages, or a basic safe fallback, and writes details
to server logs. Local dev instead shows the diagnostic page described below.

Page navigation uses native CSS view transitions: a 240 ms crossfade between
same-origin pages. A single direct-child `<main>` fades independently of the
header and admin bar; other layouts get a whole-page fade. Reduced-motion settings
disable the effect. Browsers without cross-document transitions navigate normally.
The starter links `css/navigation.css` directly in the layout head, before the
general stylesheet. A small native-CSS opt-in in each layout head makes it
available before external styles finish loading, including the reduced-motion
opt-out. Animation styling stays in ordinary editable site CSS; Flea does not
inject transitions or require npm.
This affects link navigation, not the automatic dev refresh after saving a file.

Partials are underscore-prefixed HTML files anywhere under site/. `{render footer}` includes `_footer.html` beside the calling template. `{render /foo/bar/footer}` includes `foo/bar/_footer.html`. Relative directory paths, including `./` and `../`, resolve from the calling template; nested renders resolve from the partial containing them. Paths are unquoted and omit the filename's underscore and `.html`. Paths escaping site/ are rejected. The root `main.layout.html`, `_404.html`, and `_500.html` are reserved, not partials. This replaces the former `{partial "name"}` syntax and `_partials/` convention.

Missing and recursive includes fail validation, even in unused partials. Includes inherit the surrounding loop aliases. `page.path` is the canonical route. `page.title` and `page.description` come from that route's metadata in `flea/snips.json`. Without a title, the default is `Home` at `/`, otherwise the final route component with hyphens replaced by spaces and its first letter capitalized. Description defaults to empty. Metadata is contextually escaped.

```html
<h1>{snip.home.title}</h1>
{#if app.email}
  <a href="mailto:{app.email}">Email us</a>
{:else if app.phone}
  <p>{app.phone}</p>
{:else}
  <p>Contact details coming soon.</p>
{/if}
{#each snip.home.services as service}
  <article><h2>{service.name}</h2><p>{service.description}</p></article>
{/each}
```

Values and conditions accept dotted paths only. Missing values render empty. Missing/null values, false, zero, empty strings, empty arrays, and empty objects are false in conditions. Empty or missing/null collections render no loop body; other non-array collections are errors. Nested loops and alias shadowing work; the outer alias is restored afterward. `app`, `page`, `data`, and legacy `snip` are reserved root names and remain accessible inside loops. Aliases start with a letter and contain letters, digits, or underscores.

`{snip a7Km2pQ9xB}` renders sanitized rich HTML from `flea/snips.json`, only in HTML element content. Ordinary dotted-path values remain contextually escaped, including attributes and URLs. Legacy `{snip notice}` still reads the escaped `text` field from `flea/snips/notice.json`; legacy structured snips remain supported during migration. JSON content is never parsed as template syntax. There is no unrestricted raw-HTML directive.

There are no expressions, comparisons, functions, filters, array indexes, or arbitrary code execution. Braces outside script/style elements are template directives; use `&#123;` and `&#125;` for literal text braces. Script/style bodies and their opening tags are left untouched by the FleaCMS parser, so they do not support interpolation. The internal delimiter `[[NW` is reserved and rejected in template source. Developers are trusted to author HTML, JavaScript, CSS, and public assets; snip JSON is treated as data.

Loops can include an empty-list fallback:

```html
{#each snip.team.members as member}
  <p>{member.name}</p>
{:else}
  <p>No members yet.</p>
{/each}
```

The fallback also runs for missing/null arrays. It has no current item; any
outer alias with the same name is restored. `{:else if dotted.path}` is only
valid in conditions. A final `{:else}` cannot be followed by another branch.

Use `{#unless dotted.path}` for the inverse of an if condition:

```html
{#unless app.current_admin}
  <a href="/admin">Admin sign-in</a>
{:else}
  <p>You are signed in.</p>
{/unless}
```

Missing/null values, false, zero, empty strings and empty arrays are falsey.
Unless also supports `{:else if dotted.path}` with the usual positive condition.
Expressions remain dotted paths, not JavaScript.

## JSON configuration and content

Editing starts on after sign-in and remembers your choice as you browse in the
same tab. Its on state is green; thicker green dashed boundaries with rounded
corners show editable content. Cancel and Save sit
at the right of the second toolbar row and keep editing on.

Collections are repeatable groups such as team cards. While Editing is on,
arrange them directly on-page: drag handles reorder items, X removes
one, and an Add tile creates another. Keyboard users can focus a handle and use
arrows or Home/End. Photo fields accept existing library images or new JPEG/PNG
uploads, saved with the page. Text and photo selection remain on-page, using the same Save
changes/Cancel controls. See [multisnips](docs/multisnips.md) for authoring/storage.

### Editable snip regions

Editable content uses `snip`, `{snip ID}`, and `flea/snips.json` with `snips`
JSON keys. Existing sites using `flea` content markers or `fleas.json` need these
renamed explicitly; there is no compatibility layer. Keep IDs and saved content
unchanged. The CLI, `flea.config.json`, and `<!-- flea` settings comments retain
their names.

While `flea dev` is running, add an editable region to a template:

```html
<h1 snip>Welcome to <em>FleaCMS</em></h1>
```

On save, dev extracts the inner HTML and rewrites only that region:

```html
<h1>{snip a7Km2pQ9xB}</h1>
```

IDs are cryptographically random, case-sensitive, 10-character alphanumeric
strings, checked against active and archived local IDs. IDs never contain braces,
quotes or whitespace. Use `{snip ID}` without quotes. Quoted markers are rejected;
HTML comments do not render snips. The bare attribute needs no name or value.
An optional attribute value supplies a descriptive label, never an identity.
The attribute is only an authoring marker: extraction removes it, preserving
the element and its other attributes. The directive identifies the content from
then on. Redundant attributes around existing directives are removed on the next
dev load/save without changing the ID or saved content. Stored labels remain
unless you explicitly supply a new label. Inline editing uses the rendered
directive boundaries, not a persistent HTML attribute.
Reusing an ID intentionally shares content. IDs referenced from
multiple routes, layouts, or partials are grouped under `shared`.

Images and videos use the same authoring attribute, outside text snips:

```html
<img snip src="/images/team.jpg" alt="Our team" class="rounded-lg">
<video snip src="/videos/intro.mp4" poster="/images/intro.jpg" controls></video>
```

Dev imports these local files and replaces their source/alt/poster values with
bindings such as `src="{snip ID.src}"`. The generated typed JSON entry is the
source of truth. Classes, dimensions and playback attributes stay in the template.
With Editing on, admins click the media, then Save or Cancel with the page.
JPEG/PNG images and optimized H.264 MP4 videos are supported. Nested snips,
srcset/picture combinations and competing video source elements are rejected.
See [images](docs/photos.md) and [videos](docs/videos.md) for details.

Private `flea/snips.json` stores rich content, typed media and page metadata:

```json
{
  "pages": {
    "/": {
      "title": "Home",
      "description": "Welcome to our site.",
      "snips": {
        "a7Km2pQ9xB": {
          "content": "Welcome to <em>FleaCMS</em>"
        }
      }
    }
  },
  "shared": {"snips": {}}
}
```

Use a `div` or `section` region for multiple paragraphs. Formatting includes
paragraphs, emphasis, lists, headings h2–h6, blockquotes, code, and safe links.
Scripts, styles, event handlers, images, forms, embedded content, SVG/MathML,
and arbitrary attributes are removed. Allowed link schemes are HTTP, HTTPS,
mailto, tel, and relative URLs; protocol-relative links are removed. Rich snips
cannot appear in attributes, script/style interpolation, textarea, title, or
select contexts. Ordinary `{data.name.field}` values remain escaped text.

New text regions must have explicit balanced closing tags, cannot nest, and cannot
contain FleaCMS directives. Use HTML entities for literal braces during initial
extraction. Existing JSON content can contain braces: it is never template code.
Snip files are limited to 5 MiB and each rich content value to 1 MiB.

Extraction happens only in dev, including startup. Status is read-only. Deploy
rejects unextracted regions and asks you to run dev and commit the resulting
files. Every relevant save rescans all templates, extracts new regions, updates
labels and route ownership, and moves unreferenced IDs to `archive`. Undoing a
removal restores archived content by ID. Existing referenced content and page
metadata are preserved. Dev prints `Snips updated in flea/snips.json.` when it writes
changes. Missing IDs and malformed regions are reported in the terminal.

Local content writes are coordinated in memory. A separate OS lock prevents dev,
pull and reset from writing concurrently. Starting dev again gracefully stops
the previous dev server before taking that lock; a crash releases it automatically.
Server-side deployment/content locks are unchanged.
A `flea/runtime/.snips-pending.json` recovery record saves the content and
before/after templates before any replacement. Dev resumes interrupted writes;
if files were independently edited, it stops rather than overwriting them.
Reconcile those files with the recovery record and restart dev. Do not delete
the recovery record without preserving its content.

Developer-managed structured JSON belongs in `flea/data/*.json`, available as
`{data.team.members}` and in `{#each data.team.members as member}` loops. It is
deployed with code, not treated as admin-editable rich content. Legacy
`flea/snips/*.json` remains supported; new sites no longer generate that directory.

### Site settings

Site settings must use root `.flea.config.json`. Existing projects need the full
structure described above; no old filename or directory fallback is supported.

`.flea.config.json` stores public site-wide variables in `app`:

```json
{
  "app": {
    "site_name": "FleaCMS",
    "email": "hello@example.com"
  }
}
```

Other fields within `app` are available to templates. `admin.emails` controls admin
access as described below. Other unknown top-level settings are ignored; redirects
and deploy settings have no effect here. JSON comments and trailing commas are errors.
Rich snips live in `flea/snips.json`; structured data lives in `flea/data/*.json`.
Use `{app.site_name}`, `{app.email}`, or your own `{app.phone}` in any page or
layout. Keep secrets out of `app`. Page-specific values stay in `page.*`, and
structured files stay in `data.*`; admin configuration is not exposed through `app`.

New sites start with an empty `app` object. Until a name is saved, the folder name
supplies its title-cased default (`my-project` becomes `My Project`). The first
interactive `flea dev` asks `Site name [My Project]:`; Enter accepts the default.
It saves `app.site_name` and does not ask again. Noninteractive dev uses the default
without writing or prompting. Set and commit a nonempty name before deployment.
There is no `site.*` alias or old-configuration fallback.
Legacy named snips in `flea/snips/*.json` remain supported but are not browser-editable.

## Admin editing

Repeatable, page-local content is documented in [Multisnips](docs/multisnips.md).
For developer-editable sign-in/contact messages and offline previews, see
[Email templates](docs/email-templates.md).

### Theme editing

Every approved admin can choose **Theme**, beside Editing in the topbar. A right-hand
panel groups the existing variables in `css/theme.css`: colors, typography,
sizes, spacing, layout, forms, corners, borders, and shadows. Controls share the
topbar's sizing and interaction styles, independently of the website theme.
Theme temporarily suspends page editing and replaces the second-row page actions
with controls inside the panel: a header X and a fixed-footer Save button. Escape
also closes Theme; unsaved changes require confirmation. Closing Theme restores
your editing preference.
Save stays visible while the settings scroll. The panel fades/slides in and out,
with immediate transitions when reduced motion is enabled.
Font Color offers neutral colors followed by brand colors, without semantic
colors. The starter defaults to Black; selections stay linked to theme tokens.
Page colors groups Font Color and Background Color together. Background Color
defaults to Lighter and controls the starter page background through --bg-color.

Changes preview only in that browser until **Save**. Closing restores the
previous appearance. Switching between text editing and Theme asks before
discarding unsaved changes. Failed or stale saves keep the preview available.
The editor updates supported existing custom properties in one `:root` block;
it preserves imports, selectors, and unrelated CSS. Text and spacing scales use
paired small-screen/large-screen base sizes and ratios, with screen widths under
an expandable section. Changes immediately preview the complete scale. Save
regenerates the CSS in Go and updates its settings comment in `theme.css`; no
Node or npm is needed. Custom scales require explicit replacement before using
the generator. Other complex values retain value fields. Font fields accept
font-family text values. Google Fonts load automatically in your preview; Save
updates a managed import in theme.css and Cancel removes the preview stylesheet.
Existing developer imports are preserved. System fonts need no imports. Flea
requests the regular face; browsers may synthesize bold and italic. No API key
or Node installation is needed. Info buttons explain settings, and only one Theme
section opens at a time.
Unsupported CSS values remain developer-managed rather than being rewritten.

Sign in at `/admin`, then browse your site normally. Choose Edit in the admin
bar to change the current page. Page details edits its title and description,
including pages with no editable text. There is no sidebar or page dropdown.

New sites include editable `admin/index.html`, `code.html`,
`setup.html`, and `error.html`. These are ordinary Flea HTML templates using
`admin.layout.html`, with a full-width header, no footer, and independently
styled sign-in forms. Your public `main.layout.html` is separate.
Existing sites need the admin templates, admin layout, and updated styles copied
from a freshly generated site; there is no embedded fallback.

Change the markup directly, keeping form actions, field names, and hidden
`{admin.csrf}` values intact. Dynamic values use escaped `admin.*` data.
The entire directory is excluded from public routes and asset serving: the
login handler selects the page and enforces access, not the HTML file.

Admin pages
allow HTTPS styles, fonts, and images, but block scripts and embedded frames.
Layout navigation that requires JavaScript will not run on admin pages.
Use `Referrer-Policy: same-origin` for admin forms; `no-referrer` can cause
browsers to omit their origin and fail the exact-origin security check.

From your site's folder, run `flea admin` and enter an email, or provide it directly:

```sh
flea admin admin@example.com
```

This creates or appends to `admin.emails` in `.flea.config.json`, preserves other
settings, and skips addresses already present (case-insensitive). It accepts up
to 20 admins and does not send email, commit, or deploy. Missing or invalid local
mail settings start the email wizard automatically only when preview mode is off.
Local access updates through
the dev watcher; commit and deploy to update the live site. You can also edit JSON:

```json
{
  "app": {"site_name": "My site"},
  "admin": {
    "emails": ["admin@example.com", "editor@example.com"],
    "session_idle_days": 30
  }
}
```

The starter has an empty list. Local `/admin` shows instructions for adding the
first admin; production admin access stays disabled until you configure it.
There is no registration, password, or user-management screen. All listed addresses
have the same editing access. Up to 20 unique addresses are supported; matching is
case-insensitive. Only the developer can change the list.

To view or remove admins, edit `admin.emails` in `.flea.config.json`. Dev reloads the
list automatically. Removed admins lose access; other admins stay signed in.
Access or email-setting changes clear pending codes, not valid sessions.
Commit and deploy for the live site.

After login, Flea returns you to the homepage with an admin bar. Choose **Edit**
to change the page's existing `snip` regions. Prose containers support
H1–H4, bold, italic, underline, strikethrough, links, lists, and left/center/right
alignment; individual headings and paragraphs support inline marks and links,
without nested headings or lists. Cancel restores the page.

Save publishes all changed regions and page details together without reloading
the page. Editing stays on, and Cancel restores your most recent save. Save stays
disabled until something changes. Shared content is marked in the toolbar and
updates everywhere its ID appears. Repeated instances on the page stay in sync.
The link dialog edits the current URL, offers Remove link and a new-tab switch,
and defaults bare domains to HTTPS. There are no undo/redo toolbar buttons.
The old dashboard and its save route have been removed; `/admin` is for sign-in.

Visitors receive no editing bar or editing markers. Signed-in pages disable site
JavaScript. During dev, clean pages reload after source changes; unsaved edits,
active saves and open dialogs defer reload to protect drafts. Site layouts are not
rewritten; fixed-position headers and unusual layouts still need browser checks.

Reusable toggle styling is in `css/ui.css`: use a native checkbox with
`role="switch"` and an associated label. Its appearance and keyboard behavior
need no JavaScript.

After deploying, share your site's URL with `/admin` added, along with the email
address you approved. Flea does not send invitations. Admins do not create a
separate account: they enter that email and request a sign-in code. If no email
arrives, they should check spam and ask you to check access and delivery settings.

Visit `https://your-domain/admin`. Enter an approved email, then its six-digit code.
Paste the full code into any of the six boxes, or type it one digit at a time.
The email displays the code prominently and includes a plain-text alternative.
Codes expire after 10 minutes, allow five attempts, and work once. Sessions survive
browser closure, server restarts, and deployments. Cookies are host-only,
HttpOnly, Secure in production, and SameSite=Strict. POST requests require the same
Origin and a CSRF token. Login requests have per-address, client-IP and per-site
limits (3, 10 and 40 per 15 minutes). Unlisted addresses receive the same generic
response but no email. Forwarded IP/TLS headers are not trusted.

### Staying signed in

`admin.session_idle_days` controls how long an admin stays signed in without
activity. The default is **30 days** (omitted or zero); set 1–365 days. Opening an
admin page or submitting an authenticated form renews that browser's session.
Typing, clicking, and scrolling also renew it, at most once a minute. An idle or
hidden tab does not keep a session alive. Each browser/device has its own session.
Changing this setting applies automatically in dev and on deployment in production;
a shorter window also applies to existing sessions.

Flea manages one private JSON file per site:

- Development: `<project>/flea/runtime/sessions.json`
- Production: `/var/lib/flea/sites/<site>/flea/runtime/sessions.json`, outside releases

These are independent local and live sessions, not files to copy between environments.
The file contains hashed session tokens, admin addresses, host binding, CSRF values,
and activity/expiration timestamps. Raw login tokens stay in browser cookies.
Flea creates it on first sign-in with owner-only permissions, locks updates, and
saves atomically. Expired or revoked entries are removed during use/startup.
Malformed or unwritable session storage blocks sign-in rather than resetting it.
Pending email codes are held only in memory and expire on restart.

Session files are never publicly served or included in deployment archives.
The starter ignores `/flea/runtime/sessions.json` and `/flea/runtime/.sessions-*`; dev ensures those
rules exist and refuses already tracked session files. Do not commit, edit, or
copy session state. To sign everyone out, stop the server and delete only that
site's `flea/runtime/sessions.json`, then start it again. Logout removes just the current
browser's session. Removing an admin revokes their sessions. `flea status` shows
the local session path when admins are configured.

The production service can write the site's private `flea/` directory, not its
release directory or configuration. Deployment protocol 22 requires the updated
server and these permissions; there is no compatibility fallback for older servers.

### Editing content

The toolbar is explicitly included in the layout. Add this inside `<head>`:

```html
{#if app.current_admin}
  <meta name="csrf-token" content="{app.current_admin.csrf_token}" />
{/if}
```

And place this at the start of `<body>`:

```html
{#if app.current_admin}
  {render /admin/toolbar}
{/if}
```

`admin/_toolbar.html` contains a readable declarative shadow-root toolbar shell
and an external editor script. Both toolbar rows render with the page, before
editor requests finish; initialization attaches behavior without replacing them.
Keep its version/revision attributes: they bind edits and reloads to the page
you actually received. Dialog HTML, CSS and JavaScript load as separate cached
assets; content metadata is fetched through an authenticated, CSRF-checked JSON
endpoint. No runtime-injected toolbar, inline libraries/styles or content JSON
are inserted into page source. The shell is inert until initialization succeeds.
For existing sites, copy the current starter `site/admin/_toolbar.html` partial;
do not replace your customized layout.
The editor stays where the partial is placed; use a direct child of `<body>`.

`app.current_admin` is absent for visitors and contains `email`, `csrf_token`,
`version` and the dev-only `revision` for authenticated page requests. It cannot
be set in config. CSRF tokens are deliberately visible in signed-in source, not
session cookies. Responses containing admin data are private and never cached.
Sign-in forms retain their explicit hidden CSRF fields. Signed-in pages permit
the external editor script; other site scripts and inline scripts remain blocked.

Existing sites must add these blocks and copy `admin/_toolbar.html` from the
starter. For transitions, copy `css/navigation.css` and link it in your layout head.
Flea never rewrites your customized layout or provides an implicit fallback.

Edit content directly on the page. Spans and CSS classes on supported content tags
are preserved, including `<span class="text-brand-one">Brand text</span>`.
There is no HTML-source editor; author custom markup in the template before snip
extraction, or in the snip's JSON content. The HTML allowlist removes unsupported markup
on save. Only left/center/right text alignment is allowed as CSS; new-tab links
receive `noopener noreferrer`. Page details remain escaped. Saving validates the site, atomically writes
the managed content file, and updates the live serving snapshot without restart.
Stale drafts, including drafts predating a template change, are rejected instead of
overwriting newer edits. A failed save leaves the draft in your browser with an
error message; copy it before reloading. Archived snips,
templates, access settings, deployment settings, and structured JSON are not editable.
One Save publishes the current page's changed content and metadata. Cancel restores
the page, and leaving with unsaved changes triggers a browser warning.

### Email credentials

Resend is the only implemented provider. The code uses its
[send-email API](https://resend.com/docs/api-reference/emails/send-email) through a
small provider-neutral `mailer.Sender` interface, without adding a provider SDK.
Use a sending API key and a sender address from a verified domain in your Resend account.

On deployment of an admin-enabled or contact-enabled site, FleaCMS checks the
server's mail settings. If missing, it uses local `.env` settings, with shell
variables taking precedence, and prompts for any missing key or address. The
sender name is included when configured. Existing server mail settings are not
overwritten by `flea email`. Subsequent
deployments reuse them. Settings are shared across sites on that server and stored
in `/etc/flea/mail.env` (root-owned, mode 0600), outside site archives and Git.
systemd reads that file and passes the values to the unprivileged service. The key
travels over SSH stdin to a private temporary file, never as a command argument.
The temporary file is removed on normal completion; an interrupted connection may
leave a mode-0600 file in its mode-0700 `/tmp/flea-mail.*` directory.

From your site folder, run:

```sh
flea email
```

The wizard asks for a masked Resend API key, a sender name (such as MySite
Support), and a separate from address (such as hello@your-verified-domain.com).
It checks the Resend connection without sending email, then saves
`RESEND_API_KEY`, `FLEA_EMAIL_NAME`, and `FLEA_EMAIL_FROM` to `.env` with
owner-only permissions. Unrelated settings are preserved and `.env` is excluded
from Git. A tracked `.env` must be removed from Git tracking first.

The check uses Resend's domains endpoint. A sending-only key's documented
restricted-access response confirms connection but cannot verify the sending
domain. A successful check is not a delivery test; verify your domain in Resend.
Successful setup simply confirms that settings were saved to `.env`.

Open `/admin` on your dev server. Admin sign-in rereads `.env`
automatically; explicitly set shell variables take precedence, including empty
values. There is no need to source the file. Parsing supports literal NAME=value
lines, single/double quotes, and full-line comments, with no variable expansion
or shell execution. Double-quoted escapes use JSON/Go string syntax. Only the
mail settings are consumed; this does not export arbitrary variables to commands.
Never place keys in `.flea.config.json` or commit them. No restart is needed for
admin access or sign-in email changes. Contact forms still need a dev restart.
Authentication still requires an emailed code.

To rotate server credentials, edit `/etc/flea/mail.env` as root, keeping the
following quoted assignments and mode 0600, then restart `flea.service`:

```dotenv
RESEND_API_KEY="re_your_key_here"
FLEA_EMAIL_NAME="MySite Support"
FLEA_EMAIL_FROM="login@your-verified-domain.com"
```

Production admin access requires HTTPS. An admin-enabled HTTP-only deployment is
refused. Protocol 22 requires both the updated server binary and systemd unit, a
service-owned `/etc/flea/deploy.lock`, and writable managed content/runtime directories.
Fresh setup installs these. Existing servers need an administrator-reviewed upgrade;
do not use setup's destructive reset to upgrade a server with sites. Existing
content directories must be owned by `flea:flea` (0750); release directories
and host config remain root-owned. Each site's `flea/` runtime directory must be
owned by `flea:flea` (0700). Redeployment prepares content and runtime directories.
Back up your server before changing installation permissions.

## Email-only contact forms

Editable email templates live in `mailers/`, including `main.layout.html` and
HTML/plain-text versions of sign-in and contact messages. They are private,
watched during dev, and included in deployment. See [email templates](docs/email-templates.md).

Add `"contact": { "to": "admin@example.com" }` alongside `app` in
`.flea.config.json`, then put this in a page template:

```html
<form method="post" action="/_flea/contact">
  <label>Name <input name="name" required maxlength="100"></label>
  <label>Email <input type="email" name="email" required maxlength="254"></label>
  <label>Message <textarea name="message" required maxlength="5000"></textarea></label>
  <button type="submit">Send message</button>
</form>
```

With `dev.email_preview` enabled, submissions open locally without sending.
For real local delivery, disable the toggle, run `flea email`, then restart dev.
Local email settings load automatically from `.env`; shell variables take precedence.
Restart dev after changing contact settings too.
Deployment requires HTTPS and prompts for server mail settings if missing.
Admin access is not required. Omit `contact.to` to disable the endpoint.
Existing servers need the current runtime before deploying.

Messages are plain-text email to the configured recipient, with the visitor's
address as Reply-To. The browser receives a text confirmation, not a redirect.
Provider acceptance does not guarantee inbox delivery. No submissions are stored.
No uploads, inbox, automatic replies, custom fields or delivery queue.
Only `name`, `email`, `message`, and an optional empty `website` honeypot are accepted.
A filled honeypot returns a generic confirmation without sending.

Requests require matching Origin, production HTTPS, URL-encoded fields and a
32 KiB maximum body. Limits: 3 attempts/IP/site/15 minutes, 20/site/hour and
200/server/hour; one active send/site and four/server. Aliases share limits.
Limits reset on restart. Forwarded headers do not select the client IP.
This limits abuse but does not stop determined bots or provide process isolation.
Login mail has separate limits but shares the provider account quota.
Delivery times out after 10 seconds. An uncertain failure can mean a manual retry
sends a duplicate; there are no automatic retries. General-purpose forms remain deferred.

## Styling with Prewind

Start with `css/theme.css` to change colors, fonts, and spacing. Put your site's
shared styles in `css/global.css`, and use Prewind utilities for adjustments
in individual pages. There is no CSS build step.

The starter loads Prewind 1.3.1 from unpkg and uses a global-first stylesheet
structure adapted from Preflight.
The starter `theme.css` imports Inter for headings and body text from
Google Fonts, with `display=swap` and system sans-serif fallbacks. Warm off-white
pages, white panels and forest-green actions use editable theme tokens.
The entry point `css/global.css` declares
`@layer reset, styles, prewind` and imports:

- `theme.css`: colors, fonts, fluid spacing and text scales.
- `https://unpkg.com/prewindcss@1.3.1`: the utility library, including its reset.
- `typography.css`: headings, prose, links, lists, code, and tables.
- `forms.css`: native controls and optional `.form-field` / `.form-row` wrappers.
- `ui.css`: buttons, including `.ui-button.ui-brand-1`.
- `global.css`: imports and site-specific styles.

Keep shared rules in `@layer styles` and use utilities such as `flex gap-sm`,
`mt-lg`, or `text-brand-1` for one-off adjustments. Utilities override shared
styles through the layer order. Customize the theme rather than editing the
vendored library. Form styling does not add submission processing. CSS is
embedded from `templates/site/css/` into the executable when built.

The full starter lives in `templates/`, mirroring a generated site's structure:
edit `templates/site/*.html` for markup, `templates/site/css/` for styles,
`templates/site/images/` for assets, and `templates/flea/` for JSON defaults.
`templates/README.md` is the generated site's readme. Rebuild FleaCMS after
editing these files. They are embedded in the executable, so `flea new`
works from any directory without access to this repository. The embedding
includes private *.layout.html files, error pages, and root .flea.config.json.

## Lucide icons

Use literal lowercase Lucide names directly in templates, layouts, and partials:

```html
<a href="/">{icon arrow-left size="18"} Home</a>
{icon search class="text-dark" size="24" aria-label="Search"}
{icon check width="20" height="20" stroke-width="3"}
{#each snip.home.items as item}
  {icon info aria-label="{item.title}"}
{/each}
```

Before serving, `flea dev` downloads the icons you've used from the pinned
`lucide-static@1.47.0` package
on unpkg. It stores them directly as `flea/icons/search.svg`, for example.
New sites include `.keep` so the directory stays in Git. Each SVG contains the
license notice in a comment, with no version folder or separate license file.
New references are resolved during template reloads.
No browser-side icon library, JavaScript, or icon network requests are required:
the server renders cached geometry as inline SVG. Keep the cached SVG files
in version control for reproducible, offline use. Unused cached files are retained.
Deploy uses the same resolver before packaging; no standalone site build command exists.

`flea status` remains read-only and offline. Missing icons are warnings,
not validation failures: status exits successfully if the rest of the project
is valid. Missing icons and failed downloads render nothing in their place;
the rest of the page serves normally, with no browser error message or 500.
Dev prints one warning per unresolved icon at startup or reload, including the
template location and download failure when available. Warnings are not repeated
for every request. Malformed directives and unsafe/invalid SVGs remain errors.
A first use needs internet access; cached icons work offline. Failed downloads
do not install incomplete SVGs. Fix the name or connection and save the
template again (or restart dev) to retry. More than 256 new icons in one load
requires restarting dev to resolve the next batch. Only literal names are
supported; dynamic icon names, custom sources, and automatic version upgrades
are not supported. Icons referenced in unused partials or conditional branches
are resolved too. Script/style contents and snip strings are never scanned as icons.

Defaults: 24 by 24, `viewBox="0 0 24 24"`, no fill, `currentColor` stroke,
stroke width 2, rounded joins and caps, and `focusable="false"`.
Supported quoted attributes are `class`, `id`, `size`, `width`, `height`,
`fill`, `stroke`, `stroke-width`, `stroke-linecap`, `stroke-linejoin`, `color`,
`role`, `tabindex`, and `aria-*` / `data-*`. `size` sets both dimensions;
explicit width/height take precedence. Both single and double quotes work.
Attributes can contain escaped dotted-path values, including loop aliases.
Event handlers, inline `style`, links, and structural SVG attributes are rejected.
Use CSS classes for styling. Each icon has `lucide` and `lucide-<name>` classes.

Icons are decorative (`aria-hidden="true"`) unless given `aria-label` or
`aria-labelledby`, which defaults their role to `img`. Label an icon-only button
on the button itself. Unknown icon names are omitted with warnings; invalid name
syntax, duplicate/invalid attributes, and malformed SVGs are validation errors.
Downloaded and cached SVGs are reconstructed using
only basic geometry; scripts, external references, and arbitrary HTML cannot be
introduced through this feature. Browse names at [Lucide](https://lucide.dev/icons/).

## Routing

`index.html` maps to its directory: `/` at the root or `/about` for `about/index.html`. `about.html` also maps to `/about`; having both is a validation error. Page/asset route collisions are also errors. Assets keep their exact relative paths. There are no directory listings or implicit asset indexes.

URLs are case sensitive. Trailing slashes other than `/` return 404, without redirects. URLs containing `.html` (case insensitive), dot segments, doubled slashes, backslashes, NUL, or residual percent signs return 404. Standard percent encoding is decoded once by Go's HTTP parser and then validated; encoded private names are still private. Filenames that cannot produce an accepted URL fail validation.

Files and directories beginning with `_` or `.` are private, including configuration and snips. `/admin` is reserved and cannot be a page or asset. `/_flea/reload` is a development-only version endpoint, not a filesystem route. Symlinks and nonregular files in the included site tree are rejected. Reads use confined directory handles with file identity checks. Requests serve only preloaded route maps and immutable asset bytes, never request-derived filesystem paths.

## Reload behavior

Dev logs requests to the terminal without timestamps. Successful requests use
`GET /test`; non-200 responses include the status, such as
`404 GET /no-such-page` or `500 GET /test`. Assets and HEAD requests are logged
too. Successful internal reload polls are omitted. Logs include the escaped
route path, not query strings, so query tokens are not printed and encoded
control characters cannot create extra log lines.

The dev server watches the project directory and included subdirectories using native filesystem notifications. Changes are combined with a 100 ms debounce, then trigger full validation and a new in-memory snapshot. New directories, deletions, renames, and editors' atomic saves are detected. File fingerprints are checked only at startup or after events, not periodically while idle. Excluded metadata and dependencies are not watched. Watch failures stop dev with an error rather than silently switching to polling. Network filesystems that do not provide native notifications are not supported for automatic reload.

Served public HTML still includes a small script that checks the version every 700 ms and refreshes the browser after changes. This HTTP check does not scan files. No WebSocket, bundler, or browser extension is required. With JavaScript disabled, refresh manually. Admin pages require manual refresh to protect unsaved edits. Full reloads do not preserve browser form state.

Invalid edits produce a self-contained development diagnostic with the error,
template filename and up to 11 correctly numbered source lines (five on either
side of the highlighted error) when available, and a suggested
fix. JSON syntax errors show the filename and line, but not configuration values.
Missing routes explain URL conventions and list up to twelve known pages.
Unsupported methods explain where forms should submit. The page works even when
the site's layout or styles are broken and reloads after the next saved change.
Initial startup failures still print in the terminal before a server is running.

Detailed pages require both a loopback client and a localhost/loopback Host;
production never serves them. Errors also appear in the dev terminal or production
server logs, including failures rendering a custom error page. Production keeps
the safe custom/generic response. Logs omit request queries, cookies and bodies.

Fixing the source publishes a working snapshot. Requests already using an older
snapshot can finish safely. Assets are held in memory, so disk edits cannot change
a response halfway through serving it. A multi-file edit is not a transaction;
save related changes together and allow validation to recover. Full snapshots
are intended for small local sites, not large media libraries.

## Deploy a site

### Snips and theme synchronization

Commit `flea/snips.json` and `css/theme.css` with your site. They seed the first
deployment. Subsequent plain `flea deploy` preserves existing live snips, page
metadata, and theme CSS while publishing templates, assets, and configuration.
New snip IDs receive their committed seed content; removed IDs are archived.
Keep IDs stable when changing templates. Incompatible rich content in a new
heading/paragraph wrapper fails validation instead of being stripped.

Before local content/design work, stop `flea dev`, then run:

```sh
flea pull
# Or pull just one kind of editable state:
flea pull --snips-only
flea pull --theme-only
```

Pull uses the current branch's saved destination. It preserves local-only snips,
does not copy config, credentials, sessions, or templates, and does not commit.
Commit pending edits to the selected files first. Review and commit the pulled
changes afterward. Dev and pull cannot run simultaneously on the same project.

To intentionally publish local edits, commit them and use:

```sh
flea deploy --with-snips
flea deploy --with-theme
flea deploy --with-snips --with-theme
```

An existing deployment requires a pull baseline for each selected scope. Baselines
live privately in Git metadata under `flea-sync`, keyed by destination, branch,
and local project. They are not committed or shipped. A fresh checkout must pull
before publishing existing content. First deployments seed both without a baseline.

Snips merge by ID, with titles and descriptions compared separately. Theme CSS
is compared as one file so imports and developer customizations remain intact.
If both copies changed differently, the CLI shows the conflict and asks whether
to keep live, use local, or cancel. The first pull also asks when differing copies
have no shared baseline. There is no default winner; noninteractive conflicts fail.
All choices are collected before publishing. Live state is rechecked under the
deployment lock before activation; concurrent admin edits stop a stale plan.
Pull again after publishing to refresh your baseline and local admin content.

Pull validates before writing and advances its baseline last. If a disk write
fails, it reports any partial local update; review the files before retrying.
Neither command force-overwrites conflicts. Remote rollback remains planned.

You'll need Git, SSH access to a supported Ubuntu server, and a domain pointing
at that server (or an IP address for an HTTP-only site). For HTTPS, allow inbound
TCP ports 80 and 443. Use a complete release or build with `bash scripts/build.sh`
to include the bundled servers needed by setup.

From your site's folder, on the branch you want to publish, run:

```sh
flea deploy
```

On first use for each branch, FleaCMS asks for the target server, SSH username
(default `root`), domain/IP,
WWW-to-primary redirect, and HTTPS (default yes). Deployments use the actual
branch name, with no separate environment label or question. The internal site
name is generated automatically from the project folder and branch. A new
`staging` branch suggests `staging.example.com` using the saved main mapping,
but requires approval. Staging defaults to no WWW alias; main defaults
to redirecting WWW. An IP destination uses HTTP without a WWW alias.
HTTPS uses one combined question: `Enable HTTPS and accept Let's Encrypt terms?`
(default yes), with the terms link displayed first. It then asks for a contact
email. The saved choice is reused on future deploys.
Supplying `user@host` skips the username question. Bare targets saved by older
versions ask for the username once and save it; commit that config change before
retrying. Deployment never silently chooses your Mac username.

After connecting, deploy checks whether FleaCMS is installed. If it is missing,
it offers to run setup on that same server, then continues deploying. The existing
SSH connection is reused, including the same verified bundled-server extraction.
Failed SSH authentication does not trigger setup,
and an older or incomplete installation requires an update rather than reinstall.

Selections are saved under `deploy.branches` in `.flea.config.json` before
attempting deployment, preserving app/admin/contact and other fields:

```json
{
  "app": { "site_name": "My Site" },
  "deploy": {
    "branches": {
    "main": {
      "target": "ubuntu@1.2.3.4",
      "domain": "example.com",
      "name": "my-site-main",
      "www_redirect": true,
      "https": true,
      "email": "you@example.com",
      "agree_tos": true
    },
    "staging": {
      "target": "ubuntu@1.2.3.4",
      "domain": "staging.example.com",
      "name": "my-site-staging",
      "www_redirect": false,
      "https": true,
      "email": "you@example.com",
      "agree_tos": true
    }
    }
  }
}
```

FleaCMS offers to commit pending changes, including the new configuration, then
continues deploying. It lists every changed file, asks for confirmation and a
commit message (`init` for the first commit; otherwise required with no default), and uses normal Git
hooks and signing settings. Declining or a failed commit stops deployment.
All listed repository changes are staged, not just public site files. Sites
inside a larger repository require a manual commit to avoid staging unrelated
work. Changes made later during deployment still stop it. Future deploys
use only the current branch's saved destination and print the branch/commit.
An unmapped branch prompts instead of falling back to main. Detached HEAD,
repositories without a commit, and non-Git folders cannot deploy. The wizard can
save settings before the first commit, so you can commit the site and deployment
settings together. Upload still requires that commit and a clean status. Settings are
never uploaded and contain no credentials. You can explicitly ignore the file
if you prefer local-only settings. Otherwise commit and share the branch map
across the branches that should know about it.

Use `flea deploy -configure` to change a saved mapping. Flags such as
`-target`, `-domain`, `-name`, `-https`, `-www-redirect`, `-email`, `-agree-tos`,
and `-expected-ips` seed the questions on first use or with `-configure`.
They cannot silently override an existing branch destination. Flat settings from
the previous version are offered for confirmation and migrated to a branch map.
Different branches cannot share a hostname or the same remote site name on a
target. The server also records and enforces site branch ownership.

Git must be clean, including staged, unstaged, and untracked files, before SSH.
The repository is checked again before upload and activation. The snapshot is
read from the selected commit's Git objects, never working files, so ignored
assets and uncommitted edits cannot slip into a release. Missing icons are
resolved into the temporary committed snapshot without changing the checkout.
Sites in a repository subdirectory are supported; cleanliness covers the whole
repository. Git and a valid committed site are required on the deploying Mac.

Before uploading, the target server resolves every configured hostname (including
WWW when enabled). Lookup failures, no A/AAAA records, or any address that does
not match the target fail deployment. Both IPv4 and IPv6 answers must match.
By default the expected addresses come from server interfaces. For NAT or a
proxy/CDN, set `expected_ips` to the explicitly approved public addresses in
the config or with `-configure -expected-ips`. This advanced setting is not asked
about in the normal wizard. IP destinations skip DNS lookup.
This checks the server's resolver; it does not prove public DNS propagation or
that firewall ports are open. DNS names must point at the server for ACME.

Deploy resolves referenced icons in a temporary snapshot, validates it, and
uploads it over native SSH/SCP. Only files under `site/`, private `mailers/`,
the runtime portion of `.flea.config.json`, managed `flea/media/`,
`flea/snips.json`, `flea/data/`, legacy `flea/snips/`, and
`flea/icons/` are included. Repository files and other private workspace files are
excluded. Archives are limited to 256 MiB uncompressed and 10,000 entries;
symlinks, hard links, special files, duplicates, and traversal paths are rejected.

On the server, each release gets a new directory:

```text
/var/lib/flea/sites/my-site/releases/<timestamp-random-id>/
```

The installed server validates the extracted site, atomically replaces its JSON
host configuration, restarts the service, and verifies the active revision and
hostname. All existing aliases for the same site move together; unrelated sites
remain configured. A hostname already assigned to another site is refused.
Omitting a previous alias does not delete it. Changing aliases/removing sites is
currently a manual host configuration task.

There is a brief interruption while the single service restarts. If activation
or certificate verification fails, FleaCMS restores the previous configuration
and restarts it. Concurrent operations are locked. Prior releases are retained,
and the previous successful configuration is saved as
`/etc/flea/server.previous.json`. Failed staged releases are retained too;
automatic pruning and a user-facing rollback command remain future work.

An interrupted SSH session can leave an unfinished transaction. Inspect the
service and `/etc/flea/deploy.pending.json` on the server. If the journal
exists, restore the previous configuration with:

```sh
sudo /usr/local/bin/flea-server -recover
```

Further deployments refuse to proceed until recovery succeeds. Do not delete
the recovery journal to bypass that check.

Normal snip deployment preserves remote content by ID and existing page metadata.
Local content seeds only missing IDs. Code and structured data are still replaced.
Content lives outside code releases at
`/var/lib/flea/sites/<site>/content/<release-id>.json`; the server config's
`content` map selects the active snapshot for each hostname. Aliases pointing to
the same release must select the same content snapshot; conflicting paths are
rejected. Valid aliases share one loaded site snapshot. The existing lock
and recovery journal switch code and content together. Failed activation leaves
the previous content snapshot untouched.

After successful activation, unreferenced IDs are removed from active page/shared
snip maps and retained in the content snapshot's `archive` map. References are
scanned from template source, including false branches, empty loops, and unused
partials. Reintroducing an archived ID restores its last remote content, not a
stale local seed. Archives and old snapshots are not automatically deleted.
Page metadata is retained even when a page disappears. There is no rollback command
yet. Browser saves share the deployment lock and update the live snapshot. Manual
content edits require a service restart and must not run concurrently with deployment.

## Enable HTTPS

Deployment settings are saved without a final save-confirmation prompt. Deployment
shows the branch and domain with a Ctrl-C cancellation notice, and prints the
site URL after successful activation. A `flea rollback` command is not yet
available.

After deploying a domain, run:

```sh
flea https
```

The first-deploy wizard enables HTTPS by default, so you normally won't need a
separate step. To change an existing HTTP-only branch mapping and deploy it, run:

```sh
flea deploy -configure -https
```

Confirm the settings and use the offered commit step, or commit them yourself
before retrying. `-configure` continues into deployment after saving settings.

`-agree-tos` records acceptance of the [Let's Encrypt subscriber terms](https://letsencrypt.org/repository/).
The domain must resolve publicly to the server, inbound TCP 80 and 443 must be
open, and the server needs outbound HTTPS to Let's Encrypt. HTTP challenges use
port 80; TLS-ALPN challenges use 443. No DNS-provider credentials are needed.
IP certificates and wildcard certificates are not supported. Contact email and
consent are required explicitly. They are saved within the selected branch's
settings. WWW redirect adds both hostnames to the certificate allowlist, and
redirects to the primary domain over HTTP or HTTPS while preserving path/query.

FleaCMS uses pinned `golang.org/x/crypto/acme/autocert` for certificate issuance,
caching, and automatic renewal. The daemon handles both ports without a proxy.
Only explicitly enabled site domains can obtain or use certificates. TLS uses
version 1.2 or newer and checks that the request Host matches its TLS hostname.
HTTP requests for enabled domains redirect to HTTPS with status 308, preserving
the path and query. Other sites stay on HTTP. Challenges and the generic health
endpoint remain available on HTTP. HSTS is not enabled in this first version.

Certificates and ACME account keys live under `/var/lib/flea/certs` with
restrictive permissions. Preserve this directory across server updates and back
it up securely. A restart reuses cached certificates. The service account can
write certificates but cannot modify deployed releases or host configuration.

Activation waits up to two minutes for readiness and certificate verification,
using a locally dialed HTTPS connection with normal certificate/hostname checks.
A failure restores the previous configuration. Diagnose DNS, firewall, CAA, and
ACME errors with `sudo journalctl -u flea -n 100 --no-pager` before retrying;
repeated issuance attempts can hit CA rate limits. Previously enabled HTTPS is
controlled by the saved branch preference on later deploys. Explicitly changing
that preference to false and redeploying disables TLS for that site's aliases;
cached certificates are retained. Other sites' HTTPS settings are preserved.

The resulting host config includes a `tls` object alongside `listen`, `sites`,
and the generated `revision`:

```json
{
  "listen": ":443",
  "email": "you@example.com",
  "agree_tos": true,
  "cache_dir": "/var/lib/flea/certs",
  "domains": ["example.com"]
}
```

This is the value of `tls`, not a replacement for the whole server config. The
outer HTTP `listen` stays `:80`. Test public access from your Mac:

```sh
curl -I http://example.com/
curl -I https://example.com/
```

Tests simulate SSH/SCP failures, exercise real local HTTP/TLS listeners and
deployment recovery, and issue a certificate through a local ACME test authority.
They also verify persistent certificate reuse without the CA. No real Ubuntu
server or live Let's Encrypt account was provisioned during implementation.

## Set up an Ubuntu server from your Mac

From any directory, run `flea setup`:

```sh
flea setup
```

Enter your server's hostname, IP, or SSH config alias. Standalone setup always
connects as root, even if you supply another username. Custom ports and keys
belong in `~/.ssh/config`.
FleaCMS uses the installed `ssh` and `scp`: keys, agents, passphrases, normal
host-key checks, and interactive password authentication are handled by SSH.
If you use a password, password login must be enabled on the server. Passwords are never stored by
FleaCMS, and host-key verification is not disabled. Non-root installation uses
`sudo` with a remote terminal for its password prompt. A temporary SSH control
connection reuses authentication and is closed when setup finishes.

Targets: Ubuntu LTS 22.04, 24.04, or 26.04 with systemd, on amd64 or arm64.
The target needs OpenSSH/SFTP and standard Ubuntu tools including `ss` (iproute2),
plus sudo for non-root deployment targets. No compiler is needed on either machine.
Setup checks SSH and the remote OS/architecture, extracts the matching compressed
Linux server bundled inside the CLI, verifies its bundled SHA-256 and static ELF
architecture, then uploads it. Extraction is limited to 100 MiB. No server download,
compiler or registry is needed during setup. Verification happens before upload
or reset confirmation. There is no download or compiler fallback.

`flea version` prints the CLI version; `flea-server -version` prints the
server version. The current version is recorded in `internal/buildinfo/version.go`;
see [CHANGELOG.md](CHANGELOG.md) for release notes. `bash scripts/build.sh`
produces a build of that version with both
servers from the same source. Plain `go build` does not refresh embedded servers
and is not the supported installation build. See [RELEASING.md](RELEASING.md).

Setup installs `/usr/local/bin/flea-server`, `/etc/flea/server.json`, and
`/etc/systemd/system/flea.service`. It creates a dedicated non-login `flea`
account and `/var/lib/flea`. The service starts at boot and restarts on failure.
Setup verifies systemd activity and HTTP readiness before reporting success.
Before changing the server, the installer checks TCP and UDP ports 80 and 443
on IPv4 and IPv6. It refuses occupied ports unless every listener is verified as
the existing FleaCMS systemd main process during a confirmed reinstall. Unknown
process owners or failed port checks also stop installation; other services are never stopped.

Setup connects as root; enter just the hostname or IP. When the server binary
and configuration already exist, it warns that all sites will go offline and
requires typing `DESTROY` to reset the installation. The replacement binary is
extracted and verified before this confirmation; extraction failures leave the
existing installation unchanged. Old files are moved into a
root-only `/var/backups/flea.*` directory, including certificates and releases.

Setup and deploy automatically trust and save previously unseen SSH host keys.
Changed or revoked keys stop the connection with a friendly explanation.
Verify a changed fingerprint with your server provider before updating your
known_hosts entry; FleaCMS never removes saved keys automatically.
Restoration is manual; this is a reset, not an upgrade. Unexpected partial
installations and symlinked installation paths are refused. It does not modify firewalls,
SSH settings, packages, or existing sites. If installation fails, its newly
installed service/binary/config files are removed; any created account and data
directory remain for inspection. After a connection loss, inspect remote state
before retrying, since interrupted installs may leave partial state.

The host listens publicly on HTTP port 80 and has no sites configured:

```json
{
  "listen": ":80",
  "sites": {}
}
```

The `sites` object maps hostnames to absolute site directories, for example
`"example.com": "/var/lib/flea/sites/example"`. The service account must be
able to read them. Sites/configuration load at startup; restart after changes.
The host serves validated snapshots without browser reload scripts, file watching,
or icon downloads. Resolve icons locally before copying a site. Unknown hosts
return 404, including requests to the server IP unless that IP is a configured
site key. `/_flea/health` publicly reports only the service name and readiness.
The empty listen address binds all available interfaces; explicit IPv4/IPv6
addresses and numeric ports 1-65535 are also accepted. The service remains
non-root, with only `CAP_NET_BIND_SERVICE` granted for binding low ports.
Local `flea dev` remains loopback-only.

Allow inbound TCP 80 in your server and cloud firewalls. Setup does not change
either firewall, and its local readiness check does not prove public reachability.
From your Mac, test public access (replace the example IP):

```sh
curl http://1.2.3.4/_flea/health
```

On the server:

```sh
sudo systemctl status flea --no-pager
sudo journalctl -u flea -f
sudo systemctl restart flea
```

Existing installations are not upgraded by rerunning setup. Deployment requires
the current server executable (deployment protocol 22, unquoted snip syntax).
For an older installation,
update the executable and systemd unit deliberately, preserve server.json and all
sites, and create `/var/lib/flea/certs` owned by `flea:flea` with mode
0700. The data directory itself should be root-owned, mode 0755. The current
unit permits writes within certificates, managed site state and the deployment
lock; filesystem ownership still protects releases and site parents. Preserve
content directories (flea-owned 0750) and create each managed site's private
`flea/` directory (flea-owned 0700) before starting the new server. Reload systemd and restart
after replacing the unit. For executable-only updates on a current installation,
use `flea update-server`; it does not migrate service units or configuration.

## Local installation details

For a persistent local installation, from the repository:

```sh
mkdir -p "$HOME/.local/bin"
bash scripts/build.sh
cp bin/flea "$HOME/.local/bin/flea"
export PATH="$HOME/.local/bin:$PATH"
```

Add that PATH setting to your shell profile to keep it available in new terminals.

For framework development in this checkout, create the global link once (this
does not overwrite an existing command):

```sh
sudo ln -s /Users/sean/Projects/flea/bin/flea /usr/local/bin/flea
flea version
```

Replace your old build alias in `~/.zshrc` with:

```sh
alias flea-build='bash /Users/sean/Projects/flea/scripts/build.sh'
```

Run `source ~/.zshrc`, then `flea-build` from any directory. The script finds
Go on PATH or at `/usr/local/go/bin/go`; `GO_BINARY` overrides that location.
Contributors need Go, Bash, gzip and shasum; installed users do not need them.

`new` accepts a simple name containing letters, digits, underscores, and hyphens, starting with a letter or digit. It refuses existing destinations, including empty directories and symlinks. An interrupted creation may leave a partial new directory for inspection. `status` validates configuration, routes, templates, and current content without listening on a port. Exit codes are 0 for success, 1 for project or runtime failures, and 2 for invalid CLI usage.

`flea new my-site` also initializes Git on `main`, without staging or committing
anything. The starter includes a `.gitignore` for macOS metadata, temporary editor
files, and local environment secrets. To use a parent repository instead, run
`flea new my-site --skip-git` (the flag also works before the name).
If Git is not installed, the site is still created with instructions for setting
up Git before deployment.

## Checks

From the repository root:

```sh
gofmt -w cmd internal
go test ./...
go test -race ./...
go vet ./...
bash scripts/build.sh
go test ./internal/cli -run TestExecutableWorkflow -count=1 -v
```

The executable smoke test builds the binary, creates a fresh starter, runs status, starts dev on an available loopback port, checks pages/assets/private URLs, edits snips/config/assets/templates, checks invalid-edit errors and recovery, and sends an interrupt to verify clean shutdown. Tests need permission to bind a local TCP listener.

## Current limitations

Developers designate editable images/videos with a bare `snip` attribute. Dev
imports local seeds and writes typed JSON entries, generating src/alt/poster
bindings while preserving layout attributes. Admins click the media in Edit mode
and use the page's Save/Cancel. There are no Photos/Videos topbar modes.
Nested media/text snips and competing srcset/picture/source markup are errors.
Video requires H.264 plus optional AAC-LC, fast start,
25 MiB/file; no conversion or external service. Metadata is committed, video bytes
are Git-ignored and transferred separately during deployment/pull. See
[videos](docs/videos.md) for storage, validation, fresh-clone recovery and limits.

For hosting Flea's own release binaries independently of website deployments,
see [disk-backed download hosting](docs/download-hosting.md). This optional host
setting does not change ordinary site assets or the 256 MiB site archive limit.

New sites include AGENTS.md and portable skills in `flea/skills/`
(`import-brochure` and `fleacms-site`). These are developer instructions, not a
runtime dependency or an automatically installed agent integration. They are
excluded from serving and deployment.
For substantial imports and redesigns, they guide an agent through two design
directions and a sample-page review before the full build, unless you explicitly
delegate those choices. Small maintenance edits do not require a new design review.

Editable regions, short IDs, sanitized HTML, page metadata, content-preserving
deployment, and archived orphan pruning are implemented. Deployment protocol 22
requires an updated server binary; older servers are refused before upload.
Local extraction preserves unreferenced content until deployment. There is no
Markdown renderer or schema migration engine. Inline rich-text editing and
explicit `flea pull` synchronization are supported. Keep backups before manual edits.

This release provides local development, SSH setup/deployment, automatic HTTPS,
email-code authentication, inline editing, multisnips, theme editing and a JPEG/PNG
image/video snips. Rollback, site backups/restoration and executable-only server upgrades
are supported; see [recovery](docs/recovery.md). Release pruning, remote logs/status
commands, general-purpose forms, general field schemas, per-edit revision history
and adaptive video streaming remain deferred. Deployment briefly restarts the service
and preserves remote admin content. Developer-managed images may remain in `images/`.
Unrestricted raw HTML is unsupported. Real Ubuntu installation, live Resend
delivery and live CA verification remain pending; automated tests use fake email
delivery and local HTTP/TLS. This is not a production security certification.
