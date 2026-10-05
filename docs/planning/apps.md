# Flea mini apps: design handoff

Status: managed Contact and protocol/CSRF implemented locally; broader apps remain planned
Date: 2026-10-04
Scope: the included Contact app and future Flea mini apps

## Current decisions (October 4, supersedes older sections below)

Related planning: [CodePilot service](service.md). Both documents remain design
work, except for the approved foundation and managed Contact batches below.
Current supported behavior is in [Contact's README](../../apps/contact/README.md).
That README supersedes proposed launch/email details in older sections here.
Service workflow details are being worked
out separately; Rails is not the default backend for Flea apps.

- Apps extend a mostly static website. Flea handles routing, form forwarding,
  and rendering; it is not a Rails replacement or a framework for arbitrary
  application workflows. Complex forms are a reason to reconsider an app's
  scope before expanding the core contract.
- Admin login and the editor remain in core. Disregard the older identity
  extraction proposals and phase below.
- Apps are independent JSON services, normally Go binaries, with a documented
  API for agents writing site-specific templates. Remote APIs can use the same
  contract; incompatible services may need an adapter.
- Prefer explicit template directives such as `{layout main}` over settings
  comments. Exact grammar remains proposed. App mounts are declared in site
  configuration, for example `"apps": {"blog": {"mount": "/articles"}}`.
  The mount connects routes to the app; no `{app ...}` directive is needed.
  Reading app data on unrelated pages remains a separate open question.
- Responses bind under the app name and retain the API envelope: ordinary
  values use `blog.data.*`, not flattened `blog.*` or older `api.blog.*` paths.
- Page bindings perform reads before rendering. Ordinary forms perform writes.
  HTTP methods and mounted routes select operations; no mandatory `/apps/`
  prefix, `/submit` suffix, hidden `_action`, or custom action attribute.
- An app has a configurable mount path. Contact mounted at `/contact` receives
  `POST /contact`. Success redirects with 303 to `/contact/thank-you`; validation
  failure renders the contact page with values/errors and status 422. Service
  failure renders a generic error with status 503.
- Destinations are fixed in app behavior and documented in its README, not
  programmed in site configuration. They are relative to the mount: contact at
  `/deeply/nested/route/contact` redirects to
  `/deeply/nested/route/contact/thank-you`. Flea resolves against the mount explicitly,
  independent of browser trailing-slash behavior. See the response contract
  below; remaining registration and transport details are still open.
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
  extraction is no longer a phase. See the [Contact handoff](contact-app.md).
  Packaging, dynamic route grammar, and ZJAX changes remain open.

### Packages, configuration, and lifecycle

- Included apps and their optional Go helper package live in this repository,
  with tests covering both sides of the contract. Source layout and distribution
  packaging are not yet fixed. Inclusion does not enable every app on every site.
- Packages may be discovered automatically, but only entries in the main
  `.flea.config.json` enable them and assign public mounts. Discovery alone must
  not execute code or register routes.
- A per-app `app.json` describes the package, not site-specific routing. The
  agreed fields are `name`, app release `version`, and integer `protocol`.
  Executable declaration details remain proposed.
- The contract is named **Flea App Protocol**: for example, Flea App Protocol v1
  is declared as `"protocol": 1`. App release versions are separate. Verify the
  running service's protocol, not just the manifest; reject mismatches clearly,
  without compatibility fallbacks. See the authenticated handshake below.
- Flea manages local apps as child processes in development and production,
  passing an already-open private listener and storage location. It owns startup,
  shutdown, bounded crash-restart backoff, and useful redacted logs. This
  supersedes the historical proposal for a systemd unit per app.
- Unavailable apps return a friendly 503 on affected routes; unrelated site
  pages remain available. Migrations do not run during crash recovery.
- Separately managed services, local or remote, use a configured `url` and are
  not supervised by Flea. See the connection contract below. Readiness timing,
  resource limits, and process cleanup still need design. Child processes are
  not security sandboxes. Always-running process policy remains subject to
  measuring real app overhead; the static-site benchmark did not test apps.
- Go helpers are optional: use standard HTTP routing and small handlers, with
  helpers for request decoding, response envelopes, and service startup. Do not
  create a config-driven business-logic language or require an ORM. Start with
  contact and extract helpers justified by actual use.
- Flea retains email credentials, rendering, delivery, and development previews.
  Contact validates submissions and calls a narrow private capability to send
  this site's contact notification, not arbitrary mail. Authorization is scoped
  to the app and site. Capability transport and remote-service access remain
  open; see the [contact README](contact-app.md#email-settings-and-security).
- App email settings use `apps.<name>.email` in the site's main config, never
  `mail`. An optional `from` overrides the server-wide default sender for that
  site's app; it must be provider-authorized. Contact requires an explicit `to`
  recipient. Sender selection comes from trusted configuration, not request
  JSON. Provider credentials remain separate from site configuration.

### Connections and local launch contract

Two modes share the same Flea App Protocol. These settings are proposed, not
supported by the current server.

**Managed package:** the site's main `.flea.config.json` enables an installed
app with a mount, without a service URL or port:

```json
{
  "apps": {
    "contact": {"mount": "/contact"}
  }
}
```

Flea asks the OS for an available port by binding to `127.0.0.1:0`. Port zero
means automatic assignment, not a literal destination port or custom `:auto`
syntax. Flea keeps the listener open and passes it to the child; it must not
probe a free port, close it, and race to reopen it. The Go helper accepts that
inherited listener. Flea records its address and completes an authenticated
identity/protocol handshake before forwarding app requests.

On shutdown, both Flea and the child must release their listener references.
A restart may receive a new port without changing the public mount. Exact
descriptor handoff, launch arguments, and credential delivery remain to be
specified. This local launch contract is separate from the HTTP/JSON protocol.

**External service:** a URL in the same site's main config selects a service
whose process and port are managed independently:

```json
{
  "apps": {
    "custom": {
      "mount": "/custom",
      "url": "http://127.0.0.1:5000"
    }
  }
}
```

For example, a Python/Flask service can implement the protocol without Go helpers
or listener inheritance. It binds its own port; an occupied port causes its
startup to fail. Flea verifies the expected app identity and supported protocol,
not merely that something accepts connections. It never starts, stops, or
reassigns ports for external services.

For a service on another machine, use HTTPS:

```json
{
  "apps": {
    "custom": {
      "mount": "/custom",
      "url": "https://apps.example.com:5555"
    }
  }
}
```

If this is foobar.com's config, requests to `https://foobar.com/custom` go
through Flea to that endpoint. `apps.example.com` can resolve to the remote
server's IP. A hostname simplifies certificate management; an IP URL would
still need a valid certificate for that IP. Response values appear under
`custom.data.*`. Mount stripping and query forwarding follow the same rules as
managed apps. The service address and credentials are never sent to the browser.

### Connection security

- Browser-to-Flea traffic uses production HTTPS. Flea-to-remote-app traffic
  also requires HTTPS with certificate verification.
- Local calls use ordinary HTTP on loopback. Documentation may call this
  "localhost HTTP", but the managed listener and caller use `127.0.0.1`
  explicitly to avoid IPv4/IPv6 resolution mismatches. Never bind managed apps
  to `0.0.0.0`. Localhost is not an authentication boundary.
- Flea generates a random credential scoped to one managed app and site for
  authenticated calls/handshake. The app receives a separate credential for
  any authorized callback capability, such as sending a contact notification.
- External services use explicitly provisioned credentials outside committed
  configuration. For a multi-site service, its credential must identify and
  authorize the site, not a visitor-supplied domain or site identifier.
- Keep credentials out of URLs, logs, public assets, Git, releases, and backups.
  Exact secret storage, launch-time injection, and rotation remain open.
  Never infer compatibility from a manifest
  alone or fall back to an unauthenticated service.
- This assumes a trusted server without untrusted tenants running arbitrary
  code. Loopback plus tokens is not isolation from privileged local processes.
- Remote callback access to Flea's private capabilities remains unresolved.
  Registering a remote app does not expose a public mail endpoint automatically.

### Authenticated handshake

Before forwarding submissions or data requests, Flea calls the service's
private `GET /_flea/handshake` endpoint with its app/site-scoped credential:

```http
GET /_flea/handshake
Authorization: Bearer <secret-token>
```

The app authenticates the request before returning HTTP 200 with JSON:

```json
{
  "name": "contact",
  "version": "0.1.0",
  "protocol": 1
}
```

Flea checks the expected app name and supported protocol. `version` identifies
the app release; it need not equal Flea's version. An open port or a readable
manifest is not sufficient. A missing or incorrect credential denies access
before any app operation. The local Go helper handles the authentication checks;
external services implement the same contract.

The same authorization header accompanies subsequent Flea-to-app requests.
The token is a long, cryptographically random secret: possession grants its
defined access, not proof of a human identity. Flea supplies managed apps with
the secret privately during launch. External services require matching credentials
provisioned on both sides. Browser authorization headers must not be passed
through as app credentials.

The handshake is service-internal and must never be forwarded through a public
mount, including a visitor request to `/contact/_flea/handshake`. Callback mail
access uses a separate token with narrower permissions, not this service token.
Handshake timing, response limits, and exact rejection statuses still need to
be specified before implementation.

### Human-facing setup and errors

The normal managed-app setup is to enable Contact, choose its public page and
recipient, and let Flea handle listener assignment, secrets, and connection
checks. This describes the intended experience, not an implemented setup UI.
Keep protocol mechanics out of routine prompts while retaining useful, redacted
diagnostics for developers.

Explain what failed and the next action. For an app whose newer protocol is
unsupported, say "Contact couldn't start because it needs a newer version of
Flea." Do not use that explanation for every mismatch: wrong app identity,
rejected credentials, and an unavailable service need their own accurate messages.
Never expose tokens or internal service details to public website visitors.

### Optional SQLite schema changes

Apps own their schema and numbered SQL migration files, for example
`migrations/001_create_events.sql` and `002_add_location.sql`. Applied files stay
unchanged; subsequent schema changes get new files. New databases run them all.
A small runner tracks applied files, executes each migration in a transaction,
and records success in that same transaction. Failure stops the update.

Production migrations are an explicit app-update step after a consistent
database backup, not a side effect of ordinary startup or automatic restart.
No automatic reverse migrations initially. Update locking, interrupted-update
recovery, and app/schema compatibility checks remain to be designed. Migration
SQL belongs in Git; databases do not. Contact requires neither.

### Form request contract

Flea converts ordinary form submissions to a flat JSON object and forwards it
to the registered app. There is no request `data` wrapper:

```html
<input name="email">
<textarea name="message"></textarea>
```

```json
{
  "email": "sean@example.com",
  "message": "Hello"
}
```

- A name submitted once has a string value. Repeated names have an array of
  strings. The app handles type conversion and validation.
- Missing fields stay absent, including unchecked checkboxes. A submitted empty
  field remains an empty string.
- Names are literal: `address.city` or `sessions[a].date` remain exactly those
  JSON keys. Flea does not interpret dots, brackets, or indexes as nested data.
- Flea validates and removes its `_csrf` field before forwarding. Existing
  Origin, host isolation, and other applicable security checks remain required.
- File uploads need a separate contract and are deferred; do not encode files
  into this JSON contract.

An app can interpret its own documented field names when needed. Shared nested
field parsing is not part of the initial scope; revisit only if concrete mini
apps repeatedly need it.

### Site globals and CSRF rendering

Keep `app.*` for site globals, with explicitly documented Flea-owned names
such as existing `app.current_admin`. Reject configuration that defines a
reserved name; never silently override it. CSRF adds no reserved global or
public raw-token binding. This is separate from an installed app's response
namespace, such as `contact.data.*`.

Authors explicitly include the CSRF control in forms:

```html
<form method="post" action="/contact">
  {csrf}
  <!-- App fields -->
</form>
```

The `{csrf}` directive renders the complete hidden input named `_csrf`,
with its request-local token safely escaped in the value attribute. It is an
HTML markup directive, not a data expression; permit it only in element-content
contexts, never within attributes, scripts, or styles. There is no public
`app.csrf` token binding. This supersedes the earlier raw-token proposal.

Flea does not scan forms and inject hidden fields automatically. The token is
not stored site content or app-service response data. Rendering must not
mutate the published Project or share a visitor's token through cached HTML.
Flea checks the token and exact Origin before forwarding a submission; missing
or invalid tokens are rejected without calling the app. Strip `_csrf` from the
forwarded JSON. The field name belongs to Flea, not an app's business data.

This lifecycle is implemented. Contact now requires `{csrf}`; admin forms keep
their existing controls. Public Contact mounts are enabled only by apps.contact.

### First implementation batch (approved October 4)

Implemented in `internal/formguard`, `internal/appapi` and the site renderer.
The historical scope below records the first batch, before Contact replacement.
The second approved batch now adds public Contact dispatch and managed lifecycle;
see its status below. Verification is not inferred from a checklist.

Before replacement, code review on October 4 found that the old contact handler
enforces exact Origin, HTTPS, bounded URL-encoded bodies and submission limits,
but has no visitor CSRF token. Editor login uses a cookie/form comparison and
authenticated editing uses persisted session tokens (`internal/editor/auth.go`).
Do not make public forms depend on admin sessions or change those flows.
`internal/site/server.go` renders compiled templates per request and already
defaults responses to no-store; it does not retain finished page HTML.

Build the security and protocol foundation first, using a test service to prove
the round trip. This batch does not replace the working contact handler or
advertise a ready-to-install app.

1. Add `{csrf}` compilation and request-local rendering, including layouts and
   partials. Preserve the rule that comments and raw script/style content are
   not evaluated. Reject directive use in unsupported HTML contexts. Loading,
   validation and status checks must not generate visitor cookies or secrets.
2. Add anonymous visitor protection independent of editor authentication.
   Design: a random visitor nonce in a signed, host-only HttpOnly
   cookie, and an HMAC-authenticated form token bound to that nonce, the site,
   exact origin and expiry. Use standard-library cryptography, unambiguous
   message encoding and constant-time verification, not an unsigned cookie
   comparison. Production cookies use Secure, Path=/, no Domain and a __Host-
   name; trusted HTTP dev uses a separate site-specific cookie name. SameSite=Lax
   avoids replacing a valid cookie on ordinary incoming top-level navigation.
   This follows the signed-token approach described in the
   [OWASP CSRF guidance](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html).
3. Start with a fixed 24-hour visitor/form lifetime. Reuse a valid visitor nonce
   rather than rotating on each render or submission; existing tabs and normal
   validation retries remain usable until expiry. Issue cookies only when the
   directive actually renders, and commit them only after successful rendering.
   No per-visitor server map or database. Keep a random per-site signing key in
   runtime memory, outside immutable Projects, retained across template reloads.
   A full server restart invalidates forms: return a clear 403 reload instruction,
   never automatically replay a submission. Restart persistence is deferred.
   Simultaneous first visits without a cookie can race; fail safely and test the
   reload recovery rather than promising uninterrupted first-use multi-tab forms.
4. Keep exact Origin, approved host/HTTPS and devaccess checks. Require exactly
   one cookie and one `_csrf` POST field, rejecting malformed, missing, duplicate
   or expired values before any app call. Never take the token from a query.
   Strip `_csrf` from forwarded JSON. Token-bearing HTML is private/no-store;
   ordinary pages must not acquire visitor cookies or per-visitor state.
5. Define tested internal protocol types and mount validation: reject reserved
   namespaces, core-route conflicts and overlapping app mounts. Use literal flat
   form fields, preserving repeated values. Initially accept bounded URL-encoded
   forms only; multipart and files remain unsupported. Contact's existing 32 KiB
   request bound is the starting default, not an unbounded generic decoder.
6. Exercise authenticated HTTP transport against a local test service: handshake,
   mount stripping, unchanged GET query parameters, form-to-JSON POST, and
   request-local `data`/`invalid`/`error` rendering. Never forward browser cookies
   or Authorization. Disable automatic upstream redirects and POST retries.
   Bounds: 2-second handshake, 15-second ordinary request deadline,
   4 KiB handshake and 1 MiB response bodies, at most four calls per client and
   64 process-wide, with no waiting queue. App HTML output is capped at 2 MiB
   during rendering. Malformed, oversized or unavailable upstream responses fail
   safely with 503; application validation uses 422. Production process launch
   and publicly enabled app routing are not part of this foundation batch.

Acceptance: regression tests cover token/cookie tampering, expiry, duplicate
fields, cross-host/site and visitor isolation, reload/restart behavior, concurrent
renders, context escaping, mount conflicts, upstream timeouts/size bounds and
unchanged admin/contact behavior. Run Go tests, race tests, vet, the bundled build
and a fresh generated-site status check. No deployment, live email or commits.

### Managed Contact batch (approved and implemented locally October 4)

Contact now has bundled child launch/cleanup, fixed render-only form/thank-you routes,
a private email capability, mount-relative redirects, visitor-scoped flash and
starter/config replacement. There are no legacy aliases or migration fallbacks.
The matching-platform executable is bundled before compiling each parent; the
site package does not carry redundant executables. Deployment protocol 23 prevents
old servers accepting the new site contract. See [Contact](../../apps/contact/README.md)
for inherited descriptors, credentials, bounds and configuration.

The [isolated Linux Contact check](../benchmarks/2026-10-04-contact-apps.md)
passed initial lifecycle tests and a 1/10/30-site memory probe. Real email,
longer soaks and representative app-capacity acceptance remain open.
General read-route declarations, custom/remote-service provisioning, databases,
storage moves and ZJAX remain separate work. No deployment has been performed.

### GET request contract

Flea forwards query parameters to the mounted app as query parameters, without
a JSON request body. For example, `/articles?page=2` calls the blog service with
`?page=2`. The app handles parsing, defaults, and validation; Flea does not convert
query values to numbers or interpret nested field names. Repeated parameters
remain repeated parameters.

The app returns the response envelope below, and Flea renders the page with
that data under the app namespace.

### Mount-relative service routes

Flea strips the configured public mount when calling the app service. The mount
itself maps to `/`; descendant paths keep their suffix. HTTP methods and query
parameters are preserved.

| Public request | App service request |
| --- | --- |
| `GET /articles` | `GET /` |
| `GET /articles?page=2` | `GET /?page=2` |
| `GET /articles/my-post` | `GET /my-post` |
| `POST /contact` | `POST /` |

Apps document routes relative to their own root, independent of the site's
mount depth. Match mounts on path-segment boundaries, not arbitrary string
prefixes: `/articles` must not capture `/articles-extra`. These are URL paths,
not filesystem paths.

### Response contract

The JSON-only API uses the same structure for reads and writes. Flea exposes
renderable response fields under the configured app name:

| API field | Shape and purpose | Template path |
| --- | --- | --- |
| `data` | Object containing app-owned values | `contact.data.email` |
| `invalid` | Object mapping field names to validation messages | `contact.invalid.email` |
| `error` | Single operational failure message, a string | `contact.error` |
| `flash` | Object containing temporary messages | `contact.flash.notice` |
| `redirect` | Mount-relative destination, handled by Flea as a server instruction | Not exposed as template data |

Only envelope keys are reserved. Names inside `data` belong to the app; there
is no `_meta` namespace. `invalid` replaces the proposed `validation` name,
and the single `error` string replaces the proposed `errors` array.

Validation failure example (HTTP 422):

```json
{
  "data": {"email": "not-an-email", "message": "Hello"},
  "invalid": {"email": "Enter a valid email address."}
}
```

Operational failure example (HTTP 503):

```json
{
  "error": "We couldn't send your message. Please try again."
}
```

Successful submission example (Flea emits HTTP 303):

```json
{
  "redirect": "thank-you",
  "flash": {"notice": "Your message was sent."}
}
```

Flea preserves declared redisplayable form fields on failure under `data`, not
passwords or files. Response data is request-local, never written into shared
published Projects or compiled as template source.

Flash is scoped to the visitor and app. It is available in an immediate render
or carried across a redirect for the next page render, then expires. Persistence
and exact consumption rules still need design; assets must not consume it.

Richer template conditions are a separate proposed enhancement, not a
prerequisite for this contract.

## Historical sketch

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
