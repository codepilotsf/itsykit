# Installed shared skill source

Source repository: `git@gitlab.codepilot.com:sean-schertell/master-skills.git`

Pinned source commit: `05e8168afe6497a565d92dc912e1a47e3db40ef6`.
Update mode: local review worktrees.
Selected complete directories: `core`, `efficient-collaboration`, `task-delivery`, `deepseek-orchestrator`, `copywriting`, `design`, `prewind`, `itsyweb-site`, `art-direct`, `photoreal-image`.
Shared files originate from the pinned canonical source, including licenses and
provenance. The owner-requested page-editing removal deliberately adapts the
installed `itsyweb-site` skill and its references to this website’s 0.22.0 framework
contract (HTML metadata, static assets and Blog-only content sync). Other shared
skills remain at the pinned source. The manifest records the installed files.
Keep further unrelated project adaptations outside these shared directories.

Verify the per-file SHA-256 manifest from `.agents/` with
`sha256sum -c skills-source.sha256`.

Updates are prepared on a local branch for review. The updater does not push,
commit, merge, or alter personal instructions.
