# Site app skills verification

October 13, 2026. Framework source `1feae24`, branch `feat/site-app-skills`,
worktree `/home/sean/Projects/flea-site-app-skills`, based on Blog `975b26c`.
[Draft MR 7](https://gitlab.codepilot.com/sean-schertell/flea/-/merge_requests/7)
is stacked on `feat/minimal-blog`. One writer; original dirty main was preserved.

New sites include `add-blog` and `add-contact-form`, with intent routing in
AGENTS, app setup, template values, post/form contracts, safe publishing and
verification guidance. Existing sites adopt the instructions explicitly; no
runtime keyword listener or dependency was added. Patch 0.15.1 leaves protocol 25.

## Verification

- Native Linux ARM64 Go 1.26.9: full `go test ./...` with required real server/app
  bundles and `go vet ./...` passed. The standard `scripts/build.sh` passed,
  refreshing both Linux servers and the native CLI. No source-only build was used.
- Both new skills and the revised site-building skill passed the skill validator;
  reference links resolve and diff checks passed.
- Fresh clean/starter sites contain byte-identical skills and finish with successful
  `flea status`. Documented Blog/Contact config, templates and draft JSON also
  validated in disposable sites.
- No concurrency implementation changed, so race/browser editor suites were not
  repeated for this guidance update. Mac/AMD64 native acceptance is not established.

## Authorized private deployment

The local `/home/sean/.local/bin/flea` and shared `root@fleacms.com` server report
0.15.1. Local CLI SHA256:
`eebfeb5664d8095a25d617905107a34d15590649516306be56fdf8d815274575`.
Server Linux AMD64 SHA256 matches the bundled executable:
`452269abb9916c64660e6e4bd7b24afe9564ffbad754fd2502ba02a136ba0e1f`.
Standard `flea update-server` completed its verified executable-only swap.
All seven site backups and the old server executable remain in the private host
folder `/var/backups/flea-app-skills-0151-20261013`. Previous local CLI and website
backup remain under `/tmp/flea-app-skills-rollout`.

Config and all content/media/CSS fingerprints were identical after the server
update. The website source `5380c92` was then normally deployed through a clean,
isolated main checkout, without a source merge or explicit content sync.
Activation: `20261013-201235-eb349a3b66e33aaf2fadb520`.

Live `/docs#agents` includes both skills, request examples, literal template tokens,
valid Blog/Contact links and existing-project adoption instructions. Browser
checks at 1280px and 375px found no page/section overflow. Both skill HTTP URLs
return 404. All eleven HTTPS domains return 200; saved content/media/CSS hashes
and site/domain counts are preserved. Service is active, recovery journals are
absent and the deployment-window error journal is clear. Memory/swap policy is
unchanged. No real email was sent.

Public downloads remain 0.13.0. No public artifacts, tags or source merges were
performed. Website verification records are in its RELEASES.md and handoff.md.
