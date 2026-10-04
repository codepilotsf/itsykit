# Contact app implementation handoff

Status: implemented locally, unreleased. October 4, 2026.

The canonical usage and API contract now lives in
[apps/contact/README.md](../../apps/contact/README.md). The broader
[app plan](apps.md) still contains future work; Rails is not the default backend.

## Settled in this batch

- Separate bundled Go executable, one child per enabled site, with one private
  extracted executable shared within the parent process. No compiler on the VPS.
- Main-config registration at apps.contact: mount, required email.to, optional
  email.from. Admin/editor remain core. No old config or route aliases.
- Fixed render-only form and thanks routes. POST at the mount forwards to POST /.
  Generic read-route declarations remain deferred until an app needs them.
- Flea App Protocol v1, flat form JSON, request-local data/invalid/error/flash
  envelope, explicit {csrf} and signed anonymous visitor protection.
- Mount-relative thanks redirect and bounded visitor/host/app/destination flash:
  256 entries/site, five-minute expiry, consume after successful GET rendering.
  HEAD/assets/other pages cannot consume it; failed renders release the claim.
  Concurrent tabs share one app cookie; last response cookie wins.
- Flea owns mail credentials, private templates and dev previews. The child uses
  an independent, site-bound bearer capability, admitted for exactly one matching
  submission. Visitor email is Reply-To, never a sender or recipient setting.
- Inherited bound loopback listener (fd 3), credentials (fd 4), parent-liveness
  pipe (fd 5), authenticated readiness, bounded restart and shutdown. No secrets
  in argv/environment/logs/site files. No migration during startup or recovery.
- Unknown/repeated Contact fields fail validation. A populated spam trap with
  otherwise valid fields follows the success flow without sending.
- Unavailable services or email return safe 503 with bounded name/email/message
  redisplay. Site/IP limits return 429. No automatic POST retry.

## Acceptance still requiring an appropriate environment

- Email-client HTML/text appearance and a real default-sender send.
  On October 4, the isolated Linux real-email check sent one approved message
  through Resend using the app sender override while leaving the different
  server default unchanged. Provider acceptance, HTTP 303 and the thank-you
  flash passed. Sean confirmed inbox receipt, From `sean@fleacms.com`, and
  To/Reply-To `sean@codepilot.com`. HTML/text appearance and a real default-sender
  send remain unverified. See [the opt-in check](../../scripts/contact-email-check/README.md).
- Longer Linux/systemd soaks and restart-budget exhaustion. The initial isolated
  startup, single-child crash, shutdown and parent-death checks passed; see the
  [Linux Contact report](../benchmarks/2026-10-04-contact-apps.md).
- Representative per-app memory/descriptor costs and deployment peaks.
- Real-browser form accessibility and optional ZJAX enhancement.

Local regression tests cover ordinary HTML success, invalid fields, delivery
failure, nested mounts, scoped flash, host aliases/recipients, dev preview, LAN
protection, authenticated transport, child parent-exit cleanup and extraction.
Use recorded check results for exact evidence, not this design summary.

## Remaining scope

General custom-app installation, remote registration and remote capability
access remain planned. Storage placement, SQLite schema changes, dynamic blog
routes, cross-page app reads and ZJAX are separate decisions. No new provider,
public email relay, login bypass, compatibility layer or second email-template
system is implied. No deployment or real mail is authorized by this handoff.
