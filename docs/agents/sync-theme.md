# Synchronization and theme editing

The Theme panel slides fully in and out from the right and tracks the toolbar's
measured bottom during edit-row collapse, resizing and scrolling. Opening Theme
collapses the edit row; closing restores the preferred editing mode.

Implementation requirements linked from [AGENTS.md](../../AGENTS.md).
Read this when the task touches this subsystem. Paths in prose and code are
relative to the repository root unless stated otherwise.

`flea pull` retrieves snips and theme from the current branch's saved SSH
destination; `--snips-only` and `--theme-only` are mutually exclusive selectors.
Never pull config, secrets, sessions, or source templates. Refuse uncommitted
changes in selected files and preserve local-only IDs. Do not auto-commit.
First pulls resolve differing existing values explicitly rather than assuming
local seed content is disposable. Stored baselines live in Git metadata under
flea-sync, keyed by exact destination and local project, never in site archives.
Separate baseline presence for snips and theme. Advance baselines only after
local writes succeed. Report partial local disk failures clearly.

Normal deployment preserves live snips/metadata and theme, seeding new IDs.
`--with-snips` and `--with-theme` explicitly include local changes. Existing
deployments require the corresponding pull baseline; new sites seed automatically.
Use three-way comparisons per snip and metadata field; theme merges as a whole
CSS file. Conflicts have no default: keep live, use local, or cancel. Noninteractive
conflicts stop. Collect all decisions before activation, then compare the expected
snapshot under deploy.lock. Never hold that lock while prompting. SSH-only state
export is bounded and excludes configuration/runtime credentials. Protocol 23
requires state export, reviewed sync plans, and managed themes.

Managed live CSS lives beside content as <release-id>.json.theme.css; host.Themes
maps aliases to the same file. The theme path must accompany its content snapshot.
Release files remain administrator-owned; browser edits write only managed CSS.
Candidate config and recovery switch code, snips, and theme together. Theme CSS
participates in EditingVersion; snip saves must retain the live theme overlay.

Every owner may use Theme beside Edit. The fixed right-hand panel shares topbar
control dimensions and states, with independent styling and scrollable groups.
Theme suspends content editing: retain its first-row button and disable the
Editing switch while open. Hide page actions while Theme is open. Use a header X
and a fixed panel footer with a right-aligned Save button, disabled until changes
exist. X and Escape confirm before discarding dirty previews. Closing restores the
preferred editing mode. Animate entry and exit over 220ms, immediately with reduced
motion; cancel stale exit animations before reopening.
Preview is browser-only; Save/Cancel are explicit. Guard dirty mode switches and
retain drafts after errors.
Use plain labels (Theme, Palette), inset select chevrons, and animated accordion
height transitions in both directions, with immediate reduced-motion toggles.
Open only one sibling accordion section at a time. Put brand, neutral and semantic
color sections first, with all sections closed on entry.
Font Color is a token dropdown: neutrals before brands,
no semantic colors or arbitrary values. Preserve existing values until edited;
the starter defaults to var(--black). Validate the token on save.
Group --font-color and --bg-color under Page colors. Both select neutral/brand
tokens. The starter body uses --bg-color, default var(--lighter).
Font families are plain text inputs. Google Fonts load in a debounced
private preview; Cancel removes the preview stylesheet. Save regenerates only
the managed Google Fonts import block, before CSS rules and after any @charset.
Preserve developer imports. System families need no imports. Accept family stacks,
never pasted URLs or declarations. Default Google faces load with display=swap;
bold/italic may be synthesized. No API key, catalog API or server network call.
Use accessible inline info disclosures for
ambiguous settings and inset input/select focus outlines to avoid clipping.
Text/spacing scale controls replace individual clamp fields when the complete
Prewind scale is present. Keep six settings in each Fluid Text/Space Scale CSS
comment. Verify declarations against settings; customized scales require explicit
replacement. Browser JS previews all generated values immediately; Go regenerates
and validates them on save, preserving unrelated CSS. Cancel restores original
inline values/priorities. No Node/npm execution. Test browser/Go math agreement,
including negative intercepts and decreasing endpoints.
GET/POST /admin/theme require session authentication;
POST also requires exact Origin and CSRF. Only supported existing :root variables
are editable; preserve CSS source bytes outside changed values. Reject URL,
escape, declaration, selector, and unsupported-function injection. No raw CSS
editor. The font-import writer generates only fixed-host Google Fonts URLs from
validated family names; it never accepts arbitrary stylesheet URLs.

Dev, pull and reset share a nonblocking per-project OS lock in the system temporary
directory. Stop dev before pulling; OS lock release on exit handles crashes.
A new dev invocation may request graceful shutdown from another dev instance
through its private Unix socket, then wait for this lock. It never interrupts
pull/reset or bypasses their exclusion. See the runtime guide for control ownership.
This is separate from the in-process content mutex and remote deploy.lock.

## Status comparison

Plain `flea status` automatically reads the current branch's saved deployment.
It compares code separately from editable content and theme. Only scoped pull
baselines for the exact destination establish change direction. Without one,
report differences with unknown direction, never guess that live edits caused them.
Content differences count active snip values, metadata fields, collections and
media-library entries; ignore developer labels, ownership and retained archives.
Status reads baselines without creating directories and never saves them, pulls,
resolves conflicts or publishes. Matching values need no baseline.
