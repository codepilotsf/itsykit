# Real-page editor assets

These files are embedded in Flea, not copied into generated sites. There is no
npm install or frontend build step.

- `squire.js`: unmodified Squire 2.4.9 distribution, MIT. See SQUIRE-LICENSE.
  SHA-256: `2fc7ce78b2e3a73bc81d3d143725fea4e424a778434f5982d53db233599ab981`.
- `purify.js`: unmodified DOMPurify 3.4.16 distribution. See DOMPURIFY-LICENSE.
  SHA-256: `2c90a9b46d6463f26038a29b686e82bc91de01fdac9d5229e7cfe3b360134ea2`.
- `owner.js`: authenticated editing and atomic page-save controller.
- `owner.html`: shadow-root toolbar and dialogs; switch styling uses the bundled
  canonical `templates/site/css/ui.css`. Selected Lucide paths use LUCIDE-LICENSE.

Vendor bytes match the versions already evaluated in design/inline-editor/vendor.
Update both copies together deliberately, including licenses and browser checks.
DOMPurify handles paste hygiene; it never replaces server-side save validation.

The owner controls use a shadow root with isolated CSS so site styles cannot
normally change the toolbar. Page content keeps its own styles and outer markup.
Only complete server-marked snip wrappers can become editable. The server markers
come from rich directives; the source `snip` attribute is consumed by extraction
and is not part of runtime detection. Rich containers support headings, marks,
links, lists and alignment; leaf wrappers stay text-only. Page details and changed
regions publish together through /admin/page with session, Origin and CSRF checks.
The rendered snapshot supplies the version; stale saves leave the draft intact.

Optional Chromium smoke check, using a disposable profile and synthetic owner:

```sh
FLEA_TEST_BROWSER='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' \
  go test ./internal/editor -run TestPublicPreviewBrowser -count=1 -v
```

This exercises actual Squire activation, inactive controls, protected navigation,
mixed-region exclusion, layout spacing, headings/links/alignment, metadata,
failed-save draft retention and Cancel restoration. Backend tests cover atomic
publication, metadata-only pages, shared content, stale drafts and HTTPS login/save.
It does not test real email, clipboard input or other browser engines.
