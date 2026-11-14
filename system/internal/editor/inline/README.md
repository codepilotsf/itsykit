# Blog editor assets

These files are embedded in ItsyWeb for the private Blog editor. There is no
frontend build step. Only `/admin/assets/blog.js` is served.

- `squire.js`: unmodified Squire 2.4.9, MIT; see SQUIRE-LICENSE.
  SHA-256: `2fc7ce78b2e3a73bc81d3d143725fea4e424a778434f5982d53db233599ab981`.
- `purify.js`: unmodified DOMPurify 3.4.16; see DOMPURIFY-LICENSE.
  SHA-256: `2c90a9b46d6463f26038a29b686e82bc91de01fdac9d5229e7cfe3b360134ea2`.
- `blog.js`: rich post body editing, cover selection/upload, explicit
  Save/Publish/Unpublish, draft protection and Cancel navigation.

DOMPurify handles client hygiene; server sanitization, session/Origin/CSRF,
upload limits and snapshot versions remain authoritative. Failed saves preserve
drafts. New posts save as drafts; publishing/unpublishing is explicit. The page
CMS, extraction, toolbar controllers and isolated page-editor prototype are gone.
