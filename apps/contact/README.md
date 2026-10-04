# Contact

Send a visitor's message to one email address using an ordinary HTML form.
Contact uses Flea App Protocol v1. It has no database, inbox, uploads, automatic
replies or delivery queue. Admin sign-in and editing remain in Flea core.

## Enable it

Merge this into `.flea.config.json`, preserving your other settings:

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

`email.to` is required. Optional `email.from` overrides the server's default
sender, for example `support@example.com`; the address must be authorized by
Resend. The visitor's address is Reply-To only. Credentials never belong here.
Remove the `apps.contact` entry to disable the app. No `{app ...}` tag is needed.

Create `site/contact/index.html` and `site/contact/thanks.html`, using your
normal layout. `flea new example --starter` includes both, but leaves apps
disabled until you choose a recipient. Do not also create `site/contact.html`.
Restart dev after changing app settings; deployment activates production settings.
There is no old `contact` config or `/_flea/contact` route alias.

With `dev.email_preview` on, local submissions open as private browser previews
without being sent. For real delivery, turn preview off and run `flea email`.
Production requires HTTPS and Resend. Provider acceptance does not guarantee
inbox delivery. Local tests/previews do not establish real delivery acceptance.

## Form

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

`{csrf}` writes the whole hidden `_csrf` input. Flea does not scan forms or inject
it automatically. It checks the visitor cookie, token and exact Origin before
forwarding, then removes `_csrf` from the JSON. Tokens expire after 24 hours or
a full Flea restart; reload an expired form. Template reloads keep tokens valid.
Admin forms have their own unchanged protection.

| Field | Rule |
| --- | --- |
| `name` | Required string, trimmed, up to 100 characters, no control characters |
| `email` | Required string, trimmed, one plain email address, up to 254 characters |
| `message` | Required string, trimmed, up to 5,000 characters, no NUL |
| `website` | Optional spam trap, never redisplayed |

Fields must be valid UTF-8 and scalar strings. Unknown or repeated fields fail
validation. A filled spam trap with otherwise valid fields follows the success
flow without sending mail. URL-encoded forms only, bounded to 32 KiB; no uploads,
query-supplied form settings or nested field decoding.

## Routes and JSON

| Browser request | Service request | Result |
| --- | --- | --- |
| `GET /contact` | None | Render empty form |
| `POST /contact` | `POST /` | Validate and attempt delivery |
| `GET /contact/thanks` | None | Render thanks and available flash |

HEAD renders without a body and never consumes flash. Other routes below this
mount return 404. Read-only Contact pages do not depend on a healthy child.
Contact's two read-only routes are fixed in Flea; a general read-route registry
for future apps is not implemented.

At mount `/deep/contact`, use the corresponding templates and form action.
Success always goes to `/deep/contact/thanks`, not `/thanks`. Destinations are
app behavior, not expressions in config and not selected by visitor input.

The service receives flat JSON, without a `data` wrapper:

```json
{"name":"Alex","email":"alex@example.com","message":"Hello","website":""}
```

On success the service returns HTTP 200:

```json
{"redirect":"thanks","flash":{"notice":"Your message was sent."}}
```

Flea turns this into a browser 303. The thank-you page may display
`{contact.flash.notice}` alongside a static acknowledgment. Flash is private,
bound to the visitor, host, app and destination, and consumed only by a successful
GET render. Failed renders, HEAD and unrelated pages do not consume it. It
expires after five minutes or restart. Storage is capped at 256 messages/site;
when full, success still redirects without flash. Concurrent tabs share one
app flash cookie: the last response cookie wins. These pages are private/no-store.

Validation returns HTTP 422, with named messages and safe values:

```json
{"data":{"name":"Alex","email":"invalid","message":"Hello"},"invalid":{"email":"Enter a valid email address."}}
```

Flea re-renders the same page with status 422. Operational failures return 503
and an `error` string. Flea preserves bounded name/email/message fields; it never
redisplays spam traps, provider errors or credentials. App data is escaped as
values, never compiled as template source. There are no automatic POST retries:
an uncertain timeout can mean a later manual retry sends a duplicate.

## Email and process boundaries

Flea renders the site's private `mailers/contact.email.html` and `.email.txt`.
They receive `email.name`, `email.address`, `email.message` and public `app.*`
values. They cannot select recipients, sender, subject or Reply-To.

The build bundles a separate Contact executable for its own platform. Flea
extracts one checksum-verified private copy shared by its enabled sites, starts
one child per site and passes an already-bound `127.0.0.1:0` listener on fd 3.
The OS chooses an available port without a close/rebind race. There is no VPS
compiler, per-site executable copy or site-config port to manage.

Fd 4 carries independent app and capability credentials; fd 5 signals parent
liveness. Credentials never enter argv, environment, site files or logs. The
child receives no provider environment. Private `GET /_flea/handshake` must
return matching name/protocol before the app becomes ready. Handshake and
service calls use `Authorization: Bearer <app-token>`; browser credentials are
never forwarded. Loopback HTTP is not a sandbox against other code with the
same account privileges. No remote app registration is supported yet.

The child requests mail with authenticated JSON POST to private
`/_flea/capabilities/contact-email`, using a separate bearer token. Its body has
exactly `name`, `email`, `message`. Flea requires an identical, currently admitted
submission and permits only one attempt. It chooses all email settings itself.
Success is 204; invalid/unauthorized/unavailable requests are 422/403/503.
This endpoint is never mounted publicly.

Limits are three valid submissions/IP/site/15 minutes, 20/site/hour, one active
submission/site, and a shared four active sends/200 sends per server hour.
At most 1,024 IP buckets/site are retained. Site/IP admission failures return
429; unavailable global delivery capacity returns a safe 503. Aliases share
limits; forwarded headers never select the IP. Limits reset on Flea restart.
Login mail has separate limits but shares provider quota. Delivery has a
10-second deadline, callback calls 12 seconds and Flea-to-app calls 15 seconds.
Bodies, responses and concurrent protocol calls are bounded; there is no queue.

Flea allows five process attempts with 1/2/4/8-second restart delays, then leaves
Contact unavailable until restart. Other pages remain usable. Shutdown closes
the parent-liveness pipe, waits briefly and kills a stuck child. No migration
runs on startup or recovery. Normal shutdown removes the extracted executable;
abrupt termination can leave its private temporary directory. Processes are not
OS sandboxes. The [initial Linux capacity check](../../docs/benchmarks/2026-10-04-contact-apps.md)
covered 30 lightweight Contact sites; heavier workloads and longer soaks still
need measurement before making production-capacity guarantees.
