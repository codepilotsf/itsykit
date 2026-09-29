# Photos

Developers designate editable images with a bare `snip` attribute. Signed-in
owners leave **Editing** on, then click an image to replace its JPEG/PNG or change its
alternative text. **Done** previews the draft; the topbar **Save changes** saves
it with the page. **Cancel** restores the original. There is no Photos topbar mode.
Images stay separate from rich text; uploads are public website assets, not private
file storage.

The library accepts files up to 8 MiB and 20 megapixels. Go corrects JPEG EXIF
orientation, resizes to at most 2048 pixels on the longest edge and re-encodes to
remove metadata. PNG transparency is preserved. No Node, image-service account or
external image executable is required. HEIC, WebP, GIF, SVG and video uploads are
not supported. The library allows 100 photo records and 64 MiB of stored files.

## Templates

Start with a local image under `site/`:

```html
<img snip src="/images/team.jpg" alt="Our team" class="rounded-lg"
     loading="lazy" width="640" height="480">
```

Dev imports the image and rewrites only the editable values and authoring marker:

```html
<img src="{snip a7Km2pQ9xB.src}" alt="{snip a7Km2pQ9xB.alt}" class="rounded-lg"
     loading="lazy" width="640" height="480">
```

The generated ID identifies a `type: "image"` entry with `image` metadata beside
ordinary text snips in the page's `snips` object. Source and alt text are stored
there, not duplicated in the template. The ID survives replacements. Classes,
dimensions, loading behavior and link destinations remain developer-owned.
Empty alt text is appropriate for decorative images. Images without a snip
designation are not editable. Nested text/image/video snips are errors, as are
competing sources such as srcset or picture wrappers. External seed URLs are not
downloaded. SVGs may remain ordinary developer-managed assets, not image snips.

For a multisnip, add a named image alongside its text fields:

```html
<article multisnip="team">
  <img snip="photo" src="/images/joe.jpg" alt="Joe Appleseed"
       width="160" height="160">
  <h2 snip="name">Joe Appleseed</h2>
</article>
```

Dev imports the local JPEG/PNG and extracts its stable library ID into the item's
`fields.photo`. Omit src for an initially empty photo selection. Seed image paths
are site-root-relative; external URLs are not downloaded. The extracted image
uses `{snip.photo.src}` and `{snip.photo.alt}`, with generated binding attributes.
Keep those bindings in the template. Each group still needs at least one text field.
Click the item's photo during Edit to select a library image or upload a new
JPEG/PNG (up to 8 MiB), with alternative text. Done previews the choice locally;
Save changes processes and publishes the photo and item together. Cancel discards
the draft without uploading. Saved uploads are available in the library for other
collection items. Choosing a library photo does not change its metadata.
The picker displays library thumbnails, a No photo choice and Upload photo.
Selection shows a larger preview and its description. New-upload alt text is
editable; saved descriptions are read-only. Closing the picker without Done
discards its new preview without changing the page draft.

## Files and deployment

Standalone image metadata lives in typed snip entries in `flea/snips.json`;
multisnip photo choices use its `media` map. Local processed files live in
`flea/media/`, named by their SHA-256 hash. Commit both. Original developer seed images
are not removed. Production uploads live in the site's managed content directory,
outside immutable code releases; only validated snapshot bytes are publicly served.

Normal deployment preserves live photos and seeds new ones. `flea pull` and
`flea pull --snips-only` retrieve photo metadata and referenced files together.
`--with-snips` includes reviewed local photo changes, with explicit conflicts.
Backups include live photo bytes, and restore validates them before activation.

Replacements retain previous immutable files for recovery safety. There is no
automatic deletion/pruning, crop/focal-point tool or responsive derivative set yet.
Reaching the storage limit requires developer review and archiving unused files;
do not delete files referenced by retained releases/backups.
