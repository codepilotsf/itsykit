# Email templates

Developer-owned `mailers/` contains `main.layout.html`, `sign-in.email.html`,
`sign-in.email.txt`, `contact.email.html`, and `contact.email.txt`.
They stay private and deploy with the site.
The HTML layout wraps each HTML body through `{slot}`. Plain text has no layout.

Use `{app.site_name}` or other public app values, plus:

- Sign-in: `{email.code}`, `{email.host}`.
- Contact: `{email.name}`, `{email.address}`, `{email.message}`.

Values are escaped for HTML and literal in plain text. This intentionally small
email syntax supports dotted values, not loops, snips, includes, or scripts.
Keep inline CSS suitable for email clients; the site's normal CSS is not loaded.
Templates control bodies, never recipients, sender, subject or Reply-To.
Missing or invalid templates stop delivery. Existing sites need the `mailers`
directory. Move existing customized email templates there without changing their
contents; there is no old-path or built-in site-template fallback.

## Local delivery previews

New sites include `"dev": {"email_preview": true}` in `.flea.config.json`.
While `flea dev` runs, sign-in and contact emails open as browser previews instead
of being sent. First-run dev asks for the site name and optional admin email,
never Resend credentials. Approved admins still use the normal code sign-in flow;
copy the real code from the preview. Contact forms require `apps.contact.email.to` and a configured app mount.

The email window shows subject, sender name/address, recipient, optional Reply-To,
local date/time, the rendered HTML and an expandable plain-text version. Sender
details use local `FLEA_EMAIL_NAME` and `FLEA_EMAIL_FROM` when set; otherwise the
site name and `preview@localhost` identify the simulated sender. Contact's optional
`apps.contact.email.from` overrides the sender address in previews and real mail.

Only the dev CLI selects this sender. Production ignores the toggle and requires
real delivery. Preview mode overrides even configured Resend credentials locally.
Set the toggle to false (or omit it), configure `flea email`, then restart dev
to send real mail. Restart after changing preview mode or preview sender details.

Previews are held in memory, with at most 20 messages and 10-minute expiry.
Their random URLs work only on the exact loopback dev address by default.
With `flea dev --lan` (or `-l`), they also work on the printed network addresses.
Previews still open on the computer running Flea. Enter its code on the other
device, or replace the preview URL's loopback IP with a printed network IP while
keeping its port and random path. A preview link is private access to that message;
share it only with someone you trust. No preview listing or login bypass is added.
Stopping dev
discards them; no preview files enter Git, backups or deployment. The HTML body
is sandboxed with scripts, forms and external resource loads disabled. No inbox
listing is exposed. If browser launch fails, a local preview link is printed in
the terminal; anyone with that link and local access can view it until expiry.
The previews do not test provider delivery or real email-client rendering.

## Sample rendering

To inspect a fake-data sample without triggering an actual sign-in or contact form:

```sh
flea email preview sign-in > /tmp/flea-email-preview.html
open /tmp/flea-email-preview.html
```

Use `contact` to preview a contact message. Both use fake sample data. Dev reloads
template changes. Templates are bounded to 128 KiB after layout composition, and
rendering stops at 256 KiB per body. Live mailbox rendering/delivery still needs
testing with the actual provider and email clients.

The starter includes `/contact` and `/contact/thanks`. Enable `apps.contact` with
`mount: "/contact"` and `email.to` for the intended recipient. Optional `email.from`
overrides the server default and must be authorized by Resend. Preview mode needs
no credentials; run `flea email` for real delivery. Restart dev after app/email
changes. Success redirects to `thanks` relative to the mount; no config expression
or visitor field selects it. See the [Contact contract](../apps/contact/README.md).
