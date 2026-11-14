# ItsyWeb 0.22.0 verification

## Scope and source

The owner authorized framework/website integration, release publication, shared
server upgrade and deployment. After inventory showed all 22 hosted sites used
removed page-CMS storage, the owner explicitly approved converting all 22 while
preserving their content before upgrading/deploying. Protocol 30 is intentionally
incompatible with page-CMS content; normal loading and upgrades never migrate it.
App protocol remains 2. Blog retains its existing `itsy/snips.json` (or selected
legacy `flea/snips.json`) filename with only `blog` and `media` top-level keys.

Framework removal and metadata implementation: `d601d2a`. Integration and shared
agent guidance: `9d85679`. Framework full tests, race tests, vet, rebuilt CLI and
HTTP/Blog fixture checks passed on that source. Final native release, public
artifact and live rollout evidence will be recorded below as each step completes;
this preparation record alone does not establish publication or deployment.

## Preservation preparation

The 22 supported portable backups were created sequentially using the installed
0.21.0 server. Each manifest size/SHA-256 was verified, safely extracted and loaded
with the original framework. Private inventory, backups, conversion tools and
logs are under `/home/sean/itsyweb-releases/0.22.0/`, outside source/public assets.
Original immutable releases and content snapshots remain retained. Runtime state,
sessions, custom-app storage, host indexing, redirects, TLS and branch ownership
must remain unchanged by conversion. Website redesign deployment is owned by its
separate website thread after the runtime handoff.

## Acceptance limits

Native ARM64 Linux server acceptance, clean-machine Mac acceptance, notarization
and real outbound email delivery are not claimed. The ARM64 Linux server is
cross-built; both supported CLIs require native release checks. Mac artifacts
are ad-hoc signed. Historical releases, tags and their receipts remain immutable.
