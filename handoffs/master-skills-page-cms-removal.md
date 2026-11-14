# Master-skills handoff: remove page CMS guidance, retain Blog

## Task and ownership

Update the canonical ItsyWeb site-authoring skill in master-skills to match the
owner-approved page-CMS removal. This handoff requests a shared-skill update,
not another framework implementation or website redesign. Preserve Blog's rich
admin editor, drafts, publishing, cover uploads, storage and synchronization;
preserve Contact, custom apps, ordinary HTML/JSON/static assets and developer CSS.

The ItsyWeb implementation is complete locally at **`eea4019`**, branch
`t3/remove-page-editing`, in
`/home/sean/Worktrees/itsyweb/t3-c3b38c40`. The framework implementation is
`d601d2a`; website work is `c353797`, with audit follow-ups `6ad9960` and `11ff8c3`.
Use `eea4019` as the exact reference, including its guidance/provenance fixes.
These commits have not been published or deployed. Do not edit this finished
ItsyWeb checkout as a side effect of the canonical skill update.

Read the target repository's AGENTS.md, README, core and task-delivery guidance.
Use its required skill-creator workflow for actual skill authoring. Inspect status
first, preserve unrelated work and use one agent unless delegation is authorized.

## Locate the current canonical source

Installed source repository:
`git@gitlab.codepilot.com:sean-schertell/master-skills.git`.
Both ItsyWeb consumers record upstream pin
`05e8168afe6497a565d92dc912e1a47e3db40ef6`, with explicit local adaptations for
this removal. Their manifests describe the installed adapted bytes; they do not
claim those bytes are an unchanged copy of the upstream pin.

A read-only inspection found `/home/sean/Worktrees/master-skills` at older commit
`1bd0622`. It still has `skills/fleacms-site/SKILL.md` and legacy `flea` consumer
paths in `sync/consumers.json`. It also has an unrelated untracked
`deployment-memory-guidance/` directory. Preserve that work. Confirm the intended
current checkout/branch and source before editing; do not assume this older
checkout is the latest canonical state, rename it blindly, or copy into it wholesale.
Resolve whether the canonical skill is now `itsyweb-site` or still `fleacms-site`.

## Authoritative reference files at eea4019

Prefer portable authoring guidance over website-specific adaptations:

- [Generated-site skill](../system/templates/.agents/skills/itsyweb-site/SKILL.md)
  and its `references/`: templates-content, apps, compatibility-operations,
  migration-redirects and SEO.
- [Generated-site requirements](../system/templates/.agents/site-requirements.md)
  and [README](../system/templates/README.md).
- [Blog addition guide](../system/templates/itsy/skills/add-blog/SKILL.md),
  [Contact guide](../system/templates/itsy/skills/add-contact-form/SKILL.md) and
  [brochure import guide](../system/templates/itsy/skills/import-brochure/SKILL.md).
  These are framework-owned guides, not necessarily canonical shared-skill files;
  use them to check the contract without inventing new ownership.
- [Framework architecture](../system/docs/agents/architecture.md),
  [Blog](../system/docs/blog.md), [admin](../system/docs/agents/editor.md) and
  [sync](../system/docs/agents/sync-theme.md) for implementation details.
- [Website-installed skill](../website/.agents/skills/itsyweb-site/SKILL.md)
  and its references show the consumer conversion. Do not copy website-specific
  release gates, paths or branding into generic shared instructions.

Check the actual implementation when wording disagrees. Do not restore the old
saved-page-override decision from the historical metadata handoffs.

## Supported candidate contract

### Ordinary pages and metadata

Write page copy directly in HTML. Use `itsy/data/*.json` and `#each` for structured
lists. Images and videos are ordinary static files under `site/`. Styles and
`site/css/theme.css` tokens remain developer-managed and travel with code.

One leading settings comment owns ordinary-page metadata and optional layout:

```html
<!-- itsy
title: Visit the bookshop
description: Find our opening hours and plan a visit.
layout: /custom.layout.html
-->
<h1>Visit us</h1>
```

Values are trimmed, single-line plain text, split at the first colon. Braces,
quotes, HTML characters and later colons remain data, never template source.
Unknown, duplicate and malformed settings fail, including duplicate empty values.
Keep private layout/path validation. The existing legacy `<!-- flea` spelling
remains supported; no new aliases or YAML are added.

Layouts use `{page.title}` and `{page.description}`, contextually escaped.
Missing/empty title falls back to Home at `/`, otherwise the route-derived title;
missing/empty description is empty. Published Blog posts use their saved title
and excerpt, overriding post-template comment metadata. There are no saved
ordinary-page metadata overrides, extraction writes or metadata editor.
Encode literal braces when showing template syntax in public HTML examples.

### Removed page CMS

Remove active directions/examples for text/media snips, named/structured snips,
multisnips/collections, extraction, page editing/formatting, page Save/Cancel,
page-toolbar mounts, `app.current_admin` public rendering and browser theme
controls. Public layouts need no admin CSRF meta. Blog's own Save/Cancel remains.

