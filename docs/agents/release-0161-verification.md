# FleaCMS 0.16.1 release verification

October 18, 2026. Sean authorized committing and releasing the literal-brace fix,
updating the Mac installation and the shared server. The original checkout at
`/Users/sean/Projects/flea` is based on 0.10.0 with unrelated uncommitted work;
it is preserved. The release worktree is `/Users/sean/Projects/flea-release-0.16.1`,
branch `release/0.16.1`, based on immutable 0.16.0 tag `1c85aae`.
Unreleased modular skill changes on main are excluded from this patch release.

## Scope

A shared lexical scanner distinguishes exact Flea commands and paths from
literal brace groups in compiler and rich/media/multisnip detection. Original
source is rendered with Go contextual escaping. Sanitized content stays data;
no Impeccable code, parameter format or workflow changes are included.

## Verification and rollout

Release validation and installation receipts will be recorded here before handoff.
No public or installed version is claimed updated until verified.
