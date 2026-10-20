# Working on this Flea CMS site

Use only the global Impeccable installation discovered by the agent runtime.
Never install, vendor, copy or link Impeccable into this project or its templates.
Run its global launcher with the project as cwd. Keep project design context and
`.impeccable/` artifacts here. Report a missing global installation.

At task start, read [.agents/skills/core/SKILL.md](.agents/skills/core/SKILL.md).
Its conversation preferences are defaults; personal preferences and explicit
user requests override them. Keep personal instructions/configuration intact.

Before site edits, read [site requirements](.agents/site-requirements.md),
[README.md](README.md) and [Flea authoring](.agents/skills/fleacms-site/SKILL.md).
Load [Prewind](.agents/skills/prewind/SKILL.md) for styling and
[design](.agents/skills/design/SKILL.md) for visual/UX work.
For coding delegation or independent review, read
[orchestration](.agents/skills/deepseek-orchestrator/SKILL.md).
For focused Blog, Contact or site import work, read the matching skill in
`flea/skills/` without restarting a full-site consultation.

- Preserve content, custom layouts, assets, authentication and CSRF contracts.
  Keep secrets and agent guidance out of public assets and deployment.
- Ask before adding dependencies, backward compatibility or changing access or
  deployment settings. Deployment, releases and production mutations require
  separate authorization.
- Default implementation tasks to separate worktrees using the
  [delivery skill](.agents/skills/task-delivery/SKILL.md); coordinate runtime
  resources. Use one agent unless delegation is authorized.

Follow the site's requirements and run relevant checks before claiming completion.

## Owner-authorized delivery and decisions

Routine scoped commits, task-branch pushes, PR/MRs and verified merges are
owner-authorized without per-step approval. Respect required reviews, CI and
branch protections. Honor explicit stop-at-review requests. Preserve unrelated
work and prefer scoped revert commits for requested undo work. Deployment,
releases, destructive Git operations and production mutations still require
separate authorization, including automatic effects of pushes or merges.

Follow core's consequential-direction rule: ask before unresolved product,
creative, architecture or integration choices; proceed within approved direction.
Follow task-delivery's proportionate verification guidance; preserve required checks.
