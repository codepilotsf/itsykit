# Owner-browser review

Read before showing direction boards, concept comparisons or final review artifacts.
Use the owner's normal external browser by default. Keep any established preference
in the owner's supported memory/environment settings when available; do not invent
Impeccable configuration fields. An existing project or user-selected transport wins.

## Choose the route before publishing links

Record an `internal_preview_url` for agent inspection and an `owner_review_url`
for the owner's device. Loopback works when server and owner browser share a
machine, or an established forwarding route makes it work. A remote host's
`localhost`/`127.0.0.1` is otherwise unsuitable, even if T3 can load it.

For a remote host, discover the active existing Tailscale route with read-only
`tailscale status --json`: use the running node's `Self.DNSName` and its specific
`Self.TailscaleIPs` address. The helper below does this without policy changes.
Do not hard-code example hostnames, IPs, ports or home directories. Preserve
another selected existing private route instead of replacing it with Tailscale.
Bind only the intended interface, never `0.0.0.0` or `::`.

No reachable route is a transport limitation. Prepare the portable export first,
then request only the missing access/infrastructure decision. This workflow grants
no permission to create public tunnels, provision hosting, install dependencies,
change firewall/Tailscale policy or deploy a site. Do not silently substitute
T3-only review after the owner has chosen an external browser.

## Prepare standalone concepts

Author a dedicated project-local `.impeccable/exports/<review>/` containing
`index.html`, concepts, expanded images, mobile views and required assets. Keep it
outside public site assets. Never serve the repository, source tree or credentials.
Include only intentional review files, with prompt provenance retained outside
the served directory when it contains private context. Embed or bundle images,
permitted fonts and license notices. No temporary signed URLs or private filesystem
paths. Preserve no-build projects; the helper uses Python's standard library.

Use relative navigation throughout, including CSS URLs and image expansion links;
repair production-root assumptions such as `/concepts/a`. Static pages should use
classic local scripts rather than module imports or fetches that require an HTTP
origin. Hide or label the ZIP download as online-only when opened as `file://`;
the archive cannot contain itself. Verify native text editing, every gallery
choice and mobile views. Label illustrative controls, static motion concepts and
any persistence/playback limits.

Run the shipped helper from the consuming project, substituting the runtime's
authored design skill path for `<design-skill>`:

```sh
python3 <design-skill>/scripts/browser_review.py pack \
  --project-root . --export-dir .impeccable/exports/directions
python3 <design-skill>/scripts/browser_review.py serve \
  --project-root . --export-dir .impeccable/exports/directions --route tailscale
```

These are **authored helper commands**, not new Impeccable launcher flags. `pack`
checks declarative HTML/CSS references and makes `review.zip` with `index.html` at
its root. It does not transform concepts, acquire assets or prove arbitrary JS
works offline. Inspect dynamic URLs and exercise interactions separately. A ZIP
does not grant asset redistribution rights.

For a same-machine browser use `--route local`. For another existing route supply
`--bind <specific-local-IP> --owner-base-url <existing-http-origin>` instead of
`--route`; this helper serves direct HTTP only, with no proxy, TLS or base-path
configuration. A supplied origin's port is used, or use `--port` with a matching
port. Auto-discovered routes select an available port and report its actual URL.
Forwarding, HTTPS or path-prefixed routes need their established transport's own
server integration; do not change access settings to fit this helper.

`serve` validates the export and refreshes its ZIP, then records the served root,
specific bind, actual port, PID, task owner, separate URLs and authenticated stop
command in a sibling `<review>.server.json`. It serves only the validated files,
refuses symlinks and directory listings, and reuses a healthy server with matching
configuration and owner. Keep this record private and outside the export. Review
content is accessible to devices already permitted on that interface; it is not
an authenticated application. Use a route appropriate to the content's sensitivity.

Set `--owner <task-id>` when the runtime supplies an identity; otherwise the export
path identifies the task. After changing export files, stop and re-serve to refresh
the allowlist and ZIP. Do not run concurrent pack/start/stop operations on one export.
The helper's stop command uses the recorded server's secret, never a bare PID kill.
Do not stop a server owned by another task.

## Keep the consultation intact

For live question boards, use the runtime-discovered global launcher with project
cwd. Inspect the installed version's actual capabilities before selecting a
transport; do not document guessed `serve-question --host`, `--bind` or base-URL
flags. The installed 4.8.0 reference documents `--start`, `--payload`, `--key`,
`--wait` and `--update`; it does not establish an external bind capability.

Where an established route can carry the live board, forward the same question
and its scoped assets/endpoints. Verify answer submission, callbacks, update
polling and origin/host checks end to end before presenting it. Keep its key,
desktop/mobile comparison, toggle and submission behavior. Serve before generating
declared comps; retain local assets and exact prompt sidecars. Collect using
`serve-question --wait --key <key>` and obey its exit codes: an unanswered choice
stays pending. Re-roll with `--update --key <same-key> --payload <file>`; do not
start another consultation. Apply T3's required preview opening and any narrowly
scoped component-review force override when those project rules apply.

The static helper does **not** proxy live Impeccable servers or implement their
answer protocol. If the installed launcher/established transport cannot expose
the live board, disclose this and use a standalone export for external review,
explicitly asking the owner to answer in the conversation. Retain the original
pending key and consultation record. Record the exact answer, selected option,
round and source (`conversation`) in the project's consultation record; reconcile
through a supported launcher submission method if one exists. Never fabricate
an engine answer or mark its gate accepted by editing internal state. A required
engine-only acceptance gate remains pending until supported submission works.
Revisions and re-rolls update the same export and pending consultation. Opening,
downloading, timeout and an agent recommendation are not an answer.

For app/runtime-dependent previews, provide the reachable running app through its
existing transport. Label offline material as a limited snapshot and state the
omitted runtime behavior. Do not imply a static ZIP reproduces the app.

## Verify, present and retain

Check the owner URL's index, every concept, image expansion, mobile view and ZIP.
Inline chat boards and their buttons must use that same reachable destination or
portable relative paths, including after re-rolls. An agent-side HTTP 200 or T3
browser check proves server behavior only. Owner/device evidence proves external
Safari access; state when that remains unconfirmed. A prior owner-confirmed route
is useful evidence, not proof of every new artifact or current device connection.

Extract the ZIP and open `index.html` as `file://` with networking disabled. Verify
assets, fonts, relative navigation and applicable interactions. Keep these results
separate from runtime-preview and actual-device evidence.

Leave requested review servers running while owner review is pending. Preserve
the lifecycle record for later cleanup; stop only this task's server after review
completes, using the recorded command. Close only temporary agent inspection tabs.
Preserve owner tabs, unanswered boards and unrelated processes. Visual selection
and review delivery grant no deployment or infrastructure permission.
