# Blog

Blog is an optional built-in content application. It uses ItsyWeb's current admin
login, rich-text editor, photo library and atomic content snapshot. No database,
external service or frontend build is needed.

## Enable it

Fresh sites include private example templates at `site/blog/_index.html` and
`site/blog/_post.html`, plus `site/admin/blog.html` and `blog-edit.html`. Merge this
into `itsy.config.json`, preserving other apps and settings:

```json
{"apps":{"blog":{"mount":"/blog"}}}
```

The mount owns its entire URL subtree. Ordinary pages/assets cannot occupy it.
For another mount, move the two public example templates to that directory,
keeping their private names. Blog is disabled when its app entry is absent;
content remains stored. Older sites must copy the blog and admin templates and
blog CSS from a fresh site, preserving their custom layouts. Add an admin link to
`/admin/blog` or use the URL directly. `/admin` redirects signed-in admins there when Blog is enabled.

## Write and publish

Sign in at `/admin`, then open `/admin/blog`. Choose **New post**, enter a title,
URL slug, excerpt and body, then **Save**. New posts always save as drafts. **Publish**
makes a post public. **Unpublish** makes it a draft again. Saving a published post
updates it immediately. Cancel leaves the saved version untouched. There is no
separate draft revision of a published post and no automatic deletion.

The optional cover uses a saved library photo or a JPEG/PNG upload, with the same
8 MiB/20 megapixel processing limits as [photos](photos.md). The body accepts basic
rich-text formatting and links; inline images are not supported. Photo files are
public assets, including covers attached to drafts. Draft text and metadata stay
private. Removing a cover retains its photo in the library and on disk.

Publication dates are RFC3339, such as `2026-10-13T14:00:00Z`. Leave the date blank
to assign it when first publishing. A future date does not schedule publication;
Publish takes effect immediately. The public list sorts dates newest first, using
stable IDs to break ties. Changing a slug changes the URL; no redirect is created.

## Agent and developer format

Posts live in the optional top-level `blog` array in `itsy/snips.json`, alongside the optional `media` cover-photo library. These are the only
supported fields; ordinary pages own their content and metadata in HTML.
The existing single-file transaction keeps posts and cover metadata together.

```json
{
  "blog": [
    {
      "id": "a7Km2pQ9xB",
      "title": "Welcome",
      "slug": "welcome",
      "excerpt": "A short introduction.",
      "body": "<p>Welcome to our website.</p>",
      "status": "published",
      "published_at": "2026-10-13T14:00:00Z"
    }
  ]
}
```

An ID is exactly ten case-sensitive alphanumeric characters and stays stable
through edits. Slugs are unique across drafts and published posts, at most 100
ASCII characters, using lowercase letters, digits and single separating hyphens.
Titles are required and at most 200 characters; excerpts at most 500. There are
at most 200 posts, each body at most 1 MiB, within the existing 5 MiB content-file
limit. Status is exactly `draft` or `published`. Published posts require a valid
date. Optional `cover` is an existing ID in the content file's `media` map, never
a filename or URL. See [photos](photos.md) for immutable files and metadata.

Never reuse an ID for a different post. To remove a public post, set its status to
draft. Omitting it from source does not delete its live record on deployment.

## Templates

The developer owns layouts and presentation. Blog data contains published posts
only, including when an admin views a public page. A draft slug returns 404.

```html
{#each blog.data.posts as post}
  <article>
    <h2><a href="{post.url}">{post.title}</a></h2>
    <time datetime="{post.published_at}">{post.published_label}</time>
    <p>{post.excerpt}</p>
  </article>
{:else}<p>No posts yet.</p>
{/each}
```

The current post is `blog.data.post` on a published post route; elsewhere it is
empty. It provides the same fields, plus body and an optional cover object with
`src`, `alt`, `width` and `height`. `blog.data.url` is the listing URL. Page title
and description come from the post title and excerpt, overriding template comment
metadata. No ordinary-page saved metadata or Page details editor exists. Mounted pages retain the 2 MiB app rendering limit.

```html
<h1>{blog.data.post.title}</h1>
{#if blog.data.post.cover}
<img src="{blog.data.post.cover.src}" alt="{blog.data.post.cover.alt}"
     width="{blog.data.post.cover.width}" height="{blog.data.post.cover.height}">
{/if}
<div>{blog.data.post.body}</div>
```

Only the exact `{blog.data.post.body}` directive renders sanitized rich HTML.
It requires ordinary HTML element content, never an attribute, script, textarea,
SVG or other raw-text context. Other value tokens remain escaped. Body contents
are never compiled as template source, so braces inside a post are ordinary text.

## Deploy, pull and recovery

Normal `itsy deploy` preserves live posts by ID and seeds new source-only IDs.
Use `itsy pull` before editing existing live posts in Git. Deploy reviewed changes
with `itsy deploy --with-content`; this flag includes Blog content and cover
metadata. Explicit synchronization compares each post as a whole record. Changes
on both sides conflict and require a choice; noninteractive conflicts stop. An
omitted local ID is never a deletion request. `itsy status` counts blog differences.

Backups include live posts and photos. Restore restores that content; rollback
keeps current live posts while selecting earlier code. A missing/incompatible
blog template fails validation rather than discarding posts. Protocol 30 requires
a matching server for the new content and routing contract. There is no migration.

Comments, tags, categories, search, scheduling, revisions and inline body photos
are deferred.
