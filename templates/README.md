# Your Nanoweb site

Run nanoweb status to validate, then nanoweb dev and open http://127.0.0.1:3000.

Start with html/index.html. Edit _/snips/home.json for the welcome text and
_/nanoweb.config (JSON) for your site name. Saved changes reload your browser.

Add html/about.html to create /about. Explicit .html URLs and trailing slashes
return 404. Assets keep their paths. Files beginning with _ or . are private.

The _layout.html wraps pages and contains the header and footer directly.
Optional partials can live anywhere in html/: use {render card} for a local
_card.html or {render /shared/card} from html/. Error pages stand alone.
Styles and the Nanoweb logo are in
html/css/ and html/images/mark.svg; replace them with your own.

Styling uses local Prewind 1.3.1 with Preflight-inspired defaults, without a
build step or CDN. css/style.css imports theme.css (tokens), typography.css,
forms.css, ui.css, and application.css (site rules). Keep global rules in the
styles layer; Prewind utilities override them. Edit theme.css for colors and
fluid sizing. Form styles do not add form submission handling.

This is a local development site. Editor, deployment, HTTPS, forms, schemas,
content revisions, managed snip images, and raw HTML rendering are not implemented.
