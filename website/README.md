# FleaCMS website source

This scaffold uses site/ for pages, layouts and assets, root mailers/ for email
templates, flea/snips.json for content, and root .flea.config.json for settings.
Public URLs do not include site/. Layout filenames end in .layout.html and mail
bodies in .email.html or .email.txt. Runtime belongs in ignored flea/runtime/;
managed photo/video uploads belong in flea/media/. No old-directory fallback.

This is an unpublished website scaffold. From this directory, run `flea dev`
to preview it. Public download links are intentionally absent until release
artifacts are signed and published. The framework starter remains in templates/.

First `flea dev` asks for the site name, with a title-cased folder default.
It saves `app.site_name` in `.flea.config.json`. Use `{app.site_name}` in templates;
put other public site-wide values in `app` too, such as `{app.phone}`.
Admin settings and secrets are separate. Commit the name before deploying.

Owners can choose Theme beside Edit to preview and save existing theme.css
variables. Cancel restores the previous appearance. Theme controls remain separate
from the site's styling. Fonts use the existing stacks or system fonts.

Commit snips.json and theme.css as site seeds. Normal deployment keeps live owner
changes. Stop dev, then run `flea pull` to retrieve snips and theme, or select
`--snips-only` / `--theme-only`. Review and commit the resulting changes.
To publish intentional local edits, use `flea deploy --with-snips`,
`--with-theme`, or both. Existing live sites require a pull baseline; conflicts
ask for a choice without a default. Config, secrets and sessions are never pulled.

For owner editing, add approved addresses to admin.emails in .flea.config.json.
The empty default list disables /admin. Deploy with HTTPS; FleaCMS prompts for
Resend email settings once and stores them only on the server. Owners sign in at
/admin using an emailed code, then edit text directly on the website. Edit opens
the formatting bar and Page details; Save changes publishes them together.
For local login, run flea email to save Resend settings in a private, Git-ignored
.env file. Never commit secrets. Admin access and sign-in mail settings update
automatically in dev.

Owners stay signed in across browser closure, server restarts, and deployment.
admin.session_idle_days in .flea.config.json defaults to 30 (1–365 days). Admin activity
renews the window; an idle tab does not. Each browser has its own session.
Flea stores local sessions in flea/runtime/sessions.json and live sessions in the server's
sites/<site>/flea/runtime/sessions.json, outside releases. These private runtime files are
Git-ignored and never deployed. Do not edit or copy them. Logout revokes that
browser; removing an owner's email revokes their access.

Edit admin/index.html, code.html, setup.html, and error.html
directly to customize the owner pages. Their leading flea settings comments select
admin.layout.html, separately from the public main.layout.html. Signed-in owners
return to the public page for editing. Textareas size to content, up to 400px.
Keep form actions, field names, and hidden {admin.csrf} fields intact. The server
controls authentication; these files cannot be accessed as public pages/assets.

Run flea status to validate, then flea dev and open http://127.0.0.1:3000.

Run flea admin to add an approved login email, or flea admin owner@example.com.
It preserves existing admins and starts email setup if needed; live
access changes need deployment. Resend credentials are still required for login.

Start with index.html. Edit flea/snips.json for the welcome text and page metadata, and
.flea.config.json (JSON) for your site name. Saved changes reload your browser.

Add <h1 snip>Your heading</h1> or a section with formatted HTML while dev
is running. Dev extracts its inner HTML to flea/snips.json and removes the authoring
attribute, leaving <h1>{snip a7Km2pQ9xB}</h1>. IDs contain only letters and digits
and are written without quotes. The saved directive identifies the editable content.
Each relevant save also updates labels and route ownership, archives removed
snips, and restores archived content if its ID returns. The terminal reports
snip updates. Existing content is preserved. Run one dev process per project;
local content writes use an in-memory mutex; sessions use a separate private lock.
Nested regions and template directives inside new regions
are not allowed. Rich content is sanitized; scripts, styles, forms, and embeds
are removed. Status does not rewrite files. Commit the generated content and
templates together before deploying. Structured data for loops belongs in flea/data/.

Deploy preserves remote rich content and page metadata, seeds missing IDs, and
archives orphaned IDs after successful activation. New local values do not
overwrite owner changes. The browser editor uses the same content store; a rollback
command is not built yet.

Add about.html to create /about. Explicit .html URLs and trailing slashes
return 404. Assets keep their paths. Files beginning with _ or . are private.

The main.layout.html wraps pages and contains the header and footer directly.
Optional partials can live anywhere in the project: use {render card} for a local
_card.html or {render /shared/card} from the project root. The 404 page uses the layout;
write it as page content. The 500 error page stays standalone.
Styles and the FleaCMS logo are in
css/ and images/mark.svg; replace them with your own.

Styling uses Prewind 1.3.1 from unpkg with Preflight-inspired defaults, without a
build step. css/global.css imports theme.css (tokens), typography.css,
forms.css, and ui.css. theme.css imports Roboto Mono from Google Fonts for both
headings and body text, with monospace fallbacks. Keep global rules in the
styles layer; Prewind utilities override them. Edit theme.css for colors and
fluid sizing. Form styles do not add form submission handling.

New sites start with a Git repository on main, unless you used --skip-git or Git
was unavailable. No initial commit is made automatically.

After flea setup, run the deployment wizard:

```sh
flea deploy
```

Point your domain at the server and allow inbound TCP 80/443 before enabling
HTTPS. The first run works before your first commit and asks for the branch's server/domain, WWW redirect and
HTTPS preferences, then saves deploy.branches in .flea.config.json. It offers to commit your changes
before deploying, showing the file list and asking for a commit message.
Each branch has its own saved destination; staging suggests staging.example.com.
Dirty Git status blocks deployment. Only committed content and snips are shipped.
Use flea deploy -configure to change a mapping. DNS is checked from the target
before upload. See the FleaCMS framework README for full details.

General-purpose forms, schemas, content revisions, managed snip images, and
unrestricted raw HTML rendering are not implemented.

Icons: use {icon arrow-right} or {icon search size="20" aria-label="Search"}.
flea dev downloads only referenced Lucide 1.47.0 SVGs into
flea/icons/ with matching names and embedded license notices, and renders them inline. Deploy resolves missing icons
inside its temporary committed snapshot. Keep the cache and its LICENSE
in version control for offline use. flea status never downloads icons.
Missing icons or failed downloads are omitted with terminal warnings; pages
still load normally. Invalid icon syntax or unsafe SVGs remain errors.
Classes, dimensions, stroke attributes, and aria-/data- attributes are supported;
values must be quoted. Event handlers and inline style are not supported.

Dev logs requests as GET /test, or 404 GET /missing for a non-200 response.
Successful internal reload polls and query strings are omitted from logs.
