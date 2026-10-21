# FleaCMS 0.18.0 release verification

Prepared October 21, 2026. Publication and installer acceptance are pending.

## Scope and provenance

- Base: remote `main` `7e137e445a0b9f144559bac0aff7fd83b151ed2c`.
- Adds `flea new .` and new/existing empty destination directory paths.
- Nonempty directories, including hidden files and `.git`, are refused.
- Existing display-name, deployment-name and Git defaults are preserved.
- Carries the global-only Impeccable and starter agent-guidance updates on main.
- Version: `0.18.0`; deployment protocol remains `FLEA_DEPLOY=26`, app protocol 2.
- Release worktree: `/home/sean/Worktrees/flea-release-0180`.
- Website worktree: `/home/sean/Worktrees/flea-website-release-0180`.
- Private artifact/evidence archive: `/home/sean/fleacms-releases/0.18.0`.

## Platform decision

Owner decision: discontinue the Linux ARM64 CLI; retain the cross-compiled ARM64
Linux server inside both supported CLIs. Native ARM64 server acceptance is a
separate check and does not block publication. No ARM64 Linux machine is needed
to build the server. Older published ARM64 CLI artifacts must remain available.

Supported native CLI checks: Apple Silicon macOS and x86-64 Linux. The installer,
updater, release CI and staging must reject/omit Linux ARM64 CLI releases.

## Verification and publication

Final source, platform scope, build results, artifact hashes, publication and
website activation receipts will be recorded here after they are observed.
Published versions and historical release records must remain immutable.
