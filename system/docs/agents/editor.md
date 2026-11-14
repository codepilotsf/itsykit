# Blog administration

`/admin` is the email-code sign-in entry point. After verification it returns to
`/admin`; authenticated admins go to `/admin/blog` when Blog is enabled. Without
Blog they see a status page with Sign out, not an editor for website pages.
`/admin/blog/edit` creates or edits a post, and `/admin/blog/save` saves, publishes
or unpublishes it. New posts save as drafts. Only published posts enter public
routes. Save on a published post updates it immediately; Unpublish returns it
to a private draft. See [Blog](../blog.md).

Keep `site/admin.layout.html` and the private sign-in/Blog templates. Admin forms
use their own hidden `csrf` field from `{admin.csrf}`. Blog multipart saves also
send `X-CSRF-Token`. Exact Origin, host-bound HttpOnly/SameSite sessions, production
HTTPS, upload bounds and snapshot versions remain required. Public layouts need
only an ordinary `/admin` link. They have no toolbar mount, `app.current_admin`
state or admin CSRF meta. Public Contact forms retain the separate `{csrf}` input.
The removed `/admin/page`, `/admin/editor-context`, `/admin/theme`, `editor.js`,
`editor.css` and `toolbar.html` endpoints return 404. Only Blog's `blog.js` asset
includes the pinned Squire and DOMPurify libraries.

### Developer-managed styling

Edit `site/css/theme.css` and ordinary stylesheets. Prewind/tokens, static assets
and committed CSS deployment remain supported. There is no theme editor or overlay.


Sessions retain hashed tokens only, private mode-0600 durable state, bounded OTPs,
rate limits, exact Host/Origin and CSRF. Renew only authenticated navigation or
trusted activity, never idle timers. Owner removal and logout revoke sessions.
Dev LAN admission never grants production access. Keep email sender/checker and
private local preview boundaries. Blog uploads authenticate and check header CSRF
before parsing multipart or acquiring the processing slot. Validate form CSRF
again, bound body/upload sizes and reject unexpected files. Publish immutable
photo bytes before validating and atomically writing the content candidate.

`editstore` holds the local content mutex or remote deployment lock, refuses
pending recovery and changed releases, compares the submitted version, validates
the full candidate, durably saves once and then publishes. The canonical receipt
comes from that committed snapshot under the lock. Failed or stale saves retain
the browser draft. Blog editing never mutates page HTML, metadata or CSS.
