# Working on FleaCMS

FleaCMS is a small Go website server using plain HTML, local assets and JSON.
This is the framework repository; `templates/` is the canonical starter.

## Start here

- Inspect `git status --short` before editing. Preserve existing modified,
  deleted and untracked work. Do not reset, bulk-stage or commit without permission.
- Use [next-steps.md](docs/agents/next-steps.md) for the agreed V1 scope, sprint order and
  release criteria, and [checklist.md](docs/agents/checklist.md)
  for manual QA; preserve the user's checklist answers. Neither authorizes
  unrelated implementation, deployment or publication.
- Inspect relevant code and tests. [README.md](README.md) describes supported
  behavior; [spec.md](docs/agents/spec.md) also contains future plans. Resolve disagreements
  against current code rather than assuming planned features exist.
- Use one agent unless the user authorizes delegation. Search narrowly with
  `rg`, read relevant sections, and keep logs and progress reports focused.

## Non-negotiable rules

- No new backward compatibility, legacy aliases, format fallbacks, migrations
  or support for superseded behavior without asking the user. Existing legacy
  support is not permission to add more.
- Keep Go 1.26+, minimal dependencies and ordinary HTML/JSON. No npm/frontend
  build step. Do not add dependencies without a concrete need.
- Preserve custom site layouts and the selected Skip logo
  (`templates/site/images/mark.svg`). Framework fixes belong here.
  Ask for the current test site's path before modifying it; do not assume
  `../bookworm` is the site in scope.
- Keep published Projects immutable, reads confined, and request paths separate
  from filesystem paths. Never compile JSON or owner values as template source.
  Stream large assets/photos from validated handles; preserve mutation checks
  and snapshot lifetime rules in the runtime guide.
- Preserve session, exact Origin, CSRF, CSP, host isolation and content-version
  checks. No login bypass or weakened security to make tests pass.
- Normal deployment preserves live editable content and publishes committed CSS. Explicit sync,
  recovery and uploads must retain their validation, locking and publication order.
  Never prune retained content or media automatically.
- Keep secrets, sessions and private runtime state out of Git, public assets,
  logs, releases and backups. Use `internal/sitefiles` for runtime inclusion;
  public URL privacy is a separate boundary.
- Do not deploy, push tags or publish artifacts unless requested. Passing local
  tests does not establish real email, Ubuntu/systemd, ACME or release acceptance.

## Read the requirements for your task

Before editing a subsystem, read its linked guide and applicable feature docs.
For work crossing subsystems, read each affected guide. These are requirements,
not optional background. Do not load every guide for an unrelated small change.
Read [ARCHITECTURE.md](docs/agents/architecture.md) for implementation changes involving
dependencies, trust boundaries, loading, transactions or publication.

| Task | Required guide and references |
| --- | --- |
| Find code ownership and entry points | [Repository map](docs/agents/repository.md) |
| Site structure, config, templates, extraction, snips, starter/import conventions | [Templates and content](docs/agents/templates.md) |
| Images, videos, collections, multipart saves and media transport | [Media and collections](docs/agents/media-collections.md); applicable [photos](docs/photos.md), [videos](docs/videos.md), [multisnips](docs/multisnips.md) |
| Pull and explicit snip sync; source CSS | [Synchronization and styling](docs/agents/sync-theme.md) |
| HTML/CSS, Prewind, icons, starter styling | [Styling and icons](docs/agents/styling.md); `../prewind/README.md` |
| CLI help/dev, file watching, routing, diagnostics, contact, email rendering | [Runtime](docs/agents/runtime.md); [email templates](docs/email-templates.md) when relevant |
| SSH/setup, deploy, host/TLS, archives, rollback, backup, restore, server site administration, upgrade | [Deployment](docs/agents/deployment.md); [recovery](docs/recovery.md) |
| Owner toolbar, authentication, sessions, onboarding, mail configuration, editor prototype | [Editor](docs/agents/editor.md); media guide for those modes |
| Versions, packaging and releases | [RELEASING.md](docs/agents/releasing.md) and version rules below |

## Site and UI essentials

Generated sites use `site/` for public templates/assets, `mailers/` for private
email templates, `flea/` for content/data/media/icons/skills, and
`.flea.config.json` for configuration. Runtime state is private in
`flea/runtime/`. Globals use `app.*`; there is no `site.*` alias.
Bundled Contact uses the Flea App Protocol and explicit `{csrf}` visitor forms.
Optional Blog uses the built-in content/editor snapshot; read [Blog](docs/blog.md)
before changing posts, routes, rich bodies or synchronization. See the
runtime/template guides for both applications.
Do not change admin form contracts or infer support for arbitrary app installation.
Generated-site guidance lives in `templates/AGENTS.md` and the bundled skills in
`templates/flea/skills/` (`import-brochure` and `fleacms-site`); keep it consistent
with starter syntax, committed, and excluded from serving/deployment.

New sites enable `dev.email_preview`. Only `flea dev` opens actual outgoing mail
in private browser previews; production ignores this setting. Keep the first-run
prompts limited to site name and optional admin email, asked once on first run.
User-facing copy says "admin"; code and DOM identifiers keep "owner". See the
runtime guide.
Dev defaults to loopback. A new dev invocation gracefully replaces the same site's existing
dev server through private local control; never interrupt pull/reset or bypass
the development lock. See the runtime guide.
`--lan`/`-l` explicitly enables full trusted-network development over HTTP;
preserve the shared Host/peer gate and ordinary auth/CSRF.
`flea new` creates an initial Git commit on main unless `--skip-git` is used;
see the runtime guide for failure recovery and remote setup. It defaults to a
clean layout/index. `--starter`/`-s` adds demo pages and a six-person fictional team. `flea reset` confirms before resetting all local files
except `.git`, with a private sibling recovery copy; it never deploys.
Starter navigation links Team and Contact, with editable page copy and footer.
Keep form controls developer-managed; see the templates guide for starter scope.
Contact is enabled only by apps.contact (mount and email.to); its starter form
uses `{csrf}` and mount-relative `thank-you`. See apps/contact/README.md for the
protocol, process and email boundaries. Custom/remote app registration is planned.
Keep access pages simple and visually consistent using the shared `admin-access`
layout; visual polish must not change authentication or code-autofill behavior.
Starter styling uses Inter, warm neutral surfaces and forest-green primary actions,
all driven by theme tokens. Styles are developer-managed; normal deployment publishes committed CSS.

