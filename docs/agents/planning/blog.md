# Minimal blog app

Status: implementation authorized October 13, 2026; local implementation and verification complete.
Deployment and publication remain separate tasks.
Resolve technical details against current code; ask further questions only for
a material product, architecture, security or authorization decision.

## Objective

A small optional Flea blog with a documented JSON format and ordinary template
value tokens. CodePilot Hub agents can author it through site files. A basic
admin interface is included for independent Flea users, reusing existing login,
rich text and photo selection/upload rather than introducing another editor.

Hostname indexing shipped in 0.19.0; see [verification](../release-0190-verification.md).
Earlier work still has separate follow-ups: older site toolbar updates and
public 0.14.0 publication. This blog plan does
not complete or authorize those tasks.

## Confirmed scope

- Include the basic admin interface in the first version: post list, New/Edit,
  existing login, rich text and photo selection. Keep the rest of the first
  version minimal.
- New posts start as drafts. Publish makes a post public; Unpublish hides it
  from public routes and listings. Saving a published post updates it immediately,
  without a separate pending revision or approval state. Cancel discards unsaved
  edits. Hub owns any approval workflow outside Flea.
- Post fields are title, slug, excerpt, rich body, publication date and one
  optional cover photo. Defer inline body photos.

## First version

- A public post list and one page per published post, with custom site layouts.
- Title, slug, short excerpt, rich body, publication date and optional cover photo.
- Draft and published states. Draft posts never appear on public routes or lists.
- A small admin post list with New/Edit, the existing rich editor, cover-photo
  selection, explicit Save/Cancel and Publish/Unpublish.
- Agent-authored JSON and browser edits use the same validated content model.
- No comments, tags/categories, search, scheduling, author accounts, revision
  history, external CMS, database dependency or frontend build step in this version.

Keep photos separate from rich body HTML initially, matching the existing editor
and sanitizer. Cover photos reuse managed media; inline body-photo placement is
deferred. Do not automatically delete posts or retained media. Unpublish provides
the initial way to remove a post from public view.

On first Publish, assign the publication timestamp server-side. Ordinary edits
and republishing retain it by default; define any explicit date editing during
contract design. Public list order is newest publication first. Do not expose
draft text through public URLs or template data. An unpublished post's
public URL returns the normal site 404; its admin edit view remains available.

## Implementation decisions

Implementation uses the optional `blog` array in `flea/snips.json`, reusing one
atomic content snapshot rather than adding another persistence transaction.
`apps.blog.mount` selects private `_index.html`/`_post.html` templates beneath the
mount. Drafts remain private; photos are public website assets. Existing
`--with-snips` explicitly synchronizes blog records and media. Whole-post
conflicts require review; omitted IDs are retained. See [Blog](../../blog.md)
for the finalized supported format and usage.

Verification evidence and remaining release limits are in
[Blog verification](../blog-verification.md).
