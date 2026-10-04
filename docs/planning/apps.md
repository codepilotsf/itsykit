# Flea mini apps — design handoff

Status: design only, not implemented
Date: 2026-10-03
Scope: a proposed "apps" capability for Flea

## Current decisions (October 3, supersedes older sections below)

Related planning: [CodePilot service](service.md). Both documents remain design
work, not authorization to implement. Service workflow details are being worked
out separately; Rails is not the default backend for Flea apps.

- Admin login and the editor remain in core. Disregard the older identity
  extraction proposals and phase below.
- Apps are independent JSON services, normally Go binaries, with a documented
  API for agents writing site-specific templates. Remote APIs can use the same
  contract; incompatible services may need an adapter.
- Prefer explicit template directives such as `{layout main}` and
  `{app blog.list}` over settings comments. Exact grammar remains proposed.
  Responses bind to `blog.*`, not the older `api.blog.*` examples below.
- Page bindings perform reads before rendering. Ordinary forms perform writes.
  HTTP methods and mounted routes select operations; no mandatory `/apps/`
  prefix, `/submit` suffix, hidden `_action`, or custom action attribute.
- An app has a configurable mount path. Contact mounted at `/contact` receives
  `POST /contact`. Success redirects with 303 to `/contact/thanks`; validation
  failure renders the contact page with values/errors and status 422. Service
  failure renders a generic error with status 503.
- Destinations are fixed in app behavior and documented in its README, not
  programmed in site configuration. They are relative to the mount: contact at
  `/deeply/nested/route/contact` redirects to
  `/deeply/nested/route/contact/thanks`. Flea resolves against the mount explicitly,
  independent of browser trailing-slash behavior. Exact response envelope and
  mount registration remain open.
- Prefer flat durable storage: `storage/snips.json`, `storage/blog.json`,
  `storage/events.sqlite`, and shared `storage/media/`. App directories are
  optional when needed. Whether storage is at project root or under `flea/`
  remains open. No existing files have moved.
- SQLite WAL/SHM companion files belong to the same database. Exclude databases
  and companions from Git; back up through consistent SQLite snapshots.
- Shared media should support app references and descriptive public filenames,
  while retaining stable identities. Collision, rename, replacement, reference
  tracking, and media Git rules still need design. Never automatically prune.
- Remote content becomes authoritative after initial deployment. Later normal
  deployments preserve it; local content edits begin with a fresh pull. Content
  belongs in Git except databases and private state. Conflict handling, Git
  recording, and staging promotion are covered by the separate service planning.
- A new content API/revision system is deferred, not an app prerequisite.
- First prove contact, then blog reads/storage, then events or booking. Admin
  extraction is no longer a phase. Process supervision, packaging, dynamic route
  grammar, and ZJAX changes below remain proposals requiring verification.

The remainder preserves the earlier design sketch for reference. Its "Decided"
label does not override this section. Reconcile it when drafting the concrete
contact contract; do not implement superseded examples.

This document captures a design discussion. Nothing here exists yet. Where a
decision is marked **Decided** it was agreed in the discussion; where marked
**Open** it still needs a call. Read it before starting app work.

## Goal

Add a "mini apps" capability to Flea: optional small backend services that a
site's templates can read from and submit to.

Motivating features, in the order they came up:

1. Move the contact form out of the primary server.
2. Move admin/owner sign-in out of the primary server (the hardest; see Identity).
3. Blog.
4. Events calendar.
5. Micro e-commerce (Stripe).

Constraints carried over from how Flea works today:

- Templates stay plain, mostly static HTML/CSS.
- No npm and no frontend build step.
- Frontend and backend stay separable; an app is a standalone service.
- Prefer server-side rendering and reuse the existing Flea template language
  instead of adding a client template framework.
- No client reactivity unless a concrete widget forces it.

Note: this direction intentionally breaks Flea's current front-page promise of
"no database, no server-side application code." That is a real identity change
and should be stated as such, not slipped in.

## Mental model

> An app is a local JSON service; the Flea server is the only thing that knows
> both the app and the templates. Templates read `api.<app>.*`; forms submit
> through Flea; Flea blends the app's JSON into the rendered page.

Three roles, only one of which knows the other two:

- **Template** (Flea, server-rendered): declares which app it uses, reads JSON
  under `api.<app>.*`, posts forms to a site URL.
- **App**: a small local HTTP+JSON service (Go by default; anything that speaks
  HTTP+JSON works). No HTML, no knowledge of Flea. Owns its own storage.
- **Flea server**: the adapter. On each request it calls the app, binds the JSON
  to the template, and renders HTML. On a form POST it forwards the submission
  and re-renders with the result.

The browser talks to Flea in HTML and forms; Flea talks to the app in JSON. That
split is the whole point: no client templating, and zjax needs no JSON handling.

## Decided

