# Blog cover photos

Ordinary page images are static assets under `site/`. Blog alone uses the saved
photo library. Cover uploads accept JPEG/PNG up to 8 MiB and 20 megapixels, resize
to at most 2048 pixels, strip metadata, and write immutable content-addressed files
under `itsy/media/`. The library allows 100 records and 64 MiB retained files.

`itsy/snips.json` has an optional `media` map keyed by ten-character alphanumeric
IDs. Each photo records `file`, `alt`, `name`, `width` and `height`.
Use the metadata returned by processing, never invented dimensions or filenames.
Blog posts reference the ID through `cover`; templates receive a public cover
object with src/alt/width/height. All library photos are public, including unused
covers and draft covers. Draft post bodies remain private. Removing a cover keeps
the photo. Normal deploy, pull, backup/restore and rollback preserve library data;
no automatic deletion or pruning exists. See [Blog](blog.md).
