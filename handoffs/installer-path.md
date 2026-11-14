# Installer PATH setup handoff

## Goal

Make first-time setup as close as possible to one install command followed by
`itsy new my-site`, without a separate `itsy version` step or routine manual
PATH instructions in Getting Started. The owner asked to handle the installer
fix in a separate thread. No installer change has been made in this docs task. Keep implementation in the separate framework thread.

## Confirmed current behavior

- The installer is `website/site/install.sh`, not under `system/scripts/`.
- The documented command is `curl -fsSL https://itsyweb.dev/install.sh | bash`.
- The default destination is `$HOME/.local/bin`. `ITSY_INSTALL_DIR` overrides it;
  the existing `FLEA_INSTALL_DIR` fallback remains supported.
- It checks platform, release version, checksum and the downloaded executable
  before replacing the installed command. It installs `itsy` and the existing
  retired-command link `flea -> itsy`, under an installation lock with cleanup.
- If the destination is already on PATH, it prints `itsy new my-site`.
  Otherwise it only prints a shell-profile `export PATH=...` instruction.
  It does not edit profiles or make PATH available automatically.
- A script piped to Bash runs in a child process. Exporting PATH there, or
  editing a startup file, cannot change the invoking terminal's environment.
  Profile edits help future shells; the current terminal still needs a deliberate
  solution if immediate `itsy new` is an acceptance requirement.

## Decision still needed

No implementation choice has been approved. Agree on the user experience before
changing shell configuration. Concrete options discussed:

1. Keep the installer unchanged and include PATH setup in the same documented
   install line. This works in the invoking terminal, but does not by itself
   configure future terminals.
2. Let the installer configure future terminals automatically, with clear output
   and preserved existing profile contents. Fresh installs may need a new terminal.

If both automatic future setup and immediate same-terminal use are required,
design the install invocation and installer together. Do not claim a piped child
process can export into its parent. Do not silently introduce `sudo`, change the
install destination policy, or overwrite user shell files to avoid this decision.

## Files and instructions

Read root `AGENTS.md` and the affected `website/AGENTS.md` / `system/AGENTS.md`.
Use one agent unless delegation is authorized. Inspect Git status first and
preserve this thread's docs/navigation/metadata work and any concurrent changes.

- `website/site/install.sh`: installer implementation.
- `system/internal/selfupdate/installer_test.go`: offline installer fixtures and
  regression coverage, including a destination containing spaces. Existing
  success assertions expect printed `export PATH=` instructions; update them
  only when the agreed behavior replaces those instructions.
- `system/docs/installation.md`, installer section: the current contract
  explicitly says the installer never edits profiles.
- `website/site/docs.html` and `website/site/docs/installation.html`: onboarding.
  This docs task removes the version check and mandatory separate PATH block,
  retaining a conditional note while the installer fix remains outstanding.
- `website/README.md`, `system/README.md`, `system/templates/README.md`: review
  installation wording for consistency with the final behavior.

## Verification

Use isolated temporary HOME/profile directories. Never test profile mutation on
the real user's startup files. Cover the agreed supported shells and startup
modes, PATH already present versus absent, custom destinations and paths with
spaces, repeat installation without duplicate edits, preserved profile content,
and actionable failure output for unwritable files or unsupported shells.

Keep checksum/version/platform rejection and installation-lock cleanup checks.
Failed downloads must preserve existing binaries and must not leave misleading
shell configuration. Exercise the documented command in a fresh shell and prove
exactly when `itsy new` becomes available; separately check a newly opened shell.

From `system/`, start with `go test ./internal/selfupdate -run TestInstallerOffline`;
also run `bash -n ../website/site/install.sh`. Add focused regression cases for
the agreed PATH behavior and follow the repository's proportional verification
requirements for the final diff. Do not download/install into the real HOME.

## Delivery boundary

Finish a scoped, reviewable local change. Deployment, release publication and
production mutations require separate authorization. Updating this checkout does
not update the installer already served at itsyweb.dev. Report that distinction,
the commit and observed checks. Remove obsolete conditional docs wording only
when the installer behavior makes it accurate to do so.