- **Server-side rendering, live per request.** The main server calls the app on
  each render. No caching initially.
- **Whole page is the swap payload.** No partial-only rendering and no content
  negotiation. (Flea calls reusable markup *partials*; for app swaps we return
  the full page and let zjax extract the target region from it.)
- **Apps are separate processes, one per site**, each with its own storage
  (none, JSON files, or SQLite).
- **Hybrid supervision.** `flea dev` starts a site's apps as child processes. On
  the server, Flea writes a systemd unit per app during setup/deploy; systemd owns
  restart and health; Flea's server only calls the configured loopback port. Flea
  is a caller, not a supervisor.
- **Distribution: embedded, added on demand.** Default apps ship inside the
  `flea` binary (same pattern as starter templates and server bundles).
  `flea app add <name>` extracts `flea/apps/<name>/`; `flea app remove <name>`
  reverses it.
- **Developer writes the site templates.** Apps are JSON-only and do not ship
  templates.
- **Dynamic URLs use bracket segments**: `site/blog/[slug].html`,
  `site/blog/[...path].html`. Rendered on demand; no build-time enumeration for
  now (no Astro `getStaticPaths` equivalent). Exact files win over patterns.
- **App name is decoupled from the route.** An app named `blog` can serve
  `/articles`. Recommended mechanism: bind by app + operation
  (`api: blog#list`, `api: blog#show`) with operations declared in the manifest.
- **Two identity domains.** App end-users (blog readers, shop customers) are
  entirely the app's business. Owners/admins must still be recognized by the
  main server on every page.
- **Uniform write envelope** across apps: `values`, `errors`, `form_error`,
  `sent`. Read shapes are app-specific.

## The contract

The contract is the JSON a template reads, plus two wiring rules.

**Binding.** A page declares the app and operation in its settings comment:

```html
<!-- flea
layout: /main.layout.html
api: blog#list
-->
```

**Read.** Flea calls the named operation, gets JSON, and exposes it as
`api.<app>`:

```html
<section class="container">
  <h1>Articles</h1>
  {#each api.blog.items as post}
    <article>
      <h2><a href="/articles/{post.slug}">{post.title}</a></h2>
      <time datetime="{post.published_at}">{post.published_label}</time>
    </article>
  {/each}
</section>
```

**Write.** A form posts to a site URL; Flea forwards the fields to the app's
operation and re-renders the same template from the returned JSON:

```html
{#if api.contact.sent}
  <div class="notice" role="status"><p>Thanks — we’ll reply to {api.contact.email}.</p></div>
{:else}
  <form method="post" action="/contact" novalidate>
    {#if api.contact.form_error}
      <div class="notice notice-error" role="alert"><p>{api.contact.form_error}</p></div>
    {/if}
    <label class="form-field">Email
      <input type="email" name="email" value="{api.contact.values.email}" />
    </label>
    {#if api.contact.errors.email}
      <p class="field-error">{api.contact.errors.email}</p>
    {/if}
    <button type="submit">Send</button>
  </form>
{/if}
```

Server behavior per outcome (this is what the zjax change below enables):

| Request        | Outcome    | Status               | Body                       |
| -------------- | ---------- | -------------------- | -------------------------- |
| Form POST      | success    | `303 -> /contact?sent=1` | redirect (fetch follows)   |
| Form POST      | validation | `422`                | full page, errors rendered |
| Form POST      | app down   | `503`                | full page, generic error   |

**Manifest sketch** (proposed, not final):

```jsonc
// flea/apps/blog/app.json
{
  "name": "blog",
  "version": "0.1.0",
  "protocol": 1,
  "command": "./blog",
  "operations": {
    "list":    { "method": "GET",  "view": { "items": [ { "slug": "string", "title": "string" } ] } },
    "show":    { "method": "GET",  "params": ["slug"], "view": { "post": { "title": "string", "body_html": "string" } } },
    "comment": { "method": "POST", "fields": ["name", "email", "body"] }
  },
  "storage": "sqlite"
}
```

## zjax change (required)

Flea's form flow returns honest non-2xx statuses with a renderable page. zjax
currently refuses to swap them, which forces servers to lie with `200`.

```js
// src/lib/swaps.js, getSwapFunction
if (!response.ok) {
  const $ = getDollar(trigger.node, event, response);
  handleSwapError($, trigger);
  return;
}
```

**Fix (new default, not opt-in):** stop gating on `response.ok`. Parse whatever
body comes back; swap when the response/target node exists, and fall back to
`handleSwapError` only when there is no usable node (non-HTML body, missing
selector, network error). Expose the status to a hook so a consumer can still act
on `401`/`403` (redirect to sign-in).

Why safe: generic `500` crash pages and proxy `502` pages do not contain the
target selector, so they still fall through to `handleSwapError`. Leave
`swapType`/`responseType` behavior unchanged.

