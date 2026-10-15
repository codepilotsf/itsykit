# Working on this Flea CMS site

At task start, read [.agents/skills/core/SKILL.md](.agents/skills/core/SKILL.md).
Its conversation preferences are defaults; personal preferences and explicit
user requests override them. Keep personal instructions/configuration intact.

Before site edits, read [site requirements](.agents/site-requirements.md),
[README.md](README.md) and [Flea authoring](.agents/skills/fleacms-site/SKILL.md).
Load [Prewind](.agents/skills/prewind/SKILL.md) for styling and
[design](.agents/skills/design/SKILL.md) for visual/UX work.
For focused Blog, Contact or site import work, read the matching skill in
`flea/skills/` without restarting a full-site consultation.

- Preserve content, custom layouts, assets, authentication and CSRF contracts.
  Keep secrets and agent guidance out of public assets and deployment.
- Ask before adding dependencies, backward compatibility or changing access or
  deployment settings. Do not commit, deploy or publish unless authorized.
- Default implementation tasks to separate worktrees using the
  [delivery skill](.agents/skills/task-delivery/SKILL.md); coordinate runtime
  resources. Use one agent unless delegation is authorized.

Follow the site's requirements and run relevant checks before claiming completion.
