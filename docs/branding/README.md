# FleaCMS brand assets

<img src="logo.svg" alt="FleaCMS" width="332" height="80">

The owner approved the filled Little Skip mascot, concept D from the second
logo exploration round, on November 9, 2026. The three filled shell sections,
curved gaps, friendly face and bent legs come from that exact concept. Preserve
its contours. The previously approved Nunito 800 lettering, lockup dimensions
and spacing remain unchanged. Body typography stays independent.

- `logo.svg`: full navy lockup, with outlined lettering.
- `logo-light.svg`: the same lockup for dark surfaces.
- `mark.svg`: standalone mascot. Use it for icons, not as a replacement wordmark.
- `favicon.svg`: navy mascot on a fixed light background, matching the public
  website header in both light and dark browser chrome.
- `favicon.ico`: 16px and 32px PNG images in an ICO container.
- `favicon-32.png`, `apple-touch-icon.png`, `icon-192.png`, `icon-512.png`: browser
  raster exports from the favicon SVG, using the same fixed navy-on-light
  palette. Use their native dimensions.

`approved-mascot.png` and its prompt sidecar retain the approved raster source.
The SVG traces the dark ink in the approved RGB source at a mean channel
threshold of 150/255, then simplifies below one source pixel. This removes
the white background and slight generated tonal variation without redrawing
the character. Nunito outlines come from `nunito-800.woff2`; its SIL Open Font
License is retained in `Nunito-OFL.txt`. SVGs contain vector geometry only, with no external
images, font downloads, scripts or runtime dependencies.

`export.py` records the vector export, using Python's standard library and the
existing Linux FreeType library. It is an optional asset-authoring utility, not
part of site generation or the Go build. Run it from this directory to regenerate
the vector masters. `export.json` records the conversion settings. Raster icons
are exported from `favicon.svg` through the browser canvas at their stated sizes.

The framework copies these assets into the starter and its website source. The
public website repository uses identical files. Header, footer, docs and admin
links use the full SVG with an accessible FleaCMS label. README files use the
same master. The development diagnostic page includes the static logo so it
still renders when a site's own files fail.

CLI/server updates do not replace existing website layouts, customer logos,
favicons, editable content or media. Deploy Flea-owned site changes separately.
