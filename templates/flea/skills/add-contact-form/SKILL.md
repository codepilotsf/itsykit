---
name: add-contact-form
description: Add, enable or repair FleaCMS's bundled Contact form on an existing Flea site. Use for requests such as "add a contact form", "create a contact page", "let visitors email us", or connecting a form to Resend. Covers configuration, exact field names, CSRF, feedback tokens and mail templates. Does not apply to arbitrary forms, uploads, newsletters or a contact page that only displays an address.
---

# Add a Flea Contact form

Read the site's [README](../../../README.md) and [AGENTS](../../../AGENTS.md).
Inspect Git status, existing Contact configuration, navigation and templates.
Preserve custom layouts, content and other app settings. Follow the site's Prewind
styles; adding a form does not require a full-site redesign. This skill does not
authorize deployment or sending real test mail.

## Configure and connect

Reuse an existing `apps.contact` entry. Otherwise merge this fragment into
`.flea.config.json`, preserving other fields:

```json
{"apps":{"contact":{"mount":"/contact","email":{"to":"owner@example.com"}}}}
```

The address is a placeholder: obtain the real recipient from the user or existing
approved configuration. Never infer it from the admin login or publish it in
page markup. Optional `email.from` must be a Resend-authorized sender; the visitor
email becomes Reply-To only. Provider credentials belong in private mail setup,
never configuration JSON, templates or Git.

At `/contact`, create `site/contact/index.html` and
`site/contact/thank-you.html` using the existing layout. Do not also create
`site/contact.html`. At `/inquiries`, use `site/inquiries/`, form action
`/inquiries`, and success URL `/inquiries/thank-you`. These read routes are fixed;
arbitrary child pages and custom redirects are not supported. Add the form link
to existing navigation (often `site/_navigation.html`). Starter sites include the
two pages; clean sites may need them added. Restart dev after changing app settings.
No `{app ...}` tag, old `contact` config or `/_flea/contact` alias is needed.

## Exact form contract

```html
<section class="container py-xl">
  <h1>Contact us</h1>
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
</section>
```

`{csrf}` renders the **whole hidden `_csrf` input**, not a token value. Keep it
inside the form as ordinary element content; Flea does not inject it for you.
Do not alter admin CSRF controls, remove exact Origin checks or bypass login.
Use a native URL-encoded POST, no multipart uploads or custom JSON endpoint.

Only `name`, `email`, `message`, and optional `website` are supported visitor
fields. The last is a hidden spam trap: leave it blank and never redisplay it.
Do not add submit-button names, subject fields, recipient controls, duplicate
fields or hidden settings; unknown/repeated fields fail validation. Name/email/
message are required, trimmed strings with the shown length limits. Keep form
controls developer-managed, outside snip regions. Surrounding page copy may use
optional snips or plain HTML.

| Template value | Purpose |
| --- | --- |
| `contact.data.name`, `contact.data.email`, `contact.data.message` | Safe redisplayed input after an error |
| `contact.invalid.name`, `contact.invalid.email`, `contact.invalid.message` | Field validation messages |
| `contact.error` | Safe operational failure message |
| `contact.flash.notice` | Optional success acknowledgment on thank-you |
| `app.*` | Public site globals; there is no `site.*` alias |

All values are escaped, never template source.
Validation redisplays the form with 422; operational failures preserve fields
and show a safe error. Do not replace these with an unconditional success page.
Success redirects to the mount's `thank-you` route. There is no delivery queue,
inbox, attachment support or automatic reply.

Thank-you page:

```html
<section class="container py-xl">
  <h1>Thank you</h1>
  {#if contact.flash.notice}<p role="status">{contact.flash.notice}</p>{/if}
  <p>Thank you for getting in touch.</p>
  <a href="/">Back to website</a>
</section>
```

Keep a static acknowledgment: flash is optional, expires and is consumed once.

## Mail templates and verification

Private `mailers/contact.email.html` and `mailers/contact.email.txt` receive
`email.name`, `email.address`, `email.message` and public `app.*` values. Preserve
the existing mail layout and add these files from a disposable matching-version
starter if missing. They do not control recipients, sender, subject or Reply-To.
Never put them under public `site/`.

Run `flea status`. With `dev.email_preview` enabled, verify in local dev: form and
navigation, invalid-field feedback, preserved inputs and a valid submission to
the thank-you route with a private email preview. Use synthetic data. A local
preview sends no real email and proves no production delivery. Do not write
private preview links or message contents into the repo. Stop task-owned dev
servers before handoff.

Production requires HTTPS, configured Resend credentials and an authorized
sender. For authorized real local delivery, disable email preview, use
`flea email` and restart dev; production ignores the preview setting. Ask only
for missing recipient/setup decisions and report any unverified real delivery.
Do not add dependencies or an alternate mail API to make the bundled form work.
