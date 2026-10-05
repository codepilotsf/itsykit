# Installation and updates

Version 0.9.0 is the current test release for Apple Silicon macOS and x86-64
Linux. The public installer selects the latest stable release from
`https://fleacms.com/releases/latest.txt`; until 0.9.0 publication completes,
that channel may still report the previous version.

## Upgrading to 0.9.0

0.9.0 adds the bundled Contact app and requires deployment protocol 23. Update
the shared server with `flea update-server` before deploying sites that use the
new runtime, and update existing sites from the current starter where their
templates changed. The top-level `contact` settings and the `/_flea/contact`
route are replaced by `apps.contact` with a configured `mount`; there is no alias
or automatic migration, and old settings are rejected rather than converted.
Starter Contact pages use the mount (for example `/contact`) with a
mount-relative `thank-you` target, and their form includes an explicit CSRF
control. Custom layouts and content are otherwise preserved. Real Contact
delivery, the native Linux full suite and clean-Mac acceptance remain unverified.

Install it with:

```sh
curl -fsSL https://fleacms.com/install.sh | bash
```

The installer detects macOS/Linux and processor, downloads over HTTPS, checks
SHA-256 and executable version, and installs without sudo to `~/.local/bin/flea`.
If needed it prints a PATH export to add to your shell profile; it never edits
the profile. `FLEA_INSTALL_DIR` may specify another absolute user-owned directory.
Required V1 platforms are Apple Silicon macOS and x86-64 Linux. Other platforms
must pass native verification before being included in published metadata.

The 0.8.0 public installer passed fresh installation and updates from 0.7.1 on
an existing Mac and Ubuntu host, with installed checksums matching the published
artifacts. Linux checks ran unprivileged. Site creation and validation also passed.
These are not
clean-OS tests. The Mac executable is ad-hoc signed, not notarized; independent
clean-machine testing remains pending. The website presents this as an early
test release so invited testers can provide feedback.

For 0.8.0, update existing sites' `site/admin/_toolbar.html` from the current
starter alongside the executable. Its edit-row wrapper is required for the new
sliding toolbar. Preserve each site's custom layouts and content.

`flea update` updates only the local CLI and embedded Linux servers. Downloads
are bounded and checked for checksum, platform and version before replacement.
Failed validation preserves the installed CLI; newer local versions are not
downgraded. Remote servers change only when you explicitly run this from a site:

```sh
flea update-server
```

This uses the branch's SSH destination and confirms before restarting all sites
on that server. Interrupted updates use `flea update-server --recover`.
There is no `flea upgrade` alias. See [Recovery](recovery.md) for boundaries.

Both local installation paths use `.flea-install.lock` in the destination.
After an interruption, confirm no installer/updater is running before removing
that empty directory. Neither path edits shell profiles or remote servers.

HTTPS delivery from fleacms.com is the release trust boundary. Checksums detect
corruption, not a compromised publisher; they are not independent signatures.
Protect publishing credentials and immutable versioned artifacts. Clean-machine
Mac trust/signing checks and native Linux verification remain release gates.
