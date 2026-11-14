# Multisnips

Write example items on their page:

```html
<div class="team-grid">
  <article multisnip="team">
    <h2 snip="name">Joe Appleseed</h2>
    <p snip="title">Founder</p>
    <div snip="bio"><p>Joe helps turn ideas into websites.</p></div>
  </article>
</div>
```

`flea dev` moves the seed into `pages["/team"].multisnips.team` in `itsy/snips.json`
(on the homepage, the route is `/`). The value is a direct ordered array:

```json
"multisnips": {
  "team": [
    {
      "id": "m4Jp8rT2vN",
      "fields": {
        "name": "Joe Appleseed",
        "title": "Founder",
        "bio": "<p>Joe helps turn ideas into websites.</p>"
      }
    }
  ]
}
```

The template becomes:

```html
{#multisnip team}
  <article>
    <h2>{snip.name}</h2>
    <p>{snip.title}</p>
    <div>{snip.bio}</div>
  </article>
{/multisnip}
```

ItsyWeb preserves classes and surrounding markup. To seed several existing items,
repeat adjacent same-name authoring blocks with identical markup outside their
editable fields. Dev extracts them in source order and keeps one loop template.
Different classes, fixed text or structure are rejected rather than discarded.
An existing stored array is preserved, including an intentionally empty array;
use the admin editor to add items after extraction.
Names use lowercase letters, digits and underscores,
starting with a letter. Keep one root element and one complete element per field.

The template defines fields even when no items exist. Div/section/article fields
accept rich blocks; headings, paragraphs and spans accept inline formatting.
Text and photo-library selection fields are supported; see [photos](photos.md).
Video fields, field options and custom field types are not. Do not put multisnips in layouts, partials, loops, navigation,
forms, or inert templates. They are page-local, with no sharing or visibility toggle.

With Editing on, arrange items directly on the page. There is no Collections button.
Each item gets an outline, drag handle and remove X. Each collection has its own
**+ Add item** tile, with no group dropdown or sidebar. Drag with mouse or touch,
or focus a handle and use arrow keys (Home/End jump to either end). Escape cancels
a drag. Text stays editable in place; click a photo to choose a library image or
upload a JPEG/PNG. New uploads publish only with Save changes and become reusable
library photos. Cancel discards their private previews.
Empty photos show a neutral portrait and Add photo. Grip/remove controls appear
on hover or keyboard focus, and stay visible on touch. Dragging lifts a private
visual copy with a shadow; neighboring cards move around a dashed landing space.
Topbar Save changes/Cancel applies everything
together. Removing the last item leaves its Add tile; saving preserves the empty
array and deployments do not reseed it.

Limits: 20 groups per page, 200 items per group, 30 fields per item, 64 KiB per
field, and the existing 5 MiB content-file limit. IDs remain stable on reorder;
new items receive new IDs. Rich content is sanitized on rendering and saving.

Normal deployment preserves complete live arrays. Explicit synchronization treats
each array as one value: conflicting membership or order requires a choice. This
is deliberately conservative rather than silently guessing an order. Removing a
field or template retains its old content. That is not revision history; removing
an item and saving currently deletes it, so backups remain a v1 requirement.

Run dev and commit the extracted files before deploying. All multisnip writes use
the existing version checks, content lock, validation, publication, and extraction
journal. Photos stay outside rich text. Standalone video snips are supported;
video fields inside multisnips remain deferred.