Tests: `200` with target -> swaps; `422` with target -> swaps; `422` without
target -> error; `500` generic -> error; non-HTML `4xx` -> error; network
failure -> error.

Ship as a major zjax release (changes default behavior). zjax repo:
`/Users/sean/Projects/zjax`.

## Directory and distribution

Proposed shape for an app in a project:

```
flea/apps/
  blog/
    app.json      # name, version, protocol, command, operations, storage
    blog          # the executable (prebuilt for default apps)
    README.md     # field reference + copy-paste example
    data/         # the app's storage (JSON or SQLite) — Git-ignored   [Open]
```

A bespoke app is the same shape: you write the service (Go by default) and set
`app.json`. Flea only needs the manifest to run it and bind it.

Current repo state for reference (protocol-pinned today):

```
site/                 public views + assets
flea/snips.json       editable content + page metadata (source of truth, committed)
flea/data/            developer JSON (data.*)
flea/media/           managed photos/videos
flea/icons/           Lucide cache
flea/skills/          agent guidance
flea/runtime/         private runtime state
mailers/              private email templates
.flea.config.json     settings + deploy
```

## Storage and lifecycle (**Open**)

Three kinds of state, with different Git/deploy rules:

1. Committed + deployed: `site/`, `flea/snips.json`, `flea/data/`, `flea/icons/`, `mailers/`.
2. Ignored + preserved + backed up: app DBs/uploads (and `flea/media/` video bytes).
3. Ignored + ephemeral: sessions, locks (`flea/runtime/`).

Open:

- Do not move `flea/snips.json`: it is wired into the template tooling and the
  server protocol, so moving it is a protocol/version change plus migration.
- Where does app data live: colocated `flea/apps/<app>/data/` (one dir per app)
  or top-level `storage/<app>/` (code and state separated)?
- Is app data Flea-managed runtime state (Git-ignored, preserved across deploys,
  included in `flea backup`/`restore`) or fully opaque to Flea?
- Seed vs live: some app data is developer-authored and committed (default
  categories), some is admin-created (posts, orders). Flea already has this split
  for content; decide whether apps inherit it.
- Directory restructure: incremental (leave `snips.json`; choose an app-data
  home) vs a deliberate v2 (`/storage/` holds all persistent state, `site/` is
  views, apps are code).

## Identity (**Open**)

Two domains:

- App end-users (blog readers, shop customers) are the app's own concern; the
  main server never needs to know them.
- Owners/admins must be recognized by the main server on every page to show the
  editor and gate `/admin`. Today the main server owns that store
  (`flea/runtime/sessions.json`). Moving sign-in into an app raises: who owns
  owner sessions, and how does the main server trust a login the app performed?

Options discussed:

- App owns sessions; main server verifies a signed cookie locally with a shared
  secret (fast; revocation needs a shared story).
- App owns sessions; main server calls the app's verify endpoint per request (one
  source of truth; hard dependency on every page).
- Shared store (e.g. the site's SQLite or the lock-guarded file); main server
  keeps gating, app performs login.

Because the first planned app is sign-in, this is the hardest possible first
target. Consider proving the runtime on the contact form first (Phase 0).

## Open questions

1. Directory: incremental vs v2 restructure (see Storage).
2. App data: location, lifecycle, and seed-vs-live (see Storage).
3. Owner identity ownership and trust mechanism (see Identity).
4. Manifest contents and whether it doubles as the source for load-time
   validation of template field references (like snips are validated today).
5. Protocol/version compatibility between Flea and apps; migration rules.
6. Port allocation: fixed vs assigned range; uniqueness per (site, app) on a host.
7. Server <-> app trust: shared secret/bearer, loopback binding, timeouts.
8. Confirm the operation binding syntax (`api: app#operation`) and how route
   params are passed.
9. Dev vs prod parity, logging, and observed app health.
10. Deploy/backup of app code and data.
11. Whether apps may ever be shared across sites (currently per-site).
12. E-commerce/Stripe specifics (deferred).

## Suggested phasing

- **Phase 0 — prove the loop.** Contract + `flea app add` + rewrite the contact
  form as an app. No identity, no dynamic routes, no storage. Lowest risk.
- **Phase 1 — identity.** Move owner sign-in. Security-critical; settles the
  session-trust design.
- **Phase 2 — blog.** First app with storage and bracket dynamic routes.
- **Phase 3 — events calendar.** Stateful UI; likely the first place client
  reactivity is justified.
- **Phase 4 — micro e-commerce.** Stripe; deferred.

## References

- zjax library: `/Users/sean/Projects/zjax`
- Flea architecture: `ARCHITECTURE.md`, `spec.md`, `next-steps.md`
- Flea skill references (authoring, commands, deployment): installed at
  `~/.config/opencode/skills/fleacms/references/`
