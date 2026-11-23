# Site-config domain redirect rollout

Verified on 2026-11-15 UTC after explicit approval to update the shared Itsy
Server and activate these redirects. Implementation: `b7bb5cc`; the subsequent
integration merge only preserves upstream repository documentation, licensing
and navigation changes.

## Configuration and activation

Policy is declared in `website/itsy.config.json`, under each branch's
`domain_redirects` deployment setting:

- `main`: `fleacms.com` and `www.fleacms.com` redirect to `itsykit.dev`.
- `t3/reframe-flea-html-focus`: `itsyweb.dev` redirects to `itsykit.dev`.

The pinned candidate CLI completed the normal `itsy update-server` transaction,
including installed-site validation and readiness. The installed Linux AMD64
server SHA-256 is
`336b298236b8f56b6d993390afe8c9832af7c18a97aa1d35c3f2d5958712b6e0`.
The existing socket-activated service was retained; both `flea.service` and
`flea.socket` remain active. There are no pending deployment or upgrade journals.

Normal `itsy https` activation reconciled each branch's declarations, without
republishing website content:

- Main activation: `20261115-012235-857d24777b86fb36eb61eaee`.
- Review activation: `20261115-012324-b1e65ad5c00761347b68c3a8`.

The host config was generated through those deployment transactions. No manual
host-policy patch was installed and no site-specific domains were added to
server implementation code. Comparison with the saved pre-rollout config found
only the intended redirect entries and activation revision changed. All 35
hostname registrations, active release/content paths, TLS, download stores,
indexing and branch/commit metadata were retained.

## Verification

The candidate passed `go test ./...`, `go vet ./...`, `go test -race ./...` and
`bash scripts/build.sh`. Read-only validation as the production service account
covered all 35 live hostnames before installation. After activation:

- All 35 hosts returned healthy responses with their expected hostname and the
  current activation revision; HTTPS probes verified their configured certificates.
- All 24 public redirect checks passed: three source domains over HTTP and HTTPS,
  each with root, documentation/query, escaped-path/query and versioned-download
  requests. Each returned 308 directly to HTTPS on `itsykit.dev`, preserving the
  requested path and query.
- Nine followed metadata requests across the source domains matched the canonical
  installer, latest-version pointer and 0.22.0 manifest bytes.
- A complete Linux CLI download through HTTP on `fleacms.com` followed to HTTPS
  on `itsykit.dev`. All 29,270,178 bytes matched the published SHA-256:
  `16411183735071699ea5575d0e0990836fc55559d8b5de35f5b114c43a5351d9`.
- The canonical homepage returned 200. No public release artifacts were published
  or modified.

Private pre/post host configs and detailed verification records are retained in
`~/.local/state/itsykit-config-redirects-95pbj3cb/`. See the
[deployment guide](deployment.md#domain-redirects) for future config changes.

## Stale deployment regression and repair

After the initial verification, a separate website deployment of `1196741`
activated revision `20261115-012614-fbc134f732ea1bfca7135dd7`. Its older config
omitted `domain_redirects`, which the first implementation treated as removal.
The owner's report was reproduced: HTTPS on `itsyweb.dev` returned 200, while
HTTP redirected to HTTPS on the old hostname.

The current branch config was reapplied through `itsy https`, restoring the
redirect at revision `20261115-012808-3f49f0719198cbe1ed54553a` while preserving
the newer deployed website content. Fix `19a0f8a` makes missing/null redirect
settings fail before staging whenever the selected site has custom redirects.
An explicit empty map clears the final rule; the CLI transmits and preserves
that distinction when saving config. No site-specific server rule was added.

The fix passed the full tests, race suite, vet and canonical build and was
installed through normal `itsy update-server`. The resulting Linux AMD64 server
SHA-256 is `d7d76d78e311da24b62c3707f31bdbcbb5441b6cc418bd0304e248ee58e51d79`.
All installed-site readiness checks passed. An actual HTTPS administration
request omitting redirect settings was then rejected with the expected explicit
configuration error. Host config bytes and service PID were unchanged by that
rejected request. Detailed evidence is in the private artifact directory's
`stale-deployment-guard/live-guard-verification.json`.
