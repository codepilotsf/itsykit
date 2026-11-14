# ItsyWeb

New sites use `.itsy.config.json`, `itsy/`, and `<!-- itsy` layout settings.
Existing `.flea.config.json`, `flea/`, and `<!-- flea` projects remain supported
without conversion. Keep using an existing site's selected files rather than
creating a second layout. New-layout deployment requires server protocol 29;
run `itsy update-server` explicitly before deploying to an older server.

Build pages with HTML, style them with CSS, and let site admins edit their own content. ItsyWeb builds the server, and serves multiple sites with HTTPS directly.

## Getting started

The included Contact mini app uses a normal HTML form and `{csrf}`. Enable it
only on sites that need it, as described below. Blog is an optional built-in content application, described below. Custom
service apps remain planned.

For a new project, install ItsyWeb first, then run:

```sh
itsy new foobar
cd foobar
itsy dev
```

The planned installer is `curl -fsSL https://fleacms.com/install.sh | bash`.
Install the public test release with `curl -fsSL https://fleacms.com/install.sh | bash`.
Downloads cover Apple Silicon macOS and x86-64 Linux. Linux ARM64 CLI downloads
are discontinued; deploying to ARM64 Linux servers remains supported. See
https://fleacms.com/docs#installation for PATH setup and testing limits.
This generated folder already is a site, so you can skip directly to `itsy dev`.

Use `itsy new .` to create a site in the current empty directory. Named paths
also accept existing empty directories. Hidden files and `.git` count as content;
ItsyWeb refuses nonempty directories.

`itsy new` creates a clean layout and editable homepage. Use `itsy new --starter
foobar` (or `-s`) for example pages, a six-person fictional team and a contact form.
Navigation uses a native HTML popover on mobile, with no menu JavaScript.
The clean site has Home navigation; the starter has Team and Contact.

`itsy reset` returns this site to the clean default after confirmation. It replaces
all local files/settings except `.git`, including uploads, `.env`, customizations
and deployment destinations. Stop dev first. Original files remain in the private
sibling backup path printed by the command. No commit or remote change is made.

With ItsyWeb 0.19.0 and a matching protocol-27 client/server, staging indexing can be configured
outside this site: `itsy server indexing root@server staging.example.com noindex`.
The hostname must already be registered. This sends a noindex/nofollow response
header and survives deployments without an app flag or different templates.
`itsy server list root@server` shows each host's effective policy. Keep robots.txt
crawlable so crawlers can discover the header. Earlier releases do not include
this feature; arrange the framework upgrade separately from website deployment.

### 1. Start the development server

With the `flea` command installed, run this from your site folder:

```sh
itsy dev
```

On the first run, enter your site name and an optional admin email. New sites
enable `dev.email_preview` in `.itsy.config.json`: sign-in and contact emails
open in your browser without sending or needing Resend credentials. The preview
looks like an opened email, with sender, recipient, subject and date/time.
Use its sign-in code normally. Previews expire after 10 minutes and stay only in
dev's memory (the latest 20); stopping dev discards them. If the browser cannot
open, use the local preview link printed in the terminal.

For real local delivery, set `"dev": {"email_preview": false}`, run `itsy email`
and restart dev. Production ignores preview mode and always needs real email
configuration. Restart dev after changing preview mode or its sender details.

Open the Website URL printed in the console. Dev starts at port 3000 and
explains occupied ports before trying 3001, 3002 and so on. An explicit
`--port` stays fixed; `--port 0` lets the operating system choose a port.
Leave the server running while
you work; saved changes reload the browser. Starting dev again gracefully replaces
the same site's existing dev server, including LAN mode or port changes. Pull and
reset remain protected while running. Stop older dev binaries manually once.

For another device on the same trusted network, use `itsy dev --lan` (or `-l`)
and open a printed Network or Hostname URL. Named URLs work when they resolve to
the dev machine; Tailscale IPv4 connections are supported. Sign-in, editing, media, forms and reloads
work there too. Email previews open on the computer running ItsyWeb; enter that code
on the other device. LAN mode uses unencrypted HTTP on private IPv4 networks.
Do not forward the port. Allow it through your firewall if needed and restart dev
after changing networks. Without the flag, access stays on this computer.

### 2. Make your first change

Open `site/index.html` to change the homepage structure. Edit `itsy/snips.json`
for existing welcome text and page metadata, or add a new editable heading:

```html
<h1 snip>Welcome to our site</h1>
```

