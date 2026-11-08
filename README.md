# FleaCMS

<img src="site/images/logo.svg" alt="FleaCMS" width="249" height="60">

## Temporary packaging storage

FleaCMS 0.20.1 respects `TMPDIR` for local deployment and setup. On Linux with a
small RAM-backed `/tmp`, use a private directory on verified disk-backed storage
for each job, and keep worktrees and retained review evidence outside scratch.
The public documentation includes the supported per-deployment command.

## Website sharing preview

The public layout includes Open Graph and large-image card metadata with each
page's title, description and canonical URL. The shared 1200 by 630 PNG lives at
`site/images/social-preview.png`; its editable artwork is
`design/social-preview.html`. Render that HTML in Chrome at 1200 by 630 pixels,
device scale 1, to regenerate the image. It uses the existing logo, shared theme
tokens and terminal workflow from the homepage, with no frontend build dependency.
Keep the image URL absolute.

Build pages with HTML, style them with CSS, and let site admins edit their own content. FleaCMS builds the server, and serves multiple sites with HTTPS directly.

## Migration redirects

FleaCMS 0.20.0 adds 301 redirects for old site URLs. Merge a top-level object
into `.flea.config.json`, preserving other settings:

```json
"redirects": {
  "/about-us.html": "/about",
  "/services/": "/services"
}
```

GET/HEAD match exact paths, including listed extensions and trailing slashes.
Destinations must exist as public pages/assets on this site. Incoming queries
survive unless the target supplies one; a final `?` clears them. Targets may have
fragments. Validation rejects unsafe/private paths, duplicate decoded sources,
live route/app conflicts, missing destinations, chains and loops. No wildcard,
query-specific or external-URL mappings exist.

