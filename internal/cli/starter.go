package cli

import _ "embed"

//go:embed starter/css/style.css
var starterStyle string

//go:embed starter/css/theme.css
var starterTheme string

//go:embed starter/css/prewind-1.3.1.css
var starterPrewind string

//go:embed starter/css/typography.css
var starterTypography string

//go:embed starter/css/forms.css
var starterForms string

//go:embed starter/css/ui.css
var starterUI string

//go:embed starter/css/application.css
var starterApplication string

var starter = map[string]string{
	"_/nanoweb.config": `{
  "site": {"name": "Nanoweb"}
}
`,
	"_/snips/home.json": `{
  "title": "Welcome to Nanoweb.",
  "introduction": "Your new site is ready. Make yourself at home."
}
`,
	"html/_layout.html": `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{page.title} | {site.name}</title>
    <link rel="icon" href="/images/mark.svg" type="image/svg+xml">
    <link rel="stylesheet" href="/css/style.css">
  </head>
  <body class="min-h-screen flex flex-col bg-white">
    <header class="container max-w-full py-lg">
      <a class="inline-flex items-center gap-xs text-black no-underline" href="/" aria-label="Nanoweb home">
        <img src="/images/mark.svg" width="36" height="36" alt="">
        <span class="text-xl font-semibold tracking-tight">nanoweb</span>
      </a>
    </header>
    {slot}
    <footer class="container max-w-full py-sm border-t border-light text-dark text-sm">A small starting point for your next idea.</footer>
  </body>
</html>
`,
	"html/index.html": `<main class="container max-w-full flex-1 py-xl">
  <h1 class="mt-0">{snip.home.title}</h1>
  <p class="text-lg text-dark m-0">{snip.home.introduction}</p>
  <p class="max-w-text mt-md text-dark">Start with <code>html/index.html</code>. Save a change and your browser will reload.</p>
</main>
`,
	"html/_404.html": `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Page not found | Nanoweb</title>
    <link rel="stylesheet" href="/css/style.css">
  </head>
  <body class="min-h-screen flex flex-col bg-white"><main class="container max-w-full flex-1 py-xl">
    <h1>Page not found.</h1>
    <p>This page does not exist.</p>
    <a href="/">Back home</a>
  </main></body>
</html>
`,
	"html/_500.html": `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Temporarily unavailable | Nanoweb</title>
  </head>
  <body class="min-h-screen flex flex-col bg-white"><main>
    <h1>Temporarily unavailable.</h1>
    <p>The site is temporarily unavailable. Please try again shortly.</p>
    <a href="/">Try again</a>
  </main></body>
</html>
`,
	"html/images/mark.svg": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 40 40">
  <path d="M5 33V7h8v26Z" fill="#413778"/>
  <path d="M16 7h2l17 19v7h-8L16 20Z" fill="#5746d6"/>
  <path d="M27 7h8v15l-8-9Z" fill="#413778"/>
</svg>
`,
	"html/css/style.css":         starterStyle,
	"html/css/theme.css":         starterTheme,
	"html/css/prewind-1.3.1.css": starterPrewind,
	"html/css/typography.css":    starterTypography,
	"html/css/forms.css":         starterForms,
	"html/css/ui.css":            starterUI,
	"html/css/application.css":   starterApplication,
	"README.md": `# Your Nanoweb site

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
`,
}
