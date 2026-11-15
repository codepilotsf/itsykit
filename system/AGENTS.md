# Working on ItsyWeb

ItsyWeb (formerly FleaCMS) is a small Go website server using plain HTML, local assets and JSON.
This is `system/` within the ItsyWeb repository; `templates/` is the canonical
starter. The current public website is the sibling `../website/`, with its own
AGENTS.md. Paths and Go commands in this guide are relative to `system/`.

## Start here

- Inspect `git status --short` before editing. Preserve existing modified,
  deleted and untracked work. Never reset or bulk-stage unrelated work.
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
- Preserve custom site layouts and the approved ItsyWeb mark and logo
  (`templates/site/images/mark.svg` and `logo.svg`). Framework fixes belong here.
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
- Managed serving requires systemd's retained HTTP/HTTPS sockets; never silently
  bind after an invalid handoff. Keep the approved one-time standard-unit conversion
  under executable-upgrade recovery; see the deployment guide for its exact scope.

## Read the requirements for your task

Before editing a subsystem, read its linked guide and applicable feature docs.
For work crossing subsystems, read each affected guide. These are requirements,
not optional background. Do not load every guide for an unrelated small change.
Read [ARCHITECTURE.md](docs/agents/architecture.md) for implementation changes involving
dependencies, trust boundaries, loading, transactions or publication.

| Task | Required guide and references |
| --- | --- |
| Find code ownership and entry points | [Repository map](docs/agents/repository.md) |
| Site structure, config, templates, metadata, JSON data, starter/import conventions | [Templates and content](docs/agents/templates.md) |
| App model, JSON contract, custom managed packages, app-owned sessions | [Flea apps](docs/apps.md); [Contact](apps/contact/README.md) |
| Blog covers, multipart saves and media transport | [Media and collections](docs/agents/media-collections.md); [photos](docs/photos.md) and static [videos](docs/videos.md) |
| Pull and explicit Blog sync; source CSS | [Synchronization and styling](docs/agents/sync-theme.md) |
| HTML/CSS, Prewind, icons, starter styling | [Styling and icons](docs/agents/styling.md); `../prewind/README.md` |
| CLI help/dev, file watching, routing, diagnostics, contact, email rendering | [Runtime](docs/agents/runtime.md); [email templates](docs/email-templates.md) when relevant |
| SSH/setup, deploy, host/TLS, archives, rollback, backup, restore, server site administration, upgrade | [Deployment](docs/agents/deployment.md); [recovery](docs/recovery.md) |
| Blog editor, authentication, sessions, onboarding, mail configuration | [Editor](docs/agents/editor.md); media guide for those modes |
| Versions, packaging and releases | [RELEASING.md](docs/agents/releasing.md) and version rules below |

## CLI and site naming

Use `itsy` for all commands. The retired `flea` command prints the rename notice
and exits without dispatch, except `flea update`, which installs `itsy`. Preserve
the narrow read-only version probe required by already released Flea updaters.
New sites use `itsy.config.json`, `itsy/`, and `<!-- itsy` page settings. Existing
`.flea.config.json`, `flea/`, and `<!-- flea` remain explicitly supported; prefer
Itsy names when present and otherwise use the legacy name. Writes stay in the
selected files; never silently rename or migrate a site. These naming aliases
are the owner-approved exception to the general no-compatibility rule.
Installed server paths/accounts, wire identifiers, development/install locks,
release URLs and public artifact names retain their spelling for compatibility.
Protocol 30 is required before deploying the new site layout; an explicit
`itsy update-server` validates every installed site before upgrading. Old page-CMS
sites require an explicitly approved, backed-up source/content conversion first;
there is no automatic migration.

## Site and UI essentials

Generated sites use `site/` for public templates/assets, `mailers/` for private
email templates, `itsy/` for content/data/media/icons/skills, and
`itsy.config.json` for configuration. Runtime state is private in
`itsy/runtime/`. Globals use `app.*`; there is no `site.*` alias.
Bundled Contact uses the Flea App Protocol and explicit `{csrf}` visitor forms.
Optional Blog uses the built-in content/editor snapshot; read [Blog](docs/blog.md)
before changing posts, routes, rich bodies or synchronization. Contact is
bundled; other apps are managed local executables registered under `apps.<name>`
with a validated package at `itsy/apps/<name>/`, and they own their sessions and
storage. See [Flea apps](docs/apps.md) and the runtime/template guides for both
applications. Do not change admin form contracts or add remote-service,
package-registry or arbitrary-install support.
Generated-site guidance lives in `templates/AGENTS.md` and the bundled skills in
`templates/.agents/skills/` (core, delivery, coding orchestration, design, Prewind
and Flea authoring), plus focused `templates/itsy/skills/` guides (`add-blog`,
`add-contact-form`, `import-brochure`). Keep them consistent with starter syntax,
committed, and excluded from serving/deployment. Impeccable stays globally installed
and must never be copied or linked into a site or template. Personal conversation preferences
override the shared defaults; project requirements and permissions still apply.

