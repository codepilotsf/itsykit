# FleaCMS homepage direction

Updated September 29, 2026. Two-page structure agreed; visual design pending review.

## Two pages only

Use Sean's [zjax.dev](https://www.zjax.dev/) as the structural reference:
a direct introduction, a product demonstration, then short explanations paired
with code. Its [docs](https://www.zjax.dev/docs) are a single long-form reference.
Keep Flea similarly compact, but use a fresh, bright visual style rather than
Zjax's dark background. White surfaces, generous whitespace, crisp charcoal
type and vivid green accents; avoid making this another muted starter site.

The only public content pages are `/` and `/docs`. Put installation, quickstart,
template syntax, admin editing, deployment, email, recovery, updates, limitations,
release notes and support in anchored sections of `/docs`. Use a sticky contents
list on desktop and a compact native disclosure on mobile. No separate Features,
FAQ, Contact or Getting started page. Private authentication and error templates
remain infrastructure, not additional public content pages.

## References

- [Clerk](https://mobbin.com/screens/b57c3420-a006-4fe1-b903-4ca0ca658a5f):
  concise navigation, strong headline hierarchy and a clearly prioritized action.
- [Greptile](https://mobbin.com/screens/ddd061c9-6056-4cd2-a671-4f433c84f2de):
  confident left-aligned typography and room for a distinctive product visual.
  Do not copy its illustration, decorative texture or social-proof claims.

## Proposed composition

Preserve Skip, the selected flea mark. Use warm off-white, charcoal and forest
green, with Inter for prose and monospace only for code. Give the marketing site
its own composition rather than presenting a slightly modified starter site.

Header: logo linking home, Docs, Get started linking to `/docs#installation`.
Footer release/support links point to docs sections. Two navigation destinations
do not require a hamburger. Respect reduced-motion preferences.

Hero headline: “Write the HTML. Let them handle the edits.”
Supporting copy: “A small CMS for developers who build with plain files, and
people who just want to update their website.”
Actions: Get started and See how editing works.

Follow immediately with an actual editor screenshot or recording, showing a
small example site. No simulated customer testimonials or invented usage counts.
Below it, explain the developer/admin boundary with one short HTML example and
the resulting editable content. End with the new/dev/deploy workflow and an
honest summary of hosting, email and platform requirements.

Public installation is not verified yet. Do not present a working download CTA
or encourage execution of the public installer until artifacts and metadata have
been published and tested. The design can reserve an installation section.

## Scope and next checkpoint

Build and review the homepage locally first, at desktop and mobile widths.
Extend the approved direction to the single docs page after review. Deployment requires a confirmed destination;
do not infer permission to replace the apex site from Bookworm deployments.

The requested design skill is not available in this session. Use inspected Mobbin
references and established Prewind conventions, without a new frontend framework.
The existing scaffold contains stale Homebrew, Resend onboarding, collection UI
and media-status descriptions. Reconcile them with current framework behavior
when those pages are rewritten. Keep legacy-looking sibling directories intact
until their contents and removal scope have been reviewed.
