# ItsyWeb site requirements

Read README.md and AGENTS.md. Preserve custom layouts, copy, assets and contracts.
Use plain HTML/JSON, no frontend build step. `itsy new` is minimal; `--starter`
adds ordinary HTML demo pages. Never reset or deploy without explicit authority.
`itsy dev` loads templates and watches source changes without extracting content.
Stop task-owned servers at handoff. New sites use local email previews without
real delivery; production ignores this option. LAN mode retains exact host,
Origin and CSRF checks; never forward the development port.

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

Contact keeps its app configuration, public `{csrf}` input and thank-you route.
Private admin forms keep `{admin.csrf}`. `/admin` leads to Blog after sign-in when
enabled; no public page toolbar or admin CSRF meta is needed. Custom managed apps
retain confined packages, separate sessions and private runtime state. SEO and
redirect config remain developer-owned; see README. Runtime protocol 30 is required.

## Styling: Prewind first

Follow Prewind's global-first approach, not a parallel component CSS system.
Use existing shared defaults and theme tokens for typography, colors, links,
buttons and forms. Use supported Prewind utilities for layout, spacing, sizing,
alignment and responsive adjustments. Check the loaded Prewind version and its
documentation before choosing classes; do not assume Tailwind syntax exists.

- Before adding custom CSS, check whether existing defaults or Prewind utilities
  already express the design. Do not duplicate utilities with custom selectors.
- Do not introduce BEM naming (`block__element`, `block--modifier`), page-specific
  styling frameworks or parallel utility systems. Preserve existing custom work;
  this rule does not authorize an unrelated rewrite of existing classes.
- Custom CSS is for genuinely unsupported behavior or visual details. Before
  introducing an exception, explain what Prewind cannot reasonably express and
  obtain the user's approval. Put approved shared rules in the existing CSS
  architecture and layers; do not add an overriding layer to bypass utilities.
- Preserve useful boilerplate behavior, including sticky/fixed headers,
  translucent backgrounds, backdrop blur, mobile navigation and transitions,
  unless the user explicitly requests a change. A screenshot artifact is not
  a reason to change the actual site's behavior.
- Pass these constraints to any authorized implementation worker. Delegation
  does not relax them; the lead must review the resulting markup and CSS.

Before expanding an approved sample into the full site, review its markup and
CSS against these rules. Visual approval alone is not approval of a different
styling architecture. Before calling work complete, inspect every added or
changed CSS selector, remove utility duplication, and report any remaining
custom CSS with its reason and approval. Verify the preserved layout behavior
in the browser, not only in screenshots.

Styles are developer-managed in `site/css/theme.css` using Prewind. Ordinary
`itsy deploy` publishes committed CSS. There is no browser Theme editor or CSS
synchronization; `itsy pull` retrieves Blog posts and cover photos only.

## Optional Blog

Enable `apps.blog.mount`, use private `_index.html`/`_post.html` templates under
that mount and keep private admin templates. Read `itsy/skills/add-blog/SKILL.md`.
Posts and cover photos live only in `itsy/snips.json` (`blog` and `media`). Keep
stable IDs and exact saved metadata. Pull before editing live records locally;
normal deploy preserves live content, `--with-content` explicitly synchronizes it.
