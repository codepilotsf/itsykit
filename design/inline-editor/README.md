# Inline-editor prototype

Run from the Flea repository:

```sh
go run ./design/inline-editor
```

Open http://127.0.0.1:3143. Use `-port 3144` if necessary.
Browse with the site's navigation, then choose **Edit** to edit the current page.
On the right, Edit switches to Cancel/Save changes. Sign out remains visible,
separated from the editing controls by a small vertical divider.
Save stays disabled until content or metadata differs from the starting values.
Page details appears only in edit mode, in the second bar beside formatting.
That bar stays visible throughout editing. Formatting is dimmed and disabled
without an active editable region; Page details remains available. Content has
no selection border except a keyboard-focus indicator. Dialogs fade/slide 20px
(280ms entrance, 240ms exit), remaining modal until the exit finishes,
and page navigation crossfades the main content using native View Transitions.
Both respect reduced motion. Page fades explicitly retain their final opacity
until snapshot removal, with the old root snapshot hidden to prevent stale flashes.

This is an isolated experiment, not the shipped editor. Sign-in is simulated.
Saves update only this tab's sample data; refreshing resets it. It never reads or
writes a generated site's content, sessions, or configuration. The local server
binds only to loopback and serves embedded demo assets and starter CSS/logo.
It has no authentication and must not be exposed as a production service.
No Node build step is needed. Starter font imports may access Google Fonts.

## What to try

- H1–H4 and normal text; bold, italic, underline, and strikethrough.
- Links, bullet/numbered lists, and left/center/right alignment.
- Pasting rich text, undo/redo, and switching between separate snip regions.
- Page title and description, including Privacy which has no page snips.
- Shared footer edits, which apply across sample pages after saving.
- Cancel and unsaved-navigation confirmation.
- View/edit the current link destination, or remove it from the link dialog.
  Undo/redo buttons are omitted for now; keyboard shortcuts still work.

Open **Prototype checks and saved HTML** and run the checks. All 21 checks passed
in Chrome on 2026-09-24: heading conversion and history, inline marks, lists,
links, alignment, paste cleanup, per-region history, and matching site typography.
Manual selection, bold formatting, explicit save, contextual unlink, and undo
returning Save to its disabled state were also exercised. Focusing a region does
not create an undo step. The controls share a 36px height.
Safari, Firefox, mobile keyboards, real Word/Google Docs clipboard content,
and assistive technology still need testing. Browser checks invoke Squire's paste
API with fixtures; they are not a substitute for those real clipboard checks.

## Runtime integration

The link dialog defaults bare domains to HTTPS, preserves explicit schemes and
relative paths, and offers Open in a new tab (off by default). New-tab links keep
`target="_blank" rel="noopener noreferrer"` through demo saves. Production still
preserves this restricted policy too.
The switch is a native checkbox styled in `templates/css/ui.css`, with no
JavaScript required for its appearance or keyboard toggling.

- Go's sanitizer now supports H1 and narrowly validated text alignment.
  Client-side DOMPurify is paste hygiene, never the server security boundary.
- Keep `/admin` for login, then return to the homepage in browsing mode. Add the
  authenticated bar to public pages without leaking it through shared caches.
- Expose editable-region IDs safely
  without changing template structure or introducing invalid block nesting.
- Unlinked-page discovery is deferred; there is no editing page menu.
- One runtime Save action now publishes conflict-checked, atomic metadata/snips
  changes, retaining session, Origin, CSRF, content limits and deployment locking.
- Decide how conditional or repeated rendered snips should be represented.
- Test selection across toolbar focus, mixed blocks, nested lists, keyboard
  navigation, CSP, narrow screens, and arbitrary site CSS.
- Use real authenticated identity. Joe Appleseed is a demo label, not a new
  account-name feature.

Media remains separate from rich text. Repeatable multi-snips, Small/Large Text,
tables, colors, and other advanced formatting are not part of this experiment.
The runtime editor now replaces the old dashboard; cross-browser and real clipboard
checks remain follow-up work.

## Vendored files

Pinned distribution files from npm via unpkg, retained unmodified:

- `squire-rte@2.4.9/dist/squire.js`, saved as `vendor/squire-2.4.9.js`.
  MIT license in `vendor/SQUIRE-LICENSE`.
  SHA-256: `2fc7ce78b2e3a73bc81d3d143725fea4e424a778434f5982d53db233599ab981`.
- `dompurify@3.4.16/dist/purify.min.js`, saved as `vendor/purify-3.4.16.min.js`.
  Package license in `vendor/DOMPURIFY-LICENSE`.
  SHA-256: `2c90a9b46d6463f26038a29b686e82bc91de01fdac9d5229e7cfe3b360134ea2`.
- Prewind copied from the project's sibling `../prewind/prewind.css`, unchanged.
  SHA-256: `535b341e0622e902f2f696293a0bded6c07b156532ad368314f5d772f09bc385`.

Squire is 60,254 bytes raw / 18,560 gzip; DOMPurify is 28,885 / 11,414.
Together the libraries are about 30 KB gzipped, before the demo glue, CSS,
and browser checks. The preview does not currently enable HTTP compression.
Copies of Squire and DOMPurify are now embedded into the Flea binaries for the
authenticated real-page editor, with their licenses in `internal/editor/inline/`.
They are not copied into generated sites. The demo remains isolated, while the
runtime uses real sessions and atomic saves with the same core formatting controls.
The nine toolbar SVGs in `vendor/lucide/` come from `lucide-static@1.47.0/icons/`
via unpkg; its LICENSE is retained beside them. No Lucide JavaScript or npm
installation is needed.
