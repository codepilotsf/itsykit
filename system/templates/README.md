# ItsyWeb

New sites use `itsy.config.json`, `itsy/`, and `<!-- itsy` page settings.
Existing `.flea.config.json`, `flea/`, and `<!-- flea` projects remain supported
without conversion. Keep using an existing site's selected files rather than
creating a second layout. This runtime requires server protocol 30;
run `itsy update-server` explicitly before deploying to an older server.

Build pages with HTML, style them with CSS, and optionally manage Blog posts in the admin editor. ItsyWeb builds the server, and serves multiple sites with HTTPS directly.

## Getting started

The included Contact mini app uses a normal HTML form and `{csrf}`. Enable it
only on sites that need it, as described below. Blog is an optional built-in content application, described below. Custom
managed apps use validated local executable packages.

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

`itsy new` creates a clean layout and ordinary HTML homepage. Use `itsy new --starter
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
enable `dev.email_preview` in `itsy.config.json`: sign-in and contact emails
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

Edit ordinary HTML in `site/index.html` and its leading title/description settings
comment. Edit `itsy/data/*.json` for developer-managed lists. No extraction occurs.

### 3. Check and save your work

```sh
itsy status
```

This validates the site without rewriting files. Review your changes, then commit
the changed HTML, data and assets together.

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
| `itsy.config.json` | Site settings and branch deployment destinations |
| `itsy/snips.json` | Blog posts and cover-photo library |
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

Add a top-level `redirects` object in `itsy.config.json`, alongside `app`:

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

Store public site-wide values in the `app` object in `itsy.config.json`:

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

Add a top-level `seo` object alongside `app` in `itsy.config.json` when the
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

### Page content and metadata

Pages are ordinary developer-authored HTML. Structured lists and objects live
in `itsy/data/*.json`, accessed through `data.*`; use `#each` for repeated content.
Images and videos are ordinary static files under `site/`, referenced by their
public URLs. CSS, Prewind and theme tokens remain developer-managed and deploy
with code. No page toolbar, page text/media editor, collection editor, extraction,
legacy named snips or user-editable theme exists.

Optional Blog keeps its rich post editor and cover-photo library. Its existing
private file is `itsy/snips.json`, now containing only optional `blog` and `media`
fields. An empty site uses `{}`. Old page-CMS fields and pending extraction state
are rejected without rewriting them. Back up and explicitly convert old sites
before using this incompatible runtime; no migration or deletion is automatic.

Ordinary page metadata belongs in one leading HTML settings comment, before markup:

```html
<!-- itsy
title: Visit the bookshop
description: Find our opening hours and plan a visit.
layout: /custom.layout.html
-->
<h1>Visit the bookshop</h1>
```

Each value is single-line plain text, without surrounding quotation marks. Split
at the first colon; later colons and braces stay literal. `layout` is optional and
must name an existing site-relative private `*.layout.html` file. The existing
`<!-- flea` spelling is also supported. Unknown and duplicate settings fail
validation, including duplicate empty values. This is not YAML.

`{page.title}` and `{page.description}` render with contextual escaping, for
example `<title>{page.title}</title>` and
`<meta name="description" content="{page.description}">`. Missing or empty title
uses `Home` at `/`, otherwise the final URL component with hyphens replaced by
spaces and its first letter capitalized. Missing or empty description is empty.
Comment values are never compiled as template source or copied into JSON.
Blog posts use their saved title and excerpt instead of template comment metadata.
There are no saved ordinary-page metadata overrides or page metadata editor.

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
`{:else if app.email}`. Use ordinary JSON data for developer-managed lists.

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

### Blog administration

`/admin` is the email-code sign-in entry point. After verification it returns to
`/admin`; authenticated admins go to `/admin/blog` when Blog is enabled. Without
Blog they see a status page with Sign out, not an editor for website pages.
`/admin/blog/edit` creates or edits a post, and `/admin/blog/save` saves, publishes
or unpublishes it. New posts save as drafts. Only published posts enter public
routes. Save on a published post updates it immediately; Unpublish returns it
to a private draft. See the Optional Blog section below.

Keep `site/admin.layout.html` and the private sign-in/Blog templates. Admin forms
use their own hidden `csrf` field from `{admin.csrf}`. Blog multipart saves also
send `X-CSRF-Token`. Exact Origin, host-bound HttpOnly/SameSite sessions, production
HTTPS, upload bounds and snapshot versions remain required. Public layouts need
only an ordinary `/admin` link. They have no toolbar mount, `app.current_admin`
state or admin CSRF meta. Public Contact forms retain the separate `{csrf}` input.
The removed `/admin/page`, `/admin/editor-context`, `/admin/theme`, `editor.js`,
`editor.css` and `toolbar.html` endpoints return 404. Only Blog's `blog.js` asset
includes the pinned Squire and DOMPurify libraries.

#### Developer-managed styling

Edit `site/css/theme.css` and ordinary stylesheets. Prewind/tokens, static assets
and committed CSS deployment remain supported. There is no theme editor or overlay.

### Styling

Edit `site/css/theme.css` for colors, fonts and fluid sizing. Prewind and theme
tokens remain available. Styles are developer-managed files: ordinary deployment
publishes committed CSS, and rollback restores the selected release's CSS.

### Static images and videos

Store ordinary assets under `site/` and reference their public URLs. Blog covers use its private media library.

### Contact forms

The starter includes `site/contact/index.html` and `site/contact/thank-you.html`.
Merge `"apps": {"contact": {"mount": "/contact", "email": {"to": "owner@example.com"}}}`
into `itsy.config.json`, preserving other settings. Local preview mode needs no
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

Public layouts use ordinary `/admin` links. Admin and Contact forms retain their
separate CSRF contracts; there is no page toolbar or public admin CSRF meta.

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
`itsy.config.json`. It can run before your first commit and offers to commit
reviewed changes before deploying.

Deployment ships committed source; unresolved working-tree changes block it.
Each branch has its own destination. Use `itsy deploy -configure` to change one.
Email-enabled sites prompt for server-side Resend settings when needed.

### Sync Blog changes

Commit `itsy/snips.json` when seeding Blog posts or cover metadata. Normal
`itsy deploy` preserves live posts and photos by stable ID and adds source-only
IDs. Omitting an ID never deletes a live post or photo. Publish code/CSS normally.

```sh
itsy pull
# Review and commit local Blog content and photos.
itsy deploy --with-content
```

`pull` retrieves Blog posts and cover photos from this branch's saved destination,
keeps local-only IDs and never pulls CSS, templates, configuration or sessions.
Stop dev first; selected local files must be committed. Explicit publication
requires a destination-scoped pull baseline for existing sites. Posts and photos
compare as whole records; conflicting edits require an explicit choice and stop
noninteractively. The expected remote digest is checked under the deployment lock.
`itsy status` reports content differences read-only. `--with-content` is removed
without an alias. This contract requires deployment protocol 30.

Backups include Blog content and library photos. Rollback preserves current
content while selecting earlier code/CSS. Restore requires an expected snapshot
and explicit confirmation. Retained files are never pruned automatically.

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
- **Old page-CMS markup/content blocks validation:** back up and explicitly convert it to HTML/JSON; no automatic migration occurs.
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
source edits with `--with-content`. Draft text is private; uploaded photos are public.
See https://gitlab.codepilot.com/sean-schertell/flea/-/blob/main/docs/blog.md for
the JSON/token contract.
