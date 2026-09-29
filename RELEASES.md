# Publishing FleaCMS downloads

Public downloads live in `site/releases/<version>/`. Each version directory is
immutable: never replace published executable bytes or their SHA256SUMS.
The website repository contains public binaries, not framework source.

Publish a new version's binaries and checksum manifest first. Fetch them over
HTTPS and verify their checksums before advancing `site/releases/latest.txt`.
`site/install.sh` and `flea update` read that pointer. Test actual downloads,
installation, reinstall and update before advertising the command.

Required platforms: Apple Silicon macOS (`darwin-arm64`) and x86-64 Linux
(`linux-amd64`). Do not list additional platforms without native checks.
The Mac executable is ad-hoc signed, not notarized. This is an early test release;
clean-machine testing remains pending.

## 0.4.1 release preparation

Patch release from private framework commit `10dc360`, immutable tag `v0.4.1`.
Shortens `flea new` output to the local development instructions. The preceding
0.4.0 version was a framework development snapshot, not a public download.
Source archive SHA-256:
`9bd091f3babd20891032d162cb4e56c1d8fd7cef195a924ea05fbec38877a8fa`.
Mac tests, race tests, vet, bundle extraction and new/status smoke checks passed.
Native Linux tests (with DAC bypass capabilities dropped), vet, bundle extraction
and new/status smoke checks passed. Linux race tests were not run: this host has
no C compiler. Mac code signature verification passed; notarization remains pending.
Public publication and download checks are recorded below when complete.

## 0.3.0 build provenance

Built from an isolated snapshot of the current private framework working tree,
including the 0.3.0 version update and corrected contact-page and CLI tests. The framework
worktree was not bulk-staged, committed, or pushed.

Private source archive SHA-256:
`01f86b166483c46ff6effa6a2b6ed64061e57cae437ac8e3f9209c3f3c86bbf4`.

The archive is retained locally under ignored `release-work/0.3.0/`, outside
`site/`. Never move it into a public directory. The temporary Linux build copy
was removed after verification, together with the temporary toolchain and Linux
test installation. The local private source archive remains available.

The Mac release pipeline passed tests, race tests, vet, bundle verification,
site creation and validation. Native Linux tests, vet, bundle verification,
site creation and validation also passed. Linux tests ran with DAC bypass
capabilities dropped so root could not mask permission failures; compilation
then ran with ordinary access to generated bundle files. No Linux race run was
performed because this temporary host does not have a C compiler.

Both CLIs contain both Linux server bundles. Only the CLI executables are
published.

## Public verification, September 29, 2026

- Both anonymously downloaded HTTPS executables match the public SHA256SUMS.
- The real public installer passed fresh-directory installation, reinstall,
  PATH selection, version, new-site creation, status and dev HTTP checks on
  Apple Silicon macOS and native x86-64 Ubuntu Linux.
- The Linux installer checks ran as unprivileged UID 65534, not root.
- Install directories containing spaces work on both platforms.
- Both installed CLIs report already-current correctly. A private
  `0.3.0-rc.1` build on each platform successfully updated through the real public
  channel to the exact SHA-256 of the released 0.3.0 executable.
- Installer/updater automated tests passed checksum mismatch, interrupted
  download, invalid executable/version and preservation-of-existing-file cases.
- These were isolated directories on existing machines, not clean OS installs.
  The public pages explicitly invite testing and disclose the Mac notarization
  and platform limits.

The installation pages were deployed from website commit `2205161`, active
release `20260929-202718-8ebaae1f4db520dfa3f53919`. The homepage and docs show the
command; live install.sh matches the reviewed file; latest.txt is 0.3.0. Public
links and anchors resolve. The private source URL returns 404. Bookworm retains
its original release/content/theme and returns HTTP 200.
