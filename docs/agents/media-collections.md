# Media and collections

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

Optimized MP4 video is implemented separately from photo processing. Read
docs/videos.md. internal/video validates bounded non-fragmented fast-start MP4,
one avc1 H.264 track and optional AAC-LC audio, without decoding/transcoding.
Keep real synthetic fixtures offline; FFmpeg was used to create test fixtures
only and is not a build/runtime dependency. Limits: 25 MiB/file, 64 MiB retained
video files, 50 entries. No WebM fallback, video conversion or provider integration.
Standalone media use typed snips.Entry values: Type image/video, direct Image or
Video metadata, and optional direct Poster photo metadata. Never duplicate these
values in a library record or template. Store bytes in flea/media/, Git-ignored locally
and below managed content/media in production. Deploy stages only checksum-pinned
files named by committed metadata; source code still comes from Git. Normal
deploy preserves remote video values; explicit snip sync compares video records.
Backups/pulls include referenced video bytes. Release archives allow 256 MiB and
backup compressed payloads 270 MiB; keep these limits consistent.
LoadSyncSource is pull-only: it checks templates/metadata without opening videos,
allowing fresh clones to fetch ignored bytes. It cannot serve HTTP. Pull must
install and fully validate videos before writing content or sync baselines.
The loader excludes flea/media/ from ordinary asset maps, validates catalogued
files and pins file handles for streaming via independent SectionReaders. Never
resolve a filesystem path from a request. Reject links; changed file stat fails
closed. Uploaded hash files are immutable, with no automatic pruning. os.File
cleanup closes handles when snapshots/readers are collected. Videos remain public.
POST /admin/page accepts media changes alongside text and metadata in one save.
Require session, exact Origin and header CSRF before multipart parsing/global upload
slot; bound body to 40 MiB, spill multipart data above 1 MiB and remove temporaries.
Validate form CSRF too, bound individual files, and reject unlisted file parts.
Only authenticated uploads receive extended bounded network deadlines.
Edit mode opens image/video dialogs by clicking designated media. Done previews
the draft; shared topbar Save/Cancel publishes/restores the whole page. No Photos
or Videos topbar buttons or standalone library upload endpoints. Use object URLs
for private previews and revoke them on Cancel. No video multisnip fields yet.

Read docs/photos.md for photo commands, data boundaries and limits, and
docs/recovery.md for backup/restore transport requirements.

Photos are separate from rich text. Standalone image snips hold direct metadata;
snips.Store.Media holds multisnip photo choices. Local flea/media/ and production
content/media hold immutable hash-named JPEG/PNG files. Aggregate ImageAssets and
VideoAssets include typed active/archive entries for loading and transport.
Use bounded decoding, EXIF orientation, metadata-stripping re-encoding and 2048px
resizing. Require session and header CSRF before acquiring the global upload slot
or parsing a large body; also keep normal Origin/form CSRF checks. Files are public.
Publish validated immutable bytes, never request-derived filesystem reads. Preserve
photos in ordinary deployment and include them in snip synchronization and backup.
Retain old file versions; no automatic destructive pruning.
Multisnip image authoring uses named img snip attributes. Import local seed photos;
compiled snip.photo.src/alt and binding attributes provide typed photo selectors.
Validate photo IDs against the current library and never permit arbitrary URLs or
HTML in a photo field. Photo asset writes precede extraction content/template
transactions; failure can leave an unreferenced immutable file, not a broken link.

## Multisnips

Read docs/multisnips.md and docs/email-templates.md for current syntax and limits.
Page.Multisnips maps names directly to ordered arrays of stable-ID items with
string fields. Author multisnip="team" wrappers with named inner snip fields;
adjacent seeds may repeat when their extracted markup is identical. Keep one loop,
extract all seed values in order, and reject differences outside fields. Never
overwrite an existing stored array when reseeding, including an empty one.
dev extracts it in the same content-first journal as ordinary snips. Validate the
combined rewritten source before writing. Templates use {#multisnip team},
{snip.name}, {/multisnip}; one root, complete field wrappers, page-local only.
The template is the persistent text/photo-field schema even when arrays are empty.
Reject navigation/form/inert contexts, nested groups and loops. Sanitized HTML
is inserted only in validated element text. Browser writes validate rendered group
membership, current fields, item IDs/counts and content version. Retain removed
developer fields by stable ID. Normal deploy preserves live arrays including empty
ones; explicit sync resolves whole-array conflicts rather than guessing order.
Editing automatically enables on-page arrangement, with no Collections button.
Use dashed item overlays, inset corner grip/remove controls and per-group Add item.
Controls show on hover/focus, always on coarse pointers. A drag-only inert preview
follows the pointer with a shadow; the hidden original reserves the landing space.
Animate neighboring movement and settling unless reduced motion is requested.
Support mouse/touch and arrow/Home/End keys on handles, with live announcements
and Escape cancelling a drag. Keep controls in the editor shadow root, outside
saved content. Empty groups retain an Add tile. Save/Cancel publishes/restores
content and order together, keeping Editing on. Ordinary text
gets dashed overlay boundaries without changing layout. Click named photo fields to choose a library image or
upload a JPEG/PNG. Stage uploads as private object URLs; submit only photos still
referenced by visible collection fields with the page's multipart Save. Validate
new library IDs and field membership, process immutable bytes, then publish library
metadata and item references together. Cancel revokes previews without uploading;
empty photos show a neutral portrait and Add photo during Edit. The collection
picker uses thumbnail buttons with aria-pressed selection, a No photo choice,
Upload photo and a selected-image preview. Show alt text with the preview;
saved library descriptions are read-only. Revoke dialog-only object URLs on
selection change, close and page Cancel, separately from staged page uploads.
No duplicate/move-up/down buttons. Video fields,
general field schemas and item revision history are pending.
