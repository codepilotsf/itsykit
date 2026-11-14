# Page metadata in HTML comments

## Requested outcome and scope

The owner wants page titles and descriptions to be set at the top of a page in
its existing `<!-- itsy ... -->` settings comment. Implement framework changes
in a separate thread. This docs thread is limited to the website and handoffs.
The existing docs menu, introduction, simplified install flow and directory tree
are independent website changes and should be preserved.

The framework implementation initially added in commit `5e50959` has been
reverted from this branch at the owner's request. That commit is available as
reference, not as an approved implementation or precedence decision. Do not
cherry-pick the entire commit: it also contains website work and an old handoff.

## Current behavior

- `system/internal/site/settings.go` parses one leading ItsyWeb settings comment.
  Only `layout` is currently supported. The existing legacy comment spelling
  remains supported; this task does not authorize adding more aliases or YAML.
- `system/internal/site/project.go` supplies `page.title` and `page.description`.
  Route records in the selected content JSON supply saved metadata. Without it,
  the title comes from the route and the description is empty.
- Public layouts use those escaped values in `<title>` and description metadata.
- On-page editing already saves title/description fields in the route record.
- New sites have a saved Home title and description in their generated content.
  This matters: treating comments only as defaults can make an HTML title appear
  to have no effect on a newly generated homepage.

## Example syntax

```html
<!-- itsy
title: Visit the bookshop
description: Find our opening hours and plan a visit to Little Bookshop.
-->
<section>
  <h1>Visit the bookshop</h1>
</section>
```

Support `layout` alongside the new keys in the same block. Retain the small
single-line key/value format, duplicate and unknown-key diagnostics, safe layout
paths, private templates and contextual escaping. Do not compile metadata values
as template source or extract them into content JSON merely by loading a page.

## Resolve precedence before implementation

The owner has requested comment-based authoring but has not explicitly chosen
how it interacts with saved metadata. Ask one focused question if the other
thread has no additional direction. Two concrete choices:

1. Comments provide defaults; nonempty saved fields override them independently.
   This preserves on-page edits. Address the generated Home metadata so the
   documented authoring flow behaves as expected, and explain how to return to
   a comment default after a saved edit.
2. Explicit comment fields take precedence over saved metadata. HTML edits take
   effect directly, but the editor must clearly handle fields controlled by code.

The initial attempt used option 1. Its HTTP smoke test confirmed both comment
metadata on a new `/visit` page and preservation of the generated Home title.
That is evidence of behavior, not owner approval of the tradeoff. Decide whether
empty values mean fallback or an intentional blank, and keep editor display,
Save/Cancel, deployment content preservation and reload behavior consistent.

## Relevant paths and checks

Read root and `system/AGENTS.md`, templates/content and architecture guidance.
Inspect Git status and preserve concurrent work. Use one agent unless authorized.

- `system/internal/site/settings.go`, `project.go`, `settings_test.go`, and
  `layout_compat_test.go`: parser, rendering and regression coverage.
- `system/internal/editor/assets.go` and `system/internal/editstore/store.go`:
  editor context and saved metadata. Inspect before changing ownership behavior.
- `system/templates/itsy/snips.json`, `system/templates/minimal/snips.json`, and their
  page templates: existing generated metadata and homepage defaults.
- `system/README.md`, `system/docs/agents/templates.md`, and the generated
  authoring reference: supported syntax and precedence documentation.
- `website/site/docs/templates.html`, `website/README.md`, and the website's
  authoring reference: update only when the supported behavior makes it accurate.
  The current website docs retain the existing JSON authoring instructions.

Test actual HTTP rendering with a layout, nested routes, reloads, missing fields,
chosen precedence and empty-value semantics, escaped HTML/quotes, literal braces
and colons, malformed/duplicate settings, and extraction preservation without
unrequested JSON writes. Keep current and existing legacy layout behavior.
Verify editor behavior if affected. Generate a fresh temporary site and exercise
the documented flow, including the homepage with its seeded metadata.

A local `system/bin/itsy` was built during the initial attempt and may still
contain the reverted feature. Rebuild from the other thread's actual candidate
before testing; do not infer source support from that existing binary.

The initial checks passed for site/CLI packages, vet, the contributor build,
and a fresh-site HTTP smoke. A broader run hit download fixture permission
failures under shell umask `0002`; the host package passed with command-local
`umask 0022`. Follow current repository verification requirements on the new
candidate. Do not weaken permission checks or change the machine-wide umask.

Deployment and releases need separate authorization. A source implementation or
local build does not make this syntax available in the published installer.

## Draft website section for the supported implementation

This is the fuller documentation prepared in the original docs task. Adjust it
to match the agreed precedence and released behavior before applying it:

```html
<section id="page-details">
  <h2>Page titles and descriptions</h2>
<p>Give each page a title for the browser tab and a description for search results and link previews. Put them in an ItsyWeb page settings comment at the very top of the HTML file, before any markup:</p>
<pre tabindex="0" aria-label="Page settings in site/visit.html"><code class="language-markup">&lt;!-- itsy
title: Visit the bookshop
description: Find our opening hours and plan a visit to Little Bookshop.
--&gt;
&lt;section&gt;
  &lt;h1&gt;Visit the bookshop&lt;/h1&gt;
&lt;/section&gt;</code></pre>
<p>Write each setting as a single line of plain text, without quotation marks. You can include <code>layout</code> in the same comment when the page uses a custom layout.</p>
<p>In <code>site/main.layout.html</code>, use these values in the document head:</p>
<pre tabindex="0" aria-label="Use page metadata in a layout"><code class="language-markup">&lt;title&gt;&#123;page.title&#125; | &#123;app.site_name&#125;&lt;/title&gt;
&lt;meta name="description" content="&#123;page.description&#125;"&gt;</code></pre>
<p>The comment supplies the page's defaults. Titles and descriptions saved through on-page editing live in <code>itsy/snips.json</code> and take precedence when nonempty. If no title is set, ItsyWeb uses the page's URL to make one.</p>
<p>Ordinary page text can stay in HTML. Use <a href="/docs/editing">on-page editing</a> for content an admin should change.</p>
</section>
```
