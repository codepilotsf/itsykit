# FleaCMS 0.19.0 release verification

Candidate prepared October 24, 2026. Publication and live verification are pending.
This record will be completed from actual release and rollout evidence.

## Scope

The release reconciles host indexing implementation `3b7d21d` with published
0.18.0 and current remote main `0002a84`, preserving empty-directory creation and
the two supported CLI platforms. Unrelated local shared-skill installation work
is excluded. Exact-host `indexing` is server-owned; no website app flag, template
rewrite or body buffering is required. Deployment protocol is 27 and app protocol
remains 2. Older clients/servers reject the mismatch.

## Required verification

Native Linux AMD64 and macOS ARM64 full tests, race checks, vet and bundled builds,
exact-artifact smoke, Mac ad-hoc signature, public installer/update checks, shared
server preservation and live staging header isolation must be recorded before
claiming completion. ARM64 Linux server native execution, notarization, clean-Mac
and real-email delivery remain separate acceptance limits.

## Rollout boundaries

Use the supported shared-server upgrade, then configure only the registered
staging.codepilot.com as noindex. Verify its actual header before removing the
consumer meta fallback in a separate scoped commit. Preserve unrelated domains,
content/media/source bytes, memory/swap policy, the signup URL and Hub billing.
Do not register codepilot.com or change DNS. Retain prior binaries, host config,
site backups and immutable published downloads.
