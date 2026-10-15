# Repository map

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

New sites include a short AGENTS.md, shared/authoring skills in .agents/skills/,
version-specific .agents/site-requirements.md, and focused Blog/Contact/import
guides in flea/skills/. Keep these in Git but exclude them from runtime loading,
public assets and deployment via sitefiles. This does not
install or assume a particular agent plugin. Update skill conventions alongside
template changes; test imports without publishing or using live credentials.

Read `docs/agents/architecture.md` for dependency direction, trust boundaries, transaction
ordering, and review follow-ups. Keep it current when those responsibilities change.

- `cmd/flea/main.go`: executable entry point.
- `cmd/flea-server/main.go`: standalone persistent server entry point.
- `internal/host/`: strict server JSON config, hostname routing, readiness, and shutdown.
- `internal/setup/`: interactive SSH bootstrap, installer, and systemd unit.
- `internal/serverbundle/`: compressed embedded Linux servers and verified extraction.
- `internal/setup/deploy.go`: native SSH/SCP deployment and local deployment settings.
- `internal/release/`: bounded, confined runtime-only archive packing and extraction.
- `internal/admin/`: privileged release activation, lock, recovery journal, and probes.
- `internal/admin/sites.go`: read-only site inventory and unregister (retaining files).
- `internal/setup/server.go`: SSH-only server-wide site listing and removal over an explicit target.
- `internal/editor/`: inline editing, in-memory OTPs, persistent rolling sessions,
  CSRF and external same-origin editor assets. Sign-in markup lives in templates/site/admin/.
- `internal/editstore/`: conflict-checked content saves and validated snapshot publication.
- `internal/mailer/`: provider-neutral Sender and Resend HTTP adapter; never expose provider errors or secrets.
- `internal/state/`: the server-side file lock shared by browser saves and deployment.
- `internal/cli/cli.go`: command handling and site creation.
- `templates/`: canonical starter files, mirroring the generated site structure.
- `templates/embed.go`: embeds HTML, CSS, SVG, JSON, and README files in the binary.
- `internal/cli/cli_test.go`: CLI tests and executable/server smoke test.
- `internal/site/project.go`: loading, validation, routes, snapshots, and rendering.
- `internal/site/routes.go` and `data.go`: shared route conversion and JSON-directory loading.
- `internal/site/regions.go`: source-preserving region scanner and lexical references.
- `internal/site/extraction.go`: dev-only `LoadEditing` and snip extraction orchestration.
- `internal/site/snip_transaction.go`: content-first extraction journal and replay.
- `internal/site/files.go`: pinned project reads and durable atomic writes, also used by icons.
- `internal/admin/content.go`: remote content merge/staging; activation retains the deployment lock.
- `internal/setup/ssh.go`: shared SSH/SCP connection policy and target formatting.
- `internal/release/protocol.go`: shared CLI/server archive compatibility marker.
- `internal/site/template.go`: template parser, partial resolution, and data lookup.
- `internal/formguard/`: stateless anonymous form CSRF, separate from admin auth.
- `internal/appapi/`: mini-app protocol, mounts and bounded authenticated HTTP transport.
- `apps/contact/`: included Contact validation, manifest and canonical contract.
- `internal/contact/`: site-bound single-use email capability and shared budget.
- `internal/appmanager/`: managed child launch, readiness, backoff and cleanup.
- `internal/appruntime/` and `cmd/flea-contact/`: private child transport and entry point.
- `internal/appbundle/`: matching-platform Contact executable; build before its parent.
- `internal/site/apps.go`, `app_bridge.go`, `flash.go`: registration, public dispatch and private flash.
- `internal/site/icons.go`: on-demand pinned Lucide downloads, private cache,
  geometry-only SVG validation, and icon directive compilation.
- `internal/site/icons_test.go`: offline icon resolution, escaping, and security tests.
- `internal/site/server.go`: HTTP serving, reload watching, and URL validation.
- `internal/site/site_test.go`: rendering, routing, isolation, and reload tests.
- `bin/flea`: local build output; rebuilding is necessary for CLI changes to take effect.

Source CSS is loaded from the immutable release. `internal/site/managed.go`
provides confined fixed-path file helpers. Theme editor/overlay/sync code has
been removed; snip synchronization remains in setup/admin/syncstate.

## Blog

`internal/snips/blog.go` validates post records; `internal/site/blog.go` owns
published routes/views; `internal/editor/blog.go` owns private forms and uploads;
`internal/editstore/blog.go` saves under the shared content lock.
`internal/syncstate` compares/merges whole posts. Canonical private blog/admin
templates live under `templates/site/blog/` and `templates/site/admin/`.
See [Blog](../blog.md) for the authoring contract.
