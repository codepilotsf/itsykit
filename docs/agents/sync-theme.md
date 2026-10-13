# Snip synchronization and developer-managed styling

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).

`flea pull` retrieves live snips, metadata and referenced media from the current
branch's saved SSH destination. Never pull CSS, configuration, credentials,
sessions or source templates. Refuse uncommitted selected-file changes, preserve
local-only IDs and never auto-commit. First-pull conflicts require explicit choices.
Baselines live in private Git metadata under flea-sync, keyed by exact destination
and local project. Advance baselines only after successful local writes; report
partial disk failures honestly.

Normal deployment preserves live snips/metadata, seeding new IDs. `--with-snips`
explicitly includes reviewed local content changes. Existing destinations require
a pull baseline; new sites seed automatically. Compare snips and metadata fields
three ways; multisnip arrays compare as a whole. Noninteractive conflicts stop.
Collect choices before activation and verify the expected snapshot under deploy.lock.
Never hold that lock while prompting. State export is SSH-only and bounded.

`site/css/theme.css`, Prewind and theme tokens remain developer-managed files.
Normal deployment publishes committed CSS, without a CSS pull baseline. The host
has no theme overlay map and no browser Theme editor. Backups include release CSS;
rollback restores the selected release's CSS while keeping current live snips.
Content saves retain the immutable source CSS, and EditingVersion includes it.
The removed `--with-theme`, `--theme-only` and `--snips-only` flags have no aliases.

Protocol 25 rejects removed theme host config and sync-plan fields. Before a live
upgrade, back up remote customizations using the previous runtime and explicitly
capture needed styles in Git. Host configuration/toolbar updates require reviewed
maintenance; there is no automatic migration, reset or deletion of retained themes.

Dev, pull and reset share a nonblocking per-project OS lock in the system temporary
directory. Stop dev before pulling; OS lock release on exit handles crashes.
A new dev invocation may request graceful shutdown from another dev instance
through its private Unix socket, then wait for this lock. It never interrupts
pull/reset or bypasses their exclusion. See the runtime guide for control ownership.
This is separate from the in-process content mutex and remote deploy.lock.

## Status comparison

Plain `flea status` automatically reads the current branch's saved deployment.
It compares code separately from editable content. Only scoped pull
baselines for the exact destination establish change direction. Without one,
report differences with unknown direction, never guess that live edits caused them.
Content differences count active snip values, metadata fields, collections and
media-library entries; ignore developer labels, ownership and retained archives.
Status reads baselines without creating directories and never saves them, pulls,
resolves conflicts or publishes. Matching values need no baseline.

## Blog

Read [Blog](../blog.md) for the optional built-in blog, JSON and template contract.
Posts share the snip/media snapshot, atomic save and preservation/synchronization
paths. Drafts remain outside public routes and template data. The blog admin uses
existing login, exact Origin, CSRF, upload bounds and content-version checks.
