# ItsyWeb 0.22.0 verification

## Source and authorization

Immutable tag `v0.22.0` selects `d21da420a2518f1097d79da0f80a66347b221094`,
pushed to CP Git. It integrates the framework removal/comment metadata work
(`d601d2a`), website conversion, approved homepage and shared starter guidance.
Website-only follow-on `b8019cb` preserves the already-deployed installer PATH
improvements without changing the tagged system tree or release artifacts.

The owner explicitly authorized coordination, merge, deployment and release.
After inventory found all 22 hosted sites incompatible, the owner also explicitly
approved converting all 22 while preserving content before upgrading/deploying.
These were one-time operator conversions, not automatic migration support.
Normal loading, deployment and server upgrades reject removed page-CMS formats.
App protocol remains 2; deployment protocol is 30.

## Native builds and artifacts

Native Linux AMD64 (Go 1.26.9) and Mac ARM64 (Go 1.27.1) passed full tests,
race tests, vet and release scripts with `FLEA_REQUIRE_APP_BUNDLE=1`, using a
command-local `umask 0022`. Permission checks were preserved. Scripts ran with
`CI_COMMIT_TAG=v0.22.0`: native Linux `release.sh server amd64` then
`release.sh client amd64`; native Mac `release.sh client arm64`. Both CLIs embed
the exact final Linux servers. Mac final bundling/extraction checks and ad-hoc
signature verification passed after copying those exact final server bytes.
Signed Mac minimal/starter `new`/`status` checks passed.

The initial Mac test archive omitted the sibling website installer, causing
installer/staging fixture failures. Repeating with the full combined source
archive resolved them; no code, checks or permissions were weakened.

| Artifact | SHA-256 |
| --- | --- |
| `itsy-linux-amd64` | `16411183735071699ea5575d0e0990836fc55559d8b5de35f5b114c43a5351d9` |
| `itsy-darwin-arm64` | `76428a72a66dd1ad87441f032b9c6b6d643d08eda76ebb2f1a01096f41e9e3ab` |
| `itsy-server-linux-amd64` | `a77248fca27c14263da31c92263bd6fc55a334830b8d74d4ebb063c4d04103ff` |
| `itsy-server-linux-arm64` | `cad984c049ee9b04a8b0f845d0700107b92ba0eb2268bf82538b4adb4cc6d7e4` |

Immutable `/var/lib/flea/downloads/fleacms/0.22.0/` contains the two native
CLIs, byte-identical regular-file `flea-*` aliases and `SHA256SUMS`, with directories
0755 and files 0644. All four downloads/checksums passed anonymous HTTPS checks
on both itsyweb.dev and fleacms.com before either latest.txt advanced. All 92
previous store files retained identical SHA-256 hashes. Published tags, older
artifacts and their receipts remain unchanged.

## Framework behavior and HTTP checks

The final Linux release CLI passed real HTTP checks on fresh minimal and starter
sites: ordinary Home copy and empty `{}` Blog storage; leading comment titles,
descriptions and layouts; nested routes; missing/empty values; literal braces and
colons; quote/HTML escaping; malformed/duplicate diagnostics; reloads; private
layouts; no extraction, content changes or JSON mtime writes on page loads.
Removed editor endpoints/assets return 404. Real isolated admin login, draft
save, cover upload, publish, unpublish, saved reload and status checks passed.
The full suites protect Blog deployment preservation, sync, authentication,
Origin/CSRF, permissions, snapshots and recovery boundaries.

Ordinary metadata syntax is one leading `<!-- itsy` block with single-line
`title:`, `description:` and optional `layout:` settings; existing `<!-- flea`
spelling remains supported. Empty/missing titles fall back to route titles;
empty/missing descriptions stay empty. Blog posts use their title/excerpt over
comment metadata. Values are data, never recursively compiled as template source.
No saved page metadata overrides or page metadata editor remain.

Blog keeps `itsy/snips.json` (or selected legacy `flea/snips.json`), with only
optional `blog` and `media` top-level fields; `{}` is valid. Post fields are `id`,
`title`, `slug`, `excerpt`, `body`, `status`, optional `published_at` and `cover`.
Cover records/files retain their existing media schema and paths. Normal deploy
preserves live records; `itsy deploy --with-content` explicitly synchronizes Blog
content and `itsy pull` retrieves it using the existing conflict/baseline rules.
`--with-snips` is removed without an alias. Shared package/file names retained for
Blog are intentional; they do not retain page snips, extraction or collections.

The public Blog mount remains configured by `apps.blog.mount`, with private
`_index.html`/`_post.html` templates there. `/admin` provides sign-in and redirects
signed-in admins to `/admin/blog` when Blog is enabled. Private admin layouts keep
`{admin.csrf}`, logout and form controls; Blog saves retain exact-Origin, CSRF,
session and stale-version checks. Public layouts have no toolbar mount, admin
CSRF meta or `app.current_admin` state. Contact/public `{csrf}`, mail previews,
static media, developer CSS/tokens, layouts/partials/JSON and custom apps remain.

## Backups, conversion and activation

The installed 0.21.0 server produced 22 portable backups. Every manifest size and
SHA-256 passed verification; each archive was safely extracted and loaded with
the old framework. Preflight detected a newer Cubitalk release (`9d3b237`), halted
without altering live configuration, and refreshed that backup/conversion before
proceeding. All 22 active content snapshots matched the backed-up records.