New sites enable `dev.email_preview`. Only `itsy dev` opens actual outgoing mail
in private browser previews; production ignores this setting. Keep the first-run
prompts limited to site name and optional admin email, asked once on first run.
User-facing copy says "admin"; code and DOM identifiers keep "owner". See the
runtime guide.
Dev defaults to loopback and tries consecutive ports from 3000 when occupied,
with clear console output. Explicit `--port` stays fixed; 0 selects an OS-assigned port.
A new dev invocation gracefully replaces the same site's existing
dev server through private local control; never interrupt pull/reset or bypass
the development lock. See the runtime guide.
`--lan`/`-l` explicitly enables full trusted-network development over HTTP;
preserve the shared exact Host/port and private/Tailscale peer gate, including
detected hostname/local aliases, and ordinary auth/CSRF.
`itsy new <directory>` accepts paths, including `.` for the current directory.
Existing directories must be completely empty, including hidden files and `.git`;
reject file targets and directory symlinks without writing or prompting to overwrite.
`itsy new` creates an initial Git commit on main unless `--skip-git` is used;
see the runtime guide for failure recovery and remote setup. It defaults to a
clean layout/index. `--starter`/`-s` adds demo pages and a six-person fictional team. `itsy reset` confirms before resetting all local files
except `.git`, with a private sibling recovery copy; it never deploys.
Starter navigation links Team and Contact, with ordinary HTML page copy and footer.
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
`itsy/icons/name.svg` cache; keep icon geometry in validated SVG files.

Blog admin editing uses private templates, explicit Save/Publish/Unpublish and
cover photos. Ordinary pages use HTML/JSON and leading comment metadata. There
is no page CMS or theme editor. Keep authentication, CSRF, snapshot versions and
content preservation; see the editor and Blog guides.

Host indexing belongs to optional server `indexing` policy, never site `app`
values. Preserve it on deploy/sync/code recovery; removal clears only owned hosts.
See [host indexing](docs/host-indexing.md) and the deployment guide. Host indexing
was introduced in protocol 27; keep release rollout separate from site deploy.
Site migration redirects use top-level config `redirects` and 301 GET/HEAD;
protocol 28 prevents older servers from ignoring them. See the runtime guide.

Optional `seo.url` supplies the production HTTPS origin for snapshot-generated
sitemaps, robots discovery and missing canonical tags; exact `seo.exclude` paths
remain crawlable with noindex. Preserve custom discovery files and canonical
markup, omit non-public routes and never derive production URLs from requests.
Read [SEO](docs/seo.md) before changing this behavior; host indexing stays separate.

`itsy status` shows local validation first, then automatically checks the current
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

Full checks, run from this repository’s `system/` directory:

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

For constrained temporary storage use the private, disk-backed per-job pattern
in [development storage](docs/development-storage.md); retain review artifacts
outside scratch. Never clean unrelated jobs or change global storage policy.
On this Mac use `/usr/local/go/bin/go` if needed, with a writable disk-backed
`GOCACHE` when the default cache is restricted. For the
build script, set `GO_BINARY=/usr/local/go/bin/go` if necessary.
Browser tests use
`FLEA_TEST_BROWSER='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'`.

CLI builds and public releases support Apple Silicon macOS and x86-64 Linux.
Do not publish a Linux ARM64 CLI. Both CLIs retain the cross-compiled ARM64 server;
record native ARM64 server acceptance separately rather than requiring an ARM64
Linux runner for every release. Preserve historical published downloads.
When a build is required, use `scripts/build.sh`: it refreshes
the matching Contact child before each parent, then both embedded Linux servers
before building `bin/itsy`. Plain `go build` can
embed stale bundles. Do not overlap this build with tests compiling the bundles.
For starter, packaging, loading or site-creation changes, generate a fresh
temporary site, compare relevant starter files, and finish with `itsy status`
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

## Owner-authorized Git delivery

Routine scoped commits, task-branch pushes, PR/MRs and verified merges are
owner-authorized without per-step approval. Respect branch protections, required
reviews and CI. Honor explicit stop-at-review requests. Preserve unrelated work;
use scoped revert commits when an undo is requested. Deployment, releases,
production mutations and destructive Git operations need separate authorization,
including when triggered automatically by pushing or merging.

Ask before unresolved consequential product, creative, architecture or integration
choices. Inspect existing context, recommend an option and wait before dependent
implementation. Proceed autonomously for routine details within approved direction.

## Shared agent workflow (Kirin rollout)

At task start, read [.agents/skills/core/SKILL.md](.agents/skills/core/SKILL.md).
Use its current consultation and delivery defaults for this project. This section
supersedes older generic question-batching and per-commit/per-merge approval wording;
preserve all project-specific product, stack, security and deployment constraints.

- Planning and requirements: read
  [efficient-collaboration](.agents/skills/efficient-collaboration/SKILL.md).
- Delivery and proportionate verification: read
  [task-delivery](.agents/skills/task-delivery/SKILL.md).
- Drafting or editing copy: read [copywriting](.agents/skills/copywriting/SKILL.md).
- For authorized delegation or independent review: read
  [deepseek-orchestrator](.agents/skills/deepseek-orchestrator/SKILL.md).
  Loading it does not itself authorize delegation.
- Styling: read [prewind](.agents/skills/prewind/SKILL.md).
- Design or UX: read [design](.agents/skills/design/SKILL.md), alongside any
  project-local design context. Impeccable stays globally installed.

This directory contains the framework and generator. Use its maintenance guides;
site/application authoring skills in its templates apply to generated projects.

Ask before unresolved consequential product, creative, architecture or
integration choices. Proceed autonomously within approved direction. Routine
scoped commits, task-branch pushes and verified merges have owner authorization,
subject to required reviews, CI and branch protections. Deployment, releases,
production mutations and destructive Git actions need separate authorization.
Reuse passing verification unless changes or unresolved concerns invalidate it.