Use Prewind's global-first conventions and supported utilities, not assumed
Tailwind classes. Use `container` without `max-w-full`, theme tokens and fluid
spacing. Preserve vendored CSS/JS. Use `{icon name}` and the flat
`flea/icons/name.svg` cache; never generate toolbar SVG paths in JavaScript.

Owner edits happen on the page with explicit Save/Cancel. The Edit switch defaults on
and persists per tab across pages. The second row slides beneath the top bar and
collapses outside Edit. Save/Cancel remain on the second toolbar row
and keep editing on. Page Save updates in place from the committed server receipt,
rebasing Cancel without reloading or replacing editable wrappers. Use green 2px dashed shadow-root overlays with 6px
corners, never layout-changing borders on content. Keep keyboard focus visible on controls.
Collection photos use a thumbnail picker with private upload previews; keep
dialog Done separate from publishing with page Save changes.
The declarative toolbar and its icons render before bootstrap; do not recreate
its shell. Preserve drafts after errors and during source reloads.

`flea status` shows local validation first, then automatically checks the current
branch's saved remote. Keep it read-only; see the runtime and sync guides.

## Verification

Choose verification by behavior and risk, not line count. A one-line security
change can need broad checks; a display-only change usually does not.

- Documentation, skill instructions and copy-only changes: review the diff for
  accuracy, consistency and preserved requirements; check links when changed.
  No runtime suite or build is required.
- Small, isolated implementation changes: format changed Go files, add or update
  meaningful regression tests, and run tests for affected packages. Build once
  when executable behavior changes.
- Security, persistence, deployment behavior, concurrency, shared infrastructure
  or broad changes: run the full checks below plus relevant integration checks.
  Presentation-only changes to administration commands can use the isolated-change
  path when authorization, confirmation, target selection and mutation safeguards
  remain intact and are reviewed. Broaden checks when downstream effects are unclear.

Full checks, run from this repository:

```sh
go test ./...
go vet ./...
bash scripts/build.sh
```

For concurrency, reload or shared-state changes, also run race tests for affected
packages; use `go test -race ./...` when effects cross package boundaries or are
unclear. Use targeted tests while iterating and required broader checks once on
the final implementation. After corrections, rerun checks invalidated by the
changes, not every passing check automatically. Reviewers inspect existing
evidence rather than repeat suites without a specific concern or evidence gap.
When using DeepSeek delegation, DeepSeek owns builds, tests and troubleshooting;
the lead reviews the diff and results. Report what was and was not checked.
Before publication, run the full release checks against the final candidate as
specified in [RELEASING.md](docs/agents/releasing.md).
Loopback tests need bind permission; a sandbox bind failure is not a product failure.

On this Mac use `/usr/local/go/bin/go` if needed, with
`GOCACHE=/tmp/flea-go-cache` when the default cache is restricted. For the
build script, set `GO_BINARY=/usr/local/go/bin/go` if necessary.
Browser tests use
`FLEA_TEST_BROWSER='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'`.

When a build is required, use `scripts/build.sh`: it refreshes
the matching Contact child before each parent, then both embedded Linux servers
before building `bin/flea`. Plain `go build` can
embed stale bundles. Do not overlap this build with tests compiling the bundles.
For starter, packaging, loading or site-creation changes, generate a fresh
temporary site, compare relevant starter files, and finish with `flea status`
there. This smoke test is not required for unrelated CLI tweaks. Exclude local agent guidance
from comparisons unless deliberately changing it. Restart dev after rebuilding;
site-file reloads do not reload Go code. Verify installed binary paths before
assuming they point to this checkout.

## Keep documentation current

Update this guide and affected `docs/agents/` references in the same task when
architecture, template syntax, CLI, starter conventions, supported features or
workflow changes. Keep the root concise; put subsystem requirements in the
linked guide instead of duplicating them here. Update README.md and applicable
docs/agents/spec.md sections for behavior changes, docs/agents/architecture.md
for responsibility changes, and generated-site/import guidance for starter
changes.

Keep docs/agents/next-steps.md for priorities, open decisions, pending
verification and deferred work. Remove stale notes; completed changes belong in
CHANGELOG.md, usage in README.md. Keep planned behavior distinct from implemented
behavior; public installation remains unverified until release hosting and native
checks pass.
See docs/installation.md for local update behavior and release trust.

Use Semantic Versioning from `internal/buildinfo/version.go`. Local VERSION
overrides omit the leading v; release tags are v plus the recorded version.
Keep user-facing changes under Unreleased until a dated release. During 0.x,
use patch versions for fixes and minor versions for features or breaking changes.
The deployment protocol is separate: see `internal/release/protocol.go`.

## Communication

Use natural, brief English. Avoid AI cliches, emojis and em dashes. Question
assumptions when warranted. CLI messages should explain what happened and the
next action, with readable spacing; avoid raw subprocess commands, exit codes
and Git jargon in routine errors without hiding useful diagnostic context.
Finish with what changed, what was checked and any remaining limits.
