# Templates and content

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).

`site/` owns public HTML and assets; root `mailers/` owns private email templates.
Config is `itsy.config.json`. `itsy/data/*.json` contains developer-authored objects.
`itsy/snips.json` retains only Blog posts and cover metadata. `itsy/media/` holds
immutable library photos; `itsy/runtime/` holds private sessions. Never serve or
archive runtime state, secrets or agent guidance. Keep `.agents/` and installed
skills in Git. Impeccable stays globally installed, never copied into a site.

Existing `.flea.config.json`, `flea/` and `<!-- flea` spellings remain supported;
prefer Itsy names when present and write only the selected paths. No new alias or
migration is authorized. Protocol 30 requires a matching server. Old page-CMS
JSON/markup/named content and pending extraction journals fail closed.

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

`index.html` maps to `/`; `about.html` and `about/index.html` both map to
`/about`, and cannot coexist. `.html` URLs and non-root trailing slashes are 404
unless explicitly configured as redirects. Underscore/dot paths and all
`*.layout.html` files are private. `admin/` is reserved for authenticated templates.

`main.layout.html` optionally wraps pages using `{slot}`. Layouts may render
partials but cannot select another layout or be included as partials. `_404.html`
uses the main layout; `_500.html` stays standalone. Preserve custom site layouts.
`{render footer}` includes adjacent `_footer.html`; `{render /foo/bar/footer}`
includes `site/foo/bar/_footer.html`. Paths are unquoted, omit underscore/extension
and resolve relative to the calling template. Reject traversal, missing and
recursive includes, including unused partials.

Values use `{app.site_name}`, `{data.home.title}` or loop aliases. No `site.*`
alias. Dotted-path-only conditions use `{#if app.email}...{:else}...{/if}` or
`#unless`; `:else if` is supported. Loops use
`{#each data.home.items as item}{item.name}{:else}Empty{/each}`. Missing arrays
use the fallback; non-arrays fail. Nested loops restore shadowed outer aliases.
No functions, filters, comparisons or arbitrary expressions. All ordinary values
retain Go contextual escaping. Only the exact Blog body directive renders
sanitized rich HTML, and only in ordinary element content.

Comments and complete script/style elements stay inert except the leading settings
block. Balanced ordinary literal brace groups and JSON stay literal. Recognized
commands/dotted paths retain malformed-syntax diagnostics. Encode braces as
`&#123;` and `&#125;` when showing exact syntax. JSON and Blog bodies are never
recursively compiled. Retired page directives are recognized only to reject them.

Public forms keep `{csrf}` as a whole hidden input in ordinary element content.
Contact uses `apps.contact`, a mount-relative thank-you route and its existing
email/form contract. Admin forms separately use `{admin.csrf}` and exact Origin.
Public layouts have ordinary `/admin` links and no authenticated page state or
editor mount. Blog uses private admin templates and `/admin/blog`; see [Blog](../blog.md).

`itsy new` defaults to minimal ordinary HTML with Home comment metadata and `{}`
Blog storage. `--starter` adds static demo pages and six fictional team members.
Both include admin/Blog/mail templates and authored guidance. Refuse nonempty
creation targets, hidden entries and symlinks. Git initialization, initial commit,
`--skip-git` and explicit reset/recovery safeguards remain unchanged. Dev loads and
watches templates read-only apart from explicit icon caching, site-name/admin setup
and session storage. There is no extraction or content reindexing transaction.