Old page-CMS fields, named-content directories and pending extraction journals
fail closed without automatic conversion or rewriting. Never discard old content
or media to make validation pass. Preserve a backup and require an explicitly
scoped source conversion before using the incompatible candidate on an old site.
No bulk migration, compatibility layer, automatic pruning or live-site cleanup
is authorized by this handoff.

### Retained Blog storage and commands

Keep the existing selected `itsy/snips.json`, or the supported legacy selected
`flea/snips.json`. The only optional top-level fields are `blog` and `media`;
`{}` is valid empty storage. The filename does not imply page-snips support.

- BlogPost: `id`, `title`, `slug`, `excerpt`, `body`, `status`, optional
  `published_at` and `cover`. Post IDs are ten case-sensitive alphanumeric chars.
- Photo map records: `file`, `name`, `alt`, `width`, `height`. Photo IDs are ten
  chars starting with a letter; the content hash is in the filename, not new
  `size` or `sha256` JSON fields. Files live in `itsy/media/`.
- Cover-upload input: `id`, `name`, `alt`, plus multipart cover bytes. Keep the
  existing JPEG/PNG bounds, sanitization and immutable-photo handling.

Use `itsy pull`, `itsy deploy --with-content` and `itsy status`.
`--with-snips` is removed without an alias. Normal deploy preserves existing live
posts/photos and seeds local-only IDs. Missing IDs do not delete live records.
Explicit sync requires the destination-scoped baseline and explicit conflict
resolution; posts/photos compare as whole records. Pull leaves CSS untouched.
Keep backup, restore, rollback, expected-state and locking safeguards.

`internal/snips`, `snips.json` and synchronization `Snips` wire fields still have
active Blog storage/transport uses. Do not cosmetically rename them or remove
Blog support to achieve a zero-match text search. Explicit rejection guards and
custom-app namespace restrictions are also live behavior.

### Admin, Contact and CSS

`/admin` is the emailed-code sign-in entry. Verification returns to `/admin`;
authenticated admins go to `/admin/blog` when enabled, otherwise a status page
with Sign out. `/admin/blog/edit` and POST `/admin/blog/save` remain.

Preserve private `{admin.csrf}` controls, sessions, exact Origin/Host, production
HTTPS and content-version checks. Blog multipart saves check `X-CSRF-Token`
before parsing, then form CSRF. Blog's bundled rich editor is served through
`/admin/assets/blog.js`; old page/editor-context/theme and page-editor assets
return 404. Public pages have no authenticated CMS state or global page-media
namespace; Blog covers use `blog.data.post.cover` and the admin photo list.

Contact keeps its separate `{csrf}` hidden input, email contract and private
local email previews. Developer-authored CSS/Prewind/theme tokens remain supported.
Do not turn removal of browser theme editing into removal of source CSS.

## Version and synchronization boundaries

Protocol **30** (`FLEA_DEPLOY=30`) is implemented but unreleased. The candidate
binary still prints 0.21.0; the published 0.21.0 installer/server uses protocol 29
and does not provide this new contract. Keep current-source guidance distinct
from public availability. A local commit or successful build does not authorize
a release, server upgrade or deployment.

Keep shared guidance portable and tied to the detected implementation/version.
Do not silently convert existing older sites or claim that the public installer
already contains this removal. Do not add a permanent website-only candidate
warning to every generic skill section.

Update canonical sources first and follow the target repository's checks and
consumer workflow. This handoff alone does not authorize consumer sync. Any later
sync must use a committed exact canonical revision, inspect drift and reconcile
these deliberate local adaptations rather than overwrite or bypass them.

The affected installed copies in this combined repository live at
`system/templates/.agents/` and `website/.agents/`, not a root-level
`templates/.agents/`. Inspect actual `sync/consumers.json` and current consumer
configuration before preparing a sync. Presence of an installed copy does not
establish its enrollment in automated synchronization. Preserve custom website
work and other consumers. Do not change Preflight, global Impeccable, personal settings, live
sync jobs, credentials or other sites as an incidental step.

## Verification and delivery

The implementation reference passed full framework tests, full race tests, vet,
canonical bundled builds, fresh minimal/starter HTTP Blog lifecycle checks and
T3 Blog/browser checks. The integrated website passed status, exact metadata and
sharing output on 16 routes, removed endpoint GET/POST 404 checks, Contact CSRF,
desktop/mobile navigation and literal examples. Fresh generated skill manifests
and both installed manifests pass. These are local candidate checks, not release,
real-email, Ubuntu/systemd or public-TLS acceptance.

For the canonical skill update, use the required skill validator, check relative
links and authorization boundaries, and audit active instructions/examples for
removed page-CMS guidance. Review Blog/auth/Contact/CSS preservation and
version-aware claims. Exercise sync scripts only if they change; do not rerun
unrelated application suites for prose edits. Preserve historical release receipts.

Finish with a scoped canonical change and report its revision, checks, any
remaining drift/version limits and the exact separately proposed consumer scope.
Do not publish, deploy, upgrade servers, enable live sync or alter other projects
without the relevant authorization.
