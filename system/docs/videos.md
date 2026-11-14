# Videos

Developers designate a video with a bare `snip` attribute. Signed-in admins choose
**Edit**, then click that video to replace its optimized MP4, change its name or
replace/remove its JPEG/PNG poster. **Done** previews the draft; the topbar
**Save changes** publishes it with the page after validation. **Cancel** restores
the original. Invalid or stale saves retain the draft. There is no Videos topbar
mode and no admin-created media region without a developer designation.

## Preparing a file

- MP4 container, one H.264 video track, with optional AAC-LC audio.
- Enable **Web optimized / Fast start** when exporting.
- Use standard 8-bit H.264 Baseline, Main or High profile; 1080p or less is a
  practical export target. ItsyWeb accepts dimensions up to 8192 pixels per side.
- Limit each file to **25 MiB**. The library holds up to 50 entries and **64 MiB**
  of video files, including retained replacements.
- Export a regular, non-fragmented MP4 with no extra subtitle/data tracks.
  Add captions separately with native HTML `<track>` elements where needed.

ItsyWeb checks MP4 box bounds, sample offsets/timing, codec declarations and decoder
configuration. It does not decode every frame, strip embedded metadata, compress,
resize, transcode or repair files. Preview playback after uploading. MP4 alone is
not a codec: HEVC/H.265 files renamed to .mp4 are not supported. There is no WebM
fallback, adaptive streaming, FFmpeg, Node or external video-provider requirement.
H.264/AAC playback still depends on the visitor's browser and operating-system
codecs; some Linux installations lack them.

## Putting a video on a page

Place a local optimized MP4 under `site/` and designate it in the template:

```html
<video snip controls playsinline preload="metadata"
       src="/videos/intro.mp4" poster="/images/intro.jpg"></video>
```

Dev imports the files and generates the bindings, for example:

```html
<video controls playsinline preload="metadata"
       src="{snip a7Km2pQ9xB.src}" poster="{snip a7Km2pQ9xB.poster}"></video>
```

Use the generated ID, not a hand-invented one. Source/poster metadata belong to
the typed snip in JSON. Styling, dimensions and playback attributes remain
developer-controlled. A missing poster starts empty but remains editable.
Nested snips and competing child `<source>` elements are parsing errors.
Videos stay outside rich text. There is no video multisnip field selector yet.
The browser's native controls provide playback and seeking; ItsyWeb serves Range and
HEAD requests from pinned, validated file handles rather than in-memory video
buffers. Only catalogued files are served, with MP4 content type and hash ETags.
Do not modify a hash-named file in place; upload a replacement instead.

## Storage, deployment and recovery

Metadata is in a `type: "video"` entry beside ordinary snips in `itsy/snips.json`.
Its `video` value holds validated MP4 metadata; optional `poster` holds image
metadata directly. There is only one stored source/poster value per snip.
Commit metadata and templates. Local video files live in **itsy/media/** and
are Git-ignored. Dev, uploads and pulls ensure that ignore rule; dev refuses
already tracked video files. Files use SHA-256 names and are not renamed by admins.

Deployment stages the exact local video bytes identified by the committed
metadata, without putting those files in Git. Missing or changed bytes stop
deployment. After cloning, run `flea pull --snips-only` to retrieve live video
files before starting dev. Files not yet deployed must come from the original
working copy or a backup; retain the hash filenames.
Ordinary deployment preserves live video metadata/files and seeds new IDs.
`--with-snips` includes reviewed local metadata changes with conflict checks.

`flea pull` includes videos and referenced bytes in the same
reviewed state snapshot as photos. Stop dev first. Pull can start from committed
metadata without local video bytes, but requires full file validation before it
saves content or advances the sync baseline.
Deploy and pull never silently remove
retained video files.

Production stores uploads in the site's managed content/media/ directory,
outside code releases. Backups include referenced video bytes; restore validates
them before activation. Release payloads are bounded at 256 MiB; compressed backup
payloads at 270 MiB. Download individual files using the browser's native player.

Old replacement files remain for recovery; there is no delete/prune button yet.
At capacity, ask the developer to review retention and backups before removing
anything. Videos and posters are public assets, not private access-controlled
documents. Site bandwidth is used for playback and downloads.
