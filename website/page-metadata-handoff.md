# Page metadata in HTML comments

Status: the owner approved comment defaults on November 14, 2026. Framework
implementation belongs in a separate thread. This website task does not authorize
framework edits, deployment or releases.

## Current source and website

The inspected base is `4daf897`. `system/internal/site/settings.go` supports only
`layout` in one leading `<!-- itsy` settings comment, plus the existing legacy
`<!-- flea` spelling. `system/internal/site/project.go` reads nonempty saved
route metadata from the selected content JSON; otherwise titles come from the
route and descriptions are empty. Layout values are contextually escaped.

Keep the website's existing JSON metadata instructions until a supported release
makes comment authoring accurate. Preserve the independent docs menu,
introduction, simplified installation flow and directory tree. The affected docs
are `site/docs/templates.html#page-details`, `README.md` and
`.agents/skills/itsyweb-site/references/templates-content.md`, relative to
`website/`. The existing chapter ID already has a navigation entry.

Commit `5e50959` is reference only. Its framework changes were reverted at the
owner's request. Do not cherry-pick it wholesale: it mixes website work and an
old installer handoff with the initial implementation. Its tests establish
observed behavior, not approval of metadata precedence.

## Approved precedence and empty values

The owner selected **comment defaults, saved edits override independently**.
Nonempty saved title/description fields win. Clearing a saved field restores its
comment default; an empty comment title falls back to the route title and an
empty description stays empty. An empty value is fallback, not an instruction to
suppress a nonempty default. Preserve on-page edits and ensure unrelated content
saves do not silently turn displayed defaults into saved metadata overrides.

Address the generated Home metadata so adding a comment to a fresh homepage
visibly works, without discarding existing sites' saved edits. Explicit comment
fields taking precedence was offered and was not selected. Do not reopen the
approved decision without material new evidence.

The minimal scaffold currently seeds `title: "Home"` and an empty description;
the example starter seeds `title: "Home"` and a welcome description. Inspect both
before designing the fresh-homepage behavior. Do not infer approval from the
initial attempt's preservation of the seeded Home title.

## Framework handoff

Read root and `system/AGENTS.md`, templates/content, architecture and applicable
editor/persistence guidance. Inspect Git status and preserve concurrent work.
Use one agent in the implementation thread unless separately authorized.

Support `title`, `description` and `layout` in the same leading block:

```html
<!-- itsy
title: Visit the bookshop
description: Find our opening hours and plan a visit to Little Bookshop.
-->
<section>
  <h1>Visit the bookshop</h1>
</section>
```

Retain the small single-line key/value format, unknown/duplicate diagnostics,
safe layout paths, private templates, existing legacy layout behavior and
contextual escaping. Do not add aliases or YAML, compile values as template
source, or extract metadata into JSON merely by loading a page.

Framework paths, relative to `system/`:

- `internal/site/settings.go`, `project.go`, `settings_test.go` and
  `layout_compat_test.go`: parser, rendering and regression coverage.
- `internal/editor/assets.go` and `internal/editstore/store.go`: inspect editor
  context, saved metadata, Save/Cancel and ownership before changing behavior.
- `templates/itsy/snips.json`, `templates/minimal/snips.json` and their page
  templates: generated Home metadata and defaults.
- `README.md`, `docs/agents/templates.md` and
  `templates/.agents/skills/itsyweb-site/references/templates-content.md`:
  supported syntax and agreed precedence.

Exercise actual HTTP rendering with a layout and nested routes, reloads, missing
fields, agreed precedence and empty values, HTML/quotes, literal braces/colons,
malformed/duplicate settings, extraction preservation and no unintended JSON
writes. Keep current and legacy layout behavior. Verify editor display,
Save/Cancel, saved values after reload and deployment content preservation when
affected. Use isolated fixtures, not a live deployment.

Rebuild the CLI from the actual candidate before testing: an existing
`system/bin/itsy` may still contain the reverted feature. Generate fresh minimal
and example sites and exercise the documented flow, including Home with seeded
metadata. Follow current repository checks; the previous attempt's passes do
not validate the new candidate. A previous broader run encountered download
fixture permission failures with umask `0002`; a command-local `umask 0022`
passed the host package. Do not weaken checks or change the machine-wide umask.

Return the candidate commit, precedence/empty semantics, generated Home behavior,
editor behavior, actual verification and release availability to this docs
thread. A local build or source commit does not update the published installer.

## Website follow-through after support is available

Replace the short JSON-only `page-details` section with the example above and
explain that the comment goes before markup, values are single-line plain text
without quotation marks, and `layout` can share the block. Show layout usage:

```html
<title>{page.title} | {app.site_name}</title>
<meta name="description" content="{page.description}">
```

Encode braces as `&#123;` and `&#125;` in the website's HTML examples. Explain the
approved precedence, empty values, route-title fallback, fresh-homepage behavior
and how to return to a comment default after a saved edit, if applicable. Keep
ordinary page text in HTML and the link to `/docs/editing` for admin-managed
content. Update the README and installed authoring reference consistently.

Review examples and links against the actual supported implementation and
release. For public page changes, follow the website's status/preview checks.
Commit scoped website changes before any separately authorized deployment.
