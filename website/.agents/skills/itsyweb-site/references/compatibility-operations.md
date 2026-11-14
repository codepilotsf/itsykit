# Compatibility, development and operations

Use the site's installed docs and command-specific `itsy <command> --help`.
This is site authoring guidance, not permission to install software, change
access, deploy, restore data or upgrade a shared server.

## Naming transition

The CLI is `itsy`. In the renamed binary, every `flea` invocation prints the
rename notice and exits with usage status 2, except exact `flea update`. That
exception verifies the update and installs `itsy` beside the old command, even
when already current; it refuses to replace an unrelated existing `itsy`.
Already released Flea binaries need two `flea update` runs: the first installs
the renamed binary under the old name, the second installs `itsy`. Alternatively
use the current installer at `https://itsyweb.dev/install.sh`.

The installer puts `itsy` in `~/.local/bin` and retains a relative `flea` symlink
for the notice/update exception. `ITSY_INSTALL_DIR` selects another install
directory; existing `FLEA_INSTALL_DIR` remains supported. A narrow read-only
`.flea-update-* version` candidate probe still returns `FleaCMS <version>` for
released updaters. This is not a working `flea version` alias.

New sites use `itsy.config.json`, `itsy/` and `<!-- itsy` layout settings. Existing
`.flea.config.json`, `flea/` and `<!-- flea` settings are supported without
conversion. Select config and directory independently, preferring Itsy when
present; a missing child in `itsy/` does not fall back to `flea/`. Invalid current
paths must fail validation, not fall back. Writes, sessions, media and pull keep
the selected paths. This naming support does not restore retired template syntax,
extensionless config, old Contact config or removed theme features.

Preserve server binary `/usr/local/bin/flea-server`, service/account `flea`,
`/etc/flea`, `/var/lib/flea`, health path `/_flea/health`, protocol identifiers,
cookies and locks. Keep `.flea-install.lock`, development
locks, `flea-sync` baselines and recovery names; do not rename private state.
Canonical release URLs use `https://itsyweb.dev/releases/` and `itsy-*` artifacts;
legacy fleacms.com URLs and `flea-*` artifact aliases remain for released updaters.

The published naming baseline is ItsyWeb 0.21.0 / deployment protocol 29.
The page-CMS removal contract here is checked against source commit `eea4019`
(protocol 30, `FLEA_DEPLOY=30`); it is implemented but unreleased at this handoff.
That candidate still prints 0.21.0. A version string or successful build alone
cannot establish this contract: inspect the source/protocol and matching server.
The public 0.21.0 installer/server uses protocol 29 and does not contain this
removal; running the public installer/update is not a way to obtain the candidate.
Keep current-source guidance distinct from public availability and recheck the
actual release/implementation for later tasks. App protocol remains 2; earlier
feature minimums (host indexing 27, SEO/redirects 28, naming 29) do not relax the
protocol-30 requirement for this contract. A local commit/build authorizes no
release, deployment or shared-server upgrade. `itsy update` updates only
the local CLI/bundled servers; `itsy update-server` explicitly upgrades the saved
branch's server executable after confirmation, briefly affecting every hosted
site. It preserves old sites and does not migrate systemd policy/configuration.
Never substitute setup/reinstallation for an upgrade.

## Local development

`itsy new <directory>` creates a minimal site; `--starter`/`-s` adds demos. Paths
and `.` work, but an existing directory must be completely empty, including
hidden files and `.git`; symlink targets are refused. It normally makes an
initial Git commit on main, respecting hooks/signing. `--skip-git` supports a
parent repository. No remote is created. Do not scaffold over an existing site.

`itsy dev` is loopback-only by default, trying ports from 3000 when occupied.
Explicit `--port` is fixed; `--port 0` asks the OS. Restarting dev replaces the
same site's existing dev server gracefully, so coordinate ownership first.
`--lan`/`-l` explicitly enables trusted-network HTTP with exact Host/port and
private/Tailscale peer checks. Keep normal auth/CSRF; never expose it publicly.
Preview/sender and app config changes require a restart.

## Content preservation and recovery

Authorized `itsy setup` installs the bundled server through native SSH on
supported Ubuntu/systemd hosts; it does not provision a host, DNS or firewall.
Deployment saves branch-specific server/domain/WWW/HTTPS preferences. HTTPS needs
public DNS, inbound TCP 80/443, a contact email and explicit Let's Encrypt terms
acceptance. Use `itsy https` for an already deployed site; preserve other hosts.

Normal `itsy deploy` publishes committed templates, config, assets and CSS,
seeds local-only IDs, and preserves existing live Blog posts and cover photos.
Missing source IDs never delete live records. Preserve `deploy.branches` mappings;
only the current branch is selected, with no main fallback. Deployment needs
committed clean source. In a larger Git repository commit manually; the CLI will
not offer its automatic commit there.

Stop dev before authorized `itsy pull`; selected local files must be committed.
It retrieves Blog posts and referenced cover photos, never CSS, templates, config,
secrets or sessions, and never commits.
Preserve local-only IDs and review/commit results. `itsy deploy --with-content`
explicitly publishes reviewed local content; existing destinations need an exact
destination-scoped pull baseline first. Posts and photos compare as whole records.
Conflicts require explicit choices and unattended conflicts stop. Keep baselines
in private Git metadata, scoped to exact destination/project. Check expected remote
state under the deployment lock; preserve local project locks and recovery guards.
`--with-snips` is removed without an alias. Removed theme flags `--with-theme`,
`--theme-only`, `--snips-only` also have no aliases; CSS travels with code.

`itsy status` validates locally, then automatically checks the current branch's
saved remote without changing baselines. Without that mapping it finishes locally.
No remote flag exists. Content direction needs a matching pull baseline; otherwise
status reports differences of unknown direction. Remote failures exit 1 while
preserving the local summary; ordinary differences exit 0.

Rollback, backup/restore and executable-only server upgrades are implemented.
`itsy rollback --list` lists retained code releases; rollback keeps current live
content and selects the earlier code/CSS. `itsy backup ../site.flea-backup`
writes a new private archive outside the site; the suffix is not a rename mandate.
Backups include code/config/live content/media, excluding deploy settings,
credentials, certificates, sessions, host indexing and custom-app storage.
`itsy restore` replaces live code/content after an expected-snapshot check and
explicit confirmation. Preserve backup identity/digest checks, locking and
rollback/recovery safeguards; retained files are never automatically pruned.
Interrupted server upgrades use `itsy update-server --recover`. Read recovery docs
before use.

`itsy reset` replaces all local files/settings except `.git` after confirmation,
keeping a private sibling recovery copy; it never deploys. Do not use it for
cleanup. Server removal unregisters routing and restarts the shared service but
retains files; redeploying the same name does not recover retained live content.
Release pruning, remote logs commands and automatic retention remain deferred.

Before using this incompatible runtime on an old site, preserve a backup and
require an explicitly scoped source conversion. Old page-CMS fields, named-content
directories and pending extraction journals fail closed without rewriting. Do not
discard content/media to pass validation or perform automatic migration, pruning
or live cleanup. Naming compatibility does not make old page-CMS content compatible.
Keep historical release receipts; rollback retains current Blog content rather
than converting it for an older runtime, so check code/content compatibility.
