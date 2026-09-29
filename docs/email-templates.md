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
of being sent. First-run dev asks for the site name and optional owner email,
never Resend credentials. Approved owners still use the normal code sign-in flow;
copy the real code from the preview. Contact forms still require `contact.to`.

The email window shows subject, sender name/address, recipient, optional Reply-To,
local date/time, the rendered HTML and an expandable plain-text version. Sender
details use local `FLEA_EMAIL_NAME` and `FLEA_EMAIL_FROM` when set; otherwise the
site name and `preview@localhost` identify the simulated sender.

Only the dev CLI selects this sender. Production ignores the toggle and requires
real delivery. Preview mode overrides even configured Resend credentials locally.
Set the toggle to false (or omit it), configure `flea email`, then restart dev
to send real mail. Restart after changing preview mode or preview sender details.

Previews are held in memory, with at most 20 messages and 10-minute expiry.
Their random URLs work only on the exact loopback dev address. Stopping dev
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

The starter includes `/contact` and `/thank-you`. Set `contact.to` to the intended
recipient. Preview mode needs no credentials; run `flea email` for real delivery.
Restart dev after contact
configuration changes. Optional `contact.success_path` selects a fixed local
303 redirect after successful delivery; it cannot be supplied by form visitors.
