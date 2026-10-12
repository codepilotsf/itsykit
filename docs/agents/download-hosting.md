# Disk-backed release downloads

Flea 0.7.0 supports a host-managed store for its downloadable CLIs. This is not
a general public filesystem mount or an upload API. Normal sites need no change.

In administrator-owned `/etc/flea/server.json`, `downloads` maps existing site
hostnames to a clean absolute directory. Configure all aliases of a site equally:

```json
{
  "downloads": {
    "fleacms.com": "/var/lib/flea/downloads/fleacms",
    "www.fleacms.com": "/var/lib/flea/downloads/fleacms"
  }
}
```

Merge this field into the existing configuration; never replace the whole config
with this fragment. Upgrade to 0.7.0 before adding it. Keep the existing TLS,
redirect, content, theme and session settings. Never put this field in site config.

## Store and serving boundary

The root contains stable SemVer directories such as `0.7.0/`, each containing
`SHA256SUMS` and only the binaries listed in that manifest. Allowed names are
`flea-{darwin,linux}-{amd64,arm64}`; publish only native-verified platforms.
The root is administrator-owned and read-only to the service. Use directories
0755 and files 0644, no symlinks or hardlinked files. Group/world-writable
directories and files are rejected. Never change published bytes in place.

Limits: 128 versions, four platform binaries per version, 128 MiB per binary,
8 KiB per manifest. These bound catalogue size and open handles without buffering
binary contents. SHA-256 is verified with streaming reads at load. GET, HEAD and
Range are supported through pinned handles; directories and unknown paths return
404, other methods 405. Unknown hostnames cannot access another host's store.
Versioned responses have immutable caching. Size/mtime changes fail closed.

The store owns `/releases/<version>/<filename>` on configured hosts. It does not
serve arbitrary files or fall back to site assets for unknown versioned URLs.
`/releases/latest.txt` remains a normal website asset, so deployments and website
rollback control the public update pointer. Published old binaries remain available.

## Publishing safely

1. Build and test the exact artifacts. Hash final signed bytes. Keep a private
   local release archive outside the website; binary history must not enter future
   website commits or snapshots. Old Git history may still contain past artifacts.
2. Upload a complete version directory to a fresh private staging directory on
   disk, beside (not inside) the active store. Use ordinary SSH/SCP with host-key
   checking. Transfer only binaries and SHA256SUMS, never source or private state.
3. Check `sha256sum -c SHA256SUMS` in staging, exact filenames, permissions, native
   execution and version. Do not replace an existing version directory. For an
   idempotent retry, compare the full existing version and stop if anything differs.
4. Under `/etc/flea/deploy.lock`, check there are no pending activation/upgrade
   journals. Move the complete directory into the store on the same filesystem,
   with no overwrite. Validate using `GOMEMLIMIT=320MiB flea-server -validate`.
   Restart only after successful validation. The running service keeps its old
   catalogue until restarted. A failed validation must not advance latest.txt.
5. Verify public HTTPS downloads and hashes, including an old version. Verify
   service health and all hosted sites. If restart fails, keep the old channel and
   restore the reviewed host configuration; retain all versioned bytes for recovery.
6. Only then deploy the small website change advancing latest.txt and its download
   links. Run fresh install/update checks on every advertised native platform.

On this host, use the existing 320 MiB Go target for both service and admin
processes and verify swap. Separating downloads reduces asset RAM and upload size;
it is not permission to remove the existing memory safeguards.

## Migration and rollback

Copy all existing version directories to the store without changing bytes. Verify
every manifest, upgrade the executable, enable and validate the host mapping, then
verify old URLs. Only after those checks remove the version directories from the
website's current tree and deploy it. Retain latest.txt, installer and source history.
Preserve a local artifact archive and old server releases. No published binary is
deleted from the download service. Do not rewrite Git history.

Host configuration changes require a private backup and exclusive deployment lock.
Site rollback does not roll back the external store. Site backups do not include
it: back up this directory separately and verify its manifests on restore.
Older than 0.7 server binaries reject the new host field. An executable downgrade
therefore requires restoring a compatible config and a website snapshot containing
the old downloads, or another separately verified download-serving arrangement.
Do not blindly invoke an old upgrade recovery executable after enabling this field.

The website's RELEASES.md records actual artifact checks and deployment IDs.
