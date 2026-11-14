# Installed shared skill source

Source repository: `git@gitlab.codepilot.com:sean-schertell/master-skills.git`

Pinned source commit: `b898f3afaa0e01651f9fcfa78d03869af3a3c3f9`.
Update mode: local review worktrees.
Selected complete directories: `core`, `efficient-collaboration`, `task-delivery`, `deepseek-orchestrator`, `design`, `prewind`, `itsyweb-site`, `copywriting`, `art-direct`, `photoreal-image`.
Shared files originate from the pinned canonical source, including licenses and
provenance. The installed `itsyweb-site` entrypoint and operations reference have
a narrow release-version adaptation for 0.22.0/protocol 30; other shared files
match the pin. The manifest records the installed bytes.
Keep project adaptations outside these shared directories.

Verify the per-file SHA-256 manifest from `.agents/` with
`sha256sum -c skills-source.sha256`.

Updates are prepared on a local branch for review. The updater does not push,
commit, merge, or alter personal instructions.
