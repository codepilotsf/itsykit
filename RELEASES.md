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

## 0.6.1 preparation, October 1, 2026

Framework commit `89e3472`, immutable tag `v0.6.1`. Starting dev again replaces
the same site's server gracefully, including LAN and port changes. Pull/reset
remain exclusive, and crash recovery removes only stale control sockets.

Mac tests, vet, race checks, bundled builds and replacement checks passed. The
cross-compiled Linux AMD64 artifact passed native unprivileged checks for site
creation/status, same-port dev replacement, LAN mode and crash recovery. A full
native Linux build/test run and clean-machine Mac/notarization remain unverified.
Both CLIs embed both rebuilt Linux servers. Artifacts live under
`site/releases/0.6.1/`; publication and final public installer checks are underway.

## 0.6.0 published October 1, 2026

Framework commit `04c9b3f`, tag `v0.6.0`. User-facing copy now says "admin"; the first `flea dev`
explains the optional admin email, points to `flea admin` when skipped, and asks
only once; the site-in-use message for `dev`, `pull` and `reset` is clearer.

The framework passed tests, race checks, vet, Chrome sign-in/editing regressions,
bundled build and fresh clean/starter site validation on Apple Silicon macOS.
The Mac CLI is ad-hoc signed, not notarized. The Linux AMD64 CLI is cross-compiled
with `CGO_ENABLED=0`; the actual public download passed version and site validation
on native x86-64 Linux as UID 65534. Site creation also passed with identical staged
bytes. This was not a native Linux build or a full Linux test-suite run.
Both CLIs embed the rebuilt 0.6.0 Linux servers. Published artifacts live under
`site/releases/0.6.0/`, with SHA256SUMS. Public `latest.txt`, installation links
and website copy all target `0.6.0`.

The shared server was updated from 0.2.0 to 0.6.0 with site data/configuration
preserved. Download-only release `20261002-003522-5f12907cc538fdfc0ef629ed` published
the artifacts before the update channel advanced.
Both public binaries match SHA256SUMS; the downloaded Mac binary runs and passes
ad-hoc signature verification. Final active website release:
`20261002-004416-449f529af3c1cafa5f4d8c7f`, from commit `3ebad34` on main.

The 1 GB host exhausted memory during activation. Moving the completed upload
from RAM-backed `/tmp` to `/var/tmp/flea-deploy.cqbA28hiz3` let the download-only
deployment finish. The final website activation then lost its deployment process
to the OOM killer. Sean then explicitly approved a persistent Go memory target:
`/etc/systemd/system/flea.service.d/60-memory.conf` sets `GOMEMLIMIT=320MiB`.
After daemon reload and restart, website/content/channel activation succeeded.
No security settings, site data or credentials were changed by this override.

Fresh public installs and real 0.5.0-to-0.6.0 updates passed on Apple Silicon Mac
and native Linux AMD64, matching the published hashes. Linux checks ran as UID
65534, including already-current update and site validation. The live homepage
shows the revised copy and 0.6.0; docs link the correct binaries and checksums.
The service is active at 0.6.0; Bookworm returns HTTP 200. Sean's installed Mac
CLI was also updated through the public channel to 0.6.0. Clean-machine Mac
acceptance/notarization and a full native Linux build/test run remain unverified.
Superseded candidates are preserved in the ignored release-work directory, outside
the public tree. Their earlier checks do not verify the rebuilt 0.6.0 artifacts.

## 0.5.0 published September 29, 2026

Private framework commit `0994c1f`, immutable tag `v0.5.0`. Adds full opt-in LAN
development (`--lan`/`-l`), localhost startup links and in-place page saves.
Source archive SHA-256:
`6711e8ffc05500eecb555f4c83f05000a48a7b6dc356e0f5d40352c18c33dff8`.

Mac tests, race tests, vet, bundle extraction, native build, ad-hoc signature
verification and starter smoke checks passed. The tagged source passed a real
Chrome sign-in/edit/save test on a private-network HTTP origin.
Native Linux tests, vet, bundled build and new/status smoke checks passed as
UID 65534. The 1 GB host killed a parallel compiler for memory pressure; a rerun
with `GOMAXPROCS=1 GOFLAGS=-p=1` passed. Linux race tests remain unavailable because
the temporary build host has no C compiler. No system toolchain was installed.

Both anonymous HTTPS downloads matched the staged checksum manifest before
advancing latest.txt. Asset-only release: `20260930-005958-5fbeba55cfd5fbaf51a0efe0`.
The initial activation hit host memory pressure and automatically rolled back.
A stale, inactive 188 MB Go build directory was moved from RAM-backed `/tmp`
to `/var/tmp/flea-go-build-archive-1567161428`, preserving it for recovery; the
retry succeeded. No server executable, access settings or Bookworm data changed.
Website commit `f2d3c60`, active release
`20260930-010405-485949623ee87304b3304819`, publishes the update pointer and docs.
Actual public fresh installs, reinstalls, PATH selection, new/status, local and
LAN dev, already-current updates and 0.4.1-to-0.5.0 updates passed on Apple
Silicon macOS and native x86-64 Linux. Linux checks ran as UID 65534. Both
installations used paths containing spaces and matched the published checksums.
Live homepage, LAN docs, installer bytes and private-source 404 were verified.
FleaCMS and Bookworm returned HTTP 200. The temporary Linux source, toolchain,
cache and test installation were removed; the local private archive remains.
Framework main and the annotated tag both resolve remotely to `0994c1f`.
Mac notarization and independent clean-machine/device testing remain pending.

## 0.4.1 published September 29, 2026

Patch release from private framework commit `10dc360`, immutable tag `v0.4.1`.
Shortens `flea new` output to the local development instructions. The preceding
0.4.0 version was a framework development snapshot, not a public download.
Source archive SHA-256:
`9bd091f3babd20891032d162cb4e56c1d8fd7cef195a924ea05fbec38877a8fa`.
Mac tests, race tests, vet, bundle extraction and new/status smoke checks passed.
Native Linux tests (with DAC bypass capabilities dropped), vet, bundle extraction
and new/status smoke checks passed. Linux race tests were not run: this host has
no C compiler. Mac code signature verification passed; notarization remains pending.
Both anonymous HTTPS downloads matched the staged SHA256SUMS before latest.txt
was advanced. Website commit `146fdd5`, active release
`20260929-223440-da9b0ae1b1a6a608fb7c5375`, publishes the 0.4.1 pointer and notes.
Actual 0.3.0-to-0.4.1 updates and fresh installs passed on Apple Silicon macOS and
unprivileged x86-64 Linux. Both verified the shortened new-command output and
site validation. The existing 0.3.0 assets remain unchanged. Bookworm still responds.

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
