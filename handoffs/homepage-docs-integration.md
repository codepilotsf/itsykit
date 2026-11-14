# Homepage and Docs integration

Integrated the approved homepage (`3ca55f6`) and documentation theme (`3b4bd1e`)
with the completed page-CMS-removal handoff at
`9d856797e8f8a7b9658ae82bf7ba98a227596dd9` on
`t3/modern-framework-homepage`.

The selected “Plain HTML. Superpowers included.” composition, working demos,
install-to-HTTPS workflow, coding-agent section and retained build-pipeline/server
messages remain. Homepage copy and metadata now live directly in HTML. Public
page-editor mounts and saved page metadata remain removed. Retired editing/media
guide links were removed from the homepage; Contact, Blog and managed apps remain.
The newer guide-title links, scrolling chapter selection and explanatory docs
content use the approved navy/ice theme.

## Local verification

- Framework tree is byte-for-byte identical to `9d85679`; its existing full,
  race and vet evidence is reused, not represented as a new run.
- Used the handed-off candidate binary read-only at
  `/home/sean/Worktrees/itsyweb/t3-c3b38c40/system/bin/itsy`.
  Verified SHA256:
  `a138a909be0ecda20f7c50d7749a5c9e1fcd0012d02361281f83d91d1e7154c2`.
- Candidate `itsy status`: valid configuration/templates/content, 18 pages,
  31 assets, zero Blog posts. No deployment mapping for this task branch.
- Exact title, description and sharing metadata on all 16 public content routes;
  15 guides and 87 internal links/fragments checked against real HTTP output.
- Retired editing/media guide routes return 404. Public output contains no saved
  page directives, admin CSRF meta or current-admin template references.
- Homepage HTML/Publish tab switching and keyboard Home behavior work on the
  candidate runtime. Desktop layout and three optional-feature columns inspected.
- Phone guide menu, guide-title navigation and selected ice state work. Scrolling
  selects the correct chapter; anchors clear the sticky header. Home and Docs
  fit the minimum 320px viewport; desktop and 390px phone layouts inspected.
- Both homepage and guides scripts pass `node --check`; source diff and design
  detector pass. Existing Blog/Contact/runtime acceptance from the handoff is
  reused because this integration makes no framework changes.

## Rollout remains separate

This is a local integration, not a release or deployment receipt. Protocol 30 is
unreleased while the candidate still reports 0.21.0; the public 0.21.0 pair uses
protocol 29. Preserve the historical native Mac receipt as historical only.
Current native Mac access/acceptance, supported Ubuntu/systemd/public-TLS and
real-email acceptance remain unverified. Old page-CMS storage requires inventory,
backups and owner-approved scoped conversion before an upgrade. Preserve Blog,
media, photos and developer data; do not discard unsupported fields to pass checks.
No pushes, releases, server upgrades, live inventories, live content conversion or
deployments were performed. The original handoff checkout and canonical shared
skills were not changed.
