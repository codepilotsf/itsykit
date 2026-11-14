# Contact, Blog and managed local apps

Read the relevant installed `itsy/skills/add-contact-form/SKILL.md` or
`itsy/skills/add-blog/SKILL.md`, using the selected legacy directory when needed.
Preserve existing mounts, settings, layouts and authentication. Focused additions
use the site's current design. Restart dev after app config/package changes.

## Contact and email

Contact is bundled and disabled until configured. Merge into the selected config,
preserving other fields and using an approved recipient:

```json
{"apps":{"contact":{"mount":"/contact","email":{"to":"owner@example.com"}}}}
```

Provide `site/contact/index.html` and `site/contact/thank-you.html`, or the
corresponding directory at the chosen mount. Do not also create `site/contact.html`.
Use a native URL-encoded POST to that mount with required `name`, `email`,
`message`, optional hidden empty spam trap `website`, and explicit `{csrf}` inside
the form. `{csrf}` emits the entire hidden `_csrf` control, not a token value.
Unknown/repeated fields, arbitrary forms, multipart uploads, recipient controls
and visitor-selected success URLs are unsupported. Form controls remain ordinary
developer-authored HTML.

Render safe returned values through `contact.data.name/email/message`, field
feedback through `contact.invalid.name/email/message`, operational failure through
`contact.error`, and optional thank-you feedback through `contact.flash.notice`.
Validation returns 422 with preserved inputs; success redirects to the mount's
`thank-you`. Messages are emailed, not stored or queued. Visitor email is Reply-To,
never the sender or recipient. Old top-level `contact` config and `/_flea/contact`
are not aliases; the rename does not add new routes such as `/_itsy/contact`.

Private `mailers/*.email.html` and `*.email.txt` use escaped `app.*` and `email.*`
values; the HTML layout uses `{slot}`. Recipients, subjects and senders are outside
mail templates. `itsy email preview sign-in` (or `contact`) renders synthetic mail
without sending. New sites enable `dev.email_preview` for actual local outgoing
messages; production ignores it. Real delivery requires Resend setup via
`itsy email`, an authorized sender and production HTTPS. Local preview does not
prove inbox delivery. Never add an alternate mail API as incidental setup.

Admin access uses `itsy admin [email]` and `admin.emails`; an empty list disables
live sign-in. Live access changes require an authorized commit/deploy. Preserve
`{admin.csrf}`, form names/actions, code autofill and private runtime sessions.
Never copy sessions or expose authenticated admin state on public pages.

`/admin` is the emailed-code sign-in entry. Verification returns to `/admin`;
authenticated admins go to `/admin/blog` when enabled, otherwise a status page
with Sign out. Preserve exact Origin/Host checks, host-bound HttpOnly/SameSite
sessions, production HTTPS and content-version checks. Admin forms use the hidden
`csrf` control from `{admin.csrf}`; Contact retains its separate `{csrf}` input.
Blog multipart saves must check `X-CSRF-Token` before parsing uploads, then check
form CSRF. Preserve bounded processing, sanitization, immutable photo publication
and stale-save protection; failed saves retain the browser draft.

Public layouts need no admin CSRF meta, toolbar mount or `app.current_admin`.
Removed `/admin/page`, `/admin/editor-context`, `/admin/theme` and page-editor
assets (`editor.js`, `editor.css`, `toolbar.html`) return 404.
Keep Blog's bundled rich editor at `/admin/assets/blog.js`, private login/Blog
templates and their version controls.

## Blog

Blog is an optional built-in content app, not a child executable. Merge
`{"apps":{"blog":{"mount":"/blog"}}}` and use private
`site/blog/_index.html` and `_post.html`; no ordinary route/assets may occupy its
subtree. Admin templates use `/admin/blog` and `/admin/blog/edit`. Older sites may
need missing templates/CSS from a disposable matching-version site; preserve the
custom layout instead of resetting it.

