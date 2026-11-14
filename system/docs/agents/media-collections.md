# Blog cover media

Page media editing and collections are removed. Ordinary images/videos live under
`site/` and use normal public URLs and validated asset handles. Blog retains its
JPEG/PNG cover library only. See [photos](../photos.md) and [Blog](../blog.md).

Uploads require authenticated sessions, exact Origin, header and form CSRF,
bounded multipart parsing and the existing global processing slot. Reject symlink
or hardlinked files/directories, immutable-name collisions and corrupt bytes.
Write immutable processed photos before candidate validation and atomic content
publication. Keep limits: 8 MiB input, 20 megapixels, 2048px output, 100 photos,
64 MiB retained library. Failure may leave an unreferenced photo, never a published
missing file. Never prune retained photos automatically. Cover removal retains
library metadata and bytes. Backup/pull/deployment preserve all library photos.
