# Fern & Field example

A fictional neighborhood nursery, built in HTML and captured inside the real
FleaCMS 0.5.0 editor. This source is kept outside the public site so it can be
reused for the planned overview video.

## Recreate the example

1. Create a separate temporary project with `flea new fern-and-field --skip-git`.
2. Copy this folder's `main.layout.html` and `index.html` into that project's
   `site/`, `nursery.css` into `site/css/`, and `nursery.jpg` into `site/images/`.
   Keep the other generated starter files.
3. Run `flea dev -port 3101`. Use Fern & Field as the site name and
   `hello@fernandfield.example` as its fictional owner. Keep local email preview
   enabled. No mail is sent and no external account is needed.
4. Sign in at `/admin` using the local preview code. Click the heading and change
   “green” to “greener”. Leave the change unsaved for the screenshot.

The full poster was captured at 1100 × 740 CSS pixels. The detail is a 512 × 266
crop starting at x=24, y=254. Browser screenshots are JPEG.

## Homepage assets

- `site/images/nursery-editor.jpg`: full editor still, also suitable as the
  initial poster when the video is ready.
- `site/images/nursery-edit-detail.jpg`: closer view of the selected heading.

The current homepage deliberately has no video control or playback promise.
When a video is available, replace the hero's image/link with a native video
using this poster, visible controls, and captions.

## Photograph provenance

`nursery.jpg` was generated with the built-in image-generation tool on
September 30, 2026. Only the photograph is generated; the nursery layout and
editing UI in the screenshots are rendered HTML and the real Flea editor.

Final generation prompt:

> Use case: photorealistic-natural. Asset type: editorial photograph for a fictional neighborhood plant nursery website named Fern & Field. Primary request: a beautiful, believable close-up of lush potted ferns and delicate leafy plants on a weathered wooden bench at a small neighborhood nursery. Soft morning light through greenhouse glass, muted natural greens, aged terracotta pots, subtle texture, peaceful and tactile. Landscape 3:2 composition with foliage and one large terracotta fern pot as the focal point, natural imperfect arrangement and softly blurred greenhouse background. No people, no lettering, no logos, no website UI, no watermark. This photo will occupy the right half of an elegant cream-colored nursery website. Return a saved local image file for use in the website.

## Publishing the homepage later

This redesign changes the existing two homepage snips locally. Ordinary
deployment preserves live owner content. Follow the main README's pull-baseline
and `--with-snips` workflow if the revised hero copy should also be published.
Do not deploy this branch automatically.
