# Templates and content

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

A generated site uses `site/` for public templates/assets, root `mailers/` for
private email templates, and `flea/` for content, media, data, icons and skills.
`flea/runtime/` holds sessions and extraction recovery records and is Git-ignored.
`flea/skills/` is committed agent guidance, excluded from serving and archives.
Bundled import/redesign and site-building skills use two design directions and a
sample-page review before a substantial full build, unless the user explicitly
delegates those decisions. Keep generated README/AGENTS summaries aligned.
`.flea.config.json` at the project root stores public globals in the required `app` object.
Branch destinations live under `deploy.branches` in that same file. Preserve other
JSON fields when saving destinations and strip `deploy` from releases/backups.
Templates use app.site_name and other app.* values; no site.* compatibility alias.
New sites omit site_name until interactive dev prompts with the title-cased project
folder default and saves the answer. Saved names must be nonempty strings.
Read-only/noninteractive loads use the folder default without writing; deployment
requires a committed name so temporary release directories never change it.
Use this exact filename everywhere; no extensionless configuration fallback.
Deployment protocol 23 requires the JSON-suffixed configuration filename and
inline editing with the updated rich-content sanitization policy.
Owner-editable snips live in flea/snips.json: pages map canonical routes to title,
description, and snips; shared.snips holds reusable regions; archive holds removed
IDs. Entries have content and an optional label. Bare snip attributes need no
name or value. IDs are 10 case-sensitive alphanumeric chars,
crypto/rand with rejection sampling and active/archive collision checks.
Generate {snip ID} without quotes. Quoted markers are rejected; HTML comments
are ordinary comments and never render content or count as references, except
the leading flea page-settings comment described below.
Reject old flea attributes, directives, content paths and recovery journals;
never silently ignore or migrate them. Protocol 23 enforces the snip format.
Do not add layout-changing borders to editable content. Show dashed editable
boundaries using shadow-root overlays, stronger on hover/focus. Suppress the
native editable wrapper outline and restore developer styling when editing ends.
Keep keyboard focus indicators on toolbar buttons and form controls.
Validate rich references identically in the parser and region scanner;
IDs cannot contain braces, whitespace or punctuation.
Developer-managed structured objects live in flea/data/*.json and use data.*.
Legacy flea/snips/*.json and escaped named directives remain supported.

Bare img/video snip attributes import local seed files and become typed snip
entries. Rewritten templates use src="{snip ID.src}" and alt="{snip ID.alt}"
or poster="{snip ID.poster}". JSON owns those values; classes, dimensions,
loading, playback and link attributes remain developer-owned. Require complete,
matching bindings in their proper tag attributes, contextual escaping and the
correct stored type. Reject competing srcset/picture/video source children and
nested text/media snips before extraction. Explicit image snips may be inside
links without making link destinations editable. Never infer editability from
an ordinary media tag. Keep multisnip spelling unchanged.
Owner rendering adds data-snip-media only for authenticated editing; visitors
receive ordinary attributes. Save authorization checks rendered visible bindings,
not client-provided IDs alone. Preserve immutable snapshots and version checks.

LoadEditing (dev startup/watch only) extracts inner HTML of bare text snip regions,
removes the authoring attribute, and preserves the element and other source bytes.
The directive alone identifies extracted content; never regenerate a snip attribute.
Remove redundant attributes around existing directives through the same journaled
transaction, retaining IDs/content and existing labels unless a new label is supplied.
Track exact attribute spans with quote-aware scanning, not tag reserialization or
a global regex. Deploy requires all authoring markers to be extracted.
Nested regions, malformed nesting, unsupported void/raw-text
regions, and directives in new content are errors. Existing IDs retain content;
labels/routes are reindexed on dev saves, cross-route/layout/partial IDs are shared.
Use the in-process LockContent mutex through local writes and publication; acquire
it before Server.reloadMu. No local content filesystem lock; multiple dev processes on
one project are unsupported. No legacy lock cleanup or automatic migration.
On every relevant save, archive unreferenced IDs and restore reintroduced IDs,
preserving content and metadata. Log successful snip updates. Persist
flea/runtime/.snips-pending.json with before/after
templates and content before saving content, then templates. Replay on restart;
stop if independently edited files conflict. Status and deploy never extract.
Deploy refuses unextracted regions. Do not ignore recovery records.

Reconstruct rich HTML with an allowlist using x/net/html.
Allow spans and escaped class attributes on supported content tags. Preserve them
through browser sanitization and inline editing; keep event handlers and arbitrary
inline CSS blocked. Classes are developer-controlled styling, not executable markup.
Never compile JSON into template source. Only sanitized values become template.HTML through richFlea at
execution; require element-text contexts, reject attributes/raw-text contexts.
Script/style source remains untouched. Block scripts, styles, SVG/MathML, forms,
embeds, images, arbitrary attributes, and unsafe links. Page metadata and data.*
remain escaped. Snip JSON is limited to 5 MiB and each content value to 1 MiB.

Use internal/sitefiles as the shared exclusion policy for loading, reloads,
Git snapshots, and archives. Exclude dotpaths, node_modules, and project metadata
(README.md, AGENTS.md, ARCHITECTURE.md, spec.md, next-steps.md, checklist.md, LICENSE, LICENSE.md, go.mod, go.sum).
Include only site/, mailers/, .flea.config.json and approved flea/ content, data,
media and icon files in release snapshots. The config is the sole dotfile exception.
Private runtime and skills never enter archives. Keep the project-root read boundary
separate from the site/ template boundary: logical template names remain site-relative,
while extraction journal edit paths begin with site/. No old-directory fallback or
automatic migration. Protocol 23 prevents old servers receiving this structure.

`flea new` initializes Git on main and commits the generated files as
`Initial site`, respecting normal identity, hooks and signing. Commit failure
keeps the site with recovery instructions; no remote is configured automatically.
Default creation is minimal: layout, navigation, index, styles/logo, error pages,
admin/email support, guidance and blank configuration. `--starter`/`-s` selects
the complete example site. Keep minimal overrides in templates/minimal/; they
are embedded separately and never copied as a public directory. Navigation links
live in site/_navigation.html, shared by desktop and native mobile popover.
Keep both Admin links labelled Admin and wrapped in `{#unless app.current_admin}`;
signed-in owners use the existing toolbar instead.
`flea reset` resets all local files except .git after explicit confirmation, under
the dev/pull lock. Validate the fresh tree first, then move originals into a private
sibling backup. No automatic commit, deployment, remote purge or clean alias.
Refuse non-sites, filesystem/home roots, linked site roots and Git worktrees.
Support --skip-git before or after the site name for existing repositories.
If Git is unavailable, still create the site and explain how to initialize Git
before deployment. Embed the starter .gitignore explicitly; retain deployment
settings and icon caches in version control by default.

- `main.layout.html` is the optional global wrapper; `{slot}` inserts the page.
- A leading `<!-- flea` comment with `layout: /admin.layout.html` on its own
  line and a closing `-->` selects a custom layout. Use one block before markup,
  site-relative *.layout.html paths, and only supported key:value settings.
  Reject unknown/duplicate keys, missing/unsafe layouts, nested layouts and settings
  in partials. Layouts support {slot} and relative partials, cannot be included as
  partials, and stay private. _500.html remains standalone. No YAML dependency.
- The starter's header and footer live directly in `main.layout.html`. Do not
  generate separate `_header.html` or `_footer.html` files by default.
- The starter includes a welcome page, contact/contact-thank-you pages, and a team
  multisnip example with six fictional members and neutral replaceable portraits.
  Contact delivery requires apps.contact with mount and email.to. The pages live
  at site/contact/index.html and site/contact/thank-you.html; apps are disabled in
  the starter config until a recipient is chosen.
  Navigation uses Team and Contact; keep /test as an unlinked example. Seed
  editable headings/body copy on Home, Team, Contact and Thank you, plus the
  shared footer. Keep snips outside form controls and navigation. Contact uses
  a responsive two-column introduction/form layout that stacks on narrow screens.
  Its logo is the selected
  option C (Skip): a happy side-profile flea with a folded jumping leg, in
  slate-800 rounded line art. templates/site/images/mark.svg is the shipped logo,
  shared by the header and favicon. Keep the selected design unchanged.
- Optional partials are underscore-prefixed HTML files anywhere under site/.
  There is no special `_partials/` directory.
- `{render footer}` resolves `_footer.html` beside the calling template.
  `{render /foo/bar/footer}` resolves `foo/bar/_footer.html`.
  Paths are unquoted and omit the filename's underscore and `.html` extension.
- Relative paths, including `./` and `../`, resolve from the calling template,
  including inside partials. Paths cannot escape site/. Missing and recursive
  includes fail validation, including in unused partials.
- Root `main.layout.html`, `_404.html`, and `_500.html` are reserved, not partials.
  `_404.html` uses the global layout; write it as page content for `{slot}`.
  `_500.html` remains standalone so error handling does not depend on the layout.
- Data uses dotted paths such as `{app.site_name}` and `{data.home.title}`.
  `{csrf}` renders a whole hidden `_csrf` input on public pages, not a raw token
  value. Include it in Contact forms. Keep it in ordinary element content;
  attributes, textarea and foreign/raw-text contexts are unsupported. Comments
  and script/style contents remain inert. Validation and non-public renderers
  generate no visitor token. Do not replace admin CSRF controls.
  App values retain their envelope: contact.data.name, contact.invalid.name,
  contact.error and contact.flash.notice. Never compile them as template source.
  Conditions use `{#if app.email}...{/if}`; loops use
  `{#each data.home.items as item}...{/each}`. Includes inherit loop aliases.
  Support `{#unless dotted.path}...{/unless}` with inverted if truthiness.
  Support `{:else}` in if/unless/each and `{:else if dotted.path}` in if/unless.
  Each fallbacks run for empty/missing arrays and restore any outer shadowed
  alias; no current item exists there. Reject duplicate else or branches after
  else, and retain dotted-path-only conditions and contextual escaping.
- `{snip a7Km2pQ9xB}` renders sanitized owner HTML. Legacy `{snip notice}`
  renders the escaped text field of a named snip. Preserve
  contextual escaping through Go's html/template; never reinterpret JSON as
  template code. Script and style elements do not support interpolation.
