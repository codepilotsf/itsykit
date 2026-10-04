# Contact app

Status: proposed README, not implemented. October 4, 2026.

This is an authoring specification for the first app using Flea App Protocol v1.
Examples describe proposed behavior, not commands or configuration supported by
today's Flea. See the [app plan](apps.md) for shared decisions. Items explicitly
marked proposed or open still need agreement before implementation.

## What it does

Send a website visitor's message to one configured recipient. No database,
stored submissions, background queue, attachments, or client account system.
Admin login and the editor remain in Flea core.

The browser submits an ordinary HTML form to Flea. Flea checks the request,
forwards JSON to the contact service, and renders HTML or redirects based on
the service's JSON response. The browser never calls the service directly.

## Package and registration

Proposed installed package contents, relative to the still-undecided app root:

```text
contact/
  app.json
  contact       # platform-specific executable
  README.md
```

Proposed manifest (`command` syntax remains open):

```json
{
  "name": "contact",
  "version": "0.1.0",
  "protocol": 1,
  "command": "./contact"
}
```

Merge this registration into the main `.flea.config.json`; do not replace other
site settings:

```json
{
  "apps": {
    "contact": {"mount": "/contact"}
  }
}
```

The mount enables the package. No template app directive or operation registry
is needed. Flea starts the local service and checks its protocol. Discovery
alone never starts it. Launch arguments and protocol handshake remain open.

## Pages and routes

Create site-owned templates at `site/contact/index.html` and
`site/contact/thanks.html`. These use the site's normal layout and styling.
Do not also create `site/contact.html`, which would duplicate the form route.

| Browser request | Service request | Result |
| --- | --- | --- |
| `GET /contact` | None needed | Render empty form |
| `POST /contact` | `POST /` with JSON | Validate and attempt delivery |
| `GET /contact/thanks` | None needed | Render thank-you page and available flash |

This proposes explicit render-only routes for contact, not forwarding every
GET under every mount. How Flea learns which routes require service reads is
still open. Blog list/detail routes will require service GET requests.

At mount `/deeply/nested/route/contact`, put the templates at the corresponding
site paths and post to that full public mount. The fixed success destination is
`thanks`, resolved by Flea to `/deeply/nested/route/contact/thanks`. No visitor
field or site-config redirect expression selects the destination.

## Submission fields

Proposed field rules preserve the current contact handler's basic scope:

| Field | Rule |
| --- | --- |
| `name` | Required string, trimmed, at most 100 characters; no control characters |
| `email` | Required string containing one valid email address; Reply-To only |
| `message` | Required string, trimmed, at most 5,000 characters; no NUL characters |
| `website` | Optional spam-trap string; not redisplayed |

Use valid UTF-8. All these fields are scalar; reject repeated values rather
than choosing one. Unknown-field handling and exact spam-trap outcomes need
final specification. Preserve bounded requests and existing spam protections.

Flea forwards ordinary fields without a `data` wrapper:

```json
{
  "name": "Alex",
  "email": "alex@example.com",
  "message": "Do you have availability next week?",
  "website": ""
}
```

Flea validates and removes its own CSRF field before forwarding. The precise
token-rendering mechanism is open; no new template syntax is implied here.

## Responses

All service responses below have JSON bodies. Proposed success status is 200
from the service; Flea turns its `redirect` instruction into a browser 303:

```json
{
  "redirect": "thanks",
  "flash": {"notice": "Your message was sent."}
}
```

Validation failure returns HTTP 422 without attempting delivery:

```json
{
  "data": {"name": "Alex", "email": "invalid", "message": "Hello"},
  "invalid": {"email": "Enter a valid email address."}
}
```

Flea re-renders the form page at the same URL with status 422. Name, email, and
message are redisplayable; passwords, files, and the spam-trap field are not.

Delivery failure returns HTTP 503:

```json
{
  "error": "We couldn't send your message. Please try again."
}
```

