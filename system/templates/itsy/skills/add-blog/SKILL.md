---
name: add-blog
description: Add, enable or customize ItsyKit's built-in Blog on an existing ItsyKit site. Use for requests such as "add a blog", "let's create a blog", "set up a blog", or authoring blog posts and templates. Covers app configuration, post JSON, available template values and the admin editor. Does not apply to a static article page or a full-site redesign.
---

# Add an ItsyKit Blog

Read the site's [README](../../../README.md) and [AGENTS](../../../AGENTS.md).
Inspect Git status, configuration, templates and content before editing. Preserve
custom layouts, navigation, other apps and existing content. Work within the
requested scope; this skill does not authorize publishing posts or deploying.
Use the site's Prewind conventions and existing styles. Adding a Blog does not
require a full-site design interview.

## Enable and connect

Blog requires ItsyKit 0.15.0 or newer and a matching server. Check `itsy version`;
do not assume the public download channel includes Blog yet.

Merge this fragment into `itsy.config.json`, preserving all other fields:

```json
{"apps":{"blog":{"mount":"/blog"}}}
```

Use an existing configured mount when present. Blog owns its entire URL subtree:
do not create `site/blog.html`, ordinary pages or assets at that mount. The public
routes `/blog` and `/blog/slug` use **private** `site/blog/_index.html` and
`site/blog/_post.html`. For `/journal`, use `site/journal/` and the same filenames.

Fresh matching-version sites include these templates and the authenticated
`site/admin/blog.html` and `site/admin/blog-edit.html`. For an older site, generate
a disposable matching-version site with `itsy new NAME --skip-git` outside the
working project. Copy only the missing Blog/admin templates and incorporate their
Blog CSS; the matching ItsyKit runtime supplies `/admin/assets/blog.js`. Preserve
the template's script reference, the site's layout and existing admin
authentication/CSRF/version controls. Never reset the existing site or
replace its entire layout/styles. Add the listing link to the existing navigation
(often `site/_navigation.html`). `/admin/blog` is the admin entry point; keep the private admin templates. `/admin` leads signed-in admins to Blog
when enabled. Public navigation uses an ordinary `/admin` link.

Restart the site's dev server after app configuration changes.

## Public template values

| Value | Meaning |
| --- | --- |
| `blog.data.url` | Listing URL, respecting the configured mount |
| `blog.data.posts` | Published posts, newest publication date first |
| `blog.data.post` | Current published post on its route; empty elsewhere |
| `post.title`, `post.slug`, `post.excerpt`, `post.url` | Listing loop values |
| `post.published_at`, `post.published_label` | RFC3339 date and readable label |
| `blog.data.post.cover` | Optional photo object: `src`, `alt`, `width`, `height` |
| `blog.data.post.body` | Sanitized rich HTML, with the exact directive below |
| `page.title`, `page.description` | Use post title/excerpt; template comments cannot override |

Listing template:

```html
<section class="container py-xl">
  <h1>Blog</h1>
  {#each blog.data.posts as post}
    <article>
      <h2><a href="{post.url}">{post.title}</a></h2>
      <time datetime="{post.published_at}">{post.published_label}</time>
      <p>{post.excerpt}</p>
    </article>
  {:else}<p>No posts yet.</p>
  {/each}
</section>
```

Post template:

```html
<article class="container py-xl">
  <a href="{blog.data.url}">Back to blog</a>
  <h1>{blog.data.post.title}</h1>
  {#if blog.data.post.cover}
    <img src="{blog.data.post.cover.src}" alt="{blog.data.post.cover.alt}"
         width="{blog.data.post.cover.width}" height="{blog.data.post.cover.height}">
  {/if}
  <div class="blog-body">{blog.data.post.body}</div>
</article>
```

Only exact `{blog.data.post.body}` renders rich HTML, in ordinary HTML element
content. Never use it in attributes, scripts, textarea or SVG, or replace it with
an alias such as `{post.body}`. Other values are escaped; JSON/body text is never
template source. Drafts are absent from public data even for signed-in admins.
Use `app.*` for public site globals, not `site.*`.

## Author posts

Admins sign in at `/admin`, then use `/admin/blog`: New post, Save draft, explicit
Publish/Unpublish. Saving a published post updates it immediately. Keep posts as
drafts unless publication is requested. A future date does not schedule publishing.

Agents may edit the top-level `blog` array in `itsy/snips.json`. Merge records,
preserving `pages`, `shared`, `archive`, `media`, videos and other posts. For a new
post generate a unique, secure-random ten-character case-sensitive alphanumeric
ID; check existing IDs and keep it stable through edits. Ordinary pages do not use generated content IDs. The following is a format example, not a
record to overwrite or an ID to reuse:

```json
{
  "id": "a7Km2pQ9xB",
  "title": "Welcome",
  "slug": "welcome",
  "excerpt": "A short introduction.",
  "body": "<p>Welcome to our website.</p>",
  "status": "draft"
}
```

Title is required (200 characters maximum), excerpt at most 500, and slug at most
100 ASCII characters: lowercase letters/digits with single separating hyphens,
unique across drafts and published posts. Status is `draft` or `published`;
published source records require RFC3339 `published_at`. Admin first Publish
assigns the date if blank. Slug changes create no redirect. Bodies support basic
rich formatting/links, no inline photos, scripts or forms. Limits: 200 posts,
1 MiB per body, 5 MiB for the content JSON.

Optional `cover` is an existing ID in the content's `media` map, not a file path or
URL. Reuse the admin library/upload flow instead of inventing media records.
Photos, including draft covers, are public assets; draft text remains private.
Unpublish retains the record; omission from source never deletes a live post.

## Verify and hand off

Run `itsy status`. In local dev, check the listing, a published post, empty state,
optional cover and navigation. Verify a draft URL returns 404 and the admin editor
still uses normal login. If testing publication, use disposable local content.
Review the JSON and diff, and stop any task-owned dev server before handoff.

Normal deploy preserves existing live posts by ID and seeds new IDs. When an
existing live post must be changed in source, pull the intended destination first
(with dev stopped), retain other live content, and use reviewed
`itsy deploy --with-content` only when deployment is authorized. The flag includes
Blog and media; whole-record conflicts require resolution. A plain Git merge and
normal deploy do not replace existing live posts. Report setup and checks, and
whether any posts were actually published.