Use the `fleacms-site` skill to inventory an old website from sitemaps and links,
map replacements and verify responses. Provide old URL/analytics exports for
orphaned pages; leave ambiguous mappings for a decision. Run `flea status` and
verify 301/Location and the final content locally. Update both CLI and server to
protocol 28 before deploying; public 0.19.0 and earlier ignore these rules. See
[the public guide](https://fleacms.com/docs#redirects).

## Search discovery

FleaCMS 0.20.0 can generate a sitemap, advertise it in robots.txt and add missing
canonical tags. A canonical tag identifies a page's preferred address. Update
the local CLI with `flea update`; for an older shared server, run
`flea update-server` from a site with that server saved as the current branch's
deployment destination. The server update briefly restarts every site on it.
Deploying website files does not update the server executable.

Add this object alongside `app` in the site's `.flea.config.json`, preserving
the other settings:

```json
"seo": {
  "url": "https://example.com",
  "exclude": ["/campaign/receipt"]
}
```

Use the production HTTPS address even in dev or staging. `exclude` is optional;
replace the example with a real page you want kept out of search. Entries match
exact paths, without wildcards, queries or trailing slashes. Excluded pages remain
accessible with a noindex header and are omitted from the generated sitemap and
canonical tags. Contact receipts, drafts, private routes, custom managed apps and
redirects stay out of automatic discovery.

Run `flea status`, then `flea dev`. At the printed local address, check
`/sitemap.xml`, `/robots.txt`, a public page's canonical tag in View Source and
an excluded page's `X-Robots-Tag: noindex` response header. Commit the reviewed
configuration, deploy from the intended branch with `flea deploy`, and repeat
those checks on the public domain.

Custom robots.txt/sitemap.xml files and existing canonical tags take precedence.
Maintain sitemap declarations and exclusions yourself in custom discovery files.
Without `seo`, automatic discovery is disabled. Staging protection still uses
the separate exact-host indexing policy.

Google Search Console is optional: generated robots.txt already advertises the
sitemap. To see Google's processing reports, verify ownership of the production
site and add `https://example.com/sitemap.xml` in its Sitemaps report. Submit the
URL, not an uploaded file; it stays current without resubmitting after every edit.
See [the public walkthrough](https://fleacms.com/docs#seo-discovery).

## Hostname indexing

FleaCMS 0.19.0 adds exact-host indexing policy with deployment protocol 27.
Update the CLI and explicitly upgrade the shared server before using it:

```sh
flea update
flea update-server
flea server indexing root@server staging.example.com noindex
flea server list root@server
```

The hostname must already be registered. Type it to confirm; changing policy
restarts the shared service. Staging receives `X-Robots-Tag: noindex, nofollow`
without a website app flag or different templates. Policies survive deploy, sync,
code rollback and site restore. `allow` stops adding the restriction and preserves
owner-authored restrictions. Keep robots.txt crawlable so crawlers can discover
noindex; Flea warns about detected static site-wide blocks without rewriting it.
Site backups exclude policy; removal clears only owned hosts. Verify native
headers before removing an existing meta fallback. See the
[public documentation](https://fleacms.com/docs#indexing) and RELEASES.md for
actual rollout evidence.

## Getting started

For a new project, install FleaCMS first, then run:

```sh
flea new foobar
cd foobar
flea dev
```

Install the public test release with `curl -fsSL https://fleacms.com/install.sh | bash`.
Downloads cover Apple Silicon macOS and x86-64 Linux. Linux ARM64 CLI downloads
are discontinued from 0.18.0; both supported CLIs retain ARM64 Linux server
deployment support. Older published downloads remain available. See
https://fleacms.com/docs#installation for PATH setup and current testing limits.
This generated folder already is a site, so you can skip directly to `flea dev`.
Use `flea new .` to create a site in the current empty directory, or pass a
relative or absolute directory path. Flea creates a missing directory when its
parent exists. Existing directories must be completely empty, including hidden
files and `.git`; file targets and directory symlinks are refused. The folder
name supplies the default site name, which first-run dev lets you change.

`flea new` creates a clean layout and editable homepage. Use `flea new --starter
foobar` (or `-s`) for example pages, a six-person fictional team and a contact form.
Navigation uses a native HTML popover on mobile, with no menu JavaScript.
The clean site has Home navigation; the starter has Team and Contact.

`flea reset` returns this site to the clean default after confirmation. It replaces
all local files/settings except `.git`, including uploads, `.env`, customizations
and deployment destinations. Stop dev first. Original files remain in the private
sibling backup path printed by the command. No commit or remote change is made.

### 1. Start the development server

With the `flea` command installed, run this from your site folder:

```sh
flea dev
```

On the first run, enter your site name and an optional admin email. New sites
enable `dev.email_preview` in `.flea.config.json`: sign-in and contact emails
open in your browser without sending or needing Resend credentials. The preview
looks like an opened email, with sender, recipient, subject and date/time.
Use its sign-in code normally. Previews expire after 10 minutes and stay only in
dev's memory (the latest 20); stopping dev discards them. If the browser cannot
open, use the local preview link printed in the terminal.

For real local delivery, set `"dev": {"email_preview": false}`, run `flea email`
and restart dev. Production ignores preview mode and always needs real email
configuration. Restart dev after changing preview mode or its sender details.

Open [http://localhost:3000](http://localhost:3000). Leave the server running while
you work; saved changes reload the browser. Running dev again gracefully replaces
the same site's existing dev server, including LAN mode and port changes.
Pull/reset remain protected. Stop dev processes started with older binaries once.

Use `flea dev --lan` (or `-l`) to test the full site from another device on a
trusted private IPv4 network. Open a printed Network URL. Email previews open
on the computer running Flea; use its code on the other device. LAN mode uses
unencrypted HTTP. Never forward the development port; restart after changing networks.

### 2. Make your first change

Open `site/index.html` to change the homepage structure. Edit `flea/snips.json`
for existing welcome text and page metadata, or add a new editable heading:

```html
<h1 snip>Welcome to our site</h1>
```

Dev moves the text into `flea/snips.json` and replaces it with a generated snip
directive. See [editable content](#editable-content) for details.

### 3. Check and save your work

```sh
flea status
```

This validates the site without rewriting files. Review your changes, then commit
the templates and extracted content together.

New sites have a Git repository on `main`, unless created with `--skip-git` or
without Git installed. Flea creates an `Initial site` commit using your normal
Git identity, hooks and signing settings, then prints commands to connect and
push a remote repository. A failed commit preserves the site with retry instructions.

## Project structure

| Path | Purpose |
| --- | --- |
| `site/index.html` | Homepage |
| `site/main.layout.html` | Public layout, including the header and footer |
| `site/admin.layout.html` | Sign-in layout |
| `site/css/` | Stylesheets and theme tokens |
| `site/images/` | Developer-managed images and logo |
| `.flea.config.json` | Site settings and branch deployment destinations |
| `flea/snips.json` | Editable content, page metadata, and media records |
| `flea/data/` | Optional developer-managed JSON data |
| `flea/media/` | Managed photos and videos |
| `flea/icons/` | Downloaded icon cache |
| `flea/runtime/` | Private, Git-ignored sessions and recovery files |
| `mailers/` | Private email templates |
| `flea/skills/` | Portable instructions for coding agents |

Public URLs omit `site/`. Configuration, mailers, runtime files, and agent skills
are never served as public assets. Keep secrets out of configuration and Git.

## Features

### Pages and routing

Create `site/about.html` to add `/about`. You can also use
`site/about/index.html`, but not both for the same route.

Use `/about` in links. Explicit `.html` URLs and trailing slashes return 404 unless explicitly listed in `redirects`.
Assets keep their paths: `site/images/logo.svg` is `/images/logo.svg`.
Files and directories beginning with `_` or `.` are private.

### Layouts

`site/main.layout.html` wraps public pages. Its `{slot}` receives the page content;
the starter keeps the header and footer directly in this file.

To select a custom layout, put a settings comment before the page markup:

```html
<!-- flea
layout: /custom.layout.html
-->
<main class="container">Your page content</main>
```

Create the matching `site/custom.layout.html` with a `{slot}`. Layout files are
private. `site/_404.html` uses the default layout; `site/_500.html` is standalone.

### Partials

Put reusable markup in an underscore-prefixed file, such as `site/_card.html`:

```html
{render card}
```

Relative paths resolve from the calling template. Use `{render /shared/card}` for
`site/shared/_card.html`. Omit the underscore and `.html` in render directives.

### Site settings

Store public site-wide values in the `app` object in `.flea.config.json`:

```json
{
  "app": {
    "site_name": "Our studio",
    "phone": "555-0100"
  }
}
```

Merge these values into your existing configuration, preserving other settings.
Use `{app.site_name}` or `{app.phone}` in templates. Dev saves the initial site
name here; commit it before deploying.

### Editable content

Add a bare `snip` attribute to an element whose content admins should edit:

```html
<section snip>
  <h2>About us</h2>
  <p>We build things that last.</p>
</section>
```

While dev is running, Flea extracts the inner HTML into `flea/snips.json`, removes
the attribute, and leaves a generated directive:

```html
<section>{snip a7Km2pQ9xB}</section>
```

Keep generated IDs unchanged and unquoted. Existing stored content is preserved.
Removed snips are archived; restoring their IDs restores their content.

Do not nest snip regions or put template directives inside new regions. Rich text
is sanitized; use the photo and video features for media. Commit the rewritten
templates and content together before deployment.

### Repeated content with multisnips

Use a multisnip for admin-managed lists such as team members:

```html
<article multisnip="team">
  <h2 snip="name">Joe Appleseed</h2>
  <p snip="title">Founder</p>
</article>
```

Dev extracts seeds and creates a loop. Editing starts on after login and remembers
your choice across pages in the same tab. Dashed outlines identify editable content.
Collection cards have inset drag handles, remove X icons and an Add tile. Use mouse/touch dragging
or arrow keys on a handle. Text and photo selection stay on-page; Save changes
publishes everything and Cancel restores it. Both keep Editing on.

Keep groups on individual pages, with one root element and named inner fields.
Do not nest groups or place them in layouts, partials, or other loops.
`site/team.html` includes a working example with a photo field.

### Structured data and conditions

Put developer-managed data in `flea/data/`. For example, `flea/data/team.json`
containing a `members` array is available as `data.team.members`:

```html
{#each data.team.members as member}
  <p>{member.name}</p>
{:else}
  <p>No members yet.</p>
{/each}

{#if app.phone}
  <p>Call {app.phone}</p>
{/if}
```

Values are escaped. Conditions use dotted paths and also support `{:else}` and
`{:else if app.email}`. Use multisnips when admins need to manage the list.

For negative conditions, use `{#unless app.email}...{/unless}`. It uses the
inverse of if truthiness and supports `{:else}` and `{:else if dotted.path}`.

### Styling

The starter uses Prewind 1.3.1 from unpkg, without Node or a CSS build step.

- Edit `site/css/theme.css` for colors, fonts, and fluid sizing.
- Edit `site/css/global.css` for site-wide rules in `@layer styles`.
- Use Prewind utilities for individual layout and spacing adjustments.
- Use `container` for centered page content.

`global.css` imports the theme, typography, forms, button styles, and navigation
transitions. The default font is Inter from Google Fonts, with warm off-white
backgrounds and forest-green accents controlled by theme tokens. Replace
`site/images/mark.svg` with your own logo.

The documentation loads a local Prism.js 1.30.0 bundle from `site/js/prism.js`
with core, markup, Bash and JSON components from the upstream `v1.30.0` tag.
`site/js/prism-go.js` adds the C-like and Go components from the same tag.
Its MIT license is kept in `site/js/prism.LICENSE.txt`. Label documentation code
blocks with `language-markup`, `language-bash`, `language-json` or `language-go` on the `code`
element. Token colors live in `site/css/docs-code.css`; inline code keeps the
normal site styling. No frontend build step is needed.

### Icons

Reference Lucide icons directly in HTML:

```html
{icon arrow-right}
{icon search size="20" aria-label="Search"}
```

Dev downloads only referenced icons as `flea/icons/name.svg`. The directory
includes `.keep`; each SVG includes its license notice as a comment. Commit
these files for offline use. Deploy also resolves missing icons;
`flea status` never downloads them.

Quote attribute values. Classes, dimensions, stroke attributes, and `aria-` and
`data-` attributes are supported; event handlers and inline styles are not.
Unavailable icons are omitted with terminal warnings. Invalid syntax or unsafe
SVGs fail validation.

### Admin sign-in and editing

Add an approved admin from your site folder:

```sh
flea admin admin@example.com
```

The command adds the address to `admin.emails`. It starts local email setup only
when preview mode is off and delivery is unconfigured. Run `flea email`
separately to configure Resend for real delivery. Local secrets
go in a private, Git-ignored `.env` file.

Admins visit `/admin`, receive an emailed code, and return to the site to edit.
**Edit** enables text editing and Page details; **Save changes** publishes the
draft, and **Cancel** discards it.

Live sign-in requires HTTPS and email delivery. Commit and deploy access changes
before sharing the live `/admin` link. An empty admin list disables live sign-in;
local dev shows setup guidance.

Sessions survive browser closure, server restarts, and deployments. The default
idle window is 30 days; set `admin.session_idle_days` to 1–365 to change it.
Sign out revokes the current browser session. Remove an address from `admin.emails`
to revoke that admin's access. Do not edit or copy runtime session files.

### Theme editing

Admins choose **Theme** to adjust supported colors, fonts, and spacing. Changes
preview in their browser until saved; Cancel restores the original appearance.
Google Font family names can load a preview; system fonts need no download.

The panel edits supported variables already present in `site/css/theme.css`.
Use the stylesheet itself for developer changes beyond those controls.

### Photos

Designate an editable image with `<img snip src="/images/team.jpg" alt="Our team">`.
Dev imports the local JPEG/PNG and generates `src="{snip ID.src}"` and
`alt="{snip ID.alt}"`. The typed entry in `flea/snips.json` owns those values;
classes, layout and links stay in the template. Admins click the image in Edit
mode, then use the page Save/Cancel. Files live in `flea/media/`; commit processed
photos and JSON as seeds. Images cannot be nested inside text snips.

To add an admin-selectable photo to a multisnip, use a named image field:

```html
<img snip="photo" src="/images/person.jpg" alt="Person">
```

Admins click the item's photo to choose a library image or upload a JPEG/PNG.
Done previews the choice; Save changes publishes new uploads to the library.
Cancel discards the draft without uploading. Empty photos show a
neutral portrait and an Add photo control during Edit. Normal deployment retains
live photos; `flea pull --snips-only` retrieves their metadata and files together.

### Videos

`<video snip src="/videos/intro.mp4" controls></video>` imports a local seed.
Videos require optimized MP4 files with H.264 video, optional AAC-LC audio,
and fast-start metadata. Files may be up to 25 MiB; Flea does not convert them.

Dev generates src/poster snip bindings. Admins click the video in Edit mode to
replace it or its poster, then Save changes or Cancel with the page. No separate
Photos/Videos buttons. Metadata lives in `flea/snips.json`; MP4 files in
`flea/media/` are Git-ignored.
Deployment stages referenced local files and preserves live replacements.

Pulls and backups include video files. On a fresh clone, restore missing video
files before validation, using `flea pull --snips-only` when a live destination is
configured. Storage is limited to 64 MiB including retained replacements, with no
automatic deletion. Video fields in multisnips are not supported.

### Contact forms

The starter includes `site/contact/index.html` and
`site/contact/thank-you.html`. Enable the bundled Contact app under
`apps.contact` in `.flea.config.json`, with a `mount` and the recipient in
`email.to`. This replaces the earlier top-level `contact` settings and the
`/_flea/contact` route; there is no alias or automatic migration, and deployment
requires protocol 23. Local preview mode needs no provider credentials. For real
delivery, configure email with `flea email`. Restart dev after changing contact
settings.

The form posts to the mount (for example `/contact`) with `name`, `email`,
`message`, and the hidden `website` field, and includes an explicit `{csrf}`
control. Keep those names and the control intact. Success redirects to the
mount-relative `thank-you` page with a visitor-bound notice. Messages are
emailed, not stored, and validation re-renders the form with safe values.
General-purpose form handling is not available.

### Email templates

Customize sign-in and contact messages in `mailers/*.email.html` and
`mailers/*.email.txt`. HTML messages share `mailers/main.layout.html`.

Mail bodies support escaped `app.*` and `email.*` values. Keep the layout's
`{slot}` intact. Recipients, subjects, and sender settings are configured outside
these templates.

Preview a sample sign-in email without sending it:

```sh
flea email preview sign-in > /tmp/flea-email-preview.html
```

Open the resulting file in your browser. Use `contact` instead of `sign-in` to
preview a contact message.

### Custom sign-in pages

Edit `site/admin/index.html`, `site/admin/code.html`, `site/admin/setup.html`, and
`site/admin/error.html` to change the sign-in experience. Their leading settings
comments select `site/admin.layout.html`. Keep form actions, field names, and
hidden `{admin.csrf}` fields intact.

In the public layout, preserve the `app.current_admin` conditional blocks, the
CSRF meta tag, and the `{render /admin/toolbar}` include with the toolbar's
version/revision fields. These connect your layout to the editor. The server
supplies the signed-in identity; never configure `app.current_admin` yourself.

### Agent-assisted work

Focused setup skills are available at `flea/skills/add-blog/SKILL.md` and
`flea/skills/add-contact-form/SKILL.md`. Ask your agent to "add a blog" or "add a
contact form"; `AGENTS.md` routes it to the supported setup and template values.
Blog requires Flea 0.15.0 or newer. These skills are bundled with newly generated
sites starting in 0.15.1; existing projects can copy them from a fresh site and
add the matching routes to their own `AGENTS.md`, preserving custom instructions.

To import an existing brochure site, give your coding agent the source URL or
export and `flea/skills/import-brochure/SKILL.md`. The bundled instructions cover
content, navigation, basic Prewind styling, and editable snips.

`AGENTS.md` points to this skill. Both stay in Git but are excluded from serving
and deployment; no particular agent plugin is required.

## Deployment

### Set up a server

Run `flea setup` from your Mac to configure the remote server over SSH. Point your
domain at the server and allow inbound TCP ports 80 and 443 before enabling HTTPS.

### Deploy your site

From the site folder, run:

```sh
flea deploy
```

The first run asks for the branch's server, domain, WWW redirect, and HTTPS
preferences, then saves the destination under `deploy.branches` in
`.flea.config.json`. It can run before your first commit and offers to commit
reviewed changes before deploying.

Deployment ships committed source; unresolved working-tree changes block it.
Each branch has its own destination. Use `flea deploy -configure` to change one.
Email-enabled sites prompt for server-side Resend settings when needed.

### Sync admin changes

Normal deployment preserves live editable content and theme changes, seeding new
IDs from your local site. To bring live edits back into your project, stop dev and
run:

```sh
flea pull
```

Use `--snips-only` or `--theme-only` to select one. Review and commit the result.
Pull never retrieves configuration, secrets, or sessions.

To intentionally publish local content or theme edits:

```sh
flea deploy --with-snips
flea deploy --with-theme
```

You can combine both flags. Existing live sites require a pull baseline first.
Conflicts ask you to choose explicitly between live and local changes.

## Backups and recovery

Before larger changes, create a backup at a new file path outside the site:

```sh
flea backup ../before-redesign.flea-backup
```

Use `flea rollback --list` to see retained code releases. Rolling back code retains
current admin content. `flea restore ../before-redesign.flea-backup` restores
backed-up content too and requires confirmation. Backups exclude credentials and
sessions.

## Troubleshooting

- **The site fails validation:** run `flea status` and fix the reported file.
- **A page returns 404:** check the filename and use its extensionless URL.
- **New snips block deployment:** run dev to extract them, then commit the results.
- **Local sign-in needs email setup:** run `flea email`, then return to `/admin`.
- **Pull cannot acquire the project lock:** stop the site's dev process first.

Dev shows local error details and recovers after invalid edits are fixed. Request
logs omit query strings and successful internal reload polls.

## Command help

Run `flea --help` for the command overview, or add `--help` to a command for its
options, such as `flea deploy --help`.

General field schemas, content revision history, video transcoding, and
unrestricted raw HTML rendering are not currently supported.