Posts live in the optional top-level `blog` array in the selected
`itsy/snips.json` or legacy `flea/snips.json`. Its only other optional top-level
field is `media`; `{}` is valid empty storage. Preserve existing posts/photos.
A BlogPost has `id`, `title`, `slug`, `excerpt`, sanitized rich `body`, `status`
(`draft`/`published`), optional `published_at` and `cover`. Post IDs are ten
case-sensitive alphanumeric characters; generate unique secure-random IDs only
for new posts. Published records require RFC3339 `published_at`. Title is required
(maximum 200 characters), excerpt at most 500, slug at most 100 lowercase ASCII
letters/digits separated by single hyphens and unique across all posts. Limits
are 200 posts, 1 MiB per body and 5 MiB for content JSON. Use the installed guide
for examples, resolving stale page-CMS directions against the implementation.

`cover` is a photo ID in the `media` map, never a URL/path. Photo map records have
exactly `file`, `name`, `alt`, `width`, `height`. Photo IDs are ten alphanumeric
characters starting with a letter. Files live in `itsy/media/`, served at
`/media/`; the content hash is in the filename, not new `size` or `sha256` JSON
fields. Prefer the admin library/upload flow to inventing records. Cover-upload
metadata uses only `id`, `name`, `alt`, plus multipart `cover` bytes; its ID must
match the post's cover. Preserve existing records and immutable-file handling.
JPEG/PNG inputs are bounded at 8 MiB/20 megapixels, re-encoded with metadata
stripped and resized to at most 2048px. The library allows 100 records/64 MiB.
Keep sanitization and upload bounds; ordinary static media needs no library entry.

Public templates receive `blog.data.posts`, `blog.data.post`, `blog.data.url`,
post title/slug/excerpt/url/publication fields and optional cover src/alt/dimensions
through `blog.data.post.cover`. No global page-media namespace is supplied.
Published post title/excerpt also supply `page.title`/`page.description`, overriding
post-template comment metadata.
Only exact `{blog.data.post.body}` renders sanitized rich HTML in element content;
aliased body values are escaped. Drafts are absent from public data and routes,
even for signed-in admins. Cover files remain public, including draft covers.

The rich editor keeps Save/Cancel and cover uploads. POST `/admin/blog/save`
saves, publishes or unpublishes a post. New posts save as drafts; Cancel leaves
the saved record unchanged. Publish/Unpublish is explicit; saving a published post
changes it immediately. A future date does not schedule publication. Slug changes
create no redirect; omitting an ID does not delete live content. Normal deploy
preserves live posts/photos and seeds local-only IDs. Explicit synchronization
uses `itsy pull` and authorized `itsy deploy --with-content`, comparing posts and
photos as whole records with a destination-scoped baseline and explicit conflict
resolution; see [operations](compatibility-operations.md).
Comments, taxonomy, search, scheduling, revisions and inline body images remain
planned. Authoring instructions do not authorize live publication.

## Custom managed apps

Custom local packages and app-owned sessions are implemented. Read the installed
app guide, or `system/docs/apps.md` in an available framework checkout, before
building/registering one. Register `apps.<name>` with a mount and a page there;
place a validated `app.json` and native platform executables in
`itsy/apps/<name>/`. The manifest declares exactly `name`, `version`, `protocol: 2`
and an `executable` platform-to-basename map. Only declared regular files are
packaged; no symlinks, arbitrary commands, downloads or registry installation.
Site validation checks packages without starting processes; dev/host supervise
and handshake with the local child. This confinement is not an OS sandbox.

App output uses escaped `<name>.data.*`, `.invalid.*`, `.error` and `.flash.*`
values, never template source. Preserve reserved-name restrictions, including
`snip` and `media`; removing the page CMS does not free those namespaces.
Preserve bearer authentication, exact Origin, visitor CSRF and the separate
app-session boundary. App sessions never grant admin access. The app owns session expiry/revocation and private storage, which
survives release changes but is excluded from site backups. Arrange a separate
backup when needed; do not put app data in releases or public assets.

The app wire protocol remains 2 with `/_flea/handshake`, `X-Flea-Session` and
`__Host-flea-session-<app>` cookies. Do not mechanically rename these to Itsy,
even where website examples do. Remote services, `url`/port registration,
config-declared public read routes, a public SDK and framework-provided databases,
storage helpers or uploads remain unsupported.
