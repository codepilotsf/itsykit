# ItsyKit

<img src="site/images/logo.svg" alt="ItsyKit" width="196" height="55">

Write HTML. Skip the build pipeline.

Itsy adds layouts, partials, JSON data, and clean URLs to HTML files. The same
command gives you live reload, SSH deployment, automatic HTTPS, and backups.
Add Contact or Blog when you need them.

This directory is the public website. The CLI, server, and canonical starter live
in [../system/](../system/README.md). Both share one repository.

## Framework requirement

This website requires Itsy 0.22.0 and server protocol 30 for comment metadata
and Blog-only content storage. Earlier installers cannot load it. Existing
page-CMS sites need an explicitly approved conversion and verified backups before
a shared-server upgrade. Keep historical release notes tied to their actual releases.

## Run this website

For a new site on the published test release, Apple Silicon macOS and x86-64
Linux downloads are available through:

```sh
curl -fsSL https://itsykit.dev/install.sh | bash
```

Run this command in Bash or Zsh, then open a new terminal. The installer configures
PATH for future terminals. To keep using the current terminal, run the optional
PATH command printed by the installer.
Use version 0.22.0 or newer for this checkout. From `website/`:

```sh
itsy dev
```

Open the Website URL printed in the terminal. Save a file to reload the page.
Press Ctrl-C to stop. This is an existing site; preserve its custom layout and
assets rather than copying the starter over it.

To build a new site, follow [Getting Started](https://itsykit.dev/docs).
[Installation and updates](https://itsykit.dev/docs/installation) covers PATH
setup and the current public testing limits.

## Find a guide

Start with [pages, layouts and data](https://itsykit.dev/docs/templates),
[CSS and icons](https://itsykit.dev/docs/styling), then
[local development](https://itsykit.dev/docs/development) and
[deployment](https://itsykit.dev/docs/deployment).

For optional features, see [Contact and email](https://itsykit.dev/docs/email), and
[Blog](https://itsykit.dev/docs/blog).

The remaining guides cover [SEO](https://itsykit.dev/docs/seo),
[redirects](https://itsykit.dev/docs/redirects),
[managed apps](https://itsykit.dev/docs/apps),
[coding agents](https://itsykit.dev/docs/agents), and
[recovery](https://itsykit.dev/docs/recovery).
Use `itsy --help` or a command's `--help` for options.

## Project files

| Path | Purpose |
| --- | --- |
| `site/` | Public pages, partials, layouts, and assets |
| `site/css/theme.css` | Colors, fonts, and spacing tokens |
| `site/css/global.css` | Shared styles and Prewind imports |
| `itsy.config.json` | Site settings and branch deployment destinations |
| `itsy/snips.json` | Blog posts and cover-photo records |
| `itsy/data/` | Developer-managed JSON data |
| `itsy/media/` | Blog cover-photo files |
| `itsy/icons/` | Cached Lucide icons |
| `itsy/runtime/` | Private, Git-ignored runtime state |
| `mailers/` | Private email templates |
| `.agents/` and `itsy/skills/` | Project instructions for coding agents |

Private configuration, runtime state, mailers, and agent guidance are not served.
Keep credentials out of Git. Edit shared CSS in source; normal deployment
publishes committed CSS and preserves live Blog posts and cover photos.

## Maintain the guides

`site/docs.html` is Getting Started at `/docs`. Other guides live in `site/docs/`.
`itsy/data/guides.json` defines the reading order and chapter links for the shared
accordion in `site/_guides-nav.html`. Each guide ends with one link to the next;
the last returns to Getting Started. Keep those links and chapter IDs aligned.
`site/js/guides.js` opens the current guide and marks the current chapter.

Keep Getting Started a small, working tutorial. Introduce each concept once,
show the action and its result, and favor the normal path. Give each new concept
enough explanation for a reader to understand why they are using it; avoid
reducing the guide to commands and clipped instructions. Put optional features
in their own guides and leave uncommon command options to CLI help. Keep the
advanced app contract precise without repeating it in beginner guides.

This approach draws on SvelteKit's
[command-first setup](https://svelte.dev/docs/kit/creating-a-project), Rails'
[continuous tutorial](https://guides.rubyonrails.org/getting_started.html), and
Stripe's [task-focused quickstart](https://docs.stripe.com/checkout/quickstart).

Page titles and descriptions live in one leading `<!-- itsy` settings comment
in each page, alongside an optional `layout`. Values are single-line plain text;
empty titles fall back to the route title and empty descriptions stay empty.
Encode literal template braces in HTML code examples so they display instead of executing. Label fenced
examples with the matching Prism language. The local Prism 1.30.0 bundles and
MIT license live in `site/js/`; code colors live in `site/css/docs-code.css`.
There is no frontend build step.

After edits, run `itsy status`, review links and examples, and preview the changed
pages. Commit scoped changes before deploying; this nested site cannot offer an
automatic Git commit. Preserve both saved deployment mappings in the config.
Normal deployment keeps live Blog posts and cover photos. See the Blog guide
for intentional sync with `itsy pull` and `itsy deploy --with-content`.

## Website assets

Keep the approved 8-bit wordmark, light variant and Itsy favicon in `site/images/`.
Their source and production record live in `design/8-bit-wordmark/`; the selected
concept is in `design/logo-exploration-20261114/`. The earlier woven-loop identity
is retained as historical artwork, not the current logo.
The SVGs need no fonts or build step.

The sharing image is `site/images/social-preview.png`. To regenerate it, render
`design/social-preview.html` in Chrome at 1200 × 630 pixels with device scale 1.
The layout supplies each page's sharing title, description, and canonical URL;
keep the image URL absolute.

For deployment on machines with limited temporary storage, follow the private,
disk-backed per-job pattern in
[development storage](../system/docs/development-storage.md).
