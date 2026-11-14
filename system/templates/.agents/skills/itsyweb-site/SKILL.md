---
name: itsyweb-site
description: Build or maintain ItsyWeb websites using the pinned shared authoring workflow and this version's site requirements.
---

# Build an ItsyWeb website

Read the site's AGENTS.md and `.agents/site-requirements.md`, then read
[the pinned shared authoring workflow](../fleacms-site/SKILL.md). This local entry
point adapts its legacy Flea terminology without modifying the canonical shared
skill files or their verification manifest.

Use `itsy` commands. New projects use `itsy.config.json`, `itsy/` and `<!-- itsy`
page settings. Existing `.flea.config.json`, `flea/` and `<!-- flea` remain supported;
preserve the selected filenames and content instead of converting them incidentally.
The only working old CLI command is `flea update`. New-layout deployment requires
server protocol 29 and an explicit `itsy update-server` before deployment to an
older server. Keep custom layouts, content, media, authentication and form contracts.

Use the current paths and supported behavior in the local site requirements
where the upstream workflow uses older naming. Shared workflow guidance does not
authorize deployment, external changes, or superseded template formats.
