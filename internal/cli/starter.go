package cli

var starter = map[string]string{
	"_/nanoweb.config": `{
  "site": {"name": "Fieldwork Studio", "email": "hello@example.com"}
}
`,
	"_/snips/home.json": `{
  "eyebrow": "Independent by design",
  "title": "Good things start small.",
  "introduction": "We build thoughtful spaces, useful objects, and lasting relationships. A small creative studio with room for big ideas.",
  "services": [
    {"number": "01", "name": "Strategy", "description": "Find the clear idea at the heart of your next chapter."},
    {"number": "02", "name": "Identity", "description": "Give your story a voice and a visual world of its own."},
    {"number": "03", "name": "Digital", "description": "Make a welcoming place for your work on the web."}
  ]
}
`,
	"html/_layout.html": `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><meta name="description" content="An independent creative studio for thoughtful brands and digital spaces."><title>{page.title} | {site.name}</title><link rel="icon" href="/images/mark.svg" type="image/svg+xml"><link rel="stylesheet" href="/css/style.css"></head>
<body><a class="skip" href="#main">Skip to content</a>{render header}{slot}{render footer}</body></html>
`,
	"html/_header.html":     `<header class="wrap"><a class="brand" href="/"><img src="/images/mark.svg" width="32" height="32" alt="">{site.name}</a><nav aria-label="Main navigation"><a href="/about">The studio</a><a href="mailto:{site.email}">Say hello <span aria-hidden="true">↗</span></a></nav></header>`,
	"html/_footer.html":     `<footer class="wrap"><p>{site.name}<br><span>Considered work. Human connections.</span></p><a href="mailto:{site.email}">Let's make something together ↗</a></footer>`,
	"html/index.html":       `<main id="main"><section class="hero wrap"><div><p class="eyebrow">{snip.home.eyebrow}</p><h1>{snip.home.title}</h1><p class="intro">{snip.home.introduction}</p><a class="button" href="/about">Meet the studio <span aria-hidden="true">↗</span></a></div><div class="art" role="img" aria-label="Abstract green circles and a warm orange sun"><div class="sun"></div><div class="arch"></div><p>Made with care.<br>Built to last.</p></div></section><section class="services wrap" aria-labelledby="services-title"><div class="section-heading"><p class="eyebrow">What we do</p><h2 id="services-title">A little clarity.<br>A lot of possibility.</h2></div><div class="cards">{#each snip.home.services as service}<article><span class="number">{service.number}</span><h3>{service.name}</h3><p>{service.description}</p></article>{/each}</div></section></main>`,
	"html/about/index.html": `<main id="main" class="wrap about"><p class="eyebrow">The studio</p><h1>Small team.<br>Shared curiosity.</h1><p class="intro">We believe the best work comes from paying attention. To the details, to the people, and to what really matters.</p><p>Fieldwork is an independent studio helping good ideas find their place in the world. We work closely with people who care about what they make.</p>{#if site.email}<a class="button" href="mailto:{site.email}">Tell us what you're thinking ↗</a>{/if}</main>`,
	"html/_404.html":        `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Page not found</title><link rel="stylesheet" href="/css/style.css"><main class="wrap about"><p class="eyebrow">404</p><h1>A little off<br>the path.</h1><p>This page could not be found.</p><a href="/">Back to the studio</a></main></html>`,
	"html/_500.html":        `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Temporarily unavailable</title><main><h1>We'll be right back.</h1><p>The site is temporarily unavailable. Please try again shortly.</p><a href="/">Try again</a></main></html>`,
	"html/images/mark.svg":  `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="16" fill="#244b3b"/><path d="M9 23V9h15v4H13v3h9v4h-9v3" fill="#f6f3e9"/></svg>`,
	"html/css/style.css":    `:root{color-scheme:light;--paper:#f6f3e9;--ink:#244b3b;--muted:#59695c;--line:#cdd2c3}*{box-sizing:border-box}body{margin:0;background:var(--paper);color:var(--ink);font:17px/1.65 system-ui,sans-serif}a{color:inherit;text-underline-offset:5px}a:focus-visible{outline:3px solid #ba4a2c;outline-offset:6px}.wrap{width:min(1160px,calc(100% - 64px));margin:auto}header{display:flex;align-items:center;justify-content:space-between;padding:30px 0;border-bottom:1px solid var(--line)}.brand{display:flex;gap:12px;align-items:center;font-weight:750;text-decoration:none;letter-spacing:-.7px}nav{display:flex;gap:32px;font-size:14px}nav a{text-decoration:none}.hero{display:grid;grid-template-columns:1.2fr 1fr;gap:70px;align-items:center;padding:84px 0 96px}.eyebrow{text-transform:uppercase;font-size:11px;letter-spacing:2.5px;font-weight:700}h1,h2,h3,p{margin-top:0}h1{font-family:Georgia,serif;font-weight:400;font-size:clamp(52px,6.2vw,86px);line-height:1.05;letter-spacing:-3px;margin-bottom:28px}h2{font:400 clamp(30px,4vw,46px)/1.15 Georgia,serif;letter-spacing:-1px}h3{font-size:23px;font-weight:550;margin:30px 0 10px}.intro{max-width:530px;font-size:18px;color:var(--muted);line-height:1.8}.button{display:inline-flex;gap:32px;align-items:center;padding:13px 23px;border-radius:30px;background:var(--ink);color:var(--paper);text-decoration:none;font-size:14px;margin-top:15px}.art{position:relative;isolation:isolate;height:480px;background:#dce1c9;border-radius:180px 180px 8px 8px;overflow:hidden}.sun{position:absolute;width:135px;height:135px;top:56px;right:48px;border-radius:50%;background:#cf704c}.arch{position:absolute;width:290px;height:390px;border:65px solid #648269;border-bottom:0;border-radius:160px 160px 0 0;left:-30px;bottom:-55px;transform:rotate(-24deg)}.art p{position:absolute;right:30px;bottom:20px;font:italic 23px/1.25 Georgia,serif}.services{border-top:1px solid var(--line);padding:55px 0 80px}.section-heading{display:flex;justify-content:space-between;gap:30px}.cards{display:grid;grid-template-columns:repeat(3,1fr);gap:35px;margin-top:25px}.cards article{border-top:1px solid var(--line);padding-top:20px}.number{font-size:12px;color:var(--muted)}.cards p{font-size:15px;color:var(--muted);max-width:290px}footer{display:flex;justify-content:space-between;gap:24px;border-top:1px solid var(--line);padding:32px 0 45px;font-size:13px}footer span{color:var(--muted)}.about{padding:85px 0 120px;max-width:850px}.about>p{max-width:600px}.skip{position:absolute;top:-100px;background:var(--paper);padding:12px}.skip:focus{top:0;z-index:5}@media(max-width:760px){.wrap{width:calc(100% - 40px)}header{padding:22px 0;align-items:flex-start;gap:20px}.brand{font-size:15px}nav{gap:15px;font-size:12px}.hero{grid-template-columns:1fr;gap:40px;padding:50px 0}.art{height:330px;max-width:440px;width:100%;justify-self:center}.section-heading{display:block}.cards{grid-template-columns:1fr;gap:15px}.cards p{max-width:none}h3{margin-top:14px}footer{flex-direction:column}.about{padding:50px 0 80px}h1{letter-spacing:-2px}}@media(max-width:480px){header{flex-direction:column;gap:16px}nav{width:100%;justify-content:space-between}.brand{font-size:17px}}@media(prefers-reduced-motion:no-preference){.button{transition:background .2s}.button:hover{background:#3d6550}}`,
	"README.md": `# Your Nanoweb site

Run nanoweb status to validate, then nanoweb dev and open http://127.0.0.1:3000.
Edit _/snips/home.json to change the home page. Edit _/nanoweb.config (JSON) for the site name and email. Your browser reloads after local changes.

Pages live in html/. about/index.html is /about; about/ and about.html return 404.
Assets keep their paths. Files or directories beginning with _ or . are private.
The _layout.html wraps pages. Partials such as _footer.html live anywhere in html/. Use {render footer} relative to the calling template, or {render /foo/bar/footer} from html/. Error pages stand alone.

This is a local development site. Editor, deployment, HTTPS, forms, schemas, content revisions, managed snip images, and raw HTML rendering are not implemented.
`,
}
