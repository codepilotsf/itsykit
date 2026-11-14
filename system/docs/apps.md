# ItsyWeb apps

An app is a small, separate JSON service that a ItsyWeb site can mount at a public
URL. ItsyWeb owns the public page, the HTML form, the redirect, the template values
and the email. The app owns its own logic and validates its own fields. This
guide explains the model, the shipped Contact app, the exact JSON contract, and
how a developer can write a compatible service.

Contact ships with ItsyWeb and is enabled only by `apps.contact`. You can also
register your own **managed local app**: build an executable that speaks the
current ItsyWeb App Protocol (version 2), put a validated package under
`itsy/apps/<name>/`, and enable it with `apps.<name>.mount`. No ItsyWeb source
change is needed. This guide explains
the model, the exact JSON contract, the managed launch contract, app-owned
sessions, and how to write, build, mount and test a compatible service.

This document describes the shipped 0.17.0 implementation. Custom managed apps
and app-owned sessions are published in the public 0.17.0 CLI and server.
Deploying a site that carries a custom app package requires a matching server at
deployment protocol 26; the previous 0.16.1 server is protocol 25 and refuses it,
and a protocol-26 CLI refuses a protocol-25 server before upload.
Proposed future app work lives in [planning/apps.md](agents/planning/apps.md) and
must not be read as shipped behavior.

## Contents

