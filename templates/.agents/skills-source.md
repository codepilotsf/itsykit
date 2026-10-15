# Installed shared skill source

Source repository: `git@gitlab.codepilot.com:sean-schertell/codepilot-agent-skills.git`

Pinned source commit: `daa0dfd2b51968183664e3992f38ee3e138cfca9`.
Selected complete directories: `core`, `task-delivery`, `design`, `prewind`, `fleacms-site`, `impeccable`.
Installed files match that committed source exactly, including upstream
Impeccable's license and provenance. Project adaptations live outside these
shared skill directories.

The per-file SHA-256 manifest in `skills-source.sha256` records the installed
payload (paths are relative to `.agents/`). Verify locally with
`cd .agents && sha256sum -c skills-source.sha256`.

The source commit was created locally for this task; it has not been pushed.
Publish it through an authorized review process before expecting other developers
to fetch the source revision. No ongoing synchronization is configured. Review
future updates against this revision and reconcile local edits before replacing
copies. Existing project-specific skills outside this selection are not claimed
as copies of the shared source.
