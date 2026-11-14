# ItsyWeb supported product contract

This source implements ordinary developer-authored HTML/JSON sites, optional
built-in Blog, bundled Contact and confined custom managed apps. Published
installers remain unchanged until an explicit release. This candidate requires
protocol 30. Historical release receipts retain their original scope.

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

## Blog administration

`/admin` is the email-code sign-in entry point. After verification it returns to
`/admin`; authenticated admins go to `/admin/blog` when Blog is enabled. Without
Blog they see a status page with Sign out, not an editor for website pages.
`/admin/blog/edit` creates or edits a post, and `/admin/blog/save` saves, publishes
or unpublishes it. New posts save as drafts. Only published posts enter public
routes. Save on a published post updates it immediately; Unpublish returns it
to a private draft. See [Blog](docs/blog.md).

Keep `site/admin.layout.html` and the private sign-in/Blog templates. Admin forms
use their own hidden `csrf` field from `{admin.csrf}`. Blog multipart saves also
send `X-CSRF-Token`. Exact Origin, host-bound HttpOnly/SameSite sessions, production
HTTPS, upload bounds and snapshot versions remain required. Public layouts need
only an ordinary `/admin` link. They have no toolbar mount, `app.current_admin`
state or admin CSRF meta. Public Contact forms retain the separate `{csrf}` input.
The removed `/admin/page`, `/admin/editor-context`, `/admin/theme`, `editor.js`,
`editor.css` and `toolbar.html` endpoints return 404. Only Blog's `blog.js` asset
includes the pinned Squire and DOMPurify libraries.

### Developer-managed styling

Edit `site/css/theme.css` and ordinary stylesheets. Prewind/tokens, static assets
and committed CSS deployment remain supported. There is no theme editor or overlay.

### Blog content synchronization

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
`itsy status` reports content differences read-only. `--with-snips` is removed
without an alias. This contract requires deployment protocol 30.

Backups include Blog content and library photos. Rollback preserves current
content while selecting earlier code/CSS. Restore requires an expected snapshot
and explicit confirmation. Retained files are never pruned automatically.

## Retained system contracts

Keep Go 1.26+, minimal dependencies, no frontend build, immutable Projects,
confined reads, contextual escaping, and independent Blog/content/auth storage.
Routing, layouts, includes, SEO, redirects, mail templates, Contact/form CSRF,
app isolation, developer CSS/Prewind and streamed static assets remain supported.
See templates.md, runtime.md, editor.md, deployment.md and ../blog.md for details.
Setup/deploy/pull/backup/recovery preserve exact destination/branch/commit checks,
locks, expected versions, ownership, permissions, journals and publication order.
No production/release action follows from development authorization.

Managed Ubuntu installations use systemd socket activation to retain TCP 80/443
through service restarts. New connections queue until the single server process
is ready; in-flight requests retain the five-second graceful shutdown window.
There is no guarantee against queue exhaustion, client timeouts or failed startup.
Setup installs the canonical service/socket units. The first explicit server update
converts only the exact original service without overrides, with executable and
unit recovery in one journal; customized installations require manual review.
This change does not alter the site/archive protocol or introduce a proxy.