Flea preserves redisplayable fields and renders the form with status 503. An
unreachable service also produces a safe 503. Never expose provider responses,
credentials, or internal diagnostics in the page. Automatic retry of a delivery
request is not implied: a timeout may occur after a provider accepted the mail.

## Template example

This is the form body, not a complete deployable page. It deliberately omits
the unresolved CSRF control and spam-trap markup; both must be specified before
shipping. App values are contextually escaped, never compiled as template code.

```html
<form method="post" action="/contact">
  {#if contact.error}
    <p role="alert">{contact.error}</p>
  {/if}

  <label for="contact-name">Name</label>
  <input id="contact-name" name="name" required maxlength="100"
         value="{contact.data.name}" aria-describedby="name-feedback">
  <p id="name-feedback">{contact.invalid.name}</p>

  <label for="contact-email">Email</label>
  <input id="contact-email" name="email" type="email" required
         value="{contact.data.email}" aria-describedby="email-feedback">
  <p id="email-feedback">{contact.invalid.email}</p>

  <label for="contact-message">Message</label>
  <textarea id="contact-message" name="message" required maxlength="5000"
            aria-describedby="message-feedback">{contact.data.message}</textarea>
  <p id="message-feedback">{contact.invalid.message}</p>

  <button type="submit">Send message</button>
</form>
```

Proposed render defaults provide empty strings for absent form/feedback values.
The thank-you page can always display static acknowledgment, with an optional
temporary message:

```html
<h1>Thank you</h1>
{#if contact.flash.notice}
  <p role="status">{contact.flash.notice}</p>
{/if}
```

Flash belongs to this visitor and app, survives the redirect, and expires after
the intended page render. Assets must not consume it. Cookie/session storage,
multi-tab behavior, and exact consumption semantics remain open. Do not cache
visitor-specific responses in shared page caches.

## Email settings and security

Resend remains the supported real-delivery provider. Keep one site-configured
recipient, a configured sender, a fixed subject, and the visitor's email as
Reply-To only. Visitors cannot supply recipients, sender settings, or credentials.

Preserve the requirements in the [runtime guide](../agents/runtime.md) and
[email template guide](../email-templates.md): private site-owned mailers,
bounded rendering, exact Origin/host checks, applicable CSRF protection,
production HTTPS, site/IP limits, and the shared server contact budget. Keep
message contents and secrets out of logs, Git, public files, and backups.

Flea owns email credentials, private template rendering, delivery, and the
existing development-only preview sender. The contact app validates submissions
and invokes a narrow private capability to send this site's contact notification.
It does not receive Resend credentials or choose arbitrary recipients, senders,
templates, or subjects. Core admin sign-in continues using Flea's mail machinery.

The capability is authorized for the calling app and site, not a site identifier
supplied in submission data. Flea must enforce its bounds even if the app sends
an invalid request. Only confirmed delivery acceptance (or successful local
preview creation) permits the app's success response; private errors become safe
operational feedback. Provider acceptance does not guarantee inbox delivery.

The following wiring remains undecided:

- Where app-specific recipient settings live in main config.
- The private capability's endpoint, request/response shape, and authentication.
- How capability credentials are scoped, passed, and kept out of logs and archives.
- Timeouts and limits for both Flea-to-app calls and app-to-Flea capability calls.
- Whether and how remote app services can access this capability securely.

Do not bypass these questions by introducing a new provider, public mail relay,
login bypass, or a second incompatible email-template system.

## Acceptance before implementation is complete

- Plain HTML success, validation, and delivery-failure flows work without JS.
- Nested mounts use the correct service path and thank-you destination.
- Invalid fields survive redisplay safely; scripts and template text stay inert.
- Cross-site recipients, flash, credentials, and limits remain isolated.
- Unavailable or incompatible services do not break unrelated pages.
- Current spam controls and development preview requirements are preserved.
- Protocol, route-read selection, CSRF markup, and email wiring are specified.
- Real Resend delivery is verified separately from mocks and local previews.

ZJAX enhancement is optional and must be verified separately with 422/503 HTML
and success redirects. No ZJAX changes are authorized by this draft.
