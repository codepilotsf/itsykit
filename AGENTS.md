# Working on ItsyKit

This CP Git repository has two sibling projects:

- `system/`: Go CLI/server, canonical starter, build scripts and framework docs.
  Read [system/AGENTS.md](system/AGENTS.md) before framework work.
- `website/`: current public website and documentation, replacing the former scaffold.
  Read [website/AGENTS.md](website/AGENTS.md) before website work.

Run Go commands and framework scripts from `system/`; run `itsy` website
commands from `website/`. There is one Git repository at this parent directory.
Commit changes explicitly before website deployment; the CLI does not offer
an automatic commit for a site nested in a larger repository.

Inspect `git status --short` first and preserve unrelated work. Keep both source
histories and published release tags intact. Historical tags keep their original
layouts. Do not copy framework starter files over the custom public website.
Do not deploy or publish releases without explicit authorization.

Use one agent unless the user authorizes delegation. Shared working practices
live in [core](system/.agents/skills/core/SKILL.md); delivery guidance lives in
[task-delivery](system/.agents/skills/task-delivery/SKILL.md). Global personal
instructions and project-specific security and verification requirements apply.