Conversions resolve saved sanitized rich text into ordinary HTML, named data into
ordinary JSON, collections into their current ordinary markup and page metadata
into comments. Managed page assets retain their public URLs/bytes. Twenty Sunlux
descriptions had trailing whitespace that single-line settings trim; ordinary
JSON bindings in private layouts preserve their exact original rendered values.
Every Blog/media record is retained exactly. Private per-site Git repositories
record original input, original hosted source identity and committed conversion.
Existing local customer development checkouts were not overwritten; the preserved
converted source is available for their next authoring/deployment handoff.

All 368 public pages across the 22 sites passed old/new DOM and metadata
comparisons. The same 368 routes then passed real HTTPS before/after comparisons
for status, DOM, metadata, redirects and indexing. Comparison normalized only
visitor CSRF token values and intentionally removed collection editor attributes.
There were two retained live Blog posts in CodePilot staging. All their stored
fields, covers and published rendering were preserved. Existing source CSS,
layouts, static assets and custom app packages remain intact.

New immutable release directories, separate content snapshots and linked history
receipts were staged under each existing site. The candidate configuration passed
validation as the actual `flea` service user before activation. Configuration
comparison and content comparison under the shared deployment lock prevented
stale writes. TLS, indexing, redirects, listen/download settings and all branch
ownership remained unchanged. No old release, content snapshot or media was pruned.

A one-time paired executable/configuration journal was rehearsed with an injected
activation failure. The first live attempt rolled back because its immediate
readiness check ran before systemd's asynchronous startup finished; the old pair
was restored. After adding a bounded readiness wait, recovery verified all HTTPS
identities before clearing guards, and retry succeeded with the same artifacts.
Contact cancellation diagnostics occurred during that premature shutdown; the
successful retry and subsequent real Contact page comparisons passed.

Activation revision: `20261114-212120-46e25a4f2760f891feb34b59`.
Installed binary/version/protocol matched the table, 0.22.0 and `FLEA_DEPLOY=30`.
All 34 HTTPS hostname/revision probes passed, `flea.service` was active and no
recovery journals remained. Private runtime file hashes, modes and ownership
matched immediately before/after the successful activation, preserving sessions
and custom-app storage. The 320 MiB Go memory target, swap and service policy were
unchanged; temporary administrator commands used disk-backed private scratch.

The existing website ownership remained:

- `itsy-web`: branch `t3/reframe-flea-html-focus`, itsyweb.dev and itsy.fleacms.com.
- `fleacms-website-main`: branch `main`, fleacms.com and its www redirect.

The website thread completed their sequential normal deployment after this green
runtime/download handoff. It used the exact release CLI and a task-only SSH
wrapper that applies `GOMEMLIMIT=320MiB` and private disk-backed `TMPDIR` to the
original remote command through quoted `bash -c`, including compound commands.
No explicit content sync or permanent host-policy change is needed.

## Public installation and website rollout

Website source `b8019cba5aabce1ca3fc168165c554606b2cdb3b` was deployed normally,
without content sync, to both existing destinations:

- `itsy-web`: `20261114-212920-e298c932a5abcb78ccb8b445`.
- `fleacms-website-main`: `20261114-213138-e8cc85769c3a94943766916b`.

Both public channels select 0.22.0. The website owner verified all 16 retained
metadata routes, canonical/OG/Twitter values, 87 links/fragments and nine asset
hashes on each public domain, including the installer; retired guides return 404.
Alias/WWW redirects and Docs navigation passed. See the website release receipt
for its desktop/mobile design checks and exact deployed-source evidence.

The actual deployed installer is byte-identical to source, SHA-256
`deb56fba7c8c248daa24d637521205f17495c607d59bd105b8c0aebf16740d63`.
Native Linux Bash and native Mac Zsh passed actual public fresh installation in
private homes containing spaces, profile setup, reinstall/profile idempotence,
same-terminal immediate `itsy new`/`status`, fresh interactive and login shell
PATH/new/status checks, retired-command rejection, 0.21.0 update, already-current
update, and the two-step released 0.20.5 `flea update` transition with its prominent
rename notice. The public downloaded Mac binary passed signature verification.
Normal Kirin and Mac CLIs were updated through the supported updater and match
the manifest; their prior executable bytes remain privately retained. Real
home/startup files were not used for installer/PATH fixtures.

After both ordinary website activations, the final sweep verified all 34 HTTPS
health identities against revision `20261114-213138-e8cc85769c3a94943766916b`,
all 22 registrations/branch ownership, the exact two live Blog posts and 14 media
records/files, all 26 private runtime files with identical hashes/modes/ownership,
and all 92 historical download files with identical hashes. TLS/indexing/redirect/
listen/download policies remain identical, and no pending journals remain.

## Evidence and limits

Private backups, per-site converted Git sources, reviewed operator/recovery
scripts, source archives, inventories, artifacts and native/live logs are under
`/home/sean/itsyweb-releases/0.22.0/` on Kirin and corresponding release scratch
on Sean's Mac. The production paired rollback material is private under
`/var/tmp/itsy022-release.4ijctuZn/`; retained incompatible pre-conversion code
requires that paired runtime/configuration recovery, rather than a normal
single-site rollback into the new runtime. Guards reject incompatible restores.
These private materials must not enter public website assets or Git.

Native ARM64 Linux server acceptance, clean-machine Mac acceptance, notarization
and real outbound email delivery are not claimed. ARM64 Linux is cross-built;
Mac is ad-hoc signed. This release is not a claim of completed V1 acceptance.
