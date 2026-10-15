# Provenance

This directory is an unmodified install of the upstream **Impeccable** skill,
vendored as a project skill. OpenCode and Codex discover it from `.agents/skills/`;
Claude Code reads `.claude/skills/`, which this repo bridges with a tracked symlink to
`../.agents/skills` so the skill body stays a single copy.

- Upstream: https://github.com/pbakaus/impeccable
- Pinned revision: `d631a8827f99414d2b6daba4ef08b7f8701751d7` (release 4.5.2, 2026-10-09)
- Installed with the Codex `skill-installer` from the repo path
  `.agents/skills/impeccable` into `.agents/skills/impeccable`.
- License: Apache-2.0. See [LICENSE](LICENSE) and [NOTICE.md](NOTICE.md), both copied
  from the pinned revision.

The skill body (`SKILL.md`, `reference/`, `agents/`, `scripts/`) is upstream and is
loaded only on demand. No provider-native hooks were installed; no product interview is
run automatically. To refresh, reinstall the same repo path at a newer pinned revision
and update this file.

## Engine dependency

`scripts/impeccable` is a launcher, not the engine. On first use it downloads the
platform engine binary (`engine-v0.1.14` at this pin) from the GitHub release into a user
cache (`$IMPECCABLE_HOME`, or `~/.impeccable/bin/<version>`), verifies it against the
`.sha256` sidecar, and fails closed on mismatch. First run therefore needs network egress
and a writable user cache outside the repository; later runs reuse the cache. A sandbox
without egress must preinstall the binary on `PATH` or point `IMPECCABLE_BIN` at it.

Verified at this pin by running only the `engine-probe` handshake with a private
`IMPECCABLE_HOME` under `/tmp`: it downloaded the engine and returned
`impeccable-engine 0.1.14`. No `context`, `init`, hooks, or design interview were run, and
no cache was written inside the repository.
