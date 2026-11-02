# Styling and icons

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

### Icons

Use `{icon arrow-right}` or `{icon search size="20" class="text-dark"}`.
Names are literal lowercase kebab-case; attributes are quoted HTML-style values.
Dotted-path values inside attribute quotes inherit loop aliases and use Go's
contextual escaping. Defaults are 24px, currentColor stroke, and aria-hidden;
aria-label/aria-labelledby gives the icon role img. Supported attributes and
limitations are listed in README.md. No event handlers, style, href, or raw HTML.

Do not embed the Lucide library or download it wholesale. `LoadDev` fetches only
missing referenced icons from pinned lucide-static 1.47.0 at unpkg before serving,
including on reload. `Load`/status stays read-only and never accesses the network.
Missing icons produce deduplicated `Project.Warnings` and render nothing.
Download/network failures are also warnings; status exits 0 when otherwise valid.
Print warnings at startup/reload, not on every request. Invalid directives and
unsafe or malformed downloaded/cached SVGs remain validation errors.
Cache validated SVG files directly as `flea/icons/name.svg`. New sites include
`flea/icons/.keep`; keep this directory in Git. Embed the pinned upstream license
notice in each SVG comment, with no version subdirectory or separate LICENSE.
The upstream version stays pinned in Go. Cached icons work offline.
Downloads are bounded, cancellable, and atomically saved via confined handles.
Never fetch icons inside request handlers. Reject symlinks and reconstruct SVG
using the restricted geometry vocabulary before insertion into template source.
Deploy also uses LoadDev to resolve icons before packaging. No standalone site
build command exists yet; a future build should reuse this resolver.
Tests use an injected HTTP transport, not live CDN requests.
The editor embeds only grip-vertical, x and plus from the same pinned Lucide
release with its license. Compile them through the shared {icon ...} parser and
SVG sanitizer into reusable toolbar templates, not copied paths in JavaScript.
The six visible formatting icons use ordinary {icon ...} tags in the toolbar
partial and ship as cached starter SVGs. Do not insert or replace them with JS.

### CSS

Keep boilerplate pages restrained: theme-driven typography and colors, readable
intro widths, and consistent native form controls. Owner access screens share
the narrow `admin-access` layout; keep the code field's native autofill semantics
and six-box enhancement unchanged. The standalone 500 page has self-contained
fallback styling so an asset failure cannot leave it unstyled.

The native mobile popover sits below `--flea-toolbar-height` (zero for visitors).
The editor publishes that measured height on :root and updates it on resize.
Keep both the menu and backdrop below it. Use CSS entry/exit transitions with
discrete display/overlay handling and immediate reduced-motion behavior.
The hamburger is three 30px CSS lines in a borderless 48px touch target.

Default to the Prewind method; read `../prewind/README.md` for its philosophy
and supported classes. Keep HTML readable:

- Global defaults (typography, links, forms) belong in shared CSS.
- Use Prewind classes for one-off layout, spacing, and color adjustments.
  Prefer theme tokens and fluid sizes such as `gap-md` and `p-lg`; do not
  invent custom classes that merely duplicate utilities.
- When styled markup repeats, consider a partial to reuse the markup and
  its utility classes. Keep the starter header/footer inline as specified above.
- Occasionally use inline styles for genuinely arbitrary one-off values,
  such as `style="width: 37px"`; do not invent Tailwind bracket utilities.
- Write shared CSS for behavior Prewind does not cover, such as complex grids
  and animations. Do not assume Tailwind classes or variants exist in Prewind.

Use Prewind's `container` class for inner page content by default, giving it
consistent centered width and theme-controlled padding. For full-width sections,
put `container` on an inner wrapper. Do not add `max-w-full` alongside container;
container already handles its width. Avoid wrappers duplicating its purpose.
Use global textarea `field-sizing: content`, a one-field minimum and 400px maximum
height with overflow scrolling. Do not add fixed rows or autosizing JavaScript.

The starter loads Prewind 1.3.1 from unpkg, based on the Preflight template's
global-first CSS conventions. No Node build step or Rails dependency is required.
The main stylesheet is `css/global.css`, with layer order `reset, styles,
prewind`. It imports Prewind, theme tokens, typography, native form styles, and
button styles. Put global rules in `@layer styles` so utilities can override them.
The starter `theme.css` imports Inter for both headings and body text,
with display=swap and system sans-serif fallbacks. Use warm off-white page
backgrounds, white form/card surfaces, charcoal text and forest-green primary
actions. Keep these colors in theme tokens, not page-specific overrides. Customize
colors, fonts, and fluid scales in `theme.css`; site rules belong in `global.css`.
Form styling does not implement submission handling.

Keep the vendored `prewind-1.3.1.css` unchanged; customize the other files.
The theme and general styles are adapted from `../preflight/lib/templates/app/assets/stylesheets/`.
The library matches `../prewind/prewind.css` at version 1.3.1. Preserve the simple
starter with its inline header/footer when making styling changes.

### FleaCMS identity

The FleaCMS lockup uses the approved Little Skip mascot and outlined Nunito 800.
Use `images/logo.svg` for branded headers and `images/mark.svg` for the mascot.
The outlined lettering needs no runtime font request. Keep normal site body and
heading typography unchanged. Canonical masters and export notes live in
[branding](../branding/README.md); do not redraw the mascot or retune its spacing.
