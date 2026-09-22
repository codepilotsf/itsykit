# Your Flea CMS site

For owner editing, add approved addresses to admin.emails in _/flea.config.
The empty default list disables /admin. Deploy with HTTPS; Flea CMS prompts for
Resend email settings once and stores them only on the server. Owners sign in at
/admin using an emailed code, then edit page details and snip text/simple HTML.
For local login, export RESEND_API_KEY and FLEA_EMAIL_FROM before starting dev.
Never commit these secrets. Restart dev after changing the approved addresses.

Run flea status to validate, then flea dev and open http://127.0.0.1:3000.

Run flea admin to add an approved login email, or flea admin owner@example.com.
It preserves existing admins and other settings. Restart dev afterward; live
access changes need deployment. Resend credentials are still required for login.

Start with index.html. Edit _/snips.json for the welcome text and page metadata, and
_/flea.config (JSON) for your site name. Saved changes reload your browser.

Add <h1 snip="title">Your heading</h1> or a section with formatted HTML while dev
is running. Dev extracts its inner HTML to _/snips.json and inserts a short-ID
{snip a7Km2pQ9xB} marker, without quotes. IDs contain only letters and digits.
Each relevant save also updates labels and route ownership, archives removed
snips, and restores archived content if its ID returns. The terminal reports
snip updates. Existing content is preserved. Run one dev process per project;
local writes use an in-memory mutex, not a lock file.
Nested regions and template directives inside new regions
are not allowed. Rich content is sanitized; scripts, styles, forms, and embeds
are removed. Status does not rewrite files. Commit the generated content and
templates together before deploying. Structured data for loops belongs in _/data/.

Deploy preserves remote rich content and page metadata, seeds missing IDs, and
archives orphaned IDs after successful activation. New local values do not
overwrite owner changes. The browser editor uses the same content store; a rollback
command is not built yet.

Add about.html to create /about. Explicit .html URLs and trailing slashes
return 404. Assets keep their paths. Files beginning with _ or . are private.

The _layout.html wraps pages and contains the header and footer directly.
Optional partials can live anywhere in the project: use {render card} for a local
_card.html or {render /shared/card} from the project root. The 404 page uses the layout;
write it as page content. The 500 error page stays standalone.
Styles and the Flea CMS logo are in
css/ and images/mark.svg; replace them with your own.

Styling uses Prewind 1.3.1 from unpkg with Preflight-inspired defaults, without a
build step. css/global.css imports theme.css (tokens), typography.css,
forms.css, and ui.css. theme.css imports Google Fonts at the top: Manrope for
headings and Source Sans 3 for body text, with system fallbacks. Keep global rules in the
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
HTTPS preferences, then saves _/deploy.json. It offers to commit your changes
before deploying, showing the file list and asking for a commit message.
Each branch has its own saved destination; staging suggests staging.example.com.
Dirty Git status blocks deployment. Only committed content and snips are shipped.
Use flea deploy -configure to change a mapping. DNS is checked from the target
before upload. See the Flea CMS framework README for full details.

Editor, forms, schemas,
content revisions, managed snip images, and raw HTML rendering are not implemented.

Icons: use {icon arrow-right} or {icon search size="20" aria-label="Search"}.
flea dev downloads only referenced Lucide 1.47.0 SVGs into
_/icons/lucide-1.47.0/ and renders them inline. Deploy resolves missing icons
inside its temporary committed snapshot. Keep the cache and its LICENSE
in version control for offline use. flea status never downloads icons.
Missing icons or failed downloads are omitted with terminal warnings; pages
still load normally. Invalid icon syntax or unsafe SVGs remain errors.
Classes, dimensions, stroke attributes, and aria-/data- attributes are supported;
values must be quoted. Event handlers and inline style are not supported.

Dev logs requests as GET /test, or 404 GET /missing for a non-200 response.
Successful internal reload polls and query strings are omitted from logs.