Dev moves the text into `itsy/snips.json` and replaces it with a generated snip
directive. See [editable content](#editable-content) for details.

### 3. Check and save your work

```sh
itsy status
```

This validates the site without rewriting files. Review your changes, then commit
the templates and extracted content together.

New sites have an `Initial site` commit on `main`, unless created with
`--skip-git`, without Git installed, or with a Git error reported during creation.
ItsyWeb respects your Git identity, hooks and signing settings. To connect a blank
remote project, run:

```sh
git remote add origin <repository-url>
git push -u origin main
```

No remote is configured automatically. There is no existing `origin` to rename.

## Project structure

| Path | Purpose |
| --- | --- |
| `site/index.html` | Homepage |
| `site/main.layout.html` | Public layout, including the header and footer |
| `site/admin.layout.html` | Sign-in layout |
| `site/css/` | Stylesheets and theme tokens |
| `site/images/` | Developer-managed images and logo |
| `.itsy.config.json` | Site settings and branch deployment destinations |
| `itsy/snips.json` | Editable content, page metadata, and media records |
| `itsy/data/` | Optional developer-managed JSON data |
| `itsy/media/` | Managed photos and videos |
| `itsy/icons/` | Downloaded icon cache |
| `itsy/runtime/` | Private, Git-ignored sessions and recovery files |
| `mailers/` | Private email templates |
| `.agents/skills/` | Shared core, design, Prewind and ItsyWeb authoring skills |
| `.agents/site-requirements.md` | Version-specific site requirements |
| `itsy/skills/` | Focused Blog, Contact and import guides |

Public URLs omit `site/`. Configuration, mailers, runtime files, and agent skills
are never served as public assets. Keep secrets out of configuration and Git.

## Features

### Pages and routing

Create `site/about.html` to add `/about`. You can also use
`site/about/index.html`, but not both for the same route.

Use `/about` in links. Explicit `.html` URLs and trailing slashes return 404
unless listed as migration redirects.

### Migration redirects

Add a top-level `redirects` object in `.itsy.config.json`, alongside `app`:

```json
"redirects": {
  "/about-us.html": "/about",
  "/services/": "/services"
}
```

Each exact, case-sensitive old path returns **301** on GET/HEAD. Destinations
must be existing public pages/assets on this site. Incoming queries survive unless
the destination supplies its own query; a final `?` clears the query. Destinations
may include fragments. No wildcard, query-specific or external-URL rules exist.
Private/internal paths, duplicate decoded sources, chains, loops and conflicts
with active routes or managed app subtrees fail validation. Former Blog slugs
can be listed explicitly. Update redirects when removing their destination.

Ask the `itsyweb-site` skill to inventory the old website and prepare redirects,
or use `itsy/skills/import-brochure/SKILL.md` during a full import. Preserve old URL
spellings and report URLs the crawl cannot discover or map. Run `itsy status`
and verify each 301 and its final destination on a local server. This requires
deployment protocol 28; public 0.19.0 / protocol 27 and earlier ignore these rules.
An inventory does not authorize deployment or runtime upgrades.
Assets keep their paths: `site/images/logo.svg` is `/images/logo.svg`.
Files and directories beginning with `_` or `.` are private.

### Layouts

`site/main.layout.html` wraps public pages. Its `{slot}` receives the page content;
the starter keeps the header and footer directly in this file.

To select a custom layout, put a settings comment before the page markup:

```html
<!-- itsy
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

Store public site-wide values in the `app` object in `.itsy.config.json`:

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

### Search discovery and SEO

Add a top-level `seo` object alongside `app` in `.itsy.config.json` when the
production address is known:

```json
"seo": {
  "url": "https://example.com",
  "exclude": ["/contact/thank-you"]
}
```

`url` must be an HTTPS origin without credentials, a subdirectory, query or
fragment. ItsyWeb generates `/sitemap.xml` from public pages and published blog
posts, plus `/robots.txt` with a production Sitemap line. Missing canonical tags
are inserted into public responses without changing template files; HTML
fragments use a canonical HTTP Link header. No external service is needed.

`exclude` contains exact page paths, without wildcards or trailing slashes.
Excluded pages stay accessible with a noindex header and are omitted from the
generated sitemap and canonicals. Admin routes, drafts, assets and custom managed
app routes are omitted from automatic discovery. Enabled Contact's form is
included; its thank-you page is excluded and noindex by default. Existing robots
or googlebot noindex metadata also excludes a page from the sitemap.

Custom `site/robots.txt`, `site/sitemap.xml` and canonical tags are preserved.
Add the Sitemap line and manage exclusions yourself in custom discovery files.
A page with a custom canonical enters the generated sitemap only when that
canonical exactly matches its absolute production URL. ItsyWeb does not invent
modification dates; beyond 50,000 URLs or 50 MiB, provide a custom sitemap index.

Without `seo`, automatic SEO behavior is disabled. Setting a production URL does
not register hosts, configure TLS or protect staging: keep the separate server
indexing policy described above. A supporting ItsyWeb server executable is required;
deploying this site's configuration alone does not upgrade it.

### Editable content

Add a bare `snip` attribute to an element whose content admins should edit:

```html
<section snip>
  <h2>About us</h2>
  <p>We build things that last.</p>
</section>
```

While dev is running, ItsyWeb extracts the inner HTML into `itsy/snips.json`, removes
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

Put developer-managed data in `itsy/data/`. For example, `itsy/data/team.json`
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
`site/images/logo.svg` with your own full logo, `site/images/mark.svg` with your
standalone mark, and the favicon/touch-icon assets with your site icons.

### Icons

Reference Lucide icons directly in HTML:

```html
{icon arrow-right}
{icon search size="20" aria-label="Search"}
```

Dev downloads only referenced icons as `itsy/icons/name.svg`. The directory
includes `.keep`; each SVG includes its license notice as a comment. Commit
these files for offline use. Deploy also resolves missing icons;
`itsy status` never downloads them.

Quote attribute values. Classes, dimensions, stroke attributes, and `aria-` and
`data-` attributes are supported; event handlers and inline styles are not.
Unavailable icons are omitted with terminal warnings. Invalid syntax or unsafe
SVGs fail validation.

### Admin sign-in and editing

Add an approved admin from your site folder:

```sh
itsy admin admin@example.com
```

The command adds the address to `admin.emails`. It starts local email setup only
when preview mode is off and delivery is unconfigured. Run `itsy email`
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

### Styling

Edit `site/css/theme.css` for colors, fonts and fluid sizing. Prewind and theme
tokens remain available. Styles are developer-managed files: ordinary deployment
publishes committed CSS, and rollback restores the selected release's CSS.

### Photos

Designate an editable image with `<img snip src="/images/team.jpg" alt="Our team">`.
Dev imports the local JPEG/PNG and generates `src="{snip ID.src}"` and
`alt="{snip ID.alt}"`. The typed entry in `itsy/snips.json` owns those values;
classes, layout and links stay in the template. Admins click the image in Edit
mode, then use the page Save/Cancel. Files live in `itsy/media/`; commit processed
photos and JSON as seeds. Images cannot be nested inside text snips.

To add an admin-selectable photo to a multisnip, use a named image field:

```html
<img snip="photo" src="/images/person.jpg" alt="Person">
```

Admins click the item's photo to choose a library image or upload a JPEG/PNG.
Done previews the choice; Save changes publishes new uploads to the library.
Cancel discards the draft without uploading. Empty photos show a
neutral portrait and an Add photo control during Edit. Normal deployment retains
live photos; `itsy pull` retrieves their metadata and files together.

### Videos

`<video snip src="/videos/intro.mp4" controls></video>` imports a local seed.
Videos require optimized MP4 files with H.264 video, optional AAC-LC audio,
and fast-start metadata. Files may be up to 25 MiB; ItsyWeb does not convert them.

Dev generates src/poster snip bindings. Admins click the video in Edit mode to
replace it or its poster, then Save changes or Cancel with the page. No separate
Photos/Videos buttons. Metadata lives in `itsy/snips.json`; MP4 files in
`itsy/media/` are Git-ignored.
Deployment stages referenced local files and preserves live replacements.

Pulls and backups include video files. On a fresh clone, restore missing video
files before validation, using `itsy pull` when a live destination is
configured. Storage is limited to 64 MiB including retained replacements, with no
automatic deletion. Video fields in multisnips are not supported.

### Contact forms

The starter includes `site/contact/index.html` and `site/contact/thank-you.html`.
Merge `"apps": {"contact": {"mount": "/contact", "email": {"to": "owner@example.com"}}}`
into `.itsy.config.json`, preserving other settings. Local preview mode needs no
provider credentials. For real delivery, configure email with `itsy email`.
Restart dev after changing contact settings.

The form posts to `/contact` with `{csrf}`, `name`, `email`, `message`, and the hidden
`website` field. Keep those names intact. Messages are emailed, not stored.
Removing `apps.contact` disables the app. General-purpose form handling is not
available.

Validation re-renders with `contact.data.*` and `contact.invalid.*`; delivery
failure supplies `contact.error`. Success redirects to the mount's `/thank-you`
page, where `contact.flash.notice` is available once. These are escaped values,
not template source. Optional `email.from` overrides the configured sender;
visitor email remains Reply-To. There is no old contact config/route alias.

See https://fleacms.com/docs#custom-apps for the app model, the exact request and
response JSON, custom managed app packages and app-owned sessions. Remote or
registry-based app registration is not supported.

### Email templates

Customize sign-in and contact messages in `mailers/*.email.html` and
`mailers/*.email.txt`. HTML messages share `mailers/main.layout.html`.

Mail bodies support escaped `app.*` and `email.*` values. Keep the layout's
`{slot}` intact. Recipients, subjects, and sender settings are configured outside
these templates.

Preview a sample sign-in email without sending it:

```sh
itsy email preview sign-in > /tmp/flea-email-preview.html
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

Shared core, delivery, design, Prewind, ItsyWeb authoring and upstream Impeccable
skills ship in `.agents/skills/`. Root `AGENTS.md` explicitly loads core and routes
other work on demand. Conversation defaults yield to personal preferences;
project requirements and permissions still apply. No personal instruction files
or global tool settings are changed.

Focused guides ship in `itsy/skills/`:

- **add-blog**: requests like "add a blog" or "let's create a blog"; setup,
  post JSON, Blog template values and admin publishing.
- **add-contact-form**: requests like "add a contact form"; setup, exact fields,
  CSRF, feedback values and private mail templates.
- **import-brochure**: import or redesign an existing brochure/portfolio site,
  preserving its content and navigation. Interactive: choose a direction (two mood
  boards or references) and approve a sample page before the full build.
The shared **itsyweb-site** skill covers general authoring, real business
content and verification, loading the same **prewind** and **design** skills used
by other projects. Ongoing synchronization is not configured; installed source
information is recorded in `.agents/skills-source.md`.

Point your agent at the relevant `SKILL.md`. `AGENTS.md` routes these requests to
the matching skill. They stay in Git but are excluded from serving and deployment; no particular agent plugin is
required.

## Deployment

### Set up a server

Run `itsy setup` from your Mac to configure the remote server over SSH. Point your
domain at the server and allow inbound TCP ports 80 and 443 before enabling HTTPS.

### Deploy your site

From the site folder, run:

```sh
itsy deploy
```

The first run asks for the branch's server, domain, WWW redirect, and HTTPS
preferences, then saves the destination under `deploy.branches` in
`.itsy.config.json`. It can run before your first commit and offers to commit
reviewed changes before deploying.

Deployment ships committed source; unresolved working-tree changes block it.
Each branch has its own destination. Use `itsy deploy -configure` to change one.
Email-enabled sites prompt for server-side Resend settings when needed.

### Sync admin changes

Normal deployment preserves live editable content, seeding new IDs from your
local site. It publishes committed CSS, including `site/css/theme.css`.
To bring live snip/metadata and media edits back, stop dev and run:

```sh
itsy pull
```

Review and commit the result. Pull never retrieves styles, source templates,
configuration, secrets or sessions. To intentionally publish local snip edits:

```sh
itsy deploy --with-snips
```

Existing live sites require a pull baseline first. Conflicts require an explicit
choice between live and local changes; unattended conflicts stop. There are no
Theme controls or theme synchronization flags.

## Backups and recovery

Before larger changes, create a backup at a new file path outside the site:

```sh
itsy backup ../before-redesign.flea-backup
```

Use `itsy rollback --list` to see retained code releases. Rolling back code retains
current admin content. `itsy restore ../before-redesign.flea-backup` restores
backed-up content too and requires confirmation. Backups exclude credentials and
sessions.

## Troubleshooting

- **The site fails validation:** run `itsy status` and fix the reported file.
- **A page returns 404:** check the filename and use its extensionless URL.
- **New snips block deployment:** run dev to extract them, then commit the results.
- **Local sign-in needs email setup:** run `itsy email`, then return to `/admin`.
- **Pull cannot acquire the project lock:** stop the site's dev process first.

Dev shows local error details and recovers after invalid edits are fixed. Request
logs omit query strings and successful internal reload polls.

## Command help

Run `itsy --help` for the command overview, or add `--help` to a command for its
options, such as `itsy deploy --help`.

General field schemas, content revision history, video transcoding, and
unrestricted raw HTML rendering are not currently supported.

## Optional Blog

Merge `"blog": {"mount": "/blog"}` into the configuration’s `apps` object.
The private `site/blog/_index.html` and `_post.html` templates supply the listing
and posts. Sign in, then open `/admin/blog` for drafts, rich text, cover selection
and explicit Publish/Unpublish. Posts use the top-level `blog` array in
`itsy/snips.json`. Normal deploy preserves live posts; pull before publishing
source edits with `--with-snips`. Draft text is private; uploaded photos are public.
See https://gitlab.codepilot.com/sean-schertell/flea/-/blob/main/docs/blog.md for
the JSON/token contract.
