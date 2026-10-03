# Installation and updates

Version 0.7.1 is the current published test release for Apple Silicon macOS and
x86-64 Linux. The public installer selects the latest stable test release.
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

The 0.7.1 public installer passed fresh installation and updates from 0.7.0 on
an existing Mac and Ubuntu host, with installed checksums matching the published
artifacts. Linux checks ran unprivileged. Site creation and validation also passed.
These are not
clean-OS tests. The Mac executable is ad-hoc signed, not notarized; independent
clean-machine testing remains pending. The website presents this as an early
test release so invited testers can provide feedback.

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