1. [What an app is](#what-an-app-is)
2. [Support at a glance](#support-at-a-glance)
3. [Enable Contact](#enable-contact)
4. [The Contact form and templates](#the-contact-form-and-templates)
5. [Two boundaries: browser form and service call](#two-boundaries-browser-form-and-service-call)
6. [Request JSON](#request-json)
7. [Response envelope](#response-envelope)
8. [Status codes and browser outcomes](#status-codes-and-browser-outcomes)
9. [Redirects, mounts and flash](#redirects-mounts-and-flash)
10. [Transport, handshake and limits](#transport-handshake-and-limits)
11. [Managed process contract](#managed-process-contract)
12. [Manifest, version and protocol](#manifest-version-and-protocol)
13. [Writing a compatible service](#writing-a-compatible-service)
14. [Custom managed apps](#custom-managed-apps)
15. [App-owned sessions](#app-owned-sessions)
16. [Limits and future integration](#limits-and-future-integration)

## What an app is

Three roles, and only one of them knows the other two:

- **The visitor** sees an ordinary HTML page and submits an ordinary form.
- **ItsyWeb** serves the page, checks the form, forwards a small JSON object to the
  app, receives a small JSON object back, then renders the page, redirects or
  shows an error. ItsyWeb never runs app business logic.
- **The app** is an independent HTTP service. It receives JSON, decides what to
  do, and returns JSON. It does not render pages and does not send email itself;
  it can request a narrow, site-bound capability when ItsyWeb allows it.

An app is not a general application framework, a database or a job queue. ItsyWeb
keeps public pages static and sends only the data an app needs. Values returned
by an app are request-local: they are never written into the published site or
compiled as template source.

App responses are bound under the app's name. Contact mounted at `/contact`
returns values under `contact.data.*`, `contact.invalid.*`, `contact.error` and
`contact.flash.*`.

## Support at a glance

| Capability | Status today |
| --- | --- |
| Contact app: configure, render a form, deliver one email | Shipped |
| Contact mount path, recipient and optional sender | Shipped |
| ItsyWeb App Protocol v2 types, mount rules, JSON envelope | Shipped |
| Managed child process launch | Shipped |
| Register a managed local app package from site config | Shipped in 0.17.0 |
| Site-bound email capability for Contact | Shipped |
| App-owned server-side sessions (`session`, `X-Flea-Session`) | Shipped in 0.17.0 |
| `url`/port key for an external or remote service | Not supported |
| Public SDK or importable helper package | Not available |
| Read-only routes declared in configuration | Not implemented |
| Database, storage or uploads provided by ItsyWeb | Not part of the contract |

`apps.contact` is the bundled app. Any other `apps.<name>` entry is a **custom
managed app**: ItsyWeb validates `itsy/apps/<name>/app.json`, confines its
executable to the package directory, launches it the same way as Contact, and
verifies the running service at handshake time. Remote services, package
registries, automatic downloads and arbitrary shell commands are not supported.
The protocol is documented so you can build a service in any language; ItsyWeb's Go
packages are private and are not a public SDK.

## Enable Contact

Contact sends a visitor's message to one email address. It has no database,
inbox, uploads, automatic replies or delivery queue.

Merge this into `itsy.config.json`, preserving your other settings:

```json
{
  "apps": {
    "contact": {
      "mount": "/contact",
      "email": {"to": "owner@example.com"}
    }
  }
}
```

- `mount` is required and must be a canonical public path below the site root,
  for example `/contact` or `/deep/contact`. It cannot be `/`, cannot overlap an
  asset or another app mount, and cannot conflict with `/admin`, `/_flea` or
  `/releases`.
- `email.to` is required and must be one email address.
- `email.from` is optional and overrides the server's default sender. It must be
  authorized by your mail provider.
- Credentials never belong here.

Create the app's pages at the mount and a mount-relative `thank-you` page:

- `site/contact/index.html`
- `site/contact/thank-you.html`

Do not also create `site/contact.html`. `flea new --starter` includes both
pages but leaves `apps` disabled until you choose a recipient. Remove the
`apps.contact` entry to disable the app. Restart dev after changing app
settings; deployment activates production settings.

`dev.email_preview` opens local submissions as private browser previews without
sending. For real delivery, turn preview off and run `flea email`. Production
requires HTTPS and a configured provider. Provider acceptance does not guarantee
inbox delivery.

## The Contact form and templates

The form is ordinary HTML. ItsyWeb does not scan forms or inject hidden fields, so
include `{csrf}` yourself:

```html
<form method="post" action="/contact">
  {csrf}
  {#if contact.error}<p role="alert">{contact.error}</p>{/if}
  <label>Name <input name="name" required maxlength="100"
    value="{contact.data.name}" aria-describedby="name-feedback"></label>
  <p id="name-feedback">{contact.invalid.name}</p>
  <label>Email <input name="email" type="email" required maxlength="254"
    value="{contact.data.email}" aria-describedby="email-feedback"></label>
  <p id="email-feedback">{contact.invalid.email}</p>
  <label>Message <textarea name="message" required maxlength="5000"
    aria-describedby="message-feedback">{contact.data.message}</textarea></label>
  <p id="message-feedback">{contact.invalid.message}</p>
  <div hidden><label>Leave empty <input name="website" tabindex="-1"
    autocomplete="off"></label></div>
  <button type="submit">Send message</button>
</form>
```

`{csrf}` renders the complete hidden `_csrf` input. ItsyWeb checks the visitor
cookie, the token and the exact Origin before forwarding, then removes `_csrf`
from the JSON. Tokens expire after 24 hours or a full ItsyWeb restart; reload an
expired form. Template reloads keep tokens valid. Admin forms keep their own
unchanged protection.

Contact's fields:

| Field | Rule |
| --- | --- |
| `name` | Required string, trimmed, up to 100 characters, no control characters |
| `email` | Required string, trimmed, one plain email address, up to 254 characters |
| `message` | Required string, trimmed, up to 5,000 characters, no NUL |
| `website` | Optional spam trap, never redisplayed |

Fields must be valid UTF-8 and single strings. Unknown or repeated fields fail
validation. A filled spam trap with otherwise valid fields follows the success
flow without sending mail. Forms are URL-encoded and bounded to 32 KiB; there
are no uploads, no query-supplied form settings and no nested field decoding.

At a nested mount such as `/deep/contact`, use the matching pages and form
action. Success always goes to `/deep/contact/thank-you`, not `/thank-you`.
Destinations are app behavior, not configuration and not visitor input.

The site's private `mailers/contact.email.html` and `.email.txt` render the
message. They receive `email.name`, `email.address`, `email.message` and public
`app.*` values. They cannot choose recipient, sender, subject or Reply-To.

## Two boundaries: browser form and service call

Two different boundaries are easy to confuse. Keep them separate:

1. **Browser to ItsyWeb.** The visitor submits an HTML form. ItsyWeb checks CSRF,
   Origin, method, content type, size and the rate limit. ItsyWeb never sends the
   browser's cookies or `Authorization` header to the app.
2. **ItsyWeb to app.** ItsyWeb converts the checked form into flat JSON and calls the
   app's mount-relative route with its own private bearer credential.

| Browser request | Service request | Result |
| --- | --- | --- |
| `GET /contact` | none | Render the empty form |
| `POST /contact` | `POST /` | Validate and attempt delivery |
| `GET /contact/thank-you` | none | Render the thank-you page and any flash |
| `GET /contact/list` | none | `404`; Contact only mounts the form and thank-you |

Contact's two read-only pages are fixed in ItsyWeb and never call the child, and
only Contact's mount accepts a `POST`. A **custom** app may read on `GET`: a page
that exists below its mount calls the service route with the request method and
query, so an app can render state (for example, whether a session is signed in).
`HEAD` is sent to the app as `GET` and the browser receives headers only. A
`GET` on a path with no page is `404`; a general read-route registry is not
implemented.

Mount stripping is generic: the mount itself maps to `/` and descendant paths
keep their suffix. `GET /articles?page=2` calls the service with `GET /?page=2`,
preserving query parameters and repeated values. ItsyWeb matches mounts on
path-segment boundaries, so `/articles` never captures `/articles-extra`. These
are URL paths, not filesystem paths.

## Request JSON

There is **no request envelope**. The JSON body is a flat object whose keys are
the submitted form names.

```json
{"name":"Alex","email":"alex@example.com","message":"Hello","website":""}
```

Rules that ItsyWeb enforces when building this object:

- Only `POST` submissions are forwarded with a body.
- The browser form must be `application/x-www-form-urlencoded`. A query string
  on the POST is rejected with `400`.
- `_csrf` belongs to ItsyWeb. It is verified and removed, never forwarded. An app
  must not define `_csrf`.
- A name submitted once becomes a JSON string. A name submitted more than once
  becomes an array of strings. The app decides how to handle repeats.
- A submitted empty field stays an empty string (`""`). A missing field is
  absent. Unchecked checkboxes are absent.
- Field names are literal. `address.city` and `items[0]` stay exactly those
  JSON keys; ItsyWeb does not interpret dots, brackets or indexes as nesting.
- Values are always strings or arrays of strings. ItsyWeb never converts numbers,
  booleans or `null` and never nests. The app performs its own conversion.
- The JSON body is bounded to 192 KiB (six times the 32 KiB form bound).
- Non-UTF-8 names or values fail with `400`.
- Duplicate JSON keys, non-string values, arrays shorter than two items and any
  `_csrf` key are rejected by the service-side reader. Contact reports these as
  `400`; other services choose their own status.

ItsyWeb does not validate app business fields. Contact does that in its own code.
Unknown keys, repeated values and type mismatches reach Contact and fail there,
not in ItsyWeb's transport.

## Response envelope

The app replies with **one JSON object**. Only these six keys are reserved;
every other key is rejected and treated as a service failure.

| Key | JSON type | Purpose | Template path |
| --- | --- | --- | --- |
| `data` | object | App-owned values | `contact.data.name` |
| `invalid` | object of string to string | Validation messages per field | `contact.invalid.name` |
| `error` | string | One operational failure message | `contact.error` |
| `flash` | object of string to string | Temporary messages | `contact.flash.notice` |
| `redirect` | string | Mount-relative destination | Not exposed to templates |
| `session` | object or `null` | App-owned browser cookie instruction | Not exposed to templates |

Rules for the envelope:

- The body must be a single JSON object. An empty body, `null`, an array or two
  objects in a row is invalid.
- Unknown keys are invalid, including `DATA` or a `_meta` namespace.
- An explicit `null` for any envelope key is invalid, except `session` where
  `null` is the documented way to clear the cookie. An explicit `session: null`
  is only accepted on a `200` response.
- `invalid` and `flash` values must be strings. Inside `data`, any JSON is
  allowed, including nested objects, arrays, numbers, booleans and `null`.
- `flash`, serialized, must be at most 4,096 bytes.
- Duplicate JSON keys are invalid; `encoding/json` alone would silently keep one.
- Nesting is limited to 64 levels.
- The response body is bounded to 1 MiB. Larger responses are treated as a
  service failure.
- `session` is validated before any cookie changes. See
  [App-owned sessions](#app-owned-sessions).

Only envelope keys are reserved. Names inside `data` belong to the app. ItsyWeb
renders valid JSON `data` as returned, but the included Contact bridge only ever
redisplays the bounded `name`, `email` and `message` values on failure.

## Status codes and browser outcomes

ItsyWeb accepts three application statuses and maps everything else to a safe
unavailable result. The envelope must match the status:

| Service status | Envelope requirement | Meaning |
| --- | --- | --- |
| `200` | no `invalid`, no `error` | Success or an app-rendered read |
| `422` | at least one `invalid`, no `error` | Validation failure |
| `503` | an `error`, no `invalid` | Operational failure |

Any other status, a malformed body, an unknown envelope key, a wrong content
type or a transport error is replaced by a generic `503` with the message
"This app is temporarily unavailable. Please try again later." No upstream
diagnostic is shown to the visitor.

Browser outcomes:

- `GET`/`HEAD` on a read-only route renders the page with status `200` and never
  calls the app.
- A successful `POST` that returns `redirect` becomes a browser `303` to the
  destination, with flash stored for the next render.
- A `422` re-renders the same page with status `422`. Contact copies bounded
  `name`, `email` and `message` values into `contact.data.*` and the messages
  into `contact.invalid.*`. It never redisplays the spam trap, provider errors
  or credentials.
- A `503` re-renders the page with status `503` and `contact.error`.
- Form failures before any app call use `400`, `403`, `405` or `415`.
- Contact rate limiting returns `429` with `Retry-After: 900`.

### Fully valid examples

Browser form body (`application/x-www-form-urlencoded`, `_csrf` present in the
real request):

```text
_csrf=<token>&name=Alex&email=alex%40example.com&message=Hello&website=
```

Service request:

```json
{"name":"Alex","email":"alex@example.com","message":"Hello","website":""}
```

Success response, service `200`:

```json
{"redirect":"thank-you","flash":{"notice":"Your message was sent."}}
```

Browser outcome: `303` to `/contact/thank-you`. The thank-you page can render
`{contact.flash.notice}`.

Validation response, service `422`:

```json
{"data":{"name":"Alex","email":"invalid","message":"Hello"},"invalid":{"email":"Enter a valid email address."}}
```

Browser outcome: the contact page re-renders with status `422`, showing the
message and the safe values.

Operational failure response, service `503`:

```json
{"data":{"name":"Alex","email":"alex@example.com","message":"Hello"},"error":"We couldn't confirm delivery. Please try again later or contact us directly."}
```

Browser outcome: the contact page re-renders with status `503` and the error.

Read response, service `200` (generic capability; no public read route is
registered for Contact):

```json
{"data":{"title":"Articles","items":[{"slug":"one","title":"First"}]}}
```

## Redirects, mounts and flash

- `redirect` is a **mount-relative** path. `"thank-you"` at mount `/contact`
  becomes `/contact/thank-you`. A leading `/`, a full URL, `..`, an encoded
  traversal, a query string or a fragment is rejected. The destination page must
  exist in the site, and `redirect` is only honored on a `POST` that returned
  `200`.
- Success always goes to the mount's own `thank-you` page. A nested mount never
  redirects to a top-level `/thank-you`.
- Flash is private and bound to the visitor, host, app and destination. It is
  consumed only by a successful `GET` render. Failed renders, `HEAD` and
  unrelated pages do not consume it. It expires after five minutes or a restart.
  Storage is capped at 256 messages per site; when full, success still redirects
  without flash. Concurrent tabs share one app flash cookie, so the last
  response cookie wins.
- App pages are private and `no-store`. App values are escaped as values and
  never compiled as template source.
- Contact values appear under `contact.*` in templates. Conditions and dotted
  paths work normally, for example `{#if contact.error}` and
  `{contact.invalid.email}`.

## Transport, handshake and limits

ItsyWeb calls the app over HTTP on loopback with a private bearer credential.
These are protocol facts, not site settings.

- **Handshake.** Before any app request, ItsyWeb calls `GET /_flea/handshake` with
  `Authorization: Bearer <app-token>`. The app must return `200` and JSON with
  `name` matching the expected app, `protocol` equal to `2`, and a non-empty
  `version`. A missing or wrong credential must be denied before the app
  returns anything. An open port or a readable manifest is not enough.
- **Authorization.** Every Flea-to-app call carries the same bearer token.
  Browser cookies and `Authorization` headers are never forwarded. Handshake
  is service-internal and is never exposed through the public mount.
- **Content type.** ItsyWeb sends `Content-Type: application/json` on POST and
  `Accept: application/json`. Responses must be `application/json`; a response
  with any `Content-Encoding` is rejected. Upstream redirects are not followed.
- **Bounds.** Handshake response 4 KiB; ordinary response 1 MiB; request body
  192 KiB; POST form 32 KiB. Handshake deadline 2 seconds; ordinary request
  deadline 15 seconds.
- **Concurrency.** At most four calls in flight per app client and 64 across the
  process. There is no waiting queue: excess calls fail immediately with the
  generic unavailable result.
- **Retries.** There are no automatic POST retries. The request body is sent
  with a non-rewindable reader specifically so the transport cannot retry it.
  An uncertain timeout can mean a later manual retry sends a duplicate.
- **Errors.** A malformed, oversized, wrong-typed or unavailable response is
  replaced by the generic `503`. App validation uses `422`.

## Managed process contract

Apps are launched and supervised by ItsyWeb only during an actual `flea dev` or
host run. Loading, `flea status`, validation and extraction never start an app.
This is an **internal managed contract between ItsyWeb and a confined executable,
not site configuration**. A site configures only `apps.<name>.mount` (and
`apps.contact.email`); it never sets ports, descriptors or tokens.

For a managed child (bundled Contact or a custom package):

- **fd 3** is an already-bound `127.0.0.1:0` TCP listener. The OS picks the port,
  so there is no probe/close/rebind race and no port in site config.
- **fd 4** is a one-shot JSON credential pipe with exactly five string keys:
  `{"Name":"<app>","Token":"<64 hex>","Storage":"<private dir>","CapabilityURL":"","CapabilityToken":""}`.
  Contact receives a non-empty `CapabilityURL` and `CapabilityToken`; a custom
  app receives empty strings for both. Unknown fields are rejected, so declare
  all five. The app token authenticates ItsyWeb's calls to the app.
- **fd 5** is a parent-liveness pipe. When the parent closes it, the child
  cancels and shuts down.
- The child environment is only `LANG=C`. Standard output and error are
  discarded so private payloads are not logged. Credentials never appear in
  argv, environment, site files or logs.
- Contact uses one checksum-verified private executable shared by its enabled
  sites; a custom app runs the executable confined to `itsy/apps/<name>/`. ItsyWeb
  starts one child per site and app, allows five attempts with 1, 2, 4 and 8
  second backoff, then leaves that app unavailable until restart while other
  pages stay up.
- Loopback HTTP is not a sandbox against other code running as the same user.
  Processes are not OS sandboxes and are not described as such.

The `Storage` directory is app-owned private runtime state under the stable
per-site runtime directory, a sibling of admin session state and outside every
immutable release (`<site>/itsy/runtime/apps/<name>/` in development,
`<data>/sites/<name>/itsy/runtime/apps/<name>/` in production), created `0700`.
It therefore survives a release change. ItsyWeb never reads, prunes or backs it up;
it is excluded from releases and backups, so back app data up yourself. A
production site with a custom app requires a managed content snapshot. See
[App-owned sessions](#app-owned-sessions).

### Capability request and response

The child does not hold mail credentials. It asks ItsyWeb for the one thing it is
allowed to do. Contact posts to:

```text
POST <CapabilityURL>/_flea/capabilities/contact-email
Authorization: Bearer <capability-token>
Content-Type: application/json
```

Body (exactly these three keys):

```json
{"name":"Alex","email":"alex@example.com","message":"Hello"}
```

ItsyWeb requires an identical, currently admitted submission and permits only one
attempt. It chooses recipient, sender, subject and templates itself; the visitor
address is Reply-To only. Responses: `204` on success, `422` for an invalid
body, `403` for an unauthorized or already-used notification, `503` when
delivery capacity is unavailable. The endpoint is never mounted publicly.

## Manifest, version and protocol

Three separate ideas:

- **Protocol** is the wire contract, an integer. The current ItsyWeb App Protocol is
  `2` (v1 is rejected; there is no fallback). ItsyWeb checks the running service's
  protocol at handshake time.
- **App version** is the app's own release string, for example `0.1.0`. It need
  not match ItsyWeb's version. It identifies the app build, not the contract.
- **Manifest** is a package description, not site routing. A custom package's
  `itsy/apps/<name>/app.json` declares one executable per platform, keyed
  `<goos>-<goarch>`:

```json
{
  "name": "session",
  "version": "0.1.0",
  "protocol": 2,
  "executable": {
    "darwin-arm64": "session-darwin-arm64",
    "linux-amd64": "session-linux-amd64"
  }
}
```

The bundled Contact manifest (`apps/contact/app.json`) declares no executable
because ItsyWeb extracts it from its own platform bundle. The manifest is not
trusted as proof of compatibility: the handshake is. Site configuration never
contains a version or protocol.

## Writing a compatible service

You can write an independent HTTP service that speaks the current ItsyWeb App
Protocol. It can be in any language; it does not need ItsyWeb's Go packages, which
are private and cannot be imported by an external module. There is no public SDK.
Once built, you register it as a managed local package; see
[Custom managed apps](#custom-managed-apps).

A compatible service must:

1. Require a bearer token on every request and deny others before doing work.
2. Answer `GET /_flea/handshake` with `{"name":<name>,"version":<string>,"protocol":2}`.
3. Accept `POST <mount-stripped-path>` with `Content-Type: application/json`
   and a flat object of string or string-array values.
4. Reply with `Content-Type: application/json` and the response envelope.
5. Use only the accepted statuses and matching envelopes.

The short example below is a complete, standard-library way to exercise the wire
contract manually. It is not a managed package. The fuller runnable sample,
including the inherited-fd launch contract, app-owned sessions and persistence,
is [`samples/session-app`](../samples/session-app) in this repository; it imports
no ItsyWeb package.

```go
package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

const protocol = 2

func main() {
	token := os.Getenv("APP_TOKEN") // placeholder; use a long random secret
	if len(token) < 32 {
		log.Fatal("set APP_TOKEN to at least 32 characters")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/_flea/handshake", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r, token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"name": "example", "version": "0.1.0", "protocol": protocol,
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r, token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var fields map[string]any
		if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		name, _ := fields["name"].(string)
		if name == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"data":    map[string]any{"name": name},
				"invalid": map[string]string{"name": "Enter your name."},
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"redirect": "thank-you",
			"flash":    map[string]string{"notice": "Thanks, " + name + "."},
		})
	})
	log.Fatal(http.ListenAndServe("127.0.0.1:8787", mux))
}

func authorized(r *http.Request, token string) bool {
	got := []byte(r.Header.Get("Authorization"))
	want := []byte("Bearer " + token)
	return subtle.ConstantTimeCompare(got, want) == 1
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
```

Exercise the contract locally:

```sh
export APP_TOKEN="$(openssl rand -hex 32)"
go run .

# In another terminal, reuse the same APP_TOKEN.
curl -s http://127.0.0.1:8787/_flea/handshake \
  -H "Authorization: Bearer $APP_TOKEN"
# {"name":"example","protocol":2,"version":"0.1.0"}

curl -s -X POST http://127.0.0.1:8787/ \
  -H "Authorization: Bearer $APP_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alex"}'
# {"flash":{"notice":"Thanks, Alex."},"redirect":"thank-you"}

curl -s -X POST http://127.0.0.1:8787/ \
  -H "Authorization: Bearer $APP_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
# {"data":{"name":""},"invalid":{"name":"Enter your name."}}
```

That example proves the wire contract, authentication, statuses and envelope
shapes. It does not mount the service in a site.

### Registering a managed service

Build your service for every platform you run (your development host and the
server), place the executables in a package directory, and list them in the
manifest. Then enable it in site config. The manifest is validated when the site
loads; the running service is verified by the handshake at startup.

```text
itsy/apps/session/app.json             {"name":"session","version":"0.1.0","protocol":2,"executable":{"darwin-arm64":"session-darwin-arm64","linux-amd64":"session-linux-amd64"}}
itsy/apps/session/session-darwin-arm64 the Mac executable
itsy/apps/session/session-linux-amd64  the server executable
```

```json
{
  "apps": {
    "session": {"mount": "/session"}
  }
}
```

Create a page at the mount, for example `site/session/index.html`, using
`{session.data.*}` and, for forms, `{csrf}`. Restart `flea dev` after changing app
settings or rebuilding the executable. No `url`, port, registry or install
command is involved: ItsyWeb launches only the confined executable you placed
there. See [Custom managed apps](#custom-managed-apps) for validation and
confinement rules.

## Custom managed apps

Any `apps.<name>` entry other than `contact` registers a custom managed app.
ItsyWeb loads the package at `itsy/apps/<name>/` at project validation time:

- `app.json` must be a single JSON object with exactly `name`, `version`,
  `protocol` and `executable`, with no duplicate, trailing, unknown or
  case-variant fields. It is bounded to 64 KiB.
- `name` must equal the config key and be a lowercase `[a-z][a-z0-9_]*`
  identifier that is not a reserved template namespace.
- `protocol` must be `2`. A different protocol is rejected; there is no fallback.
- `version` is a non-empty printable string of at most 64 bytes.
- `executable` maps `<goos>-<goarch>` to one filename in the same directory. The
  running platform's entry is required; the packer ships exactly the declared
  files. Each declared filename must be a single confined basename and cannot be
  `app.json`, `data` or a dotfile.
- `app.json` and every declared executable must be a single-link regular file.
  The running platform's executable must also carry an execute bit and begin a
  native binary for that platform (ELF on Linux, Mach-O on macOS). Path
  separators, `..` and nested paths are rejected, so the executable cannot escape
  the package. At launch ItsyWeb re-opens it confined, verifies its size and SHA-256
  against the value recorded at load, and copies it to a private file before
  running, so a swap after validation cannot change what runs.
- The site must have a page at the configured `mount`. Pages below the mount are
  reachable; the mount itself maps to service route `/`.
- Only `app.json` and the declared executables are packaged. The runtime `data`
  name, other unlisted files, dotfiles, nested paths and symlinks are never
  packaged, and the extractor rejects an archive containing an unlisted package
  file. A custom app takes no `email` settings.

The mount uses the same rules as Contact: a canonical public path below the site
root that does not overlap an asset, a core route or another mount. At runtime
ItsyWeb sends the app its name, token and a private storage directory through the
inherited credential pipe and supervises one child per site with bounded
backoff. An unavailable custom app only affects its own routes.

## App-owned sessions

The app owns session storage, authentication, expiry, revocation and token
generation. ItsyWeb adds no session database and keeps no app session data. It
bridges the browser and the app:

- On each request, ItsyWeb reads its own per-site, per-app cookie and forwards the
  opaque token to the app as the `X-Flea-Session` request header. It never
  forwards a browser `Cookie` or `Authorization` header, and the token is never
  put into templates, logs, URLs or public JSON.
- The app's response may include `session`:

| `session` value | Effect |
| --- | --- |
| omitted | Leave the cookie unchanged |
| `null` | Clear the cookie |
| `{"token":"<opaque>","max_age":<seconds>}` | Set or rotate the cookie |

Rules ItsyWeb enforces before changing any cookie:

- `session` is accepted only on a `200` response. On `422` or `503` it is a
  service failure and no cookie changes.
- `token` is a JSON string of 16 to 512 bytes from cookie-safe printable ASCII
  (no spaces, `;`, `,`, `"` or `\`). Anything else is rejected.
- `max_age` is a bare JSON integer from `1` to `2592000` (30 days). Strings,
  fractions and out-of-range values are rejected.
- The object has exactly `token` and `max_age`, spelled exactly that way. Unknown
  or case-variant keys, duplicate keys, a non-object value or malformed JSON are
  rejected. An explicit `null` is the only valid non-object value.

ItsyWeb owns the cookie name and attributes. In production the name is
`__Host-flea-session-<app>` with `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`
and no `Domain`; in trusted HTTP development a site-specific name is used so two
sites on one loopback host never share a cookie. Cookies are isolated by app, and
by site through host-only scoping: the browser never sends a production session
cookie to a different site's host. A malformed or duplicate cookie, including a
valid cookie accompanied by a malformed duplicate of the same name, is ignored
and no token is forwarded. A cookie for another app is never read for this app.

The session cookie is written before the response status or redirect, so a
successful redirect carries it. Concurrent responses for the same app each carry
one session cookie; ordinary browser cookie rules decide which one wins, and the
last committed response takes effect. An app session never grants ItsyWeb admin or
editor access: those use separate cookies and remain unchanged.

Store sessions in the private `Storage` directory ItsyWeb passes at launch. It is
stable per-site runtime state outside every release, never packaged and never
backed up, so keep your own copy; a normal app restart or release change leaves
it untouched.

## Limits and future integration

- Contact is bundled. Any other app must be a managed local executable under
  `itsy/apps/<name>/`; there is no install command, package registry or
  automatic download.
- Remote services, external `url`/port registration, read-route declarations in
  configuration, databases, storage platforms and file uploads are not part of
  the contract.
- ItsyWeb supervises only a confined local executable. It does not run arbitrary
  shell commands and does not sandbox processes; loopback HTTP is not isolation.
- App sessions and app storage are app-owned. ItsyWeb stores no session data and
  does not back up app storage; keep your own copy. A normal release change
  leaves the runtime directory in place. An app session never grants admin or
  editor access.
- Email delivery, limits, flash storage and process supervision are server-owned.
  An app cannot select recipients, bypass limits or send arbitrary mail.
- Provider acceptance does not guarantee inbox delivery. Local previews and
  tests do not establish real delivery acceptance.

The design sketch for remote services and richer read routes stays in
[planning/apps.md](agents/planning/apps.md); treat it as design, not shipped
behavior. For the included app's full operational contract and acceptance
limits, see [Contact](../apps/contact/README.md).
