# Your Nanoweb site

Run nanoweb status to validate, then nanoweb dev and open http://127.0.0.1:3000.

Start with html/index.html. Edit _/snips/home.json for the welcome text and
_/nanoweb.config (JSON) for your site name. Saved changes reload your browser.

Add html/about.html to create /about. Explicit .html URLs and trailing slashes
return 404. Assets keep their paths. Files beginning with _ or . are private.

The _layout.html wraps pages and contains the header and footer directly.
Optional partials can live anywhere in html/: use {render card} for a local
_card.html or {render /shared/card} from html/. The 404 page uses the layout;
write it as page content. The 500 error page stays standalone.
Styles and the Nanoweb logo are in
html/css/ and html/images/mark.svg; replace them with your own.

Styling uses Prewind 1.3.1 from unpkg with Preflight-inspired defaults, without a
build step. css/global.css imports theme.css (tokens), typography.css,
forms.css, and ui.css. theme.css imports Google Fonts at the top: Manrope for
headings and Source Sans 3 for body text, with system fallbacks. Keep global rules in the
styles layer; Prewind utilities override them. Edit theme.css for colors and
fluid sizing. Form styles do not add form submission handling.

This is a local development site. Editor, deployment, HTTPS, forms, schemas,
content revisions, managed snip images, and raw HTML rendering are not implemented.

Icons: use {icon arrow-right} or {icon search size="20" aria-label="Search"}.
nanoweb dev downloads only referenced Lucide 1.47.0 SVGs into
_/icons/lucide-1.47.0/ and renders them inline. Keep the cache and its LICENSE
in version control for offline use. nanoweb status never downloads icons.
Missing icons or failed downloads are omitted with terminal warnings; pages
still load normally. Invalid icon syntax or unsafe SVGs remain errors.
Classes, dimensions, stroke attributes, and aria-/data- attributes are supported;
values must be quoted. Event handlers and inline style are not supported.

Dev logs requests as GET /test, or 404 GET /missing for a non-200 response.
Successful internal reload polls and query strings are omitted from logs.
