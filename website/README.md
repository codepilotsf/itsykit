# ItsyWeb

<img src="site/images/logo.svg" alt="ItsyWeb" width="233" height="50">

Write HTML. Skip the build pipeline.

You wanted to make a website, not maintain a toolchain. ItsyWeb adds layouts,
partials, variables, loops and clean URLs to HTML files. Its web server assembles
the shared pieces when it serves the page. You keep writing HTML and CSS.

The same tool gives you local development with live reload, Ubuntu server setup,
SSH deployment, automatic HTTPS, backups and code rollback. Bring a server you
control, SSH access and a domain; ItsyWeb handles the web server and certificates.
There is no frontend build step or language runtime to install.

Use `itsy` for commands, `itsy.config.json` for site settings, and `itsy/` for
content and runtime files. Downloads are served from `itsyweb.dev`. This
repository contains the website and documentation, not the CLI source.

## Getting started

Install the current public test release for Apple Silicon macOS or x86-64 Linux:

```sh
curl -fsSL https://itsyweb.dev/install.sh | bash
```

See [installation notes](https://itsyweb.dev/docs#installation) for PATH setup and
current testing limits. No Intel Mac or Linux ARM64 CLI download is available;
both supported CLIs can deploy to x86-64 and ARM64 Linux servers.

Create a site and start the local server:

```sh
itsy new my-site
cd my-site
itsy dev
```

This repository already is a site, so start with `itsy dev`. Enter a site name
if prompted; the admin email is optional. Open the Website URL printed by dev.
Skip admin setup to start with HTML pages.

### Edit the HTML

Replace the contents of `site/index.html` with your page markup:

```html
<h1>Hello, from our little corner.</h1>
<p>A website made with HTML and CSS.</p>
```

The default `site/main.layout.html` supplies the document wrapper, header and
footer. Your page goes in its `{slot}`. Change shared markup once in the layout;
save and the browser reloads. Add `site/about.html` to create `/about`.

For existing sites, preserve their layouts and content rather than replacing
them with this example. The optional editor uses generated references into
`itsy/snips.json`; ordinary page text can stay directly in HTML.

### Check and save

```sh
itsy status
```

Status validates without rewriting files, shows site counts and Git state, and
checks the current branch's saved deployment over SSH when one exists. Review
and commit the changes before deploying. New sites have a Git repository on
`main` with an `Initial site` commit, unless Git is unavailable or you pass
`--skip-git`. Creation uses your normal Git identity, hooks and signing settings;
a failed commit preserves the site with retry instructions. It prints commands
to connect a remote repository without creating one for you.

### Starting options

Use `itsy new .` in a completely empty directory, or pass a relative or absolute
path. A missing directory is created if its parent exists. Existing directories
must be empty, including hidden files and `.git`; file targets and directory
symlinks are refused. The folder name supplies the default site name.

`itsy new` creates a clean homepage. `itsy new --starter my-site` (or `-s`) adds
example pages, a fictional team and a contact form. Generated navigation uses a
native mobile popover without menu JavaScript.

Running `itsy dev` again replaces this site's existing dev server, including
LAN mode and port changes. Pull and reset remain exclusive. Stop a dev process
started with an older binary manually once. Dev starts at 3000 and tries the
next free port; an explicit `--port` stays fixed and `--port 0` selects a free port.

Use `itsy dev --lan` (or `-l`) to test from another device on a trusted private
IPv4 network. Open a printed Network URL. LAN mode uses unencrypted HTTP; never
forward the development port and restart dev after changing networks.

## Project structure

| Path | Purpose |
| --- | --- |
| `site/index.html` | Homepage |
| `site/main.layout.html` | Public layout, including the header and footer |
| `site/admin.layout.html` | Sign-in layout |
| `site/css/` | Stylesheets and theme tokens |
| `site/images/` | Developer-managed images and logo |
| `itsy.config.json` | Site settings and branch deployment destinations |
| `itsy/snips.json` | Editable content, page metadata, and media records |
| `itsy/data/` | Optional developer-managed JSON data |
| `itsy/media/` | Managed photos and videos |
| `itsy/icons/` | Downloaded icon cache |
| `itsy/runtime/` | Private, Git-ignored sessions and recovery files |
| `mailers/` | Private email templates |
| `itsy/skills/` | Portable instructions for coding agents |

Public URLs omit `site/`. Configuration, mailers, runtime files, and agent skills
are never served as public assets. Keep secrets out of configuration and Git.

## HTML, with reusable pieces

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
`{:else if app.email}`. For admin-managed lists, see the optional multisnips section below.

For negative conditions, use `{#unless app.email}...{/unless}`. It uses the
inverse of if truthiness and supports `{:else}` and `{:else if dotted.path}`.

### Styling

The starter uses Prewind 1.3.1 from unpkg, without Node or a CSS build step.

- Edit `site/css/theme.css` for colors, fonts, and fluid sizing.
- Edit `site/css/global.css` for site-wide rules in `@layer styles`.
- Use Prewind utilities for individual layout and spacing adjustments.
- Use `container` for centered page content.

`global.css` imports the theme, typography, forms, button styles, and navigation
transitions. This site's font is Inter from Google Fonts, with Navy and Ice
colors controlled by shared theme tokens. Replace
`site/images/logo.svg` with your own logo.

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

Dev downloads only referenced icons as `itsy/icons/name.svg`. The directory
includes `.keep`; each SVG includes its license notice as a comment. Commit
these files for offline use. Deploy also resolves missing icons;
`itsy status` never downloads them.

Quote attribute values. Classes, dimensions, stroke attributes, and `aria-` and
`data-` attributes are supported; event handlers and inline styles are not.
Unavailable icons are omitted with terminal warnings. Invalid syntax or unsafe
SVGs fail validation.

### Page titles and descriptions

Page metadata lives under the route in `pages` in `itsy/snips.json`. Update its
`title` and `description` fields. Layouts use `{page.title}`, `{page.description}`
and `{page.path}`. Page text does not have to be editable to use metadata.

## Deployment

### Set up a server

Run `itsy setup` from your computer to configure the remote server over SSH. Point your
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

With HTTPS enabled, ItsyWeb requests and renews Let's Encrypt certificates.
Setup does not rent a server or create DNS records. One host can serve multiple
sites; the server and inbound ports must be available before enabling HTTPS.

## Backups and recovery

Before larger changes, create a backup at a new file path outside the site:

```sh
itsy backup ../before-redesign.itsy-backup
```

Use `itsy rollback --list` to see retained code releases. Rolling back code retains
current admin content. `itsy restore ../before-redesign.itsy-backup` restores
backed-up content too and requires confirmation. Backups exclude credentials and
sessions.

`itsy reset` returns the local site to the clean default after confirmation.
Stop dev first. It replaces all local files and settings except Git history,
including uploads, `.env`, customizations and deployment destinations. The
original files remain in the private sibling backup printed by the command.
It makes no commit or remote change.

## Migration redirects

ItsyWeb 0.20.0 adds 301 redirects for old site URLs. Merge a top-level object
into `itsy.config.json`, preserving other settings:

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

Use the `itsy-site` skill to inventory an old website from sitemaps and links,
map replacements and verify responses. Provide old URL/analytics exports for
orphaned pages; leave ambiguous mappings for a decision. Run `itsy status` and
verify 301/Location and the final content locally. Update both CLI and server to
protocol 28 before deploying; public 0.19.0 and earlier ignore these rules. See
[the public guide](https://itsyweb.dev/docs#redirects).

## Search discovery

ItsyWeb 0.20.0 can generate a sitemap, advertise it in robots.txt and add missing
canonical tags. A canonical tag identifies a page's preferred address. Update
the local CLI with `itsy update`; for an older shared server, run
`itsy update-server` from a site with that server saved as the current branch's
deployment destination. The server update briefly restarts every site on it.
Deploying website files does not update the server executable.

Add this object alongside `app` in the site's `itsy.config.json`, preserving
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

Run `itsy status`, then `itsy dev`. At the printed local address, check
`/sitemap.xml`, `/robots.txt`, a public page's canonical tag in View Source and
an excluded page's `X-Robots-Tag: noindex` response header. Commit the reviewed
configuration, deploy from the intended branch with `itsy deploy`, and repeat
those checks on the public domain.

Custom robots.txt/sitemap.xml files and existing canonical tags take precedence.
Maintain sitemap declarations and exclusions yourself in custom discovery files.
Without `seo`, automatic discovery is disabled. Staging protection still uses
the separate exact-host indexing policy.

Google Search Console is optional: generated robots.txt already advertises the
sitemap. To see Google's processing reports, verify ownership of the production
site and add `https://example.com/sitemap.xml` in its Sitemaps report. Submit the
URL, not an uploaded file; it stays current without resubmitting after every edit.
See [the public walkthrough](https://itsyweb.dev/docs#seo-discovery).

## Hostname indexing

ItsyWeb 0.19.0 adds exact-host indexing policy with deployment protocol 27.
Update the CLI and explicitly upgrade the shared server before using it:

```sh
itsy update
itsy update-server
itsy server indexing root@server staging.example.com noindex
itsy server list root@server
```

The hostname must already be registered. Type it to confirm; changing policy
restarts the shared service. Staging receives `X-Robots-Tag: noindex, nofollow`
without a website app flag or different templates. Policies survive deploy, sync,
code rollback and site restore. `allow` stops adding the restriction and preserves
owner-authored restrictions. Keep robots.txt crawlable so crawlers can discover
noindex; ItsyWeb warns about detected static site-wide blocks without rewriting it.
Site backups exclude policy; removal clears only owned hosts. Verify native
headers before removing an existing meta fallback. See the
[public documentation](https://itsyweb.dev/docs#indexing) and RELEASES.md for
actual rollout evidence.

## Optional editing and apps

The HTML, development and deployment workflow works without configuring an
editor or enabling an app. Leave developer-managed text in your HTML. Use the
features below when an admin needs to change content or a site needs a contact
form or blog. The public documentation covers [Blog](https://itsyweb.dev/docs#blog)
and [custom managed apps](https://itsyweb.dev/docs#custom-apps) in detail.

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
The starter includes a working team example with a photo field.

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
**Edit** controls text editing and Page details; **Save changes** publishes the
draft, and **Cancel** discards it.

Live sign-in requires HTTPS and email delivery. Commit and deploy access changes
before sharing the live `/admin` link. An empty admin list disables live sign-in;
local dev shows setup guidance.

Sessions survive browser closure, server restarts, and deployments. The default
idle window is 30 days; set `admin.session_idle_days` to 1–365 to change it.
Sign out revokes the current browser session. Remove an address from `admin.emails`
to revoke that admin's access. Do not edit or copy runtime session files.

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

The starter includes `site/contact/index.html` and
`site/contact/thank-you.html`. Enable the bundled Contact app under
`apps.contact` in `itsy.config.json`, with a `mount` and the recipient in
`email.to`. This replaces the earlier top-level `contact` settings and the
`/_itsy/contact` route; there is no alias or automatic migration, and deployment
requires matching client and server versions (0.20.5 uses protocol 28). Local preview mode needs no provider credentials. For real
delivery, configure email with `itsy email`. Restart dev after changing contact
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
itsy email preview sign-in > /tmp/itsy-email-preview.html
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

### Local email previews

New sites enable `dev.email_preview` in `itsy.config.json`. Sign-in and contact
emails open in the browser without sending or needing Resend credentials. Use
the preview's sign-in code normally. Previews expire after 10 minutes and stay
only in dev's memory (the latest 20); stopping dev discards them. If the browser
cannot open, use the private local preview link printed in the terminal.

For real local delivery, set `"dev": {"email_preview": false}`, run `itsy email`
and restart dev. Production ignores preview mode and needs real email
configuration. Restart dev after changing preview mode or sender details. In
LAN mode, previews open on the dev computer; use the code on your other device.
Do not save preview messages or sign-in codes in site files.

### Sync live edits

Normal deployment preserves live editable content and publishes committed CSS,
seeding new content IDs from source. To bring live edits and referenced media
back into the project, stop dev and run:

```sh
itsy pull
```

Review and commit the result. Pull does not retrieve configuration, secrets or
sessions. To intentionally publish reviewed local edits to existing content,
establish a pull baseline first, then run:

```sh
itsy deploy --with-snips
```

Conflicts require an explicit choice between live and local changes. Browser
Theme editing and the `--theme-only`, `--with-theme` and `--snips-only` flags were
removed in 0.14.0; edit CSS in source and publish it with a normal deployment.

## Working with coding agents

Focused setup skills are available at `itsy/skills/add-blog/SKILL.md` and
`itsy/skills/add-contact-form/SKILL.md`. Ask your agent to "add a blog" or "add a
contact form"; `AGENTS.md` routes it to the supported setup and template values.
Blog requires ItsyWeb 0.15.0 or newer. These skills are bundled with newly generated
sites starting in 0.15.1; existing projects can copy them from a fresh site and
add the matching routes to their own `AGENTS.md`, preserving custom instructions.

To import an existing brochure site, give your coding agent the source URL or
export and `itsy/skills/import-brochure/SKILL.md`. The bundled instructions cover
content, navigation, basic Prewind styling, and editable snips.

`AGENTS.md` points to this skill. Both stay in Git but are excluded from serving
and deployment; no particular agent plugin is required.

## Troubleshooting

- **The site fails validation:** run `itsy status` and fix the reported file.
- **A page returns 404:** check the filename and use its extensionless URL.
- **New snips block deployment:** run dev to extract them, then commit the results.
- **Local sign-in needs email setup:** use `dev.email_preview` for local previews,
  or run `itsy email` for real delivery; restart dev after changing mail settings.
- **Pull cannot acquire the project lock:** stop the site's dev process first.

Dev shows local error details and recovers after invalid edits are fixed. Request
logs omit query strings and successful internal reload polls.

## Command help

Run `itsy --help` for the command overview, or add `--help` to a command for its
options, such as `itsy deploy --help`.

General content field schemas, a content revision-history interface and video
transcoding are not supported. The optional rich-text editor sanitizes content;
this restriction does not apply to your developer-authored HTML templates.

## Website maintenance

### Sharing preview

The public layout includes Open Graph and large-image card metadata with each
page's title, description and canonical URL. The shared 1200 by 630 PNG lives at
`site/images/social-preview.png`; its editable artwork is
`design/social-preview.html`. Render that HTML in Chrome at 1200 by 630 pixels,
device scale 1, to regenerate the image. It uses the working product name, shared theme
tokens and terminal workflow from the homepage, with no frontend build dependency.
Keep the image URL absolute.

### Temporary packaging storage

ItsyWeb 0.20.1 respects `TMPDIR` for local deployment and setup. On Linux with a
small RAM-backed `/tmp`, use a private directory on verified disk-backed storage
for each job, and keep worktrees and retained review evidence outside scratch.
The public documentation includes the supported per-deployment command.

The approved ItsyWeb woven-loop identity and its outlined wordmark are retained
in `site/images/logo.svg`, `logo-light.svg` and `mark.svg`. The selected concept,
exact generation prompts and production record live under `design/itsy-web-logo/`
and `.impeccable/mocks/decision/`. Shipped SVGs need no fonts or build step.
