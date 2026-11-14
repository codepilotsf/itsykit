# Session app

An independent, standard-library example of a custom Flea app that owns its own
server-side sessions. It imports no Flea package and is not part of the Flea
binary. Use it to see and exercise the full contract: managed launch, handshake,
form JSON, and the `session` bridge.

> The `login` action is a demonstration only. It performs no credential check and
> must never be treated as real authentication.

## What it does

- `GET /` returns `{"data":{"authenticated":true|false}}` for the browser's
  opaque token read from the `X-Flea-Session` header.
- `POST /` accepts a flat JSON object with an `action` field:
  - `login` sets a session with `max_age` 3600.
  - `login-short` sets a one-second session (for expiry checks).
  - `rotate` revokes the current token and issues a new one.
  - `logout` clears the session with `session: null`.
  - anything else returns `422` with a field message.
- Sessions persist to `sessions.json` in the private storage directory Flea
  passes at launch, so they survive an app restart. Expiry is enforced by the
  app, not by Flea. Writes are durable and fail closed: if the store cannot be
  written, the request fails with `503` instead of silently dropping the change.
  The store is bounded to 1000 live sessions.

The token appears only in the `session` value Flea consumes; it is never placed
in `data`, templates, URLs or logs.

## Launch contract

Flea launches the executable with three inherited descriptors:

- fd 3: an already-bound `127.0.0.1` TCP listener.
- fd 4: one JSON object with exactly these five keys, all strings:
  `{"Name":..., "Token":"<64 hex>", "Storage":"<dir>", "CapabilityURL":"", "CapabilityToken":""}`.
  The last two are empty for a custom app and are used only by the bundled
  Contact app; unknown fields are rejected.
- fd 5: a parent-liveness pipe; EOF means shut down.

The app must require `Authorization: Bearer <Token>` on every request, deny
others first, answer `GET /_flea/handshake` with its name, version and
`protocol: 2`, and never expose the handshake through the public mount.

## Build and register

A package declares one executable per platform, keyed `<goos>-<goarch>`. Build
for every platform you run (your development host and the server), then place
the manifest and the built files under the site:

```sh
go build -o /path/to/site/flea/apps/session/session-darwin-arm64 ./samples/session-app   # your Mac
GOOS=linux GOARCH=amd64 go build -o /path/to/site/flea/apps/session/session-linux-amd64 ./samples/session-app
```

```text
site/flea/apps/session/app.json
site/flea/apps/session/session-darwin-arm64
site/flea/apps/session/session-linux-amd64
```

The manifest lists exactly the platforms you built:

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

Enable it in the site's `.flea.config.json`:

```json
{
  "apps": {
    "session": {"mount": "/session"}
  }
}
```

Create `site/session/index.html`. A form posts to `/session` and includes
`{csrf}`; page values use `{session.data.*}` and `{session.invalid.*}`. Restart
`flea dev` after rebuilding the binary.

The manifest is validated at load. Flea rejects a mismatched name or protocol, a
non-executable, symlinked or hardlinked file, a file that is not a native binary
for its declared platform, a path that escapes the package, unknown manifest
fields, and a package that does not provide an executable for the running
platform. The deployment preflight checks the immutable staged snapshot and
rejects a package that does not provide the server's platform, or provides a
mislabeled one, before anything is uploaded. At startup the running service is
verified by its handshake.

## Notes

- App storage lives in Flea's private per-site runtime directory (beside admin
  sessions), outside every immutable release, and is excluded from releases and
  backups. Back it up yourself.
- The sample is deliberately file-based and single-process. A real app should
  choose a storage design that suits its load and locking needs.
- An app session never grants Flea admin or editor access.
