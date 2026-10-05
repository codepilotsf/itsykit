# Flea handoff: October 5, 2026

Start here in a fresh conversation. Read AGENTS.md and the relevant subsystem
guides before acting. This handoff records completed work, not authorization for
new production changes. Do not repeat the completed rollout.

## Working arrangement

Use the global skill `/Users/sean/.codex/skills/deepseek-orchestrator/SKILL.md`.
Sean wants low Codex/Astra usage: DeepSeek owns investigation, implementation,
tests, routine fixes and docs. Use a separate read-only DeepSeek reviewer for
substantive changes, sequentially, then batch fixes. Codex handles scope,
consequential decisions and risk-focused final review. No native Codex subagents.
The skill includes a runner that captures verbose logs outside the repository.
Read concise reports first; do not duplicate exploration or repeat unchanged
checks. Keep mandatory progress updates brief and useful.

Skill runner syntax and mock success/failure/missing-report tests passed.
Metadata checked with Ruby YAML; standard Python validator could not run because
PyYAML is missing. No workers or rollout commands were left running.

## Repositories and commits

All three trees were clean immediately before creating this handoff.

- Framework `/Users/sean/Projects/flea`: `3447cb7` implements server site
  administration and sets source version 0.9.0-rc.1; `7b3dc6d` records rollout.
- Atlas `/Users/sean/Projects/atlas`: `e9dd946` interim Contact disable;
  `b2f768a` final Contact app conversion, deployed.
- Website `/Users/sean/fleacms-website`: `7ef4e6e` Contact pages, previously
  pushed and now deployed; `9eea991` rollout handoff/docs only, not deployed.

New framework/Atlas/website-doc commits have not been pushed. This handoff is
also uncommitted. Inspect status and remotes before any future Git action.

## Completed live rollout

Verified by the rollout worker October 5; refresh read-only state before a new
operation. Host `root@fleacms.com` runs **0.9.0-rc.1 / protocol 23**. Public
downloads, latest.txt and Sean's installed CLI remain **0.8.1**. No RC artifacts
or tag were published. Use framework `bin/flea` for current server commands;
`/Users/sean/.local/bin/flea` embeds the old protocol-22 server.

Sequence completed: interim Atlas deployment, authorized one-time Bookworm
routing removal, supported update-server, final Atlas deployment, website deploy.
Bookworm's actual registration was `v1-acceptance-main`, not `bookworm-main`.
It is now unregistered; all 20 releases, content, media and certificates retained.
Do not rerun the temporary protocol-22 maintenance script or unregister Atlas.

Active sites: Atlas, FleaCMS, Kent and Naves. All healthy; www redirects to apex.
Atlas and FleaCMS Contact use `/contact` and `/contact/thank-you`. Live CSRF 403,
invalid-email 422 with feedback, and honeypot 303 checks passed without sending
email. Real delivery was not tested in this rollout. Two Contact children run.
Content/themes matched backups, except Atlas's intentional thank-you route move
with snips preserved. Kent/Naves were not redeployed. No recovery was needed;
no pending journals remained. Memory/swap policy unchanged (320MiB Go target).

Activation IDs and build hash are recorded in `docs/agents/deployment-runbook.md` and
website `RELEASES.md`/`handoff.md`. The full worker evidence is
`/tmp/flea-rc-rollout.LKpchI/live-report.md` (temporary, may disappear).

Five verified private site backups: `/Users/sean/fleacms-backups`.
Remote pre-removal config:
`/var/backups/flea/server.json.before-bookworm-removal-20261005-162928`.
Never publish backups or print private configuration/credentials.

## What remains

- No new implementation task is selected. Ask Sean what to tackle next.
- Push new commits only with authorization. Public 0.9.0 release is separate.
- `flea server sites` passed live inventory. The new `server remove` command
  still needs disposable-host live QA; Bookworm used the one-time old-protocol
  script, not that command. Removal retains files but does not automatically
  reconnect retained content on redeploy.
- Native Linux full suite, clean-Mac acceptance and real Contact delivery for
  this rollout remain open. See next-steps.md and preserve checklist answers.

Framework full tests/vet/race/build passed during implementation; subsequent
review fixes passed full tests/vet/build and affected-package race checks.
RC metadata checks/build and both site validations passed. Do not repeat the
entire suite just to resume this handoff.
